package AIRepository

import (
	"context"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// TranslateHTML translates the given HTML content to the specified target language
// It safely splits long HTML content into chunks to avoid token limitations
func (r *Repository) TranslateHTML(ctx context.Context, html string, sourceLanguage, targetLanguage string, maxTokensPerChunk int) (string, int, error) {
	// HTML'i güvenli parçalara böl
	htmlChunks := splitHTMLSafely(html, maxTokensPerChunk)

	// Her parçayı ayrı ayrı çevir
	translatedChunks := make([]string, len(htmlChunks))
	totalTokensUsed := 0

	for i, chunk := range htmlChunks {
		translatedChunk, tokensUsed, err := r.translateChunk(ctx, chunk, sourceLanguage, targetLanguage)
		if err != nil {
			return "", totalTokensUsed, fmt.Errorf("error translating chunk %d: %w", i, err)
		}
		translatedChunks[i] = translatedChunk
		totalTokensUsed += tokensUsed
	}

	// Çevrilen parçaları birleştir
	return strings.Join(translatedChunks, ""), totalTokensUsed, nil
}

// translateChunk translates a single HTML chunk
func (r *Repository) translateChunk(ctx context.Context, chunk, sourceLanguage, targetLanguage string) (string, int, error) {
	systemInstruction := fmt.Sprintf(`You are a professional %s-to-%s translator.
Your task is to translate the provided HTML content.
IMPORTANT: Translate ONLY the text content while preserving ALL HTML tags, attributes, and structure.
Keep all links, formatting, and HTML elements intact.
Your translation should be natural and fluent in the target language.`, sourceLanguage, targetLanguage)

	// User message with the HTML content to translate
	userPrompt := fmt.Sprintf(`Translate the following HTML content from %s to %s.
Preserve ALL HTML tags and structure - translate ONLY the text content.
Return the translated HTML as plain text without using markdown or code blocks.

HTML Content:
%s`, sourceLanguage, targetLanguage, chunk)

	// Create the OpenAI API request
	resp, err := r.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: "gpt-4.1-nano", // or another model
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemInstruction,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPrompt,
				},
			},
			Temperature: 0.1, // Low temperature for more consistent translations
		},
	)

	if err != nil {
		return "", 0, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", 0, fmt.Errorf("empty response from OpenAI API")
	}

	// Toplam token kullanımını al
	totalTokens := resp.Usage.TotalTokens

	// Çevirilen içeriği ve token kullanımını döndür
	return resp.Choices[0].Message.Content, totalTokens, nil
}

// splitHTMLSafely splits HTML content into smaller chunks while preserving HTML structure
// It tries to split at safe points like closing tags to maintain valid HTML
func splitHTMLSafely(html string, maxChunkSize int) []string {
	if len(html) <= maxChunkSize {
		return []string{html}
	}

	var chunks []string
	remaining := html

	for len(remaining) > 0 {
		chunkSize := maxChunkSize
		if len(remaining) <= chunkSize {
			chunks = append(chunks, remaining)
			break
		}

		// Find a safe splitting point (closing tag)
		safePoint := findSafeHTMLSplitPoint(remaining, chunkSize)
		if safePoint <= 0 {
			// If no safe point found, use maxChunkSize as fallback
			safePoint = maxChunkSize
		}

		chunks = append(chunks, remaining[:safePoint])
		remaining = remaining[safePoint:]
	}

	return chunks
}

// findSafeHTMLSplitPoint locates a suitable point to split HTML content
// without breaking HTML structure, prioritizing closing tags
func findSafeHTMLSplitPoint(html string, maxPos int) int {
	if maxPos >= len(html) {
		return len(html)
	}

	// Potential safe splitting points in order of preference
	candidates := []string{"</div>", "</p>", "</h1>", "</h2>", "</h3>", "</h4>", "</h5>", "</h6>", "</span>", "</a>", "</li>", "</ul>", "</ol>", ">", ";", "."}

	// Scan backward from maxPos to find the nearest safe splitting point
	for _, candidate := range candidates {
		lastIndex := -1
		pos := 0

		// Find all occurrences of this candidate tag up to maxPos
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

		// If we found a closing tag, use this position
		if lastIndex > 0 {
			return lastIndex
		}
	}

	// If no safe tags found, try to split at a word boundary
	for i := maxPos; i > maxPos-50 && i > 0; i-- {
		if html[i] == ' ' {
			return i
		}
	}

	return -1 // No safe splitting point found
}
