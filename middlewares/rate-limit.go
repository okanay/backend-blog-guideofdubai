package middlewares

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	cache "github.com/okanay/backend-blog-guideofdubai/services"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

// Sabit değerler
const (
	MAX_REQUESTS_PER_MINUTE = 5               // Dakikada maximum 5 istek
	RATE_LIMIT_WINDOW       = 1 * time.Minute // 1 dakikalık pencere
	MAX_TOKENS              = 10000           // Token limiti (isteğe bağlı)
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
		fmt.Println("[LOG] Rate limit middleware başladı")

		// User bilgisini al
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			fmt.Println("[LOG] Kullanıcı kimliği bulunamadı")
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
			fmt.Println("[LOG] Kullanıcı kimliği geçersiz formatta")
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "internal_error",
				"message": "Kullanıcı kimliği geçersiz formatta",
			})
			c.Abort()
			return
		}

		fmt.Printf("[LOG] Kullanıcı ID: %s\n", userID.String())

		// Rate limit bilgisini kontrol et
		rateInfo, allowed, resetTime := m.checkRateLimit(userID.String())

		// Rate limit başlıklarını ekle
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", MAX_REQUESTS_PER_MINUTE))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", MAX_REQUESTS_PER_MINUTE-rateInfo.RequestCount))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

		fmt.Printf("[LOG] Rate limit kontrol sonucu - İzin: %v, Kalan istek: %d, Reset: %s\n",
			allowed, MAX_REQUESTS_PER_MINUTE-rateInfo.RequestCount, resetTime.Format(time.RFC3339))

		if !allowed {
			retryAfter := int(time.Until(resetTime).Seconds())
			if retryAfter < 0 {
				retryAfter = 0
			}

			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			fmt.Printf("[LOG] Rate limit aşıldı - Retry-After: %d saniye\n", retryAfter)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "rate_limit_exceeded",
				"message": "AI istek limiti aşıldı. Lütfen daha sonra tekrar deneyin.",
				"data": gin.H{
					"resetAt":    resetTime,
					"retryAfter": retryAfter,
				},
			})
			c.Abort()
			return
		}

		// İstek sayacını hemen artır (önden artırma)
		m.incrementRequestCount(userID.String(), rateInfo)

		requestStart := time.Now()
		fmt.Println("[LOG] İstek başlangıç zamanı:", requestStart.Format(time.RFC3339))

		// Sonraki middleware'lere devam et
		c.Next()

		// İstek başarılı ise (sadece 200 OK durumunda) token kullanımını güncelle
		// Başarılı olmadıysa, eklenen istek sayısını geri alabilirsiniz (opsiyonel)
		if c.Writer.Status() == http.StatusOK {
			// İstek süresini hesapla
			requestDuration := time.Since(requestStart)
			fmt.Printf("[LOG] İstek süresi: %v\n", requestDuration)

			// Token kullanımını güncelle (opsiyonel)
			tokensUsed := 1000
			m.updateTokenUsage(userID.String(), tokensUsed)
		} else if c.Writer.Status() != http.StatusOK && c.Writer.Status() != http.StatusTooManyRequests {
			// İstek başarısız olduğunda sayacı geri al (opsiyonel)
			// m.decrementRequestCount(userID.String())
		}
	}
}

