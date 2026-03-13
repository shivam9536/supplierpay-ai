package models

import "time"

// ── Cash Flow Forecast ──────────────────────
type CashFlowPeriod struct {
	PeriodStart       time.Time `json:"period_start"`
	PeriodEnd         time.Time `json:"period_end"`
	ScheduledOutflows float64   `json:"scheduled_outflows"`
	InvoiceCount      int       `json:"invoice_count"`
	ProjectedBalance  float64   `json:"projected_balance"`
	RiskLevel         string    `json:"risk_level"` // LOW, MEDIUM, HIGH
}

type CashFlowForecast struct {
	GeneratedAt     time.Time        `json:"generated_at"`
	ForecastDays    int              `json:"forecast_days"`
	TotalOutflows   float64          `json:"total_outflows"`
	Periods         []CashFlowPeriod `json:"periods"`
	RiskFlags       []string         `json:"risk_flags"`
	StartingBalance float64          `json:"starting_balance"`
}

// ── API Response Wrappers ───────────────────
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type PaginatedResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Total   int         `json:"total"`
	Page    int         `json:"page"`
	PerPage int         `json:"per_page"`
}

// ── SSE Event ───────────────────────────────
type SSEEvent struct {
	InvoiceID string      `json:"invoice_id"`
	Step      string      `json:"step"`
	Status    string      `json:"status"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
}
