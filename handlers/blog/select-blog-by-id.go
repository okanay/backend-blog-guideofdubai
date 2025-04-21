package BlogHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) SelectBlogByID(c *gin.Context) {
	// URL'den id parametresini al
	blogIDString := c.Param("id")

	// ID'yi UUID formatına dönüştür
	id, err := uuid.Parse(blogIDString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid_id",
			"message": "Geçersiz blog ID formatı.",
		})
		return
	}

	// Repository'den blog bilgilerini getir
	blog, err := h.BlogRepository.SelectBlogByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "blog_not_found",
			"message": "Blog yazısı bulunamadı.",
		})
		return
	}

	// Başarılı yanıt döndür
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"blog":    blog,
	})
}
