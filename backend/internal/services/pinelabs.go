package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/supplierpay/backend/internal/config"
	"go.uber.org/zap"
)

// ── Pine Labs token cache ────────────────────

type tokenCache struct {
	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

func (tc *tokenCache) get() (string, bool) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if tc.token == "" || time.Now().After(tc.expiresAt.Add(-30*time.Second)) {
		return "", false
	}
	return tc.token, true
}

func (tc *tokenCache) set(token string, expiresAt time.Time) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.token = token
	tc.expiresAt = expiresAt
}

// ── Pine Labs Payment Client ────────────────

type PineLabsClient struct {
	cfg    *config.Config
	logger *zap.Logger
	http   *http.Client
	cache  tokenCache
}

func NewPineLabsClient(cfg *config.Config, logger *zap.Logger) *PineLabsClient {
	return &PineLabsClient{
		cfg:    cfg,
		logger: logger,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

// ── Token management ─────────────────────────

type tokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   string `json:"expires_at"`
}

func (p *PineLabsClient) getAccessToken(ctx context.Context) (string, error) {
	if tok, ok := p.cache.get(); ok {
		return tok, nil
	}

	reqBody := tokenRequest{
		ClientID:     p.cfg.PineLabsClientID,
		ClientSecret: p.cfg.PineLabsClientSecret,
		GrantType:    "client_credentials",
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		p.cfg.PineLabsAPIURL+"/auth/v1/token", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build token request: %w", err)
	}
	p.setCommonHeaders(req, "")

	resp, err := p.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token API %d: %s", resp.StatusCode, string(respBytes))
	}

	var tr tokenResponse
	if err := json.Unmarshal(respBytes, &tr); err != nil {
		return "", fmt.Errorf("parse token response: %w", err)
	}

	expiresAt := time.Now().Add(55 * time.Minute) // conservative default
	if tr.ExpiresAt != "" {
		if t, err := time.Parse(time.RFC3339Nano, tr.ExpiresAt); err == nil {
			expiresAt = t
		}
	}
	p.cache.set(tr.AccessToken, expiresAt)
	p.logger.Info("Pine Labs: access token refreshed", zap.Time("expires_at", expiresAt))
	return tr.AccessToken, nil
}

// ── Payout request / response types ─────────

type pineLabsPayoutRequest struct {
	MerchantPaymentReference string             `json:"merchant_payment_reference"`
	PaymentAmount            pineLabsAmount     `json:"payment_amount"`
	PaymentMethod            string             `json:"payment_method"`
	BankAccountDetails       pineLabsBankDetail `json:"bank_account_details"`
	Remarks                  string             `json:"remarks,omitempty"`
	ScheduledAt              string             `json:"scheduled_at,omitempty"`
}

type pineLabsAmount struct {
	Value    int64  `json:"value"` // in paise (smallest unit)
	Currency string `json:"currency"`
}

type pineLabsBankDetail struct {
	AccountNumber string `json:"account_number"`
	IFSC          string `json:"ifsc"`
	AccountName   string `json:"account_name"`
}

type pineLabsPayoutResponse struct {
	PaymentID                string `json:"payment_id"`
	MerchantPaymentReference string `json:"merchant_payment_reference"`
	Status                   string `json:"status"`
	Message                  string `json:"message,omitempty"`
	CreatedAt                string `json:"created_at,omitempty"`
}

type pineLabsStatusResponse struct {
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// ── InitiateDisbursement ─────────────────────

func (p *PineLabsClient) InitiateDisbursement(ctx context.Context, req DisbursementRequest) (*DisbursementResponse, error) {
	p.logger.Info("Pine Labs: initiating disbursement",
		zap.String("invoice_id", req.InvoiceID),
		zap.Float64("amount", req.Amount),
		zap.String("vendor", req.VendorName),
	)

	token, err := p.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("Pine Labs auth: %w", err)
	}

	// Convert amount to paise (INR smallest unit)
	amountPaise := int64(req.Amount * 100)

	payoutReq := pineLabsPayoutRequest{
		MerchantPaymentReference: req.Reference,
		PaymentAmount: pineLabsAmount{
			Value:    amountPaise,
			Currency: req.Currency,
		},
		PaymentMethod: "NEFT",
		BankAccountDetails: pineLabsBankDetail{
			AccountNumber: req.AccountNumber,
			IFSC:          req.IFSC,
			AccountName:   req.VendorName,
		},
		Remarks: fmt.Sprintf("Invoice payment: %s", req.InvoiceID),
	}

	body, _ := json.Marshal(payoutReq)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		p.cfg.PineLabsAPIURL+"/payouts/v3/payments/banks", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build payout request: %w", err)
	}
	p.setCommonHeaders(httpReq, token)

	resp, err := p.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("payout request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	p.logger.Info("Pine Labs payout response",
		zap.Int("status_code", resp.StatusCode),
		zap.String("body", string(respBytes)),
	)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("Pine Labs payout API %d: %s", resp.StatusCode, string(respBytes))
	}

	var pr pineLabsPayoutResponse
	if err := json.Unmarshal(respBytes, &pr); err != nil {
		return nil, fmt.Errorf("parse payout response: %w", err)
	}

	return &DisbursementResponse{
		TransactionID: pr.PaymentID,
		Status:        pr.Status,
		Message:       pr.Message,
	}, nil
}

// ── GetTransactionStatus ─────────────────────

