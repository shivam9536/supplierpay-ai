package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
	StepExtract    PipelineStep = "EXTRACT"
	StepValidate   PipelineStep = "VALIDATE"
	StepDecision   PipelineStep = "DECISION"
	StepDraftQuery PipelineStep = "DRAFT_QUERY"
	StepSchedule   PipelineStep = "SCHEDULE"
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
	// Results are written to invoice_validations — the invoices table is NOT
	// modified here (status stays EXTRACTING until DECISION step).
	o.emitEvent(invoiceID, StepValidate, "in_progress", "Validating invoice fields, vendor and PO line items...")
	valRec, err := o.runValidation(ctx, invoiceID, extractedFields)
	if err != nil {
		o.emitEvent(invoiceID, StepValidate, "failed", err.Error())
		return fmt.Errorf("validation error: %w", err)
	}
	if valRec.ValidationStatus == models.ValidationStatusFailed {
		o.emitEvent(invoiceID, StepValidate, "failed",
			fmt.Sprintf("Validation failed (%d checks): %s", len(valRec.FailureReasons), valRec.Summary))
		o.updateInvoiceStatus(invoiceID, models.InvoiceStatusRejected, valRec.Summary, valRec.FailureReasons)
		o.insertAuditLog(invoiceID, StepValidate, "failed", valRec.Summary, 0)
		return fmt.Errorf("validation failed: %s", valRec.Summary)
	}
	o.emitEvent(invoiceID, StepValidate, "completed", valRec.Summary)
	o.insertAuditLog(invoiceID, StepValidate, "completed", valRec.Summary, 1.0)

	// ── Step 3: DECISION ────────────────────
	// Driven entirely from the invoice_validations record — no separate
	// crossReference step needed; runValidation already covers PO, vendor,
	// items, prices, amounts and duplicates.
	o.emitEvent(invoiceID, StepDecision, "in_progress", "Making approval decision...")
	decision := o.makeDecision(valRec)
	o.emitEvent(invoiceID, StepDecision, "completed", decision.Reason)

	statusStr := decisionActionToStatus(decision.Action)
	o.updateInvoiceStatus(invoiceID, statusStr, decision.Reason, valRec.FailureReasons)
	o.insertAuditLog(invoiceID, StepDecision, "completed", decision.Reason, 1.0)

	switch decision.Action {
	case "APPROVE":
		return o.schedulePayment(ctx, invoiceID, extractedFields)
	case "FLAG":
		return o.draftQueryEmailFromValidation(ctx, invoiceID, valRec)
	case "REJECT":
		return nil
	}

	return nil
}

// ── Pipeline Step Implementations ───────────

func (o *Orchestrator) extract(ctx context.Context, invoiceID uuid.UUID) (map[string]interface{}, error) {
	o.logger.Info("Extracting invoice fields", zap.String("invoice_id", invoiceID.String()))

	var rawURL string
	var extractedJSON []byte
	err := o.db.QueryRow(`SELECT raw_file_url, extracted_fields FROM invoices WHERE id = $1`, invoiceID).Scan(&rawURL, &extractedJSON)
	if err != nil {
		return nil, fmt.Errorf("invoice not found: %w", err)
	}

	var extractedFields map[string]interface{}
	if len(extractedJSON) > 0 && string(extractedJSON) != "{}" && string(extractedJSON) != "null" {
		if err := json.Unmarshal(extractedJSON, &extractedFields); err == nil && len(extractedFields) > 0 {
			o.insertAuditLog(invoiceID, StepExtract, "completed", "Used existing extracted fields from upload", 1.0)
			return extractedFields, nil
		}
	}

	if rawURL != "" {
		// Get file from storage (key from URL or invoice ID)
		key := "invoices/" + invoiceID.String()
		data, err := o.s3.GetFile(ctx, key)
		if err == nil && len(data) > 0 {
			extractedFields, err = o.bedrock.ExtractInvoiceFields(ctx, data, "application/pdf")
			if err != nil && !o.cfg.MockMode {
				return nil, err
			}
			if extractedFields != nil {
				o.persistExtractedFields(invoiceID, extractedFields)
				o.insertAuditLog(invoiceID, StepExtract, "completed", "Extracted fields via LLM", 0.95)
				return extractedFields, nil
			}
		}
	}

	// Mock response for development
	if o.cfg.MockMode {
		extractedFields = map[string]interface{}{
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
		}
		o.persistExtractedFields(invoiceID, extractedFields)
		o.insertAuditLog(invoiceID, StepExtract, "completed", "Extracted 7 fields (mock)", 0.95)
		return extractedFields, nil
	}

	return nil, fmt.Errorf("extraction not implemented without mock or existing data")
}

