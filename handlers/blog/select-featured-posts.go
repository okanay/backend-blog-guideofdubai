package BlogHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

func (h *Handler) SelectFeaturedPosts(c *gin.Context) {
	// Cache'den öne çıkan yazıları kontrol et
	blogs, exists := h.BlogCache.GetFeaturedPosts()
	if exists {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"blogs":   blogs,
			"count":   len(blogs),
			"cached":  true,
		})
		return
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

	// Öne çıkan yazıları cache'e kaydet
	h.BlogCache.SaveFeaturedPosts(blogs)

	// Cevabı döndür
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"blogs":   blogs,
		"count":   len(blogs),
		"cached":  false,
	})
}
