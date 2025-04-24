// handlers/image/confirm-upload.go
package ImageHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

func (h *Handler) ConfirmUpload(c *gin.Context) {
	var input types.ConfirmUploadInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid_input",
			"message": "Geçersiz istek formatı: " + err.Error(),
		})
		return
	}

	// Kullanıcı ID'sini al
	userID := c.MustGet("user_id").(uuid.UUID)

	// SignatureID'yi UUID'ye dönüştür
	signatureID, err := uuid.Parse(input.SignatureID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid_signature_id",
			"message": "Geçersiz imza ID'si",
		})
		return
	}

	signature, err := h.ImageRepository.GetSignatureByID(c.Request.Context(), signatureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "signature_fetch_failed",
			"message": "İmza bilgileri alınamadı: " + err.Error(),
		})
		return
	}

	if signature.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "permission_denied",
			"message": "Bu yükleme işlemi için yetkiniz yok",
		})
		return
	}

	// Resmi veritabanına kaydet
	imageInput := types.SaveImageInput{
		URL:         input.URL,
		Filename:    signature.Filename,
		AltText:     input.AltText,
		FileType:    signature.FileType,
		SizeInBytes: input.SizeInBytes,
		Width:       input.Width,
		Height:      input.Height,
	}

	imageID, err := h.ImageRepository.SaveImage(c.Request.Context(), userID, imageInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "image_save_failed",
			"message": "Resim kaydedilemedi: " + err.Error(),
		})
		return
	}

	// İmza kaydını tamamlandı olarak işaretle
	err = h.ImageRepository.CompleteUploadSignature(c.Request.Context(), signatureID, imageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "signature_update_failed",
			"message": "İmza kaydı güncellenemedi: " + err.Error(),
		})
		return
	}

	// Başarılı yanıt döndür
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":  imageID.String(),
			"url": input.URL,
		},
	})
}
