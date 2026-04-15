# GoClaw CLI - Code Standards

## File Organization

### Directory Structure

```
goclaw-cli/
├── main.go                 # Entry point
├── cmd/                    # Command definitions (Cobra)
│   ├── root.go            # Root command and global flags
│   ├── auth.go            # Authentication commands
│   ├── agents.go          # Agent CRUD commands
│   ├── chat.go            # Chat command (interactive + streaming)
│   ├── sessions.go        # Session management
│   ├── skills.go          # Skill management
│   ├── mcp.go             # MCP server management
│   ├── providers.go       # LLM provider management
│   ├── tools.go           # Custom tool management
│   ├── cron.go            # Scheduled job management
│   ├── teams.go           # Team management
│   ├── channels.go        # Channel management
│   ├── traces.go          # Trace viewing
│   ├── memory.go          # Memory document management
│   ├── config_cmd.go      # Config management
│   ├── logs.go            # Log streaming
│   ├── storage.go         # Storage browsing
│   ├── admin.go           # Admin commands
│   ├── status.go          # Health checks
│   ├── version.go         # Version display
│   ├── helpers.go         # Shared command helpers
│   └── [more commands...]
├── internal/               # Private packages
│   ├── client/            # HTTP + WebSocket clients
│   │   ├── http.go        # REST API client
│   │   ├── websocket.go   # WebSocket streaming
│   │   ├── auth.go        # Auth helpers (keyring, device pairing)
│   │   └── errors.go      # API error handling
│   ├── config/            # Configuration management
│   │   └── config.go      # Config loading, precedence, profiles
│   ├── output/            # Output formatting
│   │   └── output.go      # Table, JSON, YAML formatters
│   └── tui/               # Terminal UI
│       └── prompt.go      # Interactive prompts
├── Makefile               # Build automation
├── go.mod                 # Module definition
├── .goreleaser.yaml       # Release configuration
├── .github/workflows/     # CI/CD pipelines
│   ├── ci.yaml           # Test and build
│   └── release.yaml      # Release workflow
└── README.md              # User documentation
```

### Naming Conventions

- **Go Files:** `snake_case.go` (e.g., `config_cmd.go`, `websocket.go`)
- **Packages:** Lowercase, no underscores (e.g., `internal/client`, `internal/config`)
- **Functions:** `PascalCase` (exported), `camelCase` (unexported)
- **Variables:** `camelCase` (local), `CONSTANT_CASE` (constants)
- **Interfaces:** `Reader`, `Writer`, `Handler` (noun-based)
- **Error Types:** `ErrXxx` (e.g., `ErrInvalidToken`, `ErrNotFound`)

---

## Go Conventions

### Module & Imports

```go
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/client"
	"github.com/nextlevelbuilder/goclaw-cli/internal/config"
	"github.com/spf13/cobra"
)
```

**Rules:**
- Organize imports: stdlib → external → internal
- Use fully qualified package imports (no `.` or `_` aliases)
- Each package on its own line, no grouping with commas

### Error Handling

**Pattern: Wrap with context using `fmt.Errorf("%w")`**

```go
// Good: Wrapped error with context
if err != nil {
	return fmt.Errorf("fetch agent: %w", err)
}

// Bad: Strings without wrapping
if err != nil {
	return fmt.Errorf("error: %v", err)
}

// Bad: Losing context
if err != nil {
	return err
}
```

**Error Messages:**
- Lowercase (no "Error" prefix in message)
- Describe operation, not error type
- Chain context: `"load config: read file: %w"`

**Central Error Handler (Phase 0 — locked contract):**
All command errors bubble via `return err` to `cmd.Execute()` → `output.PrintError(err, format)` + `output.Exit(output.FromError(err))`. Do NOT print errors in individual commands. The `output.FromError()` function maps server error codes (12 known codes) to exit codes 0-6. Server errors are passed through with `code`, `message`, and `details` fields preserved in JSON mode.

**No Import Cycles:**
The `output` package uses duck-typed interfaces (`apiErrorIface`, `apiErrorWithStatus`) to inspect `client.APIError` fields without importing `client`. `APIError` implements these interfaces via exported methods (`ErrorCode()`, `HTTPStatus()`, etc.).

