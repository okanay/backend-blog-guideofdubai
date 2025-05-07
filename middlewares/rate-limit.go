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

// Ana middleware yapısı
type AIRateLimitMiddleware struct {
	cache *cache.Cache
}

// Middleware oluşturucu
func NewAIRateLimitMiddleware(cache *cache.Cache) *AIRateLimitMiddleware {
	return &AIRateLimitMiddleware{
		cache: cache,
	}
}

// 1. EN ÖNEMLİ FONKSİYON: Gin middleware olarak çalışan ana fonksiyon
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
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_MAX_REQUESTS))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", configs.AI_RATE_LIMIT_MAX_REQUESTS-rateInfo.RequestCount))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

		fmt.Printf("[LOG] Rate limit kontrol sonucu - İzin: %v, Kalan istek: %d, Reset: %s\n",
			allowed, configs.AI_RATE_LIMIT_MAX_REQUESTS-rateInfo.RequestCount, resetTime.Format(time.RFC3339))

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

		requestStart := time.Now()
		fmt.Println("[LOG] İstek başlangıç zamanı:", requestStart.Format(time.RFC3339))

		// Sonraki middleware'lere devam et
		c.Next()

		// İstek başarılı ise (sadece 200 OK durumunda)
		if c.Writer.Status() == http.StatusOK {
			// İstek süresini hesapla
			requestDuration := time.Since(requestStart)
			fmt.Printf("[LOG] İstek süresi: %v\n", requestDuration)

			// Örnek token kullanımı
			tokensUsed := 1000

			// Rate limit bilgisini güncelle
			m.updateRateLimit(userID.String(), tokensUsed)
		}
	}
}

// 2. ÖNEMLİ FONKSİYON: Rate limit kontrolü
func (m *AIRateLimitMiddleware) checkRateLimit(userID string) (*types.RateLimitInfo, bool, time.Time) {
	cacheKey := fmt.Sprintf("ai_rate_limit:%s", userID)
	fmt.Printf("[LOG] checkRateLimit - UserID: %s, Cache Key: %s\n", userID, cacheKey)

	// Cache'den rate limit bilgisini al
	data, exists := m.cache.Get(cacheKey)
	fmt.Printf("[LOG] Cache durumu - Veri var mı: %v\n", exists)

	now := time.Now()
	var rateInfo types.RateLimitInfo

	// Cache'de yoksa veya süresi dolmuşsa yeni bir rate limit kaydı oluştur
	if !exists {
		fmt.Println("[LOG] Cache boş, yeni kayıt oluşturuluyor")
		// Her istek için sıfırlanan versiyon için burada reset yapıyoruz
		rateInfo = types.RateLimitInfo{
			UserID:        userID,
			RequestCount:  0, // Her istekte sıfırlanıyor
			TokensUsed:    0, // Her istekte sıfırlanıyor
			FirstRequest:  now,
			LastRequest:   now,
			WindowResetAt: now.Add(configs.AI_RATE_LIMIT_WINDOW),
		}
	} else {
		fmt.Println("[LOG] Cache'den veri alındı, unmarshal ediliyor")
		// Var olan rate limit bilgisini unmarshal et
		if err := json.Unmarshal(data, &rateInfo); err != nil {
			fmt.Printf("[LOG] Unmarshal HATASI: %v, yeni kayıt oluşturuluyor\n", err)
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
			fmt.Printf("[LOG] Mevcut rate limit bilgisi - RequestCount: %d, TokensUsed: %d, WindowResetAt: %s\n",
				rateInfo.RequestCount, rateInfo.TokensUsed, rateInfo.WindowResetAt.Format(time.RFC3339))

			// Her istekte sıfırlanan versiyon için burada reset yapıyoruz
			// Yorum satırını kaldırarak normal çalışan versiyona dönüştürebilirsiniz
			/*
				// Zaman penceresi süresi dolmuş mu kontrol et
				if now.After(rateInfo.WindowResetAt) {
					fmt.Println("[LOG] Zaman penceresi dolmuş, pencere sıfırlanıyor")
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
			*/
		}
	}

	// Rate limit'i aşmış mı kontrol et
	isAllowed := rateInfo.RequestCount < configs.AI_RATE_LIMIT_MAX_REQUESTS &&
		rateInfo.TokensUsed < configs.AI_RATE_LIMIT_MAX_TOKENS

	fmt.Printf("[LOG] Rate limit kontrol sonucu - İzin: %v, RequestCount: %d, TokensUsed: %d, Limitler: %d istek, %d token\n",
		isAllowed, rateInfo.RequestCount, rateInfo.TokensUsed,
		configs.AI_RATE_LIMIT_MAX_REQUESTS, configs.AI_RATE_LIMIT_MAX_TOKENS)

	return &rateInfo, isAllowed, rateInfo.WindowResetAt
}

// 3. ÖNEMLİ FONKSİYON: Rate limit güncelleme
func (m *AIRateLimitMiddleware) updateRateLimit(userID string, tokensUsed int) {
	cacheKey := fmt.Sprintf("ai_rate_limit:%s", userID)
	fmt.Printf("[LOG] updateRateLimit - UserID: %s, TokensUsed: %d, Cache Key: %s\n", userID, tokensUsed, cacheKey)

	// Cache'den mevcut rate limit bilgisini al
	data, exists := m.cache.Get(cacheKey)
	fmt.Printf("[LOG] Cache durumu - Veri var mı: %v\n", exists)

	now := time.Now()
	var rateInfo types.RateLimitInfo

	if !exists {
		fmt.Println("[LOG] Cache boş, yeni kayıt oluşturuluyor")
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
		fmt.Println("[LOG] Cache'den veri alındı, unmarshal ediliyor")
		// Var olan bilgiyi unmarshal et
		if err := json.Unmarshal(data, &rateInfo); err != nil {
			fmt.Printf("[LOG] Unmarshal HATASI: %v, yeni kayıt oluşturuluyor\n", err)
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
	jsonData, err := json.Marshal(rateInfo)
	if err != nil {
		fmt.Printf("[LOG] Marshal HATASI: %v\n", err)
		return
	}

	// Burada SORUNLU KISIM: Orijinal kodda SetWithTTL kullanımı
	// Bu sorunun kaynağı olabilir. Neden? Cache'e yazarken kullanılan TTL değeri
	fmt.Printf("[LOG] Cache'e yazılıyor - TTL: %v\n", configs.AI_RATE_LIMIT_WINDOW)
	m.cache.SetWithTTL(cacheKey, jsonData, configs.AI_RATE_LIMIT_WINDOW)

	// Cache'e yazıldıktan sonra kontrol et (teşhis için)
	dataAfter, existsAfter := m.cache.Get(cacheKey)
	if existsAfter {
		var checkInfo types.RateLimitInfo
		if err := json.Unmarshal(dataAfter, &checkInfo); err == nil {
			fmt.Printf("[LOG] Cache'e yazıldıktan SONRA kontrol - RequestCount: %d, TokensUsed: %d, WindowResetAt: %s\n",
				checkInfo.RequestCount, checkInfo.TokensUsed, checkInfo.WindowResetAt.Format(time.RFC3339))
		} else {
			fmt.Printf("[LOG] Cache'e yazıldıktan sonra unmarshal HATASI: %v\n", err)
		}
	} else {
		fmt.Println("[LOG] KRİTİK HATA: Cache'e yazıldıktan hemen sonra veri bulunamadı!")
	}
}
