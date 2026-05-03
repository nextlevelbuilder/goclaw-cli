# GoClaw CLI - System Architecture

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    GoClaw CLI (User)                        │
│                                                             │
│  Interactive Terminal / CI Script / AI Agent               │
└────────────────────────┬────────────────────────────────────┘
                         │
         ┌───────────────┴────────────────┐
         │                                │
    ┌────▼──────┐                  ┌──────▼────┐
    │  REST API │                  │ WebSocket │
    │ (HTTP)    │                  │ (WS)      │
    └────┬──────┘                  └──────┬────┘
         │                                │
         │    HTTP GET/POST/PATCH         │  WS Connect
         │    /v1/agents, /v1/sessions    │  /v1/chat/stream
         │                                │  /v1/logs/stream
         └────────────────┬───────────────┘
                          │
                    ┌─────▼────────┐
                    │  GoClaw      │
                    │  Server      │
                    │              │
                    │  Port 8080   │
                    └──────────────┘
```

---

## Component Architecture

### 1. CLI Entry Point (main.go)

```
main()
  └─> cmd.Execute()
      └─> rootCmd.Execute()
          ├─> PersistentPreRunE: Load config (file→env→flags)
          ├─> Create Printer (table/json/yaml)
          └─> Dispatch to Command
```

**Flow:**
1. Load configuration (precedence: flags > env > file > defaults)
2. Initialize output printer
3. Execute command handler

### 2. Command Layer (cmd/ - 91 files, modularized)

**Pattern: Cobra Command Structure**

```go
rootCmd (goclaw)
  ├─ PersistentFlags: --server, --token, --output, --yes, --verbose, --insecure, --profile
  ├─ PersistentPreRunE: Load config, create printer
  └─ Subcommands (~40 groups)
      ├─ auth, credentials, api-keys, api-docs
      ├─ agents (CRUD + files + instances + episodic + evolution + links + skills + v3-flags + ...)
      ├─ chat, sessions, skills, mcp, providers, tools, cron
      ├─ teams (+ workspace upload/move + tasks + events + scopes)
      ├─ channels (contacts, instances, pending, writers)
      ├─ hooks (list/create/update/delete/toggle/test/history)  # event interception
      ├─ vault (documents, links, search, graph, enrichment, upload)
      ├─ memory (kg entities/dedup/graph/extract; index; chunks)
      ├─ usage (summary/detail/costs/timeseries/breakdown)
      ├─ traces, costs, files (sign), voices (list/refresh)
      ├─ tts, media, packages (+ github-releases)
      ├─ tenants, users, system-configs, oauth, pair, send, quota
      ├─ backup, restore, storage, logs, status, edition, version
      └─ approvals, delegations, activity, admin-credentials, heartbeat
```

**Command Handler Pattern:**

```go
var myCmd = &cobra.Command{
	Use:   "subcommand",
	Short: "Description",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Validate arguments
		// 2. Create HTTP client
		// 3. Make API call
		// 4. Format output
		// 5. Return error or nil
	},
}
```

**Global Variables (initialized by root.go):**
- `cfg *config.Config` — Loaded config
- `printer *output.Printer` — Output formatter

---

### 3. Client Layer (internal/client/)

#### HTTP Client (http.go)

```
HTTPClient
  ├─ BaseURL: string (e.g., "https://goclaw.example.com")
  ├─ Token: string (Auth bearer token)
  ├─ HTTPClient: *http.Client (Conn pool, timeout=30s)
  └─ Verbose: bool

Methods:
  ├─ Get(path) -> json.RawMessage
  ├─ Post(path, body) -> json.RawMessage
  ├─ Put(path, body) -> json.RawMessage
  ├─ Patch(path, body) -> json.RawMessage
  └─ Delete(path) -> json.RawMessage
```

**Request/Response Format:**

```
Request:
  GET /v1/agents HTTP/1.1
  Host: goclaw.example.com
  Authorization: Bearer {token}
  Content-Type: application/json

Response:
  HTTP/1.1 200 OK
  {
    "ok": true,
    "payload": [
      { "id": "abc123", "name": "Agent1", ... }
    ]
  }
```

**Error Handling:**

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

#### WebSocket Client (websocket.go)

```
WebSocket
  ├─ conn: *websocket.Conn
  └─ Methods:
      ├─ Stream(ctx, fn) -> error
      └─ Close()
