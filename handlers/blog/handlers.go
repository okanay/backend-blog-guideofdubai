// handlers/blog/handlers.go
package BlogHandler

import (
	BlogRepository "github.com/okanay/backend-blog-guideofdubai/repositories/blog"
	"github.com/okanay/backend-blog-guideofdubai/services/cache"
)

type Handler struct {
	BlogRepository *BlogRepository.Repository
	Cache          *cache.Cache
	BlogCache      *cache.BlogCacheService
}

func NewHandler(b *BlogRepository.Repository, c *cache.Cache) *Handler {
	return &Handler{
		BlogRepository: b,
		Cache:          c,
		BlogCache:      cache.NewBlogCacheService(c),
	}
}
