package models

import (
	"time"

	"github.com/google/uuid"
)

// ── Payment Run ─────────────────────────────
type PaymentRunStatus string

const (
	PaymentRunPending        PaymentRunStatus = "PENDING"
	PaymentRunExecuting      PaymentRunStatus = "EXECUTING"
	PaymentRunCompleted      PaymentRunStatus = "COMPLETED"
	PaymentRunPartialFailure PaymentRunStatus = "PARTIAL_FAILURE"
)

type PaymentRun struct {
	ID              uuid.UUID        `db:"id" json:"id"`
	RunDate         time.Time        `db:"run_date" json:"run_date"`
	TotalAmount     float64          `db:"total_amount" json:"total_amount"`
	InvoiceCount    int              `db:"invoice_count" json:"invoice_count"`
	Status          PaymentRunStatus `db:"status" json:"status"`
	PineLabsBatchID string           `db:"pinelabs_batch_id" json:"pinelabs_batch_id"`
	CreatedAt       time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time        `db:"updated_at" json:"updated_at"`
}

// ── Goods Receipt (mock for 3-way matching) ─
type GoodsReceipt struct {
	ID               uuid.UUID `db:"id" json:"id"`
	PONumber         string    `db:"po_number" json:"po_number"`
	ReceivedQuantity JSONB     `db:"received_quantity" json:"received_quantity"`
	ReceivedDate     time.Time `db:"received_date" json:"received_date"`
	Status           string    `db:"status" json:"status"` // RECEIVED, PARTIAL
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}