// Rate limit kontrolü yapar
func (m *AIRateLimitMiddleware) checkRateLimit(userID string) (*types.RateLimitInfo, bool, time.Time) {
	cacheKey := fmt.Sprintf("ai_rate_limit:%s", userID)
	fmt.Printf("[LOG] checkRateLimit - UserID: %s, Cache Key: %s\n", userID, cacheKey)

	now := time.Now()
	var rateInfo types.RateLimitInfo
	var windowResetAt time.Time

	// Cache'den rate limit bilgisini al
	data, exists := m.cache.Get(cacheKey)
	fmt.Printf("[LOG] Cache durumu - Veri var mı: %v\n", exists)

	if !exists {
		fmt.Println("[LOG] Cache boş, yeni kayıt oluşturuluyor")
		// Yeni bir 1 dakikalık pencere başlat
		windowResetAt = now.Add(RATE_LIMIT_WINDOW)

		rateInfo = types.RateLimitInfo{
			UserID:        userID,
			RequestCount:  0,
			TokensUsed:    0,
			FirstRequest:  now,
			LastRequest:   now,
			WindowResetAt: windowResetAt,
		}
	} else {
		fmt.Println("[LOG] Cache'den veri alındı, unmarshal ediliyor")
		// Var olan rate limit bilgisini unmarshal et
		if err := json.Unmarshal(data, &rateInfo); err != nil {
			fmt.Printf("[LOG] Unmarshal HATASI: %v, yeni kayıt oluşturuluyor\n", err)
			// Hata durumunda yeni bir kayıt oluştur
			windowResetAt = now.Add(RATE_LIMIT_WINDOW)

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
			fmt.Printf("[LOG] Mevcut rate limit bilgisi - RequestCount: %d, TokensUsed: %d, WindowResetAt: %s\n",
				rateInfo.RequestCount, rateInfo.TokensUsed, windowResetAt.Format(time.RFC3339))

			// Zaman penceresi süresi dolmuş mu kontrol et
			if now.After(windowResetAt) {
				fmt.Println("[LOG] Zaman penceresi dolmuş, pencere sıfırlanıyor")
				// Süresi dolmuşsa, pencereyi sıfırla
				windowResetAt = now.Add(RATE_LIMIT_WINDOW)

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

	// Rate limit'i aşmış mı kontrol et
	isAllowed := rateInfo.RequestCount < MAX_REQUESTS_PER_MINUTE

	fmt.Printf("[LOG] Rate limit kontrol sonucu - İzin: %v, RequestCount: %d, Limit: %d istek\n",
		isAllowed, rateInfo.RequestCount, MAX_REQUESTS_PER_MINUTE)

	return &rateInfo, isAllowed, windowResetAt
}

// İstek sayacını artırır (önden artırma)
func (m *AIRateLimitMiddleware) incrementRequestCount(userID string, currentInfo *types.RateLimitInfo) {
	cacheKey := fmt.Sprintf("ai_rate_limit:%s", userID)

	now := time.Now()

	// İstek sayacını artır
	currentInfo.RequestCount++
	currentInfo.LastRequest = now

	fmt.Printf("[LOG] İstek sayacı artırıldı - Yeni değer: %d\n", currentInfo.RequestCount)

	// Güncellenmiş bilgiyi marshal et
	jsonData, err := json.Marshal(currentInfo)
	if err != nil {
		fmt.Printf("[LOG] Marshal HATASI: %v\n", err)
		return
	}

	// Kalan süreyi hesapla
	remainingTime := currentInfo.WindowResetAt.Sub(now)
	if remainingTime <= 0 {
		remainingTime = RATE_LIMIT_WINDOW
	}

	fmt.Printf("[LOG] Cache'e yazılıyor - Kalan süre: %v\n", remainingTime)

	// Cache'e yaz (TTL olarak kalan süreyi kullan)
	m.cache.SetWithTTL(cacheKey, jsonData, remainingTime)

	// Cache'e yazıldıktan sonra kontrol et
	dataAfter, existsAfter := m.cache.Get(cacheKey)
	if existsAfter {
		var checkInfo types.RateLimitInfo
		if err := json.Unmarshal(dataAfter, &checkInfo); err == nil {
			fmt.Printf("[LOG] Cache'e yazıldıktan SONRA kontrol - RequestCount: %d, WindowResetAt: %s\n",
				checkInfo.RequestCount, checkInfo.WindowResetAt.Format(time.RFC3339))
		}
	}
}

// Token kullanımını günceller (opsiyonel)
func (m *AIRateLimitMiddleware) updateTokenUsage(userID string, tokensUsed int) {
	cacheKey := fmt.Sprintf("ai_rate_limit:%s", userID)
	fmt.Printf("[LOG] updateTokenUsage - UserID: %s, TokensUsed: %d\n", userID, tokensUsed)

	// Cache'den mevcut rate limit bilgisini al
	data, exists := m.cache.Get(cacheKey)
	if !exists {
		fmt.Println("[LOG] Cache bilgisi bulunamadı, token güncellemesi yapılamıyor")
		return
	}

	var rateInfo types.RateLimitInfo
	if err := json.Unmarshal(data, &rateInfo); err != nil {
		fmt.Printf("[LOG] Unmarshal HATASI: %v, token güncellemesi yapılamıyor\n", err)
		return
	}

	now := time.Now()

	// Token kullanımını güncelle
	oldTokens := rateInfo.TokensUsed
	rateInfo.TokensUsed += tokensUsed

	fmt.Printf("[LOG] Token güncellendi - Eski: %d, Yeni: %d\n", oldTokens, rateInfo.TokensUsed)

	// Güncellenmiş bilgiyi marshal et
	jsonData, err := json.Marshal(rateInfo)
	if err != nil {
		fmt.Printf("[LOG] Marshal HATASI: %v\n", err)
		return
	}

	// Kalan süreyi hesapla
	remainingTime := rateInfo.WindowResetAt.Sub(now)
	if remainingTime <= 0 {
		remainingTime = RATE_LIMIT_WINDOW
	}

	// Cache'e yaz (TTL olarak kalan süreyi kullan)
	m.cache.SetWithTTL(cacheKey, jsonData, remainingTime)
}

// İstek sayacını azaltır (opsiyonel, başarısız istekler için)
func (m *AIRateLimitMiddleware) decrementRequestCount(userID string) {
	cacheKey := fmt.Sprintf("ai_rate_limit:%s", userID)

	// Cache'den mevcut rate limit bilgisini al
	data, exists := m.cache.Get(cacheKey)
	if !exists {
		return
	}

	var rateInfo types.RateLimitInfo
	if err := json.Unmarshal(data, &rateInfo); err != nil {
		return
	}

	now := time.Now()

	// İstek sayacını azalt (negatif olmamak şartıyla)
	if rateInfo.RequestCount > 0 {
		rateInfo.RequestCount--
	}

	// Güncellenmiş bilgiyi marshal et
	jsonData, _ := json.Marshal(rateInfo)

	// Kalan süreyi hesapla
	remainingTime := rateInfo.WindowResetAt.Sub(now)
	if remainingTime <= 0 {
		remainingTime = RATE_LIMIT_WINDOW
	}

	// Cache'e yaz (TTL olarak kalan süreyi kullan)
	m.cache.SetWithTTL(cacheKey, jsonData, remainingTime)
}
