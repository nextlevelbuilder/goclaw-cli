package client

import (
	"testing"
)

func TestAPIError_WithStatusCode(t *testing.T) {
	err := &APIError{StatusCode: 404, Code: "not_found", Message: "agent not found"}
	expected := "[404] not_found: agent not found"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestAPIError_WithoutStatusCode(t *testing.T) {
	err := &APIError{Code: "validation", Message: "name required"}
	expected := "validation: name required"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestSentinelErrors(t *testing.T) {
	if ErrNotAuthenticated.Code != "not_authenticated" {
		t.Errorf("unexpected code: %s", ErrNotAuthenticated.Code)
	}
	if ErrServerRequired.Code != "server_required" {
		t.Errorf("unexpected code: %s", ErrServerRequired.Code)
	}
}

// TestAPIError_InterfaceMethods covers the output.apiErrorIface + apiErrorWithStatus
// methods added to APIError for duck-typed error handling.
func TestAPIError_InterfaceMethods(t *testing.T) {
	err := &APIError{
		StatusCode:   429,
		Code:         "RESOURCE_EXHAUSTED",
		Message:      "rate limited",
		Details:      map[string]any{"limit": 100},
		Retryable:    true,
		RetryAfterMs: 3000,
	}

	if err.ErrorCode() != "RESOURCE_EXHAUSTED" {
		t.Errorf("ErrorCode() = %q, want RESOURCE_EXHAUSTED", err.ErrorCode())
	}
	if err.ErrorMessage() != "rate limited" {
		t.Errorf("ErrorMessage() = %q, want 'rate limited'", err.ErrorMessage())
	}
	if err.ErrorDetails() == nil {
		t.Error("ErrorDetails() should not be nil")
	}
	if !err.IsRetryable() {
		t.Error("IsRetryable() should be true")
	}
	if err.RetryAfter() != 3000 {
		t.Errorf("RetryAfter() = %d, want 3000", err.RetryAfter())
	}
	if err.HTTPStatus() != 429 {
		t.Errorf("HTTPStatus() = %d, want 429", err.HTTPStatus())
	}
}

func TestAPIError_InterfaceMethods_ZeroValues(t *testing.T) {
	err := &APIError{Code: "INTERNAL", Message: "crash"}
	if err.ErrorDetails() != nil {
		t.Error("ErrorDetails() should be nil for zero value")
	}
	if err.IsRetryable() {
		t.Error("IsRetryable() should be false for zero value")
	}
	if err.RetryAfter() != 0 {
		t.Errorf("RetryAfter() = %d, want 0", err.RetryAfter())
	}
	if err.HTTPStatus() != 0 {
		t.Errorf("HTTPStatus() = %d, want 0", err.HTTPStatus())
	}
}
