package utils

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// ErrorResponse represents a standard error response format
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

// ErrorCode is a custom error type for application-specific errors
type ErrorCode string

// Predefined error codes
const (
	ErrorDuplicateEntry     ErrorCode = "duplicate_entry"
	ErrorInvalidReference   ErrorCode = "invalid_reference"
	ErrorMissingField       ErrorCode = "missing_required_field"
	ErrorInvalidValue       ErrorCode = "invalid_value"
	ErrorDatabaseError      ErrorCode = "database_error"
	ErrorNotFound           ErrorCode = "not_found"
	ErrorAuthenticationFail ErrorCode = "authentication_failed"
	ErrorForbidden          ErrorCode = "forbidden"
	ErrorOperationFailed    ErrorCode = "operation_failed"
)

// HttpStatus maps error codes to HTTP status codes
var HttpStatus = map[ErrorCode]int{
	ErrorDuplicateEntry:     http.StatusConflict,
	ErrorInvalidReference:   http.StatusBadRequest,
	ErrorMissingField:       http.StatusBadRequest,
	ErrorInvalidValue:       http.StatusBadRequest,
	ErrorDatabaseError:      http.StatusInternalServerError,
	ErrorNotFound:           http.StatusNotFound,
	ErrorAuthenticationFail: http.StatusUnauthorized,
	ErrorForbidden:          http.StatusForbidden,
	ErrorOperationFailed:    http.StatusBadRequest,
}

// ErrorMessages maps error codes to user-friendly messages
var ErrorMessages = map[ErrorCode]string{
	ErrorDuplicateEntry:     "Bu kayıt zaten mevcut.",
	ErrorInvalidReference:   "Geçersiz referans.",
	ErrorMissingField:       "Zorunlu alan eksik.",
	ErrorInvalidValue:       "Geçersiz değer.",
	ErrorDatabaseError:      "Veritabanı işlemi sırasında bir hata oluştu.",
	ErrorNotFound:           "Kayıt bulunamadı.",
	ErrorAuthenticationFail: "Kimlik doğrulama başarısız.",
	ErrorForbidden:          "Bu işlem için yetkiniz yok.",
	ErrorOperationFailed:    "İşlem başarısız oldu.",
}

// SpecificErrorHandler allows checking for specific PostgreSQL errors
type SpecificErrorHandler struct {
	Code          string
	ConstraintKey string
	ErrorCode     ErrorCode
	Message       string
}

// HandleDatabaseError handles common database errors and returns appropriate responses
func HandleDatabaseError(c *gin.Context, err error, operation string) bool {
	// Check if this is a PostgreSQL error
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		// Define specific error handlers for different scenarios
		specificHandlers := []SpecificErrorHandler{
			// Unique constraint violations
			{
				Code:          "23505",
				ConstraintKey: "users_username_key",
				ErrorCode:     "username_exists",
				Message:       "Bu kullanıcı adı zaten kullanımda.",
			},
			{
				Code:          "23505",
				ConstraintKey: "users_email_key",
				ErrorCode:     "email_exists",
				Message:       "Bu e-posta adresi zaten kullanımda.",
			},
			{
				Code:          "23505",
				ConstraintKey: "blog_posts_slug_language_key",
				ErrorCode:     "slug_language_exists",
				Message:       "Bu url yapısı ve dil kombinasyonu zaten kullanımda.",
			},
			{
				Code:          "23505",
				ConstraintKey: "categories_name_key",
				ErrorCode:     "category_exists",
				Message:       "Bu kategori adı zaten kullanımda.",
			},
			{
				Code:          "23505",
				ConstraintKey: "tags_name_key",
				ErrorCode:     "tag_exists",
				Message:       "Bu etiket adı zaten kullanımda.",
			},
		}

		// Check for specific error conditions first
		for _, handler := range specificHandlers {
			if pqErr.Code == pq.ErrorCode(handler.Code) && strings.Contains(pqErr.Constraint, handler.ConstraintKey) {
				c.JSON(HttpStatus[ErrorDuplicateEntry], ErrorResponse{
					Success: false,
					Error:   string(handler.ErrorCode),
					Message: handler.Message,
				})
				return true
			}
		}

		// If no specific handler matched, use generic error handling by PostgreSQL error code
		switch pqErr.Code {
		case "23505": // Unique Violation (general case)
			c.JSON(HttpStatus[ErrorDuplicateEntry], ErrorResponse{
				Success: false,
				Error:   string(ErrorDuplicateEntry),
				Message: ErrorMessages[ErrorDuplicateEntry] + " Constraint: " + pqErr.Constraint,
			})
			return true

		case "23503": // Foreign Key Violation
			c.JSON(HttpStatus[ErrorInvalidReference], ErrorResponse{
				Success: false,
				Error:   string(ErrorInvalidReference),
				Message: ErrorMessages[ErrorInvalidReference] + " " + pqErr.Detail,
			})
			return true

		case "23502": // Not Null Violation
			c.JSON(HttpStatus[ErrorMissingField], ErrorResponse{
				Success: false,
				Error:   string(ErrorMissingField),
				Message: ErrorMessages[ErrorMissingField] + " Sütun: " + pqErr.Column,
			})
			return true

		case "23514": // Check Violation
			c.JSON(HttpStatus[ErrorInvalidValue], ErrorResponse{
				Success: false,
				Error:   string(ErrorInvalidValue),
				Message: ErrorMessages[ErrorInvalidValue] + " " + pqErr.Detail,
			})
			return true

		default: // Other PostgreSQL errors
			c.JSON(HttpStatus[ErrorDatabaseError], ErrorResponse{
				Success: false,
				Error:   string(ErrorDatabaseError),
				Message: ErrorMessages[ErrorDatabaseError] + " " + pqErr.Message,
			})
			return true
		}
	}

	// For non-PostgreSQL errors, return a generic operation failed message
	c.JSON(HttpStatus[ErrorOperationFailed], ErrorResponse{
		Success: false,
		Error:   string(ErrorOperationFailed),
		Message: operation + " işlemi başarısız: " + err.Error(),
	})
	return true
}

// NotFound handles not found errors
func NotFound(c *gin.Context, resource string) {
	c.JSON(HttpStatus[ErrorNotFound], ErrorResponse{
		Success: false,
		Error:   string(ErrorNotFound),
		Message: resource + " bulunamadı.",
	})
}

// Unauthorized handles authentication failures
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = ErrorMessages[ErrorAuthenticationFail]
	}

	c.JSON(HttpStatus[ErrorAuthenticationFail], ErrorResponse{
		Success: false,
		Error:   string(ErrorAuthenticationFail),
		Message: message,
	})
}

// Forbidden handles permission errors
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = ErrorMessages[ErrorForbidden]
	}

	c.JSON(HttpStatus[ErrorForbidden], ErrorResponse{
		Success: false,
		Error:   string(ErrorForbidden),
		Message: message,
	})
}

// BadRequest handles invalid request errors
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   "invalid_request",
		Message: message,
	})
}

// SendError is a generic error response function
func SendError(c *gin.Context, errorCode ErrorCode, message string) {
	status, exists := HttpStatus[errorCode]
	if !exists {
		status = http.StatusInternalServerError
	}

	if message == "" {
		message, exists = ErrorMessages[errorCode]
		if !exists {
			message = "Bilinmeyen hata."
		}
	}

	c.JSON(status, ErrorResponse{
		Success: false,
		Error:   string(errorCode),
		Message: message,
	})
}
