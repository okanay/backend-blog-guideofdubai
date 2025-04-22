// handlers/image/create-presigned-url.go
package ImageHandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

// CreatePresignedURL dosya yüklemek için presigned URL oluşturur
func (h *Handler) CreatePresignedURL(c *gin.Context) {
	var input types.CreatePresignedURLInput
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

	// Presigned URL oluştur
	presignedOutput, err := h.R2Repository.GeneratePresignedURL(c.Request.Context(), types.PresignURLInput{
		Filename:    input.Filename,
		ContentType: input.ContentType,
		SizeInBytes: input.SizeInBytes,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "presigned_url_failed",
			"message": "Yükleme URL'si oluşturulamadı: " + err.Error(),
		})
		return
	}

	// Veritabanında signature kaydı oluştur
	signatureInput := types.UploadSignatureInput{
		PresignedURL: presignedOutput.PresignedURL,
		UploadURL:    presignedOutput.UploadURL,
		Filename:     input.Filename,
		FileType:     input.ContentType,
		ExpiresAt:    presignedOutput.ExpiresAt,
	}

	signatureID, err := h.ImageRepository.CreateUploadSignature(c.Request.Context(), userID, signatureInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "signature_creation_failed",
			"message": "Yükleme kaydı oluşturulamadı: " + err.Error(),
		})
		return
	}

	// Başarılı yanıt döndür
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": types.CreatePresignedURLResponse{
			ID:           signatureID.String(),
			PresignedURL: presignedOutput.PresignedURL,
			UploadURL:    presignedOutput.UploadURL,
			ExpiresAt:    presignedOutput.ExpiresAt,
			Filename:     input.Filename,
		},
	})
}
