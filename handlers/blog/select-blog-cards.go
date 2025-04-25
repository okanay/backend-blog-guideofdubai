package BlogHandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

func (h *Handler) SelectBlogCards(c *gin.Context) {
	// Query parametrelerini al
	queryOptions := types.BlogCardQueryOptions{
		Limit:  10,
		Offset: 0,
	}

	// ID
	if idStr := c.Query("id"); idStr != "" {
		id, err := uuid.Parse(idStr)
		if err == nil {
			queryOptions.ID = id
		}
	}

	// Title
	if title := c.Query("title"); title != "" {
		queryOptions.Title = title
	}

	// Language
	if language := c.Query("language"); language != "" {
		queryOptions.Language = language
	}

	// Category
	if category := c.Query("category"); category != "" {
		queryOptions.CategoryValue = category
	}

	// Tag
	if tag := c.Query("tag"); tag != "" {
		queryOptions.TagValue = tag
	}

	// Featured
	if featured := c.Query("featured"); featured == "true" {
		queryOptions.Featured = true
	}

	// Status
	if status := c.Query("status"); status != "" {
		queryOptions.Status = types.BlogStatus(status)
	}

	// Limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			queryOptions.Limit = limit
		}
	}

	// Offset
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			queryOptions.Offset = offset
		}
	}

	// SortBy
	if sortBy := c.Query("sortBy"); sortBy != "" {
		queryOptions.SortBy = sortBy
	} else {
		queryOptions.SortBy = "created_at"
	}

	// SortDirection
	if sortDir := c.Query("sortDirection"); sortDir != "" {
		if sortDir == "asc" {
			queryOptions.SortDirection = types.SortAsc
		} else {
			queryOptions.SortDirection = types.SortDesc
		}
	} else {
		queryOptions.SortDirection = types.SortDesc // Varsayılan sıralama yönü
	}

	cacheKey := fmt.Sprintf("blog_cards: %v", queryOptions)
	if cachedData, exists := h.Cache.Get(cacheKey); exists {
		var blogs []types.BlogPostCardView
		if err := json.Unmarshal(cachedData, &blogs); err == nil {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"blogs":   blogs,
				"count":   len(blogs),
				"cached":  true,
			})
			return
		}
	}

	// Repository fonksiyonunu çağır
	blogs, err := h.BlogRepository.SelectBlogCards(queryOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Veriyi JSON'a çevir ve cache'e kaydet
	if jsonData, err := json.Marshal(blogs); err == nil {
		h.Cache.Set(cacheKey, jsonData)
	}

	// Sonuçları ve toplam sayısını döndür
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"blogs":   blogs,
		"count":   len(blogs),
	})
}
