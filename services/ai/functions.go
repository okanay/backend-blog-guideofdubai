package AIService

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	types "github.com/okanay/backend-blog-guideofdubai/types"
)

func (s *AIService) DispatchToolCall(
	ctx context.Context,
	functionName string,
	functionArgs string,
	userID uuid.UUID,
) (string, *types.GenerateMetadataResponse, error) {
	switch functionName {
	case "get_all_categories":
		result, err := s.HandleGetAllCategories(ctx)
		return result, nil, err
	case "get_all_tags":
		result, err := s.HandleGetAllTags(ctx)
		return result, nil, err
	case "create_category":
		result, err := s.HandleCreateCategory(ctx, functionArgs, userID)
		return result, nil, err
	case "create_tag":
		result, err := s.HandleCreateTag(ctx, functionArgs, userID)
		return result, nil, err
	case "finalize_metadata":
		metadata, result, err := s.HandleFinalizeMetadata(functionArgs)
		return result, metadata, err
	default:
		return `{"error": "Unknown function"}`, nil, nil
	}
}

func (s *AIService) HandleGetAllCategories(ctx context.Context) (string, error) {
	result, err := s.BlogRepo.SelectAllCategories()
	if err != nil {
		return "", err
	}
	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

func (s *AIService) HandleGetAllTags(ctx context.Context) (string, error) {
	result, err := s.BlogRepo.SelectAllTags()
	if err != nil {
		return "", err
	}
	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

func (s *AIService) HandleCreateCategory(ctx context.Context, args string, userID uuid.UUID) (string, error) {
	var params map[string]string
	json.Unmarshal([]byte(args), &params)
	newInput := types.CategoryInput{
		Name:  params["name"],
		Value: params["value"],
	}
	result, err := s.BlogRepo.CreateBlogCategory(newInput, userID)
	if err != nil {
		return "", err
	}
	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

func (s *AIService) HandleCreateTag(ctx context.Context, args string, userID uuid.UUID) (string, error) {
	var params map[string]string
	json.Unmarshal([]byte(args), &params)
	newInput := types.TagInput{
		Name:  params["name"],
		Value: params["value"],
	}
	result, err := s.BlogRepo.CreateBlogTag(newInput, userID)
	if err != nil {
		return "", err
	}
	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

func (s *AIService) HandleFinalizeMetadata(args string) (*types.GenerateMetadataResponse, string, error) {
	var metadata types.GenerateMetadataResponse
	if err := json.Unmarshal([]byte(args), &metadata); err != nil {
		return nil, "", err
	}
	return &metadata, `{"status": "success", "message": "Metadata finalized successfully"}`, nil
}
