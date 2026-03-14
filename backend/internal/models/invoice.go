package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ── Invoice Status ──────────────────────────
type InvoiceStatus string

const (
	InvoiceStatusPending    InvoiceStatus = "PENDING"
	InvoiceStatusExtracting InvoiceStatus = "EXTRACTING"
	InvoiceStatusValidating InvoiceStatus = "VALIDATING"
	InvoiceStatusApproved   InvoiceStatus = "APPROVED"
	InvoiceStatusFlagged    InvoiceStatus = "FLAGGED"
	InvoiceStatusRejected   InvoiceStatus = "REJECTED"
	InvoiceStatusScheduled  InvoiceStatus = "SCHEDULED"
	InvoiceStatusPaid       InvoiceStatus = "PAID"
)

// ── JSONB Helper ────────────────────────────
// JSONB handles object-shaped JSON columns (e.g. extracted_fields, discrepancies).
type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// ── RawJSON handles any JSON column (object or array) ──────────────────────
type RawJSON json.RawMessage

func (r RawJSON) MarshalJSON() ([]byte, error) {
	if len(r) == 0 {
		return []byte("null"), nil
	}
	return json.RawMessage(r).MarshalJSON()
}

func (r *RawJSON) Scan(value interface{}) error {
	if value == nil {
		*r = RawJSON("null")
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	*r = make(RawJSON, len(bytes))
	copy(*r, bytes)
	return nil
}

func (r RawJSON) Value() (driver.Value, error) {
	if len(r) == 0 {
		return nil, nil
	}
	return []byte(r), nil
}

// ── Invoice ─────────────────────────────────
type Invoice struct {
	ID                    uuid.UUID     `db:"id" json:"id"`
	VendorID              uuid.UUID     `db:"vendor_id" json:"vendor_id"`
	InvoiceNumber         string        `db:"invoice_number" json:"invoice_number"`
	POReference           string        `db:"po_reference" json:"po_reference"`
	RawFileURL            *string       `db:"raw_file_url" json:"raw_file_url,omitempty"`
	ExtractedFields       JSONB         `db:"extracted_fields" json:"extracted_fields"`
	LineItems             RawJSON       `db:"line_items" json:"line_items"`
	TotalAmount           float64       `db:"total_amount" json:"total_amount"`
	TaxAmount             float64       `db:"tax_amount" json:"tax_amount"`
	Currency              string        `db:"currency" json:"currency"`
	InvoiceDate           *time.Time    `db:"invoice_date" json:"invoice_date,omitempty"`
	DueDate               *time.Time    `db:"due_date" json:"due_date,omitempty"`
	Status                InvoiceStatus `db:"status" json:"status"`
	Discrepancies         RawJSON       `db:"discrepancies" json:"discrepancies"`
	DecisionReason        string        `db:"decision_reason" json:"decision_reason"`
	ScheduledPaymentDate  *time.Time    `db:"scheduled_payment_date" json:"scheduled_payment_date"`
	PineLabsTransactionID *string       `db:"pinelabs_transaction_id" json:"pinelabs_transaction_id,omitempty"`
	CreatedAt             time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time     `db:"updated_at" json:"updated_at"`
}

// InvoiceWithVendor is used for list responses with vendor name joined.
type InvoiceWithVendor struct {
	Invoice
	VendorName string `db:"vendor_name" json:"vendor_name"`
}

// ── Invoice Create Request ──────────────────
type InvoiceUploadRequest struct {
	VendorID      string `json:"vendor_id" binding:"required"`
	InvoiceNumber string `json:"invoice_number"`
	POReference   string `json:"po_reference"`
}
