# GoClaw CLI - Codebase Summary

**Generated from:** `repomix-output.xml` (2026-03-15)
**Total Files:** 49
**Total Tokens:** 56,074
**Total Size:** 193 KB

---

## Overview

GoClaw CLI is a production-ready Go application (1,100+ lines) providing command-line management for GoClaw AI agent gateway servers. Built with Cobra framework, it supports 30 command groups across 23 command files with dual modes: interactive (human) and automation (CI/agent).

**Key Metrics:**
- **23 command files** in `cmd/`
- **7 internal packages** (client, config, output, tui)
- **4 core dependencies** (cobra, websocket, yaml, term)
- **No ORM, no external CLIs** — single statically-linked binary

---

## Directory Structure & File Inventory

### Root Level

| File | Purpose | Size |
|------|---------|------|
| `main.go` | Entry point, calls `cmd.Execute()` | 8 lines |
| `go.mod` | Module definition (Go 1.25.3) | 16 lines |
| `Makefile` | Build automation (build, test, lint, install) | 26 lines |
| `.goreleaser.yaml` | Release config (GoReleaser v2) | 37 lines |
| `README.md` | User documentation | 130 lines |

### cmd/ — Command Definitions (21 files)

All files follow Cobra pattern: root command + subcommands.

#### Core Commands

| File | Commands | LOC | Purpose |
|------|----------|-----|---------|
| `root.go` | `goclaw` (root), global flags | 52 | Root command + persistent flags |
| `auth.go` | `auth`, `credentials` | 180+ | Login, logout, profile mgmt |
| `agents.go` | `agents` (list/get/create/update/delete) | 250+ | Agent CRUD operations |
| `chat.go` | `chat` | 300+ | Interactive + streaming chat |
| `sessions.go` | `sessions` (list/get/delete/reset/label) | 200+ | Session management |
| `skills.go` | `skills` (list/upload/delete) | 200+ | Skill management |
| `mcp.go` | `mcp` (list/add/remove/grants) | 250+ | MCP server management |
| `providers.go` | `providers` (list/create/update/delete) | 200+ | LLM provider mgmt |
| `tools.go` | `tools` (list/invoke/delete) | 180+ | Custom tool management |
| `cron.go` | `cron` (list/create/delete/trigger) | 220+ | Scheduled job management |
| `teams.go` | `teams` (list/create/members) | 270+ | Team management (largest file) |
| `channels.go` | `channels` (list/contacts) | 200+ | Channel management |
| `traces.go` | `traces` (list/export) | 180+ | LLM trace viewing |
| `memory.go` | `memory` (list/search/upsert) | 180+ | Memory document management |
| `config_cmd.go` | `config` (get/apply/patch) | 150+ | Server config management |
| `logs.go` | `logs` | 120+ | Real-time log streaming |
| `storage.go` | `storage` (list/download) | 150+ | Workspace file browser |
| `admin.go` | Admin operations | 250+ | Admin commands (3rd largest) |
| `status.go` | `status` | 80+ | Server health check |
| `version.go` | `version` | 60+ | Version display |
| `api-keys.go` | `api-keys` (list/create/revoke) | 135 | API key management |
| `api-docs.go` | `api-docs` (open/spec) | 82 | API documentation viewer |
| `helpers.go` | Helper functions | 100+ | Shared utilities (newHTTP, unmarshal) |

**Command File Stats:**
- Largest: `teams.go` (4,075 tokens, 7.3%)
- 2nd: `agents.go` (4,048 tokens, 7.2%)
- 3rd: `admin.go` (2,949 tokens, 5.3%)

### internal/ — Private Packages (7 files)

#### client/ — HTTP + WebSocket Clients

```
internal/client/
├── http.go          # REST API client
├── websocket.go     # WebSocket streaming
├── auth.go          # Auth helpers (keyring, device pairing)
└── errors.go        # API error handling
```

**http.go:**
- `HTTPClient` struct: BaseURL, Token, HTTPClient, Verbose
- Methods: `Get()`, `Post()`, `Put()`, `Patch()`, `Delete()`
- Response handling: `apiResponse` struct with OK, Payload, Error
- TLS support: `--insecure` flag disables cert verification
- Timeout: 30 seconds per request
- Returns `json.RawMessage` for deferred unmarshaling

**websocket.go:**
- `WebSocket` struct: conn (*websocket.Conn)
- Methods: `Stream()` for bidirectional communication
- Used by: `chat`, `logs`, `traces` commands
- Context-aware: respects `ctx.Done()` for cancellation

**auth.go:**
- Credential management via OS keyring
- Device pairing flow
- Token validation

