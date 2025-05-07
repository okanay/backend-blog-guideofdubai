package middlewares

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/configs"
	cache "github.com/okanay/backend-blog-guideofdubai/services"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

// AIRateLimitMiddleware AI servisleri için rate limiting uygular
type AIRateLimitMiddleware struct {
	cache           *cache.Cache
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
}

// NewAIRateLimitMiddleware yeni bir AI rate limit middleware oluşturur
func NewAIRateLimitMiddleware(cache *cache.Cache) *AIRateLimitMiddleware {
	middleware := &AIRateLimitMiddleware{
		cache:           cache,
		cleanupInterval: 10 * time.Minute,
		stopCleanup:     make(chan struct{}),
	}

	// Temizleme görevini başlat
	go middleware.startCleanupRoutine()

	return middleware
}

// Temizleme rutini
func (m *AIRateLimitMiddleware) startCleanupRoutine() {
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanupExpiredRateLimits()
		case <-m.stopCleanup:
			return
		}
	}
}

// Süresi dolan rate limit kayıtlarını temizle
func (m *AIRateLimitMiddleware) cleanupExpiredRateLimits() {
	now := time.Now()
	prefix := "ai_rate_limit:"

	// Tüm rate limit kayıtlarını al
	allRateLimits := m.cache.GetAllWithPrefix(prefix)

	for key, data := range allRateLimits {
		var rateInfo types.RateLimitInfo
		if err := json.Unmarshal(data, &rateInfo); err == nil {
			// Süresi dolan kayıtları temizle
			if now.After(rateInfo.WindowResetAt) {
				m.cache.Delete(key)
			}
		}
	}
}

// Stop middleware'i durdurur (graceful shutdown için)
func (m *AIRateLimitMiddleware) Stop() {
	close(m.stopCleanup)
}

// RateLimit rate limit middleware fonksiyonunu döndürür
func (m *AIRateLimitMiddleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Kullanıcı ID'sini al
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "unauthorized",
				"message": "Kullanıcı kimliği bulunamadı",
			})
			c.Abort()
			return
		}

		userID, ok := userIDInterface.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "internal_error",
				"message": "Kullanıcı kimliği geçersiz formatta",
			})
			c.Abort()
			return
		}

		// Rate limit bilgisini al veya oluştur
		rateLimit, isLimited, resetTime := m.checkAndUpdateRateLimit(userID.String())
		if isLimited {
			// Rate limit başlıklarını ekle
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_MAX_REQUESTS))
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", 0))
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))
			c.Header("Retry-After", fmt.Sprintf("%d", int(time.Until(resetTime).Seconds())))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "rate_limit_exceeded",
				"message": fmt.Sprintf("Rate limit aşıldı. %s sonra tekrar deneyin.", time.Until(resetTime).Round(time.Second)),
				"data": gin.H{
					"reset":     resetTime,
					"remaining": 0,
					"limit":     configs.AI_RATE_LIMIT_MAX_REQUESTS,
				},
			})
			c.Abort()
			return
		}

		// Rate limit başlıklarını ekle
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_MAX_REQUESTS))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_MAX_REQUESTS-rateLimit.RequestCount))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", rateLimit.WindowResetAt.Unix()))

		// Orijinal isteği işle
		c.Next()

		// Yanıt sonrası token sayısını güncelle
		if c.Writer.Status() == http.StatusOK {
			tokensUsed := 0
			if ctxTokens, exists := c.Get("tokens_used"); exists {
				if tokensInt, ok := ctxTokens.(int); ok {
					tokensUsed = tokensInt
				}
			}

			if tokensUsed > 0 {
				m.updateTokenUsage(userID.String(), tokensUsed)
			}
		}
	}
}

