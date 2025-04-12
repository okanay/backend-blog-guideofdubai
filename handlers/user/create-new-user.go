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

func (h *Handler) CreateNewUser(c *gin.Context) {
	var request types.UserCreateRequest

	err := utils.ValidateRequest(c, &request)
	if err != nil {
		return
	}

	user, err := h.UserRepository.CreateNewUser(request)
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
						"message": "Bu kullanıcı adı zaten kullanılıyor.",
					})
				} else if strings.Contains(constraintName, "email") {
					c.JSON(http.StatusConflict, gin.H{
						"success": false,
						"error":   "email_exists",
						"message": "Bu e-posta adresi zaten kullanılıyor.",
					})
				} else {
					c.JSON(http.StatusConflict, gin.H{
						"success": false,
						"error":   "duplicate_entry",
						"message": "Bu kayıt zaten mevcut.",
					})
				}
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "server_error",
			"message": "Kullanıcı oluşturulurken bir hata oluştu.",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Kullanıcı başarıyla oluşturuldu.",
		"user":    user,
	})
}
