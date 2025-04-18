package BlogHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (h *Handler) CreateBlogCategory(c *gin.Context) {
	var request types.CategoryInput

	err := utils.ValidateRequest(c, &request)
	if err != nil {
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	category, err := h.BlogRepository.CreateBlogCategory(request, userID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":  true,
		"category": category,
	})
}
