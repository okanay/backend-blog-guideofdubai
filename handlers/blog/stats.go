// handlers/blog/stats-handlers.go
package BlogHandler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/configs"
)

// GetBlogStats tüm blog istatistiklerini getirir
func (h *Handler) GetBlogStats(c *gin.Context) {
	language := c.Query("language")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	stats, total, err := h.BlogRepository.GetBlogStats(language, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "database_error",
			"message": "İstatistikler getirilemedi: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stats":   stats,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// GetBlogStatByID tek bir blog'un istatistiklerini getirir
func (h *Handler) GetBlogStatByID(c *gin.Context) {
	blogIDStr := c.Param("id")
	blogID, err := uuid.Parse(blogIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid_id",
			"message": "Geçersiz blog ID",
		})
		return
	}

	stat, err := h.BlogRepository.GetBlogStatByID(blogID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "database_error",
			"message": "İstatistik getirilemedi: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stat":    stat,
	})
}

func (h *Handler) TrackBlogView(c *gin.Context) {
	// Blog ID'yi URL parametresinden al
	blogIDStr := c.Query("id")
	if blogIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "missing_parameter",
			"message": "id parametresi gereklidir.",
		})
		return
	}

	// Blog ID'yi UUID'ye çevir
	blogID, err := uuid.Parse(blogIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid_parameter",
			"message": "Geçersiz blog ID formatı.",
		})
		return
	}

	// Gerçek IP adresini al
	ip := c.ClientIP()

	// Cache anahtarı oluştur
	cacheKey := fmt.Sprintf("track_view::blog-id:%s:user-ip:%s", blogID.String(), ip)

	// Cache içinde olup olmadığını kontrol et
	if _, exists := h.Cache.Get(cacheKey); exists {
		// Aynı IP adresinden kısa süre içinde tekrar görüntüleme yapılmış
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"tracked": false,
			"reason":  "recently_viewed",
		})
		return
	}

	// Görüntüleme sayısını artır
	go func(id uuid.UUID) {
		if err := h.BlogRepository.IncrementViewCount(id); err != nil {
			// Hata loglama
			fmt.Printf("Görüntüleme sayımı artırılırken hata: %v\n", err)
		}
	}(blogID)

	// IP adresini önbelleğe al (1 dakika TTL)
	h.Cache.SetWithTTL(cacheKey, []byte(cacheKey), configs.VIEW_CACHE_EXPIRATION)

	// İşlem başarılı cevabını döndür
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"tracked": true,
	})
}
