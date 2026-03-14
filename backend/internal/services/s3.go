package services

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/supplierpay/backend/internal/config"
	"go.uber.org/zap"
)

// ── S3 Storage Client ───────────────────────
type S3Client struct {
	cfg    *config.Config
	logger *zap.Logger
	client *s3.Client
}

func NewS3Client(cfg *config.Config, logger *zap.Logger) *S3Client {
	region := cfg.S3Region
	if region == "" {
		region = cfg.AWSRegion
	}
	if region == "" {
		region = "us-east-1"
	}

	var awsCfg aws.Config
	var err error

	if cfg.AWSAccessKeyID != "" && cfg.AWSSecretAccessKey != "" {
		logger.Info("S3: using IAM key/secret from environment")
		staticCreds := credentials.NewStaticCredentialsProvider(
			cfg.AWSAccessKeyID,
			cfg.AWSSecretAccessKey,
			"",
		)
		awsCfg, err = awsconfig.LoadDefaultConfig(context.Background(),
			awsconfig.WithRegion(region),
			awsconfig.WithCredentialsProvider(staticCreds),
		)
	} else {
		logger.Info("S3: using default AWS credential chain")
		awsCfg, err = awsconfig.LoadDefaultConfig(context.Background(),
			awsconfig.WithRegion(region),
		)
	}

	if err != nil {
		logger.Error("S3: failed to build AWS config", zap.Error(err))
	}

	return &S3Client{
		cfg:    cfg,
		logger: logger,
		client: s3.NewFromConfig(awsCfg),
	}
}

func (s *S3Client) UploadFile(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	s.logger.Info("Uploading to S3",
		zap.String("bucket", s.cfg.S3BucketName),
		zap.String("key", key),
		zap.Int("size", len(data)),
	)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.cfg.S3BucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("s3 upload failed: %w", err)
	}

	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.cfg.S3BucketName, s.cfg.S3Region, key)
	s.logger.Info("S3 upload successful", zap.String("url", url))
	return url, nil
}

func (s *S3Client) GetFile(ctx context.Context, key string) ([]byte, error) {
	s.logger.Info("Getting from S3",
		zap.String("bucket", s.cfg.S3BucketName),
		zap.String("key", key),
	)

	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.cfg.S3BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("s3 get failed: %w", err)
	}
	defer out.Body.Close()

	data, err := io.ReadAll(out.Body)
	if err != nil {
		return nil, fmt.Errorf("s3 read body failed: %w", err)
	}
	return data, nil
}

func (s *S3Client) GetPresignedURL(ctx context.Context, key string) (string, error) {
	presignClient := s3.NewPresignClient(s.client)
	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.cfg.S3BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", fmt.Errorf("s3 presign failed: %w", err)
	}
	return req.URL, nil
}

// ── Mock Storage (local dev) ─────────────────
type MockStorageClient struct {
	logger *zap.Logger
}

func NewMockStorageClient(logger *zap.Logger) *MockStorageClient {
	return &MockStorageClient{logger: logger}
}

func (m *MockStorageClient) UploadFile(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	m.logger.Info("[MOCK] File uploaded", zap.String("key", key), zap.Int("size", len(data)))
	return fmt.Sprintf("http://localhost:8080/mock-files/%s", key), nil
}

func (m *MockStorageClient) GetFile(ctx context.Context, key string) ([]byte, error) {
	m.logger.Info("[MOCK] Getting file", zap.String("key", key))
	return []byte("mock-file-data"), nil
}

func (m *MockStorageClient) GetPresignedURL(ctx context.Context, key string) (string, error) {
	return fmt.Sprintf("http://localhost:8080/mock-files/%s", key), nil
}
