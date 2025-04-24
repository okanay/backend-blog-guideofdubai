package utils

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidateRequest, request verilerinin doğruluğunu kontrol eder ve hataları işler
func ValidateRequest(c *gin.Context, req any) error {
	if err := c.ShouldBindJSON(&req); err != nil {
		// JSON ayrıştırma hatası
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "invalid_json",
			Message: "Geçersiz JSON formatı: " + err.Error(),
		})
		return err
	}

	validate := validator.New()
	err := validate.Struct(req)

	if err != nil {
		var errorMessages []string

		// Doğrulama hatalarını al
		validationErrors := err.(validator.ValidationErrors)
		for _, e := range validationErrors {
			// Daha kullanıcı dostu hata mesajları hazırla
			field := e.Field()
			switch e.Tag() {
			case "required":
				errorMessages = append(errorMessages, field+" alanı zorunludur")
			case "email":
				errorMessages = append(errorMessages, field+" alanı geçerli bir e-posta adresi olmalıdır")
			case "min":
				errorMessages = append(errorMessages, field+" alanı en az "+e.Param()+" karakter olmalıdır")
			case "max":
				errorMessages = append(errorMessages, field+" alanı en fazla "+e.Param()+" karakter olmalıdır")
			default:
				errorMessages = append(errorMessages, field+" alanı geçersiz: "+e.Tag())
			}
		}

		// Tüm hata mesajlarını birleştir
		message := strings.Join(errorMessages, ", ")

		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "validation_error",
			Message: message,
		})
		return err
	}

	return nil
}

// ValidateParam, URL parametrelerini doğrular
func ValidateParam(c *gin.Context, paramName string) (string, bool) {
	param := c.Param(paramName)
	if param == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "missing_parameter",
			Message: paramName + " parametresi eksik.",
		})
		return "", false
	}
	return param, true
}

// ValidateQuery, sorgu parametrelerini doğrular
func ValidateQuery(c *gin.Context, queryName string) (string, bool) {
	query := c.Query(queryName)
	if query == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "missing_query",
			Message: queryName + " sorgu parametresi eksik.",
		})
		return "", false
	}
	return query, true
}

// ValidateFileType, dosya türünü doğrular
func ValidateFileType(contentType string) bool {
	validTypes := map[string]bool{
		"image/jpeg":    true,
		"image/png":     true,
		"image/webp":    true,
		"image/gif":     true,
		"image/svg+xml": true,
	}

	return validTypes[contentType]
}
