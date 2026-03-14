package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/supplierpay/backend/internal/models"
	"go.uber.org/zap"
)

type VendorHandler struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewVendorHandler(db *sqlx.DB, logger *zap.Logger) *VendorHandler {
	return &VendorHandler{db: db, logger: logger}
}

// List returns all vendors
func (h *VendorHandler) List(c *gin.Context) {
	var vendors []models.Vendor
	err := h.db.Select(&vendors, `SELECT id, name, email, bank_account_number, bank_ifsc,
		payment_terms_days, early_payment_discount, early_payment_days, criticality_score, created_at, updated_at
		FROM vendors ORDER BY name`)
	if err != nil {
		h.logger.Error("Failed to list vendors", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Failed to list vendors",
		})
		return
	}
	if vendors == nil {
		vendors = []models.Vendor{}
	}
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    vendors,
	})
}

// GetByID returns a single vendor
func (h *VendorHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Invalid vendor ID",
		})
		return
	}
	var vendor models.Vendor
	err = h.db.Get(&vendor, `SELECT id, name, email, bank_account_number, bank_ifsc,
		payment_terms_days, early_payment_discount, early_payment_days, criticality_score, created_at, updated_at
		FROM vendors WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false, Error: "Vendor not found",
		})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    vendor,
	})
}

// Create adds a new vendor
func (h *VendorHandler) Create(c *gin.Context) {
	var vendor models.Vendor
	if err := c.ShouldBindJSON(&vendor); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Invalid vendor data",
		})
		return
	}
	vendor.ID = uuid.New()
	_, err := h.db.Exec(`INSERT INTO vendors (id, name, email, bank_account_number, bank_ifsc,
		payment_terms_days, early_payment_discount, early_payment_days, criticality_score)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		vendor.ID, vendor.Name, vendor.Email, vendor.BankAccountNumber, vendor.BankIFSC,
		vendor.PaymentTermsDays, vendor.EarlyPaymentDiscount, vendor.EarlyPaymentDays, vendor.CriticalityScore)
	if err != nil {
		h.logger.Error("Failed to create vendor", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Failed to create vendor",
		})
		return
	}
	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Vendor created",
		Data:    vendor,
	})
}
