package AIHandler

import (
	AIRepository "github.com/okanay/backend-blog-guideofdubai/repositories/ai"
)

type Handler struct {
	AIRepository *AIRepository.Repository
}

func NewHandler(ai *AIRepository.Repository) *Handler {
	return &Handler{
		AIRepository: ai,
	}
}
