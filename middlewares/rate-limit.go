package middlewares

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/configs"
	cache "github.com/okanay/backend-blog-guideofdubai/services"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

// AIRateLimitMiddleware yapısına temizleme interval'i ekleyelim
type AIRateLimitMiddleware struct {
	cache           *cache.Cache
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
	mu              sync.RWMutex // Senkronizasyon için mutex
}

// Constructor güncellendi
func NewAIRateLimitMiddleware(cache *cache.Cache) *AIRateLimitMiddleware {
	middleware := &AIRateLimitMiddleware{
		cache:           cache,
		cleanupInterval: 10 * time.Minute, // Her 10 dakika temizleme yap
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

// Graceful shutdown için Stop metodu
func (m *AIRateLimitMiddleware) Stop() {
	close(m.stopCleanup)
}

func (m *AIRateLimitMiddleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// User bilgisini al (AuthMiddleware tarafından set edilmiş olmalı)
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

		// Rate limit bilgisini al
		rateInfo := m.getRateLimitInfo(userID.String())

		// Limit kontrolü
		isAllowed := true
		limitMessage := ""

		if rateInfo.RequestCount >= configs.AI_RATE_LIMIT_MAX_REQUESTS {
			isAllowed = false
			limitMessage = fmt.Sprintf("Maksimum istek limiti aşıldı (%d/%d). Limit %s sonra sıfırlanacak.",
				rateInfo.RequestCount, configs.AI_RATE_LIMIT_MAX_REQUESTS,
				time.Until(rateInfo.WindowResetAt).Round(time.Second))
		} else if rateInfo.TokensUsed >= configs.AI_RATE_LIMIT_MAX_TOKENS {
			isAllowed = false
			limitMessage = fmt.Sprintf("Maksimum token limiti aşıldı (%d/%d). Limit %s sonra sıfırlanacak.",
				rateInfo.TokensUsed, configs.AI_RATE_LIMIT_MAX_TOKENS,
				time.Until(rateInfo.WindowResetAt).Round(time.Second))
		} else if rateInfo.RequestsPerMin >= configs.AI_RATE_LIMIT_REQ_PER_MINUTE {
			isAllowed = false
			limitMessage = fmt.Sprintf("Dakika başına istek limiti aşıldı (%d/%d). Lütfen %s sonra tekrar deneyin.",
				rateInfo.RequestsPerMin, configs.AI_RATE_LIMIT_REQ_PER_MINUTE,
				time.Until(rateInfo.MinuteStartedAt.Add(time.Minute)).Round(time.Second))
		}

		// Rate limit başlıklarını ekle
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_MAX_REQUESTS))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_MAX_REQUESTS-rateInfo.RequestCount))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", rateInfo.WindowResetAt.Unix()))
		c.Header("X-RateLimit-Minute-Limit", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_REQ_PER_MINUTE))
		c.Header("X-RateLimit-Minute-Remaining", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_REQ_PER_MINUTE-rateInfo.RequestsPerMin))
		c.Header("X-RateLimit-Minute-Reset", fmt.Sprintf("%d", rateInfo.MinuteStartedAt.Add(time.Minute).Unix()))

		if !isAllowed {
			retryAfter := 60 // Varsayılan değer (1 dakika)

			// Doğru Retry-After değerini belirle
			if strings.Contains(limitMessage, "Dakika başına") {
				retryAfter = int(time.Until(rateInfo.MinuteStartedAt.Add(time.Minute)).Seconds())
			} else {
				retryAfter = int(time.Until(rateInfo.WindowResetAt).Seconds())
			}

			if retryAfter < 0 {
				retryAfter = 0
			}

			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "rate_limit_exceeded",
				"message": limitMessage,
				"data": gin.H{
					"resetAt":    rateInfo.WindowResetAt,
					"retryAfter": retryAfter,
					"limits": gin.H{
						"requests": gin.H{
							"total": gin.H{
								"limit":     configs.AI_RATE_LIMIT_MAX_REQUESTS,
								"used":      rateInfo.RequestCount,
								"remaining": configs.AI_RATE_LIMIT_MAX_REQUESTS - rateInfo.RequestCount,
								"reset":     rateInfo.WindowResetAt,
							},
							"perMinute": gin.H{
								"limit":     configs.AI_RATE_LIMIT_REQ_PER_MINUTE,
								"used":      rateInfo.RequestsPerMin,
								"remaining": configs.AI_RATE_LIMIT_REQ_PER_MINUTE - rateInfo.RequestsPerMin,
								"reset":     rateInfo.MinuteStartedAt.Add(time.Minute),
							},
						},
						"tokens": gin.H{
							"limit":     configs.AI_RATE_LIMIT_MAX_TOKENS,
							"used":      rateInfo.TokensUsed,
							"remaining": configs.AI_RATE_LIMIT_MAX_TOKENS - rateInfo.TokensUsed,
						},
					},
				},
			})
			c.Abort()
			return
		}

		// İstek sayaçlarını güncelle (istek öncesi)
		rateInfo.RequestCount++
		rateInfo.RequestsPerMin++
		rateInfo.LastRequest = time.Now()

		// Rate limit bilgisini kaydet
		m.saveRateLimitInfo(rateInfo)

		// Request öncesi zamanı kaydet
		requestStart := time.Now()

		// Sonraki middleware'lere devam et
		c.Next()

		// İstek başarılı ise (sadece 200 OK durumunda)
		if c.Writer.Status() == http.StatusOK {
			// İstek süresini hesapla
			requestDuration := time.Since(requestStart)

			// Log kaydı oluştur (opsiyonel)
			fmt.Printf("AI Request from user %s took %v\n", userID.String(), requestDuration)

			// Yanıt gövdesinden kullanılan token sayısını al
			// Önce rate limit bilgisini yeniden yükle
			updatedRateInfo := m.getRateLimitInfo(userID.String())

			// Response body'den token kullanım bilgisini al (mümkünse)
			tokensUsed := 0

			// Yanıt içeriğinden token kullanımını bulmaya çalış
			if ctxTokens, exists := c.Get("tokens_used"); exists {
				if tokensInt, ok := ctxTokens.(int); ok {
					tokensUsed = tokensInt
				}
			}

			// Eğer token bilgisi bulunamadıysa tahmin et
			if tokensUsed == 0 {
				// Varsayılan değer - AI handler'da güncellenecek
				tokensUsed = 1000
			}

			// Token kullanımını güncelle
			updatedRateInfo.TokensUsed += tokensUsed

			// Güncellenmiş rate limit bilgisini kaydet
			m.saveRateLimitInfo(updatedRateInfo)
		}
	}
}

