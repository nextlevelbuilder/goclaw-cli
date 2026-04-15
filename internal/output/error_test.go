package output

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
)

// --- ParseHTTPError ---

func TestParseHTTPError_ValidEnvelope(t *testing.T) {
	body := []byte(`{"error":{"code":"UNAUTHORIZED","message":"token expired"}}`)
	d := ParseHTTPError(body, 401)
	if d.Code != "UNAUTHORIZED" {
		t.Errorf("Code = %q, want UNAUTHORIZED", d.Code)
	}
	if d.Message != "token expired" {
		t.Errorf("Message = %q, want 'token expired'", d.Message)
	}
}

func TestParseHTTPError_PlainText(t *testing.T) {
	body := []byte("internal server error")
	d := ParseHTTPError(body, 500)
	if d.Code != "INTERNAL" {
		t.Errorf("Code = %q, want INTERNAL", d.Code)
	}
	if d.Message != "internal server error" {
		t.Errorf("Message = %q, want 'internal server error'", d.Message)
	}
}

func TestParseHTTPError_EmptyBody(t *testing.T) {
	d := ParseHTTPError([]byte(""), 503)
	if d.Code != "INTERNAL" {
		t.Errorf("Code = %q, want INTERNAL", d.Code)
	}
}

func TestParseHTTPError_MissingCodeField(t *testing.T) {
	// Valid JSON but missing "code" — should fall back to HTTP status mapping
	body := []byte(`{"error":{"message":"oops"}}`)
	d := ParseHTTPError(body, 404)
	if d.Code != "NOT_FOUND" {
		t.Errorf("Code = %q, want NOT_FOUND", d.Code)
	}
}

func TestParseHTTPError_WithRetryable(t *testing.T) {
	body := []byte(`{"error":{"code":"RESOURCE_EXHAUSTED","message":"rate limited","retryable":true,"retryAfterMs":2000}}`)
	d := ParseHTTPError(body, 429)
	if !d.Retryable {
		t.Error("expected Retryable=true")
	}
	if d.RetryAfterMs != 2000 {
		t.Errorf("RetryAfterMs = %d, want 2000", d.RetryAfterMs)
	}
}

// --- ErrorDetail.Error ---

func TestErrorDetail_Error(t *testing.T) {
	d := &ErrorDetail{Code: "NOT_FOUND", Message: "agent not found"}
	if d.Error() != "NOT_FOUND: agent not found" {
		t.Errorf("Error() = %q", d.Error())
	}
}

// --- PrintError JSON output ---

func TestPrintError_JSONMode_APIError(t *testing.T) {
	// We test toErrorDetail + JSON serialisation without capturing stdout
	// (stdout capture in tests is fragile). Instead verify toErrorDetail produces correct shape.
	err := &ErrorDetail{Code: "UNAUTHORIZED", Message: "bad token", Retryable: false}
	detail := toErrorDetail(err)
	if detail.Code != "UNAUTHORIZED" {
		t.Errorf("Code = %q, want UNAUTHORIZED", detail.Code)
	}
}

func TestPrintError_PlainError(t *testing.T) {
	err := errors.New("something went wrong")
	detail := toErrorDetail(err)
	if detail.Code != "UNKNOWN" {
		t.Errorf("Code = %q, want UNKNOWN", detail.Code)
	}
	if detail.Message != "something went wrong" {
		t.Errorf("Message = %q", detail.Message)
	}
}

func TestPrintError_NilError(t *testing.T) {
	detail := toErrorDetail(nil)
	if detail.Code != "UNKNOWN" {
		t.Errorf("Code = %q, want UNKNOWN for nil", detail.Code)
	}
}

// --- FromError exit code mapping via apiErrorIface ---

// fakeAPIError implements apiErrorIface + apiErrorWithStatus for testing.
type fakeAPIError struct {
	code       string
	msg        string
	statusCode int
}

func (f *fakeAPIError) Error() string       { return f.msg }
func (f *fakeAPIError) ErrorCode() string   { return f.code }
func (f *fakeAPIError) ErrorMessage() string { return f.msg }
func (f *fakeAPIError) ErrorDetails() any   { return nil }
func (f *fakeAPIError) IsRetryable() bool   { return false }
func (f *fakeAPIError) RetryAfter() int     { return 0 }
func (f *fakeAPIError) HTTPStatus() int     { return f.statusCode }

func TestFromError_KnownServerCode(t *testing.T) {
	cases := []struct {
		code string
		want int
	}{
		{"UNAUTHORIZED", ExitAuth},
		{"NOT_PAIRED", ExitAuth},
		{"TENANT_ACCESS_REVOKED", ExitAuth},
		{"NOT_FOUND", ExitNotFound},
		{"NOT_LINKED", ExitNotFound},
		{"INVALID_REQUEST", ExitValidation},
		{"FAILED_PRECONDITION", ExitValidation},
		{"ALREADY_EXISTS", ExitValidation},
		{"INTERNAL", ExitServer},
		{"UNAVAILABLE", ExitServer},
		{"AGENT_TIMEOUT", ExitServer},
		{"RESOURCE_EXHAUSTED", ExitResource},
	}
	for _, tc := range cases {
		err := &fakeAPIError{code: tc.code}
		if got := FromError(err); got != tc.want {
			t.Errorf("FromError(%q) = %d, want %d", tc.code, got, tc.want)
		}
	}
}

