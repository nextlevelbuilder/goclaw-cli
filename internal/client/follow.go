package client

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// FollowHandler is called for every event received from the stream.
// Returning a non-nil error stops the stream immediately (no reconnect).
type FollowHandler func(event *WSEvent) error

// FollowConfig configures FollowStream behaviour.
// Pass nil for defaults (MaxRetries=5, BaseDelay=1s).
// MaxRetries=0 explicitly means zero reconnects (fail on first drop).
// Use a negative value for unlimited retries.
type FollowConfig struct {
	MaxRetries int           // reconnect attempts; 0 = no retries; negative = unlimited
	BaseDelay  time.Duration // initial backoff delay (default 1s)
}

// handlerErr wraps errors returned by the user-supplied handler.
// It is NOT retryable — FollowStream stops immediately when it sees this.
type handlerErr struct{ err error }

func (h *handlerErr) Error() string { return h.err.Error() }
func (h *handlerErr) Unwrap() error { return h.err }

// FollowStream calls method on a fresh WebSocket connection and delivers events
// to handler until ctx is cancelled, handler returns error, or retries exhausted.
//
// Reconnects with exponential backoff on transient connection errors only.
// Handler errors stop the stream immediately (no reconnect attempt).
// Each successful reconnect re-dials and re-sends the initial RPC call.
func FollowStream(ctx context.Context, serverURL, token, userID string, insecure bool,
	method string, params any, handler FollowHandler, cfg *FollowConfig,
) error {
	// Default config when cfg is nil. cfg.MaxRetries=0 literal means no retries.
	maxRetries := 5
	baseDelay := time.Second
	if cfg != nil {
		maxRetries = cfg.MaxRetries
		if cfg.BaseDelay > 0 {
			baseDelay = cfg.BaseDelay
		}
	}

	var attempt int
	for {
		err := followOnce(ctx, serverURL, token, userID, insecure, method, params, handler)
		if err == nil {
			return nil // clean stop
		}

		// Handler-originated error: stop immediately, do NOT retry.
		var hErr *handlerErr
		if errors.As(err, &hErr) {
			return hErr.err
		}

		// ctx cancelled — stop immediately
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Unlimited retries if maxRetries negative
		if maxRetries >= 0 {
			attempt++
			if attempt > maxRetries {
				return fmt.Errorf("follow stream: max retries (%d) exceeded: %w", maxRetries, err)
			}
		}

		delay := baseDelay * (1 << (attempt - 1)) // 1s, 2s, 4s, 8s, 16s
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
}

// followOnce dials, connects, calls method, and delivers events to handler.
// Returns when ctx is cancelled, handler returns error, or connection drops.
func followOnce(ctx context.Context, serverURL, token, userID string, insecure bool,
	method string, params any, handler FollowHandler,
) error {
	ws := NewWSClient(serverURL, token, userID, insecure)
	if _, err := ws.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer ws.Close()

	// Channel to propagate handler errors or connection done signal
	errCh := make(chan error, 1)

	ws.Subscribe("*", func(e *WSEvent) {
		if err := handler(e); err != nil {
			select {
			case errCh <- &handlerErr{err: err}:
			default:
			}
		}
	})

	// Send the initial RPC call — server starts pushing events
	if _, err := ws.Call(method, params); err != nil {
		return fmt.Errorf("call %s: %w", method, err)
	}

	// Block until ctx cancelled, handler error, or connection dropped
	select {
	case <-ctx.Done():
		return nil // clean cancellation
	case err := <-errCh:
		return err
	case <-ws.done:
		return fmt.Errorf("connection closed unexpectedly")
	}
}
