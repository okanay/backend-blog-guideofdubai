package BlogHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

func (h *Handler) SelectBlogByGroupID(c *gin.Context) {
	// Query parametrelerini al
	slug := c.Query("slug")
	lang := c.Query("lang")

	// GroupID parametresinin varlığını kontrol et
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "missing_parameter",
			"message": "groupId parametresi gereklidir.",
		})
		return
	}

	// Language parametresinin varlığını kontrol et
	if lang == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "missing_parameter",
			"message": "lang parametresi gereklidir.",
		})
		return
	}

	// Request nesnesini oluştur
	request := types.BlogSelectByGroupIDInput{
		GroupID:  slug,
		Language: lang,
	}

	// Repository'den blog bilgilerini getir
	blogView, err := h.BlogRepository.SelectBlogByGroupID(request)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "blog_not_found",
			"message": err.Error(),
		})
		return
	}

	// Başarılı yanıt döndür (HTTP 200 OK kullanılıyor - HTTP 201 Created yerine)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"blog":    blogView,
	})
}