**errors.go:**
- `APIError` struct: Code, Message
- Error parsing and user-friendly messages

#### config/ — Configuration Management

```
internal/config/
└── config.go
```

**Features:**
- `Config` struct: Server, Token, OutputFormat, Profile, Insecure, Verbose, Yes
- `Profile` struct: Name, Server, Token, DefaultAgent, OutputFormat
- `FileConfig` struct: ActiveProfile, Profiles
- `Load()` function: Implements precedence: flags > env > file > defaults
- `Dir()`: Returns ~/.goclaw/
- `FilePath()`: Returns ~/.goclaw/config.yaml
- Multi-profile support with `FindProfile()`
- Environment variables: GOCLAW_SERVER, GOCLAW_TOKEN, GOCLAW_OUTPUT

#### output/ — Output Formatting

```
internal/output/
└── output.go
```

**Formats:**
- `table`: Human-readable tables (default)
- `json`: Compact JSON for machines
- `yaml`: Configuration-friendly format

**Printer struct:**
- Methods: `Print()`, `PrintTable()`, `PrintJSON()`, `PrintYAML()`

**Table struct:**
- Headers, rows, alignment
- Used by all list commands for human output

#### tui/ — Terminal UI

```
internal/tui/
└── prompt.go
```

**Features:**
- Interactive prompts for user input
- TUI integration with `golang.org/x/term`
- Raw mode for streaming (chat, logs)

---

## Dependencies

### go.mod Analysis

```
go 1.25.3

require (
	github.com/gorilla/websocket v1.5.3
	github.com/spf13/cobra v1.10.2
	golang.org/x/term v0.41.0
	gopkg.in/yaml.v3 v3.0.1
)

indirect (
	github.com/inconshreveable/mousetrap v1.1.0
	github.com/spf13/pflag v1.0.9
	golang.org/x/sys v0.42.0
)
```

**Core Dependencies:**

| Package | Version | Purpose |
|---------|---------|---------|
| `cobra` | v1.10.2 | CLI framework (commands, flags, help) |
| `gorilla/websocket` | v1.5.3 | WebSocket streaming client |
| `golang.org/x/term` | v0.41.0 | Terminal utilities (raw mode, prompt) |
| `yaml.v3` | v3.0.1 | YAML parsing/serialization |

**Why No ORM?**
- HTTP API is the primary interface to GoClaw
- No database layer needed in CLI
- Keeps binary size small (~8 MB)

**Why No External CLIs?**
- Statically-linked Go binary
- No shell or system dependencies
- Easy to distribute and install

---

## Largest Files by Complexity

| Rank | File | Tokens | % | Reason |
|------|------|--------|---|--------|
| 1 | `cmd/teams.go` | 4,075 | 7.3% | Complex team operations |
| 2 | `cmd/agents.go` | 4,048 | 7.2% | Full CRUD + sharing |
| 3 | `cmd/admin.go` | 2,949 | 5.3% | Multi-operation admin |
| 4 | `cmd/mcp.go` | 2,940 | 5.2% | MCP grants + access |
| 5 | `cmd/skills.go` | 2,635 | 4.7% | Skill upload + mgmt |

**Total cmd/ code:** ~47,000 tokens (84% of codebase)
**Total internal/ code:** ~9,000 tokens (16% of codebase)

---

## Command Hierarchy

```
goclaw (root)
├── auth
│   ├── login
│   ├── logout
│   └── use-context
├── credentials
│   ├── get
│   └── set
├── agents
│   ├── list
│   ├── get
│   ├── create
│   ├── update
│   ├── delete
│   ├── share
│   └── delegation-link
├── api-keys (list, create, revoke)
├── api-docs (open, spec)
├── chat
├── sessions (list, get, delete, reset, label)
├── skills (list, upload, delete)
├── mcp (list, add, remove, grants, access-requests)
├── providers (list, create, update, delete, models)
├── tools (list, invoke, delete)
├── cron (list, create, update, delete, trigger, history)
├── teams (list, create, members, task-board, workspace)
├── channels (list, contacts, pending-messages)
├── traces (list, export)
├── memory (list, search, upsert)
├── knowledge-graph (entities, links, query)
├── usage (summary, cost-breakdown)
├── config (get, apply, patch)
├── logs
├── storage (list, download)
├── approvals (list, approve, deny)
├── delegations
├── tts (synthesize, list-voices)
├── media (upload, download)
├── activity
├── status
└── version
```

**Total: 30 command groups**

---

## Key Patterns & Conventions

### Error Handling
```go
if err != nil {
	return fmt.Errorf("operation: %w", err)
}
```
All errors wrapped with context.

