package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/robfig/cron/v3"
	"github.com/supplierpay/backend/internal/agent"
	"github.com/supplierpay/backend/internal/config"
	"github.com/supplierpay/backend/internal/services"
	"go.uber.org/zap"
)

// PaymentScheduler runs nightly payment jobs and polls for pending invoices.
type PaymentScheduler struct {
	db      *sqlx.DB
	cfg     *config.Config
	logger  *zap.Logger
	orch    *agent.Orchestrator
	payment services.PaymentService
	cron    *cron.Cron

	// invoice-poll state
	pollStop chan struct{}
	pollWg   sync.WaitGroup

	// tracks invoices currently being processed so we don't double-dispatch
	inFlight   map[uuid.UUID]struct{}
	inFlightMu sync.Mutex
}

func NewPaymentScheduler(db *sqlx.DB, cfg *config.Config, logger *zap.Logger, orch *agent.Orchestrator, payment services.PaymentService) *PaymentScheduler {
	return &PaymentScheduler{
		db:       db,
		cfg:      cfg,
		logger:   logger,
		orch:     orch,
		payment:  payment,
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

// pickAndProcessPendingInvoice fetches the oldest PENDING invoice (one at a
// time, FIFO by created_at) and dispatches it to the agent pipeline.
// Using SELECT … FOR UPDATE SKIP LOCKED ensures safe concurrent execution if
// multiple scheduler instances ever run side-by-side.
func (s *PaymentScheduler) pickAndProcessPendingInvoice() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Atomically claim exactly one PENDING invoice and transition it to
	// EXTRACTING so no other poller picks it up.
	var invoiceID uuid.UUID
	err := s.db.QueryRowContext(ctx, `
		UPDATE invoices
		   SET status = 'EXTRACTING', updated_at = NOW()
		 WHERE id = (
		       SELECT id FROM invoices
		        WHERE status = 'PENDING'
		        ORDER BY created_at ASC
		        LIMIT 1
		        FOR UPDATE SKIP LOCKED
		 )
		RETURNING id
	`).Scan(&invoiceID)

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

// scheduledInvoice holds the minimal fields needed for a payment run.
type scheduledInvoice struct {
	ID            uuid.UUID `db:"id"`
	InvoiceNumber *string   `db:"invoice_number"`
	TotalAmount   float64   `db:"total_amount"`
	Currency      *string   `db:"currency"`
	VendorName    string    `db:"vendor_name"`
	AccountNumber string    `db:"bank_account_number"`
	IFSC          string    `db:"bank_ifsc"`
}

func (s *PaymentScheduler) runPayments() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	s.logger.Info("Running nightly payment batch")

	// 1. Fetch all SCHEDULED invoices due today
	var invoices []scheduledInvoice
	err := s.db.SelectContext(ctx, &invoices, `
		SELECT i.id,
		       i.invoice_number,
		       i.total_amount,
		       i.currency,
		       v.name            AS vendor_name,
		       v.bank_account_number,
		       v.bank_ifsc
		  FROM invoices i
		  JOIN vendors v ON i.vendor_id = v.id
		 WHERE i.status = 'SCHEDULED'
		   AND i.scheduled_payment_date <= CURRENT_DATE
		   AND i.pinelabs_transaction_id IS NULL
		 ORDER BY i.scheduled_payment_date ASC
	`)
	if err != nil {
		s.logger.Error("Failed to fetch scheduled invoices", zap.Error(err))
		return
	}
	if len(invoices) == 0 {
		s.logger.Info("No invoices due for payment today")
		return
	}

	s.logger.Info("Payment batch: invoices to process", zap.Int("count", len(invoices)))

	// 2. Create a PaymentRun record
	runID := uuid.New()
	totalAmount := 0.0
	for _, inv := range invoices {
		totalAmount += inv.TotalAmount
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO payment_runs (id, run_date, total_amount, invoice_count, status)
		VALUES ($1, CURRENT_DATE, $2, $3, 'EXECUTING')`,
		runID, totalAmount, len(invoices),
	)
	if err != nil {
		s.logger.Error("Failed to create payment run record", zap.Error(err))
		return
	}

	// 3. Disburse each invoice via Pine Labs
	successCount := 0
	failCount := 0

	for _, inv := range invoices {
		invNum := ""
		if inv.InvoiceNumber != nil {
			invNum = *inv.InvoiceNumber
		}
		currency := "INR"
		if inv.Currency != nil && *inv.Currency != "" {
			currency = *inv.Currency
		}

		ref := fmt.Sprintf("SUPPAY-%s", inv.ID.String()[:8])
		disbResp, disbErr := s.payment.InitiateDisbursement(ctx, services.DisbursementRequest{
			InvoiceID:     inv.ID.String(),
			VendorName:    inv.VendorName,
			AccountNumber: inv.AccountNumber,
			IFSC:          inv.IFSC,
			Amount:        inv.TotalAmount,
			Currency:      currency,
			Reference:     ref,
		})

		if disbErr != nil {
			s.logger.Error("Disbursement failed",
				zap.String("invoice_id", inv.ID.String()),
				zap.String("invoice_number", invNum),
				zap.Error(disbErr),
			)
			failCount++
			// Mark invoice with a failure note but keep status SCHEDULED for retry
			_, _ = s.db.ExecContext(ctx,
				`UPDATE invoices SET decision_reason = $1, updated_at = NOW() WHERE id = $2`,
				fmt.Sprintf("Disbursement failed: %v", disbErr), inv.ID,
			)
			continue
		}

		// 4. Update invoice: store transaction ID and mark PAID
		_, updateErr := s.db.ExecContext(ctx, `
			UPDATE invoices
			   SET status = 'PAID',
			       pinelabs_transaction_id = $1,
			       updated_at = NOW()
			 WHERE id = $2`,
			disbResp.TransactionID, inv.ID,
		)
		if updateErr != nil {
			s.logger.Error("Failed to update invoice after disbursement",
				zap.String("invoice_id", inv.ID.String()),
				zap.Error(updateErr),
			)
		} else {
			s.logger.Info("Invoice paid",
				zap.String("invoice_id", inv.ID.String()),
				zap.String("invoice_number", invNum),
				zap.String("transaction_id", disbResp.TransactionID),
				zap.Float64("amount", inv.TotalAmount),
			)
			successCount++
		}
	}

	// 5. Update PaymentRun with final status
	runStatus := "COMPLETED"
	if failCount > 0 && successCount == 0 {
		runStatus = "PARTIAL_FAILURE"
	} else if failCount > 0 {
		runStatus = "PARTIAL_FAILURE"
	}

	_, _ = s.db.ExecContext(ctx, `
		UPDATE payment_runs
		   SET status = $1, updated_at = NOW()
		 WHERE id = $2`,
		runStatus, runID,
	)

	s.logger.Info("Payment batch complete",
		zap.String("run_id", runID.String()),
		zap.Int("success", successCount),
		zap.Int("failed", failCount),
		zap.String("status", runStatus),
	)
}

func (s *PaymentScheduler) updateForecast() {
	s.logger.Info("Updating cash flow forecast")

	// TODO: Dev 3 — Regenerate 90-day forecast
	// 1. Query all APPROVED + SCHEDULED invoices
	// 2. Group payments by week
	// 3. Calculate running balance
	// 4. Detect risk periods
}