func (m *AIRateLimitMiddleware) checkRateLimit(userID string) (*types.RateLimitInfo, bool, time.Time, string) {
	// Cache eriişimi için lock kullan
	m.mu.RLock()
	defer m.mu.RUnlock()

	cacheKey := fmt.Sprintf("ai_rate_limit:%s", userID)

	// Cache'den rate limit bilgisini al
	data, exists := m.cache.Get(cacheKey)

	now := time.Now()
	var rateInfo types.RateLimitInfo
	var limitMessage string

	// Cache'de yoksa veya süresi dolmuşsa yeni bir rate limit kaydı oluştur
	if !exists {
		rateInfo = types.RateLimitInfo{
			UserID:          userID,
			RequestCount:    0,
			TokensUsed:      0,
			RequestsPerMin:  0,
			FirstRequest:    now,
			LastRequest:     now,
			WindowResetAt:   now.Add(configs.AI_RATE_LIMIT_WINDOW),
			MinuteStartedAt: now,
		}
	} else {
		// Var olan rate limit bilgisini unmarshal et
		if err := json.Unmarshal(data, &rateInfo); err != nil {
			// Hata durumunda yeni bir kayıt oluştur
			rateInfo = types.RateLimitInfo{
				UserID:          userID,
				RequestCount:    0,
				TokensUsed:      0,
				RequestsPerMin:  0,
				FirstRequest:    now,
				LastRequest:     now,
				WindowResetAt:   now.Add(configs.AI_RATE_LIMIT_WINDOW),
				MinuteStartedAt: now,
			}
		}

		// Zaman penceresi süresi dolmuş mu kontrol et
		if now.After(rateInfo.WindowResetAt) {
			// Süresi dolmuşsa, pencereyi sıfırla
			rateInfo = types.RateLimitInfo{
				UserID:          userID,
				RequestCount:    0,
				TokensUsed:      0,
				RequestsPerMin:  0,
				FirstRequest:    now,
				LastRequest:     now,
				WindowResetAt:   now.Add(configs.AI_RATE_LIMIT_WINDOW),
				MinuteStartedAt: now,
			}
		} else {
			// Dakika başına istek sayısını kontrol et
			if now.Sub(rateInfo.MinuteStartedAt) > time.Minute {
				// Dakika dolmuşsa dakika sayacını sıfırla
				rateInfo.RequestsPerMin = 0
				rateInfo.MinuteStartedAt = now
			}
		}
	}

	// Rate limit'i aşmış mı kontrol et (OR operatörü kullanıldı)
	isAllowed := true

	if rateInfo.RequestCount >= configs.AI_RATE_LIMIT_MAX_REQUESTS {
		isAllowed = false
		limitMessage = fmt.Sprintf("Maksimum istek limiti aşıldı (%d/%d). Limit %s sonra sıfırlanacak.",
			rateInfo.RequestCount, configs.AI_RATE_LIMIT_MAX_REQUESTS,
			time.Until(rateInfo.WindowResetAt).Round(time.Second))
	} else if rateInfo.TokensUsed >= configs.AI_RATE_LIMIT_MAX_TOKENS {
		isAllowed = false
		limitMessage = fmt.Sprintf("Maksimum token limiti aşıldı (%d/%d). Limit %s sonra sıfırlanacak.",
			rateInfo.TokensUsed, configs.AI_RATE_LIMIT_MAX_TOKENS,
			time.Until(rateInfo.WindowResetAt).Round(time.Second))
	} else if rateInfo.RequestsPerMin >= configs.AI_RATE_LIMIT_REQ_PER_MINUTE {
		isAllowed = false
		limitMessage = fmt.Sprintf("Dakika başına istek limiti aşıldı (%d/%d). Lütfen %s sonra tekrar deneyin.",
			rateInfo.RequestsPerMin, configs.AI_RATE_LIMIT_REQ_PER_MINUTE,
			time.Until(rateInfo.MinuteStartedAt.Add(time.Minute)).Round(time.Second))
	}

	return &rateInfo, isAllowed, rateInfo.WindowResetAt, limitMessage
}

