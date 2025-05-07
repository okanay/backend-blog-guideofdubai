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
	translatedHTML, tokensUsed, err := h.AIRepository.TranslateHTML(
		c.Request.Context(),
		request.HTML,
		request.SourceLanguage,
		request.TargetLanguage,
		750, // Yaklaşık token sayısı (ayarlanabilir)
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "translation_failed",
			"message": "Çeviri sırasında bir hata oluştu: " + err.Error(),
		})
		return
	}

	// Token kullanımını context'e kaydet (rate limiter için)
	c.Set("tokens_used", tokensUsed)

	// İşlem maliyeti bilgisini ekleyelim
	costInfo := calculateCost(tokensUsed)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"translatedHTML": translatedHTML,
			"sourceLanguage": request.SourceLanguage,
			"targetLanguage": request.TargetLanguage,
			"tokensUsed":     tokensUsed,
			"cost":           costInfo,
		},
	})
}

// Maliyet hesaplama yardımcı fonksiyonu
func calculateCost(tokensUsed int) map[string]any {
	// Fiyatlandırma: Input $0.05, Output $0.20 (milyon token başına)
	inputCost := float64(tokensUsed) * 0.05 / 1000000.0
	outputCost := float64(tokensUsed) * 0.20 / 1000000.0
	totalCost := inputCost + outputCost

	return map[string]any{
		"inputTokens":  tokensUsed,
		"outputTokens": tokensUsed, // Yaklaşık bir değer, gerçek durumda farklı olabilir
		"inputCost":    inputCost,
		"outputCost":   outputCost,
		"totalCost":    totalCost,
	}
}
