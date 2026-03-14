# 🤖 SupplierPay AI — Autonomous B2B Invoice & Payment Agent

> Upload invoices (PDF or JSON). The agent extracts, validates, approves, and schedules payments — end to end.

## 🏗️ Architecture

```
Frontend (React + Vite)  →  Backend (Go + Gin)  →  AWS Bedrock (Claude)
         ↕                        ↕                      ↕
    Tailwind + Recharts       Postgres DB         Pine Labs / Plural
         ↕                        ↕                      ↕
    SSE real-time updates    Agent Orchestrator      Payments
                             (Extract → Validate → Decide → Pay)
```

## 📁 Project Structure

```
supplierpay-ai/
├── docker-compose.yml          # Full stack (Postgres, backend, frontend)
├── .env.example                # Environment template — copy to .env
├── Makefile                     # Dev shortcuts (up, down, db-reset, etc.)
│
├── backend/                     # Go + Gin API
│   ├── cmd/server/main.go      # Entry point
│   ├── internal/
│   │   ├── config/             # Env config (DB, AWS, Pine Labs, JWT)
│   │   ├── database/           # Postgres connection
│   │   ├── models/             # Invoice, PO, Vendor, Validation, etc.
│   │   ├── handlers/           # HTTP handlers (invoices, vendors, POs, payments)
│   │   ├── middleware/         # JWT auth
│   │   ├── router/             # Routes
│   │   ├── agent/              # 🤖 Orchestrator: extract → validate → decision → schedule
│   │   ├── services/           # Bedrock, S3, SES, Pine Labs (+ mocks)
│   │   └── scheduler/          # Cron payment runs
│   ├── db/
│   │   ├── migrations/         # 001_init.sql, 002_invoice_validations, 003_nullable_vendor_id
│   │   └── seed/               # 001_seed.sql — vendors, POs, demo users
│   └── Dockerfile
│
├── frontend/                    # React 18 + Vite + Tailwind
│   ├── src/
│   │   ├── pages/               # Dashboard, Invoices, Upload, InvoiceDetail, Vendors, POs, etc.
│   │   ├── components/          # Layout, shared UI
│   │   ├── context/             # Auth
│   │   └── services/            # API client (axios)
│   └── Dockerfile
│
├── scripts/
│   └── seed_samples.sh         # Login + upload sample PDFs from samples/pdf (INV_*.pdf)
│
└── samples/
    └── pdf/                     # Sample invoice PDFs for demo
```

## 🚀 Quick Start

### Prerequisites

- **Docker & Docker Compose**
- **Go 1.22+** (optional, for local backend)
- **Node.js 18+** or **Bun** (optional, for local frontend)

### 1. Clone & env

```bash
git clone <repo-url>
cd supplierpay-ai
cp .env.example .env
# Edit .env: set MOCK_MODE=true for no AWS/Pine Labs; set credentials for real services
```

### 2. Start stack

```bash
make up
```

- **Frontend:** http://localhost:3000  
- **Backend API:** http://localhost:8080  
- **Postgres:** localhost:5432  

### 3. Login

Open http://localhost:3000 and sign in (e.g. `demo@supplierpay.ai` / `demo` or credentials from seed).

### 4. Optional: upload sample invoices

With the stack running and after at least one login has been done (so JWT works), you can upload sample PDFs:

```bash
./scripts/seed_samples.sh
```

This seeds POs, gets a token, and uploads all `samples/pdf/INV_*.pdf` via the API. Then open http://localhost:3000/invoices to see them in the pipeline.

### 5. Useful commands

```bash
make up              # Start all services
make down            # Stop all
make logs            # Tail all logs
make logs-backend    # Backend only
make db-shell        # Postgres psql
make db-reset        # Re-run init + seed (see Makefile for exact steps)
make clean           # Remove containers, volumes, local build artifacts
```

## 🔧 Local development (without Docker)

### Backend

```bash
cd backend
go mod tidy
go run cmd/server/main.go
```

Set `DB_HOST=localhost` (and ensure Postgres is running, e.g. only `docker compose up postgres -d`).

### Frontend

```bash
cd frontend
npm install   # or: bun install
npm run dev   # or: bun run dev
```

Set `VITE_API_URL=http://localhost:8080` if needed.

### Database (Docker only)

```bash
docker compose up postgres -d
# Schema + seed are applied via mounted init scripts (01_init.sql, 02_seed.sql)
# To re-run: make db-reset (or run the SQL files manually)
```

