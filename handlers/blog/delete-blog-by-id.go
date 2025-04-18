package BlogHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) DeleteBlogByID(c *gin.Context) {
	blogIDString := c.Param("id")
	id, err := uuid.Parse(blogIDString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	err = h.BlogRepository.HardDeleteBlogByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
	})
}
