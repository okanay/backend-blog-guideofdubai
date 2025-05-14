package BlogHandler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

func (h *Handler) SelectRecentPosts(c *gin.Context) {
	// Cache'den son eklenen yazıları kontrol et
	blogs, exists := h.BlogCache.GetRecentPosts()
	if exists {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"blogs":   blogs,
			"count":   len(blogs),
			"cached":  true,
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "4")
	parsedLimit, _ := strconv.Atoi(limitStr)

	language := c.DefaultQuery("language", "en")

	// Cache'te yoksa veritabanından getir
	queryOptions := types.BlogCardQueryOptions{
		Status:        types.BlogStatusPublished,
		Limit:         parsedLimit,
		SortBy:        "created_at",
		SortDirection: types.SortDesc,
		Language:      language,
	}
	blogs, _, err := h.BlogRepository.SelectBlogCards(queryOptions)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Son eklenen yazıları cache'e kaydet
	h.BlogCache.SaveRecentPosts(blogs)

	// Cevabı döndür
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"blogs":   blogs,
		"count":   len(blogs),
		"cached":  false,
	})
}
