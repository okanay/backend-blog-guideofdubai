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
	// 1. Sistem talimatını oluştur
	systemPrompt := fmt.Sprintf(`You are a professional %s-to-%s translator.
Your task is to translate the provided HTML content.
IMPORTANT: Translate ONLY the text content while preserving ALL HTML tags, attributes, and structure.
Keep all links, formatting, and HTML elements intact.
Your translation should be natural and fluent in the target language.`,
		sourceLanguage,
		targetLanguage,
	)

	// 2. Kullanıcı mesajını oluştur
	userPrompt := fmt.Sprintf(`Translate the following HTML content from %s to %s.
Preserve ALL HTML tags and structure - translate ONLY the text content.
Return the translated HTML as plain text without using markdown or code blocks.

HTML Content:
%s`,
		sourceLanguage,
		targetLanguage,
		chunk,
	)

	// 3. OpenAI API isteğini oluştur ve gönder
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

	// 4. Hata kontrolü
	if err != nil {
		return "", 0, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", 0, fmt.Errorf("empty response from OpenAI API")
	}

	// 5. Toplam token kullanımını al
	totalTokens := resp.Usage.TotalTokens

	// 6. Çevrilen içeriği ve token kullanımını döndür
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
