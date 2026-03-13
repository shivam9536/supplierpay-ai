package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/supplierpay/backend/internal/config"
	"github.com/supplierpay/backend/internal/models"
	"go.uber.org/zap"
)

type InvoiceHandler struct {
	db     *sqlx.DB
	cfg    *config.Config
	logger *zap.Logger
}

func NewInvoiceHandler(db *sqlx.DB, cfg *config.Config, logger *zap.Logger) *InvoiceHandler {
	return &InvoiceHandler{db: db, cfg: cfg, logger: logger}
}

// Upload handles PDF invoice upload
func (h *InvoiceHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("invoice")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "No file provided",
		})
		return
	}

	vendorID := c.PostForm("vendor_id")
	poReference := c.PostForm("po_reference")

	h.logger.Info("Invoice upload received",
		zap.String("filename", file.Filename),
		zap.String("vendor_id", vendorID),
		zap.String("po_reference", poReference),
	)

	// TODO: Dev 1 — Upload file to S3, create invoice record, trigger agent pipeline
	// 1. Upload to S3 → get raw_file_url
	// 2. Insert invoice with status PENDING
	// 3. Trigger agent orchestrator (async goroutine)
	// 4. Return invoice ID for SSE tracking

	invoiceID := uuid.New()

	c.JSON(http.StatusAccepted, models.APIResponse{
		Success: true,
		Message: "Invoice uploaded and processing started",
		Data: gin.H{
			"invoice_id": invoiceID,
			"status":     models.InvoiceStatusPending,
			"filename":   file.Filename,
		},
	})
}

// UploadJSON handles JSON invoice payload (hackathon demo path)
func (h *InvoiceHandler) UploadJSON(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Invalid JSON payload",
		})
		return
	}

	h.logger.Info("JSON invoice received", zap.Any("payload", payload))

	// TODO: Dev 2 — Create invoice from JSON, skip S3, trigger agent pipeline
	invoiceID := uuid.New()

	c.JSON(http.StatusAccepted, models.APIResponse{
		Success: true,
		Message: "Invoice received and processing started",
		Data: gin.H{
			"invoice_id": invoiceID,
			"status":     models.InvoiceStatusPending,
		},
	})
}

// List returns all invoices with optional filters
func (h *InvoiceHandler) List(c *gin.Context) {
	status := c.Query("status")
	vendorID := c.Query("vendor_id")

	h.logger.Info("Listing invoices", zap.String("status", status), zap.String("vendor_id", vendorID))

	// TODO: Dev 3 — Query invoices with filters, pagination
	invoices := []models.Invoice{}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    invoices,
	})
}

// GetByID returns a single invoice with full details
func (h *InvoiceHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	h.logger.Info("Getting invoice", zap.String("id", id))

	// TODO: Dev 3 — Fetch invoice by ID with vendor details
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    nil,
	})
}

// GetAuditLog returns the agent processing audit trail for an invoice
func (h *InvoiceHandler) GetAuditLog(c *gin.Context) {
	id := c.Param("id")

	h.logger.Info("Getting audit log", zap.String("invoice_id", id))

	// TODO: Dev 2 — Fetch audit logs for invoice
	auditLogs := []models.AuditLog{}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    auditLogs,
	})
}

// Reprocess re-triggers the agent pipeline for a flagged invoice
func (h *InvoiceHandler) Reprocess(c *gin.Context) {
	id := c.Param("id")

	h.logger.Info("Reprocessing invoice", zap.String("id", id))

	// TODO: Dev 2 — Reset status to PENDING, re-trigger agent orchestrator
	c.JSON(http.StatusAccepted, models.APIResponse{
		Success: true,
		Message: "Invoice reprocessing started",
	})
}
