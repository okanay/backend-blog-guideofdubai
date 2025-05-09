// services/ai_service.go

package AIService

import (
	"context"
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
) (*types.GenerateMetadataResponse, error) {

	messages := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf(`You are an AI assistant that generates SEO metadata for blogs. Your task is to analyze blog content and suggest appropriate title, description, categories and tags. Follow these guidelines:

			- Generate an SEO-friendly title (max 60 chars)
			- Generate a compelling description (max 160 chars)
			- First retrieve all existing categories and tags
			- For categories:
			  * Select 1-2 most relevant categories
			  * IMPORTANT: The "name" field should be the URL-friendly slug (lowercase with hyphens) and the "value" field should be the display name with proper capitalization
			  * Example: {"name": "rent-a-car", "value": "Rent A Car"}
			  * If an existing category is semantically similar to what you want to suggest, USE THE EXISTING ONE
			  * Only create a new category if there is absolutely no existing category that matches the content

			- For tags:
			  * Select 3-8 most relevant tags
			  * IMPORTANT: The "name" field should be the URL-friendly slug (lowercase with hyphens) and the "value" field should be the display name with proper capitalization
			  * Example: {"name": "tourist-attractions", "value": "Tourist Attractions"}
			  * If an existing tag is semantically similar to what you want to suggest, USE THE EXISTING ONE
			  * Only create a new tag if there is absolutely no existing tag that matches the concept

			- All metadata should be in %s language
			- You must use the available tools in this order: 1) get categories and tags, 2) create new ones ONLY IF NECESSARY, 3) finalize metadata
			- REMEMBER: Prefer using existing categories and tags whenever possible. Creating new ones should be a last resort.`, language),
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: fmt.Sprintf("Please generate appropriate SEO metadata for this blog content:\n\n%s", html),
		},
	}

	var finalMetadata *types.GenerateMetadataResponse
	maxIterations := 10

	for range make([]struct{}, maxIterations) {
		if finalMetadata != nil {
			break
		}

		resp, err := s.AIRepo.Client().CreateChatCompletion(
			ctx,
			openai.ChatCompletionRequest{
				Model:       "gpt-4.1-nano",
				Messages:    messages,
				Tools:       AITools,
				ToolChoice:  "auto",
				Temperature: 0.1,
			},
		)

		if err != nil {
			return nil, fmt.Errorf("OpenAI API error: %w", err)
		}
		if len(resp.Choices) == 0 {
			return nil, fmt.Errorf("empty response from OpenAI API")
		}

		assistantMessage := resp.Choices[0].Message
		messages = append(messages, assistantMessage)

		if assistantMessage.ToolCalls != nil && len(assistantMessage.ToolCalls) > 0 {
			for _, toolCall := range assistantMessage.ToolCalls {
				functionName := toolCall.Function.Name
				functionArgs := toolCall.Function.Arguments

				log.Printf("Calling function: %s", functionName)

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

	if finalMetadata == nil {
		return nil, fmt.Errorf("failed to generate metadata: maximum iterations reached")
	}

	return finalMetadata, nil
}
