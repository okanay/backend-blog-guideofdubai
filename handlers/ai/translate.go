package AIHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

// TranslateBlogPost bir blog yazısının içeriğini çevirir
func (h *Handler) TranslateBlogPost(c *gin.Context) {
	var request types.TranslateRequest
	err := utils.ValidateRequest(c, &request)
	if err != nil {
		return
	}

	// HTML içeriğini çevir
	translatedHTML, err := h.AIRepository.TranslateHTML(
		c.Request.Context(),
		request.HTML,
		request.SourceLanguage,
		request.TargetLanguage,
		3000, // Yaklaşık token sayısı (ayarlanabilir)
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "translation_failed",
			"message": "Çeviri sırasında bir hata oluştu: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"translatedHTML": translatedHTML,
			"sourceLanguage": request.SourceLanguage,
			"targetLanguage": request.TargetLanguage,
		},
	})
}
