package AIHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

// GenerateBlogMetadata bir blog yazısı için metadata oluşturur
func (h *Handler) GenerateBlogMetadata(c *gin.Context) {
	var request types.GenerateMetadataRequest
	err := utils.ValidateRequest(c, &request)
	if err != nil {
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	// AI Service'i kullanarak metadata oluştur
	metadata, tokensUsed, err := h.AIService.GenerateMetadataWithTools(
		c.Request.Context(),
		request.HTML,
		request.Language,
		userID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "metadata_generation_failed",
			"message": "Metadata oluşturulurken bir hata oluştu: " + err.Error(),
		})
		return
	}

	// Token kullanımını context'e kaydet (rate limiter için)
	c.Set("tokens_used", tokensUsed)

	// Global utils fonksiyonunu kullanarak maliyet bilgisini al
	costInfo := utils.CalculateAICost(tokensUsed)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metadata,
		"cost":    costInfo,
	})
}
