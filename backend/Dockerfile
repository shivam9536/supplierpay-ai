# ── Build Stage ──────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/server ./cmd/server

# ── Dev Stage ─────────────────────────────────
# Runs the server; code is mounted from host. Restart container to pick up changes.
FROM golang:1.24-alpine AS dev

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["go", "run", "./cmd/server"]

# ── Production Stage ─────────────────────────
FROM alpine:3.19 AS production

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/bin/server .
COPY --from=builder /app/db ./db

EXPOSE 8080

CMD ["./server"]
