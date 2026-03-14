package events

import (
	"sync"

	"github.com/supplierpay/backend/internal/models"
)

// Broadcaster fans out SSE events from the agent to clients subscribed by invoice ID.
type Broadcaster struct {
	mu      sync.RWMutex
	subs    map[string]map[chan models.SSEEvent]struct{} // invoiceID -> set of channels
	source  <-chan models.SSEEvent
	done    chan struct{}
	started bool
}

// NewBroadcaster creates a broadcaster that reads from the given source channel.
func NewBroadcaster(source <-chan models.SSEEvent) *Broadcaster {
	return &Broadcaster{
		subs:   make(map[string]map[chan models.SSEEvent]struct{}),
		source:  source,
		done:   make(chan struct{}),
	}
}

// Start begins forwarding events from source to subscribers. Call once.
func (b *Broadcaster) Start() {
	b.mu.Lock()
	if b.started {
		b.mu.Unlock()
		return
	}
	b.started = true
	b.mu.Unlock()

	go func() {
		for {
			select {
			case <-b.done:
				return
			case evt, ok := <-b.source:
				if !ok {
					return
				}
				b.mu.RLock()
				chans := b.subs[evt.InvoiceID]
				// Copy to avoid holding lock while sending
				var list []chan models.SSEEvent
				for ch := range chans {
					list = append(list, ch)
				}
				b.mu.RUnlock()
				for _, ch := range list {
					select {
					case ch <- evt:
					default:
						// Client slow; skip
					}
				}
			}
		}
	}()
}

// Stop stops the broadcaster.
func (b *Broadcaster) Stop() {
	close(b.done)
}

// Subscribe returns a channel that receives events for the given invoice ID.
// Caller must call Unsubscribe when done to avoid leaks.
func (b *Broadcaster) Subscribe(invoiceID string) chan models.SSEEvent {
	ch := make(chan models.SSEEvent, 10)
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.subs[invoiceID] == nil {
		b.subs[invoiceID] = make(map[chan models.SSEEvent]struct{})
	}
	b.subs[invoiceID][ch] = struct{}{}
	return ch
}

// Unsubscribe removes the channel for the given invoice ID.
func (b *Broadcaster) Unsubscribe(invoiceID string, ch chan models.SSEEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if m := b.subs[invoiceID]; m != nil {
		delete(m, ch)
		if len(m) == 0 {
			delete(b.subs, invoiceID)
		}
	}
	close(ch)
}
