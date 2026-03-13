package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/supplierpay/backend/internal/config"
	"go.uber.org/zap"
)

// ── Bedrock LLM Client ──────────────────────
type BedrockClient struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewBedrockClient(cfg *config.Config, logger *zap.Logger) *BedrockClient {
	return &BedrockClient{cfg: cfg, logger: logger}
}

func (b *BedrockClient) ExtractInvoiceFields(ctx context.Context, invoiceData []byte, mimeType string) (map[string]interface{}, error) {
	b.logger.Info("Calling Bedrock for invoice extraction",
		zap.String("model", b.cfg.BedrockModelID),
		zap.Int("data_size", len(invoiceData)),
	)

	// TODO: Dev 2 — Implement actual Bedrock API call
	// Prompt template:
	// "Extract the following fields from this invoice:
	//  - vendor_name, invoice_number, po_reference
	//  - line_items (description, quantity, unit_price, total)
	//  - total_amount, tax_amount, currency
	//  - invoice_date, due_date
	//  Return as JSON with confidence scores."

	return nil, fmt.Errorf("Bedrock extraction not yet implemented — use MOCK_MODE=true")
}

func (b *BedrockClient) GenerateQueryEmail(ctx context.Context, invoiceDetails map[string]interface{}, discrepancies []string) (string, error) {
	b.logger.Info("Calling Bedrock for query email generation")

	// TODO: Dev 2 — Implement email generation
	// Prompt: "Draft a professional email to supplier about invoice discrepancies..."

	return "", fmt.Errorf("Bedrock email generation not yet implemented — use MOCK_MODE=true")
}

func (b *BedrockClient) ExplainDecision(ctx context.Context, matchResult map[string]interface{}) (string, error) {
	b.logger.Info("Calling Bedrock for decision explanation")

	// TODO: Dev 2 — Implement decision explanation
	return "", fmt.Errorf("Bedrock decision explanation not yet implemented — use MOCK_MODE=true")
}

// ── Mock LLM Client (for local dev) ─────────
type MockLLMClient struct {
	logger *zap.Logger
}

func NewMockLLMClient(logger *zap.Logger) *MockLLMClient {
	return &MockLLMClient{logger: logger}
}

func (m *MockLLMClient) ExtractInvoiceFields(ctx context.Context, invoiceData []byte, mimeType string) (map[string]interface{}, error) {
	m.logger.Info("[MOCK] Extracting invoice fields")

	fields := map[string]interface{}{
		"vendor_name":    "Acme Corp",
		"invoice_number": "INV-2026-001",
		"po_reference":   "PO-2026-100",
		"total_amount":   50000.00,
		"tax_amount":     9000.00,
		"currency":       "INR",
		"invoice_date":   "2026-03-01",
		"due_date":       "2026-03-31",
		"confidence":     0.95,
		"line_items": []map[string]interface{}{
			{"description": "Cloud Hosting - March", "quantity": 1, "unit_price": 41000.00, "total": 41000.00},
			{"description": "Support Services", "quantity": 1, "unit_price": 9000.00, "total": 9000.00},
		},
	}
	return fields, nil
}

func (m *MockLLMClient) GenerateQueryEmail(ctx context.Context, invoiceDetails map[string]interface{}, discrepancies []string) (string, error) {
	m.logger.Info("[MOCK] Generating query email")

	detailsJSON, _ := json.MarshalIndent(invoiceDetails, "", "  ")

	email := fmt.Sprintf(`Subject: Invoice Discrepancy — Action Required

Dear Supplier,

We have identified the following discrepancies in your recent invoice:

%s

Invoice Details:
%s

Please review and provide a corrected invoice or clarification at your earliest convenience.

Best regards,
SupplierPay AI — Accounts Payable`, discrepancies, string(detailsJSON))

	return email, nil
}

func (m *MockLLMClient) ExplainDecision(ctx context.Context, matchResult map[string]interface{}) (string, error) {
	m.logger.Info("[MOCK] Explaining decision")
	return "Invoice auto-approved: PO matched, amounts verified, line items consistent.", nil
}
