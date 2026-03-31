package aws

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/iancenry/jarvis/internal/server"
)

type S3Client struct {
	server *server.Server
	client *s3.Client
}

func NewS3Client(server *server.Server, cfg aws.Config) *S3Client {
	return &S3Client{
		server: server,
		client: s3.NewFromConfig(cfg),
	}
}

func (s *S3Client) UploadFile(ctx context.Context, bucket, fileKey string, file io.Reader, fileSize int64, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(fileKey),
		Body:          file,
		ContentLength: aws.Int64(fileSize),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return fileKey, nil
}

// CreatePresignedURL generates a presigned URL for downloading a file from S3. The URL will expire after 15 minutes.
func (s *S3Client) CreatePresignedURL(ctx context.Context, bucket, fileKey string) (string, error) {
	presignClient := s3.NewPresignClient(s.client)
	expiresIn := 15 * time.Minute

	presignedURL, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileKey),
	}, s3.WithPresignExpires(expiresIn))
	if err != nil {
		return "", fmt.Errorf("failed to create presigned URL: %w", err)
	}

	return presignedURL.URL, nil
}

func (s *S3Client) DeleteFile(ctx context.Context, bucket, fileKey string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	return nil
}
