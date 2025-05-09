package AIService

import "github.com/sashabaranov/go-openai"

var AITools = []openai.Tool{
	{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "get_all_categories",
			Description: "Retrieve all existing blog categories from the database",
			Parameters: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
				"required":   []string{},
			},
		},
	},
	{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "get_all_tags",
			Description: "Retrieve all existing blog tags from the database",
			Parameters: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
				"required":   []string{},
			},
		},
	},
	{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "create_category",
			Description: "Create a new blog category. You MUST call this if the suggested category does not exist in the list retrieved from get_all_categories.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "The URL-friendly slug of the category (must be lowercase, no spaces, use hyphens)",
					},
					"value": map[string]any{
						"type":        "string",
						"description": "The display name of the category",
					},
				},
				"required": []string{"name", "value"},
			},
		},
	},
	{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "create_tag",
			Description: "Create a new blog tag. You MUST call this if the suggested tag does not exist in the list retrieved from get_all_tags.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "The URL-friendly slug of the tag (must be lowercase, no spaces, use hyphens)",
					},
					"value": map[string]any{
						"type":        "string",
						"description": "The display name of the tag",
					},
				},
				"required": []string{"name", "value"},
			},
		},
	},
	{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "finalize_metadata",
			Description: "Generate the final metadata based on analysis",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"title": map[string]any{
						"type":        "string",
						"description": "SEO-friendly title, maximum 60 characters",
					},
					"description": map[string]any{
						"type":        "string",
						"description": "SEO-friendly description, maximum 160 characters",
					},
					"categories": map[string]any{
						"type":        "array",
						"description": "List of 1-2 categories that best match the content",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"name": map[string]any{
									"type":        "string",
									"description": "Category name",
								},
								"value": map[string]any{
									"type":        "string",
									"description": "Category slug value",
								},
							},
						},
					},
					"tags": map[string]any{
						"type":        "array",
						"description": "List of 3-8 tags that best match the content",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"name": map[string]any{
									"type":        "string",
									"description": "Tag name",
								},
								"value": map[string]any{
									"type":        "string",
									"description": "Tag slug value",
								},
							},
						},
					},
				},
				"required": []string{"title", "description", "categories", "tags"},
			},
		},
	},
}
