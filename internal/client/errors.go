package client

import "fmt"

// APIError represents a structured error from the GoClaw server.
type APIError struct {
	StatusCode int    `json:"status_code,omitempty"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("[%d] %s: %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ErrNotAuthenticated is returned when no credentials are configured.
var ErrNotAuthenticated = &APIError{Code: "not_authenticated", Message: "not authenticated — run 'goclaw auth login'"}

// ErrServerRequired is returned when no server URL is configured.
var ErrServerRequired = &APIError{Code: "server_required", Message: "server URL required — use --server or GOCLAW_SERVER"}
