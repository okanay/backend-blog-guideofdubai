package BlogHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (h *Handler) UpdateBlogStatus(c *gin.Context) {
	var request types.BlogUpdateStatusInput

	err := utils.ValidateRequest(c, &request)
	if err != nil {
		return
	}

	// Blog ID'yi doğrula
	blogID, err := uuid.Parse(request.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid_id",
			"message": "Geçersiz blog ID formatı.",
		})
		return
	}

	// Blog durumunu güncelle
	err = h.BlogRepository.UpdateBlogStatus(blogID, request.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	h.BlogCache.InvalidateAllBlogs()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Blog yazısı durumu başarıyla güncellendi.",
	})
}
