package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/supplierpay/backend/internal/config"
	"github.com/supplierpay/backend/internal/models"
	"github.com/supplierpay/backend/internal/services"
	"go.uber.org/zap"
)

type PaymentHandler struct {
	db      *sqlx.DB
	cfg     *config.Config
	logger  *zap.Logger
	payment services.PaymentService
}

func NewPaymentHandler(db *sqlx.DB, cfg *config.Config, logger *zap.Logger, payment services.PaymentService) *PaymentHandler {
	return &PaymentHandler{db: db, cfg: cfg, logger: logger, payment: payment}
}

// GetSchedule returns upcoming scheduled payments (invoices with status SCHEDULED, scheduled_payment_date >= today)
func (h *PaymentHandler) GetSchedule(c *gin.Context) {
	var invoices []models.Invoice
	err := h.db.Select(&invoices, `SELECT id, vendor_id, invoice_number, po_reference, raw_file_url, extracted_fields, line_items,
		total_amount, tax_amount, currency, invoice_date, due_date, status, discrepancies, decision_reason,
		scheduled_payment_date, pinelabs_transaction_id, created_at, updated_at
		FROM invoices
		WHERE status = $1 AND scheduled_payment_date >= CURRENT_DATE
		ORDER BY scheduled_payment_date ASC`,
		models.InvoiceStatusScheduled)
	if err != nil {
		h.logger.Error("Failed to get payment schedule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Failed to get schedule",
		})
		return
	}
	if invoices == nil {
		invoices = []models.Invoice{}
	}
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    invoices,
	})
}

// TriggerRun manually triggers a payment run for all SCHEDULED invoices due today
func (h *PaymentHandler) TriggerRun(c *gin.Context) {
	h.logger.Info("Manual payment run triggered")
	ctx := c.Request.Context()

	type scheduledInvoice struct {
		ID            uuid.UUID `db:"id"`
		InvoiceNumber *string   `db:"invoice_number"`
		TotalAmount   float64   `db:"total_amount"`
		Currency      *string   `db:"currency"`
		VendorName    string    `db:"vendor_name"`
		AccountNumber string    `db:"bank_account_number"`
		IFSC          string    `db:"bank_ifsc"`
	}

	var invoices []scheduledInvoice
	err := h.db.SelectContext(ctx, &invoices, `
		SELECT i.id,
		       i.invoice_number,
		       i.total_amount,
		       i.currency,
		       v.name            AS vendor_name,
		       v.bank_account_number,
		       v.bank_ifsc
		  FROM invoices i
		  JOIN vendors v ON i.vendor_id = v.id
		 WHERE i.status = 'SCHEDULED'
		   AND i.scheduled_payment_date <= CURRENT_DATE
		   AND i.pinelabs_transaction_id IS NULL
		 ORDER BY i.scheduled_payment_date ASC
	`)
	if err != nil {
		h.logger.Error("Failed to fetch scheduled invoices", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Failed to fetch scheduled invoices",
		})
		return
	}
	if len(invoices) == 0 {
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Message: "No invoices due for payment today",
		})
		return
	}

	runID := uuid.New()
	totalAmount := 0.0
	for _, inv := range invoices {
		totalAmount += inv.TotalAmount
	}

	_, err = h.db.ExecContext(ctx, `
		INSERT INTO payment_runs (id, run_date, total_amount, invoice_count, status)
		VALUES ($1, CURRENT_DATE, $2, $3, 'EXECUTING')`,
		runID, totalAmount, len(invoices),
	)
	if err != nil {
		h.logger.Error("Failed to create payment run", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Failed to create payment run",
		})
		return
	}

	// Disburse asynchronously so the HTTP response is immediate
	go func() {
		successCount := 0
		failCount := 0

		for _, inv := range invoices {
			currency := "INR"
			if inv.Currency != nil && *inv.Currency != "" {
				currency = *inv.Currency
			}
			ref := fmt.Sprintf("SUPPAY-%s", inv.ID.String()[:8])

			disbResp, disbErr := h.payment.InitiateDisbursement(ctx, services.DisbursementRequest{
				InvoiceID:     inv.ID.String(),
				VendorName:    inv.VendorName,
				AccountNumber: inv.AccountNumber,
				IFSC:          inv.IFSC,
				Amount:        inv.TotalAmount,
				Currency:      currency,
				Reference:     ref,
			})

			if disbErr != nil {
				h.logger.Error("Disbursement failed",
					zap.String("invoice_id", inv.ID.String()),
					zap.Error(disbErr),
				)
				failCount++
				continue
			}

			_, _ = h.db.Exec(`
				UPDATE invoices
				   SET status = 'PAID',
				       pinelabs_transaction_id = $1,
				       updated_at = NOW()
				 WHERE id = $2`,
				disbResp.TransactionID, inv.ID,
			)
			h.logger.Info("Invoice paid via manual run",
				zap.String("invoice_id", inv.ID.String()),
				zap.String("transaction_id", disbResp.TransactionID),
			)
			successCount++
		}

		runStatus := "COMPLETED"
		if failCount > 0 {
			runStatus = "PARTIAL_FAILURE"
		}
		_, _ = h.db.Exec(`UPDATE payment_runs SET status = $1, updated_at = NOW() WHERE id = $2`,
			runStatus, runID)

		h.logger.Info("Manual payment run complete",
			zap.String("run_id", runID.String()),
			zap.Int("success", successCount),
			zap.Int("failed", failCount),
		)
	}()

	c.JSON(http.StatusAccepted, models.APIResponse{
		Success: true,
		Message: fmt.Sprintf("Payment run initiated for %d invoice(s) totalling ₹%.2f", len(invoices), totalAmount),
		Data:    gin.H{"run_id": runID, "invoice_count": len(invoices), "total_amount": totalAmount},
	})
}

