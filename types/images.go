package types

import (
	"time"

	"github.com/google/uuid"
)

// Image görüntü tablosundaki kayıtlar için
type Image struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"userId"`
	URL         string    `json:"url"`
	Filename    string    `json:"filename"`
	AltText     string    `json:"altText"`
	FileType    string    `json:"fileType"`
	SizeInBytes int64     `json:"sizeInBytes"`
	Width       int       `json:"width,omitempty"`
	Height      int       `json:"height,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ImageCreateInput bir görüntü oluşturmak için girdi
type ImageCreateInput struct {
	URL         string `json:"url" binding:"required"`
	Filename    string `json:"filename" binding:"required"`
	AltText     string `json:"altText"`
	FileType    string `json:"fileType" binding:"required"`
	SizeInBytes int64  `json:"sizeInBytes" binding:"required"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
}

// UploadSignature imza tablosundaki kayıtlar için
type UploadSignature struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"userId"`
	ImageID      uuid.UUID `json:"imageId"`
	PresignedURL string    `json:"presignedUrl"`
	UploadURL    string    `json:"uploadUrl"`
	Filename     string    `json:"filename"`
	FileType     string    `json:"fileType"`
	ExpiresAt    time.Time `json:"expiresAt"`
	Completed    bool      `json:"completed"`
	CreatedAt    time.Time `json:"createdAt"`
}

// SignatureCreateInput bir imza oluşturmak için girdi
type SignatureCreateInput struct {
	ImageID      uuid.UUID `json:"imageId" binding:"required"`
	PresignedURL string    `json:"presignedUrl" binding:"required"`
	UploadURL    string    `json:"uploadUrl" binding:"required"`
	Filename     string    `json:"filename" binding:"required"`
	FileType     string    `json:"fileType" binding:"required"`
	ExpiresAt    time.Time `json:"expiresAt" binding:"required"`
}

// PresignURLInput Presigned URL oluşturmak için girdi
type PresignURLInput struct {
	Filename    string `json:"filename" binding:"required"`
	ContentType string `json:"contentType" binding:"required"`
	Folder      string `json:"folder" binding:"required"`
}

// PresignedURLOutput Presigned URL oluşturma çıktısı
type PresignedURLOutput struct {
	PresignedURL string    `json:"presignedUrl"`
	UploadURL    string    `json:"uploadUrl"`
	ObjectKey    string    `json:"objectKey"`
	ExpiresAt    time.Time `json:"expiresAt"`
}
