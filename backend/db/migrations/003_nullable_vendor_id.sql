-- Allow invoices to be uploaded without a known vendor_id.
-- The AI extraction pipeline will resolve the vendor from the PDF content.
ALTER TABLE invoices ALTER COLUMN vendor_id DROP NOT NULL;
