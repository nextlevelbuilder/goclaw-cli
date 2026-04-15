package output

import "os"

// Exit code constants — locked contract for AI/automation consumers.
const (
	ExitSuccess    = 0 // Normal completion
	ExitGeneric    = 1 // Unknown/unmapped error
	ExitAuth       = 2 // Auth failure: UNAUTHORIZED, NOT_PAIRED, TENANT_ACCESS_REVOKED, HTTP 401/403
	ExitNotFound   = 3 // NOT_FOUND, NOT_LINKED, HTTP 404
	ExitValidation = 4 // INVALID_REQUEST, FAILED_PRECONDITION, ALREADY_EXISTS, HTTP 400/409/422
	ExitServer     = 5 // INTERNAL, UNAVAILABLE, AGENT_TIMEOUT, HTTP 5xx
	ExitResource   = 6 // RESOURCE_EXHAUSTED, HTTP 429, connection timeout, DNS fail
)

// serverCodeMap maps server error codes to CLI exit codes.
var serverCodeMap = map[string]int{
	// Auth (2)
	"UNAUTHORIZED":          ExitAuth,
	"NOT_PAIRED":            ExitAuth,
	"TENANT_ACCESS_REVOKED": ExitAuth,

	// Not found (3)
	"NOT_FOUND":  ExitNotFound,
	"NOT_LINKED": ExitNotFound,

	// Validation (4)
	"INVALID_REQUEST":    ExitValidation,
	"FAILED_PRECONDITION": ExitValidation,
	"ALREADY_EXISTS":     ExitValidation,

	// Server (5)
	"INTERNAL":      ExitServer,
	"UNAVAILABLE":   ExitServer,
	"AGENT_TIMEOUT": ExitServer,

	// Resource/network (6)
	"RESOURCE_EXHAUSTED": ExitResource,
}

// MapServerCode maps a server error code string to a CLI exit code.
// Returns ExitGeneric (1) for unknown codes.
func MapServerCode(code string) int {
	if c, ok := serverCodeMap[code]; ok {
		return c
	}
	return ExitGeneric
}

// MapHTTPStatus maps an HTTP status code to a CLI exit code as fallback.
func MapHTTPStatus(status int) int {
	switch {
	case status == 401 || status == 403:
		return ExitAuth
	case status == 404:
		return ExitNotFound
	case status == 400 || status == 409 || status == 422:
		return ExitValidation
	case status == 429:
		return ExitResource
	case status >= 500:
		return ExitServer
	default:
		return ExitGeneric
	}
}

// Exit terminates the process with the given exit code.
func Exit(code int) {
	os.Exit(code)
}
