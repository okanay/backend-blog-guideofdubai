package types

import (
	"encoding/json"
	"fmt"
)

type GenerateMetadataRequest struct {
	HTML     string `json:"html" binding:"required"`
	Language string `json:"language" binding:"required"`
}

type GenerateMetadataResponse struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func GetMetadataPromptSchema() ([]byte, error) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"title": map[string]any{
				"type":        "string",
				"description": "SEO-friendly title, maximum 120 characters",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "SEO-friendly description, maximum 200 characters",
			},
		},
		"required":             []string{"title", "description"},
		"additionalProperties": false,
	}

	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("Failed to convert schema to JSON: %v", err)
	}

	return schemaBytes, nil
}
