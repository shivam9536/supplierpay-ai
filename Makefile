.PHONY: help up down build logs restart db-reset seed backend frontend

# Default target
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ── Docker Commands ──────────────────────────
up: ## Start all services
	docker compose up --build -d

up-logs: ## Start all services with logs
	docker compose up --build

down: ## Stop all services
	docker compose down

restart: ## Restart all services
	docker compose down && docker compose up --build -d

logs: ## Tail all logs
	docker compose logs -f

logs-backend: ## Tail backend logs
	docker compose logs -f backend

logs-frontend: ## Tail frontend logs
	docker compose logs -f frontend

# ── Database Commands ────────────────────────
db-shell: ## Open psql shell
	docker compose exec postgres psql -U supplierpay -d supplierpay

	db-reset: ## Reset database (drop + recreate + seed)
	docker compose down -v
	docker compose up postgres -d
	@echo "Waiting for Postgres to be ready..."
	@sleep 5
	docker compose exec postgres psql -U supplierpay -d supplierpay -f /docker-entrypoint-initdb.d/01_init.sql
	docker compose exec postgres psql -U supplierpay -d supplierpay -f /docker-entrypoint-initdb.d/02_seed.sql
	@echo "Database reset complete!"

# ── Individual Service Commands ──────────────
backend: ## Start only backend
	docker compose up --build backend -d

frontend: ## Start only frontend
	docker compose up --build frontend -d

# ── Development Shortcuts ────────────────────
install-backend: ## Install Go dependencies
	cd backend && go mod tidy

install-frontend: ## Install frontend dependencies
	cd frontend && bun install

test-backend: ## Run backend tests
	cd backend && go test ./...

lint-backend: ## Lint backend code
	cd backend && golangci-lint run

# ── Clean Up ─────────────────────────────────
clean: ## Remove all containers, volumes, and build artifacts
	docker compose down -v --rmi local
	rm -rf backend/tmp backend/bin
	rm -rf frontend/dist frontend/node_modules
