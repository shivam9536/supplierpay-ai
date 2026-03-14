-- ============================================
-- SupplierPay AI — Seed Data for Demo
-- Matches dataset/invoices and dataset/purchase_orders
-- ============================================

-- ── Vendors ─────────────────────────────────
-- Vendors matching the invoice/PO dataset files
INSERT INTO vendors (id, name, email, bank_account_number, bank_ifsc, payment_terms_days, early_payment_discount, early_payment_days, criticality_score) VALUES
-- INV-ACME-2026-050.pdf / PO-2026-100.pdf
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'Acme Cloud Solutions',       'billing@acmecloud.com',       '1234567890', 'HDFC0001234', 30, 2.00, 10, 9),
-- INV_01_corporate.pdf / PO_01_corporate.pdf
('b1000001-0001-0001-0001-000000000001', 'Northern Electronics Ltd',   'accounts@northern-elec.com',  '1111111111', 'HDFC0001111', 30, 1.50, 10, 8),
-- INV_02_trading.pdf / PO_02_trading.pdf
('b2000002-0002-0002-0002-000000000002', 'Global Trading Co',          'billing@globaltrading.com',   '2222222222', 'ICIC0002222', 45, 2.00, 15, 7),
-- INV_03_office_supplies.pdf / PO_03_office_supplies.pdf
('b3000003-0003-0003-0003-000000000003', 'Office Essentials Pvt Ltd',  'finance@officeessentials.in', '3333333333', 'SBIN0003333', 30, 0.00,  0, 5),
-- INV_04_consulting.pdf / PO_04_consulting.pdf
('b4000004-0004-0004-0004-000000000004', 'Strategic Consulting Group', 'invoices@strategiccg.com',    '4444444444', 'AXIS0004444', 60, 3.00, 20, 9),
-- INV_05_manufacturing.pdf / PO_05_manufacturing.pdf
('b5000005-0005-0005-0005-000000000005', 'Precision Manufacturing Co', 'ap@precisionmfg.in',          '5555555555', 'KOTK0005555', 30, 1.00, 10, 8),
-- INV_06_small_business.pdf / PO_06_small_business.pdf
('b6000006-0006-0006-0006-000000000006', 'Quick Services Ltd',         'billing@quickservices.com',   '6666666666', 'HDFC0006666', 15, 0.00,  0, 4)
ON CONFLICT DO NOTHING;

