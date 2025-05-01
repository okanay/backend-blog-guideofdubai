// handlers/blog/select-blog-by-groupID.go
package BlogHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

func (h *Handler) SelectBlogByGroupID(c *gin.Context) {
	slugOrGroupID := c.Query("slug")
	lang := c.Query("lang") // Kullanım için dil parametresini alıyoruz ama repository'de kullanmıyoruz

	if slugOrGroupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "missing_parameter",
			"message": "slug parametresi gereklidir.",
		})
		return
	}

	// Cache key oluştur: "blog_slug:{slugOrGroupID}"
	cacheKey := "blog_slug:" + slugOrGroupID

	// Cache'den kontrol et
	cachedPost, cachedAlternatives, exists := h.BlogCache.GetBlogAndAlternativesBySlug(cacheKey)
	if exists {
		// Alternatifleri sadeleştir
		var altLinks []map[string]string
		for _, alt := range cachedAlternatives {
			altLinks = append(altLinks, map[string]string{
				"language": alt.Language,
				"slug":     alt.Slug,
				"title":    alt.Metadata.Title,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"success":      true,
			"blog":         cachedPost,
			"cached":       true,
			"alternatives": altLinks,
		})
		return
	}

	// Cache'de yoksa veritabanından getir
	request := types.BlogSelectByGroupIDInput{
		SlugOrGroupID: slugOrGroupID,
		Language:      lang, // Dili alıyoruz ama repository'de kullanmıyoruz
	}

	// Repository'yi çağır - slug odaklı arama için
	post, alternatives, err := h.BlogRepository.SelectBlogByGroupID(request)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "blog_not_found",
			"message": err.Error(),
		})
		return
	}

	// Cache'e kaydet - lang parametresini kullanmadan, sadece slug bazlı
	h.BlogCache.SaveBlogAndAlternativesBySlug(cacheKey, post, alternatives)

	// Alternatifleri sadeleştir
	var altLinks []map[string]string
	for _, alt := range alternatives {
		altLinks = append(altLinks, map[string]string{
			"language": alt.Language,
			"slug":     alt.Slug,
			"title":    alt.Metadata.Title,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"blog":         post,
		"cached":       false,
		"alternatives": altLinks,
	})
}
