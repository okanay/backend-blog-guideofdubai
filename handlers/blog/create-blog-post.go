package BlogHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (h *Handler) CreateBlogPost(c *gin.Context) {
	var request types.BlogPostCreateInput

	err := utils.ValidateRequest(c, &request)
	if err != nil {
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	blog, err := h.BlogRepository.CreateBlogPost(request, userID)

	if err != nil {
		if utils.HandleDatabaseError(c, err, "Blog yazısı oluşturma") {
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "unexpected_error",
			"message": "Beklenmeyen bir hata oluştu.",
		})
		return
	}

	h.Cache.Clear()

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"blog":    blog,
	})
}
