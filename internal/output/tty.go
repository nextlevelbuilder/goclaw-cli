package output

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// IsTTY reports whether the given file descriptor is connected to a terminal.
func IsTTY(fd int) bool {
	return term.IsTerminal(fd)
}

// validFormats is the set of supported output formats.
var validFormats = map[string]bool{"table": true, "json": true, "yaml": true}

// isValidFormat reports whether s is a supported output format.
func isValidFormat(s string) bool { return validFormats[s] }

// ResolveFormat determines the output format using the precedence chain:
//  1. Explicit --output flag value (non-empty flagVal)
//  2. GOCLAW_OUTPUT environment variable
//  3. stdout is a TTY → "table"
//  4. else → "json"
//
// Valid values: "table", "json", "yaml".
// Invalid flag values are passed through (cobra/caller validation catches them).
// Invalid GOCLAW_OUTPUT values emit a warning on stderr and fall through to
// TTY detection, to avoid silently mis-formatting AI-consumer output.
func ResolveFormat(flagVal string) string {
	if flagVal != "" {
		return flagVal
	}
	if env := os.Getenv("GOCLAW_OUTPUT"); env != "" {
		if isValidFormat(env) {
			return env
		}
		fmt.Fprintf(os.Stderr, "goclaw: warning: invalid GOCLAW_OUTPUT=%q, falling back to auto-detect (valid: table|json|yaml)\n", env)
	}
	if IsTTY(int(os.Stdout.Fd())) {
		return "table"
	}
	return "json"
}
