package BlogHandler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) SelectRelatedPosts(c *gin.Context) {
	// Parametreleri al
	blogID := c.Query("blogId")
	language := c.Query("language")
	limit := 4 // Varsayılan

	if limitParam, err := strconv.Atoi(c.Query("limit")); err == nil && limitParam > 0 {
		limit = limitParam
	}

	// Kategori ve etiketleri parse et
	categoriesParam := c.Query("categories")
	tagsParam := c.Query("tags")

	var categories, tags []string
	if categoriesParam != "" {
		categories = strings.Split(categoriesParam, ",")
	}
	if tagsParam != "" {
		tags = strings.Split(tagsParam, ",")
	}

	// Blog ID'yi doğrula
	blogUUID, err := uuid.Parse(blogID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid_blog_id",
			"message": "Geçersiz blog ID formatı",
		})
		return
	}

	// Cache kontrolü
	relatedPosts, exists := h.BlogCache.GetRelatedPosts(blogUUID, categories, tags, language)
	if exists {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"blogs":   relatedPosts,
			"cached":  true,
		})
		return
	}

	// İlgili içerikleri getir
	relatedPosts, err = h.BlogRepository.SelectRelatedPosts(blogUUID, categories, tags, language, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "database_error",
			"message": "İlgili içerikler getirilemedi: " + err.Error(),
		})
		return
	}

	// Cache'e kaydet
	h.BlogCache.SaveRelatedPosts(blogUUID, categories, tags, language, relatedPosts)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"blogs":   relatedPosts,
		"cached":  false,
	})
}
