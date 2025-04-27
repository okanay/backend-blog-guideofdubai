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

	// Parametrelerin varlığını kontrol et
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "missing_parameter",
			"message": "slug parametresi gereklidir.",
		})
		return
	}

	if lang == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "missing_parameter",
			"message": "lang parametresi gereklidir.",
		})
		return
	}

	// Cache'den blogu kontrol et
	blog, exists := h.BlogCache.GetBlogByGroupIDAndLang(slug, lang)
	if exists {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"blog":    blog,
			"cached":  true,
		})
		return
	}

	// Request nesnesini oluştur
	request := types.BlogSelectByGroupIDInput{
		GroupID:  slug,
		Language: lang,
	}

	// Repository'den blog bilgilerini getir
	blog, err := h.BlogRepository.SelectBlogByGroupID(request)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "blog_not_found",
			"message": err.Error(),
		})
		return
	}

	// Blog'u cache'e kaydet
	h.BlogCache.SaveBlogByGroupIDAndLang(slug, lang, blog)

	// Başarılı yanıt döndür
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"blog":    blog,
		"cached":  false,
	})
}
