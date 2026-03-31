package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	appconfig "github.com/iancenry/jarvis/internal/config"
)

type AWS struct {
	S3 *S3Client
}

func NewAWS(awsConfig appconfig.AWSConfig) (*AWS, error) {
	configOptions := []func(*config.LoadOptions) error{
		config.WithRegion(awsConfig.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			awsConfig.AccessKeyID,
			awsConfig.SecretAccessKey,
			"",
		)),
	}

	// Add custom endpoint if provided (for S3-compatible services like Sevalla)
	if awsConfig.EndpointURL != "" {
		configOptions = append(configOptions, config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string,
				options ...interface{},
			) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           awsConfig.EndpointURL,
					SigningRegion: awsConfig.Region,
				}, nil
			}),
		))
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), configOptions...)
	if err != nil {
		return nil, err
	}

	return &AWS{
		S3: NewS3Client(cfg),
	}, nil
}