## 📡 API overview

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/auth/login` | JWT login |
| `POST` | `/api/v1/invoices/upload` | Upload PDF (multipart; optional `vendor_id`, `po_reference`) |
| `POST` | `/api/v1/invoices/upload-json` | Submit invoice as JSON (for demos; requires `vendor_id`) |
| `GET`  | `/api/v1/invoices` | List invoices (optional `status`, `vendor_id`) |
| `GET`  | `/api/v1/invoices/:id` | Invoice detail |
| `GET`  | `/api/v1/invoices/:id/audit-log` | Agent audit trail |
| `POST` | `/api/v1/invoices/:id/reprocess` | Re-run pipeline |
| `GET`  | `/api/v1/vendors` | List vendors |
| `GET`  | `/api/v1/purchase-orders` | List POs |
| `GET`  | `/api/v1/payments/schedule` | Upcoming payments |
| `POST` | `/api/v1/payments/run` | Trigger payment run |
| `GET`  | `/api/v1/forecast` | Cash flow forecast |
| `GET`  | `/api/v1/events/invoices/:id` | SSE updates for an invoice |
| `POST` | `/api/v1/webhooks/pinelabs` | Pine Labs webhook |

All invoice/payment/vendor/PO endpoints require `Authorization: Bearer <token>` except login and webhook.

## 🤖 Agent pipeline

1. **Upload** — PDF stored in S3 (or mock); invoice row created with status `PENDING`.
2. **Extract** — Fields from existing JSON, or from PDF via Bedrock (or mock). Vendor resolved from `vendor_name` and written to invoice.
3. **Validate** — Runs all checks and writes result to `invoice_validations`:
   - Required fields, positive total
   - **Vendor exists** (in `vendors`)
   - **PO exists** and is **open** (or partially matched)
   - **Vendor matches PO**
   - **No duplicate** invoice number
   - **Amount within PO** (invoice total ≤ PO remaining, 2% tolerance)
   - **Line items** match PO (description, quantity, unit price within 2%)
4. **Decision** — From validation: **APPROVE** → schedule payment; **FLAG** → draft query email; **REJECT** → stop.
5. **Schedule / Disburse** — Payment terms, schedule date, then Pine Labs (or mock) for disbursement.

```
UPLOAD → EXTRACT → RESOLVE_VENDOR → VALIDATE → DECISION → SCHEDULE → DISBURSE
                                              │
                              ┌───────────────┼───────────────┐
                              ▼               ▼               ▼
                         APPROVE          FLAG            REJECT
                              │               │               │
                      SCHEDULE_PAYMENT   DRAFT_QUERY     (stop)
                              │           EMAIL
                         DISBURSE (Pine Labs)
```

## ⚙️ Environment

Key variables (see `.env.example`):

| Group | Variables |
|-------|-----------|
| **App** | `APP_ENV`, `APP_PORT`, `FRONTEND_URL`, `MOCK_MODE` |
| **DB** | `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE` |
| **AWS** | `AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` |
| **Bedrock** | `BEDROCK_API_KEY` or IAM; `BEDROCK_MODEL_ID`, `BEDROCK_MAX_TOKENS` |
| **S3** | `S3_BUCKET_NAME`, `S3_REGION` |
| **SES** | `SES_SENDER_EMAIL`, `SES_REGION` |
| **Pine Labs / Plural** | `PINELABS_API_URL`, `PINELABS_CLIENT_ID`, `PINELABS_CLIENT_SECRET`, `PINELABS_MERCHANT_ID` |
| **JWT** | `JWT_SECRET`, `JWT_EXPIRY_HOURS` |

## 🧪 Mock mode

Set **`MOCK_MODE=true`** in `.env` to use in-memory mocks:

- **Bedrock** — No API calls; returns fixed extraction.
- **S3** — No upload; returns a fake URL.
- **SES** — No email sent; logs only.
- **Pine Labs** — No real payments; returns mock transaction IDs.

Use this when you don’t have AWS or Pine Labs credentials.

## 📊 Seed data

- **Vendors** — 5 demo vendors (Acme Cloud, TechParts, Global Office Supplies, etc.).
- **Purchase orders** — 7 POs (OPEN / PARTIALLY_MATCHED) with line items.
- **Users** — Demo login (e.g. `demo@supplierpay.ai` / `demo`).
- **Sample PDFs** — In `samples/pdf/`; upload via UI or `scripts/seed_samples.sh`.

Invoice numbers in seed are unique; for JSON upload use a new `invoice_number` (e.g. `INV-2026-PASS-001`) so the duplicate check passes.

## 🧩 Main code areas

| Area | Focus | Key paths |
|------|--------|-----------|
| **Infra & payments** | S3, SES, Pine Labs, scheduler | `services/s3.go`, `ses.go`, `pinelabs.go`, `scheduler/` |
| **Agent** | Orchestrator, validation, decision | `agent/orchestrator.go`, `services/bedrock.go` |
| **Backend API** | Handlers, DB, forecast | `handlers/*.go`, `db/` |
| **Frontend** | Pages, API client, SSE | `frontend/src/pages/`, `services/api.js` |

---

Built for **Pine Labs Hackathon** 🚀