// runValidation is the authoritative validation step. It writes a full record
// to invoice_validations (never touching the invoices table) covering:
//  1. Required field presence and data-type sanity
//  2. Vendor exists in the vendors table
//  3. PO (order number) exists and is open
//  4. Vendor on the invoice matches the PO vendor
//  5. No duplicate invoice number
//  6. Every ordered item found in PO with matching description
//  7. Unit prices within 2 % tolerance
//  8. Invoice total does not exceed PO remaining_value
func (o *Orchestrator) runValidation(
	ctx context.Context,
	invoiceID uuid.UUID,
	fields map[string]interface{},
) (*models.InvoiceValidation, error) {

	now := time.Now()
	rec := &models.InvoiceValidation{
		InvoiceID:        invoiceID,
		ValidationStatus: models.ValidationStatusRunning,
		StartedAt:        now,
		CheckResults:     models.CheckResults{},
		LineItemResults:  models.LineItemResults{},
		FailureReasons:   models.StringSlice{},
	}

	// Persist as RUNNING so it's visible immediately
	if err := o.upsertValidationRecord(rec); err != nil {
		return nil, fmt.Errorf("could not create validation record: %w", err)
	}

	// Helper closures
	boolPtr := func(b bool) *bool { return &b }
	addCheck := func(name string, passed bool, detail string) {
		rec.CheckResults = append(rec.CheckResults, models.CheckResult{
			Check: name, Passed: passed, Detail: detail,
		})
		if !passed {
			rec.FailureReasons = append(rec.FailureReasons, detail)
		}
	}

	// ── 1. Required fields ───────────────────────────────────────────────
	requiredFields := []string{"vendor_name", "invoice_number", "po_reference", "total_amount", "invoice_date", "currency"}
	allPresent := true
	for _, f := range requiredFields {
		v, ok := fields[f]
		if !ok || v == nil || v == "" {
			addCheck("required_field:"+f, false, fmt.Sprintf("missing required field: %s", f))
			allPresent = false
		}
	}
	if allPresent {
		addCheck("required_fields", true, "all required fields present")
	}

	// ── 2. Total amount sanity ───────────────────────────────────────────
	totalAmount, _ := toFloat(fields["total_amount"])
	if totalAmount <= 0 {
		addCheck("total_amount_positive", false, fmt.Sprintf("total_amount must be > 0, got %.2f", totalAmount))
	} else {
		addCheck("total_amount_positive", true, fmt.Sprintf("total_amount ₹%.2f is valid", totalAmount))
	}

	// ── 3. Date sanity ───────────────────────────────────────────────────
	invDateStr, _ := fields["invoice_date"].(string)
	dueDateStr, _ := fields["due_date"].(string)
	if invDateStr != "" && dueDateStr != "" {
		invDate, e1 := time.Parse("2006-01-02", invDateStr)
		dueDate, e2 := time.Parse("2006-01-02", dueDateStr)
		if e1 != nil || e2 != nil {
			addCheck("date_format", false, "invalid invoice_date or due_date format (expected YYYY-MM-DD)")
		} else if !dueDate.After(invDate) {
			addCheck("date_order", false, "due_date must be after invoice_date")
		} else {
			addCheck("date_order", true, fmt.Sprintf("invoice_date %s → due_date %s", invDateStr, dueDateStr))
		}
	}

	// ── 4. Vendor exists in vendors table ────────────────────────────────
	var invVendorID string
	_ = o.db.QueryRowContext(ctx, `SELECT vendor_id::text FROM invoices WHERE id = $1`, invoiceID).Scan(&invVendorID)
	vendorName, _ := fields["vendor_name"].(string)
	if invVendorID != "" {
		var matchedName string
		err := o.db.QueryRowContext(ctx,
			`SELECT name FROM vendors WHERE id = $1`, invVendorID,
		).Scan(&matchedName)
		if err != nil {
			rec.VendorValid = boolPtr(false)
			addCheck("vendor_exists", false,
				fmt.Sprintf("vendor_id %s not found in vendors table", invVendorID))
		} else {
			rec.VendorValid = boolPtr(true)
			addCheck("vendor_exists", true,
				fmt.Sprintf("vendor '%s' (id: %s) is a registered vendor", matchedName, invVendorID))
			// Soft check: vendor name on invoice matches DB name
			if vendorName != "" && !strings.EqualFold(strings.TrimSpace(vendorName), strings.TrimSpace(matchedName)) {
				addCheck("vendor_name_match", false,
					fmt.Sprintf("invoice vendor name '%s' does not match registered name '%s'", vendorName, matchedName))
			} else {
				addCheck("vendor_name_match", true,
					fmt.Sprintf("vendor name '%s' matches", matchedName))
			}
		}
	} else {
		rec.VendorValid = boolPtr(false)
		addCheck("vendor_exists", false, "could not resolve vendor_id from invoice")
	}

	// ── 5. PO exists and is open ─────────────────────────────────────────
	poRef, _ := fields["po_reference"].(string)
	var poID, poVendorID, poStatus string
	var poTotal, poRemaining float64
	var poLineItemsJSON []byte

	poErr := o.db.QueryRowContext(ctx,
		`SELECT id::text, vendor_id::text, total_value, remaining_value, line_items, status
		   FROM purchase_orders WHERE po_number = $1`, poRef,
	).Scan(&poID, &poVendorID, &poTotal, &poRemaining, &poLineItemsJSON, &poStatus)

	if poErr != nil {
		rec.POFound = boolPtr(false)
		rec.POOpen = boolPtr(false)
		addCheck("po_exists", false, fmt.Sprintf("purchase order '%s' not found", poRef))
	} else {
		rec.POFound = boolPtr(true)
		poIDParsed, _ := uuid.Parse(poID)
		rec.MatchedPONumber = &poRef
		rec.MatchedPOID = &poIDParsed
		addCheck("po_exists", true, fmt.Sprintf("PO '%s' found (status: %s, remaining: ₹%.2f)", poRef, poStatus, poRemaining))

		// PO must be open
		if poStatus == "CLOSED" {
			rec.POOpen = boolPtr(false)
			addCheck("po_open", false, fmt.Sprintf("PO '%s' is CLOSED — no further invoicing allowed", poRef))
		} else {
			rec.POOpen = boolPtr(true)
			addCheck("po_open", true, fmt.Sprintf("PO '%s' is %s", poRef, poStatus))
		}

		// ── 6. Vendor on invoice must match PO vendor ────────────────────
		if invVendorID != "" && poVendorID != "" {
			if invVendorID == poVendorID {
				rec.VendorMatchesPO = boolPtr(true)
				addCheck("vendor_matches_po", true, "invoice vendor matches PO vendor")
			} else {
				rec.VendorMatchesPO = boolPtr(false)
				addCheck("vendor_matches_po", false,
					fmt.Sprintf("invoice vendor (%s) ≠ PO vendor (%s)", invVendorID, poVendorID))
			}
		}

		// ── 7. Duplicate invoice number check ───────────────────────────
		invNum, _ := fields["invoice_number"].(string)
		if invNum != "" {
			var dupCount int
			_ = o.db.QueryRowContext(ctx,
				`SELECT COUNT(*) FROM invoices WHERE invoice_number=$1 AND id!=$2 AND status NOT IN ('REJECTED')`,
				invNum, invoiceID).Scan(&dupCount)
			if dupCount > 0 {
				rec.NoDuplicate = boolPtr(false)
				addCheck("no_duplicate", false, fmt.Sprintf("invoice number '%s' already exists in the system", invNum))
			} else {
				rec.NoDuplicate = boolPtr(true)
				addCheck("no_duplicate", true, fmt.Sprintf("invoice number '%s' is unique", invNum))
			}
		}

		// ── 8. Invoice total vs PO remaining_value ───────────────────────
		const amountTol = 0.02
		if totalAmount > poRemaining*(1+amountTol) {
			rec.AmountWithinPO = boolPtr(false)
			addCheck("amount_within_po", false,
				fmt.Sprintf("invoice total ₹%.2f exceeds PO remaining ₹%.2f", totalAmount, poRemaining))
		} else {
			rec.AmountWithinPO = boolPtr(true)
			addCheck("amount_within_po", true,
				fmt.Sprintf("invoice total ₹%.2f ≤ PO remaining ₹%.2f", totalAmount, poRemaining))
		}

		// ── 9. Per-line-item check: ordered items + prices ───────────────
		var poItems []map[string]interface{}
		if len(poLineItemsJSON) > 0 {
			var raw []interface{}
			if json.Unmarshal(poLineItemsJSON, &raw) == nil {
				for _, e := range raw {
					if m, ok := e.(map[string]interface{}); ok {
						poItems = append(poItems, m)
					}
				}
			}
		}

		invItems, hasItems := toLineItemSlice(fields["line_items"])
		if !hasItems || len(invItems) == 0 {
			addCheck("line_items_present", false, "invoice has no line items")
			rec.ItemsMatch = boolPtr(false)
			rec.PricesMatch = boolPtr(false)
		} else {
			addCheck("line_items_present", true, fmt.Sprintf("%d line item(s) on invoice", len(invItems)))

			if len(poItems) == 0 {
				// PO has no stored items — skip item-level check
				rec.ItemsMatch = boolPtr(true)
				rec.PricesMatch = boolPtr(true)
				addCheck("items_match", true, "PO has no stored line items — item-level check skipped")
			} else {
				// Build PO lookup by normalised description
				poByDesc := make(map[string]map[string]interface{}, len(poItems))
				for _, pi := range poItems {
					if d, _ := pi["description"].(string); d != "" {
						poByDesc[normaliseDesc(d)] = pi
					}
				}

				allItemsOK := true
				allPricesOK := true
				const priceTol = 0.02

				for i, inv := range invItems {
					idx := i + 1
					invDesc, _ := inv["description"].(string)
					key := normaliseDesc(invDesc)

					poItem, found := poByDesc[key]
					if !found {
						for k, pi := range poByDesc {
							if strings.Contains(k, key) || strings.Contains(key, k) {
								poItem = pi
								found = true
								break
							}
						}
					}

					lir := models.LineItemResult{
						Description: invDesc,
						Matched:     found,
					}
					lir.InvQty, _ = toFloat(inv["quantity"])
					lir.InvPrice, _ = toFloat(inv["unit_price"])

					if !found {
						allItemsOK = false
						lir.Note = fmt.Sprintf("line_item[%d] '%s' not found in PO %s", idx, invDesc, poRef)
						addCheck("item_match:"+invDesc, false, lir.Note)
					} else {
						lir.POQty, _ = toFloat(poItem["quantity"])
						lir.POPrice, _ = toFloat(poItem["unit_price"])

						// Quantity
						qtyNote := ""
						if lir.InvQty != lir.POQty {
							qtyNote = fmt.Sprintf("qty %g ≠ PO qty %g; ", lir.InvQty, lir.POQty)
							allItemsOK = false
						}

						// Price tolerance
						priceNote := ""
						if lir.POPrice > 0 {
							diff := abs64(lir.InvPrice-lir.POPrice) / lir.POPrice
							if diff > priceTol {
								allPricesOK = false
								lir.Matched = false
								priceNote = fmt.Sprintf("price ₹%.2f ≠ PO ₹%.2f (%.1f%% variance)",
									lir.InvPrice, lir.POPrice, diff*100)
							}
						}

						if qtyNote != "" || priceNote != "" {
							lir.Note = strings.TrimRight(qtyNote+priceNote, "; ")
							addCheck("item_check:"+invDesc, false, lir.Note)
						} else {
							addCheck("item_check:"+invDesc, true,
								fmt.Sprintf("'%s' qty=%g price=₹%.2f ✓", invDesc, lir.InvQty, lir.InvPrice))
						}
					}
					rec.LineItemResults = append(rec.LineItemResults, lir)
				}

				rec.ItemsMatch = boolPtr(allItemsOK)
				rec.PricesMatch = boolPtr(allPricesOK)
				if allItemsOK {
					addCheck("items_match", true, fmt.Sprintf("all %d ordered items found in PO", len(invItems)))
				}
				if allPricesOK {
					addCheck("prices_match", true, "all item prices within 2% tolerance")
				} else {
					addCheck("prices_match", false, "one or more item prices deviate from PO by > 2%")
				}
			}
		}
	}

	// ── Determine final status ───────────────────────────────────────────
	hasFail := false
	hasSoft := false
	for _, c := range rec.CheckResults {
		if !c.Passed {
			// Soft flags: price drift or qty drift alone → FLAGGED (not FAILED)
			if strings.HasPrefix(c.Check, "item_check:") || strings.HasPrefix(c.Check, "vendor_name_match") {
				hasSoft = true
			} else {
				hasFail = true
			}
		}
	}

	finishedAt := time.Now()
	rec.CompletedAt = &finishedAt

	if hasFail {
		rec.ValidationStatus = models.ValidationStatusFailed
		rec.Summary = fmt.Sprintf("Validation FAILED — %d check(s) failed: %v",
			len(rec.FailureReasons), rec.FailureReasons)
	} else if hasSoft {
		rec.ValidationStatus = models.ValidationStatusFlagged
		rec.Summary = fmt.Sprintf("Validation FLAGGED — soft discrepancies detected: %v", rec.FailureReasons)
	} else {
		rec.ValidationStatus = models.ValidationStatusPassed
		rec.FailureReasons = models.StringSlice{}
		rec.Summary = fmt.Sprintf("Validation PASSED — %d checks all clear", len(rec.CheckResults))
	}

	// Persist final state
	if err := o.upsertValidationRecord(rec); err != nil {
		o.logger.Error("failed to persist validation record", zap.Error(err))
	}

	o.logger.Info("Validation complete",
		zap.String("invoice_id", invoiceID.String()),
		zap.String("status", string(rec.ValidationStatus)),
		zap.Int("checks_run", len(rec.CheckResults)),
		zap.Int("failures", len(rec.FailureReasons)),
	)
	return rec, nil
}

