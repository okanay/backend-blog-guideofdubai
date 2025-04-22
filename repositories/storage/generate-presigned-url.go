package StorageRepository

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/okanay/backend-blog-guideofdubai/types"
)

func (r *Repository) GeneratePresignedURL(ctx context.Context, input types.PresignURLInput) (*types.PresignedURLOutput, error) {
	objectKey := "key"

	presignClient := s3.NewPresignClient(r.client)

	putObjectRequest, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(objectKey),
		ContentType: aws.String(input.ContentType),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(5 * time.Minute)
	})

	if err != nil {
		return nil, fmt.Errorf("presigned URL oluşturulamadı: %w", err)
	}

	publicURL := ""

	return &types.PresignedURLOutput{
		PresignedURL: putObjectRequest.URL,
		UploadURL:    publicURL,
		ObjectKey:    objectKey,
		ExpiresAt:    time.Now().Add(5 * time.Minute),
	}, nil
}
