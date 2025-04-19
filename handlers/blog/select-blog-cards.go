package BlogHandler

import (
	"errors"
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
	options, err := parseQueryParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// En az bir filtre gerekli kontrolü
	if !options.HasFilter() {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "En az bir filtre gereklidir (id, ids, category, tag, language, featured)",
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

// Query parametrelerini parse eden yardımcı fonksiyon
func parseQueryParams(c *gin.Context) (types.BlogCardQueryOptions, error) {
	var options types.BlogCardQueryOptions

	// ID parametresi
	if idStr := c.Query("id"); idStr != "" {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return options, errors.New("Geçersiz ID formatı")
		}
		options.ID = id
	}

	// Çoklu ID'ler (virgülle ayrılmış)
	if idsStr := c.Query("ids"); idsStr != "" {
		idStrings := strings.Split(idsStr, ",")
		for _, idStr := range idStrings {
			id, err := uuid.Parse(strings.TrimSpace(idStr))
			if err != nil {
				return options, fmt.Errorf("Geçersiz ID formatı: %s", idStr)
			}
			options.IDs = append(options.IDs, id)
		}
	}

	// Diğer filtre parametreleri
	if category := c.Query("category"); category != "" {
		options.CategoryValue = category
	}

	if tag := c.Query("tag"); tag != "" {
		options.TagValue = tag
	}

	if language := c.Query("language"); language != "" {
		options.Language = language
	}

	if featured := c.Query("featured"); featured == "true" {
		options.Featured = true
	}

	if status := c.Query("status"); status != "" {
		options.Status = types.BlogStatus(status)
	}

	// Sayfalama parametreleri
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			return options, errors.New("Geçersiz limit değeri")
		}
		options.Limit = limit
	} else {
		// Varsayılan limit
		options.Limit = 10
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return options, errors.New("Geçersiz offset değeri")
		}
		options.Offset = offset
	}

	// Tarih filtresi parametreleri
	err := parseDateParams(c, &options)
	if err != nil {
		return options, err
	}

	// Sıralama parametreleri
	if sortBy := c.Query("sort_by"); sortBy != "" {
		options.SortBy = sortBy
	}

	if sortDir := c.Query("sort_dir"); sortDir != "" {
		if sortDir == "asc" {
			options.SortDirection = types.SortAsc
		} else if sortDir == "desc" {
			options.SortDirection = types.SortDesc
		} else {
			return options, errors.New("Geçersiz sort_dir değeri. 'asc' veya 'desc' kullanın")
		}
	} else {
		options.SortDirection = types.SortDesc // Varsayılan olarak en yeniden eskiye
	}

	return options, nil
}

// Tarih ile ilgili parametreleri parse eden yardımcı fonksiyon
func parseDateParams(c *gin.Context, options *types.BlogCardQueryOptions) error {
	// Başlangıç tarihi
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return errors.New("Geçersiz start_date formatı. YYYY-MM-DD kullanın")
		}
		options.StartDate = &startDate
	}

	// Bitiş tarihi
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return errors.New("Geçersiz end_date formatı. YYYY-MM-DD kullanın")
		}
		// Günün sonuna kadar
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		options.EndDate = &endDate
	}

	return nil
}
