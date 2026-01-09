package httpapi

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]struct{}),
	}
}

func (h *Hub) Add(conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) Remove(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
}

func (h *Hub) Broadcast(message []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			conn.Close()
			delete(h.clients, conn)
		}
	}
}

type WSHandler struct {
	hub *Hub
}

func NewWSHandler(hub *Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "upgrade failed", http.StatusBadRequest)
		return
	}
	h.hub.Add(conn)

	go func() {
		defer func() {
			h.hub.Remove(conn)
			conn.Close()
		}()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}
