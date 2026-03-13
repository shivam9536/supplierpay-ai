package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/supplierpay/backend/internal/config"
	"go.uber.org/zap"
)

// ── Pine Labs Payment Client ────────────────
type PineLabsClient struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewPineLabsClient(cfg *config.Config, logger *zap.Logger) *PineLabsClient {
	return &PineLabsClient{cfg: cfg, logger: logger}
}

func (p *PineLabsClient) InitiateDisbursement(ctx context.Context, req DisbursementRequest) (*DisbursementResponse, error) {
	p.logger.Info("Initiating Pine Labs disbursement",
		zap.String("invoice_id", req.InvoiceID),
		zap.Float64("amount", req.Amount),
		zap.String("vendor", req.VendorName),
	)

	// TODO: Dev 1 — Implement Pine Labs API call
	// POST to PINELABS_API_URL/disbursement
	// Headers: X-API-Key, X-Merchant-ID
	// Body: account details, amount, reference

	return nil, fmt.Errorf("Pine Labs integration not yet implemented — use MOCK_MODE=true")
}

func (p *PineLabsClient) GetTransactionStatus(ctx context.Context, transactionID string) (*TransactionStatus, error) {
	p.logger.Info("Checking Pine Labs transaction status", zap.String("transaction_id", transactionID))

	// TODO: Dev 1 — GET PINELABS_API_URL/transactions/{transactionID}
	return nil, fmt.Errorf("Pine Labs status check not yet implemented")
}

// ── Mock Payment Client ─────────────────────
type MockPaymentClient struct {
	logger *zap.Logger
}

func NewMockPaymentClient(logger *zap.Logger) *MockPaymentClient {
	return &MockPaymentClient{logger: logger}
}

func (m *MockPaymentClient) InitiateDisbursement(ctx context.Context, req DisbursementRequest) (*DisbursementResponse, error) {
	txnID := uuid.New().String()
	m.logger.Info("[MOCK] Disbursement initiated",
		zap.String("transaction_id", txnID),
		zap.String("invoice_id", req.InvoiceID),
		zap.Float64("amount", req.Amount),
	)

	return &DisbursementResponse{
		TransactionID: txnID,
		Status:        "PENDING",
		Message:       "Mock disbursement initiated successfully",
	}, nil
}

func (m *MockPaymentClient) GetTransactionStatus(ctx context.Context, transactionID string) (*TransactionStatus, error) {
	m.logger.Info("[MOCK] Transaction status check", zap.String("transaction_id", transactionID))

	return &TransactionStatus{
		TransactionID: transactionID,
		Status:        "SUCCESS",
		Timestamp:     "2026-03-12T10:00:00Z",
	}, nil
}