```

**Usage Pattern:**

```
WebSocket Stream (Bidirectional)
  │
  ├─> Connect to /v1/chat/stream
  ├─> Client sends: { "message": "Hello" }
  ├─> Server streams: { "type": "message", "text": "..." }
  ├─> Client streams: { "type": "message", "text": "..." }
  └─> Client sends EOF or timeout
```

**Commands Using WebSocket:**
- `goclaw chat` (interactive mode)
- `goclaw logs` (real-time tailing with -f)
- `goclaw traces` (live trace streaming)

#### Authentication (auth.go)

**Credential Storage:**
1. OS Keyring (preferred)
   - macOS: Keychain
   - Linux: Secret Service / pass
   - Windows: Credential Manager
2. Fallback: ~/.goclaw/credentials (encrypted)

**Device Pairing Flow:**
```
goclaw auth login --pair
  │
  ├─> Show pairing code (e.g., "ABC123")
  ├─> User visits https://goclaw.example.com/pair?code=ABC123
  ├─> Server validates and generates token
  ├─> Poll server for token completion
  └─> Store token in keyring
```

---

### 4. Configuration Management (internal/config/)

**Precedence (Highest to Lowest):**

```
1. CLI Flags
   goclaw --server https://custom.com agents list

2. Environment Variables
   export GOCLAW_SERVER=https://custom.com
   goclaw agents list

3. Config File (~/.goclaw/config.yaml)
   active_profile: production
   profiles:
     - name: production
       server: https://goclaw.example.com
       token: {in-keyring}

4. Defaults
   OutputFormat: "table"
```

**Config Struct:**

```go
type Config struct {
	Server       string
	Token        string
	OutputFormat string  // "table", "json", "yaml"
	Profile      string
	Insecure     bool    // Skip TLS cert check (testing only)
	Verbose      bool    // Debug logging
	Yes          bool    // Skip confirmation prompts
}

type Profile struct {
	Name         string
	Server       string
	Token        string
	DefaultAgent string
	OutputFormat string
}
```

**Load Flow:**

```
Load(cmd *cobra.Command) -> Config
  │
  ├─> 1. Read ~/.goclaw/config.yaml
  ├─> 2. Merge active profile or specified profile
  ├─> 3. Overlay env vars: GOCLAW_SERVER, GOCLAW_TOKEN, GOCLAW_OUTPUT
  ├─> 4. Overlay CLI flags (only if explicitly set)
  └─> Return merged Config
```

---

### 5. Output Formatting & Error Handling (internal/output/)

**Package Structure (Phase 0):**

```
internal/output/
├── output.go      # Printer, Table, format dispatching
├── exit.go        # ExitCode constants (0-6) + mappers
├── error.go       # ErrorDetail, PrintError, FromError
└── tty.go         # IsTTY, ResolveFormat (TTY auto-detect)
```

**Exit Codes (locked contract for AI/automation consumers):**

```go
const (
    ExitSuccess     = 0  // Success
    ExitGeneric     = 1  // Unknown/unmapped
    ExitAuth        = 2  // UNAUTHORIZED, NOT_PAIRED, etc.
    ExitNotFound    = 3  // NOT_FOUND, HTTP 404
    ExitValidation  = 4  // INVALID_REQUEST, HTTP 400/422
    ExitServer      = 5  // INTERNAL, UNAVAILABLE, HTTP 5xx
    ExitResource    = 6  // RESOURCE_EXHAUSTED, HTTP 429, timeouts
)
```

Maps 12 known server error codes to exit codes; HTTP status fallback for envelope-less responses.

**Error Output Shape (JSON mode):**

```json
{"error": {"code": "UNAUTHORIZED", "message": "...", "details": {...}}}
```

**TTY-Aware Format Resolution (precedence):**
1. `--output` flag (explicit)
2. `GOCLAW_OUTPUT` env var
3. stdout is TTY → `"table"`
4. else → `"json"`

**Printer Interface:**

```go
type Printer struct {
	Format string
}

func (p *Printer) Print(data any)   // Dispatch to table/json/yaml
func (p *Printer) Error(err error)  // Format error for output
func (p *Printer) Success(msg string) // Print success message
```

**Output Examples:**

```bash
# Table (default, human-readable)
goclaw agents list
│ ID       │ NAME    │ PROVIDER │ MODEL     │ STATUS │
├──────────┼─────────┼──────────┼───────────┼────────┤
│ abc123   │ Agent1  │ openai   │ gpt-4     │ active │