// checkAndUpdateRateLimit rate limit kontrolü yapar ve istek sayısını günceller
func (m *AIRateLimitMiddleware) checkAndUpdateRateLimit(userID string) (*types.RateLimitInfo, bool, time.Time) {
	cacheKey := fmt.Sprintf("ai_rate_limit:%s", userID)
	now := time.Now()

	// Cache'den mevcut bilgiyi al
	data, exists := m.cache.Get(cacheKey)
	var rateInfo types.RateLimitInfo

	if !exists {
		// Yeni kullanıcı, ilk kayıt oluştur
		rateInfo = types.RateLimitInfo{
			UserID:          userID,
			RequestCount:    1, // İlk istek
			TokensUsed:      0,
			RequestsPerMin:  1,
			FirstRequest:    now,
			LastRequest:     now,
			WindowResetAt:   now.Add(configs.AI_RATE_LIMIT_WINDOW),
			MinuteStartedAt: now,
		}
	} else {
		// Mevcut kaydı yükle
		if err := json.Unmarshal(data, &rateInfo); err != nil {
			// Hata durumunda yeni kayıt oluştur
			rateInfo = types.RateLimitInfo{
				UserID:          userID,
				RequestCount:    1,
				TokensUsed:      0,
				RequestsPerMin:  1,
				FirstRequest:    now,
				LastRequest:     now,
				WindowResetAt:   now.Add(configs.AI_RATE_LIMIT_WINDOW),
				MinuteStartedAt: now,
			}
		} else {
			// Zaman penceresi sıfırlama kontrolü
			if now.After(rateInfo.WindowResetAt) {
				// Pencere süresi dolmuş, yeni pencere başlat
				rateInfo = types.RateLimitInfo{
					UserID:          userID,
					RequestCount:    1,
					TokensUsed:      0,
					RequestsPerMin:  1,
					FirstRequest:    now,
					LastRequest:     now,
					WindowResetAt:   now.Add(configs.AI_RATE_LIMIT_WINDOW),
					MinuteStartedAt: now,
				}
			} else {
				// Dakika başına kontrol
				if now.Sub(rateInfo.MinuteStartedAt) > time.Minute {
					rateInfo.RequestsPerMin = 1
					rateInfo.MinuteStartedAt = now
				} else {
					rateInfo.RequestsPerMin++
				}

				// İstek sayısını artır
				rateInfo.RequestCount++
				rateInfo.LastRequest = now
			}
		}
	}

	// Rate limit kontrolü
	isLimited := false

	// Ana limit kontrolü
	if rateInfo.RequestCount > configs.AI_RATE_LIMIT_MAX_REQUESTS {
		isLimited = true
	}

	// Token limit kontrolü
	if rateInfo.TokensUsed >= configs.AI_RATE_LIMIT_MAX_TOKENS {
		isLimited = true
	}

	// Dakika başına limit kontrolü
	if rateInfo.RequestsPerMin > configs.AI_RATE_LIMIT_REQ_PER_MINUTE {
		isLimited = true
	}

	// Cache'i güncelle (sınır aşılsa bile kaydedelim)
	jsonData, _ := json.Marshal(rateInfo)
	m.cache.SetWithTTL(cacheKey, jsonData, configs.AI_RATE_LIMIT_WINDOW)

	return &rateInfo, isLimited, rateInfo.WindowResetAt
}

// updateTokenUsage kullanıcının token kullanımını günceller
func (m *AIRateLimitMiddleware) updateTokenUsage(userID string, tokensUsed int) {
	cacheKey := fmt.Sprintf("ai_rate_limit:%s", userID)

	// Cache'den mevcut bilgiyi al
	data, exists := m.cache.Get(cacheKey)
	if !exists {
		return // Kullanıcı kaydı yoksa çık
	}

	var rateInfo types.RateLimitInfo
	if err := json.Unmarshal(data, &rateInfo); err != nil {
		return // Hata varsa çık
	}

	// Token kullanımını güncelle
	rateInfo.TokensUsed += tokensUsed

	// Cache'i güncelle
	jsonData, _ := json.Marshal(rateInfo)
	m.cache.SetWithTTL(cacheKey, jsonData, configs.AI_RATE_LIMIT_WINDOW)
}
