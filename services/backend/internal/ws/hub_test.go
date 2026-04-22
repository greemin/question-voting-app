package ws

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestHub_Register(t *testing.T) {
	hub := NewHub(false)
	client := &Client{
		SessionID: "session1",
		Send:      make(chan []byte, 256),
	}

	hub.Register(client)

	hub.mu.RLock()
	defer hub.mu.RUnlock()

	if _, ok := hub.rooms["session1"]; !ok {
		t.Fatal("expected room 'session1' to exist")
	}
	if !hub.rooms["session1"][client] {
		t.Fatal("expected client to be registered in 'session1'")
	}
}

func TestHub_Unregister(t *testing.T) {
	hub := NewHub(false)
	client1 := &Client{SessionID: "session1", Send: make(chan []byte, 256)}
	client2 := &Client{SessionID: "session1", Send: make(chan []byte, 256)}

	hub.Register(client1)
	hub.Register(client2)

	// Unregister client1
	hub.Unregister(client1)

	hub.mu.RLock()
	if hub.rooms["session1"][client1] {
		t.Error("expected client1 to be removed")
	}
	if !hub.rooms["session1"][client2] {
		t.Error("expected client2 to remain")
	}
	hub.mu.RUnlock()

	// client1's Send channel should be closed
	select {
	case _, ok := <-client1.Send:
		if ok {
			t.Error("expected client1.Send to be closed")
		}
	default:
		t.Error("expected client1.Send to be closed, but it blocked")
	}

	// Unregister client2 (the last client in the room)
	hub.Unregister(client2)

	hub.mu.RLock()
	if _, ok := hub.rooms["session1"]; ok {
		t.Error("expected room 'session1' to be deleted when empty to free memory")
	}
	hub.mu.RUnlock()
}

func TestHub_Broadcast(t *testing.T) {
	hub := NewHub(false)
	client1 := &Client{SessionID: "session1", Send: make(chan []byte, 256)}
	client2 := &Client{SessionID: "session1", Send: make(chan []byte, 256)}
	client3 := &Client{SessionID: "session2", Send: make(chan []byte, 256)}

	hub.Register(client1)
	hub.Register(client2)
	hub.Register(client3)

	msg := []byte("hello session 1")
	hub.Broadcast("session1", msg)

	// Check client 1 (should receive)
	select {
	case received := <-client1.Send:
		if string(received) != string(msg) {
			t.Errorf("client1 received %q, expected %q", received, msg)
		}
	default:
		t.Error("client1 did not receive message")
	}

	// Check client 3 (should NOT receive, wrong session room)
	select {
	case <-client3.Send:
		t.Error("client3 should not have received message")
	default:
		// Expected behavior
	}
}

func TestHub_Broadcast_BlockedClient(t *testing.T) {
	hub := NewHub(false)
	// Create a client with a buffer size of 1
	client := &Client{SessionID: "session1", Send: make(chan []byte, 1)}
	hub.Register(client)

	// Fill the buffer so the next send will block
	client.Send <- []byte("first message")

	// This broadcast should encounter a blocked channel, and proactively drop the client
	hub.Broadcast("session1", []byte("second message"))

	hub.mu.RLock()
	defer hub.mu.RUnlock()

	if hub.rooms["session1"][client] {
		t.Error("expected client to be removed due to a full/blocked send buffer")
	}
}

// newTestServer starts an httptest server that upgrades connections and serves WS for sessionID.
func newTestServer(t *testing.T, hub *Hub, sessionID string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub.ServeWS(w, r, sessionID)
	}))
}

// dialTestServer connects a gorilla WS client to an httptest server.
func dialTestServer(t *testing.T, server *httptest.Server) *websocket.Conn {
	t.Helper()
	u := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	return conn
}

func TestWritePump_SendsPing(t *testing.T) {
	pingPeriod = 60 * time.Millisecond
	pongWait = 500 * time.Millisecond
	writeWait = 100 * time.Millisecond
	t.Cleanup(func() {
		pingPeriod = 54 * time.Second
		pongWait = 60 * time.Second
		writeWait = 10 * time.Second
	})

	hub := NewHub(false)
	server := newTestServer(t, hub, "s1")
	defer server.Close()

	conn := dialTestServer(t, server)
	defer conn.Close()

	// Gorilla handles ping frames internally inside ReadMessage — they are never
	// returned as a message type. Override the ping handler to capture the event.
	pingReceived := make(chan struct{}, 1)
	conn.SetPingHandler(func(appData string) error {
		select {
		case pingReceived <- struct{}{}:
		default:
		}
		return conn.WriteMessage(websocket.PongMessage, []byte(appData))
	})

	// ReadMessage must be running for the ping handler to fire.
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	select {
	case <-pingReceived:
		// success
	case <-time.After(2 * pingPeriod):
		t.Error("expected ping within 2x ping period, none received")
	}
}

func TestReadPump_ClosesConnectionOnDeadline(t *testing.T) {
	pongWait = 80 * time.Millisecond
	pingPeriod = 60 * time.Millisecond
	writeWait = 100 * time.Millisecond
	t.Cleanup(func() {
		pongWait = 60 * time.Second
		pingPeriod = 54 * time.Second
		writeWait = 10 * time.Second
	})

	hub := NewHub(false)
	server := newTestServer(t, hub, "s2")
	defer server.Close()

	conn := dialTestServer(t, server)
	defer conn.Close()

	// Ignore pings so the server never receives a pong — deadline should fire.
	conn.SetPingHandler(func(string) error { return nil })

	// The server should close the connection after pongWait.
	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, _, err := conn.ReadMessage()
	if err == nil {
		t.Fatal("expected connection to be closed by server, but ReadMessage succeeded")
	}
}
