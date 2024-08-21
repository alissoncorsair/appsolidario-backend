package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/alissoncorsair/appsolidario-backend/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type R2Storage struct {
	client     *s3.Client
	bucketName string
}

func NewR2Storage(accountID, bucketName string) (*R2Storage, error) {
	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion("auto"),
		awsconfig.WithEndpointResolver(aws.EndpointResolverFunc(
			func(service, region string) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID),
				}, nil
			},
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &R2Storage{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (s *R2Storage) UploadFile(file io.Reader, filename string) (string, error) {
	uniqueFilename := uuid.New().String() + "-" + filename

	_, err := s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(uniqueFilename),
		Body:   file,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s/%s", config.Envs.R2AccountID, s.bucketName, uniqueFilename), nil
}
