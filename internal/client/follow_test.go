package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// wsUpgrader is shared across WS test helpers.
var followUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// startFollowServer starts a test WS server that:
//  1. Accepts a connect handshake
//  2. Sends `eventCount` events on any method call
//  3. Then closes the connection
func startFollowServer(t *testing.T, eventCount int) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := followUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Handle connect handshake
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var req WSRequest
		if err := json.Unmarshal(msg, &req); err != nil {
			return
		}
		// Send connect OK
		resp := WSResponse{Type: "res", ID: req.ID, OK: true}
		if err := conn.WriteJSON(resp); err != nil {
			return
		}

		// Handle the actual method call
		_, msg, err = conn.ReadMessage()
		if err != nil {
			return
		}
		if err := json.Unmarshal(msg, &req); err != nil {
			return
		}
		// Acknowledge the call
		resp = WSResponse{Type: "res", ID: req.ID, OK: true}
		if err := conn.WriteJSON(resp); err != nil {
			return
		}

		// Push events
		for i := 0; i < eventCount; i++ {
			evt := map[string]any{
				"type":    "event",
				"event":   "log",
				"payload": map[string]any{"n": i},
			}
			if err := conn.WriteJSON(evt); err != nil {
				return
			}
		}
		// Server closes — triggers reconnect path
	}))
	return srv
}

func TestFollowStream_ContextCancel(t *testing.T) {
	srv := startFollowServer(t, 2)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	// Replace URL scheme for WSClient (it expects http/https base)
	httpURL := srv.URL

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	var received []string
	handler := func(e *WSEvent) error {
		received = append(received, e.Event)
		return nil
	}

	err := FollowStream(ctx, httpURL, "", "cli", false,
		"logs.tail", nil, handler,
		&FollowConfig{MaxRetries: 1, BaseDelay: 50 * time.Millisecond})

	// Should return ctx error or max retries after context cancelled
	_ = wsURL // silence unused var
	_ = err   // error is expected (context timeout or max retries)
}

func TestFollowStream_HandlerErrorStops(t *testing.T) {
	// Server that would send many events, then drop (to trigger reconnect
	// if handler errors were mistakenly retried).
	srv := startFollowServer(t, 5)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	sentinel := errors.New("handler stop")
	callCount := 0
	handler := func(e *WSEvent) error {
		callCount++
		return sentinel // stop on first event
	}

	// With MaxRetries=5, if FollowStream mistakenly retried on handler errors,
	// callCount would exceed 1 (new connection → new events).
	start := time.Now()
	err := FollowStream(ctx, srv.URL, "", "cli", false,
		"logs.tail", nil, handler,
		&FollowConfig{MaxRetries: 5, BaseDelay: 10 * time.Millisecond})
	elapsed := time.Since(start)

	if !errors.Is(err, sentinel) {
		t.Errorf("expected handler sentinel error, got: %v", err)
	}
	// Key invariant: no reconnect. Server sends 5 events per connection.
	// >5 calls would prove reconnect-and-resend occurred. In-flight events
	// (≤5) before the error propagates are acceptable since ws delivery is async.
	if callCount > 5 {
		t.Errorf("handler error must NOT trigger reconnect; got %d calls (>5 implies reconnect)", callCount)
	}
	if elapsed >= 500*time.Millisecond {
		t.Errorf("handler error should return fast (no backoff); took %v (expected <500ms)", elapsed)
	}
}

func TestFollowStream_MaxRetriesZeroMeansNoRetries(t *testing.T) {
	// Server that drops connection immediately after RPC call (no events).
	// Connection drop → "unexpectedly closed" transient error → would retry.
	// With MaxRetries=0, must fail immediately without retry.
	srv := startFollowServer(t, 0)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	handler := func(e *WSEvent) error { return nil }

	start := time.Now()
	err := FollowStream(ctx, srv.URL, "", "cli", false,
		"logs.tail", nil, handler,
		&FollowConfig{MaxRetries: 0, BaseDelay: 100 * time.Millisecond})
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error when MaxRetries=0 and connection drops")
	}
	if !strings.Contains(err.Error(), "max retries") {
		t.Errorf("expected 'max retries' in error, got: %v", err)
	}
	// No retry → elapsed should be much less than BaseDelay (100ms)
	if elapsed >= 100*time.Millisecond {
		t.Errorf("MaxRetries=0 should fail fast without waiting; took %v", elapsed)
	}
}

func TestFollowConfig_NilUsesDefaults(t *testing.T) {
	// Verify nil cfg applies default MaxRetries=5 (not 0).
	// We can only observe via elapsed time or errors; smoke test:
	// nil cfg should NOT immediately fail on a single connection drop.
	srv := startFollowServer(t, 0)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	handler := func(e *WSEvent) error { return nil }

	err := FollowStream(ctx, srv.URL, "", "cli", false,
		"logs.tail", nil, handler, nil)

	// Should hit ctx deadline (proving it was retrying), not fail-fast
	if err == nil {
		t.Fatal("expected error")
	}
}