### Command Structure
```go
var myCmd = &cobra.Command{
	Use:   "command",
	Short: "Description",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Implementation
		return nil
	},
}
```

### HTTP Client Usage
```go
c, err := newHTTP()          // Helper function (helpers.go)
data, err := c.Get("/path")  // Returns json.RawMessage
unmarshaled := unmarshalList(data)  // Parse JSON
```

### Output Formatting
```go
if cfg.OutputFormat == "table" {
	tbl := output.NewTable("Col1", "Col2")
	tbl.AddRow("val1", "val2")
	printer.Print(tbl)
} else {
	printer.Print(unmarshalMap(data))  // JSON/YAML
}
```

### WebSocket Streaming
```go
ws, _ := client.NewWebSocket(cfg.Server, cfg.Token, "/path")
ws.Stream(ctx, func(msg []byte) error {
	// Process message
	return nil
})
```

---

## CI/CD & Build

### GitHub Actions

**ci.yaml** (triggered on push/PR to main):
- Go 1.25
- Build: `go build ./...`
- Vet: `go vet ./...`
- Test: `go test -race ./...`

**release.yaml** (triggered on tag):
- Uses GoReleaser v2
- Builds: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64, windows/arm64
- Generates: tar.gz (Unix), zip (Windows), checksums.txt

### Local Build

```bash
make build      # Binary: ./goclaw
make test       # Run tests with race detector
make lint       # go vet ./...
make install    # Install to GOPATH/bin
make clean      # Remove binaries
```

---

## Configuration Hierarchy

1. **Defaults:** `OutputFormat: "table"`
2. **Config File:** `~/.goclaw/config.yaml`
3. **Environment:** `GOCLAW_SERVER`, `GOCLAW_TOKEN`, `GOCLAW_OUTPUT`
4. **CLI Flags:** `--server`, `--token`, `--output`, `--yes`, `--profile`

Each level overrides the previous.

---

## Security Model

- **Credentials:** OS keyring integration (not in config file)
- **TLS:** Enabled by default, `--insecure` only for testing
- **Token:** Accept via `GOCLAW_TOKEN` env var, not CLI args
- **Logging:** Never log sensitive data
- **Process Security:** No credentials visible in `ps` output

---

## Testing Strategy

**Current Status:** In Progress (Phases 1-9 complete, testing is Phase 10)

**Planned Approach:**
- Table-driven tests for all commands
- Mock HTTP responses for unit tests
- Integration tests for critical paths (auth, chat, agents)
- Race detector: `go test -race ./...`
- Target: >80% coverage

---

## Notable Implementation Details

### No Global State
- Commands create client instances (`newHTTP()`) on demand
- Config loaded per-command via `PersistentPreRunE`
- Printer instance created per-root execution

### Deferred JSON Parsing
- HTTP client returns `json.RawMessage`
- Commands unmarshal only what's needed
- Saves memory for large responses

### Profile Management
- Multiple profiles in `~/.goclaw/config.yaml`
- Set active via `goclaw auth use-context <profile>`
- Override per-command: `goclaw --profile staging agents list`

### Automation Mode
- Flags: `--yes` (skip prompts), `--output json` (machine output), `--verbose` (debug)
- Perfect for CI/CD, AI agents, scripts
- Environment variables eliminate token in command history

---

## File Statistics

| Category | Count | Est. LOC |
|----------|-------|---------|
| Command files | 23 | 2,717+ |
| Internal packages | 7 | 600+ |
| Build/CI configs | 3 | 80+ |
| Docs | 5 | 600+ |
| **Total** | **38** | **3,997+** |

---

## Design Principles

1. **YAGNI:** Features requested by GoClaw dashboard, no speculative additions
2. **KISS:** Cobra + raw HTTP, no complex frameworks
3. **DRY:** Shared helpers in `helpers.go`, reusable client methods
4. **Security-First:** Keyring by default, TLS required, no plaintext secrets
5. **Dual Mode:** Works for humans (interactive) and machines (automation)

---

## Known Limitations & Future Work

**Phase 1-9 (Complete):**
- 28 command groups
- Full API coverage
- Dual mode (interactive + automation)
- Multi-profile support
- WebSocket streaming

**Phase 10+ (Future):**
- Unit test coverage >80%
- Integration tests
- Shell completion scripts (bash, zsh, fish)
- Homebrew tap
- Man pages

---

## Last Updated

- **Date:** 2026-03-15
- **Status:** Production Ready
- **Phases Complete:** 1-9 (All feature implementation)
- **Next Focus:** Testing & completion (Phase 10+)
