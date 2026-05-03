# GoClaw CLI - Codebase Summary

**Generated from:** `repomix-output.xml` (2026-04-15)
**Phase Status:** P0-P4 Complete (AI-First Expansion)
**Total Files:** 65+
**Estimated Tokens:** 72,000+
**Total Size:** 220+ KB

---

## Overview

GoClaw CLI is a production-ready Go application (3,500+ lines) providing comprehensive command-line management for GoClaw AI agent gateway servers. Built with Cobra framework, it supports 30 command groups across 50+ command files with dual modes: interactive (human) and automation (CI/agent). Phases 0-4 (AI-first expansion) add AI ergonomics, admin/ops, migration, vault, and advanced agent/team/memory support.

**Key Metrics:**
- **50+ command files** in `cmd/` (modularized for maintainability)
- **7 internal packages** (client, config, output, tui) with Phase 0 AI additions
- **4 core dependencies** (cobra, websocket, yaml, term)
- **No ORM, no external CLIs** — single statically-linked binary
- **Phase 0 locked contracts:** Exit codes 0-6, TTY auto-detect, central error handler, FollowStream reconnect

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

### cmd/ — Command Definitions (26 files)

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
| `config_cmd.go` | `config` (get/apply/patch/permissions) | 230+ | Server config + permissions |
| `logs.go` | `logs` | 120+ | Real-time log streaming |
| `storage.go` | `storage` (list/download) | 150+ | Workspace file browser |
| `admin.go` | Admin operations | 250+ | Admin commands |
| `status.go` | `status` | 80+ | Server health check |
| `version.go` | `version` | 60+ | Version display |
| `api_keys.go` | `api-keys` (list/create/revoke) | 135 | API key management |
| `api_docs.go` | `api-docs` (open/spec) | 82 | API documentation viewer |
| `backup.go` | `backup system/tenant` (create, preflight, download) | 210 | System + tenant backup commands |
| `backup_s3.go` | `backup s3` (config get/set, list, upload, backup) | 155 | S3 backup integration |
| `restore.go` | `restore system/tenant` (with typed confirmation) | 115 | Destructive restore with safety guards |
| `agents_export.go` | `agents export/import/import-merge` | 100 | Agent export/import subcommands |
| `teams_export.go` | `teams export/import` | 75 | Team export/import subcommands |
| `skills_export.go` | `skills export/import` | 75 | Skills export/import subcommands |
| `mcp_export.go` | `mcp export/import` | 75 | MCP export/import subcommands |
| `io_helpers.go` | Shared I/O utilities (copy, progress, file helpers) | 45 | Streaming file I/O helpers |
| `tenants.go` | `tenants` (list/get/create/update/mine/users) | 200 | Tenant CRUD + membership (HTTP) |
| `heartbeat.go` | `heartbeat` (get/set/toggle/test/targets/logs) | 190 | Agent heartbeat monitoring (WS) |
| `heartbeat_checklist.go` | `heartbeat checklist` (get/set) | 75 | Heartbeat checklist split file (WS) |
| `system_configs.go` | `system-configs` (list/get/set/delete) | 120 | System key-value config (HTTP) |
| `edition.go` | `edition` | 35 | Server edition info, no auth (HTTP) |
| `helpers.go` | Helper functions | 100+ | Shared utilities (newHTTP, unmarshal) |

**Command File Stats:**
- Largest: `teams.go` (4,075 tokens, 7.3%)
- 2nd: `agents.go` (4,048 tokens, 7.2%)
- 3rd: `admin.go` (2,949 tokens, 5.3%)

### internal/ — Private Packages (7 files)

#### client/ — HTTP + WebSocket Clients

