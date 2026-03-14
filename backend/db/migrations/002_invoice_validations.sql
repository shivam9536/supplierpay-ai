-- ============================================
-- SupplierPay AI — Invoice Validations Table
-- ============================================
-- This table stores the full result of every validation run against an
-- invoice.  The invoices table is intentionally NOT modified by the
-- validation step; this table is the authoritative record of validation
-- state, checks performed, and per-check results.
-- ============================================

CREATE TABLE IF NOT EXISTS invoice_validations (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_id          UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,

    -- Overall validation outcome
    -- PENDING   → row created, checks not yet run
    -- RUNNING   → checks in progress
    -- PASSED    → all checks passed
    -- FAILED    → one or more checks failed
    -- FLAGGED   → passed structurally but has soft discrepancies (price/qty drift)
    validation_status   VARCHAR(20) NOT NULL DEFAULT 'PENDING'
                        CHECK (validation_status IN ('PENDING','RUNNING','PASSED','FAILED','FLAGGED')),

    -- ── Per-check boolean outcomes ───────────────────────────
    vendor_valid        BOOLEAN,          -- vendor_id exists in vendors table
    po_found            BOOLEAN,          -- po_reference matches a purchase_order row
    po_open             BOOLEAN,          -- PO status is OPEN or PARTIALLY_MATCHED
    vendor_matches_po   BOOLEAN,          -- invoice.vendor_id == purchase_order.vendor_id
    items_match         BOOLEAN,          -- every invoice line item found in PO
    prices_match        BOOLEAN,          -- all matched item prices within tolerance
    amount_within_po    BOOLEAN,          -- invoice total <= PO remaining_value (+2% grace)
    no_duplicate        BOOLEAN,          -- no other non-rejected invoice with same number

    -- ── Detailed check results (JSON arrays) ─────────────────
    -- Each entry: {"check": "...", "passed": true/false, "detail": "..."}
    check_results       JSONB NOT NULL DEFAULT '[]',

    -- ── Matched PO snapshot (for auditability) ───────────────
    matched_po_number   VARCHAR(50),
    matched_po_id       UUID,

    -- ── Line-item diff (what matched, what didn't) ───────────
    -- Array of: {"description":"...","inv_qty":1,"po_qty":1,
    --            "inv_price":41000,"po_price":41000,"matched":true,"note":""}
    line_item_results   JSONB NOT NULL DEFAULT '[]',

    -- ── Human-readable summary ───────────────────────────────
    summary             TEXT,
    failure_reasons     JSONB NOT NULL DEFAULT '[]',   -- []string of error messages

    -- ── Timing ───────────────────────────────────────────────
    started_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at        TIMESTAMP WITH TIME ZONE,
    created_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- One active validation record per invoice (latest wins via ON CONFLICT upsert)
CREATE UNIQUE INDEX IF NOT EXISTS idx_inv_validations_invoice ON invoice_validations(invoice_id);
CREATE INDEX IF NOT EXISTS idx_inv_validations_status   ON invoice_validations(validation_status);
CREATE INDEX IF NOT EXISTS idx_inv_validations_po       ON invoice_validations(matched_po_number);

-- updated_at auto-trigger
CREATE TRIGGER trigger_invoice_validations_updated_at
    BEFORE UPDATE ON invoice_validations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
