package AIService

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/sashabaranov/go-openai"
)

// --- DATA STRUCTS ---

type TextItem struct {
	Index           int      `json:"index"`
	Path            []string `json:"path"`
	Original        string   `json:"original"`
	StartsWithSpace bool     `json:"startsWithSpace,omitempty"`
	EndsWithSpace   bool     `json:"endsWithSpace,omitempty"`
}

type TextItemTranslation struct {
	Index           int      `json:"index"`
	Path            []string `json:"path"`
	Original        string   `json:"original"`
	Translated      string   `json:"translated"`
	StartsWithSpace bool     `json:"startsWithSpace,omitempty"`
	EndsWithSpace   bool     `json:"endsWithSpace,omitempty"`
}

type TextItemTranslationResponse struct {
	Items []TextItemTranslation `json:"items"`
}

// --- TRANSLATABLE PATHS ---

// Her bir component tipi için çevrilecek alanların yollarını tanımlar
var translatablePaths = []struct {
	Type  string   // Component tipi
	Paths []string // İçindeki çevrilecek alanların yolları
}{
	{Type: "text", Paths: []string{"text"}},
	{Type: "enhancedImage", Paths: []string{"attrs.alt", "attrs.title", "attrs.caption"}},
	{Type: "image", Paths: []string{"attrs.alt", "attrs.title", "attrs.caption"}},
	{Type: "instagramCarousel", Paths: []string{"attrs.cards.*.caption", "attrs.cards.*.alt", "attrs.cards.*.title"}},
	// İleride yeni component tipleri buraya eklenebilir
}

// --- MAIN TRANSLATION FUNCTION ---

func (s *AIService) TranslateBlogPostJSON(
	ctx context.Context,
	jsonContent string,
	sourceLanguage string,
	targetLanguage string,
) (translatedJSON string, inputTokens int, outputTokens int, err error) {
	// JSON'ı ayrıştır
	var doc any
	if err := json.Unmarshal([]byte(jsonContent), &doc); err != nil {
		return "", 0, 0, fmt.Errorf("JSON parse error: %w", err)
	}

	// Çevirilecek metinleri topla
	items := extractTranslatableTexts(doc)
	if len(items) == 0 {
		return jsonContent, 0, 0, nil // Çevirilecek metin yoksa aynısını geri döndür
	}

	// Metinleri batchlere ayır
	batchSize := 8
	batches := chunkTextItems(items, batchSize)

	// Her batch'i paralel olarak çevir
	allTranslations := make([]TextItemTranslation, 0, len(items))
	var totalInputTokens, totalOutputTokens int
	var wg sync.WaitGroup
	errs := make([]error, len(batches))
	results := make([][]TextItemTranslation, len(batches))

	for i, batch := range batches {
		wg.Add(1)
		go func(i int, batch []TextItem) {
			defer wg.Done()
			trans, inTok, outTok, err := s.translateTextItemBatch(ctx, batch, sourceLanguage, targetLanguage, i)
			if err != nil {
				errs[i] = err
				return
			}

			results[i] = trans
			totalInputTokens += inTok
			totalOutputTokens += outTok
		}(i, batch)
	}
	wg.Wait()

	// Hataları kontrol et
	for _, err := range errs {
		if err != nil {
			return "", 0, 0, err
		}
	}

	// Tüm çevirileri birleştir
	for _, batchTrans := range results {
		allTranslations = append(allTranslations, batchTrans...)
	}

	// Çevirileri orijinal JSON'a yerleştir
	for _, trans := range allTranslations {
		// Boşlukları geri ekle
		translatedWithSpaces := restoreSpaces(trans.Translated, trans.StartsWithSpace, trans.EndsWithSpace)

		if err := setValueAtPath(doc, trans.Path, translatedWithSpaces); err != nil {
			return "", 0, 0, fmt.Errorf("Failed to set translation at path %v: %w", trans.Path, err)
		}
	}

	// Çevirisi tamamlanan JSON'ı string olarak döndür
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", 0, 0, fmt.Errorf("JSON marshal error: %w", err)
	}

	return string(out), totalInputTokens, totalOutputTokens, nil
}

