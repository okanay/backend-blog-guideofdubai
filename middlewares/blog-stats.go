// middlewares/blog_stats.go
package middlewares

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	BlogRepository "github.com/okanay/backend-blog-guideofdubai/repositories/blog"
	cache "github.com/okanay/backend-blog-guideofdubai/services"
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
		// Handler'ı önce çalıştır
		c.Next()

		// Sadece başarılı isteklerde devam et
		if c.Writer.Status() != 200 {
			return
		}

		// Blog ID'yi kontrol et
		blogIDInterface, exists := c.Get("blog_id")
		if !exists {
			return
		}

		blogID, ok := blogIDInterface.(uuid.UUID)
		if !ok {
			return
		}

		// Cache key oluştur
		cacheKey := fmt.Sprintf("blog_view:%s", blogID.String())
		// Son 15 dakika içinde görüntülendi mi?
		if _, exists := m.cache.Get(cacheKey); exists {
			return
		}

		// Görüntülenme sayısını artır (asenkron)
		go func(id uuid.UUID) {
			m.blogRepo.IncrementViewCount(id)
		}(blogID)

		// Cache'e kaydet (15 dakika)
		m.cache.SetWithTTL(cacheKey, []byte("1"), m.duration)
	}
}