// ListRuns returns history of payment runs
func (h *PaymentHandler) ListRuns(c *gin.Context) {
	var runs []models.PaymentRun
	err := h.db.Select(&runs, `SELECT id, run_date, total_amount, invoice_count, status, pinelabs_batch_id, created_at, updated_at
		FROM payment_runs ORDER BY run_date DESC LIMIT 50`)
	if err != nil {
		h.logger.Error("Failed to list payment runs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Failed to list runs",
		})
		return
	}
	if runs == nil {
		runs = []models.PaymentRun{}
	}
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    runs,
	})
}

// PineLabsWebhook handles payment confirmation from Pine Labs.
// Pine Labs sends HMAC-SHA256 signature in the X-Signature header.
func (h *PaymentHandler) PineLabsWebhook(c *gin.Context) {
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot read body"})
		return
	}

	// Verify webhook signature when using the real Pine Labs client
	if !h.cfg.MockMode {
		if plClient, ok := h.payment.(*services.PineLabsClient); ok {
			sig := c.GetHeader("X-Signature")
			if sig == "" {
				sig = c.GetHeader("x-signature")
			}
			if !plClient.VerifyWebhookSignature(rawBody, sig) {
				h.logger.Warn("Pine Labs webhook: invalid signature")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
				return
			}
		}
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	h.logger.Info("Pine Labs webhook received", zap.Any("payload", payload))

	// Update invoice status based on Pine Labs payment outcome
	paymentID, _ := payload["payment_id"].(string)
	status, _ := payload["status"].(string)

	if paymentID != "" && status != "" {
		if normalisePineLabsWebhookStatus(status) == "PAID" {
			_, dbErr := h.db.Exec(`
				UPDATE invoices
				   SET status = 'PAID', updated_at = NOW()
				 WHERE pinelabs_transaction_id = $1
				   AND status != 'PAID'`,
				paymentID,
			)
			if dbErr != nil {
				h.logger.Error("Failed to update invoice from webhook",
					zap.String("payment_id", paymentID),
					zap.Error(dbErr),
				)
			} else {
				h.logger.Info("Invoice marked PAID via webhook",
					zap.String("payment_id", paymentID),
				)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"received": true, "timestamp": time.Now().UTC()})
}

func normalisePineLabsWebhookStatus(s string) string {
	switch s {
	case "SUCCESS", "COMPLETED", "PROCESSED":
		return "PAID"
	default:
		return s
	}
}
