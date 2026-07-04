package client

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

func (e *APIError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("[%d] %s: %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ErrorCode satisfies output.apiErrorIface — used by PrintError/FromError without import cycle.
func (e *APIError) ErrorCode() string { return e.Code }

// ErrorMessage satisfies output.apiErrorIface.
func (e *APIError) ErrorMessage() string { return e.Message }

// ErrorDetails satisfies output.apiErrorIface.
func (e *APIError) ErrorDetails() any { return e.Details }

// IsRetryable satisfies output.apiErrorIface.
func (e *APIError) IsRetryable() bool { return e.Retryable }

// RetryAfter satisfies output.apiErrorIface.
func (e *APIError) RetryAfter() int { return e.RetryAfterMs }

// HTTPStatus satisfies output.apiErrorWithStatus — returns the HTTP status code.
func (e *APIError) HTTPStatus() int { return e.StatusCode }

// ErrNotAuthenticated is returned when no credentials are configured.
var ErrNotAuthenticated = &APIError{Code: "not_authenticated", Message: "not authenticated — run 'goclaw auth login'"}

// ErrServerRequired is returned when no server URL is configured.
var ErrServerRequired = &APIError{Code: "server_required", Message: "server URL required — use --server or GOCLAW_SERVER"}