-- ── Purchase Orders ─────────────────────────
-- POs matching the dataset/purchase_orders files
INSERT INTO purchase_orders (id, po_number, vendor_id, total_value, remaining_value, line_items, approved_by, status) VALUES
-- PO-2026-100.pdf → Acme Cloud Solutions
(
    '11111111-1111-1111-1111-111111111111',
    'PO-2026-100',
    'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
    59000.00,
    59000.00,
    '[{"description": "Cloud Hosting - March", "quantity": 1, "unit_price": 41000.00, "total": 41000.00}, {"description": "Support Services", "quantity": 1, "unit_price": 9000.00, "total": 9000.00}, {"description": "Tax (GST 18%)", "quantity": 1, "unit_price": 9000.00, "total": 9000.00}]',
    'Rahul Sharma',
    'OPEN'
),
-- PO_01_corporate.pdf → Northern Electronics Ltd
(
    '01000001-0001-0001-0001-000000000001',
    'PO-CORP-2026-001',
    'b1000001-0001-0001-0001-000000000001',
    250000.00,
    250000.00,
    '[{"description": "Enterprise Server Equipment", "quantity": 2, "unit_price": 100000.00, "total": 200000.00}, {"description": "Installation & Setup", "quantity": 1, "unit_price": 50000.00, "total": 50000.00}]',
    'Priya Patel',
    'OPEN'
),
-- PO_02_trading.pdf → Global Trading Co
(
    '02000002-0002-0002-0002-000000000002',
    'PO-TRADE-2026-002',
    'b2000002-0002-0002-0002-000000000002',
    175000.00,
    175000.00,
    '[{"description": "Imported Electronics Components", "quantity": 500, "unit_price": 300.00, "total": 150000.00}, {"description": "Shipping & Handling", "quantity": 1, "unit_price": 25000.00, "total": 25000.00}]',
    'Amit Kumar',
    'OPEN'
),
-- PO_03_office_supplies.pdf → Office Essentials Pvt Ltd
(
    '03000003-0003-0003-0003-000000000003',
    'PO-OFFICE-2026-003',
    'b3000003-0003-0003-0003-000000000003',
    45000.00,
    45000.00,
    '[{"description": "Ergonomic Office Chairs", "quantity": 10, "unit_price": 3500.00, "total": 35000.00}, {"description": "Standing Desks", "quantity": 2, "unit_price": 5000.00, "total": 10000.00}]',
    'Sneha Reddy',
    'OPEN'
),
-- PO_04_consulting.pdf → Strategic Consulting Group
(
    '04000004-0004-0004-0004-000000000004',
    'PO-CONSULT-2026-004',
    'b4000004-0004-0004-0004-000000000004',
    500000.00,
    500000.00,
    '[{"description": "Business Strategy Consulting", "quantity": 1, "unit_price": 350000.00, "total": 350000.00}, {"description": "Market Research & Analysis", "quantity": 1, "unit_price": 150000.00, "total": 150000.00}]',
    'Rahul Sharma',
    'OPEN'
),
-- PO_05_manufacturing.pdf → Precision Manufacturing Co
(
    '05000005-0005-0005-0005-000000000005',
    'PO-MFG-2026-005',
    'b5000005-0005-0005-0005-000000000005',
    320000.00,
    320000.00,
    '[{"description": "CNC Machined Parts", "quantity": 100, "unit_price": 2500.00, "total": 250000.00}, {"description": "Quality Inspection", "quantity": 1, "unit_price": 30000.00, "total": 30000.00}, {"description": "Packaging & Delivery", "quantity": 1, "unit_price": 40000.00, "total": 40000.00}]',
    'Priya Patel',
    'OPEN'
),
-- PO_06_small_business.pdf → Quick Services Ltd
(
    '06000006-0006-0006-0006-000000000006',
    'PO-SB-2026-006',
    'b6000006-0006-0006-0006-000000000006',
    25000.00,
    25000.00,
    '[{"description": "Cleaning Services - Monthly", "quantity": 1, "unit_price": 15000.00, "total": 15000.00}, {"description": "Maintenance Support", "quantity": 1, "unit_price": 10000.00, "total": 10000.00}]',
    'Amit Kumar',
    'OPEN'
)
ON CONFLICT DO NOTHING;

-- ── Goods Receipts (for 3-way matching) ─────
-- GRs matching the POs above - enables 3-way matching during invoice validation
INSERT INTO goods_receipts (po_number, received_quantity, received_date, status) VALUES
('PO-2026-100',       '[{"description": "Cloud Hosting - March", "quantity": 1}, {"description": "Support Services", "quantity": 1}]', '2026-03-05', 'RECEIVED'),
('PO-CORP-2026-001',  '[{"description": "Enterprise Server Equipment", "quantity": 2}, {"description": "Installation & Setup", "quantity": 1}]', '2026-03-08', 'RECEIVED'),
('PO-TRADE-2026-002', '[{"description": "Imported Electronics Components", "quantity": 500}, {"description": "Shipping & Handling", "quantity": 1}]', '2026-03-10', 'RECEIVED'),
('PO-OFFICE-2026-003','[{"description": "Ergonomic Office Chairs", "quantity": 10}, {"description": "Standing Desks", "quantity": 2}]', '2026-03-12', 'RECEIVED'),
('PO-CONSULT-2026-004','[{"description": "Business Strategy Consulting", "quantity": 1}, {"description": "Market Research & Analysis", "quantity": 1}]', '2026-03-01', 'RECEIVED'),
('PO-MFG-2026-005',   '[{"description": "CNC Machined Parts", "quantity": 100}, {"description": "Quality Inspection", "quantity": 1}, {"description": "Packaging & Delivery", "quantity": 1}]', '2026-03-11', 'RECEIVED'),
('PO-SB-2026-006',    '[{"description": "Cleaning Services - Monthly", "quantity": 1}, {"description": "Maintenance Support", "quantity": 1}]', '2026-03-01', 'RECEIVED')
ON CONFLICT DO NOTHING;

-- ══════════════════════════════════════════════════════════════
-- NOTE: No sample invoices are seeded.
-- Invoices will be added via the upload workflow to test AI processing.
-- ══════════════════════════════════════════════════════════════
