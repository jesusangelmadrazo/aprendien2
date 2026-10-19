package ws

import (
	"context"
	"log/slog"
	"sync"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

// Hub keeps a single active connection per user so the backend can push
// events (offers, status changes, live location) without polling.
type Hub struct {
	mu    sync.RWMutex
	conns map[string]*websocket.Conn
}

func NewHub() *Hub {
	return &Hub{conns: make(map[string]*websocket.Conn)}
}

func (h *Hub) Register(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[userID] = conn
}

func (h *Hub) Unregister(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conns[userID] == conn {
		delete(h.conns, userID)
	}
}

// SendToUser is a best-effort push: if the user isn't connected the event is dropped.
// State that matters (lesson/offer status) always lives in Postgres, so a missed
// push just means the client picks it up on its next poll/reconnect.
func (h *Hub) SendToUser(ctx context.Context, userID string, event Event) {
	h.mu.RLock()
	conn, ok := h.conns[userID]
	h.mu.RUnlock()
	if !ok {
		return
	}
	if err := wsjson.Write(ctx, conn, event); err != nil {
		slog.Warn("failed to push ws event", "user_id", userID, "type", event.Type, "error", err)
	}
}