// --- TEXT EXTRACTION ---

// JSON içinden çevirilecek tüm metinleri çıkarır
func extractTranslatableTexts(doc any) []TextItem {
	items := []TextItem{}
	index := 0
	traverseJSON(doc, []string{}, &items, &index)
	return items
}

// JSON'ı recursive olarak dolaşıp çevirilecek metinleri bulur
func traverseJSON(data any, path []string, items *[]TextItem, index *int) {
	switch v := data.(type) {
	case map[string]any:
		// Component tipini kontrol et
		typeVal, hasType := v["type"].(string)
		if hasType {
			// Bu component tipi için çevirilecek alanları bul
			for _, translatable := range translatablePaths {
				if translatable.Type == typeVal {
					// Bu tip için belirtilen tüm alanları kontrol et
					for _, fieldPath := range translatable.Paths {
						extractPathValue(v, fieldPath, path, items, index)
					}
				}
			}
		}

		// Alt öğeleri dolaş (her alan için recursive çağrı)
		for key, val := range v {
			traverseJSON(val, append(path, key), items, index)
		}

	case []any:
		// Dizi elemanlarını dolaş
		for i, item := range v {
			traverseJSON(item, append(path, strconv.Itoa(i)), items, index)
		}
	}
}

// Belirtilen path'deki değeri çıkarır (wildcard destekli)
func extractPathValue(obj map[string]any, fieldPath string, basePath []string, items *[]TextItem, index *int) {
	parts := strings.Split(fieldPath, ".")
	extractPathRecursive(obj, parts, 0, basePath, items, index)
}

// Path'i recursive olarak izleyerek değeri çıkarır
func extractPathRecursive(obj any, parts []string, partIndex int, basePath []string, items *[]TextItem, index *int) {
	if partIndex >= len(parts) {
		return
	}

	part := parts[partIndex]
	isLastPart := partIndex == len(parts)-1

	switch v := obj.(type) {
	case map[string]any:
		if part == "*" {
			// Wildcard: Tüm map elemanlarını dolaş
			for key, val := range v {
				newPath := append(basePath, key)
				if isLastPart {
					extractFinalValue(val, newPath, items, index)
				} else {
					extractPathRecursive(val, parts, partIndex+1, newPath, items, index)
				}
			}
		} else {
			// Normal key: Belirtilen anahtarı kontrol et
			if val, ok := v[part]; ok {
				newPath := append(basePath, part)
				if isLastPart {
					extractFinalValue(val, newPath, items, index)
				} else {
					extractPathRecursive(val, parts, partIndex+1, newPath, items, index)
				}
			}
		}

	case []any:
		if part == "*" {
			// Wildcard: Tüm dizi elemanlarını dolaş
			for i, item := range v {
				indexStr := strconv.Itoa(i)
				newPath := append(basePath, indexStr)
				if isLastPart {
					extractFinalValue(item, newPath, items, index)
				} else {
					extractPathRecursive(item, parts, partIndex+1, newPath, items, index)
				}
			}
		} else {
			// Normal index: Belirtilen indeksi kontrol et
			idx, err := strconv.Atoi(part)
			if err == nil && idx >= 0 && idx < len(v) {
				newPath := append(basePath, part)
				if isLastPart {
					extractFinalValue(v[idx], newPath, items, index)
				} else {
					extractPathRecursive(v[idx], parts, partIndex+1, newPath, items, index)
				}
			}
		}
	}
}

