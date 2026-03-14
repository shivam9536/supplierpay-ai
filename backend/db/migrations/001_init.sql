-- ============================================
-- SupplierPay AI — Database Schema
-- ============================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ── Vendors ─────────────────────────────────
CREATE TABLE IF NOT EXISTS vendors (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name                    VARCHAR(255) NOT NULL,
    email                   VARCHAR(255) NOT NULL,
    bank_account_number     VARCHAR(50),
    bank_ifsc               VARCHAR(20),
    payment_terms_days      INTEGER DEFAULT 30,
    early_payment_discount  DECIMAL(5,2) DEFAULT 0,      -- e.g., 2.00 for 2%
    early_payment_days      INTEGER DEFAULT 10,           -- e.g., 10 for "2/10 net 30"
    criticality_score       INTEGER DEFAULT 5 CHECK (criticality_score BETWEEN 1 AND 10),
    created_at              TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at              TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ── Purchase Orders ─────────────────────────
CREATE TABLE IF NOT EXISTS purchase_orders (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    po_number         VARCHAR(50) UNIQUE NOT NULL,
    vendor_id         UUID NOT NULL REFERENCES vendors(id),
    total_value       DECIMAL(15,2) NOT NULL,
    remaining_value   DECIMAL(15,2) NOT NULL,
    line_items        JSONB DEFAULT '[]',
    approved_by       VARCHAR(255),
    status            VARCHAR(20) DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'PARTIALLY_MATCHED', 'CLOSED')),
    created_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ── Goods Receipts (for 3-way matching) ─────
CREATE TABLE IF NOT EXISTS goods_receipts (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    po_number           VARCHAR(50) NOT NULL REFERENCES purchase_orders(po_number),
    received_quantity   JSONB DEFAULT '[]',
    received_date       DATE NOT NULL,
    status              VARCHAR(20) DEFAULT 'RECEIVED' CHECK (status IN ('RECEIVED', 'PARTIAL')),
    created_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ── Invoices ────────────────────────────────
CREATE TABLE IF NOT EXISTS invoices (
    id                        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vendor_id                 UUID REFERENCES vendors(id),
    invoice_number            VARCHAR(100),
    po_reference              VARCHAR(50),
    raw_file_url              TEXT,
    extracted_fields          JSONB DEFAULT '{}',
    line_items                JSONB DEFAULT '[]',
    total_amount              DECIMAL(15,2) DEFAULT 0,
    tax_amount                DECIMAL(15,2) DEFAULT 0,
    currency                  VARCHAR(10) DEFAULT 'INR',
    invoice_date              DATE,
    due_date                  DATE,
    status                    VARCHAR(20) DEFAULT 'PENDING'
                              CHECK (status IN ('PENDING', 'EXTRACTING', 'VALIDATING', 'APPROVED', 'FLAGGED', 'REJECTED', 'SCHEDULED', 'PAID')),
    discrepancies             JSONB DEFAULT '[]',
    decision_reason           TEXT,
    scheduled_payment_date    DATE,
    pinelabs_transaction_id   VARCHAR(100),
    created_at                TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at                TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ── Invoice Validations ─────────────────────
CREATE TABLE IF NOT EXISTS invoice_validations (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_id          UUID NOT NULL UNIQUE REFERENCES invoices(id) ON DELETE CASCADE,
    validation_status   VARCHAR(20) NOT NULL DEFAULT 'PENDING'
                        CHECK (validation_status IN ('PENDING','RUNNING','PASSED','FAILED','FLAGGED')),
    vendor_valid        BOOLEAN,
    po_found            BOOLEAN,
    po_open             BOOLEAN,
    vendor_matches_po   BOOLEAN,
    items_match         BOOLEAN,
    prices_match        BOOLEAN,
    amount_within_po    BOOLEAN,
    no_duplicate        BOOLEAN,
    check_results       JSONB DEFAULT '[]',
    line_item_results   JSONB DEFAULT '[]',
    matched_po_number   VARCHAR(50),
    matched_po_id       UUID,
    summary             TEXT DEFAULT '',
    failure_reasons     JSONB DEFAULT '[]',
    started_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at        TIMESTAMP WITH TIME ZONE,
    created_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invoice_validations_invoice ON invoice_validations(invoice_id);
CREATE INDEX IF NOT EXISTS idx_invoice_validations_status ON invoice_validations(validation_status);

-- ── Audit Log ───────────────────────────────
CREATE TABLE IF NOT EXISTS audit_logs (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_id        UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    step              VARCHAR(30) NOT NULL
                      CHECK (step IN ('EXTRACT', 'VALIDATE', 'CROSS_REFERENCE', 'DECISION', 'DRAFT_QUERY', 'SCHEDULE', 'DISBURSE')),
    result            VARCHAR(20) NOT NULL,         -- completed, failed, skipped
    reasoning         TEXT,                         -- LLM explanation or rule description
    confidence_score  DECIMAL(5,4) DEFAULT 0,       -- 0.0000 to 1.0000
    duration_ms       INTEGER DEFAULT 0,
    created_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ── Payment Runs ────────────────────────────
CREATE TABLE IF NOT EXISTS payment_runs (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    run_date            DATE NOT NULL,
    total_amount        DECIMAL(15,2) DEFAULT 0,
    invoice_count       INTEGER DEFAULT 0,
    status              VARCHAR(20) DEFAULT 'PENDING'
                        CHECK (status IN ('PENDING', 'EXECUTING', 'COMPLETED', 'PARTIAL_FAILURE')),
    pinelabs_batch_id   VARCHAR(100),
    created_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ── Payment Run ↔ Invoice Junction ──────────
CREATE TABLE IF NOT EXISTS payment_run_invoices (
    payment_run_id    UUID NOT NULL REFERENCES payment_runs(id),
    invoice_id        UUID NOT NULL REFERENCES invoices(id),
    status            VARCHAR(20) DEFAULT 'PENDING',
    PRIMARY KEY (payment_run_id, invoice_id)
);

-- ── Indexes ─────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status);
CREATE INDEX IF NOT EXISTS idx_invoices_vendor ON invoices(vendor_id);
CREATE INDEX IF NOT EXISTS idx_invoices_po_ref ON invoices(po_reference);
CREATE INDEX IF NOT EXISTS idx_invoices_payment_date ON invoices(scheduled_payment_date);
CREATE INDEX IF NOT EXISTS idx_invoices_created ON invoices(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_invoice ON audit_logs(invoice_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_step ON audit_logs(step);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_vendor ON purchase_orders(vendor_id);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_number ON purchase_orders(po_number);
CREATE INDEX IF NOT EXISTS idx_goods_receipts_po ON goods_receipts(po_number);
CREATE INDEX IF NOT EXISTS idx_payment_runs_date ON payment_runs(run_date);

-- ── Updated At Trigger ──────────────────────
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_vendors_updated_at
    BEFORE UPDATE ON vendors
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_purchase_orders_updated_at
    BEFORE UPDATE ON purchase_orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_invoices_updated_at
    BEFORE UPDATE ON invoices
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_payment_runs_updated_at
    BEFORE UPDATE ON payment_runs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_invoice_validations_updated_at
    BEFORE UPDATE ON invoice_validations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
