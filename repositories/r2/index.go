package R2Repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Repository struct {
	client        *s3.Client
	bucketName    string
	folderName    string
	publicURLBase string
}

func NewRepository(accountID, accessKeyID, accessKeySecret, bucketName, folderName, publicURLBase, endpoint string) *Repository {
	// SDK konfigürasyonu oluştur
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("auto"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID,
			accessKeySecret,
			"",
		)),
	)
	if err != nil {
		panic(fmt.Sprintf("R2 configuration error: %s", err))
	}

	// S3 istemcisini özel endpoint ile oluştur
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true // R2 ile path-style URL kullanımı genellikle tercih edilir
	})

	return &Repository{
		client:        s3Client,
		bucketName:    bucketName,
		folderName:    folderName,
		publicURLBase: publicURLBase,
	}
}