# JSON (automation)
goclaw agents list -o json
[{"id":"abc123","name":"Agent1","provider":"openai"}]

# YAML (config friendly)
goclaw agents list -o yaml
- id: abc123
  name: Agent1
  provider: openai
```

---

### 6. Terminal UI (internal/tui/)

**Interactive Features:**

```go
// Prompt for user input
response, err := tui.Prompt("Enter agent name: ")

// Raw mode (chat streaming)
tui.RawMode(func() {
	// Read terminal input character by character
	// Print server responses without buffering
})
```

**Used By:**
- `goclaw auth login` (credential input)
- `goclaw chat` (interactive mode)
- Confirmation prompts (unless `--yes` is set)

---

## Data Flow Examples

### Example 1: List Agents

```
User Command:
  $ goclaw agents list -o json

Flow:
  1. Cobra: Parse command "agents list", flag "--output json"
  2. PersistentPreRunE: Load config, create printer with format=json
  3. agentsListCmd.RunE():
     a. Create HTTP client: c := newHTTP()
     b. GET /v1/agents: data, err := c.Get("/v1/agents")
     c. Parse response: agents := unmarshalList(data)
     d. Output: printer.Print(agents)

Response Sequence:
  HTTP: GET /v1/agents + Bearer token
        ↓
  Server: Return 200 OK with agent list
        ↓
  Client: Parse JSON into []map[string]any
        ↓
  Printer: Marshal to JSON
        ↓
  Output: Print to stdout as JSON
```

### Example 2: Interactive Chat

```
User Command:
  $ goclaw chat myagent

Flow:
  1. Cobra: Parse command "chat", arg "myagent"
  2. PersistentPreRunE: Load config, create printer
  3. chatCmd.RunE():
     a. Resolve agent ID from name "myagent"
     b. Create WebSocket: ws := NewWebSocket("/v1/chat/stream?agent=abc123")
     c. Enter raw mode (terminal)
     d. Loop:
        ├─ Read stdin
        ├─ Send via WebSocket
        ├─ Receive from WebSocket
        ├─ Print to stdout (streaming)
        └─ Repeat until EOF/exit

Message Flow:
  User (stdin) → Marshal to JSON
                  ↓
              Send via WebSocket
                  ↓
              GoClaw Server
                  ↓
              Receive response (streaming)
                  ↓
              Parse JSON
                  ↓
              Print to stdout (real-time)
```

### Example 3: Create Agent (Automation)

```
Script:
  $ curl -H "Authorization: Bearer $TOKEN" \
      -d '{"name":"Bot1","provider":"openai","model":"gpt-4"}' \
      https://goclaw.example.com/v1/agents | \
    goclaw --token $TOKEN agents create --name Bot1 --provider openai --model gpt-4 -y -o json

Flow:
  1. Parse flags: --name, --provider, --model
  2. Set yes=true (skip confirmation)
  3. Set output=json
  4. Create HTTP client with token from env/flags
  5. POST /v1/agents with body: {"name":"Bot1",...}
  6. Receive response
  7. Output JSON (no confirmation prompt due to --yes)
  8. Exit with status code

Response:
  {
    "id": "xyz789",
    "name": "Bot1",
    "provider": "openai",
    "model": "gpt-4",
    "status": "active"
  }
```

---

## Configuration Precedence (Detailed)

### Scenario 1: Default Profile

```bash
$ cat ~/.goclaw/config.yaml
active_profile: production
profiles:
  - name: production
    server: https://goclaw.example.com

$ goclaw agents list
# Uses: https://goclaw.example.com
```

### Scenario 2: Environment Override

```bash
$ export GOCLAW_SERVER=https://staging.example.com
$ goclaw agents list
# Uses: https://staging.example.com (env overrides config)
```

### Scenario 3: Flag Override

```bash
$ export GOCLAW_SERVER=https://staging.example.com
$ goclaw --server https://custom.example.com agents list
# Uses: https://custom.example.com (flag overrides env)
```

### Scenario 4: Profile Switch

```bash
$ goclaw --profile staging agents list
# Loads profile named "staging" from config file
# Env and flags still override profile values
```

---

## Error Handling Strategy

### HTTP Errors

```
API Response:
  {
    "ok": false,
    "error": {
      "code": "ERR_NOT_FOUND",
      "message": "Agent not found"
    }
  }

