package ImageHandler

import (
	ImageRepository "github.com/okanay/backend-blog-guideofdubai/repositories/image"
	StorageRepository "github.com/okanay/backend-blog-guideofdubai/repositories/storage"
)

type Handler struct {
	ImageRepository   *ImageRepository.Repository
	StorageRepository *StorageRepository.Repository
}

func NewHandler(i *ImageRepository.Repository, s *StorageRepository.Repository) *Handler {
	return &Handler{
		ImageRepository:   i,
		StorageRepository: s,
	}
}
