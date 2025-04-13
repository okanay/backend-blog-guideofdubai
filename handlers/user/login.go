package UserHandler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-blog-guideofdubai/configs"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (h *Handler) Login(c *gin.Context) {
	var request types.UserLoginRequest

	// Validate request
	err := utils.ValidateRequest(c, &request)
	if err != nil {
		return
	}

	// Retrieve user information from the database
	user, err := h.UserRepository.SelectByUsername(request.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "invalid_credentials",
			"message": "Invalid username or password.",
		})
		return
	}

	// Validate password
	if !utils.CheckPassword(request.Password, user.HashedPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "invalid_credentials",
			"message": "Invalid username or password.",
		})
		return
	}

	// Check user status
	if user.Status != types.UserStatusActive {
		var statusMessage string
		switch user.Status {
		case types.UserStatusSuspended:
			statusMessage = "Your account has been suspended."
		case types.UserStatusDeleted:
			statusMessage = "Your account has been deleted."
		default:
			statusMessage = "Your account is not active."
		}

		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "account_inactive",
			"message": statusMessage,
		})
		return
	}

	// Create token claims
	tokenClaims := types.TokenClaims{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Membership,
	}

	// Generate access token
	accessToken, err := utils.GenerateAccessToken(tokenClaims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "token_generation_failed",
			"message": "An error occurred while creating the session.",
		})
		return
	}

	// Generate refresh token
	refreshToken := utils.GenerateRefreshToken()

	// Set expiration date for refresh token
	expiresAt := time.Now().Add(configs.REFRESH_TOKEN_DURATION)

	// Save refresh token to the database
	tokenRequest := types.TokenCreateRequest{
		UserID:       user.ID,
		UserEmail:    user.Email,
		UserUsername: user.Username,
		Token:        refreshToken,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
		ExpiresAt:    expiresAt,
	}

	_, err = h.TokenRepository.CreateRefreshToken(tokenRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "token_save_failed",
			"message": "An error occurred while creating the session.",
		})
		return
	}

	// Update user's last login time
	now := time.Now()
	err = h.UserRepository.UpdateLastLogin(user.Email, now)
	if err != nil {
		// This error should not prevent session creation, just log it
		// log.Printf("Error updating last login time: %v", err)
	}

	// Set cookies
	// Access Token Cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		configs.ACCESS_TOKEN_NAME,
		accessToken,
		int(configs.ACCESS_TOKEN_DURATION.Seconds()),
		"/",
		"",    // Domain - can be left empty, browser will use the current domain
		false, // Secure - should be true in production
		true,  // HttpOnly
	)

	// Refresh Token Cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		configs.REFRESH_TOKEN_NAME,
		refreshToken,
		int(configs.REFRESH_TOKEN_DURATION.Seconds()),
		"/",
		"",    // Domain
		false, // Secure
		true,  // HttpOnly
	)

	// Return user information securely
	userProfile := types.UserProfileResponse{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		Membership:    user.Membership,
		EmailVerified: user.EmailVerified,
		Status:        user.Status,
		CreatedAt:     user.CreatedAt,
		LastLogin:     now, // Newly updated login time
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful.",
		"user":    userProfile,
	})
}
