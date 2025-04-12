package BlogHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) BlogPageIndex(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"page": "blog-page-index"})
}
