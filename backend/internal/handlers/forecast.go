package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/supplierpay/backend/internal/models"
	"go.uber.org/zap"
)

type ForecastHandler struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewForecastHandler(db *sqlx.DB, logger *zap.Logger) *ForecastHandler {
	return &ForecastHandler{db: db, logger: logger}
}

// GetForecast returns 90-day cash flow forecast
func (h *ForecastHandler) GetForecast(c *gin.Context) {
	h.logger.Info("Generating cash flow forecast")

	// TODO: Dev 3 — Implement real forecast logic
	// 1. Query all APPROVED/SCHEDULED invoices
	// 2. Group by week
	// 3. Calculate cumulative outflows
	// 4. Flag "crunch weeks"

	forecast := models.CashFlowForecast{
		GeneratedAt:     time.Now(),
		ForecastDays:    90,
		TotalOutflows:   0,
		StartingBalance: 1000000, // Mock starting balance
		Periods:         []models.CashFlowPeriod{},
		RiskFlags:       []string{},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    forecast,
	})
}
