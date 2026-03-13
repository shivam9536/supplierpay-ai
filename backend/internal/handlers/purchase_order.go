package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

// List returns all purchase orders
func (h *PurchaseOrderHandler) List(c *gin.Context) {
	vendorID := c.Query("vendor_id")
	status := c.Query("status")

	h.logger.Info("Listing POs", zap.String("vendor_id", vendorID), zap.String("status", status))

	// TODO: Dev 3 — Query POs with filters
	pos := []models.PurchaseOrder{}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    pos,
	})
}

// GetByID returns a single purchase order
func (h *PurchaseOrderHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	h.logger.Info("Getting PO", zap.String("id", id))

	// TODO: Dev 3 — Fetch PO by ID with line items
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    nil,
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

	// TODO: Dev 3 — Insert PO into DB
	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Purchase order created",
		Data:    po,
	})
}
