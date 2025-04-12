package middlewares

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-blog-guideofdubai/configs"
	UserRepository "github.com/okanay/backend-blog-guideofdubai/repositories/user"
)

func AuthMiddleware(ur *UserRepository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		refreshToken, err := c.Cookie(configs.SESSION_REFRESH_TOKEN_NAME)
		if err != nil {
			handleUnauthorized(c, "Session not found.")
			return
		}

		accessToken, err := c.Cookie(configs.SESSION_ACCESS_TOKEN_NAME)
		if err != nil {
			handleUnauthorized(c, "Session not found.")
			return
		}

		c.Set("refresh_token", refreshToken)
		c.Set("access_token", accessToken)
		c.Next()
	}
}

func handleUnauthorized(c *gin.Context, message string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(os.Getenv("SESSION_COOKIE_NAME"), "", -1, "/", os.Getenv("SESSION_DOMAIN"), false, false)

	c.JSON(http.StatusUnauthorized, gin.H{"error": message})
	c.Abort()
}
