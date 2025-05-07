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
		rateInfo, allowed, resetTime, minuteLimit := m.checkRateLimit(userID.String())

		// Rate limit başlıklarını ekle - ConfigUREABLE_RATE_LIMIT_MAX_REQUESTS tam limitini göster
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_MAX_REQUESTS))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_MAX_REQUESTS-rateInfo.RequestCount))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))
		// Dakika limitini ekle
		c.Header("X-RateLimit-Minute-Limit", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_REQ_PER_MINUTE))
		c.Header("X-RateLimit-Minute-Remaining", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_REQ_PER_MINUTE-minuteLimit))

		fmt.Printf("[LOG] Rate limit kontrol sonucu - İzin: %v, Toplam Kalan: %d/%d, Dakika Kalan: %d/%d, Reset: %s\n",
			allowed,
			configs.AI_RATE_LIMIT_MAX_REQUESTS-rateInfo.RequestCount, configs.AI_RATE_LIMIT_MAX_REQUESTS,
			configs.AI_RATE_LIMIT_REQ_PER_MINUTE-minuteLimit, configs.AI_RATE_LIMIT_REQ_PER_MINUTE,
			resetTime.Format(time.RFC3339))

		if !allowed {
			retryAfter := int(time.Until(resetTime).Seconds())
			retryAfter = max(retryAfter, 0)

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
		if c.Writer.Status() == http.StatusOK {
			// İstek süresini hesapla
			requestDuration := time.Since(requestStart)
			fmt.Printf("[LOG] İstek süresi: %v\n", requestDuration)

			// Token kullanımını güncelle (gerçek API yanıtından alınmalı)
			tokensUsed := 1000 // Örnek değer, gerçekte OpenAI yanıtından hesaplanmalı
			m.updateTokenUsage(userID.String(), tokensUsed)
		}
	}
}

