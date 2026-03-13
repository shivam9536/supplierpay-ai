package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/supplierpay/backend/internal/config"
	"github.com/supplierpay/backend/internal/models"
	"github.com/supplierpay/backend/internal/services"
	"go.uber.org/zap"
)

// ── Agent Pipeline States ───────────────────
type PipelineStep string

const (
	StepExtract        PipelineStep = "EXTRACT"
	StepValidate       PipelineStep = "VALIDATE"
	StepCrossReference PipelineStep = "CROSS_REFERENCE"
	StepDecision       PipelineStep = "DECISION"
	StepDraftQuery     PipelineStep = "DRAFT_QUERY"
	StepSchedule       PipelineStep = "SCHEDULE"
)

// ── Orchestrator ────────────────────────────
type Orchestrator struct {
	db        *sqlx.DB
	cfg       *config.Config
	logger    *zap.Logger
	bedrock   services.LLMService
	s3        services.StorageService
	ses       services.EmailService
	eventChan chan models.SSEEvent // For real-time updates
}

func NewOrchestrator(
	db *sqlx.DB,
	cfg *config.Config,
	logger *zap.Logger,
	bedrock services.LLMService,
	s3 services.StorageService,
	ses services.EmailService,
) *Orchestrator {
	return &Orchestrator{
		db:        db,
		cfg:       cfg,
		logger:    logger,
		bedrock:   bedrock,
		s3:        s3,
		ses:       ses,
		eventChan: make(chan models.SSEEvent, 100),
	}
}

// ProcessInvoice runs the full agent pipeline for an invoice
func (o *Orchestrator) ProcessInvoice(ctx context.Context, invoiceID uuid.UUID) error {
	o.logger.Info("Agent pipeline started", zap.String("invoice_id", invoiceID.String()))

	// ── Step 1: EXTRACT ─────────────────────
	o.emitEvent(invoiceID, StepExtract, "in_progress", "Extracting fields from invoice...")
	extractedFields, err := o.extract(ctx, invoiceID)
	if err != nil {
		o.emitEvent(invoiceID, StepExtract, "failed", err.Error())
		return fmt.Errorf("extraction failed: %w", err)
	}
	o.emitEvent(invoiceID, StepExtract, "completed", "Fields extracted successfully")

	// ── Step 2: VALIDATE ────────────────────
	o.emitEvent(invoiceID, StepValidate, "in_progress", "Validating invoice data...")
	validationErrors := o.validate(extractedFields)
	if len(validationErrors) > 0 {
		o.emitEvent(invoiceID, StepValidate, "failed", fmt.Sprintf("Validation failed: %v", validationErrors))
		// TODO: Update invoice status to REJECTED
		return fmt.Errorf("validation failed: %v", validationErrors)
	}
	o.emitEvent(invoiceID, StepValidate, "completed", "Validation passed")

	// ── Step 3: CROSS-REFERENCE ─────────────
	o.emitEvent(invoiceID, StepCrossReference, "in_progress", "Matching against purchase orders...")
	matchResult, err := o.crossReference(ctx, invoiceID, extractedFields)
	if err != nil {
		o.emitEvent(invoiceID, StepCrossReference, "failed", err.Error())
		return fmt.Errorf("cross-reference failed: %w", err)
	}
	o.emitEvent(invoiceID, StepCrossReference, "completed", matchResult.Summary)

	// ── Step 4: DECISION ────────────────────
	o.emitEvent(invoiceID, StepDecision, "in_progress", "Making approval decision...")
	decision := o.makeDecision(matchResult)
	o.emitEvent(invoiceID, StepDecision, "completed", decision.Reason)

	// ── Step 5: ACTION ──────────────────────
	switch decision.Action {
	case "APPROVE":
		return o.schedulePayment(ctx, invoiceID, extractedFields)
	case "FLAG":
		return o.draftQueryEmail(ctx, invoiceID, matchResult)
	case "REJECT":
		// TODO: Update invoice status to REJECTED, notify vendor
		return nil
	}

	return nil
}

// ── Pipeline Step Implementations ───────────

func (o *Orchestrator) extract(ctx context.Context, invoiceID uuid.UUID) (map[string]interface{}, error) {
	// TODO: Dev 2 — Implement Bedrock extraction
	// 1. Fetch invoice file from S3
	// 2. Send to Bedrock Claude with extraction prompt
	// 3. Parse response into structured fields
	// 4. Store extracted_fields in DB
	// 5. Log to audit_log

	o.logger.Info("Extracting invoice fields", zap.String("invoice_id", invoiceID.String()))

	// Mock response for development
	if o.cfg.MockMode {
		return map[string]interface{}{
			"vendor_name":    "Acme Corp",
			"invoice_number": "INV-2026-001",
			"po_reference":   "PO-2026-100",
			"total_amount":   50000.00,
			"tax_amount":     9000.00,
			"currency":       "INR",
			"invoice_date":   "2026-03-01",
			"due_date":       "2026-03-31",
			"line_items": []map[string]interface{}{
				{"description": "Cloud Hosting - March", "quantity": 1, "unit_price": 41000.00, "total": 41000.00},
				{"description": "Support Services", "quantity": 1, "unit_price": 9000.00, "total": 9000.00},
			},
		}, nil
	}

	// Real Bedrock call
	// extractedFields, err := o.bedrock.ExtractInvoiceFields(ctx, invoiceData)
	return nil, fmt.Errorf("not implemented")
}

