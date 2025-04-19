package BlogHandler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

func (h *Handler) SelectBlogCards(c *gin.Context) {
	var options types.BlogCardQueryOptions
	var hasFilter bool

	// Query parametrelerini işle
	if idStr := c.Query("id"); idStr != "" {
		id, err := uuid.Parse(idStr)
		if err == nil {
			options.ID = id
			hasFilter = true
		} else {
			// Geçersiz ID formatı
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid ID format",
			})
			return
		}
	}

	// Çoklu ID'ler (virgülle ayrılmış olarak)
	if idsStr := c.Query("ids"); idsStr != "" {
		idStrings := strings.Split(idsStr, ",")
		for _, idStr := range idStrings {
			if id, err := uuid.Parse(strings.TrimSpace(idStr)); err == nil {
				options.IDs = append(options.IDs, id)
			} else {
				// Geçersiz ID formatı
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Invalid ID format: %s", idStr),
				})
				return
			}
		}
		if len(options.IDs) > 0 {
			hasFilter = true
		}
	}

	// Kategori parametresi
	if category := c.Query("category"); category != "" {
		options.CategoryValue = category
		hasFilter = true
	}

	// Etiket parametresi
	if tag := c.Query("tag"); tag != "" {
		options.TagValue = tag
		hasFilter = true
	}

	// Dil parametresi
	if language := c.Query("language"); language != "" {
		options.Language = language
		hasFilter = true
	}

	// Öne çıkanlar parametresi
	if featured := c.Query("featured"); featured == "true" {
		options.Featured = true
		hasFilter = true
	}

	// Status parametresi
	if status := c.Query("status"); status != "" {
		options.Status = types.BlogStatus(status)
		hasFilter = true
	}

	// Sayfalama parametreleri
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			options.Limit = limit
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid limit value",
			})
			return
		}
	} else {
		// Varsayılan limit
		options.Limit = 10
	}

	// Başlangıç tarihi
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			options.StartDate = &startDate
			hasFilter = true
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid start_date format. Use YYYY-MM-DD",
			})
			return
		}
	}

	// Bitiş tarihi
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// Günün sonuna kadar
			endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			options.EndDate = &endDate
			hasFilter = true
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid end_date format. Use YYYY-MM-DD",
			})
			return
		}
	}

	// Son X gün
	if lastDaysStr := c.Query("last_days"); lastDaysStr != "" {
		if lastDays, err := strconv.Atoi(lastDaysStr); err == nil && lastDays > 0 {
			options.LastDays = lastDays
			hasFilter = true
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid last_days value",
			})
			return
		}
	}

	// Son X hafta
	if lastWeeksStr := c.Query("last_weeks"); lastWeeksStr != "" {
		if lastWeeks, err := strconv.Atoi(lastWeeksStr); err == nil && lastWeeks > 0 {
			options.LastWeeks = lastWeeks
			hasFilter = true
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid last_weeks value",
			})
			return
		}
	}

	// Son X ay
	if lastMonthsStr := c.Query("last_months"); lastMonthsStr != "" {
		if lastMonths, err := strconv.Atoi(lastMonthsStr); err == nil && lastMonths > 0 {
			options.LastMonths = lastMonths
			hasFilter = true
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid last_months value",
			})
			return
		}
	}

	// Sıralama alanı
	if sortBy := c.Query("sort_by"); sortBy != "" {
		options.SortBy = sortBy
	}

	// Sıralama yönü
	if sortDir := c.Query("sort_dir"); sortDir != "" {
		if sortDir == "asc" {
			options.SortDirection = types.SortAsc
		} else if sortDir == "desc" {
			options.SortDirection = types.SortDesc
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid sort_dir value. Use 'asc' or 'desc'",
			})
			return
		}
	} else {
		options.SortDirection = types.SortDesc // Varsayılan olarak en yeniden eskiye
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			options.Offset = offset
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid offset value",
			})
			return
		}
	}

	// En az bir filtre gerekli
	if !hasFilter {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "At least one filter is required (id, ids, category, tag, featured)",
		})
		return
	}

	// Blog kartlarını getir
	blogCards, err := h.BlogRepository.SelectBlogCards(options)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"blogs":   blogCards,
		"count":   len(blogCards),
	})
}
