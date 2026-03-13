package scheduler

import (
	"github.com/jmoiron/sqlx"
	"github.com/robfig/cron/v3"
	"github.com/supplierpay/backend/internal/config"
	"go.uber.org/zap"
)

// PaymentScheduler runs nightly payment jobs
type PaymentScheduler struct {
	db     *sqlx.DB
	cfg    *config.Config
	logger *zap.Logger
	cron   *cron.Cron
}

func NewPaymentScheduler(db *sqlx.DB, cfg *config.Config, logger *zap.Logger) *PaymentScheduler {
	return &PaymentScheduler{
		db:     db,
		cfg:    cfg,
		logger: logger,
		cron:   cron.New(),
	}
}

func (s *PaymentScheduler) Start() {
	// Run every night at midnight
	s.cron.AddFunc("0 0 * * *", s.runPayments)

	// Run cash flow forecast update every 6 hours
	s.cron.AddFunc("0 */6 * * *", s.updateForecast)

	s.cron.Start()
	s.logger.Info("Payment scheduler started",
		zap.String("payment_schedule", "daily at 00:00"),
		zap.String("forecast_schedule", "every 6 hours"),
	)
}

func (s *PaymentScheduler) Stop() {
	s.cron.Stop()
	s.logger.Info("Payment scheduler stopped")
}

func (s *PaymentScheduler) runPayments() {
	s.logger.Info("Running nightly payment batch")

	// TODO: Dev 1 — Implement payment run
	// 1. SELECT * FROM invoices WHERE scheduled_payment_date = CURRENT_DATE AND status = 'SCHEDULED'
	// 2. Group by vendor for batch processing
	// 3. Call Pine Labs disbursement API for each batch
	// 4. Update invoice statuses
	// 5. Create PaymentRun record
	// 6. Handle failures with retry logic
}

func (s *PaymentScheduler) updateForecast() {
	s.logger.Info("Updating cash flow forecast")

	// TODO: Dev 3 — Regenerate 90-day forecast
	// 1. Query all APPROVED + SCHEDULED invoices
	// 2. Group payments by week
	// 3. Calculate running balance
	// 4. Detect risk periods
}
