package services

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	appconfig "github.com/supplierpay/backend/internal/config"
	"go.uber.org/zap"
)

// ── S3 Storage Client ───────────────────────
type S3Client struct {
	cfg    *appconfig.Config
	logger *zap.Logger
	client *s3.Client
}

func NewS3Client(cfg *appconfig.Config, logger *zap.Logger) *S3Client {
	return &S3Client{cfg: cfg, logger: logger}
}

func (s *S3Client) initClient(ctx context.Context) error {
	if s.client != nil {
		return nil
	}

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(s.cfg.S3Region),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	s.client = s3.NewFromConfig(awsCfg)
	return nil
}

func (s *S3Client) UploadFile(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	s.logger.Info("Uploading to S3",
		zap.String("bucket", s.cfg.S3BucketName),
		zap.String("key", key),
		zap.Int("size", len(data)),
	)

	if err := s.initClient(ctx); err != nil {
		return "", err
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.cfg.S3BucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		s.logger.Error("Failed to upload to S3", zap.Error(err))
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.cfg.S3BucketName, s.cfg.S3Region, key)
	s.logger.Info("Successfully uploaded to S3", zap.String("url", url))
	return url, nil
}

func (s *S3Client) GetFile(ctx context.Context, key string) ([]byte, error) {
	s.logger.Info("Getting from S3", zap.String("key", key))

	if err := s.initClient(ctx); err != nil {
		return nil, err
	}

	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.cfg.S3BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		s.logger.Error("Failed to get from S3", zap.Error(err))
		return nil, fmt.Errorf("failed to get from S3: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read S3 object body: %w", err)
	}

	return data, nil
}

func (s *S3Client) GetPresignedURL(ctx context.Context, key string) (string, error) {
	s.logger.Info("Generating presigned URL", zap.String("key", key))

	// For simplicity, return direct URL (in production, use presigner)
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.cfg.S3BucketName, s.cfg.S3Region, key), nil
}

// ── Mock Storage (local filesystem) ─────────
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