Client Handling:
  1. Parse apiResponse
  2. Check OK field
  3. If false, wrap Error in command context
  4. Return: fmt.Errorf("fetch agent: %w", apiErr)
  5. Cobra prints to stderr and exits(1)
```

### WebSocket Errors

```
1. Connection Error: fmt.Errorf("dial websocket: %w", err)
2. Stream Error: fmt.Errorf("read message: %w", err)
3. Context Cancelled: fmt.Errorf("stream cancelled: %w", ctx.Err())
```

### Validation Errors

```
Before API Call:
  ├─ Arguments: cobra.ExactArgs(1), cobra.MinimumNArgs(1)
  ├─ Flags: cmd.MarkFlagRequired("name")
  └─ Values: Manual validation + fmt.Errorf()
```

---

## Security Architecture

### Credential Flow

```
User Input (--token flag, env var, or interactive prompt)
  ↓
OS Keyring (macOS Keychain, Linux Secret Service, Windows Credential Manager)
  ↓
At Runtime: Load from keyring into memory (Config.Token)
  ↓
HTTP Client: Add "Authorization: Bearer {Token}" header
  ↓
TLS Encryption (HTTPS by default, --insecure only for testing)
```

### Threat Mitigations

| Threat | Mitigation |
|--------|-----------|
| Token in history | Use env var instead of flag |
| Token in logs | Never log Config.Token |
| Token on disk | Store in OS keyring |
| MITM attack | TLS required by default |
| Process inspection | Credentials not in argv (use env var) |
| Config file exposure | Only metadata stored, not credentials |

---

## Performance Characteristics

### HTTP Requests

- **Timeout:** 30 seconds per request
- **Connection Pooling:** net/http default (reuses TCP connections)
- **Response Parsing:** Deferred (json.RawMessage), unmarshaled only when needed

### WebSocket Streams

- **Latency:** <100ms for messages
- **Throughput:** Limited by server and network
- **Memory:** Streaming mode (no buffering entire response)

### CLI Performance

- **Startup:** ~10ms
- **Config Load:** ~5ms
- **First Request:** ~100ms (includes TLS handshake)
- **Subsequent:** ~30ms (connection reuse)

---

## Extensibility Points

### Adding New Commands

1. Create `cmd/newfeature.go`
2. Define `var newFeatureCmd = &cobra.Command{...}`
3. Define subcommands and flags
4. Register in `init()`: `rootCmd.AddCommand(newFeatureCmd)`

### Custom Output Formats

1. Extend `internal/output/output.go`
2. Add format case to `Printer.Print()`
3. Implement marshal function
4. Update help text and output flag validation

### Custom Auth Methods

1. Extend `internal/client/auth.go`
2. Implement credential storage backend
3. Register in auth command
4. Update login flow

---

## Deployment Topology

```
Developer Workstation
  └─> goclaw CLI (statically-linked binary)
      └─> HTTPS/WSS → GoClaw Server
          ├─> Port 443 (HTTPS)
          └─> Port 443 (WSS via TLS)

CI/CD Pipeline
  └─> goclaw CLI (via go install or release binary)
      └─> HTTPS → GoClaw Server (via env vars)

Container/Docker
  └─> COPY goclaw binary
      └─> ENV GOCLAW_SERVER=... GOCLAW_TOKEN=...
          └─> ENTRYPOINT ["/usr/bin/goclaw", "command"]
```

---

## Version Management

### Build-Time Injection

```bash
VERSION=$(git describe --tags --always --dirty)
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

