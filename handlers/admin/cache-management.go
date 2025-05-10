// handlers/admin/cache-management.go
package AdminHandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-blog-guideofdubai/configs"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

// GetCacheStats tüm cache istatistiklerini gösterir
func (h *Handler) GetCacheStats(c *gin.Context) {
	stats := h.Cache.GetStats()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stats":   stats,
	})
}

// ClearAllCache belirli önekleri koruyarak tüm cache'i temizler
func (h *Handler) ClearAllCache(c *gin.Context) {
	// Korunacak önekleri al (varsayılan olarak AI rate limitleri korunur)
	protectedPrefixes := []string{"ai_rate_limit:", "ai_rate_limit_minute:"}

	// URL parametresi ile ek önekler belirtilebilir
	if additionalProtected := c.Query("protect"); additionalProtected != "" {
		for _, prefix := range strings.Split(additionalProtected, ",") {
			protectedPrefixes = append(protectedPrefixes, strings.TrimSpace(prefix))
		}
	}

	// Blog cache'ini temizle, korunan önekleri atla
	h.Cache.ClearExceptPrefixes(protectedPrefixes)

	c.JSON(http.StatusOK, gin.H{
		"success":            true,
		"message":            "Cache başarıyla temizlendi (korunan önekler hariç)",
		"protected_prefixes": protectedPrefixes,
	})
}

// ClearCacheWithPrefix belirtilen öneke sahip cache'leri temizler
func (h *Handler) ClearCacheWithPrefix(c *gin.Context) {
	prefix := c.Query("prefix")
	if prefix == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "missing_parameter",
			"message": "prefix parametresi gereklidir",
		})
		return
	}

	// Belirtilen önekle başlayan tüm cache anahtarlarını temizle
	h.Cache.ClearPrefix(prefix)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("'%s' önekine sahip cache anahtarları temizlendi", prefix),
	})
}

// GetCacheItems belirli bir önekteki tüm cache öğelerini listeler
func (h *Handler) GetCacheItems(c *gin.Context) {
	prefix := c.DefaultQuery("prefix", "")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	// Çok fazla veri dönmemesi için limit kontrolü
	if limit > 500 {
		limit = 500
	}

	// Tüm cache öğelerini getir
	allItems := h.Cache.GetAllWithPrefix(prefix)

	totalCount := len(allItems)

	// Sonuçları limit ile sınırla
	items := make(map[string]any)
	count := 0

	for key, value := range allItems {
		if count >= limit {
			break
		}

		// JSON olarak parse etmeyi dene
		var parsedValue any
		if err := json.Unmarshal(value, &parsedValue); err == nil {
			items[key] = parsedValue
		} else {
			// JSON parse edemiyorsak string olarak göster
			items[key] = string(value)
		}

		count++
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"items":   items,
		"count":   count,
		"total":   totalCount,
		"prefix":  prefix,
		"limit":   limit,
	})
}

// ClearAIRateLimits AI rate limit cache'lerini temizler
func (h *Handler) ClearAIRateLimits(c *gin.Context) {
	// Sadece AI rate limit cache'lerini temizle
	h.Cache.ClearAIRateLimits()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "AI rate limit cache'leri başarıyla temizlendi",
	})
}

// GetAIRateLimits tüm AI rate limit kayıtlarını gösterir
func (h *Handler) GetAIRateLimits(c *gin.Context) {
	// Tüm rate limit kayıtlarını al
	rateLimits := []map[string]any{}

	// Cache'den "ai_rate_limit:" önekine sahip tüm anahtarları ara
	allRateLimits := h.Cache.GetAllWithPrefix("ai_rate_limit:")

	for key, data := range allRateLimits {
		var rateInfo types.RateLimitInfo
		if err := json.Unmarshal(data, &rateInfo); err == nil {
			// Kalan zamanı hesapla
			now := time.Now()
			remainingTime := rateInfo.WindowResetAt.Sub(now)

			rateLimits = append(rateLimits, map[string]any{
				"cacheKey":        key,
				"userId":          rateInfo.UserID,
				"requestCount":    rateInfo.RequestCount,
				"tokensUsed":      rateInfo.TokensUsed,
				"firstRequest":    rateInfo.FirstRequest,
				"lastRequest":     rateInfo.LastRequest,
				"windowResetAt":   rateInfo.WindowResetAt,
				"requestsPerMin":  rateInfo.RequestsPerMin,
				"minuteStartedAt": rateInfo.MinuteStartedAt,
				"remaining": gin.H{
					"requests":    configs.AI_RATE_LIMIT_MAX_REQUESTS - rateInfo.RequestCount,
					"tokens":      configs.AI_RATE_LIMIT_MAX_TOKENS - rateInfo.TokensUsed,
					"timeSeconds": int(remainingTime.Seconds()),
					"timeHuman":   formatDuration(remainingTime),
				},
				"limits": gin.H{
					"maxRequests":       configs.AI_RATE_LIMIT_MAX_REQUESTS,
					"maxTokens":         configs.AI_RATE_LIMIT_MAX_TOKENS,
					"windowDuration":    configs.AI_RATE_LIMIT_WINDOW.String(),
					"reqPerMinute":      configs.AI_RATE_LIMIT_REQ_PER_MINUTE,
					"windowDurationMin": int(configs.AI_RATE_LIMIT_WINDOW.Minutes()),
				},
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rateLimits,
		"count":   len(rateLimits),
	})
}

// ResetUserRateLimit belirli bir kullanıcının rate limit bilgilerini sıfırlar
func (h *Handler) ResetUserRateLimit(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "missing_parameter",
			"message": "userId parametresi gereklidir",
		})
		return
	}

	// Kullanıcının rate limit anahtarlarını temizle
	rateKey := fmt.Sprintf("ai_rate_limit:%s", userID)
	minuteKey := fmt.Sprintf("ai_rate_limit_minute:%s", userID)

	h.Cache.Delete(rateKey)
	h.Cache.Delete(minuteKey)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Kullanıcı %s için AI rate limit bilgileri sıfırlandı", userID),
	})
}

// formatDuration süreyi insan dostu bir formata çevirir
func formatDuration(d time.Duration) string {
	if d < 0 {
		return "Süresi dolmuş"
	}

	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%d saat %d dakika", h, m)
	} else if m > 0 {
		return fmt.Sprintf("%d dakika %d saniye", m, s)
	}
	return fmt.Sprintf("%d saniye", s)
}
