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
		utils.Unauthorized(c, "Geçersiz kullanıcı adı veya şifre.")
		return
	}

	// Validate password
	if !utils.CheckPassword(request.Password, user.HashedPassword) {
		utils.Unauthorized(c, "Geçersiz kullanıcı adı veya şifre.")
		return
	}

	// Check user status
	if user.Status != types.UserStatusActive {
		var statusMessage string
		switch user.Status {
		case types.UserStatusSuspended:
			statusMessage = "Hesabınız askıya alındı."
		case types.UserStatusDeleted:
			statusMessage = "Hesabınız silindi."
		default:
			statusMessage = "Hesabınız aktif değil."
		}

		utils.Forbidden(c, statusMessage)
		return
	}

	// Token işlemleri...
	tokenClaims := types.TokenClaims{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}

	// Generate access token
	accessToken, err := utils.GenerateAccessToken(tokenClaims)
	if err != nil {
		utils.SendError(c, "token_generation_failed", "Oturum oluşturulurken bir hata oluştu.")
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
		if utils.HandleDatabaseError(c, err, "Token kaydetme") {
			return
		}
		return
	}

	// Update user's last login time
	now := time.Now()
	err = h.UserRepository.UpdateLastLogin(user.Email, now)
	if err != nil {
		// Bu hata kritik değil, session oluşmaya devam edebilir
		// Log yapılabilir
	}

	// Set cookies
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		configs.ACCESS_TOKEN_NAME,
		accessToken,
		int(configs.ACCESS_TOKEN_DURATION.Seconds()),
		"/",
		"",    // Domain
		false, // Secure
		true,  // HttpOnly
	)

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
	userProfile := types.UserView{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		Role:          user.Role,
		EmailVerified: user.EmailVerified,
		Status:        user.Status,
		CreatedAt:     user.CreatedAt,
		LastLogin:     now, // Newly updated login time
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Giriş başarılı.",
		"user":    userProfile,
	})
}
