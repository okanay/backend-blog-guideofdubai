package BlogHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (h *Handler) UpdateBlogPost(c *gin.Context) {
	var request types.BlogUpdateInput

	err := utils.ValidateRequest(c, &request)
	if err != nil {
		return
	}

	blog, err := h.BlogRepository.UpdateBlogPost(request)
	if err != nil {
		if utils.HandleDatabaseError(c, err, "Blog yazısı güncelleme") {
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "unexpected_error",
			"message": "Beklenmeyen bir hata oluştu: " + err.Error(),
		})
		return
	}

	// Tüm cache'i temizle - Daha hassas bir yaklaşım olarak sadece ilgili blogun cache'ini temizleyebiliriz
	blogID, _ := uuid.Parse(request.ID)
	if blogID != uuid.Nil {
		// İlgili blogun cache'ini temizle
		h.BlogCache.InvalidateBlogByID(blogID)
	} else {
		// ID geçersizse tüm cache'i temizle
		h.BlogCache.InvalidateAllBlogs()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Blog yazısı başarıyla güncellendi.",
		"blog":    blog,
	})
}