### API Client Pattern

**HTTP Client (internal/client/http.go):**

```go
type HTTPClient struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
	Verbose    bool
}

// Methods: Get, Post, Put, Patch, Delete
func (c *HTTPClient) Get(path string) (json.RawMessage, error) {
	// Implementation
}

func (c *HTTPClient) Post(path string, body any) (json.RawMessage, error) {
	// Implementation
}
```

**Response Handling:**
```go
type apiResponse struct {
	OK      bool            `json:"ok"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Error   *APIError       `json:"error,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
```

**Usage in Commands:**
```go
c, err := newHTTP()  // Helper creates client from config
if err != nil {
	return err
}
data, err := c.Get("/v1/agents")
if err != nil {
	return err
}
// Unmarshal raw JSON
agents := unmarshalList(data)
```

### Command Structure (Cobra)

**Pattern: One command group per file**

```go
// agents.go
var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Manage agents",
}

var agentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Implementation
		return nil
	},
}

var agentsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get agent details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Implementation
		return nil
	},
}

func init() {
	agentsCmd.AddCommand(agentsListCmd, agentsGetCmd)
	// Add flags
	agentsCreateCmd.Flags().String("name", "", "Agent name")
	agentsCreateCmd.Flags().String("provider", "", "LLM provider")
}
```

**Command Flags:**
```go
// Persistent flags (all subcommands)
rootCmd.PersistentFlags().String("server", "", "Server URL")

// Command-specific flags
createCmd.Flags().String("name", "", "Resource name")
createCmd.MarkFlagRequired("name")
```

### Configuration Loading

**Precedence: CLI flags > Environment > Config file > Defaults**

```go
func Load(cmd *cobra.Command) (*Config, error) {
	cfg := &Config{OutputFormat: "table"}

	// 1. Load from file
	if fc, err := loadFile(); err == nil {
		cfg.Server = fc.GetServer()
	}

	// 2. Overlay environment variables
	if v := os.Getenv("GOCLAW_SERVER"); v != "" {
		cfg.Server = v
	}

	// 3. Overlay CLI flags (only if set)
	if cmd.Flags().Changed("server") {
		server, _ := cmd.Flags().GetString("server")
		cfg.Server = server
	}

	return cfg, nil
}
```

### WebSocket Streaming

**Pattern: Send/receive on channels**

```go
type WebSocket struct {
	conn *websocket.Conn
}

func (ws *WebSocket) Stream(ctx context.Context, fn func(msg []byte) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		_, data, err := ws.conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("read message: %w", err)
		}

		if err := fn(data); err != nil {
			return fmt.Errorf("process message: %w", err)
		}
	}
}
```

### Output Formatting

**Internal/output/output.go:**

```go
type Printer struct {
	Format string // "table", "json", "yaml"
}

// Print() dispatches to table, JSON, or YAML
func (p *Printer) Print(data any) {
	switch p.Format {
	case "json":
		p.printJSON(data)
	case "yaml":
		p.printYAML(data)
	default:
		p.printTable(data)
	}
}

type Table struct {
	headers []string
	rows    [][]string
}

func NewTable(headers ...string) *Table {
	return &Table{headers: headers}
}

func (t *Table) AddRow(values ...string) {
	t.rows = append(t.rows, values)
}
```

### Testing

**Table-Driven Tests:**

```go
func TestGet(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantCode  int
		wantError bool
		wantBody  string
	}{
		{
			name:     "success",
			path:     "/v1/agents",
			wantCode: 200,
			wantBody: `{"ok":true,"payload":[]}`,
		},
		{
			name:      "not found",
			path:      "/v1/agents/invalid",
			wantCode:  404,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test implementation
		})
	}
}
```

**Race Detection:**
```bash
go test -race ./...
```

---

## Code Quality

### Linting

```bash
make lint       # go vet ./...
```

**Rules:**
- No unused variables or imports
- No unreachable code
- Type-safe operations
- Proper error handling

### Documentation

