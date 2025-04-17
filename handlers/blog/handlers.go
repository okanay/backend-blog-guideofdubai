package BlogHandler

import (
	BlogRepository "github.com/okanay/backend-blog-guideofdubai/repositories/blog"
)

type Handler struct {
	BlogRepository *BlogRepository.Repository
}

func NewHandler(b *BlogRepository.Repository) *Handler {
	return &Handler{
		BlogRepository: b,
	}
}
