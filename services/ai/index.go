// services/ai_service.go

package AIService

import (
	AIRepository "github.com/okanay/backend-blog-guideofdubai/repositories/ai"
	BlogRepository "github.com/okanay/backend-blog-guideofdubai/repositories/blog"
)

type AIService struct {
	AIRepo   *AIRepository.Repository
	BlogRepo *BlogRepository.Repository
}

func NewAIService(aiRepo *AIRepository.Repository, blogRepo *BlogRepository.Repository) *AIService {
	return &AIService{
		AIRepo:   aiRepo,
		BlogRepo: blogRepo,
	}
}

// Tool tanımları (functions) için yapı
type Function struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}
