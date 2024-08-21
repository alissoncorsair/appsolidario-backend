package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/alissoncorsair/appsolidario-backend/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type R2Storage struct {
	client     *s3.Client
	bucketName string
	accountID  string
}

func NewR2Storage(accountID, bucketName string) (*R2Storage, error) {
	var accessKeyId = config.Envs.R2AccessKeyID
	var accessKeySecret = config.Envs.R2AccessKeySecret

	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")),
		awsconfig.WithRegion("auto"),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID))
	})

	return &R2Storage{
		client:     client,
		bucketName: bucketName,
		accountID:  accountID,
	}, nil
}

func (s *R2Storage) UploadFile(ctx context.Context, file io.Reader, filename string) (string, error) {
	uniqueFilename := uuid.New().String() + "-" + filename

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(uniqueFilename),
		Body:   file,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return s.generateFileURL(uniqueFilename), nil
}

func (s *R2Storage) generateFileURL(filename string) string {
	return fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s/%s", s.accountID, s.bucketName, filename)
}
