package models

import (
	"time"

	"github.com/google/uuid"
)

// ── PO Status ───────────────────────────────
type POStatus string

const (
	POStatusOpen             POStatus = "OPEN"
	POStatusPartiallyMatched POStatus = "PARTIALLY_MATCHED"
	POStatusClosed           POStatus = "CLOSED"
)

// ── Purchase Order ──────────────────────────
type PurchaseOrder struct {
	ID             uuid.UUID `db:"id" json:"id"`
	PONumber       string    `db:"po_number" json:"po_number"`
	VendorID       uuid.UUID `db:"vendor_id" json:"vendor_id"`
	TotalValue     float64   `db:"total_value" json:"total_value"`
	RemainingValue float64   `db:"remaining_value" json:"remaining_value"`
	LineItems      RawJSON   `db:"line_items" json:"line_items"`
	ApprovedBy     *string   `db:"approved_by" json:"approved_by"`
	Status         POStatus  `db:"status" json:"status"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}
