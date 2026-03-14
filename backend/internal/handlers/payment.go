package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/supplierpay/backend/internal/config"
	"github.com/supplierpay/backend/internal/models"
	"go.uber.org/zap"
)

type PaymentHandler struct {
	db     *sqlx.DB
	cfg    *config.Config
	logger *zap.Logger
}

func NewPaymentHandler(db *sqlx.DB, cfg *config.Config, logger *zap.Logger) *PaymentHandler {
	return &PaymentHandler{db: db, cfg: cfg, logger: logger}
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

// TriggerRun manually triggers a payment run
func (h *PaymentHandler) TriggerRun(c *gin.Context) {
	h.logger.Info("Manual payment run triggered")

	// TODO: Dev 1 — Execute payment run via Pine Labs
	// 1. Get all SCHEDULED invoices for today
	// 2. Batch call Pine Labs disbursement API
	// 3. Update invoice statuses
	// 4. Create PaymentRun record

	c.JSON(http.StatusAccepted, models.APIResponse{
		Success: true,
		Message: "Payment run initiated",
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

// PineLabsWebhook handles payment confirmation from Pine Labs
func (h *PaymentHandler) PineLabsWebhook(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	h.logger.Info("Pine Labs webhook received", zap.Any("payload", payload))

	// TODO: Dev 1 — Validate webhook signature, update invoice status to PAID
	c.JSON(http.StatusOK, gin.H{"received": true})
}
