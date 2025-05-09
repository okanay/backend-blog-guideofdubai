package AIService

import (
	AIRepository "github.com/okanay/backend-blog-guideofdubai/repositories/ai"
	BlogRepository "github.com/okanay/backend-blog-guideofdubai/repositories/blog"
	"github.com/sashabaranov/go-openai"
)

// AIService handles AI operations between repository and handlers
type AIService struct {
	AIRepo   *AIRepository.Repository
	BlogRepo *BlogRepository.Repository
	Tools    []openai.Tool
}

func NewAIService(aiRepo *AIRepository.Repository, blogRepo *BlogRepository.Repository) *AIService {
	return &AIService{
		AIRepo:   aiRepo,
		BlogRepo: blogRepo,
		Tools:    AITools,
	}
}