**Function Comments:**
```go
// Get retrieves data from the API at the given path.
// It returns raw JSON payload or an error.
func (c *HTTPClient) Get(path string) (json.RawMessage, error) {
```

**Package Comments:**
```go
// Package client provides HTTP and WebSocket clients for GoClaw API.
package client
```

### Performance Considerations

- Reuse HTTP client (connection pooling)
- Stream large responses (WebSocket)
- Avoid unnecessary unmarshaling
- Use `json.RawMessage` to defer unmarshaling

### Security Best Practices

- **Credentials:** Pass via env var `GOCLAW_TOKEN`, not CLI args
- **TLS:** Default to secure, `--insecure` only for testing
- **Keyring:** Use OS keyring for persistent storage
- **Logging:** Never log sensitive data
- **Input Validation:** Validate user input before API calls

---

## Build & Deployment

### Build Process

```bash
make build      # Build goclaw binary locally
make test       # Run all tests with race detector
make lint       # Run go vet
make install    # Install to GOPATH/bin
make clean      # Remove binaries and dist/
```

### Environment Variables (Build)

```bash
VERSION=$(git describe --tags --always --dirty)
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

go build -ldflags "-X github.com/nextlevelbuilder/goclaw-cli/cmd.Version=$VERSION"
```

### Release via GoReleaser

```bash
git tag v1.0.0
git push origin v1.0.0  # Triggers CI/CD
```

Generated artifacts:
- `goclaw_X.X.X_linux_amd64.tar.gz`
- `goclaw_X.X.X_darwin_amd64.tar.gz`
- `goclaw_X.X.X_windows_amd64.zip`
- (arm64 variants for each OS)

---

## Dependencies

### Core Libraries

| Dependency | Version | Use Case |
|------------|---------|----------|
| `cobra` | v1.10+ | CLI framework |
| `gorilla/websocket` | v1.5+ | WebSocket streaming |
| `golang.org/x/term` | Latest | Terminal utilities (raw mode) |
| `yaml.v3` | v3.0+ | YAML parsing |

### No ORM
Raw HTTP calls with JSON marshaling—no database abstraction layer.

### No External CLIs
Single statically-linked binary; no shell dependencies.

---

## Configuration Standards

### File: ~/.goclaw/config.yaml

```yaml
active_profile: production
profiles:
  - name: production
    server: https://goclaw.example.com
    default_agent: myagent
    output: table
```

**Credentials:** Stored in OS keyring, not in config file.

### Environment Variables

| Variable | Purpose | Example |
|----------|---------|---------|
| `GOCLAW_SERVER` | Server URL | `https://goclaw.example.com` |
| `GOCLAW_TOKEN` | Auth token | `sk_prod_abc123xyz` |
| `GOCLAW_OUTPUT` | Output format | `json` |

---

## Common Patterns

### Creating a Command Handler

```go
var myCmd = &cobra.Command{
	Use:   "subcommand",
	Short: "Brief description",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate args
		if len(args) != 1 {
			return fmt.Errorf("expected 1 argument, got %d", len(args))
		}

		// Create HTTP client
		c, err := newHTTP()
		if err != nil {
			return err
		}

		// Make API call
		data, err := c.Get("/v1/resource/" + args[0])
		if err != nil {
			return fmt.Errorf("fetch resource: %w", err)
		}

		// Format and output
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(myCmd)
	myCmd.Flags().String("option", "", "Option description")
}
```

### Interactive Prompt

```go
import "github.com/nextlevelbuilder/goclaw-cli/internal/tui"

response, err := tui.Prompt("Enter agent name: ")
if err != nil {
	return fmt.Errorf("prompt: %w", err)
}
```

### Streaming with WebSocket

```go
ws, err := client.NewWebSocket(cfg.Server, cfg.Token, "/v1/chat/stream")
if err != nil {
	return err
}

ctx := context.Background()
err = ws.Stream(ctx, func(msg []byte) error {
	fmt.Println(string(msg))
	return nil
})
```

---

## Code Review Checklist

Before submitting PR:

