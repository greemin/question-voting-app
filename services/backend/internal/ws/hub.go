package ws

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Client represents a single connected user.
type Client struct {
	Hub       *Hub
	SessionID string
	Conn      *websocket.Conn
	Send      chan []byte
}

// Hub manages all active clients and broadcasts messages to session rooms.
type Hub struct {
	// rooms maps a sessionID to a set of active clients
	rooms    map[string]map[*Client]bool
	mu       sync.RWMutex
	upgrader websocket.Upgrader
}

// NewHub creates a new WebSocket Hub.
func NewHub(isProduction bool) *Hub {
	up := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	if !isProduction {
		up.CheckOrigin = func(r *http.Request) bool {
			return true // Allow all origins for local development
		}
	}

	return &Hub{
		rooms:    make(map[string]map[*Client]bool),
		upgrader: up,
	}
}

// Register adds a client to a session's room.
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[client.SessionID] == nil {
		h.rooms[client.SessionID] = make(map[*Client]bool)
	}
	h.rooms[client.SessionID][client] = true
	log.Printf("WS client registered for session %q (total: %d)", client.SessionID, len(h.rooms[client.SessionID]))
}

// Unregister removes a client from a session's room.
func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.rooms[client.SessionID]; ok {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			close(client.Send)
			if len(clients) == 0 {
				delete(h.rooms, client.SessionID)
			}
		}
	}
	log.Printf("WS client unregistered for session %q", client.SessionID)
}

// Broadcast sends a message to all connected clients in a specific session.
func (h *Hub) Broadcast(sessionID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := h.rooms[sessionID]
	log.Printf("WS broadcast to session %q: %d client(s)", sessionID, len(clients))
	for client := range clients {
		select {
		case client.Send <- message:
		default:
			// If send buffer is full or blocked, assume client is disconnected
			close(client.Send)
			delete(clients, client)
		}
	}
}

// ServeWS upgrades the HTTP connection and registers the client.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request, sessionID string) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	client := &Client{
		Hub:       h,
		SessionID: sessionID,
		Conn:      conn,
		Send:      make(chan []byte, 256),
	}

	h.Register(client)

	// Start pump goroutines
	go client.writePump()
	go client.readPump()
}

// readPump listens for incoming messages from the client.
func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
	}()
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}
	}
}

// writePump sends messages from the hub to the client.
func (c *Client) writePump() {
	defer c.Conn.Close()
	for {
		message, ok := <-c.Send
		if !ok {
			// Hub closed the channel
			c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		w, err := c.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			log.Printf("WS writePump NextWriter error for session %q: %v", c.SessionID, err)
			return
		}
		w.Write(message)

		if err := w.Close(); err != nil {
			log.Printf("WS writePump Close error for session %q: %v", c.SessionID, err)
			return
		}
	}
}
