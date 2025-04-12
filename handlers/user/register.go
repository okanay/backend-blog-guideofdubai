package UserHandler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (h *Handler) Register(c *gin.Context) {
	var request types.UserCreateRequest

	err := utils.ValidateRequest(c, &request)
	if err != nil {
		return
	}

	user, err := h.UserRepository.CreateUser(request)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505":
				constraintName := pqErr.Constraint
				if strings.Contains(constraintName, "username") {
					c.JSON(http.StatusConflict, gin.H{
						"success": false,
						"error":   "username_exists",
						"message": "This username is already in use.",
					})
				} else if strings.Contains(constraintName, "email") {
					c.JSON(http.StatusConflict, gin.H{
						"success": false,
						"error":   "email_exists",
						"message": "This email address is already in use.",
					})
				} else {
					c.JSON(http.StatusConflict, gin.H{
						"success": false,
						"error":   "duplicate_entry",
						"message": "This entry already exists.",
					})
				}
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "server_error",
			"message": "An error occurred while creating the user.",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User created successfully.",
		"user":    user,
	})
}
