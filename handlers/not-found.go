package handlers

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) NotFound(c *gin.Context) {
	c.JSON(404, gin.H{"message": "The requested " + c.Request.URL.Path + " was not found."})
}