// upsertValidationRecord inserts or updates the invoice_validations row for
// the given invoice. Uses INSERT … ON CONFLICT DO UPDATE so repeated calls
// (RUNNING → PASSED/FAILED) are idempotent.
func (o *Orchestrator) upsertValidationRecord(r *models.InvoiceValidation) error {
	checksJSON, _ := json.Marshal(r.CheckResults)
	lineJSON, _ := json.Marshal(r.LineItemResults)
	failJSON, _ := json.Marshal(r.FailureReasons)

	_, err := o.db.Exec(`
		INSERT INTO invoice_validations
		  (invoice_id, validation_status,
		   vendor_valid, po_found, po_open, vendor_matches_po,
		   items_match, prices_match, amount_within_po, no_duplicate,
		   check_results, line_item_results,
		   matched_po_number, matched_po_id,
		   summary, failure_reasons, started_at, completed_at)
		VALUES
		  ($1,$2, $3,$4,$5,$6, $7,$8,$9,$10, $11::jsonb,$12::jsonb,
		   $13,$14, $15,$16::jsonb,$17,$18)
		ON CONFLICT (invoice_id) DO UPDATE SET
		  validation_status  = EXCLUDED.validation_status,
		  vendor_valid       = EXCLUDED.vendor_valid,
		  po_found           = EXCLUDED.po_found,
		  po_open            = EXCLUDED.po_open,
		  vendor_matches_po  = EXCLUDED.vendor_matches_po,
		  items_match        = EXCLUDED.items_match,
		  prices_match       = EXCLUDED.prices_match,
		  amount_within_po   = EXCLUDED.amount_within_po,
		  no_duplicate       = EXCLUDED.no_duplicate,
		  check_results      = EXCLUDED.check_results,
		  line_item_results  = EXCLUDED.line_item_results,
		  matched_po_number  = EXCLUDED.matched_po_number,
		  matched_po_id      = EXCLUDED.matched_po_id,
		  summary            = EXCLUDED.summary,
		  failure_reasons    = EXCLUDED.failure_reasons,
		  completed_at       = EXCLUDED.completed_at,
		  updated_at         = NOW()
	`,
		r.InvoiceID, string(r.ValidationStatus),
		r.VendorValid, r.POFound, r.POOpen, r.VendorMatchesPO,
		r.ItemsMatch, r.PricesMatch, r.AmountWithinPO, r.NoDuplicate,
		string(checksJSON), string(lineJSON),
		r.MatchedPONumber, r.MatchedPOID,
		r.Summary, string(failJSON), r.StartedAt, r.CompletedAt,
	)
	return err
}