func TestFromError_HTTPStatusFallback(t *testing.T) {
	// Unknown server code, but HTTP status maps
	err := &fakeAPIError{code: "UNKNOWN_CODE", statusCode: 403}
	if got := FromError(err); got != ExitAuth {
		t.Errorf("FromError via HTTP 403 = %d, want %d", got, ExitAuth)
	}
}

func TestFromError_NilError(t *testing.T) {
	if got := FromError(nil); got != ExitSuccess {
		t.Errorf("FromError(nil) = %d, want %d", got, ExitSuccess)
	}
}

func TestFromError_PlainError(t *testing.T) {
	if got := FromError(errors.New("oops")); got != ExitGeneric {
		t.Errorf("FromError(plain) = %d, want %d", got, ExitGeneric)
	}
}

// --- ErrorEnvelope JSON shape ---

func TestErrorEnvelope_JSONShape(t *testing.T) {
	env := ErrorEnvelope{Error: &ErrorDetail{Code: "NOT_FOUND", Message: "missing"}}
	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var out map[string]any
	_ = json.Unmarshal(data, &out)
	errObj, ok := out["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected 'error' object, got %T", out["error"])
	}
	if errObj["code"] != "NOT_FOUND" {
		t.Errorf("code = %v, want NOT_FOUND", errObj["code"])
	}
}

// --- PrintError stdout capture ---

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	r.Close()
	return string(buf[:n])
}

func TestPrintError_JSONMode_WritesEnvelope(t *testing.T) {
	err := &ErrorDetail{Code: "UNAUTHORIZED", Message: "bad token"}
	out := captureStdout(t, func() {
		PrintError(err, "json")
	})
	var env ErrorEnvelope
	if e := json.Unmarshal([]byte(out), &env); e != nil {
		t.Fatalf("output not valid JSON: %v — got: %s", e, out)
	}
	if env.Error == nil || env.Error.Code != "UNAUTHORIZED" {
		t.Errorf("unexpected envelope: %+v", env.Error)
	}
}

func TestPrintError_JSONMode_WithAPIError(t *testing.T) {
	err := &fakeAPIError{code: "NOT_FOUND", msg: "agent missing"}
	out := captureStdout(t, func() {
		PrintError(err, "json")
	})
	if !json.Valid([]byte(out)) {
		t.Errorf("output not valid JSON: %s", out)
	}
}

func TestPrintError_TableMode_WritesToStderr(t *testing.T) {
	// Just verify it doesn't panic and doesn't write to stdout
	err := &ErrorDetail{Code: "INTERNAL", Message: "server error"}
	out := captureStdout(t, func() {
		PrintError(err, "table")
	})
	if out != "" {
		t.Errorf("table mode should not write to stdout, got: %q", out)
	}
}

// --- httpStatusCode coverage for remaining branches ---

func TestParseHTTPError_400(t *testing.T) {
	d := ParseHTTPError([]byte("bad input"), 400)
	if d.Code != "INVALID_REQUEST" {
		t.Errorf("400 → %q, want INVALID_REQUEST", d.Code)
	}
}

func TestParseHTTPError_401(t *testing.T) {
	d := ParseHTTPError([]byte(""), 401)
	if d.Code != "UNAUTHORIZED" {
		t.Errorf("401 → %q, want UNAUTHORIZED", d.Code)
	}
}

func TestParseHTTPError_403(t *testing.T) {
	d := ParseHTTPError([]byte("forbidden"), 403)
	if d.Code != "TENANT_ACCESS_REVOKED" {
		t.Errorf("403 → %q, want TENANT_ACCESS_REVOKED", d.Code)
	}
}

func TestParseHTTPError_409(t *testing.T) {
	d := ParseHTTPError([]byte("conflict"), 409)
	if d.Code != "ALREADY_EXISTS" {
		t.Errorf("409 → %q, want ALREADY_EXISTS", d.Code)
	}
}

func TestParseHTTPError_422(t *testing.T) {
	d := ParseHTTPError([]byte("unprocessable"), 422)
	if d.Code != "INVALID_REQUEST" {
		t.Errorf("422 → %q, want INVALID_REQUEST", d.Code)
	}
}

func TestParseHTTPError_429(t *testing.T) {
	d := ParseHTTPError([]byte("rate limited"), 429)
	if d.Code != "RESOURCE_EXHAUSTED" {
		t.Errorf("429 → %q, want RESOURCE_EXHAUSTED", d.Code)
	}
}

func TestParseHTTPError_UnknownStatus(t *testing.T) {
	d := ParseHTTPError([]byte("weird"), 418)
	if d.Code != "UNKNOWN" {
		t.Errorf("418 → %q, want UNKNOWN", d.Code)
	}
}

// --- FromError via ErrorDetail ---

func TestFromError_ErrorDetailKnownCode(t *testing.T) {
	err := &ErrorDetail{Code: "INTERNAL", Message: "crash"}
	if got := FromError(err); got != ExitServer {
		t.Errorf("FromError(ErrorDetail{INTERNAL}) = %d, want %d", got, ExitServer)
	}
}

func TestFromError_ErrorDetailUnknownCode(t *testing.T) {
	err := &ErrorDetail{Code: "WEIRD", Message: "something"}
	if got := FromError(err); got != ExitGeneric {
		t.Errorf("FromError(ErrorDetail{WEIRD}) = %d, want %d", got, ExitGeneric)
	}
}
