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
		rateInfo, allowed, resetTime, minuteLimit := m.checkRateLimit(userID.String())
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_MAX_REQUESTS))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_MAX_REQUESTS-rateInfo.RequestCount))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))
		c.Header("X-RateLimit-Minute-Limit", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_REQ_PER_MINUTE))
		c.Header("X-RateLimit-Minute-Remaining", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_REQ_PER_MINUTE-minuteLimit))

		if !allowed {
			retryAfter := int(time.Until(resetTime).Seconds())
			retryAfter = max(retryAfter, 0)
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))

			// Hangi limitin aşıldığını belirle
			var limitType string
			var limitMessage string
			var resetMessage string

			isMinuteLimitExceeded := minuteLimit >= configs.AI_RATE_LIMIT_REQ_PER_MINUTE
			isTotalLimitExceeded := rateInfo.RequestCount >= configs.AI_RATE_LIMIT_MAX_REQUESTS
			isTokenLimitExceeded := rateInfo.TokensUsed >= configs.AI_RATE_LIMIT_MAX_TOKENS

			minuteResetTime := time.Now().Add(1 * time.Minute)
			retryAfterMinutes := 1

			if isMinuteLimitExceeded && (retryAfter > 60) {
				// Eğer toplam limit de aşıldıysa ve dakika limiti de aşıldıysa
				limitType = "hem dakika hem toplam"
				limitMessage = fmt.Sprintf("Dakika başına %d istek ve toplam %d dakikada %d istek",
					configs.AI_RATE_LIMIT_REQ_PER_MINUTE,
					int(configs.AI_RATE_LIMIT_WINDOW.Minutes()),
					configs.AI_RATE_LIMIT_MAX_REQUESTS)
				resetMessage = fmt.Sprintf("Bir sonraki isteği %d dakika sonra yapabilirsiniz, tam limitlere ise %s tarihinde ulaşacaksınız",
					retryAfterMinutes,
					resetTime.Format("15:04:05"))
			} else if isMinuteLimitExceeded {
				// Sadece dakika limiti aşıldıysa
				limitType = "dakika"
				limitMessage = fmt.Sprintf("Dakika başına maksimum %d istek yapabilirsiniz",
					configs.AI_RATE_LIMIT_REQ_PER_MINUTE)
				resetMessage = fmt.Sprintf("Bir sonraki isteği %d dakika sonra (%s) yapabilirsiniz",
					retryAfterMinutes,
					minuteResetTime.Format("15:04:05"))
			} else if isTotalLimitExceeded {
				// Sadece toplam limit aşıldıysa
				limitType = "toplam"
				limitMessage = fmt.Sprintf("%d dakikalık süre içinde toplam %d istek yapabilirsiniz",
					int(configs.AI_RATE_LIMIT_WINDOW.Minutes()),
					configs.AI_RATE_LIMIT_MAX_REQUESTS)
				resetMessage = fmt.Sprintf("Limitiniz %s tarihinde yenilenecektir",
					resetTime.Format("15:04:05"))
			} else if isTokenLimitExceeded {
				// Token limiti aşıldıysa
				limitType = "token"
				limitMessage = fmt.Sprintf("%d dakikalık süre içinde toplam %d token kullanabilirsiniz",
					int(configs.AI_RATE_LIMIT_WINDOW.Minutes()),
					configs.AI_RATE_LIMIT_MAX_TOKENS)
				resetMessage = fmt.Sprintf("Token limitiniz %s tarihinde yenilenecektir",
					resetTime.Format("15:04:05"))
			} else {
				// Genel durum
				limitType = "istek"
				limitMessage = fmt.Sprintf("Dakika başına %d istek ve toplam %d dakikada %d istek yapabilirsiniz",
					configs.AI_RATE_LIMIT_REQ_PER_MINUTE,
					int(configs.AI_RATE_LIMIT_WINDOW.Minutes()),
					configs.AI_RATE_LIMIT_MAX_REQUESTS)
				resetMessage = fmt.Sprintf("Bir sonraki isteği %d saniye sonra yapabilirsiniz", retryAfter)
			}

			// Detaylı rate limit mesajı
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "rate_limit_exceeded",
				"message": fmt.Sprintf("AI %s limiti aşıldı. %s. %s.", limitType, limitMessage, resetMessage),
				"data": gin.H{
					"resetAt":         resetTime,
					"retryAfter":      retryAfter,
					"limitType":       limitType,
					"totalLimit":      configs.AI_RATE_LIMIT_MAX_REQUESTS,
					"totalRemaining":  configs.AI_RATE_LIMIT_MAX_REQUESTS - rateInfo.RequestCount,
					"minuteLimit":     configs.AI_RATE_LIMIT_REQ_PER_MINUTE,
					"minuteRemaining": configs.AI_RATE_LIMIT_REQ_PER_MINUTE - minuteLimit,
					"tokenLimit":      configs.AI_RATE_LIMIT_MAX_TOKENS,
					"tokenRemaining":  configs.AI_RATE_LIMIT_MAX_TOKENS - rateInfo.TokensUsed,
					"windowMinutes":   int(configs.AI_RATE_LIMIT_WINDOW.Minutes()),
					"windowReset":     resetTime.Format("15:04:05"),
					"minuteReset":     minuteResetTime.Format("15:04:05"),
					"currentTime":     time.Now().Format("15:04:05"),
					"explanation":     "AI servisleri maliyetli olduğundan, adil kullanım ve kaynak yönetimi için API istek limitlerini uyguluyoruz. Bu, tüm kullanıcıların servisten adil bir şekilde yararlanmasını sağlar.",
				},
			})
			c.Abort()
			return
		}
		m.incrementRequestCount(userID.String(), rateInfo)
		c.Next()
		if c.Writer.Status() == http.StatusOK {
			tokensUsed := 1000
			m.updateTokenUsage(userID.String(), tokensUsed)
		}
	}
}