```
internal/client/
├── http.go                # REST API client
├── websocket.go           # WebSocket streaming
├── auth.go                # Auth helpers (keyring, device pairing)
├── errors.go              # APIError struct (matches server ErrorShape) + interface methods
├── follow.go              # FollowStream() — persistent streaming with exponential backoff
├── signed_download.go     # DownloadSigned() — unauthenticated binary download (signed token flow)
└── multipart_upload.go    # UploadFile() / DrainResponse() — streaming multipart POST
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
- `APIError` struct: Code, Message, Details, Retryable, RetryAfterMs, StatusCode
- Implements `output.apiErrorIface` + `output.apiErrorWithStatus` (duck-typed, no import cycle)
- Sentinel errors: `ErrNotAuthenticated`, `ErrServerRequired`

**follow.go:**
- `FollowStream(ctx, serverURL, token, ...)` — reconnects on drop with exponential backoff (1s→2s→4s→8s→16s, max 5 retries)
- `FollowHandler func(*WSEvent) error` — returning non-nil stops stream

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

#### output/ — Output Formatting + Error Handling

```
internal/output/
├── output.go      # Printer, TableData — format-agnostic output
├── exit.go        # Exit code constants (0-6) + MapServerCode/MapHTTPStatus/Exit
├── error.go       # ErrorDetail/ErrorEnvelope, ParseHTTPError, PrintError, FromError
└── tty.go         # IsTTY(fd), ResolveFormat(flag) with TTY auto-detection
```

**Formats:**
- `table`: Human-readable tables (when stdout is a TTY)
- `json`: Compact JSON for machines (default when piped/CI)
- `yaml`: Configuration-friendly format

**TTY-aware format resolution (precedence):**
1. `--output` flag (explicit)
2. `GOCLAW_OUTPUT` environment variable
3. stdout is TTY → `"table"`
4. else → `"json"`

**Printer struct:**
- Methods: `Print()`, `Error()`, `Success()`

**Table struct:**
- Headers, rows, alignment
- Used by all list commands for human output

**Exit codes (locked contract for AI/automation consumers):**
| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Generic/unknown |
| 2 | Auth failure |
| 3 | Not found |
| 4 | Validation error |
| 5 | Server error |
| 6 | Resource/network |

**Error format (JSON mode):**
```json
{"error": {"code": "UNAUTHORIZED", "message": "..."}}
```

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
│   ├── list / get / create / update / delete
│   ├── share / unshare / regenerate / resummon / cancel-summon
│   ├── delegation-link
│   ├── export / import / import-merge
│   ├── files (list, get, set)            # global AGENTS.md, SOUL.md, IDENTITY.md, ...
│   ├── instances (list, get-file, set-file, metadata, update-metadata)
│   ├── episodic (list, search)
│   ├── evolution (metrics, suggestions, update)
│   ├── orchestration / codex-pool-activity
│   ├── skills list                       # skills granted to agent
│   ├── v3-flags (get, toggle)
│   ├── wake / wait / identity
│   └── prompt-preview
├── api-keys (list, create, revoke)
├── api-docs (open, spec)
├── backup
│   ├── system [--wait --file]
│   ├── system-preflight
│   ├── system-download <token> --file
│   ├── tenant [--tenant-id]
│   ├── tenant-preflight
│   ├── tenant-download <token> --file
│   └── s3
│       ├── config get [--show-secret]
│       ├── config set --bucket --access-key --secret-key
│       ├── list
│       ├── upload <token>
│       └── backup
├── restore
│   ├── system <file> --yes --confirm=<basename>
│   └── tenant <file> --tenant-id --yes --confirm=<tenantID>
├── chat
├── sessions (list, get, delete, reset, label)
├── skills (list, upload, delete, export, import [--apply])
├── mcp (list, add, remove, grants, access-requests, export, import [--apply])
├── providers (list, create, update, delete, models)
├── tools (list, invoke, delete)
├── cron (list, create, update, delete, trigger, history)
├── teams (list, create, members, task-board, export, import [--apply])
│   └── workspace (list, read, delete, upload, move)
├── channels (list, contacts, pending-messages)
├── traces (list, export)
├── memory (list, search, upsert)
├── knowledge-graph (entities, links, query)
├── usage (summary, detail, costs, timeseries, breakdown)
├── costs (server-side cost summary; alias under traces.go)
├── files (sign)                                # signed URL helper
├── voices (list, refresh)                      # voice catalog
├── hooks (list, create, update, delete, toggle, test, history)
├── tenants (list, get, create, update, mine, users list/add/remove)
├── heartbeat (get, set, toggle, test, targets, logs, checklist get/set)
├── system-configs (list, get, set, delete)
├── edition
├── config (get, apply, patch, permissions list/grant/revoke)
├── logs
├── storage (list, download)
├── approvals (list, approve, deny)
├── delegations
├── tts (status, enable, disable, providers, set-provider, test-connection)
├── media (upload, download)
├── activity
├── status
└── version
```

**Total: ~40 command groups** (after AI-first expansion P0–P2 reaching CLI parity with server admin surface)

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

## Migration System (Phase 2)

### Backup/Restore Safety

