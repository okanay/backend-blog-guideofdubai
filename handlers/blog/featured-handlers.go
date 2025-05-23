package BlogHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

// AddToFeatured bir blogu featured listesine ekler
func (h *Handler) AddToFeatured(c *gin.Context) {
	var request types.FeaturedBlogInput

	err := utils.ValidateRequest(c, &request)
	if err != nil {
		return
	}

	err = h.BlogRepository.AddToFeaturedList(request.BlogID, request.Language)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Cache'i temizle
	h.BlogCache.InvalidateAllBlogs()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Blog başarıyla featured listesine eklendi",
	})
}

// RemoveFromFeatured bir blogu featured listesinden çıkarır
func (h *Handler) RemoveFromFeatured(c *gin.Context) {
	blogIDString := c.Param("id")
	blogID, err := uuid.Parse(blogIDString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Geçersiz blog ID",
		})
		return
	}

	err = h.BlogRepository.RemoveFromFeaturedList(blogID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Tüm dillerdeki cache'i temizle
	h.BlogCache.InvalidateAllBlogs()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Blog tüm featured listelerinden çıkarıldı",
	})
}

// UpdateFeaturedOrdering featured blog sıralamasını günceller
func (h *Handler) UpdateFeaturedOrdering(c *gin.Context) {
	var request types.FeaturedBlogOrderingInput

	err := utils.ValidateRequest(c, &request)
	if err != nil {
		return
	}

	err = h.BlogRepository.UpdateFeaturedOrdering(request.Language, request.BlogIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// İlgili dilin cache'ini temizle
	h.BlogCache.InvalidateAllBlogs()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Featured blog sıralaması güncellendi",
	})
}

// GetFeaturedBlogs belirli bir dil için featured blogları getirir
func (h *Handler) GetFeaturedBlogs(c *gin.Context) {
	language := c.DefaultQuery("language", "en")

	// Cache'den kontrol et
	blogs, exists := h.BlogCache.GetFeaturedPostsByLanguage(language)
	if exists {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"blogs":   blogs,
			"count":   len(blogs),
			"cached":  true,
		})
		return
	}

	// Veritabanından getir
	blogs, err := h.BlogRepository.GetFeaturedBlogs(language)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Cache'e kaydet
	h.BlogCache.SaveFeaturedPostsByLanguage(language, blogs)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"blogs":   blogs,
		"count":   len(blogs),
		"cached":  false,
	})
}