func (m *AIRateLimitMiddleware) checkRateLimit(userID string) (*types.RateLimitInfo, bool, time.Time, int) {
	cacheKey := fmt.Sprintf("ai_rate_limit:%s", userID)
	minuteKey := fmt.Sprintf("ai_rate_limit_minute:%s", userID)

	now := time.Now()
	var rateInfo types.RateLimitInfo
	var windowResetAt time.Time

	minuteCount := 0
	minuteData, minuteExists := m.cache.Get(minuteKey)

	if minuteExists {
		if count, err := parseMinuteCount(minuteData); err == nil {
			minuteCount = count
		}
	}

	data, exists := m.cache.Get(cacheKey)

	if !exists {
		windowResetAt = now.Add(configs.AI_RATE_LIMIT_WINDOW)

		rateInfo = types.RateLimitInfo{
			UserID:        userID,
			RequestCount:  0,
			TokensUsed:    0,
			FirstRequest:  now,
			LastRequest:   now,
			WindowResetAt: windowResetAt,
		}
	} else {
		if err := json.Unmarshal(data, &rateInfo); err != nil {
			windowResetAt = now.Add(configs.AI_RATE_LIMIT_WINDOW)

			rateInfo = types.RateLimitInfo{
				UserID:        userID,
				RequestCount:  0,
				TokensUsed:    0,
				FirstRequest:  now,
				LastRequest:   now,
				WindowResetAt: windowResetAt,
			}
		} else {
			windowResetAt = rateInfo.WindowResetAt

			if now.After(windowResetAt) {
				windowResetAt = now.Add(configs.AI_RATE_LIMIT_WINDOW)

				rateInfo = types.RateLimitInfo{
					UserID:        userID,
					RequestCount:  0,
					TokensUsed:    0,
					FirstRequest:  now,
					LastRequest:   now,
					WindowResetAt: windowResetAt,
				}
			}
		}
	}

	isAllowed := rateInfo.RequestCount < configs.AI_RATE_LIMIT_MAX_REQUESTS &&
		minuteCount < configs.AI_RATE_LIMIT_REQ_PER_MINUTE &&
		rateInfo.TokensUsed < configs.AI_RATE_LIMIT_MAX_TOKENS

	return &rateInfo, isAllowed, windowResetAt, minuteCount
}

func (m *AIRateLimitMiddleware) incrementRequestCount(userID string, currentInfo *types.RateLimitInfo) {
	cacheKey := fmt.Sprintf("ai_rate_limit:%s", userID)
	minuteKey := fmt.Sprintf("ai_rate_limit_minute:%s", userID)

	now := time.Now()

	currentInfo.RequestCount++
	currentInfo.LastRequest = now

	jsonData, err := json.Marshal(currentInfo)
	if err != nil {
		return
	}

	remainingTime := currentInfo.WindowResetAt.Sub(now)
	if remainingTime <= 0 {
		remainingTime = configs.AI_RATE_LIMIT_WINDOW
	}

	m.cache.SetWithTTL(cacheKey, jsonData, remainingTime)

	minuteCount := 1
	minuteData, minuteExists := m.cache.Get(minuteKey)

	if minuteExists {
		if count, err := parseMinuteCount(minuteData); err == nil {
			minuteCount = count + 1
		}
	}

	countData := fmt.Sprintf("%d", minuteCount)
	m.cache.SetWithTTL(minuteKey, []byte(countData), 1*time.Minute)
}

func (m *AIRateLimitMiddleware) updateTokenUsage(userID string, tokensUsed int) {
	cacheKey := fmt.Sprintf("ai_rate_limit:%s", userID)

	data, exists := m.cache.Get(cacheKey)
	if !exists {
		return
	}

	var rateInfo types.RateLimitInfo
	if err := json.Unmarshal(data, &rateInfo); err != nil {
		return
	}

	now := time.Now()

	rateInfo.TokensUsed += tokensUsed

	jsonData, err := json.Marshal(rateInfo)
	if err != nil {
		return
	}

	remainingTime := rateInfo.WindowResetAt.Sub(now)
	if remainingTime <= 0 {
		remainingTime = configs.AI_RATE_LIMIT_WINDOW
	}

	m.cache.SetWithTTL(cacheKey, jsonData, remainingTime)
}

func parseMinuteCount(data []byte) (int, error) {
	count := 0
	err := json.Unmarshal(data, &count)
	if err != nil {
		if n, err := fmt.Sscanf(string(data), "%d", &count); err != nil || n != 1 {
			return 0, fmt.Errorf("dakika sayacı okunamadı: %v", err)
		}
	}
	return count, nil
}
