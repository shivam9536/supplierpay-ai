package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/robfig/cron/v3"
	"github.com/supplierpay/backend/internal/agent"
	"github.com/supplierpay/backend/internal/config"
	"go.uber.org/zap"
)

// PaymentScheduler runs nightly payment jobs and polls for pending invoices.
type PaymentScheduler struct {
	db     *sqlx.DB
	cfg    *config.Config
	logger *zap.Logger
	orch   *agent.Orchestrator
	cron   *cron.Cron

	// invoice-poll state
	pollStop chan struct{}
	pollWg   sync.WaitGroup

	// tracks invoices currently being processed so we don't double-dispatch
	inFlight   map[uuid.UUID]struct{}
	inFlightMu sync.Mutex
}

func NewPaymentScheduler(db *sqlx.DB, cfg *config.Config, logger *zap.Logger, orch *agent.Orchestrator) *PaymentScheduler {
	return &PaymentScheduler{
		db:       db,
		cfg:      cfg,
		logger:   logger,
		orch:     orch,
		cron:     cron.New(),
		pollStop: make(chan struct{}),
		inFlight: make(map[uuid.UUID]struct{}),
	}
}

// Start launches all scheduled jobs and the invoice-poll loop.
func (s *PaymentScheduler) Start() {
	// ── Nightly payment run (00:00 every day) ──────────────────────────
	s.cron.AddFunc("0 0 * * *", s.runPayments)

	// ── Cash-flow forecast refresh every 6 hours ──────────────────────
	s.cron.AddFunc("0 */6 * * *", s.updateForecast)

	s.cron.Start()
	s.logger.Info("Payment scheduler started",
		zap.String("payment_schedule", "daily at 00:00"),
		zap.String("forecast_schedule", "every 6 hours"),
	)

	// ── Invoice validation poll every 3 seconds ────────────────────────
	s.pollWg.Add(1)
	go s.pollPendingInvoices()
	s.logger.Info("Invoice poll loop started", zap.Duration("interval", 3*time.Second))
}

// Stop gracefully shuts down all jobs and the poll loop.
func (s *PaymentScheduler) Stop() {
	close(s.pollStop)
	s.pollWg.Wait()
	s.cron.Stop()
	s.logger.Info("Payment scheduler stopped")
}

// ── Invoice Poll Loop ──────────────────────────────────────────────────────

// pollPendingInvoices queries for PENDING invoices every 3 seconds and
// dispatches each one through the agent orchestrator pipeline.
func (s *PaymentScheduler) pollPendingInvoices() {
	defer s.pollWg.Done()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.pollStop:
			s.logger.Info("Invoice poll loop stopped")
			return
		case <-ticker.C:
			s.pickAndProcessPendingInvoice()
		}
	}
}

// stallRecoveryTimeout is how long an invoice can remain in a transient
// processing state (EXTRACTING / VALIDATING) before it is considered stalled
// and eligible for re-processing.
const stallRecoveryTimeout = 2 * time.Minute

// pickAndProcessPendingInvoice fetches the oldest PENDING invoice (one at a
// time, FIFO by created_at) and dispatches it to the agent pipeline.
// It also recovers invoices that got stuck in EXTRACTING or VALIDATING states
// (e.g. due to a server crash) for longer than stallRecoveryTimeout.
// Using SELECT … FOR UPDATE SKIP LOCKED ensures safe concurrent execution if
// multiple scheduler instances ever run side-by-side.
func (s *PaymentScheduler) pickAndProcessPendingInvoice() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Atomically claim one invoice that is either:
	//   a) PENDING (normal path), or
	//   b) stuck in EXTRACTING / VALIDATING for > stallRecoveryTimeout (crash recovery)
	// Reset it to PENDING so the full pipeline runs from the beginning.
	var invoiceID uuid.UUID
	err := s.db.QueryRowContext(ctx, `
		UPDATE invoices
		   SET status = 'EXTRACTING', updated_at = NOW()
		 WHERE id = (
		       SELECT id FROM invoices
		        WHERE status = 'PENDING'
		           OR (status IN ('EXTRACTING', 'VALIDATING')
		               AND updated_at < NOW() - $1::interval)
		        ORDER BY created_at ASC
		        LIMIT 1
		        FOR UPDATE SKIP LOCKED
		 )
		RETURNING id
	`, stallRecoveryTimeout.String()).Scan(&invoiceID)

	if err != nil {
		// No rows → nothing to do; any other error is transient, log and move on.
		return
	}

	// Guard against double-dispatch (belt-and-suspenders)
	s.inFlightMu.Lock()
	if _, busy := s.inFlight[invoiceID]; busy {
		s.inFlightMu.Unlock()
		return
	}
	s.inFlight[invoiceID] = struct{}{}
	s.inFlightMu.Unlock()

	s.logger.Info("Invoice poll: picked up invoice for processing",
		zap.String("invoice_id", invoiceID.String()),
	)

	go func(id uuid.UUID) {
		defer func() {
			s.inFlightMu.Lock()
			delete(s.inFlight, id)
			s.inFlightMu.Unlock()
		}()

		if err := s.orch.ProcessInvoice(context.Background(), id); err != nil {
			s.logger.Error("Invoice pipeline failed",
				zap.String("invoice_id", id.String()),
				zap.Error(err),
			)
		} else {
			s.logger.Info("Invoice pipeline completed",
				zap.String("invoice_id", id.String()),
			)
		}
	}(invoiceID)
}

// ── Nightly / periodic jobs ───────────────────────────────────────────────

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
