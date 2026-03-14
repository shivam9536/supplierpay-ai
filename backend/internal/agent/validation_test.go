package agent

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/supplierpay/backend/internal/config"
	"github.com/supplierpay/backend/internal/models"
	"github.com/supplierpay/backend/internal/services"
	"go.uber.org/zap"
)

// ── Stub LLM that always returns APPROVE with no extra flags ─────────────────

type stubLLM struct{}

func (s *stubLLM) ExtractInvoiceFields(_ context.Context, _ []byte, _ string) (map[string]interface{}, error) {
	return nil, nil
}
func (s *stubLLM) GenerateQueryEmail(_ context.Context, _ map[string]interface{}, _ []string) (string, error) {
	return "", nil
}
func (s *stubLLM) ExplainDecision(_ context.Context, _ map[string]interface{}) (string, error) {
	return "", nil
}
func (s *stubLLM) ValidateWithLLM(_ context.Context, _ map[string]interface{}, _ map[string]interface{}, _ string) (*services.LLMValidationResult, error) {
	return &services.LLMValidationResult{
		OverallAssessment:       "APPROVE",
		Confidence:              1.0,
		AdditionalDiscrepancies: nil,
		RiskFlags:               nil,
		Explanation:             "stub: all good",
	}, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func newTestOrchestrator(t *testing.T) (*Orchestrator, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	sqlxDB := sqlx.NewDb(db, "postgres")
	o := &Orchestrator{
		db:        sqlxDB,
		cfg:       &config.Config{MockMode: false},
		logger:    zap.NewNop(),
		bedrock:   &stubLLM{},
		eventChan: make(chan models.SSEEvent, 100),
	}
	return o, mock
}

func mustUUID(t *testing.T) uuid.UUID {
	t.Helper()
	id, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// poLineItemsJSON encodes a slice of line items for use in mock rows.
func poLineItemsJSON(items []map[string]interface{}) []byte {
	b, _ := json.Marshal(items)
	return b
}

// expectFirstUpsert registers the initial RUNNING upsert that happens before
// any DB lookups in runValidation.
func expectFirstUpsert(mock sqlmock.Sqlmock) {
	mock.ExpectExec(`INSERT INTO invoice_validations`).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

// expectFinalUpsert registers the final upsert that persists the completed result.
func expectFinalUpsert(mock sqlmock.Sqlmock) {
	mock.ExpectExec(`INSERT INTO invoice_validations`).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

// ── runValidation table tests ─────────────────────────────────────────────────

func TestRunValidation(t *testing.T) {
	vendorID := mustUUID(t)
	invoiceID := mustUUID(t)

	validFields := map[string]interface{}{
		"vendor_name":    "Acme Corp",
		"invoice_number": "INV-001",
		"po_reference":   "PO-100",
		"total_amount":   float64(50000),
		"tax_amount":     float64(9000),
		"currency":       "INR",
		"invoice_date":   "2026-03-01",
		"due_date":       "2026-03-31",
		"line_items": []interface{}{
			map[string]interface{}{"description": "Cloud Hosting", "quantity": float64(1), "unit_price": float64(41000)},
			map[string]interface{}{"description": "Support Services", "quantity": float64(1), "unit_price": float64(9000)},
		},
	}

	poItems := []map[string]interface{}{
		{"description": "Cloud Hosting", "quantity": float64(1), "unit_price": float64(41000)},
		{"description": "Support Services", "quantity": float64(1), "unit_price": float64(9000)},
	}

	tests := []struct {
		name           string
		fields         map[string]interface{}
		setupMock      func(mock sqlmock.Sqlmock, invID, vendID uuid.UUID)
		wantStatus     models.ValidationStatus
		wantFailureLen int // minimum number of failure reasons expected
	}{
		{
			name:   "happy path — all checks pass",
			fields: validFields,
			setupMock: func(mock sqlmock.Sqlmock, invID, vendID uuid.UUID) {
				expectFirstUpsert(mock)
				mock.ExpectQuery(`SELECT vendor_id`).
					WithArgs(invID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"vendor_id"}).AddRow(vendID.String()))
				mock.ExpectQuery(`SELECT name FROM vendors`).
					WithArgs(vendID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Acme Corp"))
				mock.ExpectQuery(`SELECT id`).
					WithArgs("PO-100").
					WillReturnRows(sqlmock.NewRows([]string{"id", "vendor_id", "total_value", "remaining_value", "line_items", "status"}).
						AddRow(mustUUID(t).String(), vendID.String(), float64(100000), float64(100000), poLineItemsJSON(poItems), "OPEN"))
				mock.ExpectQuery(`SELECT COUNT`).
					WithArgs("INV-001", invID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				expectFinalUpsert(mock)
			},
			wantStatus:     models.ValidationStatusPassed,
			wantFailureLen: 0,
		},
		{
			name: "missing required field — currency",
			fields: func() map[string]interface{} {
				f := copyFields(validFields)
				delete(f, "currency")
				return f
			}(),
			setupMock: func(mock sqlmock.Sqlmock, invID, vendID uuid.UUID) {
				expectFirstUpsert(mock)
				mock.ExpectQuery(`SELECT vendor_id`).
					WithArgs(invID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"vendor_id"}).AddRow(vendID.String()))
				mock.ExpectQuery(`SELECT name FROM vendors`).
					WithArgs(vendID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Acme Corp"))
				mock.ExpectQuery(`SELECT id`).
					WithArgs("PO-100").
					WillReturnRows(sqlmock.NewRows([]string{"id", "vendor_id", "total_value", "remaining_value", "line_items", "status"}).
						AddRow(mustUUID(t).String(), vendID.String(), float64(100000), float64(100000), poLineItemsJSON(poItems), "OPEN"))
				mock.ExpectQuery(`SELECT COUNT`).
					WithArgs("INV-001", invID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				expectFinalUpsert(mock)
			},
			wantStatus:     models.ValidationStatusFailed,
			wantFailureLen: 1,
		},
		{
			name: "total_amount is zero",
			fields: func() map[string]interface{} {
				f := copyFields(validFields)
				f["total_amount"] = float64(0)
				return f
			}(),
			setupMock: func(mock sqlmock.Sqlmock, invID, vendID uuid.UUID) {
				expectFirstUpsert(mock)
				mock.ExpectQuery(`SELECT vendor_id`).
					WithArgs(invID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"vendor_id"}).AddRow(vendID.String()))
				mock.ExpectQuery(`SELECT name FROM vendors`).
					WithArgs(vendID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Acme Corp"))
				mock.ExpectQuery(`SELECT id`).
					WithArgs("PO-100").
					WillReturnRows(sqlmock.NewRows([]string{"id", "vendor_id", "total_value", "remaining_value", "line_items", "status"}).
						AddRow(mustUUID(t).String(), vendID.String(), float64(100000), float64(100000), poLineItemsJSON(poItems), "OPEN"))
				mock.ExpectQuery(`SELECT COUNT`).
					WithArgs("INV-001", invID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				expectFinalUpsert(mock)
			},
			wantStatus:     models.ValidationStatusFailed,
			wantFailureLen: 1,
		},
		{
			name:   "PO not found",
			fields: validFields,
			setupMock: func(mock sqlmock.Sqlmock, invID, vendID uuid.UUID) {
				expectFirstUpsert(mock)
				mock.ExpectQuery(`SELECT vendor_id`).
					WithArgs(invID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"vendor_id"}).AddRow(vendID.String()))
				mock.ExpectQuery(`SELECT name FROM vendors`).
					WithArgs(vendID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Acme Corp"))
				mock.ExpectQuery(`SELECT id`).
					WithArgs("PO-100").
					WillReturnError(sql.ErrNoRows)
				expectFinalUpsert(mock)
			},
			wantStatus:     models.ValidationStatusFailed,
			wantFailureLen: 1,
		},
		{
			name:   "PO is CLOSED",
			fields: validFields,
			setupMock: func(mock sqlmock.Sqlmock, invID, vendID uuid.UUID) {
				expectFirstUpsert(mock)
				mock.ExpectQuery(`SELECT vendor_id`).
					WithArgs(invID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"vendor_id"}).AddRow(vendID.String()))
				mock.ExpectQuery(`SELECT name FROM vendors`).
					WithArgs(vendID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Acme Corp"))
				mock.ExpectQuery(`SELECT id`).
					WithArgs("PO-100").
					WillReturnRows(sqlmock.NewRows([]string{"id", "vendor_id", "total_value", "remaining_value", "line_items", "status"}).
						AddRow(mustUUID(t).String(), vendID.String(), float64(100000), float64(100000), poLineItemsJSON(poItems), "CLOSED"))
				mock.ExpectQuery(`SELECT COUNT`).
					WithArgs("INV-001", invID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				expectFinalUpsert(mock)
			},
			wantStatus:     models.ValidationStatusFailed,
			wantFailureLen: 1,
		},
		{
			name:   "vendor mismatch with PO",
			fields: validFields,
			setupMock: func(mock sqlmock.Sqlmock, invID, vendID uuid.UUID) {
				otherVendorID := mustUUID(t)
				expectFirstUpsert(mock)
				mock.ExpectQuery(`SELECT vendor_id`).
					WithArgs(invID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"vendor_id"}).AddRow(vendID.String()))
				mock.ExpectQuery(`SELECT name FROM vendors`).
					WithArgs(vendID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Acme Corp"))
				mock.ExpectQuery(`SELECT id`).
					WithArgs("PO-100").
					WillReturnRows(sqlmock.NewRows([]string{"id", "vendor_id", "total_value", "remaining_value", "line_items", "status"}).
						AddRow(mustUUID(t).String(), otherVendorID.String(), float64(100000), float64(100000), poLineItemsJSON(poItems), "OPEN"))
				mock.ExpectQuery(`SELECT COUNT`).
					WithArgs("INV-001", invID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				expectFinalUpsert(mock)
			},
			wantStatus:     models.ValidationStatusFailed,
			wantFailureLen: 1,
		},
		{
			name:   "duplicate invoice number",
			fields: validFields,
			setupMock: func(mock sqlmock.Sqlmock, invID, vendID uuid.UUID) {
				expectFirstUpsert(mock)
				mock.ExpectQuery(`SELECT vendor_id`).
					WithArgs(invID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"vendor_id"}).AddRow(vendID.String()))
				mock.ExpectQuery(`SELECT name FROM vendors`).
					WithArgs(vendID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Acme Corp"))
				mock.ExpectQuery(`SELECT id`).
					WithArgs("PO-100").
					WillReturnRows(sqlmock.NewRows([]string{"id", "vendor_id", "total_value", "remaining_value", "line_items", "status"}).
						AddRow(mustUUID(t).String(), vendID.String(), float64(100000), float64(100000), poLineItemsJSON(poItems), "OPEN"))
				mock.ExpectQuery(`SELECT COUNT`).
					WithArgs("INV-001", invID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				expectFinalUpsert(mock)
			},
			wantStatus:     models.ValidationStatusFailed,
			wantFailureLen: 1,
		},
		{
			name:   "invoice total exceeds PO remaining",
			fields: validFields, // total_amount = 50000
			setupMock: func(mock sqlmock.Sqlmock, invID, vendID uuid.UUID) {
				expectFirstUpsert(mock)
				mock.ExpectQuery(`SELECT vendor_id`).
					WithArgs(invID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"vendor_id"}).AddRow(vendID.String()))
				mock.ExpectQuery(`SELECT name FROM vendors`).
					WithArgs(vendID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Acme Corp"))
				// remaining_value = 10000, total_amount = 50000 → should fail
				mock.ExpectQuery(`SELECT id`).
					WithArgs("PO-100").
					WillReturnRows(sqlmock.NewRows([]string{"id", "vendor_id", "total_value", "remaining_value", "line_items", "status"}).
						AddRow(mustUUID(t).String(), vendID.String(), float64(100000), float64(10000), poLineItemsJSON(poItems), "OPEN"))
				mock.ExpectQuery(`SELECT COUNT`).
					WithArgs("INV-001", invID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				expectFinalUpsert(mock)
			},
			wantStatus:     models.ValidationStatusFailed,
			wantFailureLen: 1,
		},
		{
			name: "price drift > 2% → FLAGGED",
			fields: func() map[string]interface{} {
				f := copyFields(validFields)
				// Cloud Hosting price drifted by 5%
				f["line_items"] = []interface{}{
					map[string]interface{}{"description": "Cloud Hosting", "quantity": float64(1), "unit_price": float64(43050)}, // 5% above 41000
					map[string]interface{}{"description": "Support Services", "quantity": float64(1), "unit_price": float64(9000)},
				}
				return f
			}(),
			setupMock: func(mock sqlmock.Sqlmock, invID, vendID uuid.UUID) {
				expectFirstUpsert(mock)
				mock.ExpectQuery(`SELECT vendor_id`).
					WithArgs(invID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"vendor_id"}).AddRow(vendID.String()))
				mock.ExpectQuery(`SELECT name FROM vendors`).
					WithArgs(vendID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Acme Corp"))
				mock.ExpectQuery(`SELECT id`).
					WithArgs("PO-100").
					WillReturnRows(sqlmock.NewRows([]string{"id", "vendor_id", "total_value", "remaining_value", "line_items", "status"}).
						AddRow(mustUUID(t).String(), vendID.String(), float64(100000), float64(100000), poLineItemsJSON(poItems), "OPEN"))
				mock.ExpectQuery(`SELECT COUNT`).
					WithArgs("INV-001", invID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				expectFinalUpsert(mock)
			},
			wantStatus:     models.ValidationStatusFlagged,
			wantFailureLen: 0, // soft — not in FailureReasons for FLAGGED
		},
		{
			name: "vendor name mismatch (soft) → FLAGGED",
			fields: func() map[string]interface{} {
				f := copyFields(validFields)
				f["vendor_name"] = "Acme Corporation" // different from DB "Acme Corp"
				return f
			}(),
			setupMock: func(mock sqlmock.Sqlmock, invID, vendID uuid.UUID) {
				expectFirstUpsert(mock)
				mock.ExpectQuery(`SELECT vendor_id`).
					WithArgs(invID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"vendor_id"}).AddRow(vendID.String()))
				mock.ExpectQuery(`SELECT name FROM vendors`).
					WithArgs(vendID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Acme Corp"))
				mock.ExpectQuery(`SELECT id`).
					WithArgs("PO-100").
					WillReturnRows(sqlmock.NewRows([]string{"id", "vendor_id", "total_value", "remaining_value", "line_items", "status"}).
						AddRow(mustUUID(t).String(), vendID.String(), float64(100000), float64(100000), poLineItemsJSON(poItems), "OPEN"))
				mock.ExpectQuery(`SELECT COUNT`).
					WithArgs("INV-001", invID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				expectFinalUpsert(mock)
			},
			wantStatus:     models.ValidationStatusFlagged,
			wantFailureLen: 0,
		},
		{
			name: "due_date before invoice_date",
			fields: func() map[string]interface{} {
				f := copyFields(validFields)
				f["invoice_date"] = "2026-03-31"
				f["due_date"] = "2026-03-01" // before invoice_date
				return f
			}(),
			setupMock: func(mock sqlmock.Sqlmock, invID, vendID uuid.UUID) {
				expectFirstUpsert(mock)
				mock.ExpectQuery(`SELECT vendor_id`).
					WithArgs(invID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"vendor_id"}).AddRow(vendID.String()))
				mock.ExpectQuery(`SELECT name FROM vendors`).
					WithArgs(vendID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Acme Corp"))
				mock.ExpectQuery(`SELECT id`).
					WithArgs("PO-100").
					WillReturnRows(sqlmock.NewRows([]string{"id", "vendor_id", "total_value", "remaining_value", "line_items", "status"}).
						AddRow(mustUUID(t).String(), vendID.String(), float64(100000), float64(100000), poLineItemsJSON(poItems), "OPEN"))
				mock.ExpectQuery(`SELECT COUNT`).
					WithArgs("INV-001", invID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				expectFinalUpsert(mock)
			},
			wantStatus:     models.ValidationStatusFailed,
			wantFailureLen: 1,
		},
		{
			name: "line item not in PO → FAILED",
			fields: func() map[string]interface{} {
				f := copyFields(validFields)
				f["line_items"] = []interface{}{
					map[string]interface{}{"description": "Unknown Widget", "quantity": float64(1), "unit_price": float64(5000)},
				}
				return f
			}(),
			setupMock: func(mock sqlmock.Sqlmock, invID, vendID uuid.UUID) {
				expectFirstUpsert(mock)
				mock.ExpectQuery(`SELECT vendor_id`).
					WithArgs(invID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"vendor_id"}).AddRow(vendID.String()))
				mock.ExpectQuery(`SELECT name FROM vendors`).
					WithArgs(vendID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Acme Corp"))
				mock.ExpectQuery(`SELECT id`).
					WithArgs("PO-100").
					WillReturnRows(sqlmock.NewRows([]string{"id", "vendor_id", "total_value", "remaining_value", "line_items", "status"}).
						AddRow(mustUUID(t).String(), vendID.String(), float64(100000), float64(100000), poLineItemsJSON(poItems), "OPEN"))
				mock.ExpectQuery(`SELECT COUNT`).
					WithArgs("INV-001", invID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				expectFinalUpsert(mock)
			},
			wantStatus:     models.ValidationStatusFailed,
			wantFailureLen: 1,
		},
		{
			name:   "vendor not found in DB",
			fields: validFields,
			setupMock: func(mock sqlmock.Sqlmock, invID, vendID uuid.UUID) {
				expectFirstUpsert(mock)
				mock.ExpectQuery(`SELECT vendor_id`).
					WithArgs(invID.String()).
					WillReturnRows(sqlmock.NewRows([]string{"vendor_id"}).AddRow(vendID.String()))
				mock.ExpectQuery(`SELECT name FROM vendors`).
					WithArgs(vendID.String()).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectQuery(`SELECT id`).
					WithArgs("PO-100").
					WillReturnRows(sqlmock.NewRows([]string{"id", "vendor_id", "total_value", "remaining_value", "line_items", "status"}).
						AddRow(mustUUID(t).String(), vendID.String(), float64(100000), float64(100000), poLineItemsJSON(poItems), "OPEN"))
				mock.ExpectQuery(`SELECT COUNT`).
					WithArgs("INV-001", invID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				expectFinalUpsert(mock)
			},
			wantStatus:     models.ValidationStatusFailed,
			wantFailureLen: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o, mock := newTestOrchestrator(t)
			tc.setupMock(mock, invoiceID, vendorID)

			rec, err := o.runValidation(context.Background(), invoiceID, tc.fields)
			if err != nil {
				t.Fatalf("runValidation returned unexpected error: %v", err)
			}
			if rec.ValidationStatus != tc.wantStatus {
				t.Errorf("ValidationStatus = %q, want %q\nSummary: %s\nFailures: %v",
					rec.ValidationStatus, tc.wantStatus, rec.Summary, rec.FailureReasons)
			}
			if len(rec.FailureReasons) < tc.wantFailureLen {
				t.Errorf("FailureReasons len = %d, want >= %d; reasons: %v",
					len(rec.FailureReasons), tc.wantFailureLen, rec.FailureReasons)
			}
			if rec.CompletedAt == nil {
				t.Error("CompletedAt should be set after runValidation")
			}
			if len(rec.CheckResults) == 0 {
				t.Error("CheckResults should not be empty")
			}
		})
	}
}

// ── makeDecision ──────────────────────────────────────────────────────────────

func TestMakeDecision(t *testing.T) {
	o := &Orchestrator{logger: zap.NewNop()}

	tests := []struct {
		status     models.ValidationStatus
		wantAction string
	}{
		{models.ValidationStatusFailed, "REJECT"},
		{models.ValidationStatusFlagged, "FLAG"},
		{models.ValidationStatusPassed, "APPROVE"},
		{models.ValidationStatusRunning, "APPROVE"}, // default branch
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("status=%s", tc.status), func(t *testing.T) {
			v := &models.InvoiceValidation{
				ValidationStatus: tc.status,
				Summary:          "test summary",
			}
			d := o.makeDecision(v)
			if d.Action != tc.wantAction {
				t.Errorf("makeDecision(%s).Action = %q, want %q", tc.status, d.Action, tc.wantAction)
			}
			if d.Reason == "" {
				t.Error("Decision.Reason should not be empty")
			}
		})
	}
}

// ── decisionActionToStatus ────────────────────────────────────────────────────

func TestDecisionActionToStatus(t *testing.T) {
	tests := []struct {
		action string
		want   models.InvoiceStatus
	}{
		{"APPROVE", models.InvoiceStatusApproved},
		{"FLAG", models.InvoiceStatusFlagged},
		{"REJECT", models.InvoiceStatusRejected},
		{"UNKNOWN", models.InvoiceStatusFlagged}, // default
		{"", models.InvoiceStatusFlagged},
	}

	for _, tc := range tests {
		got := decisionActionToStatus(tc.action)
		if got != tc.want {
			t.Errorf("decisionActionToStatus(%q) = %q, want %q", tc.action, got, tc.want)
		}
	}
}

// ── amount tolerance boundary ─────────────────────────────────────────────────

func TestAmountWithinPOTolerance(t *testing.T) {
	// 2% tolerance: invoice total up to poRemaining * 1.02 is accepted
	vendorID := mustUUID(t)
	invoiceID := mustUUID(t)

	poItems := []map[string]interface{}{
		{"description": "Widget", "quantity": float64(1), "unit_price": float64(100000)},
	}

	tests := []struct {
		name        string
		invoiceAmt  float64
		poRemaining float64
		wantPass    bool
	}{
		{"exactly equal", 100000, 100000, true},
		{"within 2%", 101999, 100000, true},  // 101999 / 100000 = 1.01999 < 1.02
		{"exactly 2%", 102000, 100000, true},  // 102000 / 100000 = 1.02 — boundary (not strictly greater)
		{"just over 2%", 102001, 100000, false},
		{"well under", 50000, 100000, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fields := map[string]interface{}{
				"vendor_name":    "Acme Corp",
				"invoice_number": "INV-TOL",
				"po_reference":   "PO-TOL",
				"total_amount":   tc.invoiceAmt,
				"currency":       "INR",
				"invoice_date":   "2026-03-01",
				"due_date":       "2026-03-31",
				"line_items": []interface{}{
					map[string]interface{}{"description": "Widget", "quantity": float64(1), "unit_price": tc.invoiceAmt},
				},
			}

			o, mock := newTestOrchestrator(t)
			expectFirstUpsert(mock)
			mock.ExpectQuery(`SELECT vendor_id`).WithArgs(invoiceID.String()).
				WillReturnRows(sqlmock.NewRows([]string{"vendor_id"}).AddRow(vendorID.String()))
			mock.ExpectQuery(`SELECT name FROM vendors`).WithArgs(vendorID.String()).
				WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Acme Corp"))
			mock.ExpectQuery(`SELECT id`).WithArgs("PO-TOL").
				WillReturnRows(sqlmock.NewRows([]string{"id", "vendor_id", "total_value", "remaining_value", "line_items", "status"}).
					AddRow(mustUUID(t).String(), vendorID.String(), tc.poRemaining, tc.poRemaining, poLineItemsJSON(poItems), "OPEN"))
			mock.ExpectQuery(`SELECT COUNT`).WithArgs("INV-TOL", invoiceID).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			expectFinalUpsert(mock)

			rec, err := o.runValidation(context.Background(), invoiceID, fields)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Find the amount_within_po check result
			var amountCheck *models.CheckResult
			for i := range rec.CheckResults {
				if rec.CheckResults[i].Check == "amount_within_po" {
					amountCheck = &rec.CheckResults[i]
					break
				}
			}
			if amountCheck == nil {
				t.Fatal("amount_within_po check not found in results")
			}
			if amountCheck.Passed != tc.wantPass {
				t.Errorf("amount_within_po.Passed = %v, want %v (inv=%.0f, remaining=%.0f)",
					amountCheck.Passed, tc.wantPass, tc.invoiceAmt, tc.poRemaining)
			}
		})
	}
}

// ── copyFields makes a shallow copy of the fields map ────────────────────────

func copyFields(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
