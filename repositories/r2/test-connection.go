// repositories/r2/test-connection.go
package R2Repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// TestConnection bağlantının çalışıp çalışmadığını test eder
func (r *Repository) TestConnection(ctx context.Context) error {
	listResult, err := r.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  &r.bucketName,
		Prefix:  &r.folderName,
		MaxKeys: aws.Int32(1),
	})

	if err != nil {
		return fmt.Errorf("R2 bağlantı testi başarısız: %w", err)
	}

	// Başarılı bağlantı, nesne sayısını kontrol et
	fmt.Printf("R2 bağlantı testi başarılı! %d nesne listelendi.\n", len(listResult.Contents))
	return nil
}