go build -ldflags "-X github.com/nextlevelbuilder/goclaw-cli/cmd.Version=$VERSION ..."
```

### Runtime Access

```go
$ goclaw version
GoClaw CLI v1.0.0 (commit: abc1234, built: 2026-03-15T10:00:00Z)
```

---

---

## AI Ergonomics (Phase 0 Locked Contracts)

### Design Principles for AI/Automation Consumers

**Contract 1: Deterministic Exit Codes**
- Exit codes 0-6 map to semantic error categories
- Enables AI agents to implement retry logic based on error type
- Example: Exit 2 (auth) → re-authenticate; Exit 6 (rate-limit) → exponential backoff

**Contract 2: TTY-Aware Output (No Special Flags Needed)**
- Detects piped output automatically → JSON
- Human terminal → table (pretty columns)
- CI/scripts get machine-readable output without `--output json`
- Env var override: `GOCLAW_OUTPUT=json` forces format

**Contract 3: Structured Error JSON**
- Errors always include code + message + details (in JSON mode)
- Parseable by AI agents for decision-making
- HTTP status codes mapped to exit codes for envelope-less responses

**Contract 4: Streaming with Automatic Reconnection**
- `FollowStream` pattern: exponential backoff on disconnect
- Used by `logs --follow`, `agents wait`, `teams events --follow`
- Handler errors stop immediately (no retry loop)
- Context cancellation respected for clean shutdown

### TTY Detection Flow

```
┌─────────────────────────────────────────────┐
│ Command Execution                           │
└────────────────┬────────────────────────────┘
                 │
        ┌────────▼────────────┐
        │ Check --output flag │
        └────────┬────────────┘
                 │
          Yes (explicit) → Use flag value
          │
          No → ┌─────────────────────┐
               │ Check GOCLAW_OUTPUT │
               └────────┬────────────┘
                        │
                 Yes (env set) → Use env value
                 │
                 No → ┌─────────────────────┐
                      │ Check stdout is TTY │
                      └────────┬────────────┘
                               │
                        Yes → "table" (human)
                        │
                        No → "json" (machine/CI)
```

### Error Mapping for AI Agents

**Server Error Code → Exit Code + Action:**

| Server Code | Exit Code | Meaning | AI Action |
|-------------|-----------|---------|-----------|
| `UNAUTHORIZED` | 2 | Auth required | Re-authenticate |
| `NOT_PAIRED` | 2 | Device pairing needed | Run `goclaw auth login --pair` |
| `NOT_FOUND` | 3 | Resource missing | Fail soft (not transient) |
| `INVALID_REQUEST` | 4 | Bad input | Fix request, don't retry |
| `FAILED_PRECONDITION` | 4 | State error | Wait + retry (state dependent) |
| `INTERNAL` | 5 | Server error | Hard fail (server down?) |
| `UNAVAILABLE` | 5 | Server unavailable | Exponential backoff retry |
| `AGENT_TIMEOUT` | 5 | Agent unresponsive | Check agent status |
| `RESOURCE_EXHAUSTED` | 6 | Rate-limited | Backoff (retry_after_ms respected) |
| HTTP 429 | 6 | Rate-limited | Respect Retry-After header |
| Connection timeout | 6 | Network issue | Transient; exponential backoff |

### AI-Critical Commands (Phase 4 MAX POLISH)

**Fully polished for AI agent orchestration:**

1. **`chat history`** — Retrieve conversation history
   - JSON schema in help text
   - Structured message array (role + content + timestamp)
   - Used by: AI agents reviewing context

2. **`chat inject`** — Inject context without triggering response
   - Role validation (system/user/assistant)
   - Content validation (required)
   - Used by: Orchestration tools injecting state/facts

3. **`chat session-status`** — Snapshot of session state
   - Ready-to-use for state machines
   - Used by: Workflow engines checking preconditions

4. **`agents wait --timeout=30s --state=ready`** — Blocking wait for agent state
   - Exit code 6 on timeout (retryable)
   - Used by: Orchestration waiting for agent availability

5. **`agents identity`** — Agent persona/identity snapshot
   - Used by: Multi-agent systems with role delegation

6. **`memory kg` subsystem** — Full knowledge graph CRUD
   - Entities, traversal, deduplication, graph export
   - Used by: RAG + semantic search workflows

### No Manual Format Flags Needed in CI

**Before (old style):**
```bash
goclaw agents list -o json | jq '.[] | select(.status == "active")'
```

**After (new style, equally readable):**
```bash
goclaw agents list | jq '.[] | select(.status == "active")'
# Still outputs JSON because stdout is piped!
```

### Error Handling Example (AI Agent)

```bash
#!/bin/bash
set -e

# Call with `set -e`: any non-zero exit stops script
goclaw agents wake myagent --timeout=10s

case $? in
  0)  echo "Agent ready" ;;
  2)  echo "Auth failed, re-authenticate" ;;
  3)  echo "Agent not found" ;;
  5)  echo "Server error, check GoClaw status" ;;
  6)  echo "Timeout, retrying..." && sleep 5 && retry ;;
esac
```

---

## Last Updated

- **Date:** 2026-04-15
- **Phases:** Legacy 1-9 Complete + P0-P4 Complete + P5 Deferred
- **Diagram Language:** ASCII (no external rendering needed)
- **Status:** Production Ready
- **AI Ergonomics:** Phase 0 locked contracts in place (exit codes, TTY-detect, error structs, streaming reconnect)
