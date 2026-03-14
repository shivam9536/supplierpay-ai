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

// GetForecast returns 90-day cash flow forecast from APPROVED + SCHEDULED invoices
func (h *ForecastHandler) GetForecast(c *gin.Context) {
	h.logger.Info("Generating cash flow forecast")

	type row struct {
		PeriodStart string  `db:"period_start"`
		Total       float64 `db:"total"`
		Count       int     `db:"cnt"`
	}
	var rows []row
	err := h.db.Select(&rows, `
		SELECT date_trunc('week', scheduled_payment_date)::date AS period_start,
		       COALESCE(SUM(total_amount + tax_amount), 0) AS total,
		       COUNT(*)::int AS cnt
		FROM invoices
		WHERE status IN ($1, $2) AND scheduled_payment_date IS NOT NULL
		  AND scheduled_payment_date >= CURRENT_DATE
		  AND scheduled_payment_date <= CURRENT_DATE + 90
		GROUP BY date_trunc('week', scheduled_payment_date)
		ORDER BY period_start`,
		models.InvoiceStatusApproved, models.InvoiceStatusScheduled)
	if err != nil {
		h.logger.Error("Failed to get forecast", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Failed to generate forecast",
		})
		return
	}

	periods := make([]models.CashFlowPeriod, 0, len(rows))
	var totalOutflows float64
	balance := 1000000.0
	var riskFlags []string
	for _, r := range rows {
		totalOutflows += r.Total
		balance -= r.Total
		tStart := parseDate(r.PeriodStart)
		tEnd := tStart.AddDate(0, 0, 6)
		riskLevel := "LOW"
		if balance < 100000 {
			riskLevel = "HIGH"
			riskFlags = append(riskFlags, "Low projected balance in week starting "+r.PeriodStart)
		} else if balance < 300000 {
			riskLevel = "MEDIUM"
		}
		periods = append(periods, models.CashFlowPeriod{
			PeriodStart:       tStart,
			PeriodEnd:         tEnd,
			ScheduledOutflows: r.Total,
			InvoiceCount:      r.Count,
			ProjectedBalance:  balance,
			RiskLevel:         riskLevel,
		})
	}

	forecast := models.CashFlowForecast{
		GeneratedAt:     time.Now(),
		ForecastDays:    90,
		TotalOutflows:   totalOutflows,
		StartingBalance: 1000000,
		Periods:         periods,
		RiskFlags:       riskFlags,
	}
	if forecast.RiskFlags == nil {
		forecast.RiskFlags = []string{}
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    forecast,
	})
}

func parseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}