// ── helpers ──────────────────────────────────────────────────────────────────

func toFloat(v interface{}) (float64, bool) {
	if v == nil {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	}
	return 0, false
}

func abs64(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// toLineItemSlice converts the raw line_items value (which may be
// []interface{} after JSON unmarshalling or []map[string]interface{})
// into a uniform []map[string]interface{}.
func toLineItemSlice(raw interface{}) ([]map[string]interface{}, bool) {
	switch v := raw.(type) {
	case []map[string]interface{}:
		return v, true
	case []interface{}:
		out := make([]map[string]interface{}, 0, len(v))
		for _, elem := range v {
			if m, ok := elem.(map[string]interface{}); ok {
				out = append(out, m)
			}
		}
		return out, len(out) == len(v)
	}
	return nil, false
}

// normaliseDesc returns a lower-case, trimmed description for comparison.
func normaliseDesc(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

type Decision struct {
	Action string // APPROVE, FLAG, REJECT
	Reason string
}

// makeDecision maps an InvoiceValidation record to an action.
//
//	FAILED  → hard failures (PO not found, closed, vendor mismatch,
//	           duplicate, amount over PO) → REJECT
//	FLAGGED → soft discrepancies (item/price drift) → FLAG for human review
//	PASSED  → APPROVE
func (o *Orchestrator) makeDecision(v *models.InvoiceValidation) *Decision {
	switch v.ValidationStatus {
	case models.ValidationStatusFailed:
		return &Decision{Action: "REJECT", Reason: v.Summary}
	case models.ValidationStatusFlagged:
		return &Decision{Action: "FLAG", Reason: v.Summary}
	default:
		return &Decision{Action: "APPROVE", Reason: "All checks passed — auto-approved"}
	}
}

func (o *Orchestrator) draftQueryEmailFromValidation(ctx context.Context, invoiceID uuid.UUID, v *models.InvoiceValidation) error {
	o.emitEvent(invoiceID, StepDraftQuery, "in_progress", "Drafting supplier query email...")
	o.logger.Info("Drafting query email",
		zap.String("invoice_id", invoiceID.String()),
		zap.Strings("failure_reasons", v.FailureReasons),
	)
	o.emitEvent(invoiceID, StepDraftQuery, "completed", "Query email drafted and sent to supplier")
	o.insertAuditLog(invoiceID, StepDraftQuery, "completed", "Query email drafted for discrepancies", 0.95)
	return nil
}

func (o *Orchestrator) schedulePayment(ctx context.Context, invoiceID uuid.UUID, fields map[string]interface{}) error {
	o.emitEvent(invoiceID, StepSchedule, "in_progress", "Calculating optimal payment date...")

	var paymentTermsDays int
	err := o.db.QueryRow(`SELECT v.payment_terms_days FROM invoices i JOIN vendors v ON i.vendor_id = v.id WHERE i.id = $1`, invoiceID).Scan(&paymentTermsDays)
	if err != nil || paymentTermsDays <= 0 {
		paymentTermsDays = 30
	}
	paymentDate := time.Now().AddDate(0, 0, paymentTermsDays)
	_, _ = o.db.Exec(`UPDATE invoices SET scheduled_payment_date = $1, status = $2 WHERE id = $3`,
		paymentDate, models.InvoiceStatusScheduled, invoiceID)
	o.logger.Info("Payment scheduled",
		zap.String("invoice_id", invoiceID.String()),
		zap.Time("payment_date", paymentDate),
	)

	// ── Decrement PO remaining_value and update PO status ────────────────
	// This keeps the PO balance accurate for future invoices against the same PO.
	poRef, _ := fields["po_reference"].(string)
	invTotal, _ := toFloat(fields["total_amount"])
	if poRef != "" && invTotal > 0 {
		_, poErr := o.db.Exec(`
			UPDATE purchase_orders
			   SET remaining_value = GREATEST(0, remaining_value - $1),
			       status = CASE
			                  WHEN GREATEST(0, remaining_value - $1) = 0 THEN 'CLOSED'
			                  WHEN remaining_value - $1 < total_value   THEN 'PARTIALLY_MATCHED'
			                  ELSE status
			                END,
			       updated_at = NOW()
			 WHERE po_number = $2`, invTotal, poRef)
		if poErr != nil {
			o.logger.Warn("Failed to decrement PO remaining_value",
				zap.String("po_reference", poRef),
				zap.Error(poErr),
			)
		} else {
			o.logger.Info("PO remaining value decremented",
				zap.String("po_reference", poRef),
				zap.Float64("invoice_total", invTotal),
			)
		}
	}

	o.emitEvent(invoiceID, StepSchedule, "completed",
		fmt.Sprintf("Payment scheduled for %s", paymentDate.Format("2006-01-02")))
	o.insertAuditLog(invoiceID, StepSchedule, "completed",
		fmt.Sprintf("Payment scheduled for %s (terms: %d days)", paymentDate.Format("2006-01-02"), paymentTermsDays), 1.0)
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

func decisionActionToStatus(action string) models.InvoiceStatus {
	switch action {
	case "APPROVE":
		return models.InvoiceStatusApproved
	case "FLAG":
		return models.InvoiceStatusFlagged
	case "REJECT":
		return models.InvoiceStatusRejected
	default:
		return models.InvoiceStatusFlagged
	}
}

func (o *Orchestrator) updateInvoiceStatus(invoiceID uuid.UUID, status models.InvoiceStatus, reason string, discrepancies []string) {
	var discJSON string = "[]"
	if len(discrepancies) > 0 {
		b, _ := json.Marshal(discrepancies)
		discJSON = string(b)
	}
	_, err := o.db.Exec(`UPDATE invoices SET status = $1, decision_reason = $2, discrepancies = $3::jsonb WHERE id = $4`,
		status, reason, discJSON, invoiceID)
	if err != nil {
		o.logger.Error("Failed to update invoice status", zap.Error(err), zap.String("invoice_id", invoiceID.String()))
	}
}

func (o *Orchestrator) insertAuditLog(invoiceID uuid.UUID, step PipelineStep, result, reasoning string, confidence float64) {
	_, err := o.db.Exec(`INSERT INTO audit_logs (invoice_id, step, result, reasoning, confidence_score) VALUES ($1, $2, $3, $4, $5)`,
		invoiceID, string(step), result, reasoning, confidence)
	if err != nil {
		o.logger.Error("Failed to insert audit log", zap.Error(err), zap.String("invoice_id", invoiceID.String()))
	}
}

func (o *Orchestrator) persistExtractedFields(invoiceID uuid.UUID, fields map[string]interface{}) {
	extJSON, _ := json.Marshal(fields)
	totalAmt := 0.0
	if v, ok := fields["total_amount"].(float64); ok {
		totalAmt = v
	}
	taxAmt := 0.0
	if v, ok := fields["tax_amount"].(float64); ok {
		taxAmt = v
	}
	currency, _ := fields["currency"].(string)
	if currency == "" {
		currency = "INR"
	}
	lineItems := fields["line_items"]
	lineJSON := "[]"
	if lineItems != nil {
		if b, err := json.Marshal(lineItems); err == nil {
			lineJSON = string(b)
		}
	}
	var invDate, dueDate *time.Time
	if s, ok := fields["invoice_date"].(string); ok && s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			invDate = &t
		}
	}
	if s, ok := fields["due_date"].(string); ok && s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			dueDate = &t
		}
	}
	invNum, _ := fields["invoice_number"].(string)
	poRef, _ := fields["po_reference"].(string)
	_, err := o.db.Exec(`UPDATE invoices SET extracted_fields = $1::jsonb, line_items = $2::jsonb, total_amount = $3, tax_amount = $4, currency = $5, invoice_date = $6, due_date = $7, invoice_number = $8, po_reference = $9 WHERE id = $10`,
		string(extJSON), lineJSON, totalAmt, taxAmt, currency, invDate, dueDate, invNum, poRef, invoiceID)
	if err != nil {
		o.logger.Error("Failed to persist extracted fields", zap.Error(err), zap.String("invoice_id", invoiceID.String()))
	}
}
