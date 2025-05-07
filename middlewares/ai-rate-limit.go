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
		now := time.Now()
		rateInfo, minuteCount := m.getRateLimitInfo(userID.String())

		// Dakika limitinin yenileneceği zaman
		minuteResetTime := now.Add(1 * time.Minute).Truncate(time.Minute)

		// Kalan süreleri hesapla
		windowResetDuration := rateInfo.WindowResetAt.Sub(now)
		minuteResetDuration := minuteResetTime.Sub(now)

		// Rate limit başlıkları
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_MAX_REQUESTS))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_MAX_REQUESTS-rateInfo.RequestCount))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", rateInfo.WindowResetAt.Unix()))
		c.Header("X-RateLimit-Minute-Limit", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_REQ_PER_MINUTE))
		c.Header("X-RateLimit-Minute-Remaining", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_REQ_PER_MINUTE-minuteCount))

		// Limit kontrolleri
		isMinuteLimitExceeded := minuteCount >= configs.AI_RATE_LIMIT_REQ_PER_MINUTE
		isTotalLimitExceeded := rateInfo.RequestCount >= configs.AI_RATE_LIMIT_MAX_REQUESTS
		isTokenLimitExceeded := rateInfo.TokensUsed >= configs.AI_RATE_LIMIT_MAX_TOKENS

		// Herhangi bir limit aşıldıysa
		if isMinuteLimitExceeded || isTotalLimitExceeded || isTokenLimitExceeded {
			message, limitType, retryAfter := m.generateLimitMessage(
				isMinuteLimitExceeded,
				isTotalLimitExceeded,
				isTokenLimitExceeded,
				windowResetDuration,
				minuteResetDuration,
			)

			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "rate_limit_exceeded",
				"message": message,
				"data": gin.H{
					"limitType":       limitType,
					"totalLimit":      configs.AI_RATE_LIMIT_MAX_REQUESTS,
					"totalRemaining":  configs.AI_RATE_LIMIT_MAX_REQUESTS - rateInfo.RequestCount,
					"minuteLimit":     configs.AI_RATE_LIMIT_REQ_PER_MINUTE,
					"minuteRemaining": configs.AI_RATE_LIMIT_REQ_PER_MINUTE - minuteCount,
					"tokenLimit":      configs.AI_RATE_LIMIT_MAX_TOKENS,
					"tokenRemaining":  configs.AI_RATE_LIMIT_MAX_TOKENS - rateInfo.TokensUsed,
					"windowDuration":  int(configs.AI_RATE_LIMIT_WINDOW.Minutes()),
					"retryAfter":      m.formatDuration(time.Duration(retryAfter) * time.Second),
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

// formatDuration, süreyi insan dostu bir şekilde formatlar (ör. "2 dakika 30 saniye")
func (m *AIRateLimitMiddleware) formatDuration(d time.Duration) string {
	d = d.Round(time.Second)

	hours := d / time.Hour
	d -= hours * time.Hour

	minutes := d / time.Minute
	d -= minutes * time.Minute

	seconds := d / time.Second

	if hours > 0 {
		return fmt.Sprintf("%d saat %d dakika %d saniye", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%d dakika %d saniye", minutes, seconds)
	}
	return fmt.Sprintf("%d saniye", seconds)
}

// generateLimitMessage, limit durumuna göre kullanıcıya gösterilecek mesajı ve ilgili bilgileri oluşturur
func (m *AIRateLimitMiddleware) generateLimitMessage(
	isMinuteLimitExceeded bool,
	isTotalLimitExceeded bool,
	isTokenLimitExceeded bool,
	windowResetDuration time.Duration,
	minuteResetDuration time.Duration,
) (message string, limitType string, retryAfter int) {

	// Süreleri biçimlendir
	windowResetStr := m.formatDuration(windowResetDuration)
	minuteResetStr := m.formatDuration(minuteResetDuration)

	// Yalnızca dakika limiti aşılmış
	if isMinuteLimitExceeded && !isTotalLimitExceeded && !isTokenLimitExceeded {
		return fmt.Sprintf(
			"Dakika limiti aşıldı. Dakika başına en fazla %d istek yapabilirsiniz. "+
				"Limitiniz %s sonra yenilenecek.",
			configs.AI_RATE_LIMIT_REQ_PER_MINUTE,
			minuteResetStr,
		), "minute", int(minuteResetDuration.Seconds())
	}

	// Yalnızca toplam limit aşılmış
	if isTotalLimitExceeded && !isMinuteLimitExceeded {
		return fmt.Sprintf(
			"Toplam limit aşıldı. %d dakikalık süre içinde en fazla %d istek yapabilirsiniz. "+
				"Limitiniz %s sonra yenilenecek.",
			int(configs.AI_RATE_LIMIT_WINDOW.Minutes()),
			configs.AI_RATE_LIMIT_MAX_REQUESTS,
			windowResetStr,
		), "total", int(windowResetDuration.Seconds())
	}

	// Hem dakika hem toplam limit aşılmış
	if isMinuteLimitExceeded && isTotalLimitExceeded {
		// Hangi limit daha uzun sürecek?
		minuteRetry := int(minuteResetDuration.Seconds())
		totalRetry := int(windowResetDuration.Seconds())

		if totalRetry > minuteRetry {
			return fmt.Sprintf(
				"Hem dakika hem toplam limit aşıldı. "+
					"Dakika limiti (%d istek/dk) %s sonra, "+
					"toplam limit (%d istek/%d dk) ise %s sonra yenilenecek.",
				configs.AI_RATE_LIMIT_REQ_PER_MINUTE,
				minuteResetStr,
				configs.AI_RATE_LIMIT_MAX_REQUESTS,
				int(configs.AI_RATE_LIMIT_WINDOW.Minutes()),
				windowResetStr,
			), "both", totalRetry
		} else {
			return fmt.Sprintf(
				"Hem dakika hem toplam limit aşıldı. "+
					"Dakika başına %d istek yapabilirsiniz. "+
					"Limitiniz %s sonra yenilenecek.",
				configs.AI_RATE_LIMIT_REQ_PER_MINUTE,
				minuteResetStr,
			), "both", minuteRetry
		}
	}

	// Yalnızca token limiti aşılmış
	if isTokenLimitExceeded {
		return fmt.Sprintf(
			"Token limiti aşıldı. %d dakikalık süre içinde en fazla %d token kullanabilirsiniz. "+
				"Token limitiniz %s sonra yenilenecek.",
			int(configs.AI_RATE_LIMIT_WINDOW.Minutes()),
			configs.AI_RATE_LIMIT_MAX_TOKENS,
			windowResetStr,
		), "token", int(windowResetDuration.Seconds())
	}

	// Varsayılan mesaj
	return "Rate limit aşıldı. Lütfen daha sonra tekrar deneyiniz.", "unknown", 60
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
