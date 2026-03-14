package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/supplierpay/backend/internal/agent"
	"github.com/supplierpay/backend/internal/config"
	"github.com/supplierpay/backend/internal/models"
	"github.com/supplierpay/backend/internal/services"
	"go.uber.org/zap"
)

type InvoiceHandler struct {
	db      *sqlx.DB
	cfg     *config.Config
	logger  *zap.Logger
	orch    *agent.Orchestrator
	storage services.StorageService
}

func NewInvoiceHandler(db *sqlx.DB, cfg *config.Config, logger *zap.Logger, orch *agent.Orchestrator, storage services.StorageService) *InvoiceHandler {
	return &InvoiceHandler{db: db, cfg: cfg, logger: logger, orch: orch, storage: storage}
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
	vendorIDStr := c.PostForm("vendor_id")
	var vendorID *uuid.UUID
	if vendorIDStr != "" {
		parsed, err := uuid.Parse(vendorIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false, Error: "Invalid vendor_id",
			})
			return
		}
		vendorID = &parsed
	}
	poReference := c.PostForm("po_reference")

	f, err := file.Open()
	if err != nil {
		h.logger.Error("Failed to open uploaded file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Failed to read file",
		})
		return
	}
	defer f.Close()
	// Read file content for storage
	data := make([]byte, file.Size)
	if _, err := f.Read(data); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Failed to read file",
		})
		return
	}

	invoiceID := uuid.New()
	key := fmt.Sprintf("invoices/%s%s", invoiceID.String(), filepath.Ext(file.Filename))
	ctx := context.Background()
	rawURL, err := h.storage.UploadFile(ctx, key, data, file.Header.Get("Content-Type"))
	if err != nil {
		h.logger.Error("Failed to upload file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Failed to store file",
		})
		return
	}

	_, err = h.db.Exec(`INSERT INTO invoices (id, vendor_id, invoice_number, po_reference, raw_file_url, status)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		invoiceID, vendorID, "", poReference, rawURL, models.InvoiceStatusPending)

	if err != nil {
		h.logger.Error("Failed to create invoice record", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Failed to create invoice",
		})
		return
	}

	go func() {
		if err := h.orch.ProcessInvoice(context.Background(), invoiceID); err != nil {
			h.logger.Error("Agent pipeline failed", zap.String("invoice_id", invoiceID.String()), zap.Error(err))
		}
	}()

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
	vendorIDRaw, ok := payload["vendor_id"]
	if !ok {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "vendor_id is required",
		})
		return
	}
	vendorIDStr, _ := vendorIDRaw.(string)
	vendorID, err := uuid.Parse(vendorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Invalid vendor_id",
		})
		return
	}

	invoiceID := uuid.New()
	invNum, _ := payload["invoice_number"].(string)
	poRef, _ := payload["po_reference"].(string)
	totalAmt := float64(0)
	if v, ok := payload["total_amount"].(float64); ok {
		totalAmt = v
	}
	taxAmt := float64(0)
	if v, ok := payload["tax_amount"].(float64); ok {
		taxAmt = v
	}
	currency, _ := payload["currency"].(string)
	if currency == "" {
		currency = "INR"
	}
	extractedJSON := "{}"
	lineItemsJSON := "[]"
	if b, err := json.Marshal(payload); err == nil {
		extractedJSON = string(b)
	}
	if li, ok := payload["line_items"]; ok {
		if b, err := json.Marshal(li); err == nil {
			lineItemsJSON = string(b)
		}
	}
	var invDate, dueDate *time.Time
	if s, ok := payload["invoice_date"].(string); ok && s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			invDate = &t
		}
	}
	if s, ok := payload["due_date"].(string); ok && s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			dueDate = &t
		}
	}

	query := `INSERT INTO invoices (id, vendor_id, invoice_number, po_reference, extracted_fields, line_items, total_amount, tax_amount, currency, invoice_date, due_date, status)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7, $8, $9, $10, $11, $12)`
	_, err = h.db.Exec(query, invoiceID, vendorID, invNum, poRef, extractedJSON, lineItemsJSON, totalAmt, taxAmt, currency, invDate, dueDate, models.InvoiceStatusPending)
	if err != nil {
		h.logger.Error("Failed to create invoice from JSON", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Failed to create invoice",
		})
		return
	}

	go func() {
		if err := h.orch.ProcessInvoice(context.Background(), invoiceID); err != nil {
			h.logger.Error("Agent pipeline failed", zap.String("invoice_id", invoiceID.String()), zap.Error(err))
		}
	}()

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

	query := `SELECT i.id, i.vendor_id, i.invoice_number, i.po_reference, i.raw_file_url, i.extracted_fields, i.line_items,
		i.total_amount, i.tax_amount, i.currency, i.invoice_date, i.due_date, i.status, i.discrepancies, i.decision_reason,
		i.scheduled_payment_date, i.pinelabs_transaction_id, i.created_at, i.updated_at, v.name as vendor_name
		FROM invoices i LEFT JOIN vendors v ON i.vendor_id = v.id
		WHERE ($1::text = '' OR i.status = $1) AND ($2::text = '' OR i.vendor_id::text = $2)
		ORDER BY i.created_at DESC`
	var list []models.InvoiceWithVendor
	err := h.db.Select(&list, query, status, vendorID)
	if err != nil {
		h.logger.Error("Failed to list invoices", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Failed to list invoices",
		})
		return
	}
	if list == nil {
		list = []models.InvoiceWithVendor{}
	}
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    list,
	})
}

// GetByID returns a single invoice with full details
func (h *InvoiceHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Invalid invoice ID",
		})
		return
	}
	var inv models.Invoice
	err = h.db.Get(&inv, `SELECT id, vendor_id, invoice_number, po_reference, raw_file_url, extracted_fields, line_items,
		total_amount, tax_amount, currency, invoice_date, due_date, status, discrepancies, decision_reason,
		scheduled_payment_date, pinelabs_transaction_id, created_at, updated_at
		FROM invoices WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false, Error: "Invoice not found",
		})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    inv,
	})
}

// GetAuditLog returns the agent processing audit trail for an invoice
func (h *InvoiceHandler) GetAuditLog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Invalid invoice ID",
		})
		return
	}
	var auditLogs []models.AuditLog
	err = h.db.Select(&auditLogs, `SELECT id, invoice_id, step, result, reasoning, confidence_score, duration_ms, created_at
		FROM audit_logs WHERE invoice_id = $1 ORDER BY created_at ASC`, id)
	if err != nil {
		h.logger.Error("Failed to get audit log", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Failed to get audit log",
		})
		return
	}
	if auditLogs == nil {
		auditLogs = []models.AuditLog{}
	}
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    auditLogs,
	})
}

// Reprocess re-triggers the agent pipeline for a flagged invoice
func (h *InvoiceHandler) Reprocess(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Invalid invoice ID",
		})
		return
	}
	result, err := h.db.Exec(`UPDATE invoices SET status = $1, decision_reason = '', discrepancies = '[]'::jsonb WHERE id = $2`, models.InvoiceStatusPending, id)
	if err != nil {
		h.logger.Error("Failed to reset invoice for reprocess", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Failed to reprocess",
		})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false, Error: "Invoice not found",
		})
		return
	}
	go func() {
		if err := h.orch.ProcessInvoice(context.Background(), id); err != nil {
			h.logger.Error("Agent pipeline failed on reprocess", zap.String("invoice_id", id.String()), zap.Error(err))
		}
	}()
	c.JSON(http.StatusAccepted, models.APIResponse{
		Success: true,
		Message: "Invoice reprocessing started",
	})
}
