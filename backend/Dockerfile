# ── Build Stage ──────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/server ./cmd/server

# ── Dev Stage (with hot reload) ──────────────
FROM golang:1.22-alpine AS dev

WORKDIR /app

RUN apk add --no-cache gcc musl-dev
RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["air", "-c", ".air.toml"]

# ── Production Stage ─────────────────────────
FROM alpine:3.19 AS production

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/bin/server .
COPY --from=builder /app/db ./db

EXPOSE 8080

CMD ["./server"]