// Rate limit kontrolü yapar
func (m *AIRateLimitMiddleware) checkRateLimit(userID string) (*types.RateLimitInfo, bool, time.Time, int) {
	cacheKey := fmt.Sprintf("ai_rate_limit:%s", userID)
	minuteKey := fmt.Sprintf("ai_rate_limit_minute:%s", userID)

	fmt.Printf("[LOG] checkRateLimit - UserID: %s, Cache Key: %s\n", userID, cacheKey)

	now := time.Now()
	var rateInfo types.RateLimitInfo
	var windowResetAt time.Time

	// Dakikalık istek sayısını kontrol et (1 dakikalık pencere)
	minuteCount := 0
	minuteData, minuteExists := m.cache.Get(minuteKey)

	if minuteExists {
		// Dakikalık sayacı oku
		if count, err := parseMinuteCount(minuteData); err == nil {
			minuteCount = count
			fmt.Printf("[LOG] Dakikalık istek sayısı: %d/%d\n", minuteCount, configs.AI_RATE_LIMIT_REQ_PER_MINUTE)
		}
	}

	// Toplam istek sayısını kontrol et (10 dakikalık pencere)
	data, exists := m.cache.Get(cacheKey)
	fmt.Printf("[LOG] Cache durumu - Veri var mı: %v\n", exists)

	if !exists {
		fmt.Println("[LOG] Cache boş, yeni kayıt oluşturuluyor")
		// Yeni bir pencere başlat
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
		fmt.Println("[LOG] Cache'den veri alındı, unmarshal ediliyor")
		// Var olan rate limit bilgisini unmarshal et
		if err := json.Unmarshal(data, &rateInfo); err != nil {
			fmt.Printf("[LOG] Unmarshal HATASI: %v, yeni kayıt oluşturuluyor\n", err)
			// Hata durumunda yeni bir kayıt oluştur
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
			fmt.Printf("[LOG] Mevcut rate limit bilgisi - RequestCount: %d, TokensUsed: %d, WindowResetAt: %s\n",
				rateInfo.RequestCount, rateInfo.TokensUsed, windowResetAt.Format(time.RFC3339))

			// Zaman penceresi süresi dolmuş mu kontrol et
			if now.After(windowResetAt) {
				fmt.Println("[LOG] Zaman penceresi dolmuş, pencere sıfırlanıyor")
				// Süresi dolmuşsa, pencereyi sıfırla
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

	// Rate limit'i aşmış mı kontrol et (toplam limit VE dakikalık limit)
	isAllowed := rateInfo.RequestCount < configs.AI_RATE_LIMIT_MAX_REQUESTS &&
		minuteCount < configs.AI_RATE_LIMIT_REQ_PER_MINUTE &&
		rateInfo.TokensUsed < configs.AI_RATE_LIMIT_MAX_TOKENS

	fmt.Printf("[LOG] Rate limit kontrol sonucu - İzin: %v, RequestCount: %d/%d, MinuteCount: %d/%d, TokensUsed: %d/%d\n",
		isAllowed,
		rateInfo.RequestCount, configs.AI_RATE_LIMIT_MAX_REQUESTS,
		minuteCount, configs.AI_RATE_LIMIT_REQ_PER_MINUTE,
		rateInfo.TokensUsed, configs.AI_RATE_LIMIT_MAX_TOKENS)

	return &rateInfo, isAllowed, windowResetAt, minuteCount
}

// İstek sayacını artırır (önden artırma)
func (m *AIRateLimitMiddleware) incrementRequestCount(userID string, currentInfo *types.RateLimitInfo) {
	cacheKey := fmt.Sprintf("ai_rate_limit:%s", userID)
	minuteKey := fmt.Sprintf("ai_rate_limit_minute:%s", userID)

	now := time.Now()

	// 1. Ana istek sayacını artır (10 dakikalık pencere)
	currentInfo.RequestCount++
	currentInfo.LastRequest = now

	fmt.Printf("[LOG] Toplam istek sayacı artırıldı - Yeni değer: %d/%d\n",
		currentInfo.RequestCount, configs.AI_RATE_LIMIT_MAX_REQUESTS)

	// Güncellenmiş bilgiyi marshal et
	jsonData, err := json.Marshal(currentInfo)
	if err != nil {
		fmt.Printf("[LOG] Marshal HATASI: %v\n", err)
		return
	}

	// Kalan süreyi hesapla
	remainingTime := currentInfo.WindowResetAt.Sub(now)
	if remainingTime <= 0 {
		remainingTime = configs.AI_RATE_LIMIT_WINDOW
	}

	fmt.Printf("[LOG] Cache'e yazılıyor - Kalan süre: %v\n", remainingTime)

	// Cache'e yaz (TTL olarak kalan süreyi kullan)
	m.cache.SetWithTTL(cacheKey, jsonData, remainingTime)

	// 2. Dakikalık istek sayacını artır (1 dakikalık pencere)
	minuteCount := 1
	minuteData, minuteExists := m.cache.Get(minuteKey)

	if minuteExists {
		if count, err := parseMinuteCount(minuteData); err == nil {
			minuteCount = count + 1
		}
	}

	// Dakikalık sayacı kaydet
	countData := fmt.Sprintf("%d", minuteCount)
	m.cache.SetWithTTL(minuteKey, []byte(countData), 1*time.Minute) // Dakikalık pencere

	fmt.Printf("[LOG] Dakikalık istek sayacı artırıldı - Yeni değer: %d/%d\n",
		minuteCount, configs.AI_RATE_LIMIT_REQ_PER_MINUTE)

	// Cache kontrolü
	dataAfter, existsAfter := m.cache.Get(cacheKey)
	if existsAfter {
		var checkInfo types.RateLimitInfo
		if err := json.Unmarshal(dataAfter, &checkInfo); err == nil {
			fmt.Printf("[LOG] Cache'e yazıldıktan SONRA kontrol - RequestCount: %d, WindowResetAt: %s\n",
				checkInfo.RequestCount, checkInfo.WindowResetAt.Format(time.RFC3339))
		}
	}
}

// Token kullanımını günceller
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

	fmt.Printf("[LOG] Token güncellendi - Eski: %d, Yeni: %d/%d\n",
		oldTokens, rateInfo.TokensUsed, configs.AI_RATE_LIMIT_MAX_TOKENS)

	// Güncellenmiş bilgiyi marshal et
	jsonData, err := json.Marshal(rateInfo)
	if err != nil {
		fmt.Printf("[LOG] Marshal HATASI: %v\n", err)
		return
	}

	// Kalan süreyi hesapla
	remainingTime := rateInfo.WindowResetAt.Sub(now)
	if remainingTime <= 0 {
		remainingTime = configs.AI_RATE_LIMIT_WINDOW
	}

	// Cache'e yaz (TTL olarak kalan süreyi kullan)
	m.cache.SetWithTTL(cacheKey, jsonData, remainingTime)
}

// Dakikalık istek sayacını parse etme yardımcı fonksiyonu
func parseMinuteCount(data []byte) (int, error) {
	count := 0
	err := json.Unmarshal(data, &count)
	if err != nil {
		// JSON değilse, string olarak dene
		if n, err := fmt.Sscanf(string(data), "%d", &count); err != nil || n != 1 {
			return 0, fmt.Errorf("dakika sayacı okunamadı: %v", err)
		}
	}
	return count, nil
}
