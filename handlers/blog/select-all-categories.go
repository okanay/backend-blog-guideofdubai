package BlogHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) SelectAllCategories(c *gin.Context) {
	categories, err := h.BlogRepository.SelectAllCategories()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":    true,
		"categories": categories,
	})
}
