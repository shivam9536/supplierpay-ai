package models

import (
	"time"

	"github.com/google/uuid"
)

// ── Audit Log ───────────────────────────────
type AuditLogStep string

const (
	AuditStepExtract        AuditLogStep = "EXTRACT"
	AuditStepValidate       AuditLogStep = "VALIDATE"
	AuditStepCrossReference AuditLogStep = "CROSS_REFERENCE"
	AuditStepDecision       AuditLogStep = "DECISION"
	AuditStepDraftQuery     AuditLogStep = "DRAFT_QUERY"
	AuditStepSchedule       AuditLogStep = "SCHEDULE"
	AuditStepDisburse       AuditLogStep = "DISBURSE"
)

type AuditLog struct {
	ID              uuid.UUID    `db:"id" json:"id"`
	InvoiceID       uuid.UUID    `db:"invoice_id" json:"invoice_id"`
	Step            AuditLogStep `db:"step" json:"step"`
	Result          string       `db:"result" json:"result"`
	Reasoning       string       `db:"reasoning" json:"reasoning"`
	ConfidenceScore float64      `db:"confidence_score" json:"confidence_score"`
	DurationMs      int          `db:"duration_ms" json:"duration_ms"`
	CreatedAt       time.Time    `db:"created_at" json:"created_at"`
}
