-- ============================================
-- SupplierPay AI — Seed Data for Demo
-- ============================================

-- ── Vendors ─────────────────────────────────
INSERT INTO vendors (id, name, email, bank_account_number, bank_ifsc, payment_terms_days, early_payment_discount, early_payment_days, criticality_score) VALUES
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'Acme Cloud Solutions',     'billing@acmecloud.com',     '1234567890', 'HDFC0001234', 30, 2.00, 10, 9),
('b2c3d4e5-f6a7-8901-bcde-f12345678901', 'TechParts India Pvt Ltd',  'accounts@techparts.in',     '2345678901', 'ICIC0002345', 30, 0.00,  0, 7),
('c3d4e5f6-a7b8-9012-cdef-123456789012', 'Global Office Supplies',   'finance@globalsupply.com',  '3456789012', 'SBIN0003456', 45, 1.50, 15, 5),
('d4e5f6a7-b8c9-0123-defa-234567890123', 'SecureNet Cybersecurity',  'invoices@securenet.io',     '4567890123', 'AXIS0004567', 30, 2.50, 10, 10),
('e5f6a7b8-c9d0-1234-efab-345678901234', 'DataFlow Analytics',       'ap@dataflow.co.in',        '5678901234', 'KOTK0005678', 60, 0.00,  0, 6)
ON CONFLICT DO NOTHING;

-- ── Purchase Orders ─────────────────────────
INSERT INTO purchase_orders (id, po_number, vendor_id, total_value, remaining_value, line_items, approved_by, status) VALUES
(
    '11111111-1111-1111-1111-111111111111',
    'PO-2026-100',
    'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
    50000.00,
    50000.00,
    '[{"description": "Cloud Hosting - Monthly", "quantity": 1, "unit_price": 41000.00, "total": 41000.00}, {"description": "Support Services", "quantity": 1, "unit_price": 9000.00, "total": 9000.00}]',
    'Rahul Sharma',
    'OPEN'
),
(
    '22222222-2222-2222-2222-222222222222',
    'PO-2026-101',
    'b2c3d4e5-f6a7-8901-bcde-f12345678901',
    125000.00,
    125000.00,
    '[{"description": "Server Rack Components", "quantity": 10, "unit_price": 8500.00, "total": 85000.00}, {"description": "Network Cables (Cat6)", "quantity": 200, "unit_price": 200.00, "total": 40000.00}]',
    'Priya Patel',
    'OPEN'
),
(
    '33333333-3333-3333-3333-333333333333',
    'PO-2026-102',
    'c3d4e5f6-a7b8-9012-cdef-123456789012',
    35000.00,
    35000.00,
    '[{"description": "Office Chairs (Ergonomic)", "quantity": 10, "unit_price": 2500.00, "total": 25000.00}, {"description": "Standing Desks", "quantity": 5, "unit_price": 2000.00, "total": 10000.00}]',
    'Amit Kumar',
    'OPEN'
),
(
    '44444444-4444-4444-4444-444444444444',
    'PO-2026-103',
    'd4e5f6a7-b8c9-0123-defa-234567890123',
    200000.00,
    200000.00,
    '[{"description": "Annual Security Audit", "quantity": 1, "unit_price": 150000.00, "total": 150000.00}, {"description": "Penetration Testing", "quantity": 1, "unit_price": 50000.00, "total": 50000.00}]',
    'Sneha Reddy',
    'OPEN'
),
(
    '55555555-5555-5555-5555-555555555555',
    'PO-2026-104',
    'e5f6a7b8-c9d0-1234-efab-345678901234',
    75000.00,
    75000.00,
    '[{"description": "Data Pipeline Setup", "quantity": 1, "unit_price": 50000.00, "total": 50000.00}, {"description": "Dashboard Development", "quantity": 1, "unit_price": 25000.00, "total": 25000.00}]',
    'Rahul Sharma',
    'OPEN'
),
-- Extra POs for testing edge cases
(
    '66666666-6666-6666-6666-666666666666',
    'PO-2026-105',
    'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
    80000.00,
    30000.00,
    '[{"description": "Cloud Migration Services", "quantity": 1, "unit_price": 80000.00, "total": 80000.00}]',
    'Priya Patel',
    'PARTIALLY_MATCHED'
),
(
    '77777777-7777-7777-7777-777777777777',
    'PO-2026-106',
    'b2c3d4e5-f6a7-8901-bcde-f12345678901',
    45000.00,
    0.00,
    '[{"description": "Laptop Batteries", "quantity": 50, "unit_price": 900.00, "total": 45000.00}]',
    'Amit Kumar',
    'CLOSED'
)
ON CONFLICT DO NOTHING;

-- ── Goods Receipts (for 3-way matching) ─────
INSERT INTO goods_receipts (po_number, received_quantity, received_date, status) VALUES
('PO-2026-100', '[{"description": "Cloud Hosting - Monthly", "quantity": 1}, {"description": "Support Services", "quantity": 1}]', '2026-03-05', 'RECEIVED'),
('PO-2026-101', '[{"description": "Server Rack Components", "quantity": 10}, {"description": "Network Cables (Cat6)", "quantity": 200}]', '2026-03-08', 'RECEIVED'),
('PO-2026-102', '[{"description": "Office Chairs (Ergonomic)", "quantity": 10}, {"description": "Standing Desks", "quantity": 5}]', '2026-03-10', 'RECEIVED'),
('PO-2026-103', '[{"description": "Annual Security Audit", "quantity": 1}, {"description": "Penetration Testing", "quantity": 1}]', '2026-03-01', 'RECEIVED'),
('PO-2026-104', '[{"description": "Data Pipeline Setup", "quantity": 1}]', '2026-03-11', 'PARTIAL')
ON CONFLICT DO NOTHING;

