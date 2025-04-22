// handlers/image/get-user-images.go
package ImageHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetUserImages kullanıcıya ait resimleri getirir
func (h *Handler) GetUserImages(c *gin.Context) {
	// Kullanıcı ID'sini al
	userID := c.MustGet("user_id").(uuid.UUID)

	// Resimleri getir
	images, err := h.ImageRepository.GetImagesByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "image_fetch_failed",
			"message": "Resimler getirilemedi: " + err.Error(),
		})
		return
	}

	// Yanıt için dönüştür
	var response []gin.H
	for _, img := range images {
		response = append(response, gin.H{
			"id":          img.ID.String(),
			"url":         img.URL,
			"filename":    img.Filename,
			"altText":     img.AltText,
			"fileType":    img.FileType,
			"sizeInBytes": img.SizeInBytes,
			"width":       img.Width,
			"height":      img.Height,
			"createdAt":   img.CreatedAt,
		})
	}

	// Başarılı yanıt döndür
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}
