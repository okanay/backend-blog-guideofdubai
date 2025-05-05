package BlogHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

func (h *Handler) SelectBlogBySlugID(c *gin.Context) {
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

	// Blog'u getir (cache veya DB'den)
	post, alternatives, cached, err := h.getBlogBySlugOrGroupID(slugOrGroupID, lang)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "blog_not_found",
			"message": err.Error(),
		})
		return
	}

	// Response hazırla ve gönder
	h.sendBlogResponse(c, post, alternatives, cached)
}

// Helper fonksiyonlar
func (h *Handler) getBlogBySlugOrGroupID(slugOrGroupID, lang string) (*types.BlogPostView, []*types.BlogPostView, bool, error) {
	cacheKey := "blog_slug:" + slugOrGroupID

	// Cache'den kontrol et
	if cachedPost, cachedAlternatives, exists := h.BlogCache.GetBlogAndAlternativesBySlug(cacheKey); exists {
		return cachedPost, cachedAlternatives, true, nil
	}

	// Cache'de yoksa veritabanından getir
	request := types.BlogSelectByGroupIDInput{
		SlugOrGroupID: slugOrGroupID,
		Language:      lang,
	}

	post, alternatives, err := h.BlogRepository.SelectBlogByGroupID(request)
	if err != nil {
		return nil, nil, false, err
	}

	// Cache'e kaydet
	h.BlogCache.SaveBlogAndAlternativesBySlug(cacheKey, post, alternatives)

	return post, alternatives, false, nil
}

func (h *Handler) sendBlogResponse(c *gin.Context, post *types.BlogPostView, alternatives []*types.BlogPostView, cached bool) {
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
		"cached":       cached,
		"alternatives": altLinks,
	})
}
