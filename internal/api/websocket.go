package api

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

// WSHub manages WebSocket connections for log streaming
type WSHub struct {
	clients         map[string]map[*websocket.Conn]bool
	broadcast       chan WSMessage
	register        chan *WSSubscription
	unregister      chan *WSSubscription
	mu              sync.RWMutex
	allowedOrigins  string
}

// WSMessage represents a message to broadcast
type WSMessage struct {
	Type      string      `json:"type"` // "log", "metric", "status"
	RunID     string      `json:"run_id"`
	Timestamp string      `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
}

// WSSubscription represents a client subscribing to a run's logs
type WSSubscription struct {
	RunID string
	Conn  *websocket.Conn
}

// isOriginAllowed checks if a WebSocket origin is allowed
func isOriginAllowed(origin string, allowedOrigins string) bool {
	if allowedOrigins == "*" {
		return true
	}

	// Parse origin URL
	originURL, err := url.Parse(origin)
	if err != nil {
		return false
	}

	// Split allowed origins by comma
	allowed := strings.Split(allowedOrigins, ",")
	for _, ao := range allowed {
		ao = strings.TrimSpace(ao)
		allowedURL, err := url.Parse(ao)
		if err != nil {
			continue
		}
		if originURL.Scheme == allowedURL.Scheme && originURL.Host == allowedURL.Host {
			return true
		}
	}
	return false
}

// NewWSHub creates a new WebSocket hub with CORS validation
func NewWSHub(allowedOrigins string) *WSHub {
	return &WSHub{
		clients:        make(map[string]map[*websocket.Conn]bool),
		broadcast:      make(chan WSMessage, 100),
		register:       make(chan *WSSubscription),
		unregister:     make(chan *WSSubscription),
		allowedOrigins: allowedOrigins,
	}
}

// Run starts the WebSocket hub
func (h *WSHub) Run() {
	for {
		select {
		case sub := <-h.register:
			h.mu.Lock()
			if h.clients[sub.RunID] == nil {
				h.clients[sub.RunID] = make(map[*websocket.Conn]bool)
			}
			h.clients[sub.RunID][sub.Conn] = true
			h.mu.Unlock()
			log.Printf("Client registered for run %s\n", sub.RunID)

		case unsub := <-h.unregister:
			h.mu.Lock()
			if conns, ok := h.clients[unsub.RunID]; ok {
				if _, ok := conns[unsub.Conn]; ok {
					delete(conns, unsub.Conn)
					unsub.Conn.Close()
					if len(conns) == 0 {
						delete(h.clients, unsub.RunID)
					}
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			if conns, ok := h.clients[msg.RunID]; ok {
				for conn := range conns {
					if err := conn.WriteJSON(msg); err != nil {
						h.unregister <- &WSSubscription{
							RunID: msg.RunID,
							Conn:  conn,
						}
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all clients for a run
func (h *WSHub) Broadcast(msg WSMessage) {
	h.broadcast <- msg
}

// HandleLogsWebSocket handles WebSocket upgrade for log streaming
func (h *WSHub) HandleLogsWebSocket(w http.ResponseWriter, r *http.Request) {
	// Validate origin for WebSocket connection
	origin := r.Header.Get("Origin")
	if origin != "" && !isOriginAllowed(origin, h.allowedOrigins) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	runID := r.URL.Query().Get("run_id")
	if runID == "" {
		http.Error(w, "Missing run_id parameter", http.StatusBadRequest)
		return
	}

	// Create upgrader with proper origin check
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return origin == "" || isOriginAllowed(origin, h.allowedOrigins)
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket: %v\n", err)
		return
	}

	sub := &WSSubscription{
		RunID: runID,
		Conn:  conn,
	}

	h.register <- sub

	// Read messages from client (for keep-alive pings)
	go func() {
		defer func() {
			h.unregister <- sub
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v\n", err)
				}
				return
			}
		}
	}()
}
