package AIHandler

import (
	AIRepository "github.com/okanay/backend-blog-guideofdubai/repositories/ai"
	BlogRepository "github.com/okanay/backend-blog-guideofdubai/repositories/blog"
	AIService "github.com/okanay/backend-blog-guideofdubai/services/ai"
)

type Handler struct {
	AIRepository   *AIRepository.Repository
	BlogRepository *BlogRepository.Repository
	AIService      *AIService.AIService
}

func NewHandler(ai *AIRepository.Repository, blog *BlogRepository.Repository, ais *AIService.AIService) *Handler {
	return &Handler{
		AIRepository:   ai,
		BlogRepository: blog,
		AIService:      ais,
	}
}
