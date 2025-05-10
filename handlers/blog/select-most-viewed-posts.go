// handlers/blog/select-most-viewed-posts.go
package BlogHandler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) SelectMostViewedPosts(c *gin.Context) {
	// Query parametrelerini al
	language := c.DefaultQuery("language", "")
	limitStr := c.DefaultQuery("limit", "10")
	period := c.DefaultQuery("period", "all") // all, day, week, month, year

	// Limit değerini dönüştür
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50 // Maksimum limit
	}

	// Period parametresini doğrula
	validPeriods := map[string]bool{"all": true, "day": true, "week": true, "month": true, "year": true}
	if !validPeriods[period] {
		period = "all"
	}

	// Cache'den kontrol et
	blogs, exists := h.BlogCache.GetMostViewedPosts(language, period)
	if exists {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"blogs":   blogs,
			"count":   len(blogs),
			"period":  period,
			"cached":  true,
		})
		return
	}

	// Veritabanından getir
	blogs, err = h.BlogRepository.SelectMostViewedPosts(language, limit, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "database_error",
			"message": "Blog yazıları getirilirken bir hata oluştu: " + err.Error(),
		})
		return
	}

	// Cache'e kaydet
	h.BlogCache.SaveMostViewedPosts(language, period, blogs)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"blogs":   blogs,
		"count":   len(blogs),
		"period":  period,
		"cached":  false,
	})
}
