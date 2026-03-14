package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// WSClient is a WebSocket RPC client implementing GoClaw protocol v3.
type WSClient struct {
	conn      *websocket.Conn
	serverURL string
	token     string
	userID    string
	senderID  string
	insecure  bool

	nextID    atomic.Int64
	mu        sync.Mutex   // protects pending and listeners
	writeMu   sync.Mutex   // protects concurrent writes (gorilla requirement)
	pending   map[string]chan *WSResponse
	listeners map[string][]func(*WSEvent)
	done      chan struct{}
}

// WSRequest is a v3 protocol request frame.
type WSRequest struct {
	Type   string `json:"type"`
	ID     string `json:"id"`
	Method string `json:"method"`
	Params any    `json:"params,omitempty"`
}

// WSResponse is a v3 protocol response frame.
type WSResponse struct {
	Type    string          `json:"type"`
	ID      string          `json:"id"`
	OK      bool            `json:"ok"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Error   *APIError       `json:"error,omitempty"`
}

// WSEvent is a v3 protocol server push event.
type WSEvent struct {
	Type    string          `json:"type"`
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// NewWSClient creates a WebSocket client.
func NewWSClient(serverURL, token, userID string, insecure bool) *WSClient {
	return &WSClient{
		serverURL: serverURL,
		token:     token,
		userID:    userID,
		insecure:  insecure,
		pending:   make(map[string]chan *WSResponse),
		listeners: make(map[string][]func(*WSEvent)),
		done:      make(chan struct{}),
	}
}

// Connect establishes WebSocket connection and performs handshake.
func (ws *WSClient) Connect() (*json.RawMessage, error) {
	wsURL := strings.Replace(ws.serverURL, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)
	wsURL = strings.TrimRight(wsURL, "/") + "/ws"

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	if ws.insecure {
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}

	header := http.Header{}
	conn, _, err := dialer.Dial(wsURL, header)
	if err != nil {
		return nil, fmt.Errorf("websocket dial: %w", err)
	}
	ws.conn = conn

	// Start read loop
	go ws.readLoop()

	// Send connect handshake
	params := map[string]any{
		"user_id": ws.userID,
	}
	if ws.token != "" {
		params["token"] = ws.token
	}
	if ws.senderID != "" {
		params["sender_id"] = ws.senderID
	}

	resp, err := ws.Call("connect", params)
	if err != nil {
		ws.Close()
		return nil, fmt.Errorf("handshake failed: %w", err)
	}
	return &resp, nil
}

// Call sends an RPC request and waits for the matching response.
func (ws *WSClient) Call(method string, params any) (json.RawMessage, error) {
	id := fmt.Sprintf("%d", ws.nextID.Add(1))

	req := WSRequest{
		Type:   "req",
		ID:     id,
		Method: method,
		Params: params,
	}

	ch := make(chan *WSResponse, 1)
	ws.mu.Lock()
	ws.pending[id] = ch
	ws.mu.Unlock()

	defer func() {
		ws.mu.Lock()
		delete(ws.pending, id)
		ws.mu.Unlock()
	}()

	// Serialize writes to prevent concurrent write panics
	ws.writeMu.Lock()
	err := ws.conn.WriteJSON(req)
	ws.writeMu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	select {
	case resp := <-ch:
		if !resp.OK && resp.Error != nil {
			return nil, resp.Error
		}
		return resp.Payload, nil
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("timeout waiting for response to %s", method)
	case <-ws.done:
		return nil, fmt.Errorf("connection closed")
	}
}

// Subscribe registers a handler for server push events.
func (ws *WSClient) Subscribe(eventType string, handler func(*WSEvent)) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.listeners[eventType] = append(ws.listeners[eventType], handler)
}

// ClearListeners removes all event listeners (call between Stream invocations).
func (ws *WSClient) ClearListeners() {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.listeners = make(map[string][]func(*WSEvent))
}

// Stream sends an RPC request and delivers events to the handler until run completes.
// Returns the final response payload. Cleans up listeners when done.
func (ws *WSClient) Stream(method string, params any, onEvent func(*WSEvent)) (json.RawMessage, error) {
	// Clear previous listeners to prevent accumulation
	ws.ClearListeners()

	// Use sync.Once to safely close the done channel exactly once
	done := make(chan struct{})
	var closeOnce sync.Once

	ws.Subscribe("chunk", func(e *WSEvent) { onEvent(e) })
	ws.Subscribe("tool.call", func(e *WSEvent) { onEvent(e) })
	ws.Subscribe("tool.result", func(e *WSEvent) { onEvent(e) })
	ws.Subscribe("run.started", func(e *WSEvent) { onEvent(e) })
	ws.Subscribe("run.completed", func(e *WSEvent) {
		onEvent(e)
		closeOnce.Do(func() { close(done) })
	})

	resp, err := ws.Call(method, params)
	if err != nil {
		return nil, err
	}

	// Wait for run.completed or timeout
	select {
	case <-done:
	case <-time.After(10 * time.Minute):
		return nil, fmt.Errorf("stream timeout")
	case <-ws.done:
		return nil, fmt.Errorf("connection closed")
	}

	return resp, nil
}

// SetSenderID sets the sender_id for device pairing reconnection.
func (ws *WSClient) SetSenderID(id string) {
	ws.senderID = id
}

// Close shuts down the WebSocket connection.
func (ws *WSClient) Close() {
	select {
	case <-ws.done:
	default:
		close(ws.done)
	}
	if ws.conn != nil {
		ws.conn.Close()
	}
}

func (ws *WSClient) readLoop() {
	defer ws.Close()
	for {
		_, msg, err := ws.conn.ReadMessage()
		if err != nil {
			return
		}

		// Peek at the type field
		var frame struct {
			Type  string `json:"type"`
			ID    string `json:"id"`
			Event string `json:"event"`
		}
		if err := json.Unmarshal(msg, &frame); err != nil {
			continue
		}

		switch frame.Type {
		case "res":
			var resp WSResponse
			if err := json.Unmarshal(msg, &resp); err != nil {
				continue
			}
			ws.mu.Lock()
			if ch, ok := ws.pending[resp.ID]; ok {
				ch <- &resp
			}
			ws.mu.Unlock()

		case "event":
			var evt WSEvent
			if err := json.Unmarshal(msg, &evt); err != nil {
				continue
			}
			ws.mu.Lock()
			// Copy handlers to avoid holding lock during callback
			handlers := make([]func(*WSEvent), 0, len(ws.listeners[evt.Event])+len(ws.listeners["*"]))
			handlers = append(handlers, ws.listeners[evt.Event]...)
			handlers = append(handlers, ws.listeners["*"]...)
			ws.mu.Unlock()
			for _, h := range handlers {
				h(&evt)
			}
		}
	}
}
