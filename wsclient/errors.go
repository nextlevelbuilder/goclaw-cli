// Package wsclient provides an exported WebSocket RPC client for the GoClaw protocol.
// It can be imported by external Go modules to interact with a GoClaw server over WebSocket.
package wsclient

import "fmt"

// APIError represents a structured error from the GoClaw server.
// Fields match ErrorShape from the server's protocol package.
type APIError struct {
	StatusCode   int    `json:"status_code,omitempty"`
	Code         string `json:"code"`
	Message      string `json:"message"`
	Details      any    `json:"details,omitempty"`
	Retryable    bool   `json:"retryable,omitempty"`
	RetryAfterMs int    `json:"retryAfterMs,omitempty"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("[%d] %s: %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ErrorCode returns the machine-readable error code.
func (e *APIError) ErrorCode() string { return e.Code }

// ErrorMessage returns the human-readable error message.
func (e *APIError) ErrorMessage() string { return e.Message }

// ErrorDetails returns additional error details.
func (e *APIError) ErrorDetails() any { return e.Details }

// IsRetryable reports whether the operation should be retried.
func (e *APIError) IsRetryable() bool { return e.Retryable }

// RetryAfter returns the suggested retry delay in milliseconds.
func (e *APIError) RetryAfter() int { return e.RetryAfterMs }

// HTTPStatus returns the HTTP status code associated with the error.
func (e *APIError) HTTPStatus() int { return e.StatusCode }

// ErrNotAuthenticated is returned when no credentials are configured.
var ErrNotAuthenticated = &APIError{Code: "not_authenticated", Message: "not authenticated — run 'goclaw auth login'"}

// ErrServerRequired is returned when no server URL is configured.
var ErrServerRequired = &APIError{Code: "server_required", Message: "server URL required — use --server or GOCLAW_SERVER"}
