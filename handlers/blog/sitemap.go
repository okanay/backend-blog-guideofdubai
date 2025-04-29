package BlogHandler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

func (h *Handler) SelectBlogSitemap(c *gin.Context) {
	// Cache'den sitemap verilerini kontrol et
	sitemapData, exists := h.BlogCache.GetSitemap()
	if exists {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"pages":   sitemapData,
			"cached":  true,
		})
		return
	}

	// Tüm yayınlanmış blog yazılarını al
	queryOptions := types.BlogCardQueryOptions{
		Status:        types.BlogStatusPublished,
		SortBy:        "updated_at",
		SortDirection: types.SortDesc,
		Limit:         99999, // Tüm blog yazılarını almak için yüksek limit
	}

	blogs, err := h.BlogRepository.SelectBlogCards(queryOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "sitemap_generation_failed",
			"message": "Sitemap verileri alınırken bir hata oluştu: " + err.Error(),
		})
		return
	}

	// Blog yazıları için sitemap girişleri
	var sitemapPages []gin.H

	for _, blog := range blogs {
		// Featured blog yazıları için daha yüksek öncelik
		priority := 0.8
		if blog.Featured {
			priority = 0.9
		}

		sitemapPages = append(sitemapPages, gin.H{
			"groupID":    blog.GroupID,
			"slug":       blog.Slug,
			"language":   blog.Language,
			"lastmod":    blog.UpdatedAt.Format(time.RFC3339),
			"priority":   priority,
			"changefreq": "weekly",
		})
	}

	// Cache'e sitemap verilerini kaydet
	h.BlogCache.SaveSitemap(sitemapPages)

	// JSON yanıtı döndür
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"pages":   sitemapPages,
		"cached":  false,
	})
}
