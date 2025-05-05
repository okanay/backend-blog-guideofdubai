// middlewares/blog_stats.go
package middlewares

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	BlogRepository "github.com/okanay/backend-blog-guideofdubai/repositories/blog"
	cache "github.com/okanay/backend-blog-guideofdubai/services"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

type BlogStatsMiddleware struct {
	blogRepo *BlogRepository.Repository
	cache    *cache.Cache
	duration time.Duration
}

func NewBlogStatsMiddleware(blogRepo *BlogRepository.Repository, cache *cache.Cache, duration time.Duration) *BlogStatsMiddleware {
	return &BlogStatsMiddleware{
		blogRepo: blogRepo,
		cache:    cache,
		duration: duration,
	}
}

func (m *BlogStatsMiddleware) TrackView() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Writer.Status() != 200 {
			return
		}

		blogIDInterface, exists := c.Get("blog_id")
		if !exists {
			return
		}

		blogID, ok := blogIDInterface.(uuid.UUID)
		if !ok {
			return
		}

		ip := utils.GetTrueClientIP(c)

		cacheKey := fmt.Sprintf("track_view::blog-id:%s:user-ip:%s", blogID.String(), ip)
		if _, exists := m.cache.Get(cacheKey); exists {
			fmt.Println("Cache hit", ip)
			return
		}

		go func(id uuid.UUID) {
			m.blogRepo.IncrementViewCount(id)
		}(blogID)

		m.cache.SetWithTTL(cacheKey, []byte(cacheKey), m.duration)
	}
}
