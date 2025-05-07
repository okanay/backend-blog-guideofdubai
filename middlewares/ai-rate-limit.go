package middlewares

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/configs"
	cache "github.com/okanay/backend-blog-guideofdubai/services"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

type AIRateLimitMiddleware struct {
	cache *cache.Cache
}

func NewAIRateLimitMiddleware(cache *cache.Cache) *AIRateLimitMiddleware {
	return &AIRateLimitMiddleware{
		cache: cache,
	}
}

func (m *AIRateLimitMiddleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Kullanıcı doğrulama
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

		// Rate limit kontrolü
		rateInfo, minuteLimit := m.getRateLimitInfo(userID.String())

		// Rate limit başlıkları
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_MAX_REQUESTS))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_MAX_REQUESTS-rateInfo.RequestCount))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", rateInfo.WindowResetAt.Unix()))
		c.Header("X-RateLimit-Minute-Limit", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_REQ_PER_MINUTE))
		c.Header("X-RateLimit-Minute-Remaining", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_REQ_PER_MINUTE-minuteLimit))

		// Limit kontrolleri ve hata mesajları
		limitStatus := m.checkRateLimitStatus(rateInfo, minuteLimit)

		if limitStatus.IsLimited {
			c.Header("Retry-After", fmt.Sprintf("%d", limitStatus.RetryAfter))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "rate_limit_exceeded",
				"message": limitStatus.Message,
				"data": gin.H{
					"resetAt":         rateInfo.WindowResetAt,
					"retryAfter":      limitStatus.RetryAfter,
					"limitType":       limitStatus.LimitType,
					"totalLimit":      configs.AI_RATE_LIMIT_MAX_REQUESTS,
					"totalRemaining":  configs.AI_RATE_LIMIT_MAX_REQUESTS - rateInfo.RequestCount,
					"minuteLimit":     configs.AI_RATE_LIMIT_REQ_PER_MINUTE,
					"minuteRemaining": configs.AI_RATE_LIMIT_REQ_PER_MINUTE - minuteLimit,
					"tokenLimit":      configs.AI_RATE_LIMIT_MAX_TOKENS,
					"tokenRemaining":  configs.AI_RATE_LIMIT_MAX_TOKENS - rateInfo.TokensUsed,
					"windowMinutes":   int(configs.AI_RATE_LIMIT_WINDOW.Minutes()),
					"windowReset":     rateInfo.WindowResetAt.Format("15:04:05"),
					"minuteReset":     limitStatus.MinuteResetTime.Format("15:04:05"),
					"currentTime":     time.Now().Format("15:04:05"),
					"explanation":     "AI servisleri hem maliyetli olduğundan hem de adil kullanımı sağlamak için istek limitlerini uyguluyoruz.",
				},
			})
			c.Abort()
			return
		}

		// İstek sayacını artır
		m.incrementRequestCount(userID.String(), rateInfo)

		// İsteği işle
		c.Next()

		// Başarılı istek sonrası token kullanımını güncelle
		if c.Writer.Status() == http.StatusOK {
			tokensUsed := 1000
			m.updateTokenUsage(userID.String(), tokensUsed)
		}
	}
}