- [ ] No unused imports or variables
- [ ] Errors wrapped with context (`fmt.Errorf("%w")`)
- [ ] No hardcoded secrets (use env vars)
- [ ] Command flags validated and documented
- [ ] Tests added for new code
- [ ] `go vet` passes
- [ ] `go test -race` passes
- [ ] No breaking changes without migration guide

---

---

## Phase 0-4: AI-First Patterns & Locked Contracts

### TTY-Aware Output (Phase 0 Locked Contract)

**Format Auto-Detection (precedence):**
1. `--output` flag (explicit)
2. `GOCLAW_OUTPUT` environment variable
3. `stdout` is a TTY → `"table"`
4. else (piped/CI) → `"json"`

**Implementation:**
```go
// internal/output/tty.go
func ResolveFormat(flagValue string) string {
    if flagValue != "" {
        return flagValue  // --output flag wins
    }
    if env := os.Getenv("GOCLAW_OUTPUT"); env != "" {
        return env  // env override
    }
    if IsTTY(os.Stdout) {
        return "table"  // Human-friendly
    }
    return "json"  // Machine-friendly (default piped)
}
```

**Impact on Automation:**
- Scripts using `--output=table` explicitly still work
- Piped output defaults to JSON (machine-readable)
- CI/agents get clean JSON without --output flag

**--quiet Flag:**
- Gates banners/tips in non-automation contexts
- `logs` command banner only shown if TTY + not --quiet

---

### Exit Codes (Phase 0 Locked Contract)

**Exit Code Mapping (0-6):**

| Code | Meaning | When to use | Examples |
|------|---------|------------|----------|
| 0 | Success | Normal completion | All successful commands |
| 1 | Generic error | Unmapped errors | Unknown/unforeseen errors |
| 2 | Auth failure | Auth-related errors | UNAUTHORIZED, NOT_PAIRED, 401/403 HTTP |
| 3 | Not found | Resource not found | NOT_FOUND, NOT_LINKED, 404 HTTP |
| 4 | Validation error | Input/config errors | INVALID_REQUEST, FAILED_PRECONDITION, 400/409/422 HTTP |
| 5 | Server error | Server-side errors | INTERNAL, UNAVAILABLE, AGENT_TIMEOUT, 5xx HTTP |
| 6 | Resource/network | Rate-limit/timeout | RESOURCE_EXHAUSTED, 429 HTTP, connection timeout |

**Implementation (internal/output/exit.go):**
```go
func MapServerCode(code string) int {
    switch code {
    case "UNAUTHORIZED", "NOT_PAIRED", "TENANT_ACCESS_REVOKED":
        return ExitAuth  // 2
    case "NOT_FOUND", "NOT_LINKED":
        return ExitNotFound  // 3
    // ...etc
    case "RESOURCE_EXHAUSTED":
        return ExitResource  // 6
    default:
        return ExitGeneric  // 1
    }
}
```

**AI/Automation Usage:**
- Parse exit codes to determine retry strategy
- Exit 2 → re-authenticate
- Exit 6 → backoff retry (rate-limited)
- Exit 5 → hard failure (server down)

---

### Structured Error Output (Phase 0)

