package R2Repository

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Repository struct {
	client        *s3.Client
	bucketName    string
	publicURLBase string
}

func NewRepository(accountID, accessKeyID, accessKeySecret, bucketName, folderName, publicURLBase, endpoint string) *Repository {

	// S3 istemcisini yapılandır
	s3Client := s3.New(s3.Options{
		Region: "eu-central-1",
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
			accessKeyID,
			accessKeySecret,
			"",
		)),
		BaseEndpoint: aws.String(endpoint),
	})

	return &Repository{
		publicURLBase: publicURLBase,
		bucketName:    bucketName,
		client:        s3Client,
	}
}
