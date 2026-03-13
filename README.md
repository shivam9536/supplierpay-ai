# 🤖 SupplierPay AI — Autonomous B2B Invoice & Payment Agent

> Upload invoices. The agent validates, approves, and schedules payments — no human needed.

## 🏗️ Architecture

```
Frontend (Bun + React)  →  Backend (Go + Gin)  →  AWS Bedrock (Claude)
         ↕                        ↕                      ↕
    Recharts UI              Postgres DB           Pine Labs Payments
                                 ↕
                          Agent Orchestrator
                    (Extract → Validate → Decide → Pay)
```

## 📁 Project Structure

```
Project/
├── docker-compose.yml          # Full stack orchestration
├── .env.example                # Environment variables template
├── Makefile                    # Development shortcuts
│
├── backend/                    # Go + Gin API server
│   ├── cmd/server/main.go      # Entry point
│   ├── internal/
│   │   ├── config/             # Environment config loader
│   │   ├── database/           # Postgres connection
│   │   ├── models/             # Data models (Invoice, PO, Vendor, etc.)
│   │   ├── handlers/           # HTTP route handlers
│   │   ├── middleware/         # JWT auth middleware
│   │   ├── router/             # Route definitions
│   │   ├── agent/              # 🤖 AI Agent orchestrator (FSM)
│   │   ├── services/           # External service clients
│   │   │   ├── interfaces.go   # Service interfaces
│   │   │   ├── bedrock.go      # AWS Bedrock LLM + mock
│   │   │   ├── s3.go           # AWS S3 storage + mock
│   │   │   ├── ses.go          # AWS SES email + mock
│   │   │   └── pinelabs.go     # Pine Labs payments + mock
│   │   └── scheduler/          # Cron payment scheduler
│   ├── db/
│   │   ├── migrations/         # SQL schema
│   │   └── seed/               # Demo data
│   └── Dockerfile
│
└── frontend/                   # React + Vite + Tailwind
    ├── src/
    │   ├── pages/              # Dashboard, Invoices, Upload, etc.
    │   ├── components/         # Layout, shared components
    │   ├── context/            # Auth context
    │   └── services/           # API client (axios)
    └── Dockerfile
```

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.22+ (for local backend dev)
- Bun (for local frontend dev)

### 1. Clone & Setup
```bash
git clone <repo-url>
cd Project
cp .env.example .env
```

### 2. Start Everything (Docker)
```bash
make up
```

This starts:
- **Frontend** → http://localhost:3000
- **Backend API** → http://localhost:8080
- **Postgres** → localhost:5432

### 3. Login
Open http://localhost:3000 and login with any email/password (hackathon mode).

### 4. Useful Commands
```bash
make up              # Start all services
make down            # Stop all services
make logs            # View all logs
make logs-backend    # View backend logs only
make db-shell        # Open Postgres CLI
make db-reset        # Reset & reseed database
make clean           # Nuclear cleanup
```

## 🔧 Local Development (Without Docker)

### Backend
```bash
cd backend
go mod tidy
go run cmd/server/main.go
```

### Frontend
```bash
cd frontend
bun install
bun run dev
```

### Database
```bash
# Start only Postgres in Docker
docker compose up postgres -d

# Run migrations
docker compose exec postgres psql -U supplierpay -d supplierpay -f /docker-entrypoint-initdb.d/migrations/001_init.sql

# Run seed data
docker compose exec postgres psql -U supplierpay -d supplierpay -f /docker-entrypoint-initdb.d/seed/001_seed.sql
```

## 📡 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/auth/login` | Get JWT token |
| `POST` | `/api/v1/invoices/upload` | Upload PDF invoice |
| `POST` | `/api/v1/invoices/upload-json` | Submit JSON invoice |
| `GET` | `/api/v1/invoices` | List all invoices |
| `GET` | `/api/v1/invoices/:id` | Get invoice detail |
| `GET` | `/api/v1/invoices/:id/audit-log` | Agent pipeline audit trail |
| `POST` | `/api/v1/invoices/:id/reprocess` | Re-run agent pipeline |
| `GET` | `/api/v1/vendors` | List vendors |
| `GET` | `/api/v1/purchase-orders` | List POs |
| `GET` | `/api/v1/payments/schedule` | Upcoming payments |
| `POST` | `/api/v1/payments/run` | Trigger payment run |
| `GET` | `/api/v1/forecast` | 90-day cash flow forecast |
| `GET` | `/api/v1/events/invoices/:id` | SSE real-time updates |
| `POST` | `/api/v1/webhooks/pinelabs` | Pine Labs payment webhook |

## 🤖 Agent Pipeline

```
UPLOAD → EXTRACT → VALIDATE → CROSS_REFERENCE → DECISION → ACTION
                                                    │
                                    ┌───────────────┼───────────────┐
                                    ▼               ▼               ▼
                               APPROVE          FLAG            REJECT
                                    │               │               │
                            SCHEDULE_PAYMENT  DRAFT_QUERY    NOTIFY_VENDOR
                                    │           EMAIL
                               DISBURSE
                            (Pine Labs API)
```

## 👥 Dev Assignment

| Dev | Focus Area | Key Files |
|-----|-----------|-----------|
| **Dev 1** | Infra + Payments | `services/s3.go`, `services/ses.go`, `services/pinelabs.go`, `scheduler/` |
| **Dev 2** | AI Agent Core | `agent/orchestrator.go`, `services/bedrock.go`, handlers logic |
| **Dev 3** | Backend APIs | `handlers/*.go`, DB queries, forecast engine |
| **Dev 4** | Frontend | `frontend/src/pages/*`, API integration, SSE |

## 🧪 Mock Mode

Set `MOCK_MODE=true` in `.env` to use mock implementations of:
- Bedrock (returns hardcoded extraction results)
- S3 (logs upload, returns fake URL)
- SES (logs email, doesn't send)
- Pine Labs (returns mock transaction IDs)

This lets all devs work without AWS/Pine Labs credentials.

## 📊 Demo Seed Data

The seed includes:
- **5 vendors** with varying payment terms and discount structures
- **7 purchase orders** (open, partially matched, closed)
- **3 sample invoices** (approved, flagged, scheduled)
- **Full audit logs** showing agent reasoning for each invoice
- **Goods receipts** for 3-way matching demo

---

Built for **Pine Labs Hackathon** 🚀
