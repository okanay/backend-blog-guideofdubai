package AIHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (h *Handler) TranslateBlogPostJSON(c *gin.Context) {
	var request types.TranslateRequest
	err := utils.ValidateRequest(c, &request)
	if err != nil {
		return
	}

	translatedJSON, inputTokens, outputTokens, err := h.AIService.TranslateBlogPostJSON(
		c.Request.Context(),
		request.TiptapJSON,
		request.SourceLanguage,
		request.TargetLanguage,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "translation_failed",
			"message": "Çeviri sırasında bir hata oluştu: " + err.Error(),
		})
		return
	}

	cost := utils.CalculateAICostWithOutput(inputTokens, outputTokens)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"translatedJSON": translatedJSON,
			"tokensUsed":     inputTokens + outputTokens,
			"cost":           cost,
		},
	})
}
