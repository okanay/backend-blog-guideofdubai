package middlewares

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/configs"
	TokenRepository "github.com/okanay/backend-blog-guideofdubai/repositories/token"
	UserRepository "github.com/okanay/backend-blog-guideofdubai/repositories/user"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func AuthMiddleware(ur *UserRepository.Repository, tr *TokenRepository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Check access token
		accessToken, err := c.Cookie(configs.ACCESS_TOKEN_NAME)
		if err != nil {
			// If there is no access token, check the refresh token
			handleTokenRenewal(c, ur, tr)
			return
		}

		claims, err := utils.ValidateAccessToken(accessToken)
		if err != nil {
			// Hatanın türünü kontrol et
			if errors.Is(err, jwt.ErrTokenExpired) {
				handleTokenRenewal(c, ur, tr)
				return
			}

			handleUnauthorized(c, "Invalid session token.") // Daha spesifik mesaj
			return
		}

		setContextValues(c, claims.ID, claims.Username, claims.Email, claims.Role, claims.EmailVerified, claims.Status, claims.CreatedAt, claims.LastLogin)

		// 4. Continue processing
		c.Next()
	}
}

func handleTokenRenewal(c *gin.Context, ur *UserRepository.Repository, tr *TokenRepository.Repository) {
	defer utils.TimeTrack(time.Now(), "Token -> Renewal User Token")

	// 1. Retrieve the refresh token
	refreshToken, err := c.Cookie(configs.REFRESH_TOKEN_NAME)
	if err != nil {
		handleUnauthorized(c, "Session not found.")
		return
	}

	// 2. Check the refresh token in the database
	dbToken, err := tr.SelectRefreshTokenByToken(refreshToken)
	if err != nil {
		handleUnauthorized(c, "Invalid session.")
		return
	}

	// 3. Validate the refresh token
	if dbToken.IsRevoked {
		handleUnauthorized(c, "Session has been revoked.")
		return
	}

	if dbToken.ExpiresAt.Before(time.Now()) {
		handleUnauthorized(c, "Session has expired.")
		return
	}

	// 4. Retrieve the user from the database
	user, err := ur.SelectByUsername(dbToken.UserUsername)
	if err != nil {
		handleUnauthorized(c, "User not found.")
		return
	}

	// 5. Check the user's status
	if user.Status != types.UserStatusActive {
		handleUnauthorized(c, "Your account is not active.")
		return
	}

	// 6. Create token claims
	tokenClaims := types.TokenClaims{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
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
		configs.ACCESS_TOKEN_NAME,
		newAccessToken,
		int(configs.ACCESS_TOKEN_DURATION.Seconds()),
		"/",
		"",
		false,
		true,
	)

	// 10. Add user information to the context
	setContextValues(c, user.ID, user.Username, user.Email, user.Role, user.EmailVerified, user.Status, user.CreatedAt, user.LastLogin)
	// 11. Continue processing
	c.Next()
}

func handleUnauthorized(c *gin.Context, message string) {
	// Clear cookies
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(configs.ACCESS_TOKEN_NAME, "", -1, "/", "", false, true)
	c.SetCookie(configs.REFRESH_TOKEN_NAME, "", -1, "/", "", false, true)

	// Return error
	c.JSON(http.StatusUnauthorized, gin.H{
		"success": false,
		"error":   "unauthorized",
		"message": message,
	})
	c.Abort()
}

func setContextValues(c *gin.Context, userID uuid.UUID, username string, email string, role types.Role, emailVerified bool, status types.UserStatus, createdAt time.Time, lastLogin time.Time) {
	c.Set("user_id", userID)
	c.Set("username", username)
	c.Set("email", email)
	c.Set("role", role)
	c.Set("email_verified", emailVerified)
	c.Set("status", status)
	c.Set("created_at", createdAt)
	c.Set("last_login", lastLogin)
}
