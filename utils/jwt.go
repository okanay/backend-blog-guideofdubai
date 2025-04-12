package utils

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/okanay/backend-blog-guideofdubai/configs"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

// JWTClaims, standart JWT claims'i genişleterek özelleştirilmiş claim yapısı oluşturur
type JWTClaims struct {
	jwt.RegisteredClaims
	types.TokenClaims
}

// GenerateAccessToken, kullanıcı bilgileriyle yeni bir access token oluşturur
func GenerateAccessToken(claims types.TokenClaims) (string, error) {
	// JWT için secret key'i çevresel değişkenlerden al
	secretKey := os.Getenv("JWT_ACCESS_SECRET")
	if secretKey == "" {
		return "", errors.New("JWT_ACCESS_SECRET environment variable is not set")
	}

	expiryMinutes := configs.ACCESS_TOKEN_DURATION

	// JWT claims yapısını oluştur
	tokenClaims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiryMinutes) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    configs.JWT_ISSUER,
			Subject:   claims.Email,
		},
		TokenClaims: claims,
	}

	// JWT token'ı oluştur
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)

	// Token'ı imzala
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ValidateAccessToken, bir JWT token'ını doğrular ve içerisindeki claims'i döndürür
func ValidateAccessToken(tokenString string) (*types.TokenClaims, error) {
	// JWT için secret key'i çevresel değişkenlerden al
	secretKey := os.Getenv("JWT_ACCESS_SECRET")
	if secretKey == "" {
		return nil, errors.New("JWT_ACCESS_SECRET environment variable is not set")
	}

	// Claims yapısı oluştur
	claims := &JWTClaims{}

	// Token'ı parse et ve doğrula
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Algoritma kontrolü
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Token geçerliliğini kontrol et
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return &claims.TokenClaims, nil
}

// IsTokenExpired, token'ın süresinin dolup dolmadığını kontrol eder
func IsTokenExpired(tokenString string) (bool, error) {
	// JWT için secret key'i çevresel değişkenlerden al
	secretKey := os.Getenv("JWT_ACCESS_SECRET")
	if secretKey == "" {
		return true, errors.New("JWT_ACCESS_SECRET environment variable is not set")
	}

	// Claims yapısı oluştur
	claims := &JWTClaims{}

	// Token'ı parse et
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	// Parse hatası JWT'nin süresi dolduysa meydana gelebilir
	if err != nil {
		return true, fmt.Errorf("failed to parse token: %w", err)
	}

	// Token geçerliliğini kontrol et
	if !token.Valid {
		return true, errors.New("invalid token")
	}

	// Token geçerli ve süresi dolmamış
	return false, nil
}

// ExtractClaims, token string'inden claims verisini çıkarır (doğrulama yapmadan)
func ExtractClaims(tokenString string) (*types.TokenClaims, error) {
	// Token'ı parse etmeden önce yapısını tanımla
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &JWTClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Claims'i çıkar
	if claims, ok := token.Claims.(*JWTClaims); ok {
		return &claims.TokenClaims, nil
	}

	return nil, errors.New("invalid claims format")
}

// GenerateRefreshToken, benzersiz bir refresh token string'i oluşturur
func GenerateRefreshToken() string {
	// Güvenli random string oluştur
	return GenerateRandomString(configs.REFRESH_TOKEN_LENGTH)
}

// ShouldRefreshToken, token'ın yenilenip yenilenmemesi gerektiğini kontrol eder
// Örneğin, token süresinin %75'i dolduysa yenileme yapılmalı
func ShouldRefreshToken(tokenString string) (bool, error) {
	// JWT için secret key'i çevresel değişkenlerden al
	secretKey := os.Getenv("JWT_ACCESS_SECRET")
	if secretKey == "" {
		return false, errors.New("JWT_ACCESS_SECRET environment variable is not set")
	}

	// Claims yapısı oluştur
	claims := &JWTClaims{}

	// Token'ı parse et
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	// Parse hatası varsa
	if err != nil {
		return false, fmt.Errorf("failed to parse token: %w", err)
	}

	// Token süresinin ne kadarının kaldığını hesapla
	expiresAt := claims.ExpiresAt.Time
	issuedAt := claims.IssuedAt.Time
	totalDuration := expiresAt.Sub(issuedAt)
	remainingDuration := expiresAt.Sub(time.Now())

	// Toplam sürenin %25'inden az kaldıysa true döndür
	return remainingDuration < (totalDuration / 4), nil
}
