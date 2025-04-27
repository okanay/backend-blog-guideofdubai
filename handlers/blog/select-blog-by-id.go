package BlogHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) SelectBlogByID(c *gin.Context) {
	blogIDString := c.Param("id")

	// UUID'yi doğrula
	blogID, err := uuid.Parse(blogIDString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid_id",
			"message": "Geçersiz blog ID formatı.",
		})
		return
	}

	// Cache'den blogu kontrol et
	blog, exists := h.BlogCache.GetBlogByID(blogID)
	if exists {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"blog":    blog,
			"cached":  true,
		})
		return
	}

	// Cache'de yoksa veritabanından getir
	blog, err = h.BlogRepository.SelectBlogByID(blogID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "blog_not_found",
			"message": "Blog yazısı bulunamadı.",
		})
		return
	}

	// Blog'u cache'e kaydet
	h.BlogCache.SaveBlogByID(blogID, blog)

	// Başarılı yanıt döndür
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"blog":    blog,
		"cached":  false,
	})
}
