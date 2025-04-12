package middlewares

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-blog-guideofdubai/configs"
	TokenRepository "github.com/okanay/backend-blog-guideofdubai/repositories/token"
	UserRepository "github.com/okanay/backend-blog-guideofdubai/repositories/user"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func AuthMiddleware(ur *UserRepository.Repository, tr *TokenRepository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Check access token
		accessToken, err := c.Cookie(configs.SESSION_ACCESS_TOKEN_NAME)
		if err != nil {
			// If there is no access token, check the refresh token
			handleTokenRenewal(c, ur, tr)
			return
		}

		// 2. Validate the access token
		claims, err := utils.ValidateAccessToken(accessToken)
		if err != nil {
			// If the access token is invalid or expired, check the refresh token
			expired, _ := utils.IsTokenExpired(accessToken)
			if expired {
				// If the token is expired, attempt renewal
				handleTokenRenewal(c, ur, tr)
				return
			}

			// If the token is invalid and it's not an expiration issue, terminate the session
			handleUnauthorized(c, "Invalid session.")
			return
		}

		// 3. Token is valid, add user information to the context
		c.Set("user_id", claims.UniqueID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		// 4. Continue processing
		c.Next()
	}
}

func handleTokenRenewal(c *gin.Context, ur *UserRepository.Repository, tr *TokenRepository.Repository) {
	// 1. Retrieve the refresh token
	refreshToken, err := c.Cookie(configs.SESSION_REFRESH_TOKEN_NAME)
	if err != nil {
		handleUnauthorized(c, "Oturum bulunamadı.")
		return
	}

	// 2. Check the refresh token in the database
	dbToken, err := tr.SelectRefreshTokenByToken(refreshToken)
	if err != nil {
		handleUnauthorized(c, "Oturum geçersiz.")
		return
	}

	// 3. Validate the refresh token
	if dbToken.IsRevoked {
		handleUnauthorized(c, "Oturum iptal edilmiş.")
		return
	}

	if dbToken.ExpiresAt.Before(time.Now()) {
		handleUnauthorized(c, "Oturum süresi dolmuş.")
		return
	}

	// 4. Retrieve the user from the database
	user, err := ur.SelectUserByUsername(dbToken.UserUsername)
	if err != nil {
		handleUnauthorized(c, "Kullanıcı bulunamadı.")
		return
	}

	// 5. Check the user's status
	if user.Status != types.UserStatusActive {
		handleUnauthorized(c, "Hesabınız aktif değil.")
		return
	}

	// 6. Create token claims
	tokenClaims := types.TokenClaims{
		UniqueID: user.UniqueID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Membership,
	}

	// 7. Generate a new access token
	newAccessToken, err := utils.GenerateAccessToken(tokenClaims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "token_generation_failed",
			"message": "An error occurred while renewing the session.",
		})
		c.Abort()
		return
	}

	// 8. Update the last used time of the refresh token
	err = tr.UpdateRefreshTokenLastUsed(refreshToken)
	if err != nil {
		// Logging can be done but it won't block the process
	}

	// 9. Set the new access token cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		configs.SESSION_ACCESS_TOKEN_NAME,
		newAccessToken,
		int(configs.JWT_ACCESS_TOKEN_EXPIRATION.Seconds()),
		"/",
		"",
		false,
		true,
	)

	// 10. Add user information to the context
	c.Set("user_id", user.UniqueID)
	c.Set("username", user.Username)
	c.Set("email", user.Email)
	c.Set("role", user.Membership)

	// 11. Continue processing
	c.Next()
}

func handleUnauthorized(c *gin.Context, message string) {
	// Clear cookies
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(configs.SESSION_ACCESS_TOKEN_NAME, "", -1, "/", "", false, true)
	c.SetCookie(configs.SESSION_REFRESH_TOKEN_NAME, "", -1, "/", "", false, true)

	// Return error
	c.JSON(http.StatusUnauthorized, gin.H{
		"success": false,
		"error":   "unauthorized",
		"message": message,
	})
	c.Abort()
}
