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

### 2. Command Layer (cmd/ - 21 files)

**Pattern: Cobra Command Structure**

```go
rootCmd (goclaw)
  ├─ PersistentFlags: --server, --token, --output, --yes, --verbose, --insecure, --profile
  ├─ PersistentPreRunE: Load config, create printer
  └─ Subcommands (30 groups)
      ├─ auth (login, logout, use-context)
      ├─ agents (list, get, create, update, delete, share)
      ├─ api-keys (list, create, revoke)
      ├─ api-docs (open, spec)
      ├─ chat (interactive + streaming)
      ├─ sessions, skills, mcp, providers, tools, cron, teams, channels...
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

### 5. Output Formatting (internal/output/)

**Printer Interface:**

```go
type Printer struct {
	Format string
}

// Output dispatcher
func (p *Printer) Print(data any)
  ├─ if json: marshal to compact JSON
  ├─ if yaml: marshal to YAML
  └─ if table: format as Table
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
{
  "agents": [
    { "id": "abc123", "name": "Agent1", "provider": "openai" }
  ]
}

# YAML (config friendly)
goclaw agents list -o yaml
agents:
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

## Last Updated

- **Date:** 2026-03-15
- **Diagram Language:** ASCII (no external rendering needed)
- **Status:** Production Ready