-- ── Sample Invoices (various states for demo) ────
INSERT INTO invoices (id, vendor_id, invoice_number, po_reference, total_amount, tax_amount, currency, invoice_date, due_date, status, line_items, decision_reason, scheduled_payment_date) VALUES
(
    'aaaa1111-1111-1111-1111-111111111111',
    'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
    'INV-ACME-2026-042',
    'PO-2026-100',
    50000.00, 9000.00, 'INR',
    '2026-03-01', '2026-03-31',
    'APPROVED',
    '[{"description": "Cloud Hosting - Monthly", "quantity": 1, "unit_price": 41000.00, "total": 41000.00}, {"description": "Support Services", "quantity": 1, "unit_price": 9000.00, "total": 9000.00}]',
    'All checks passed — PO matched, amounts verified, line items consistent. Auto-approved.',
    '2026-03-29'
),
(
    'bbbb2222-2222-2222-2222-222222222222',
    'b2c3d4e5-f6a7-8901-bcde-f12345678901',
    'INV-TP-2026-118',
    'PO-2026-101',
    135000.00, 24300.00, 'INR',
    '2026-03-08', '2026-04-07',
    'FLAGGED',
    '[{"description": "Server Rack Components", "quantity": 10, "unit_price": 9200.00, "total": 92000.00}, {"description": "Network Cables (Cat6)", "quantity": 200, "unit_price": 215.00, "total": 43000.00}]',
    'Amount mismatch: Invoice total ₹135,000 exceeds PO total ₹125,000 by 8%. Flagged for review.',
    NULL
),
(
    'cccc3333-3333-3333-3333-333333333333',
    'd4e5f6a7-b8c9-0123-defa-234567890123',
    'INV-SN-2026-007',
    'PO-2026-103',
    200000.00, 36000.00, 'INR',
    '2026-03-05', '2026-04-04',
    'SCHEDULED',
    '[{"description": "Annual Security Audit", "quantity": 1, "unit_price": 150000.00, "total": 150000.00}, {"description": "Penetration Testing", "quantity": 1, "unit_price": 50000.00, "total": 50000.00}]',
    'All checks passed. Early payment discount available: 2.5% if paid within 10 days (ROI: 30.4% annualised). Scheduled for early payment.',
    '2026-03-15'
)
ON CONFLICT DO NOTHING;

-- ── Audit Logs for sample invoices ──────────
INSERT INTO audit_logs (invoice_id, step, result, reasoning, confidence_score, duration_ms) VALUES
('aaaa1111-1111-1111-1111-111111111111', 'EXTRACT',         'completed', 'Extracted 7 fields with 95% average confidence', 0.9500, 1250),
('aaaa1111-1111-1111-1111-111111111111', 'VALIDATE',        'completed', 'All required fields present, amounts valid, dates in range', 1.0000, 45),
('aaaa1111-1111-1111-1111-111111111111', 'CROSS_REFERENCE', 'completed', 'PO-2026-100 matched. Amount: ₹50,000 = PO value. Line items: 2/2 matched.', 1.0000, 120),
('aaaa1111-1111-1111-1111-111111111111', 'DECISION',        'completed', 'Auto-approved: All checks passed.', 1.0000, 30),
('aaaa1111-1111-1111-1111-111111111111', 'SCHEDULE',        'completed', 'Payment scheduled for Day 28 (2026-03-29). No early discount benefit.', 1.0000, 15),

('bbbb2222-2222-2222-2222-222222222222', 'EXTRACT',         'completed', 'Extracted 7 fields with 92% average confidence', 0.9200, 1380),
('bbbb2222-2222-2222-2222-222222222222', 'VALIDATE',        'completed', 'All required fields present, amounts valid', 1.0000, 40),
('bbbb2222-2222-2222-2222-222222222222', 'CROSS_REFERENCE', 'completed', 'PO-2026-101 found. AMOUNT_MISMATCH: Invoice ₹135,000 > PO ₹125,000 (+8%)', 0.8500, 135),
('bbbb2222-2222-2222-2222-222222222222', 'DECISION',        'completed', 'Flagged: Amount exceeds PO by >5%. Requires human review.', 0.9000, 25),
('bbbb2222-2222-2222-2222-222222222222', 'DRAFT_QUERY',     'completed', 'Query email drafted and sent to accounts@techparts.in', 0.9500, 2100),

('cccc3333-3333-3333-3333-333333333333', 'EXTRACT',         'completed', 'Extracted 7 fields with 97% average confidence', 0.9700, 1100),
('cccc3333-3333-3333-3333-333333333333', 'VALIDATE',        'completed', 'All required fields present, amounts valid', 1.0000, 38),
('cccc3333-3333-3333-3333-333333333333', 'CROSS_REFERENCE', 'completed', 'PO-2026-103 matched. Amount: ₹200,000 = PO value. Line items: 2/2 matched.', 1.0000, 110),
('cccc3333-3333-3333-3333-333333333333', 'DECISION',        'completed', 'Auto-approved: All checks passed.', 1.0000, 28),
('cccc3333-3333-3333-3333-333333333333', 'SCHEDULE',        'completed', 'Early payment: 2.5% discount within 10 days = ₹5,000 saved. ROI 30.4% annualised. Scheduled for 2026-03-15.', 1.0000, 22)
ON CONFLICT DO NOTHING;
