package services

import (
	"context"
	"fmt"

	"github.com/supplierpay/backend/internal/config"
	"go.uber.org/zap"
)

// ── S3 Storage Client ───────────────────────
type S3Client struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewS3Client(cfg *config.Config, logger *zap.Logger) *S3Client {
	return &S3Client{cfg: cfg, logger: logger}
}

func (s *S3Client) UploadFile(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	s.logger.Info("Uploading to S3",
		zap.String("bucket", s.cfg.S3BucketName),
		zap.String("key", key),
		zap.Int("size", len(data)),
	)

	// TODO: Dev 1 — Implement S3 upload
	// Use aws-sdk-go-v2/service/s3 PutObject

	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.cfg.S3BucketName, s.cfg.S3Region, key), nil
}

func (s *S3Client) GetFile(ctx context.Context, key string) ([]byte, error) {
	s.logger.Info("Getting from S3", zap.String("key", key))

	// TODO: Dev 1 — Implement S3 download
	return nil, fmt.Errorf("S3 download not yet implemented")
}

func (s *S3Client) GetPresignedURL(ctx context.Context, key string) (string, error) {
	// TODO: Dev 1 — Implement presigned URL generation
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s?presigned=true", s.cfg.S3BucketName, s.cfg.S3Region, key), nil
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
