package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
	// TODO: Dev 3 — Query all vendors
	vendors := []models.Vendor{}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    vendors,
	})
}

// GetByID returns a single vendor
func (h *VendorHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	h.logger.Info("Getting vendor", zap.String("id", id))

	// TODO: Dev 3 — Fetch vendor by ID
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    nil,
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

	// TODO: Dev 3 — Insert vendor into DB
	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Vendor created",
		Data:    vendor,
	})
}
