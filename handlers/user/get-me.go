package UserHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

func (h *Handler) GetMe(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	user, err := h.UserRepository.SelectByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "User not found.",
		})
		return
	}

	c.JSON(http.StatusOK, types.UserProfileResponse{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		Membership:    user.Membership,
		EmailVerified: user.EmailVerified,
		Status:        user.Status,
		CreatedAt:     user.CreatedAt,
		LastLogin:     user.LastLogin,
	})
}
