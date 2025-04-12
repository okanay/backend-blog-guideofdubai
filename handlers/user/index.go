package UserHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) UserPageIndex(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"page": "user-page-index"})
}