// Son değeri çıkarır (metin ise çeviri listesine ekler)
func extractFinalValue(val any, path []string, items *[]TextItem, index *int) {
	// Null kontrolü ekle
	if val == nil {
		return
	}

	// String değeri kontrol et
	if str, ok := val.(string); ok {
		// Boş string kontrolü
		if strings.TrimSpace(str) == "" {
			return // Boş stringler için işlem yapma
		}

		trimmed, startsWithSpace, endsWithSpace := analyzeSpaces(str)
		*items = append(*items, TextItem{
			Index:           *index,
			Path:            path,
			Original:        trimmed,
			StartsWithSpace: startsWithSpace,
			EndsWithSpace:   endsWithSpace,
		})
		*index++
	}
}

// --- BATCH TRANSLATION WITH JSON SCHEMA ---

func (s *AIService) translateTextItemBatch(
	ctx context.Context,
	batch []TextItem,
	sourceLanguage, targetLanguage string,
	batchIndex int,
) ([]TextItemTranslation, int, int, error) {
	prompt := buildTextItemTranslationPrompt(batch, sourceLanguage, targetLanguage)

	schemaBytes, err := getTextItemTranslationSchema()
	if err != nil {
		return nil, 0, 0, fmt.Errorf("Schema error: %w", err)
	}

	resp, err := s.AIRepo.Client().CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: "gpt-4.1-nano",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a professional translator that PRESERVES EXACT WHITESPACE POSITIONS in your translations.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.1,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
				JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
					Name:        "TextItemTranslation",
					Description: "Translation of text items with index, path, original, translated, and whitespace flags.",
					Schema:      json.RawMessage(schemaBytes),
					Strict:      true,
				},
			},
		},
	)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("OpenAI API error: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, 0, 0, fmt.Errorf("empty response from OpenAI API")
	}
	content := strings.TrimSpace(resp.Choices[0].Message.Content)

	var respObj TextItemTranslationResponse
	if err := json.Unmarshal([]byte(content), &respObj); err != nil {
		return nil, 0, 0, fmt.Errorf("JSON parse error: %w\nRaw content: %s", err, content)
	}
	if len(respObj.Items) != len(batch) {
		return nil, 0, 0, fmt.Errorf("Translation count mismatch: got %d, expected %d", len(respObj.Items), len(batch))
	}

	// StartsWithSpace ve EndsWithSpace değerlerini orijinal batch'ten kopyala
	for i := range respObj.Items {
		respObj.Items[i].StartsWithSpace = batch[i].StartsWithSpace
		respObj.Items[i].EndsWithSpace = batch[i].EndsWithSpace
	}

	return respObj.Items, resp.Usage.PromptTokens, resp.Usage.CompletionTokens, nil
}

// --- UTILITY FUNCTIONS ---

// Boşluk analizi fonksiyonu
func analyzeSpaces(text string) (string, bool, bool) {
	trimmed := strings.TrimSpace(text)
	startsWithSpace := len(text) > 0 && unicode.IsSpace(rune(text[0]))
	endsWithSpace := len(text) > 0 && unicode.IsSpace(rune(text[len(text)-1]))
	return trimmed, startsWithSpace, endsWithSpace
}

// Boşlukları geri ekleme fonksiyonu
func restoreSpaces(text string, startsWithSpace, endsWithSpace bool) string {
	if startsWithSpace {
		text = " " + text
	}
	if endsWithSpace {
		text = text + " "
	}
	return text
}

// TextItem'ları batchlere ayırır
func chunkTextItems(items []TextItem, size int) [][]TextItem {
	var batches [][]TextItem
	for i := 0; i < len(items); i += size {
		end := i + size
		if end > len(items) {
			end = len(items)
		}
		batches = append(batches, items[i:end])
	}
	return batches
}

