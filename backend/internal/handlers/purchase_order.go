package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/supplierpay/backend/internal/models"
	"go.uber.org/zap"
)

type PurchaseOrderHandler struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewPurchaseOrderHandler(db *sqlx.DB, logger *zap.Logger) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{db: db, logger: logger}
}

// List returns all purchase orders (optional filter by vendor_id, status)
func (h *PurchaseOrderHandler) List(c *gin.Context) {
	vendorID := c.Query("vendor_id")
	status := c.Query("status")

	query := `SELECT id, po_number, vendor_id, total_value, remaining_value, line_items, approved_by, status, created_at, updated_at
		FROM purchase_orders
		WHERE ($1::text = '' OR vendor_id::text = $1) AND ($2::text = '' OR status = $2)
		ORDER BY created_at DESC`
	var pos []models.PurchaseOrder
	err := h.db.Select(&pos, query, vendorID, status)
	if err != nil {
		h.logger.Error("Failed to list POs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Failed to list purchase orders",
		})
		return
	}
	if pos == nil {
		pos = []models.PurchaseOrder{}
	}
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    pos,
	})
}

// GetByID returns a single purchase order
func (h *PurchaseOrderHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Invalid PO ID",
		})
		return
	}
	var po models.PurchaseOrder
	err = h.db.Get(&po, `SELECT id, po_number, vendor_id, total_value, remaining_value, line_items, approved_by, status, created_at, updated_at
		FROM purchase_orders WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false, Error: "Purchase order not found",
		})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    po,
	})
}

// Create adds a new purchase order
func (h *PurchaseOrderHandler) Create(c *gin.Context) {
	var po models.PurchaseOrder
	if err := c.ShouldBindJSON(&po); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Invalid PO data",
		})
		return
	}
	po.ID = uuid.New()
	lineItemsJSON := "[]"
	if po.LineItems.Data != nil {
		b, _ := json.Marshal(po.LineItems.Data)
		lineItemsJSON = string(b)
	}
	_, err := h.db.Exec(`INSERT INTO purchase_orders (id, po_number, vendor_id, total_value, remaining_value, line_items, approved_by, status)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8)`,
		po.ID, po.PONumber, po.VendorID, po.TotalValue, po.RemainingValue, lineItemsJSON, po.ApprovedBy, po.Status)
	if err != nil {
		h.logger.Error("Failed to create PO", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Failed to create purchase order",
		})
		return
	}
	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Purchase order created",
		Data:    po,
	})
}
