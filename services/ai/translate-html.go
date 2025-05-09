package AIService

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/sashabaranov/go-openai"
)

// TranslateHTML, HTML içeriğini belirtilen kaynak dilden hedef dile çevirir
// Token limitlerini aşmamak için içeriği güvenli şekilde parçalara böler
func (s *AIService) TranslateHTML(
	ctx context.Context,
	html string,
	sourceLanguage string,
	targetLanguage string,
	maxTokensPerChunk int,
	maxChunkCount int,
) (string, int, error) {
	// 1. İçeriği güvenli şekilde parçalara böl
	htmlChunks := s.splitHTMLSafely(html, maxTokensPerChunk)

	// 2. Parça sayısı limiti aşıyorsa hata döndür
	if len(htmlChunks) > maxChunkCount {
		return "", 0, fmt.Errorf(
			"HTML içeriği çok uzun: maksimum %d token destekleniyor.",
			maxChunkCount*maxTokensPerChunk,
		)
	}

	// 3. Sonuçları tutacak değişkenleri oluştur
	translatedChunks := make([]string, len(htmlChunks))
	errorChunks := make([]error, len(htmlChunks))
	tokenCounts := make([]int, len(htmlChunks))

	// 4. Paralel işlem için WaitGroup oluştur
	var wg sync.WaitGroup
	wg.Add(len(htmlChunks))

	// 5. Her parçayı paralel olarak çevir
	for i, chunk := range htmlChunks {
		go func(index int, htmlChunk string) {
			defer wg.Done()

			// Çeviri işlemini gerçekleştir
			translatedChunk, tokensUsed, err := s.translateChunk(
				ctx,
				htmlChunk,
				sourceLanguage,
				targetLanguage,
			)

			// Sonuçları kaydet
			translatedChunks[index] = translatedChunk
			tokenCounts[index] = tokensUsed
			errorChunks[index] = err
		}(i, chunk)
	}

	// 6. Tüm goroutine'lerin tamamlanmasını bekle
	wg.Wait()

	// 7. Herhangi bir hata oluştuysa döndür
	for i, err := range errorChunks {
		if err != nil {
			return "", s.sum(tokenCounts), fmt.Errorf("chunk %d çevirme hatası: %w", i, err)
		}
	}

	// 8. Toplam token kullanımını hesapla
	totalTokensUsed := s.sum(tokenCounts)

	// 9. Çevrilen parçaları birleştir ve sonucu döndür
	return strings.Join(translatedChunks, ""), totalTokensUsed, nil
}

// sum, bir sayı dizisindeki tüm değerleri toplar
func (s *AIService) sum(numbers []int) int {
	total := 0
	for _, n := range numbers {
		total += n
	}
	return total
}

// translateChunk, tek bir HTML parçasını çevirir
func (s *AIService) translateChunk(
	ctx context.Context,
	chunk string,
	sourceLanguage string,
	targetLanguage string,
) (string, int, error) {
	// 1. Sistem talimatını oluştur - daha detaylı ve kesin
	systemPrompt := fmt.Sprintf(`You are a professional %s-to-%s translator specialized in blog posts with complex HTML structures.

Your task is ABSOLUTELY CRITICAL: Translate a blog post's HTML content while preserving its EXACT structure.

IMPORTANT RULES TO FOLLOW STRICTLY:
1. TRANSLATE EVERY SINGLE TEXT NODE, no matter how small or where it appears in the HTML.
2. NEVER omit, skip or remove ANY HTML elements, including complex ones like React components, carousels, or interactive elements.
3. PRESERVE ALL HTML tags, attributes, class names, data attributes, and structure EXACTLY as they are.
4. DO NOT modify any URLs, file paths, image sources, or technical attributes.
5. Keep all formatting, styling, and functionality intact.
6. Translate button texts, labels, alt texts, and any human-readable content.
7. Your translation must be accurate and natural in %s.
8. Remember: This is a blog editor's content being translated for readers - ALL text content must be translated!

IF YOU SEE COMPLEX COMPONENTS (like Instagram embeds, React components, or carousels):
- These are ESPECIALLY IMPORTANT to preserve completely
- Translate ONLY the visible text within them
- Keep all class names, attributes and structure exactly as they are
- DO NOT skip or remove them thinking they are too complex

This is a critical professional task - every single element must be preserved and all visible text translated!`,
		sourceLanguage,
		targetLanguage,
		targetLanguage,
	)

	// 2. Kullanıcı mesajını oluştur - daha spesifik ve direktif
	userPrompt := fmt.Sprintf(`Translate the following blog post HTML content from %s to %s.

CRITICAL INSTRUCTIONS:
- This is a real blog post from an editor that needs EVERY TEXT ELEMENT translated
- Preserve ALL HTML structure completely intact
- Translate ALL text content, including:
  * Headings, paragraphs, and lists
  * Button texts and navigation labels
  * Image alt texts and descriptions
  * Small UI texts inside components
  * Placeholders and interactive element texts
  * ANY text that would be visible to a reader

Return the complete translated HTML with ALL elements preserved. Do not omit or skip ANY section of the HTML no matter how complex it appears.

HTML Content to translate:
%s`,
		sourceLanguage,
		targetLanguage,
		chunk,
	)

	// Geri kalan kod aynı...
	resp, err := s.AIRepo.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: "gpt-4.1-nano",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPrompt,
				},
			},
			Temperature: 0.1, // Tutarlı çeviriler için düşük sıcaklık
		},
	)

	// Hata kontrolü...
	if err != nil {
		return "", 0, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", 0, fmt.Errorf("empty response from OpenAI API")
	}

	// Toplam token kullanımını al
	totalTokens := resp.Usage.TotalTokens

	// Çevrilen içeriği ve token kullanımını döndür
	return resp.Choices[0].Message.Content, totalTokens, nil
}