**JSON Error Format:**
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "authentication required",
    "details": {
      "hint": "run 'goclaw auth login' first",
      "retry_after_ms": 0
    }
  }
}
```

**Error Envelope Shape:**
```go
type ErrorEnvelope struct {
    Code       string            `json:"code"`
    Message    string            `json:"message"`
    Details    map[string]any    `json:"details,omitempty"`
    Retryable  bool              `json:"retryable,omitempty"`
    RetryAfter int64             `json:"retry_after_ms,omitempty"`
}
```

**No Double-Printing Contract:**
- Commands return `error` only, no `fmt.Println(err)`
- Central handler in `cmd.Execute()` prints error once
- `output.PrintError(err, format)` handles formatting
- Error bubbles up unchanged through call stack

---

### Destructive Operation Safety Gates (Phase 1+ Systemic)

**Two-Gate Pattern for Destructive Ops:**

Gate 1: Interactive confirmation (if TTY + no --yes)
```go
if !tui.Confirm(msg, cfg.Yes) {
    return nil  // User declined
}
```

Gate 2: Typed confirmation (restore, critical deletes)
```go
// e.g., restore requires both --yes AND --confirm=<filename>
if !cfg.Yes {
    return fmt.Errorf("requires --yes flag")
}
if cfg.Confirm != expectedValue {
    return fmt.Errorf("confirmation mismatch (typed value does not match)")
}
// Only then make API call
```

**Examples:**
- `restore system backup.tar.gz --yes --confirm=backup.tar.gz`
- `vault documents delete <id> --yes`
- `config permissions revoke --agent=X --user=Y` (gated with tui.Confirm)

**CI Behavior:**
- Non-TTY stdin without `--yes` → command refuses (error)
- Non-TTY stdin with `--yes` + correct `--confirm` → proceeds
- Always pre-flight before HTTP call (fast failure)

---

### Streaming & Reconnect (Phase 0 + P2-P4)

**FollowStream Pattern:**
```go
// internal/client/follow.go
func FollowStream(ctx context.Context, serverURL, token, ..., handler FollowHandler) error {
    // Reconnect on drop with exponential backoff: 1s → 2s → 4s → 8s → 16s (max 5 retries)
    // If handler returns error, stop immediately (no retry)
    // If server closes, reconnect and re-send call
    // Respects ctx.Done() for cancellation
}

type FollowHandler func(event *WSEvent) error
```

**Used in Commands:**
- `logs --follow` (real-time log streaming)
- `heartbeat logs --follow` (agent health monitoring)
- `agents wait --timeout=30s` (blocking wait with timeout)
- `teams events` (team event stream, optional --follow)
- `chat` (interactive streaming input/output)

**No RAM Buffering:**
- All uploads use `io.Copy(dst, src)`
- No full-file buffering in memory
- Used in: vault upload, backup download, restore upload

---

### Modularization for Maintainability (Phase 4)

**Per-Group Extraction Pattern:**
- **agents.go** (196 LoC) → Core CRUD only
  - agents_lifecycle.go (172 LoC) → wake/wait/identity
  - agents_admin.go (60 LoC) → admin-only ops
  - agents_sharing.go (96 LoC) → share/unshare
  - agents_instances.go (186 LoC) → per-user instances
  - agents_links.go (127 LoC) → delegation links
  - agents_evolution.go (110 LoC) → evolution feedback
  - agents_episodic.go (80 LoC) → episodic memory
  - agents_v3_flags.go (81 LoC) → feature flags
  - agents_misc.go (62 LoC) → orchestration + pool

- **teams.go** (150 LoC) → Core CRUD only
  - teams_members.go (93 LoC) → membership
  - teams_tasks.go (167 LoC) → core task ops
  - teams_tasks_review.go (98 LoC) → review workflow
  - teams_tasks_advanced.go (200 LoC) → advanced + delete-bulk + events
  - teams_workspace.go (87 LoC) → workspace files
  - teams_events.go (85 LoC) → team events stream
  - teams_scopes.go (41 LoC) → permission scopes

**LoC Compliance:** All files ≤200 LoC (chat files 214 = docstrings only for MAX POLISH)

---

### AI-Critical Commands (MAX POLISH, Phase 4)

**Fully documented, ≥80% test coverage:**
- `chat history` — structured message array via WS
- `chat inject` — context injection without response trigger
- `chat session-status` — session state snapshot
- `agents wait --timeout=30s --state=ready` — blocking wait
- `agents identity` — agent persona/identity retrieval
- `memory kg entities` — full KG entity CRUD + traversal

**Schema Documentation in --help:**
```go
chatHistoryCmd.Long = `Retrieve message history for current session.

Output schema:
{
  "messages": [
    {"role": "user|assistant|system", "content": "...", "timestamp": "..."}
  ]
}
`
```

---

## Version Info

- **Go Version:** 1.25.3+
- **Last Updated:** 2026-04-15
- **Status:** Production Ready (Phases 1-9 + P0-P4 Complete)
- **Phase Status:** P0-P4 ✓ Complete; P5 Deferred
