package models

import (
	"time"

	"github.com/google/uuid"
)

// ── Vendor ──────────────────────────────────
type Vendor struct {
	ID                   uuid.UUID `db:"id" json:"id"`
	Name                 string    `db:"name" json:"name"`
	Email                string    `db:"email" json:"email"`
	BankAccountNumber    string    `db:"bank_account_number" json:"bank_account_number"`
	BankIFSC             string    `db:"bank_ifsc" json:"bank_ifsc"`
	PaymentTermsDays     int       `db:"payment_terms_days" json:"payment_terms_days"`
	EarlyPaymentDiscount float64   `db:"early_payment_discount" json:"early_payment_discount"`
	EarlyPaymentDays     int       `db:"early_payment_days" json:"early_payment_days"`
	CriticalityScore     int       `db:"criticality_score" json:"criticality_score"` // 1-10
	CreatedAt            time.Time `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time `db:"updated_at" json:"updated_at"`
}
