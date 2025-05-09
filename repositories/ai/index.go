package AIRepository

import (
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
