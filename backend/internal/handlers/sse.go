package handlers

import (
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/supplierpay/backend/internal/events"
	"go.uber.org/zap"
)

type SSEHandler struct {
	logger      *zap.Logger
	broadcaster *events.Broadcaster
}

func NewSSEHandler(logger *zap.Logger, broadcaster *events.Broadcaster) *SSEHandler {
	return &SSEHandler{logger: logger, broadcaster: broadcaster}
}

// StreamInvoiceUpdates streams real-time agent pipeline updates for an invoice
func (h *SSEHandler) StreamInvoiceUpdates(c *gin.Context) {
	invoiceID := c.Param("id")
	h.logger.Info("SSE connection opened", zap.String("invoice_id", invoiceID))

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	ch := h.broadcaster.Subscribe(invoiceID)
	defer h.broadcaster.Unsubscribe(invoiceID, ch)

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Request.Context().Done():
			h.logger.Info("SSE connection closed", zap.String("invoice_id", invoiceID))
			return false
		case evt := <-ch:
			data, _ := json.Marshal(evt)
			c.SSEvent("", string(data))
			c.Writer.Flush()
			return true
		case <-heartbeat.C:
			c.SSEvent("heartbeat", "{}")
			c.Writer.Flush()
			return true
		}
	})
}
