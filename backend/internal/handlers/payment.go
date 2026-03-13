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

// GetSchedule returns upcoming scheduled payments
func (h *PaymentHandler) GetSchedule(c *gin.Context) {
	// TODO: Dev 1 — Query invoices with status SCHEDULED, grouped by date
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    []interface{}{},
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
	// TODO: Dev 1 — Query payment runs
	runs := []models.PaymentRun{}

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
