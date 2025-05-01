package BlogHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

	if lang == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "missing_parameter",
			"message": "lang parametresi gereklidir.",
		})
		return
	}

	request := types.BlogSelectByGroupIDInput{
		SlugOrGroupID: slugOrGroupID,
		Language:      lang,
	}

	// Yeni fonksiyon ile hem ana postu hem alternatifleri çekiyoruz
	post, alternatives, err := h.BlogRepository.SelectBlogByGroupID(request)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "blog_not_found",
			"message": err.Error(),
		})
		return
	}

	// Alternatifleri sadeleştir (dilersen sadece dil, slug, title dönebilirsin)
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