func (m *AIRateLimitMiddleware) updateRateLimit(userID string, tokensUsed int) {
	cacheKey := fmt.Sprintf("ai_rate_limit:%s", userID)

	// Cache'den mevcut rate limit bilgisini al
	data, exists := m.cache.Get(cacheKey)

	now := time.Now()
	var rateInfo types.RateLimitInfo

	if !exists {
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
			rateInfo.RequestCount++
			rateInfo.TokensUsed += tokensUsed
			rateInfo.LastRequest = now
		}
	}

	// Güncellenmiş bilgiyi marshal et ve cache'e yaz
	jsonData, _ := json.Marshal(rateInfo)
	m.cache.SetWithTTL(cacheKey, jsonData, configs.AI_RATE_LIMIT_WINDOW)
}

// getRateLimitInfo cache'den rate limit bilgisini alır veya yeni oluşturur
func (m *AIRateLimitMiddleware) getRateLimitInfo(userID string) *types.RateLimitInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cacheKey := fmt.Sprintf("ai_rate_limit:%s", userID)
	data, exists := m.cache.Get(cacheKey)

	now := time.Now()

	if !exists {
		return &types.RateLimitInfo{
			UserID:          userID,
			RequestCount:    0,
			TokensUsed:      0,
			RequestsPerMin:  0,
			FirstRequest:    now,
			LastRequest:     now,
			WindowResetAt:   now.Add(configs.AI_RATE_LIMIT_WINDOW),
			MinuteStartedAt: now,
		}
	}

	var rateInfo types.RateLimitInfo
	if err := json.Unmarshal(data, &rateInfo); err != nil {
		return &types.RateLimitInfo{
			UserID:          userID,
			RequestCount:    0,
			TokensUsed:      0,
			RequestsPerMin:  0,
			FirstRequest:    now,
			LastRequest:     now,
			WindowResetAt:   now.Add(configs.AI_RATE_LIMIT_WINDOW),
			MinuteStartedAt: now,
		}
	}

	// Zaman penceresi kontrolü
	if now.After(rateInfo.WindowResetAt) {
		return &types.RateLimitInfo{
			UserID:          userID,
			RequestCount:    0,
			TokensUsed:      0,
			RequestsPerMin:  0,
			FirstRequest:    now,
			LastRequest:     now,
			WindowResetAt:   now.Add(configs.AI_RATE_LIMIT_WINDOW),
			MinuteStartedAt: now,
		}
	}

	// Dakika kontrolü
	if now.Sub(rateInfo.MinuteStartedAt) > time.Minute {
		rateInfo.RequestsPerMin = 0
		rateInfo.MinuteStartedAt = now
	}

	return &rateInfo
}

// saveRateLimitInfo rate limit bilgisini cache'e kaydeder
func (m *AIRateLimitMiddleware) saveRateLimitInfo(rateInfo *types.RateLimitInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cacheKey := fmt.Sprintf("ai_rate_limit:%s", rateInfo.UserID)
	jsonData, err := json.Marshal(rateInfo)
	if err != nil {
		return err
	}

	m.cache.SetWithTTL(cacheKey, jsonData, configs.AI_RATE_LIMIT_WINDOW)
	return nil
}
