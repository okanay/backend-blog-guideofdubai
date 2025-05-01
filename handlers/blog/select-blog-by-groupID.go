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

	// Cache kontrolü (opsiyonel)
	blog, exists := h.BlogCache.GetBlogByGroupIDAndLang(slugOrGroupID, lang)
	if exists {
		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"blog":     blog,
			"cached":   true,
			"priority": 1,
		})
		return
	}

	request := types.BlogSelectByGroupIDInput{
		SlugOrGroupID: slugOrGroupID,
		Language:      lang,
	}

	blog, priority, err := h.BlogRepository.SelectBlogByGroupID(request)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "blog_not_found",
			"message": err.Error(),
		})
		return
	}

	h.BlogCache.SaveBlogByGroupIDAndLang(slugOrGroupID, lang, blog)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"blog":     blog,
		"cached":   false,
		"priority": priority,
		"fallback": priority > 2,
	})
}
