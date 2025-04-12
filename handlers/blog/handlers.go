package BlogHandler

import BlogRepository "github.com/okanay/go-websocket-backend/repositories/blog"

type Handler struct {
	BlogRepository *BlogRepository.Repository
}

func NewHandler(B *BlogRepository.Repository) *Handler {
	return &Handler{
		BlogRepository: B,
	}
}
