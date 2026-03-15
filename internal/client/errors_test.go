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
