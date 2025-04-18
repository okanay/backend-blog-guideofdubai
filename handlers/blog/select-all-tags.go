package BlogHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) SelectAllTags(c *gin.Context) {
	tags, err := h.BlogRepository.SelectAllTags()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"tags":    tags,
	})
}
