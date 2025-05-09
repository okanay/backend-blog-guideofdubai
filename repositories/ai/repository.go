package AIRepository

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type Repository struct {
	client *openai.Client
}

func NewRepository(apiKey string) *Repository {
	client := openai.NewClient(apiKey)
	return &Repository{
		client: client,
	}
}

// Client returns the OpenAI client instance
func (r *Repository) Client() *openai.Client {
	return r.client
}

// CreateChatCompletion encapsulates the OpenAI API call
func (r *Repository) CreateChatCompletion(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return r.client.CreateChatCompletion(ctx, request)
}
