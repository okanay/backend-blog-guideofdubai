// services/ai_service.go

package AIService

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/sashabaranov/go-openai"
)

func (s *AIService) GenerateMetadataWithTools(
	ctx context.Context,
	html string,
	language string,
	userID uuid.UUID,
) (*types.GenerateMetadataResponse, int, error) {

	messages := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf(`
			You are an AI assistant responsible for generating SEO metadata for blog posts.

			Your task is to analyze the provided blog content and generate:
			- An SEO-friendly title (maximum 60 characters)
			- An SEO-friendly description (maximum 160 characters)

			**IMPORTANT:**
			- Both the title and description MUST be written in the following language: %s.
			- Do NOT use any other language, even partially. All output must be in %s only.
			- The title and description should be clear, compelling, and relevant to the blog content.
			- The title should be concise and attractive for search engines.
			- The description should summarize the content and encourage users to read the blog post.

			**Output format:**
			{
			  "title": "<SEO-friendly title in %s>",
			  "description": "<SEO-friendly description in %s>"
			}

			Now, analyze the following blog content and generate the title and description according to the instructions above.
			`, language, language, language, language),
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: fmt.Sprintf("Please generate appropriate SEO metadata for this blog content:\n\n%s", html),
		},
	}

	var finalMetadata *types.GenerateMetadataResponse
	maxIterations := 10
	totalTokensUsed := 0

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
		return nil, totalTokensUsed, fmt.Errorf("Failed to convert schema to JSON: %v", err)
	}

	for range make([]struct{}, maxIterations) {
		if finalMetadata != nil {
			break
		}

		resp, err := s.AIRepo.Client().CreateChatCompletion(
			ctx,
			openai.ChatCompletionRequest{
				Model:       "gpt-4.1-nano",
				Messages:    messages,
				Tools:       s.Tools,
				ToolChoice:  "auto",
				Temperature: 0.1,
				ResponseFormat: &openai.ChatCompletionResponseFormat{
					Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
					JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
						Name:        "BlogMetadata",
						Description: "SEO metadata for a blog post",
						Schema:      json.RawMessage(schemaBytes),
						Strict:      true,
					},
				},
			},
		)

		if err != nil {
			return nil, totalTokensUsed, fmt.Errorf("OpenAI API error: %w", err)
		}
		if len(resp.Choices) == 0 {
			return nil, totalTokensUsed, fmt.Errorf("empty response from OpenAI API")
		}

		totalTokensUsed += resp.Usage.TotalTokens
		assistantMessage := resp.Choices[0].Message
		messages = append(messages, assistantMessage)

		if assistantMessage.ToolCalls != nil && len(assistantMessage.ToolCalls) > 0 {
			for _, toolCall := range assistantMessage.ToolCalls {
				functionName := toolCall.Function.Name
				functionArgs := toolCall.Function.Arguments
				log.Println("[AI] :: Calling This Function :", functionName)

				functionResult, metadata, err := s.DispatchToolCall(ctx, functionName, functionArgs, userID)
				if err != nil {
					functionResult = fmt.Sprintf(`{"error": "%s"}`, err.Error())
				}
				if metadata != nil {
					finalMetadata = metadata
				}

				messages = append(messages, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    functionResult,
					ToolCallID: toolCall.ID,
				})
			}
		} else if assistantMessage.Content != "" {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: "Please use the tools to complete the metadata generation. Call finalize_metadata when you're ready to submit the final metadata.",
			})
		}
	}

	return finalMetadata, totalTokensUsed, nil
}
