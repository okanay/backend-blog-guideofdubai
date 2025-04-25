// handlers/blog/handlers.go
package BlogHandler

import (
	BlogRepository "github.com/okanay/backend-blog-guideofdubai/repositories/blog"
	cache "github.com/okanay/backend-blog-guideofdubai/services"
)

type Handler struct {
	BlogRepository *BlogRepository.Repository
	Cache          *cache.Cache
}

func NewHandler(b *BlogRepository.Repository, c *cache.Cache) *Handler {
	return &Handler{
		BlogRepository: b,
		Cache:          c,
	}
}
