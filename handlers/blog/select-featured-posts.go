package BlogHandler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

func (h *Handler) SelectFeaturedPosts(c *gin.Context) {
	// Cache key oluştur
	cacheKey := "featured_posts"

	// Cache'te var mı kontrol et
	if cachedData, exists := h.Cache.Get(cacheKey); exists {
		// Cache'ten veriyi JSON'a dönüştür
		var blogs []types.BlogPostCardView
		if err := json.Unmarshal(cachedData, &blogs); err == nil {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"blogs":   blogs,
				"count":   len(blogs),
				"cached":  true,
			})
			return
		}
	}

	// Cache'te yoksa veritabanından getir
	queryOptions := types.BlogCardQueryOptions{
		Featured:      true,
		Status:        types.BlogStatusPublished,
		Limit:         6,
		SortBy:        "created_at",
		SortDirection: types.SortDesc,
	}

	blogs, err := h.BlogRepository.SelectBlogCards(queryOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Veriyi JSON'a çevir ve cache'e kaydet
	if jsonData, err := json.Marshal(blogs); err == nil {
		h.Cache.Set(cacheKey, jsonData)
	}

	// Cevabı döndür
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"blogs":   blogs,
		"count":   len(blogs),
		"cached":  false,
	})
}