// Rate limit bilgilerini alır
func (m *AIRateLimitMiddleware) getRateLimitInfo(userID string) (*types.RateLimitInfo, int) {
	cacheKey := fmt.Sprintf("ai_rate_limit:%s", userID)
	minuteKey := fmt.Sprintf("ai_rate_limit_minute:%s", userID)

	now := time.Now()
	var rateInfo types.RateLimitInfo

	// Dakika limitini kontrol et
	minuteCount := 0
	minuteData, minuteExists := m.cache.Get(minuteKey)

	if minuteExists {
		if count, err := strconv.Atoi(string(minuteData)); err == nil {
			minuteCount = count
		}
	}

	// Toplam limiti kontrol et
	data, exists := m.cache.Get(cacheKey)

	if !exists {
		// Yeni kayıt oluştur
		rateInfo = types.RateLimitInfo{
			UserID:        userID,
			RequestCount:  0,
			TokensUsed:    0,
			FirstRequest:  now,
			LastRequest:   now,
			WindowResetAt: now.Add(configs.AI_RATE_LIMIT_WINDOW),
		}
	} else {
		if err := json.Unmarshal(data, &rateInfo); err != nil {
			// Hata durumunda yeni kayıt
			rateInfo = types.RateLimitInfo{
				UserID:        userID,
				RequestCount:  0,
				TokensUsed:    0,
				FirstRequest:  now,
				LastRequest:   now,
				WindowResetAt: now.Add(configs.AI_RATE_LIMIT_WINDOW),
			}
		} else {
			// Zaman penceresi dolmuş mu kontrol et
			if now.After(rateInfo.WindowResetAt) {
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
	}

	return &rateInfo, minuteCount
}

// Rate limit durumunu kontrol eder
type LimitStatus struct {
	IsLimited       bool
	LimitType       string
	Message         string
	RetryAfter      int
	MinuteResetTime time.Time
}

func (m *AIRateLimitMiddleware) checkRateLimitStatus(rateInfo *types.RateLimitInfo, minuteCount int) LimitStatus {
	now := time.Now()
	status := LimitStatus{
		IsLimited:       false,
		MinuteResetTime: now.Add(1 * time.Minute).Truncate(time.Minute),
	}

	// Dakikalık limit kontrolü
	isMinuteLimitExceeded := minuteCount >= configs.AI_RATE_LIMIT_REQ_PER_MINUTE

	// Toplam limit kontrolü
	isTotalLimitExceeded := rateInfo.RequestCount >= configs.AI_RATE_LIMIT_MAX_REQUESTS

	// Token limit kontrolü
	isTokenLimitExceeded := rateInfo.TokensUsed >= configs.AI_RATE_LIMIT_MAX_TOKENS

	// Yalnızca dakika limiti aşılmış
	if isMinuteLimitExceeded && !isTotalLimitExceeded && !isTokenLimitExceeded {
		status.IsLimited = true
		status.LimitType = "minute"
		status.RetryAfter = int(status.MinuteResetTime.Sub(now).Seconds())
		status.Message = fmt.Sprintf(
			"Dakika limiti aşıldı. Dakika başına en fazla %d istek yapabilirsiniz. "+
				"Bir sonraki isteğinizi %s'de yapabilirsiniz.",
			configs.AI_RATE_LIMIT_REQ_PER_MINUTE,
			status.MinuteResetTime.Format("15:04:05"),
		)
		return status
	}

	// Yalnızca toplam limit aşılmış
	if isTotalLimitExceeded && !isMinuteLimitExceeded {
		status.IsLimited = true
		status.LimitType = "total"
		status.RetryAfter = int(rateInfo.WindowResetAt.Sub(now).Seconds())
		status.Message = fmt.Sprintf(
			"Toplam limit aşıldı. %d dakikalık süre içinde en fazla %d istek yapabilirsiniz. "+
				"Limitiniz %s'de yenilenecek.",
			int(configs.AI_RATE_LIMIT_WINDOW.Minutes()),
			configs.AI_RATE_LIMIT_MAX_REQUESTS,
			rateInfo.WindowResetAt.Format("15:04:05"),
		)
		return status
	}

	// Hem dakika hem toplam limit aşılmış
	if isMinuteLimitExceeded && isTotalLimitExceeded {
		status.IsLimited = true
		status.LimitType = "both"

		// Hangi limit daha uzun sürecek?
		minuteRetry := int(status.MinuteResetTime.Sub(now).Seconds())
		totalRetry := int(rateInfo.WindowResetAt.Sub(now).Seconds())

		if totalRetry > minuteRetry {
			status.RetryAfter = totalRetry
			status.Message = fmt.Sprintf(
				"Hem dakika hem toplam limit aşıldı. "+
					"Dakika limiti (%d istek/dk) %s'de, "+
					"toplam limit (%d istek/%d dk) ise %s'de yenilenecek.",
				configs.AI_RATE_LIMIT_REQ_PER_MINUTE,
				status.MinuteResetTime.Format("15:04:05"),
				configs.AI_RATE_LIMIT_MAX_REQUESTS,
				int(configs.AI_RATE_LIMIT_WINDOW.Minutes()),
				rateInfo.WindowResetAt.Format("15:04:05"),
			)
		} else {
			status.RetryAfter = minuteRetry
			status.Message = fmt.Sprintf(
				"Hem dakika hem toplam limit aşıldı. "+
					"Dakika başına %d istek yapabilirsiniz. "+
					"Bir sonraki isteğinizi %s'de yapabilirsiniz.",
				configs.AI_RATE_LIMIT_REQ_PER_MINUTE,
				status.MinuteResetTime.Format("15:04:05"),
			)
		}
		return status
	}

	// Yalnızca token limiti aşılmış
	if isTokenLimitExceeded {
		status.IsLimited = true
		status.LimitType = "token"
		status.RetryAfter = int(rateInfo.WindowResetAt.Sub(now).Seconds())
		status.Message = fmt.Sprintf(
			"Token limiti aşıldı. %d dakikalık süre içinde en fazla %d token kullanabilirsiniz. "+
				"Token limitiniz %s'de yenilenecek.",
			int(configs.AI_RATE_LIMIT_WINDOW.Minutes()),
			configs.AI_RATE_LIMIT_MAX_TOKENS,
			rateInfo.WindowResetAt.Format("15:04:05"),
		)
		return status
	}

	return status
}

// İstek sayacını artırır
func (m *AIRateLimitMiddleware) incrementRequestCount(userID string, rateInfo *types.RateLimitInfo) {
	cacheKey := fmt.Sprintf("ai_rate_limit:%s", userID)
	minuteKey := fmt.Sprintf("ai_rate_limit_minute:%s", userID)

	now := time.Now()

	// Ana sayacı artır
	rateInfo.RequestCount++
	rateInfo.LastRequest = now

	// Bilgiyi marshal et
	if jsonData, err := json.Marshal(rateInfo); err == nil {
		// Kalan süreyi hesapla
		remainingTime := rateInfo.WindowResetAt.Sub(now)
		if remainingTime <= 0 {
			remainingTime = configs.AI_RATE_LIMIT_WINDOW
		}

		// Cache'e yaz
		m.cache.SetWithTTL(cacheKey, jsonData, remainingTime)
	}

	// Dakika sayacını artır
	minuteCount := 1
	if minuteData, exists := m.cache.Get(minuteKey); exists {
		if count, err := strconv.Atoi(string(minuteData)); err == nil {
			minuteCount = count + 1
		}
	}

	// Dakika sayacını kaydet
	m.cache.SetWithTTL(minuteKey, []byte(strconv.Itoa(minuteCount)), 1*time.Minute)
}

// Token kullanımını günceller
func (m *AIRateLimitMiddleware) updateTokenUsage(userID string, tokensUsed int) {
	cacheKey := fmt.Sprintf("ai_rate_limit:%s", userID)

	// Cache'den bilgiyi al
	data, exists := m.cache.Get(cacheKey)
	if !exists {
		return
	}

	var rateInfo types.RateLimitInfo
	if err := json.Unmarshal(data, &rateInfo); err != nil {
		return
	}

	now := time.Now()

	// Token kullanımını güncelle
	rateInfo.TokensUsed += tokensUsed

	// Güncellenmiş bilgiyi marshal et
	if jsonData, err := json.Marshal(rateInfo); err == nil {
		// Kalan süreyi hesapla
		remainingTime := rateInfo.WindowResetAt.Sub(now)
		if remainingTime <= 0 {
			remainingTime = configs.AI_RATE_LIMIT_WINDOW
		}

		// Cache'e yaz
		m.cache.SetWithTTL(cacheKey, jsonData, remainingTime)
	}
}
