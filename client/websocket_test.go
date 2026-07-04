package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// mockWSServer creates a test WebSocket server that handles connect + custom methods.
func mockWSServer(t *testing.T, handler func(conn *websocket.Conn)) *httptest.Server {
	t.Helper()
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade: %v", err)
		}
		defer conn.Close()
		handler(conn)
	}))
}

func TestWSClient_ConnectAndCall(t *testing.T) {
	srv := mockWSServer(t, func(conn *websocket.Conn) {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var req WSRequest
			json.Unmarshal(msg, &req)

			resp := WSResponse{Type: "res", ID: req.ID, OK: true}
			switch req.Method {
			case "connect":
				resp.Payload = json.RawMessage(`{"role":"admin"}`)
			case "status":
				resp.Payload = json.RawMessage(`{"version":"1.0"}`)
			}
			conn.WriteJSON(resp)
		}
	})
	defer srv.Close()

	wsURL := strings.Replace(srv.URL, "http://", "ws://", 1)
	// Strip /ws since Connect() adds it
	wsURL = strings.TrimSuffix(wsURL, "/ws")
	// NewWSClient expects http URL, it converts internally
	httpURL := strings.Replace(wsURL, "ws://", "http://", 1)

	ws := NewWSClient(httpURL, "test-tok", "user1", false)
	connResp, err := ws.Connect()
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if connResp == nil {
		t.Fatal("expected connect response")
	}

	// Call status
	data, err := ws.Call("status", nil)
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	var result map[string]any
	json.Unmarshal(data, &result)
	if result["version"] != "1.0" {
		t.Errorf("expected version 1.0, got %v", result["version"])
	}

	ws.Close()
}

func TestWSClient_CallError(t *testing.T) {
	srv := mockWSServer(t, func(conn *websocket.Conn) {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var req WSRequest
			json.Unmarshal(msg, &req)

			if req.Method == "connect" {
				conn.WriteJSON(WSResponse{Type: "res", ID: req.ID, OK: true,
					Payload: json.RawMessage(`{}`)})
				continue
			}

			conn.WriteJSON(WSResponse{Type: "res", ID: req.ID, OK: false,
				Error: &APIError{Code: "not_found", Message: "agent not found"}})
		}
	})
	defer srv.Close()

	httpURL := strings.Replace(srv.URL, "http://", "http://", 1)
	ws := NewWSClient(httpURL, "tok", "u", false)
	_, err := ws.Connect()
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer ws.Close()

	_, err = ws.Call("agents.get", map[string]any{"id": "999"})
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Code != "not_found" {
		t.Errorf("expected not_found, got %s", apiErr.Code)
	}
}

func TestWSClient_EventSubscription(t *testing.T) {
	srv := mockWSServer(t, func(conn *websocket.Conn) {
		// Handle connect
		_, msg, _ := conn.ReadMessage()
		var req WSRequest
		json.Unmarshal(msg, &req)
		conn.WriteJSON(WSResponse{Type: "res", ID: req.ID, OK: true, Payload: json.RawMessage(`{}`)})

		// Push an event
		time.Sleep(50 * time.Millisecond)
		conn.WriteJSON(WSEvent{Type: "event", Event: "test.event", Payload: json.RawMessage(`{"data":"hello"}`)})
	})
	defer srv.Close()

	httpURL := strings.Replace(srv.URL, "http://", "http://", 1)
	ws := NewWSClient(httpURL, "tok", "u", false)
	_, err := ws.Connect()
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer ws.Close()

	received := make(chan string, 1)
	ws.Subscribe("test.event", func(e *WSEvent) {
		var d struct{ Data string }
		json.Unmarshal(e.Payload, &d)
		received <- d.Data
	})

	select {
	case data := <-received:
		if data != "hello" {
			t.Errorf("expected hello, got %s", data)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestWSClient_ClearListeners(t *testing.T) {
	ws := &WSClient{
		listeners: map[string][]func(*WSEvent){
			"chunk": {func(e *WSEvent) {}},
			"*":     {func(e *WSEvent) {}},
		},
	}
	ws.ClearListeners()
	if len(ws.listeners) != 0 {
		t.Errorf("expected empty listeners after clear, got %d", len(ws.listeners))
	}
}

func TestWSClient_SetSenderID(t *testing.T) {
	ws := NewWSClient("http://example.com", "", "u", false)
	ws.SetSenderID("sender-123")
	if ws.senderID != "sender-123" {
		t.Errorf("expected sender-123, got %s", ws.senderID)
	}
}
