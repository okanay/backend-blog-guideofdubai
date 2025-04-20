package BlogHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) SelectBlogCards(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"blogs":   "",
		"count":   0,
	})
}
