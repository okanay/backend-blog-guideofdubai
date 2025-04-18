package BlogHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (h *Handler) SelectBlogByGroupID(c *gin.Context) {
	var request types.BlogSelectByGroupIDInput

	err := utils.ValidateRequest(c, &request)
	if err != nil {
		return
	}

	blogView, err := h.BlogRepository.SelectBlogByGroupID(request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"blog":    blogView,
	})
}
