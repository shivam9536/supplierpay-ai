#!/bin/bash
set -e

API="http://localhost:8080/api/v1"

echo "=== Step 1: Seed POs ==="
docker compose exec -T postgres psql -U supplierpay -d supplierpay << 'EOSQL'
INSERT INTO purchase_orders (id, po_number, vendor_id, total_value, remaining_value, line_items, approved_by, status) VALUES
('11111111-1111-1111-1111-111111111111','PO-2026-100','a1b2c3d4-e5f6-7890-abcd-ef1234567890',50000.00,50000.00,'[{"description":"Cloud Hosting - Monthly","quantity":1,"unit_price":41000.00,"total":41000.00},{"description":"Support Services","quantity":1,"unit_price":9000.00,"total":9000.00}]','Rahul Sharma','OPEN'),
('22222222-2222-2222-2222-222222222222','PO-2026-101','b2c3d4e5-f6a7-8901-bcde-f12345678901',125000.00,125000.00,'[{"description":"Server Rack Components","quantity":10,"unit_price":8500.00,"total":85000.00},{"description":"Network Cables (Cat6)","quantity":200,"unit_price":200.00,"total":40000.00}]','Priya Patel','OPEN'),
('33333333-3333-3333-3333-333333333333','PO-2026-102','c3d4e5f6-a7b8-9012-cdef-123456789012',35000.00,35000.00,'[{"description":"Office Chairs (Ergonomic)","quantity":10,"unit_price":2500.00,"total":25000.00},{"description":"Standing Desks","quantity":5,"unit_price":2000.00,"total":10000.00}]','Amit Kumar','OPEN'),
('44444444-4444-4444-4444-444444444444','PO-2026-103','d4e5f6a7-b8c9-0123-defa-234567890123',200000.00,200000.00,'[{"description":"Annual Security Audit","quantity":1,"unit_price":150000.00,"total":150000.00},{"description":"Penetration Testing","quantity":1,"unit_price":50000.00,"total":50000.00}]','Sneha Reddy','OPEN'),
('55555555-5555-5555-5555-555555555555','PO-2026-104','e5f6a7b8-c9d0-1234-efab-345678901234',75000.00,75000.00,'[{"description":"Data Pipeline Setup","quantity":1,"unit_price":50000.00,"total":50000.00},{"description":"Dashboard Development","quantity":1,"unit_price":25000.00,"total":25000.00}]','Rahul Sharma','OPEN'),
('66666666-6666-6666-6666-666666666666','PO-2026-105','a1b2c3d4-e5f6-7890-abcd-ef1234567890',80000.00,80000.00,'[{"description":"Cloud Migration Services","quantity":1,"unit_price":80000.00,"total":80000.00}]','Priya Patel','OPEN'),
('77777777-7777-7777-7777-777777777777','PO-2026-106','b2c3d4e5-f6a7-8901-bcde-f12345678901',45000.00,45000.00,'[{"description":"Laptop Batteries","quantity":50,"unit_price":900.00,"total":45000.00}]','Amit Kumar','OPEN')
ON CONFLICT DO NOTHING;

SELECT po_number, status, remaining_value FROM purchase_orders ORDER BY po_number;
EOSQL

echo ""
echo "=== Step 2: Get auth token ==="
TOKEN=$(curl -s -X POST "$API/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@supplierpay.ai","password":"demo"}' | python3 -c "import sys,json; print(json.load(sys.stdin).get('token',''))")

if [ -z "$TOKEN" ]; then
  echo "Login failed, trying fallback credentials..."
  TOKEN=$(curl -s -X POST "$API/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"test@test.com","password":"test"}' | python3 -c "import sys,json; print(json.load(sys.stdin).get('token',''))")
fi
echo "Token: ${TOKEN:0:40}..."

echo ""
echo "=== Step 3: Upload sample PDFs ==="
SAMPLES_DIR="$(dirname "$0")/../samples/pdf"

for pdf in "$SAMPLES_DIR"/INV_*.pdf; do
  filename=$(basename "$pdf")
  echo -n "Uploading $filename ... "
  result=$(curl -s -X POST "$API/invoices/upload" \
    -H "Authorization: Bearer $TOKEN" \
    -F "invoice=@$pdf")
  invoice_id=$(echo "$result" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('invoice_id','ERROR'))" 2>/dev/null)
  echo "→ invoice_id: $invoice_id"
done

echo ""
echo "=== Done! Uploaded all sample invoices ==="
echo "Visit http://localhost:3000/invoices to see them processing"