// Path'e göre JSON içindeki değeri günceller, yol bulunamazsa hata yerine log bırakır
func setValueAtPath(obj any, path []string, value string) error {
	if len(path) == 0 {
		return fmt.Errorf("empty path")
	}

	current := obj
	for i, key := range path {
		isLast := i == len(path)-1

		switch v := current.(type) {
		case map[string]any:
			if isLast {
				v[key] = value
				return nil
			}

			next, ok := v[key]
			if !ok {
				// Anahtar bulunamadı - hata vermek yerine durumu loglayıp devam et
				fmt.Printf("Warning: Path %v could not be fully traversed, key %s not found. Skipping this translation.\n", path, key)
				return nil // Hata değil, null dönüyor böylece işlem devam eder
			}
			current = next

		case []any:
			idx, err := strconv.Atoi(key)
			if err != nil || idx < 0 || idx >= len(v) {
				// Geçersiz dizi indeksi - hata vermek yerine durumu loglayıp devam et
				fmt.Printf("Warning: Path %v could not be fully traversed, invalid array index: %s. Skipping this translation.\n", path, key)
				return nil
			}

			if isLast {
				v[idx] = value
				return nil
			}
			current = v[idx]

		default:
			// Beklenmeyen tip - hata vermek yerine durumu loglayıp devam et
			fmt.Printf("Warning: Path %v could not be fully traversed, unexpected type at %v. Skipping this translation.\n", path, path[:i+1])
			return nil
		}
	}

	fmt.Printf("Warning: Could not set value at path: %v. Skipping this translation.\n", path)
	return nil
}

// Çeviri isteği prompt'unu oluşturur
func buildTextItemTranslationPrompt(items []TextItem, sourceLanguage, targetLanguage string) string {
	inputJSON, _ := json.MarshalIndent(map[string]any{"items": items}, "", "  ")
	return fmt.Sprintf(
		`You are a professional translator. IMPORTANT: You must preserve the exact meaning, style, and formatting of the text.

		Below is a JSON object with an "items" array. Each item has "index", "path", "original", "startsWithSpace", and "endsWithSpace" fields.
Translate the "original" field of each item from %s to %s.

STRICT TRANSLATION RULES:
1. EXACTLY preserve all symbols, special characters, and punctuation like &, -, +, @, /, etc.
2. DO NOT replace "&" with numbers like "0", "06" or any other characters
3. DO NOT add any numbers, digits, or random characters that aren't in the original
4. For example, "Dubai Aquarium & Underwater Zoo" should be translated WITHOUT changing the "&" symbol
5. Preserve formatting elements exactly as they appear
6. Maintain all HTML tags and links unchanged
7. CRITICAL: Do not add or insert any numbers in place of symbols

Example of CORRECT translation:
Original: "Visit the Dubai Aquarium & Underwater Zoo"
Correct: "Dubai Akvaryumu & Su Altı Hayvanat Bahçesini Ziyaret Edin"
INCORRECT: "Dubai Akvaryumu 0 ve Su Altı Hayvanat Bahçesini Ziyaret Edin"
INCORRECT: "Dubai Akvaryumu 06 Su Altı Hayvanat Bahçesini Ziyaret Edin"

Input:
%s

Output:`, sourceLanguage, targetLanguage, string(inputJSON))
}

// TextItemTranslation şemasını oluşturur
// TextItemTranslation şemasını oluşturur
func getTextItemTranslationSchema() ([]byte, error) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"items": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"index": map[string]any{
							"type":        "integer",
							"description": "Index of the text item in the original array",
						},
						"path": map[string]any{
							"type":        "array",
							"items":       map[string]any{"type": "string"},
							"description": "Path to the text item in the original JSON",
						},
						"original": map[string]any{
							"type":        "string",
							"description": "Original text to be translated",
						},
						"translated": map[string]any{
							"type":        "string",
							"description": "Translated text",
						},
						"startsWithSpace": map[string]any{
							"type":        "boolean",
							"description": "Whether the original text starts with a space",
						},
						"endsWithSpace": map[string]any{
							"type":        "boolean",
							"description": "Whether the original text ends with a space",
						},
					},
					"required":             []string{"index", "path", "original", "translated", "startsWithSpace", "endsWithSpace"},
					"additionalProperties": false,
				},
			},
		},
		"required":             []string{"items"},
		"additionalProperties": false,
	}

	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("Failed to convert schema to JSON: %v", err)
	}

	return schemaBytes, nil
}
