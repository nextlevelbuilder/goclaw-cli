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

## Version Info

- **Go Version:** 1.25.3+
- **Last Updated:** 2026-03-15
- **Status:** Production Ready
