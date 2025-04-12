package UserHandler

import UserRepository "github.com/okanay/go-websocket-backend/repositories/user"

type Handler struct {
	UserRepository *UserRepository.Repository
}

func NewHandler(u *UserRepository.Repository) *Handler {
	return &Handler{
		UserRepository: u,
	}
}
