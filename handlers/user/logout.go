package UserHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-blog-guideofdubai/configs"
)

func (h *Handler) Logout(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		configs.ACCESS_TOKEN_NAME,
		"",
		-1,
		"/",
		"",    // Domain - can be left empty, browser will use the current domain
		false, // Secure - should be true in production
		true,  // HttpOnly
	)

	// Refresh Token Cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		configs.REFRESH_TOKEN_NAME,
		"",
		-1,
		"/",
		"",    // Domain
		false, // Secure
		true,  // HttpOnly
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logout successful.",
	})
}
