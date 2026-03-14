# SupplierPay AI — Implementation Summary & How to Run

## What Was Done

The project was completed so the full stack runs end-to-end with mock services. Summary of changes:

### 1. **Backend wiring**
- **Main** (`cmd/server/main.go`): Instantiates mock or real LLM, storage, and email services based on `MOCK_MODE`. Creates the agent **Orchestrator** and an **Event Broadcaster** that fans out SSE events by invoice ID. Passes orchestrator, broadcaster, and storage into the router.
- **Router** (`internal/router/router.go`): Accepts orchestrator, broadcaster, and storage; passes them into the invoice and SSE handlers.

### 2. **Agent orchestrator** (`internal/agent/orchestrator.go`)
- **Extract**: Reads invoice from DB (uses existing `extracted_fields` from JSON upload or mock extraction). Persists extracted fields and writes an audit log entry.
- **Validate**: Keeps current rules; on failure sets invoice status to REJECTED and writes audit log.
- **Cross-reference**: Loads PO by `po_reference` from DB, checks amount vs PO total, duplicate invoice number; sets discrepancies and writes audit log.
- **Decision**: Updates invoice status (APPROVED/FLAGGED/REJECTED) and `decision_reason`; writes audit log.
- **Schedule payment**: For APPROVE, computes payment date from vendor terms, updates `scheduled_payment_date` and status to SCHEDULED; writes audit log.
- **Draft query**: For FLAG, writes audit log.
- Helpers added: `updateInvoiceStatus`, `insertAuditLog`, `persistExtractedFields`, `persistDiscrepancies`, `decisionActionToStatus`.

### 3. **Handlers (real DB)**
- **Vendor**: List, GetByID, Create with Postgres.
- **Purchase order**: List (optional filters), GetByID, Create with Postgres.
- **Invoice**: List (with vendor name join), GetByID, GetAuditLog; **Upload** (store file via storage, insert invoice, trigger orchestrator in background); **UploadJSON** (insert from JSON, trigger orchestrator); **Reprocess** (reset status, trigger orchestrator).
- **Payment**: GetSchedule (invoices with status SCHEDULED and `scheduled_payment_date >= today`), ListRuns (from `payment_runs` table).
- **Forecast**: GetForecast (90-day outlook from APPROVED/SCHEDULED invoices, grouped by week, with risk flags).

### 4. **SSE (Server-Sent Events)**
- **Event broadcaster** (`internal/events/broadcaster.go`): Subscribes by invoice ID; forwards events from the orchestrator’s channel to subscribed clients.
- **SSE handler**: Subscribes to the broadcaster for the requested invoice ID and streams JSON events; sends a heartbeat every 15s.

### 5. **Frontend**
- **Invoices page**: Uses `getInvoices()` API, shows loading/error, formats dates; displays vendor name from list response.

### 6. **Other**
- **Makefile** `db-reset`: Fixed paths to init/seed scripts (`01_init.sql`, `02_seed.sql`).
- **Models**: Added `InvoiceWithVendor` for list response with `vendor_name`.

---

## How to Run

### Option A: Full stack with Docker

1. **Env**
   ```bash
   cp .env.example .env
   ```
   Keep `MOCK_MODE=true` so Bedrock/S3/SES/Pine Labs are mocked.

2. **Start**
   ```bash
   make up
   ```
   - Frontend: http://localhost:3000  
   - Backend: http://localhost:8080  
   - Postgres: localhost:5432 (user `supplierpay`, password `supplierpay_dev`, db `supplierpay`)

3. **Login**
   - Open http://localhost:3000 and sign in with any email/password (hackathon mode).

4. **Try**
   - **Invoices**: List shows seed invoices (and any new ones).
   - **Upload**: Use “Upload Invoice” with a file and `vendor_id` (e.g. `a1b2c3d4-e5f6-7890-abcd-ef1234567890` from seed); then open the invoice detail and watch SSE for pipeline steps.
   - **Upload JSON**: `POST /api/v1/invoices/upload-json` with JSON body including `vendor_id`, optional `invoice_number`, `po_reference`, `total_amount`, `line_items`, etc.; pipeline runs in background.
   - **Payment schedule**: Invoices with status SCHEDULED and due today or later.
   - **Cash flow**: 90-day forecast from APPROVED/SCHEDULED invoices.

### Option B: Local backend + Docker Postgres

1. Start Postgres:
   ```bash
   docker compose up postgres -d
   ```
   Wait a few seconds for init (migrations + seed).

2. Backend:
   ```bash
   cd backend
   export MOCK_MODE=true
   export DB_HOST=localhost
   # optional: GOPROXY=https://proxy.golang.org,direct if your env uses a custom proxy
   go run ./cmd/server
   ```

3. Frontend:
   ```bash
   cd frontend
   bun install && bun run dev
   ```
   Ensure `VITE_API_URL=http://localhost:8080` (or set in `.env`).

### Database reset (Docker)

```bash
make db-reset
```

Then start the rest of the stack again (`make up` or backend + frontend as above).

---

## API Quick Reference

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/login` | Body: `{ "email", "password" }` → JWT |
| GET | `/api/v1/invoices` | List (optional query: `status`, `vendor_id`) |
| GET | `/api/v1/invoices/:id` | Invoice detail |
| GET | `/api/v1/invoices/:id/audit-log` | Agent audit trail |
| POST | `/api/v1/invoices/upload` | Form: `invoice` (file), `vendor_id`, optional `po_reference` |
| POST | `/api/v1/invoices/upload-json` | JSON body with `vendor_id` + invoice fields |
| POST | `/api/v1/invoices/:id/reprocess` | Re-run agent pipeline |
| GET | `/api/v1/events/invoices/:id` | SSE stream for pipeline updates |
| GET | `/api/v1/vendors` | List vendors |
| GET | `/api/v1/purchase-orders` | List POs |
| GET | `/api/v1/payments/schedule` | Upcoming scheduled payments |
| GET | `/api/v1/forecast` | 90-day cash flow forecast |

All except login and Pine Labs webhook require header: `Authorization: Bearer <token>`.

---

## Optional Next Enhancements

- **Real AWS Bedrock**: Set `MOCK_MODE=false` and configure AWS credentials; implement `ExtractInvoiceFields` in `services/bedrock.go`.
- **Real S3/SES**: Implement upload/send in `services/s3.go` and `services/ses.go`.
- **Pine Labs**: Implement `runPayments` in `scheduler/scheduler.go` and webhook signature validation in the payment handler.
- **Frontend**: Wire Dashboard, InvoiceDetail (with SSE), UploadInvoice (form + redirect to detail), PaymentSchedule, CashFlow, Vendors, and PurchaseOrders to the same APIs where they are still mock or TODO.