- **Signed token download flow:** Server returns token → CLI downloads with auth (no signed URL needed per implementation)
- **Restore confirmation:** Two-factor guard — `--yes` (intent) + `--confirm=<value>` (typed match). Mismatch = immediate refusal, no server call made.
- **S3 secret masking:** `backup s3 config get` always masks `secret_key` as `***` unless `--show-secret` is passed
- **Preview-first imports:** All `import` subcommands default to preview endpoint. `--apply` required to mutate data.
- **Streaming:** All downloads/uploads use `io.Copy` streaming — no full-file buffering

### New Internal Packages

| File | Purpose |
|------|---------|
| `internal/client/signed_download.go` | `DownloadSigned(url, dst, insecure, progress)` — GET with NO auth header |
| `internal/client/multipart_upload.go` | `UploadFile(path, field, file)` — streaming pipe-based multipart POST |
| `cmd/io_helpers.go` | `copyProgress`, `writeToFile`, `printProgress` shared helpers |

---

## Vault Subsystem (Phase 3)

### Command Files

| File | Commands | LOC | Purpose |
|------|----------|-----|---------|
| `cmd/vault.go` | `vault tree/search/rescan/graph` | ~150 | Root group + simple queries, DOT transform |
| `cmd/vault_documents.go` | `vault documents list/get/create/update/delete/links` | ~200 | Document CRUD + link listing |
| `cmd/vault_links.go` | `vault links create/delete/batch-get` | ~100 | Link management |
| `cmd/vault_upload.go` | `vault upload <file>` | ~95 | Streaming multipart file upload |
| `cmd/vault_enrichment.go` | `vault enrichment status/stop` | ~55 | Enrichment pipeline control |
| `cmd/vault_multipart_helper.go` | Internal helper | ~50 | Multipart writer for vault uploads |

### New Internal Files

| File | Purpose |
|------|---------|
| `internal/output/tree.go` | `TreeNode`, `PrintTree`, `PrintTreeRoot`, `VaultEntriesToTreeNode` — ASCII tree rendering |

### Endpoints Covered

| Method | Path | Command |
|--------|------|---------|
| `GET` | `/v1/vault/documents` | `vault documents list` |
| `GET` | `/v1/vault/documents/{id}` | `vault documents get` |
| `POST` | `/v1/vault/documents` | `vault documents create` |
| `PUT` | `/v1/vault/documents/{id}` | `vault documents update` |
| `DELETE` | `/v1/vault/documents/{id}` | `vault documents delete` |
| `GET` | `/v1/vault/documents/{id}/links` | `vault documents links` |
| `POST` | `/v1/vault/links` | `vault links create` |
| `DELETE` | `/v1/vault/links/{id}` | `vault links delete` |
| `POST` | `/v1/vault/links/batch` | `vault links batch-get` |
| `POST` | `/v1/vault/upload` | `vault upload` |
| `POST` | `/v1/vault/rescan` | `vault rescan` |
| `GET` | `/v1/vault/tree` | `vault tree` |
| `POST` | `/v1/vault/search` | `vault search` |
| `GET` | `/v1/vault/enrichment/status` | `vault enrichment status` |
| `POST` | `/v1/vault/enrichment/stop` | `vault enrichment stop` |
| `GET` | `/v1/vault/graph` | `vault graph` |

### Key Design Decisions

- **DOT format:** `graphJSONToDOT()` transforms `{nodes, edges}` JSON → Graphviz DOT; skips edges with empty IDs.
- **Tree rendering:** `internal/output/tree.go` with `PrintTreeRoot` for TTY, JSON pass-through otherwise.
- **Upload:** Pipe-based streaming via `io.Pipe` + goroutine; `contentType()` captured before goroutine to avoid races.
- **Destructive ops:** `documents delete`, `links delete`, `enrichment stop`, `rescan` all require `--yes`.
- **Search default:** `--limit=20` (matches server default `max_results`).

### Command Hierarchy (Vault)

```
goclaw vault
├── documents
│   ├── list [--q] [--limit] [--offset]
│   ├── get <docID>
│   ├── create --title --path [--doc-type] [--scope] [--content|--file]
│   ├── update <docID> [--title] [--doc-type] [--scope]
│   ├── delete <docID> --yes
│   └── links <docID>
├── links
│   ├── create --from --to [--type]
│   ├── delete <linkID> --yes
│   └── batch-get <docID> [docID...]
├── upload <file> [--title] [--tags]
├── rescan --yes
├── tree [--path]
├── search <query> [--limit] [--offset]
├── enrichment
│   ├── status
│   └── stop --yes
└── graph [--format=json|dot]
```