func (p *PineLabsClient) GetTransactionStatus(ctx context.Context, transactionID string) (*TransactionStatus, error) {
	p.logger.Info("Pine Labs: checking transaction status", zap.String("transaction_id", transactionID))

	token, err := p.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("Pine Labs auth: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
		p.cfg.PineLabsAPIURL+"/payouts/v3/payments/"+transactionID, nil)
	if err != nil {
		return nil, fmt.Errorf("build status request: %w", err)
	}
	p.setCommonHeaders(httpReq, token)

	resp, err := p.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("status request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Pine Labs status API %d: %s", resp.StatusCode, string(respBytes))
	}

	var sr pineLabsStatusResponse
	if err := json.Unmarshal(respBytes, &sr); err != nil {
		return nil, fmt.Errorf("parse status response: %w", err)
	}

	return &TransactionStatus{
		TransactionID: sr.PaymentID,
		Status:        normalisePineLabsStatus(sr.Status),
		Timestamp:     sr.UpdatedAt,
	}, nil
}

// ── Payment Link types ───────────────────────

type paymentLinkAmount struct {
	Value    int64  `json:"value"`
	Currency string `json:"currency"`
}

type paymentLinkCustomer struct {
	EmailID string `json:"email_id,omitempty"`
}

type paymentLinkAPIRequest struct {
	Amount                       paymentLinkAmount   `json:"amount"`
	Description                  string              `json:"description,omitempty"`
	MerchantPaymentLinkReference string              `json:"merchant_payment_link_reference"`
	Customer                     paymentLinkCustomer `json:"customer,omitempty"`
	ExpireBy                     string              `json:"expire_by,omitempty"`
}

// PaymentLinkRequest is the public DTO used by callers.
type PaymentLinkRequest struct {
	AmountValue                  int64
	Currency                     string
	Description                  string
	MerchantPaymentLinkReference string
	CustomerEmail                string
	ExpireBy                     string
}

type PaymentLinkResponse struct {
	PaymentLinkID  string `json:"payment_link_id"`
	PaymentLinkURL string `json:"payment_link"`     // API returns URL in "payment_link" field
	Status         string `json:"status"`
	Message        string `json:"message,omitempty"`
}

// CreatePaymentLink creates a Pine Labs payment link via the REST API.
// Endpoint: POST /api/pay/v1/paymentlink
func (p *PineLabsClient) CreatePaymentLink(ctx context.Context, req PaymentLinkRequest) (*PaymentLinkResponse, error) {
	currency := req.Currency
	if currency == "" {
		currency = "INR"
	}

	p.logger.Info("Pine Labs: creating payment link",
		zap.String("reference", req.MerchantPaymentLinkReference),
		zap.Int64("amount_paise", req.AmountValue),
	)

	token, err := p.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("Pine Labs auth: %w", err)
	}

	apiReq := paymentLinkAPIRequest{
		Amount: paymentLinkAmount{
			Value:    req.AmountValue,
			Currency: currency,
		},
		Description:                  req.Description,
		MerchantPaymentLinkReference: req.MerchantPaymentLinkReference,
		Customer: paymentLinkCustomer{
			EmailID: req.CustomerEmail,
		},
		ExpireBy: req.ExpireBy,
	}

	body, _ := json.Marshal(apiReq)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		p.cfg.PineLabsAPIURL+"/pay/v1/paymentlink", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build payment link request: %w", err)
	}
	p.setCommonHeaders(httpReq, token)

	resp, err := p.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("payment link request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	p.logger.Info("Pine Labs payment link response",
		zap.Int("status_code", resp.StatusCode),
		zap.String("body", string(respBytes)),
	)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Pine Labs payment link API %d: %s", resp.StatusCode, string(respBytes))
	}

	var plr PaymentLinkResponse
	if err := json.Unmarshal(respBytes, &plr); err != nil {
		return nil, fmt.Errorf("parse payment link response: %w", err)
	}

	return &plr, nil
}

// ── VerifyWebhookSignature ───────────────────

// VerifyWebhookSignature validates the HMAC-SHA256 signature on incoming Pine Labs webhooks.
// Pine Labs sends the signature in the X-Signature header as a hex-encoded HMAC-SHA256
// of the raw request body using the webhook secret.
func (p *PineLabsClient) VerifyWebhookSignature(payload []byte, signature string) bool {
	if p.cfg.PineLabsWebhookSecret == "" {
		p.logger.Warn("Pine Labs webhook secret not configured — skipping signature verification")
		return true
	}
	mac := hmac.New(sha256.New, []byte(p.cfg.PineLabsWebhookSecret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// ── helpers ──────────────────────────────────

func (p *PineLabsClient) setCommonHeaders(req *http.Request, bearerToken string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Request-ID", uuid.New().String())
	req.Header.Set("Request-Timestamp", time.Now().UTC().Format(time.RFC3339Nano))
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}
}

// normalisePineLabsStatus maps Pine Labs status strings to our internal values.
func normalisePineLabsStatus(s string) string {
	switch s {
	case "SUCCESS", "COMPLETED", "PROCESSED":
		return "SUCCESS"
	case "FAILED", "REJECTED", "CANCELLED":
		return "FAILED"
	default:
		return "PENDING"
	}
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
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (m *MockPaymentClient) CreatePaymentLink(ctx context.Context, req PaymentLinkRequest) (*PaymentLinkResponse, error) {
	linkID := uuid.New().String()
	m.logger.Info("[MOCK] Payment link created",
		zap.String("link_id", linkID),
		zap.String("reference", req.MerchantPaymentLinkReference),
	)
	return &PaymentLinkResponse{
		PaymentLinkID:  linkID,
		PaymentLinkURL: "https://pay.pinelabs.com/mock/" + linkID,
		Status:         "CREATED",
	}, nil
}