// splitHTMLSafely, HTML içeriğini daha küçük parçalara böler ve HTML yapısını korur
// HTML'i güvenli noktalarda (kapatma etiketleri gibi) bölmeye çalışır
func (s *AIService) splitHTMLSafely(html string, maxChunkSize int) []string {
	// 1. İçerik zaten küçükse doğrudan döndür
	if len(html) <= maxChunkSize {
		return []string{html}
	}

	// 2. Parçaları tutacak dizi ve kalan içerik
	var chunks []string
	remaining := html

	// 3. Tüm içerik parçalanana kadar devam et
	for len(remaining) > 0 {
		// Kalan içerik maksimum boyuttan küçükse
		if len(remaining) <= maxChunkSize {
			chunks = append(chunks, remaining)
			break
		}

		// Güvenli bir bölme noktası bul
		safePoint := s.findSafeHTMLSplitPoint(remaining, maxChunkSize)
		if safePoint <= 0 {
			// Güvenli nokta bulunamazsa maxChunkSize'ı kullan
			safePoint = maxChunkSize
		}

		// Parçayı ekle ve kalanı güncelle
		chunks = append(chunks, remaining[:safePoint])
		remaining = remaining[safePoint:]
	}

	return chunks
}

// findSafeHTMLSplitPoint, HTML içeriğini bölmek için uygun bir nokta bulur
// HTML yapısını bozmadan, öncelikle kapatma etiketlerini kullanarak
func (s *AIService) findSafeHTMLSplitPoint(html string, maxPos int) int {
	// 1. maxPos içerik uzunluğundan büyükse, içeriğin tamamını döndür
	if maxPos >= len(html) {
		return len(html)
	}

	// 2. Olası güvenli bölme noktaları (öncelik sırasına göre)
	candidates := []string{
		"</div>", "</p>", "</h1>", "</h2>", "</h3>", "</h4>", "</h5>", "</h6>",
		"</span>", "</a>", "</li>", "</ul>", "</ol>", ">", ";", ".",
	}

	// 3. Her bir aday için maxPos'dan geriye doğru tara
	for _, candidate := range candidates {
		lastIndex := -1
		pos := 0

		// maxPos'a kadar olan tüm aday etiket oluşumlarını bul
		for pos < maxPos {
			index := strings.Index(html[pos:], candidate)
			if index == -1 {
				break
			}

			foundPos := pos + index + len(candidate)
			if foundPos <= maxPos {
				lastIndex = foundPos
				pos = foundPos
			} else {
				break
			}
		}

		// Kapatma etiketi bulunduğunda bu konumu kullan
		if lastIndex > 0 {
			return lastIndex
		}
	}

	// 4. Güvenli etiket bulunamazsa, bir kelime sınırında bölmeyi dene
	for i := maxPos; i > maxPos-50 && i > 0; i-- {
		if html[i] == ' ' {
			return i
		}
	}

	// 5. Güvenli bölme noktası bulunamadı
	return -1
}
