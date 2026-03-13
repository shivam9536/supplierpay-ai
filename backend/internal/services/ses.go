package services

import (
	"context"
	"fmt"

	"github.com/supplierpay/backend/internal/config"
	"go.uber.org/zap"
)

// ── SES Email Client ────────────────────────
type SESClient struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewSESClient(cfg *config.Config, logger *zap.Logger) *SESClient {
	return &SESClient{cfg: cfg, logger: logger}
}

func (s *SESClient) SendEmail(ctx context.Context, to, subject, body string) error {
	s.logger.Info("Sending email via SES",
		zap.String("from", s.cfg.SESSenderEmail),
		zap.String("to", to),
		zap.String("subject", subject),
	)

	// TODO: Dev 1 — Implement SES send email
	// Use aws-sdk-go-v2/service/ses SendEmail

	return fmt.Errorf("SES not yet implemented — use MOCK_MODE=true")
}

// ── Mock Email Client ───────────────────────
type MockEmailClient struct {
	logger *zap.Logger
}

func NewMockEmailClient(logger *zap.Logger) *MockEmailClient {
	return &MockEmailClient{logger: logger}
}

func (m *MockEmailClient) SendEmail(ctx context.Context, to, subject, body string) error {
	m.logger.Info("[MOCK] Email sent",
		zap.String("to", to),
		zap.String("subject", subject),
		zap.Int("body_length", len(body)),
	)
	// Just log it — no real email sent
	return nil
}