func (o *Orchestrator) validate(fields map[string]interface{}) []string {
	// TODO: Dev 2 — Implement validation rules
	var errors []string

	// Amount > 0
	if amount, ok := fields["total_amount"].(float64); ok && amount <= 0 {
		errors = append(errors, "Total amount must be greater than 0")
	}

	// Required fields present
	requiredFields := []string{"vendor_name", "invoice_number", "total_amount", "due_date"}
	for _, field := range requiredFields {
		if _, ok := fields[field]; !ok {
			errors = append(errors, fmt.Sprintf("Missing required field: %s", field))
		}
	}

	return errors
}

type MatchResult struct {
	POFound         bool
	AmountMatch     bool
	LineItemsMatch  bool
	Discrepancies   []string
	Summary         string
	DiscrepancyType string // AMOUNT_MISMATCH, PO_NOT_FOUND, DUPLICATE_INVOICE, LINE_ITEM_MISMATCH
}

func (o *Orchestrator) crossReference(ctx context.Context, invoiceID uuid.UUID, fields map[string]interface{}) (*MatchResult, error) {
	// TODO: Dev 2 — Implement PO cross-reference
	// 1. Look up PO by po_reference
	// 2. Compare amounts (invoice ≤ PO)
	// 3. Fuzzy match line items
	// 4. Check for duplicate invoice numbers
	// 5. Check goods receipt (3-way match)

	o.logger.Info("Cross-referencing invoice", zap.String("invoice_id", invoiceID.String()))

	// Mock result
	return &MatchResult{
		POFound:        true,
		AmountMatch:    true,
		LineItemsMatch: true,
		Discrepancies:  nil,
		Summary:        "All checks passed — PO matched, amounts verified",
	}, nil
}

type Decision struct {
	Action string // APPROVE, FLAG, REJECT
	Reason string
}

func (o *Orchestrator) makeDecision(match *MatchResult) *Decision {
	// Decision tree from the design doc
	if !match.POFound {
		return &Decision{Action: "REJECT", Reason: "Purchase order not found in system"}
	}

	if match.DiscrepancyType == "DUPLICATE_INVOICE" {
		return &Decision{Action: "REJECT", Reason: "Duplicate invoice detected"}
	}

	if !match.AmountMatch {
		return &Decision{Action: "FLAG", Reason: fmt.Sprintf("Amount mismatch: %s", match.Discrepancies)}
	}

	if !match.LineItemsMatch {
		return &Decision{Action: "FLAG", Reason: "Line items do not match purchase order"}
	}

	return &Decision{Action: "APPROVE", Reason: "All checks passed — auto-approved"}
}

func (o *Orchestrator) schedulePayment(ctx context.Context, invoiceID uuid.UUID, fields map[string]interface{}) error {
	// TODO: Dev 2 — Implement payment scheduling logic
	// 1. Get vendor payment terms
	// 2. Check for early payment discount ROI
	// 3. Calculate optimal payment date
	// 4. Update invoice with scheduled_payment_date and status SCHEDULED

	o.emitEvent(invoiceID, StepSchedule, "in_progress", "Calculating optimal payment date...")

	// Mock: Schedule for day 28
	paymentDate := time.Now().AddDate(0, 0, 28)
	o.logger.Info("Payment scheduled",
		zap.String("invoice_id", invoiceID.String()),
		zap.Time("payment_date", paymentDate),
	)

	o.emitEvent(invoiceID, StepSchedule, "completed",
		fmt.Sprintf("Payment scheduled for %s", paymentDate.Format("2006-01-02")))

	return nil
}

func (o *Orchestrator) draftQueryEmail(ctx context.Context, invoiceID uuid.UUID, match *MatchResult) error {
	// TODO: Dev 2 — Use Bedrock to generate supplier query email
	o.emitEvent(invoiceID, StepDraftQuery, "in_progress", "Drafting supplier query email...")

	// Mock email draft
	o.logger.Info("Drafting query email",
		zap.String("invoice_id", invoiceID.String()),
		zap.Strings("discrepancies", match.Discrepancies),
	)

	o.emitEvent(invoiceID, StepDraftQuery, "completed", "Query email drafted and sent to supplier")

	return nil
}

// emitEvent sends a real-time update via SSE
func (o *Orchestrator) emitEvent(invoiceID uuid.UUID, step PipelineStep, status, message string) {
	event := models.SSEEvent{
		InvoiceID: invoiceID.String(),
		Step:      string(step),
		Status:    status,
		Message:   message,
	}

	select {
	case o.eventChan <- event:
	default:
		o.logger.Warn("Event channel full, dropping event")
	}
}

// GetEventChannel returns the SSE event channel for streaming
func (o *Orchestrator) GetEventChannel() <-chan models.SSEEvent {
	return o.eventChan
}
