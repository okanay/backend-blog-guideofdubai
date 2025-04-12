package handlers

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) Index(c *gin.Context) {

	c.JSON(200, gin.H{
		"Project":   "Guide Of Dubai - Blog Backend",
		"Language":  "Golang",
		"Framework": "Gin",
		"Database":  "PostgreSQL",
		"Status":    "System is running successfully.",
	})
}
