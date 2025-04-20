package BlogHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (h *Handler) CreateBlogTag(c *gin.Context) {
	var request types.TagInput

	err := utils.ValidateRequest(c, &request)
	if err != nil {
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	tag, err := h.BlogRepository.CreateBlogTag(request, userID)

	if err != nil {
		// Aynı hata işleme fonksiyonu tekrar kullanılır
		if utils.HandleDatabaseError(c, err, "Blog etiketi oluşturma") {
			return
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"tag":     tag,
	})
}
