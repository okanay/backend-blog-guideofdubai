package BlogHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
	"github.com/okanay/backend-blog-guideofdubai/utils"
)

func (h *Handler) CreateBlogCategory(c *gin.Context) {
	var request types.CategoryInput

	err := utils.ValidateRequest(c, &request)
	if err != nil {
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	category, err := h.BlogRepository.CreateBlogCategory(request, userID)

	if err != nil {
		if utils.HandleDatabaseError(c, err, "Blog kategori oluşturma") {
			return
		}
		return
	}

	// Kategori oluşturulduğunda kategori ile ilgili önbellekleri temizle
	// Bu durumda tüm blog önbelleklerini temizlemek en güvenli yaklaşımdır
	h.BlogCache.InvalidateAllBlogs()

	c.JSON(http.StatusCreated, gin.H{
		"success":  true,
		"category": category,
	})
}
