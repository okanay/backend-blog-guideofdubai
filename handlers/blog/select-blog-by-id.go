package BlogHandler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

func (h *Handler) SelectBlogByID(c *gin.Context) {
	blogIDString := c.Param("id")

	// Cache key oluştur
	cacheKey := "blog_id:" + blogIDString

	// Cache'te var mı kontrol et
	if cachedData, exists := h.Cache.Get(cacheKey); exists {
		// Cache'ten veriyi JSON'a dönüştür
		var blog types.BlogPostView
		if err := json.Unmarshal(cachedData, &blog); err == nil {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"blog":    blog,
				"cached":  true,
			})
			return
		}
	}

	// Cache'te yoksa veritabanından getir
	blogID, err := uuid.Parse(blogIDString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid_id",
			"message": "Geçersiz blog ID formatı.",
		})
		return
	}

	blog, err := h.BlogRepository.SelectBlogByID(blogID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "blog_not_found",
			"message": "Blog yazısı bulunamadı.",
		})
		return
	}

	// Veriyi JSON'a çevir ve cache'e kaydet
	if jsonData, err := json.Marshal(blog); err == nil {
		h.Cache.Set(cacheKey, jsonData)
	}

	// Başarılı yanıt döndür
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"blog":    blog,
		"cached":  false,
	})
}
