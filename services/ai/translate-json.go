package AIService

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/sashabaranov/go-openai"
)

// --- DATA STRUCTS ---

type TextItem struct {
	Index    int      `json:"index"`
	Path     []string `json:"path"`
	Original string   `json:"original"`
}

type TextItemTranslation struct {
	Index      int      `json:"index"`
	Path       []string `json:"path"`
	Original   string   `json:"original"`
	Translated string   `json:"translated"`
}

type TextItemTranslationResponse struct {
	Items []TextItemTranslation `json:"items"`
}

// --- JSON SCHEMA ---

func GetTextItemTranslationSchema() ([]byte, error) {
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
					},
					"required":             []string{"index", "path", "original", "translated"},
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

// --- TRANSLATABLE KEYS & PROTECTED KEYS ---

var translatableKeys = []string{
	"text", "caption", "alt", "title", "label", "description",
}

// Çevrilmeyecek, korunması gereken teknik alan ve zaman-tarih anahtarları
var protectedKeys = map[string]bool{
	// Teknik alanlar
	"src": true, "href": true, "imageUrl": true, "postUrl": true,
	"userProfileImage": true, "url": true, "link": true,

	// Tarih ve zaman alanları
	"timestamp": true, "date": true, "time": true, "datetime": true,
	"publishedAt": true, "createdAt": true, "updatedAt": true,

	// Sayısal alanlar
	"likesCount": true, "commentsCount": true, "viewsCount": true, "sharesCount": true,

	// Kullanıcı ve kimlik alanları
	"username": true, "userId": true, "id": true, "uuid": true, "email": true,

	// Konum alanları
	"location": true, "coordinates": true, "latitude": true, "longitude": true,

	// Diğer korunması gereken alanlar
	"type": true, "status": true, "language": true, "code": true,
}

// --- MAIN TRANSLATION FUNCTION ---

func (s *AIService) TranslateBlogPostJSON(
	ctx context.Context,
	jsonContent string,
	sourceLanguage string,
	targetLanguage string,
) (translatedJSON string, inputTokens int, outputTokens int, err error) {
	var doc any
	if err := json.Unmarshal([]byte(jsonContent), &doc); err != nil {
		return "", 0, 0, fmt.Errorf("JSON parse error: %w", err)
	}

	items := make([]TextItem, 0)
	collectTextItems(doc, []string{}, &items, 0)

	if len(items) == 0 {
		return jsonContent, 0, 0, nil
	}

	batchSize := 10
	batches := chunkTextItems(items, batchSize)

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

	for _, err := range errs {
		if err != nil {
			return "", 0, 0, err
		}
	}
	for _, batchTrans := range results {
		allTranslations = append(allTranslations, batchTrans...)
	}

	// Çevirileri orijinal JSON'a yerleştir
	for _, trans := range allTranslations {
		if err := setTextByPath(doc, trans.Path, trans.Translated); err != nil {
			return "", 0, 0, fmt.Errorf("Failed to set translation at path %v: %w", trans.Path, err)
		}
	}

	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", 0, 0, fmt.Errorf("JSON marshal error: %w", err)
	}

	return string(out), totalInputTokens, totalOutputTokens, nil
}

// --- BATCH TRANSLATION WITH JSON SCHEMA ---

func (s *AIService) translateTextItemBatch(
	ctx context.Context,
	batch []TextItem,
	sourceLanguage, targetLanguage string,
	batchIndex int,
) ([]TextItemTranslation, int, int, error) {
	prompt := buildTextItemTranslationPrompt(batch, sourceLanguage, targetLanguage)

	schemaBytes, err := GetTextItemTranslationSchema()
	if err != nil {
		return nil, 0, 0, fmt.Errorf("Schema error: %w", err)
	}

	resp, err := s.AIRepo.Client().CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: "gpt-4.1-mini",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a professional translator.",
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
					Description: "Translation of text items with index, path, original, and translated fields.",
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
	return respObj.Items, resp.Usage.PromptTokens, resp.Usage.CompletionTokens, nil
}

// --- PROMPT BUILDER ---

func buildTextItemTranslationPrompt(items []TextItem, sourceLanguage, targetLanguage string) string {
	inputJSON, _ := json.MarshalIndent(map[string]any{"items": items}, "", "  ")
	return fmt.Sprintf(
		`You are a professional translator.

Below is a JSON object with an "items" array. Each item has "index", "path", and "original".
Translate the "original" field of each item from %s to %s.
Return the same JSON object, adding a "translated" field to each item.
Do NOT change the order or any other fields.
Do NOT add, remove, merge, or split any items.
Do NOT return any explanation, only the JSON object.

Input:
%s

Output:`, sourceLanguage, targetLanguage, string(inputJSON))
}

// --- HELPER: COLLECT TEXT ITEMS ---

func collectTextItems(obj any, path []string, items *[]TextItem, idx int) int {
	switch v := obj.(type) {
	case map[string]any:
		for k, val := range v {
			// Sadece çevrilebilir anahtarları işle ve korunması gereken anahtarları atla
			if contains(translatableKeys, k) && !protectedKeys[k] {
				if str, ok := val.(string); ok {
					if strings.TrimSpace(str) != "" {
						*items = append(*items, TextItem{
							Path:     append(path, k),
							Original: str,
							Index:    idx,
						})
						idx++
					}
				}
			} else if !protectedKeys[k] {
				// Korunması gereken anahtarlar değilse, alt öğeleri işlemeye devam et
				idx = collectTextItems(val, append(path, k), items, idx)
			}
		}
	case []any:
		for i, val := range v {
			idx = collectTextItems(val, append(path, strconv.Itoa(i)), items, idx)
		}
	}
	return idx
}

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

func setTextByPath(obj any, path []string, value string) error {
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
				return fmt.Errorf("key not found: %s", key)
			}
			current = next
		case []any:
			idx, err := strconv.Atoi(key)
			if err != nil || idx < 0 || idx >= len(v) {
				return fmt.Errorf("invalid array index: %s", key)
			}
			if isLast {
				v[idx] = value
				return nil
			}
			current = v[idx]
		default:
			return fmt.Errorf("unexpected type at %v", path[:i+1])
		}
	}
	return fmt.Errorf("could not set value at path: %v", path)
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

// --- HELPER: CHUNK TEXT ITEMS ---
