package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ── Validation Status ────────────────────────────────────────────────────────
type ValidationStatus string

const (
	ValidationStatusPending ValidationStatus = "PENDING"
	ValidationStatusRunning ValidationStatus = "RUNNING"
	ValidationStatusPassed  ValidationStatus = "PASSED"
	ValidationStatusFailed  ValidationStatus = "FAILED"
	ValidationStatusFlagged ValidationStatus = "FLAGGED" // soft discrepancies
)

// ── CheckResult is a single named check stored in check_results JSONB ────────
type CheckResult struct {
	Check  string `json:"check"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail"`
}

// CheckResults is a slice with custom DB scan / value support.
type CheckResults []CheckResult

func (c CheckResults) Value() (driver.Value, error) {
	return json.Marshal(c)
}
func (c *CheckResults) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// ── LineItemResult is one row of the per-item diff ───────────────────────────
type LineItemResult struct {
	Description string  `json:"description"`
	InvQty      float64 `json:"inv_qty"`
	POQty       float64 `json:"po_qty"`
	InvPrice    float64 `json:"inv_price"`
	POPrice     float64 `json:"po_price"`
	Matched     bool    `json:"matched"`
	Note        string  `json:"note,omitempty"`
}

// LineItemResults is a slice with custom DB scan / value support.
type LineItemResults []LineItemResult

func (l LineItemResults) Value() (driver.Value, error) {
	return json.Marshal(l)
}
func (l *LineItemResults) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, l)
}

// ── StringSlice for failure_reasons JSONB ────────────────────────────────────
type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	return json.Marshal(s)
}
func (s *StringSlice) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, s)
}

// ── InvoiceValidation maps to the invoice_validations table ─────────────────
type InvoiceValidation struct {
	ID               uuid.UUID        `db:"id"                json:"id"`
	InvoiceID        uuid.UUID        `db:"invoice_id"        json:"invoice_id"`
	ValidationStatus ValidationStatus `db:"validation_status" json:"validation_status"`

	// Per-check booleans (nullable — nil means "not yet evaluated")
	VendorValid     *bool `db:"vendor_valid"      json:"vendor_valid"`
	POFound         *bool `db:"po_found"          json:"po_found"`
	POOpen          *bool `db:"po_open"           json:"po_open"`
	VendorMatchesPO *bool `db:"vendor_matches_po" json:"vendor_matches_po"`
	ItemsMatch      *bool `db:"items_match"       json:"items_match"`
	PricesMatch     *bool `db:"prices_match"      json:"prices_match"`
	AmountWithinPO  *bool `db:"amount_within_po"  json:"amount_within_po"`
	NoDuplicate     *bool `db:"no_duplicate"      json:"no_duplicate"`

	// Detailed results
	CheckResults    CheckResults    `db:"check_results"    json:"check_results"`
	LineItemResults LineItemResults `db:"line_item_results" json:"line_item_results"`

	// Matched PO snapshot
	MatchedPONumber *string    `db:"matched_po_number" json:"matched_po_number,omitempty"`
	MatchedPOID     *uuid.UUID `db:"matched_po_id"     json:"matched_po_id,omitempty"`

	// Summary
	Summary        string      `db:"summary"         json:"summary"`
	FailureReasons StringSlice `db:"failure_reasons" json:"failure_reasons"`

	// Timing
	StartedAt   time.Time  `db:"started_at"   json:"started_at"`
	CompletedAt *time.Time `db:"completed_at" json:"completed_at,omitempty"`
	CreatedAt   time.Time  `db:"created_at"   json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"   json:"updated_at"`
}
