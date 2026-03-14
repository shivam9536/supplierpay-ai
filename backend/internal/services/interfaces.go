package services

import "context"

// ── LLM Service Interface ───────────────────
// Implemented by Bedrock client and mock client
type LLMService interface {
	// ExtractInvoiceFields extracts structured data from invoice text/image
	ExtractInvoiceFields(ctx context.Context, invoiceData []byte, mimeType string) (map[string]interface{}, error)

	// GenerateQueryEmail drafts a supplier query email for discrepancies
	GenerateQueryEmail(ctx context.Context, invoiceDetails map[string]interface{}, discrepancies []string) (string, error)

	// ExplainDecision generates a human-readable explanation of the agent's decision
	ExplainDecision(ctx context.Context, matchResult map[string]interface{}) (string, error)

	// ValidateWithLLM performs a semantic LLM-powered validation pass on the invoice
	ValidateWithLLM(ctx context.Context, invoiceFields map[string]interface{}, poDetails map[string]interface{}, ruleCheckSummary string) (*LLMValidationResult, error)
}

// ── Storage Service Interface ───────────────
// Implemented by S3 client and local file system
type StorageService interface {
	// UploadFile uploads a file and returns the URL
	UploadFile(ctx context.Context, key string, data []byte, contentType string) (string, error)

	// GetFile retrieves a file by key
	GetFile(ctx context.Context, key string) ([]byte, error)

	// GetPresignedURL generates a temporary download URL
	GetPresignedURL(ctx context.Context, key string) (string, error)
}

// ── Email Service Interface ─────────────────
// Implemented by SES client and mock logger
type EmailService interface {
	// SendEmail sends an email
	SendEmail(ctx context.Context, to, subject, body string) error
}

// ── Payment Service Interface ───────────────
// Implemented by Pine Labs client and mock client
type PaymentService interface {
	// InitiateDisbursement starts a B2B payment
	InitiateDisbursement(ctx context.Context, req DisbursementRequest) (*DisbursementResponse, error)

	// GetTransactionStatus checks status of a payment
	GetTransactionStatus(ctx context.Context, transactionID string) (*TransactionStatus, error)

	// CreatePaymentLink creates a Pine Labs payment link for an invoice
	CreatePaymentLink(ctx context.Context, req PaymentLinkRequest) (*PaymentLinkResponse, error)
}

// ── Payment Service DTOs ────────────────────
type DisbursementRequest struct {
	InvoiceID     string  `json:"invoice_id"`
	VendorName    string  `json:"vendor_name"`
	AccountNumber string  `json:"account_number"`
	IFSC          string  `json:"ifsc"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Reference     string  `json:"reference"`
}

type DisbursementResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	Message       string `json:"message"`
}

type TransactionStatus struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"` // PENDING, SUCCESS, FAILED
	Timestamp     string `json:"timestamp"`
}
