package R2Repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// DeleteObject R2 bucket'tan bir nesneyi siler
func (r *Repository) DeleteObject(ctx context.Context, objectKey string) error {
	// S3 DeleteObject API'sini çağır
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(objectKey),
	})

	if err != nil {
		return fmt.Errorf("nesne silinemedi (key: %s): %w", objectKey, err)
	}

	return nil
}
