package handlers

import (
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SSEHandler struct {
	logger *zap.Logger
}

func NewSSEHandler(logger *zap.Logger) *SSEHandler {
	return &SSEHandler{logger: logger}
}

// StreamInvoiceUpdates streams real-time agent pipeline updates for an invoice
func (h *SSEHandler) StreamInvoiceUpdates(c *gin.Context) {
	invoiceID := c.Param("id")
	h.logger.Info("SSE connection opened", zap.String("invoice_id", invoiceID))

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// TODO: Dev 2 — Implement real SSE channel per invoice
	// For now, this is a placeholder that keeps the connection open
	// The agent orchestrator will push events to a channel
	// that gets consumed here and sent to the client

	c.Stream(func(w io.Writer) bool {
		// This will be replaced with channel-based event delivery
		// Example event format:
		// data: {"invoice_id":"xxx","step":"EXTRACT","status":"completed","message":"Extracted 5 fields"}
		select {
		case <-c.Request.Context().Done():
			h.logger.Info("SSE connection closed", zap.String("invoice_id", invoiceID))
			return false
		default:
			// Heartbeat to keep connection alive
			fmt.Fprintf(w, "event: heartbeat\ndata: {}\n\n")
			c.Writer.Flush()
			return true
		}
	})
}
