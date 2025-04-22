// handlers/image/delete-image.go
package ImageHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DeleteImage bir resmi siler
func (h *Handler) DeleteImage(c *gin.Context) {
	// Resim ID'sini al
	imageIDStr := c.Param("id")
	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid_image_id",
			"message": "Geçersiz resim ID'si",
		})
		return
	}

	// Kullanıcı ID'sini al
	userID := c.MustGet("user_id").(uuid.UUID)

	// Resim bilgilerini getir
	image, err := h.ImageRepository.GetImageByID(c.Request.Context(), imageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "image_fetch_failed",
			"message": "Resim bilgileri getirilemedi: " + err.Error(),
		})
		return
	}

	if image == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "image_not_found",
			"message": "Resim bulunamadı",
		})
		return
	}

	// Resmin sahibini kontrol et
	if image.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "permission_denied",
			"message": "Bu resmi silme yetkiniz yok",
		})
		return
	}

	// R2'den resmi sil
	// URL'den object key'i çıkarmak gerekiyor, eğer URL: https://base/uploads/image.jpg ise
	// objectKey: uploads/image.jpg olmalı
	objectKey := extractObjectKeyFromURL(image.URL)

	err = h.R2Repository.DeleteObject(c.Request.Context(), objectKey)
	if err != nil {
		// Sadece loglama yap, silme işlemine devam et
		// log.Printf("R2 object deletion failed: %v", err)
	}

	// Veritabanından resmi sil
	err = h.ImageRepository.DeleteImage(c.Request.Context(), imageID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "image_delete_failed",
			"message": "Resim silinemedi: " + err.Error(),
		})
		return
	}

	// Başarılı yanıt döndür
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Resim başarıyla silindi",
	})
}

// URL'den object key'i çıkaran yardımcı fonksiyon
func extractObjectKeyFromURL(url string) string {
	// Base URL'yi çıkar, örn: https://blog-assets.project-test.info/uploads/image.jpg -> uploads/image.jpg
	// Bu örnek için basit bir implementasyon, gerçek uygulamada daha sağlam bir çözüm gerekebilir
	baseURL := "https://blog-assets.project-test.info/"
	if len(url) > len(baseURL) && url[:len(baseURL)] == baseURL {
		return url[len(baseURL):]
	}
	return url
}
