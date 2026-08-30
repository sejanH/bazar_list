package live

import (
	"encoding/json"
	"sync"
)

// Event represents a real-time event sent across the live stream.
type Event struct {
	Type string `json:"type"`
	Data []byte `json:"data"`
}

// Hub manages active SSE subscribers grouped by list ID.
type Hub struct {
	mu          sync.RWMutex
	subscribers map[uint]map[chan Event]struct{}
}

// NewHub creates a new in-memory Hub.
func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[uint]map[chan Event]struct{}),
	}
}

// Subscribe registers a subscriber channel for a given list ID.
// It returns a read-only channel and a thread-safe idempotent unsubscribe cleanup function.
func (h *Hub) Subscribe(listID uint) (<-chan Event, func()) {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan Event, 32)
	if _, ok := h.subscribers[listID]; !ok {
		h.subscribers[listID] = make(map[chan Event]struct{})
	}
	h.subscribers[listID][ch] = struct{}{}

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			h.mu.Lock()
			defer h.mu.Unlock()
			if subs, ok := h.subscribers[listID]; ok {
				delete(subs, ch)
				close(ch)
				if len(subs) == 0 {
					delete(h.subscribers, listID)
				}
			}
		})
	}

	return ch, unsubscribe
}

// Broadcast marshals the payload to JSON and dispatches it non-blocking to all subscribers of the list.
func (h *Hub) Broadcast(listID uint, eventType string, payload any) {
	var data []byte
	if b, ok := payload.([]byte); ok {
		data = b
	} else {
		var err error
		data, err = json.Marshal(payload)
		if err != nil {
			return
		}
	}

	event := Event{
		Type: eventType,
		Data: data,
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	subs, ok := h.subscribers[listID]
	if !ok {
		return
	}

	for ch := range subs {
		select {
		case ch <- event:
		default:
			// Buffer full, drop non-critical event to prevent slow client head-of-line blocking
		}
	}
}
