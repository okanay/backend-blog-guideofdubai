// handlers/blog/select-blog-by-groupID.go
package BlogHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

func (h *Handler) SelectBlogByGroupID(c *gin.Context) {
	slugOrGroupID := c.Query("slug")
	lang := c.Query("lang")

	if slugOrGroupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "missing_parameter",
			"message": "slug parametresi gereklidir.",
		})
		return
	}

	// Cache key oluştur
	cacheKey := "blog_slug:" + slugOrGroupID

	// Blog'u al (cache veya DB'den)
	var post *types.BlogPostView
	var alternatives []*types.BlogPostView
	var cached bool

	// Cache kontrolü
	cachedPost, cachedAlternatives, exists := h.BlogCache.GetBlogAndAlternativesBySlug(cacheKey)
	if exists {
		post = cachedPost
		alternatives = cachedAlternatives
		cached = true
	} else {
		// Veritabanından getir
		request := types.BlogSelectByGroupIDInput{
			SlugOrGroupID: slugOrGroupID,
			Language:      lang,
		}

		var err error
		post, alternatives, err = h.BlogRepository.SelectBlogByGroupID(request)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "blog_not_found",
				"message": err.Error(),
			})
			return
		}

		// Cache'e kaydet
		h.BlogCache.SaveBlogAndAlternativesBySlug(cacheKey, post, alternatives)
		cached = false
	}

	// Blog ID'yi middleware için set et
	if post != nil && post.ID != "" {
		if id, err := uuid.Parse(post.ID); err == nil {
			c.Set("blog_id", id)
		}
	}

	// Alternatifleri sadeleştir
	var altLinks []map[string]string
	for _, alt := range alternatives {
		altLinks = append(altLinks, map[string]string{
			"language": alt.Language,
			"slug":     alt.Slug,
			"title":    alt.Metadata.Title,
		})
	}

	// Response'u gönder
	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"blog":         post,
		"cached":       cached,
		"alternatives": altLinks,
	})
}
