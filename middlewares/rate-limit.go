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
		rateLimit, isLimited, resetTime := m.checkRateLimit(userID.String())
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
				m.updateRateLimit(userID.String(), tokensUsed)
			}
		}
	}
}

func (m *AIRateLimitMiddleware) checkRateLimit(userID string) (*types.RateLimitInfo, bool, time.Time) {
	cacheKey := fmt.Sprintf("ai_rate_limit:%s", userID)

	// Cache'den rate limit bilgisini al
	data, exists := m.cache.Get(cacheKey)

	now := time.Now()
	fmt.Printf("[LOG] checkRateLimit - UserID: %s, Time: %s\n", userID, now.Format(time.RFC3339))
	fmt.Printf("[LOG] Cache Key: %s, Exists: %v\n", cacheKey, exists)

	var rateInfo types.RateLimitInfo

	// Cache'de yoksa veya süresi dolmuşsa yeni bir rate limit kaydı oluştur
	if !exists {
		fmt.Printf("[LOG] Cache için bilgi bulunamadı, yeni bir kayıt oluşturuluyor\n")
		rateInfo = types.RateLimitInfo{
			UserID:        userID,
			RequestCount:  0,
			TokensUsed:    0,
			FirstRequest:  now,
			LastRequest:   now,
			WindowResetAt: now.Add(configs.AI_RATE_LIMIT_WINDOW),
		}
	} else {
		// Var olan rate limit bilgisini unmarshal et
		if err := json.Unmarshal(data, &rateInfo); err != nil {
			fmt.Printf("[LOG] Unmarshal hatası: %v\n", err)
			// Hata durumunda yeni bir kayıt oluştur
			rateInfo = types.RateLimitInfo{
				UserID:        userID,
				RequestCount:  0,
				TokensUsed:    0,
				FirstRequest:  now,
				LastRequest:   now,
				WindowResetAt: now.Add(configs.AI_RATE_LIMIT_WINDOW),
			}
		} else {
			fmt.Printf("[LOG] Mevcut rate info - RequestCount: %d, TokensUsed: %d, WindowResetAt: %s\n",
				rateInfo.RequestCount, rateInfo.TokensUsed, rateInfo.WindowResetAt.Format(time.RFC3339))
		}

		// Zaman penceresi süresi dolmuş mu kontrol et
		if now.After(rateInfo.WindowResetAt) {
			fmt.Printf("[LOG] Zaman penceresi dolmuş, pencere sıfırlanıyor\n")
			// Süresi dolmuşsa, pencereyi sıfırla
			rateInfo = types.RateLimitInfo{
				UserID:        userID,
				RequestCount:  0,
				TokensUsed:    0,
				FirstRequest:  now,
				LastRequest:   now,
				WindowResetAt: now.Add(configs.AI_RATE_LIMIT_WINDOW),
			}
		}
	}

	// Rate limit'i aşmış mı kontrol et
	isAllowed := rateInfo.RequestCount < configs.AI_RATE_LIMIT_MAX_REQUESTS &&
		rateInfo.TokensUsed < configs.AI_RATE_LIMIT_MAX_TOKENS

	fmt.Printf("[LOG] Rate limit kontrolü - İzin veriliyor mu: %v\n", isAllowed)

	return &rateInfo, isAllowed, rateInfo.WindowResetAt
}

func (m *AIRateLimitMiddleware) updateRateLimit(userID string, tokensUsed int) {
	cacheKey := fmt.Sprintf("ai_rate_limit:%s", userID)
	fmt.Printf("[LOG] updateRateLimit - UserID: %s, TokensUsed: %d\n", userID, tokensUsed)

	// Cache'den mevcut rate limit bilgisini al
	data, exists := m.cache.Get(cacheKey)
	fmt.Printf("[LOG] Cache Key: %s, Exists: %v\n", cacheKey, exists)

	now := time.Now()
	var rateInfo types.RateLimitInfo

	if !exists {
		fmt.Printf("[LOG] Güncellenecek cache bilgisi bulunamadı, yeni bir kayıt oluşturuluyor\n")
		// Cache'de yoksa yeni bir kayıt oluştur
		rateInfo = types.RateLimitInfo{
			UserID:        userID,
			RequestCount:  1,
			TokensUsed:    tokensUsed,
			FirstRequest:  now,
			LastRequest:   now,
			WindowResetAt: now.Add(configs.AI_RATE_LIMIT_WINDOW),
		}
	} else {
		// Var olan bilgiyi unmarshal et
		if err := json.Unmarshal(data, &rateInfo); err != nil {
			fmt.Printf("[LOG] Unmarshal hatası: %v\n", err)
			// Hata durumunda yeni bir kayıt oluştur
			rateInfo = types.RateLimitInfo{
				UserID:        userID,
				RequestCount:  1,
				TokensUsed:    tokensUsed,
				FirstRequest:  now,
				LastRequest:   now,
				WindowResetAt: now.Add(configs.AI_RATE_LIMIT_WINDOW),
			}
		} else {
			// Mevcut bilgiyi güncelle
			fmt.Printf("[LOG] Güncelleme öncesi - RequestCount: %d, TokensUsed: %d, WindowResetAt: %s\n",
				rateInfo.RequestCount, rateInfo.TokensUsed, rateInfo.WindowResetAt.Format(time.RFC3339))

			rateInfo.RequestCount++
			rateInfo.TokensUsed += tokensUsed
			rateInfo.LastRequest = now

			fmt.Printf("[LOG] Güncelleme sonrası - RequestCount: %d, TokensUsed: %d, WindowResetAt: %s\n",
				rateInfo.RequestCount, rateInfo.TokensUsed, rateInfo.WindowResetAt.Format(time.RFC3339))
		}
	}

	// Güncellenmiş bilgiyi marshal et ve cache'e yaz
	jsonData, _ := json.Marshal(rateInfo)
	cacheExpiry := rateInfo.WindowResetAt.Sub(now)
	fmt.Printf("[LOG] Cache'e yazılıyor - Expiry Duration: %v\n", cacheExpiry)

	m.cache.SetWithTTL(cacheKey, jsonData, configs.AI_RATE_LIMIT_WINDOW)

	// Cache'e yazıldıktan sonra kontrol
	dataAfter, existsAfter := m.cache.Get(cacheKey)
	if existsAfter {
		var checkInfo types.RateLimitInfo
		json.Unmarshal(dataAfter, &checkInfo)
		fmt.Printf("[LOG] Cache'e yazıldıktan sonra kontrol - RequestCount: %d, WindowResetAt: %s\n",
			checkInfo.RequestCount, checkInfo.WindowResetAt.Format(time.RFC3339))
	} else {
		fmt.Printf("[LOG] HATA: Cache'e yazıldıktan sonra veri bulunamadı!\n")
	}
}
