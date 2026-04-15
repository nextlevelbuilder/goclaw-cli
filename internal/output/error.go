package output

import (
	"encoding/json"
	"fmt"
	"os"
)

// ErrorEnvelope is the JSON shape emitted to stdout/stderr in json mode.
// Matches the server's HTTP error envelope: {"error": {"code": "...", "message": "..."}}
type ErrorEnvelope struct {
	Error *ErrorDetail `json:"error"`
}

// ErrorDetail carries the structured error fields passed through from the server.
type ErrorDetail struct {
	Code         string `json:"code"`
	Message      string `json:"message"`
	Details      any    `json:"details,omitempty"`
	Retryable    bool   `json:"retryable,omitempty"`
	RetryAfterMs int    `json:"retryAfterMs,omitempty"`
}

// Error implements the error interface so ErrorDetail can be returned as error.
func (e *ErrorDetail) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }

// httpErrorEnvelope is used to decode the server HTTP error body.
type httpErrorEnvelope struct {
	Error *ErrorDetail `json:"error"`
}

// ParseHTTPError decodes an HTTP error response body into an ErrorDetail.
// Falls back to a plain-text error if the body is not a valid JSON envelope.
// The status code is used to derive a fallback error code when the body has none.
func ParseHTTPError(body []byte, status int) *ErrorDetail {
	var env httpErrorEnvelope
	if err := json.Unmarshal(body, &env); err == nil && env.Error != nil && env.Error.Code != "" {
		return env.Error
	}
	// Fallback: plain text body or unexpected JSON shape
	msg := string(body)
	if msg == "" {
		msg = fmt.Sprintf("HTTP %d", status)
	}
	return &ErrorDetail{
		Code:    httpStatusCode(status),
		Message: msg,
	}
}

// httpStatusCode returns a canonical error code string for an HTTP status.
func httpStatusCode(status int) string {
	switch status {
	case 400:
		return "INVALID_REQUEST"
	case 401:
		return "UNAUTHORIZED"
	case 403:
		return "TENANT_ACCESS_REVOKED"
	case 404:
		return "NOT_FOUND"
	case 409:
		return "ALREADY_EXISTS"
	case 422:
		return "INVALID_REQUEST"
	case 429:
		return "RESOURCE_EXHAUSTED"
	default:
		if status >= 500 {
			return "INTERNAL"
		}
		return "UNKNOWN"
	}
}

// PrintError writes a structured error to the appropriate output stream.
// In json mode: writes {"error":{...}} to stdout (machine-readable).
// In table/default mode: writes "Error: <message>" to stderr (human-readable).
// err may be an *ErrorDetail, *client.APIError (duck-typed via interface), or a plain error.
func PrintError(err error, format string) {
	detail := toErrorDetail(err)

	if format == "json" {
		env := ErrorEnvelope{Error: detail}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(env)
		return
	}
	fmt.Fprintf(os.Stderr, "Error: %s\n", detail.Message)
}

// apiErrorIface is satisfied by client.APIError without importing the client package
// (avoids circular dependency). We duck-type it.
type apiErrorIface interface {
	error
	ErrorCode() string
	ErrorMessage() string
	ErrorDetails() any
	IsRetryable() bool
	RetryAfter() int
}

// toErrorDetail converts any error to an ErrorDetail, preferring structured fields.
func toErrorDetail(err error) *ErrorDetail {
	if err == nil {
		return &ErrorDetail{Code: "UNKNOWN", Message: "unknown error"}
	}
	// Prefer the rich interface if available
	if ae, ok := err.(apiErrorIface); ok {
		return &ErrorDetail{
			Code:         ae.ErrorCode(),
			Message:      ae.ErrorMessage(),
			Details:      ae.ErrorDetails(),
			Retryable:    ae.IsRetryable(),
			RetryAfterMs: ae.RetryAfter(),
		}
	}
	// Check for *ErrorDetail itself
	if d, ok := err.(*ErrorDetail); ok {
		return d
	}
	// Plain error
	return &ErrorDetail{Code: "UNKNOWN", Message: err.Error()}
}

// apiErrorWithStatus extends apiErrorIface with HTTP status access.
type apiErrorWithStatus interface {
	apiErrorIface
	HTTPStatus() int
}

// FromError maps any error to a CLI exit code.
// Prefers structured server code → HTTP status fallback → generic (1).
func FromError(err error) int {
	if err == nil {
		return ExitSuccess
	}
	if ae, ok := err.(apiErrorIface); ok {
		code := ae.ErrorCode()
		if c := MapServerCode(code); c != ExitGeneric {
			return c
		}
		// Try HTTP status fallback
		if aws, ok := err.(apiErrorWithStatus); ok {
			if s := aws.HTTPStatus(); s > 0 {
				return MapHTTPStatus(s)
			}
		}
	}
	// Try ErrorDetail
	if d, ok := err.(*ErrorDetail); ok {
		if c := MapServerCode(d.Code); c != ExitGeneric {
			return c
		}
	}
	return ExitGeneric
}