---

---

## Phase 4 — Agent Lifecycle + Chat + Teams + Memory KG

### New Command Files

| File | Commands | Purpose |
|------|----------|---------|
| `agents_lifecycle.go` | `agents wake/wait/identity` | AI-critical: blocking wait + identity (MAX POLISH) |
| `agents_admin.go` | `agents sync-workspace/prompt-preview` | Admin-only agent ops |
| `agents_sharing.go` | `agents share/unshare/regenerate/resummon` | Agent sharing lifecycle |
| `agents_instances.go` | `agents instances list/get-file/set-file/update-metadata/metadata` | Per-user instance management |
| `agents_links.go` | `agents links list/create/update/delete` | Delegation link management |
| `agents_evolution.go` | `agents evolution metrics/suggestions/update` | Evolution feedback loop |
| `agents_episodic.go` | `agents episodic list/search` | Episodic memory (semantic search) |
| `agents_v3_flags.go` | `agents v3-flags get/toggle` | Experimental feature flags |
| `agents_misc.go` | `agents orchestration/codex-pool-activity` | Orchestration + pool status |
| `chat_ai_commands.go` | `chat history/inject/session-status` | AI-critical MAX POLISH chat ops |
| `teams_members.go` | `teams members list/add/remove` | Team membership |
| `teams_tasks.go` | `teams tasks list/get/get-light/create/assign` | Core task CRUD |
| `teams_tasks_review.go` | `teams tasks approve/reject/comment/comments` | Task review workflow |
| `teams_tasks_advanced.go` | `teams tasks delete/delete-bulk/events/active` | Advanced task ops + follow stream |
| `teams_workspace.go` | `teams workspace list/read/delete` | Team workspace files |
| `teams_events.go` | `teams events list [--follow]` | Team event stream |
| `teams_scopes.go` | `teams scopes <teamID>` | Permission scopes |
| `memory_kg.go` | `memory kg entities list/get/upsert/delete` | KG entity CRUD |
| `memory_kg_graph.go` | `memory kg traverse/stats/graph [--compact]` | KG traversal + graph |
| `memory_kg_dedup.go` | `memory kg dedup scan/list/merge/dismiss` | KG deduplication |
| `memory_kg_legacy.go` | `memory kg query/extract/link` | Legacy KG compat wrappers |
| `memory_index.go` | `memory chunks/index/index-all/documents-global` | Memory indexing + global docs |

### AI-Critical Commands (MAX POLISH)

All have JSON schema in `--help`, full validation, ≥80% test coverage:
- `chat history` — retrieves structured message array via WS `chat.history`
- `chat inject` — injects context without triggering response (admin-only, role validation)
- `chat session-status` — current session state via WS `chat.session.status`
- `agents wait` — blocking WS call with `--timeout` + `--state`; exits 6 on timeout
- `agents identity` — identity/persona retrieval via WS `agent.identity.get`
- `memory kg` full suite — entities CRUD + traverse + stats + graph + dedup

### Modularization Results (LoC compliance)

All `cmd/` files now ≤200 LoC (chat files are 214 lines — overage is entirely docstrings for AI-critical help text).

### New Test Files

| File | Tests | Coverage |
|------|-------|---------|
| `agents_lifecycle_test.go` | 18 tests | wake, identity, wait (success+timeout+invalid), sync, preview, evolution, episodic, v3-flags, orchestration, codex, instances |
| `chat_extensions_test.go` | 11 tests | history (3), inject (5 inc. validation), session-status (2) |
| `teams_tasks_test.go` | 16 tests | list, get, get-light, create, assign, delete (yes+declined), delete-bulk (ids+missing), events, active (success+missing), scopes, events-list |
| `memory_kg_test.go` | 15 tests | entities (list/get/delete/delete-with-yes), traverse (from-required+success), stats, graph (full+compact), dedup (scan/list/merge/dismiss), chunks, index, index-all, documents-global |

### Exit Codes (Phase 4 additions)
- `agents wait` timeout → exit 6 (`output.ExitResource`)

## Last Updated

- **Date:** 2026-04-15
- **Status:** Production Ready
- **Phases Complete:** Phase 2 (Backup/Restore), Phase 3 (Vault), Phase 4 (Agents+Chat+Teams+Memory KG)
- **Total cmd/ files:** 50+
- **Total tests:** 146 passing
