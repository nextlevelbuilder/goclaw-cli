# GoClaw CLI

Go CLI for managing GoClaw AI agent gateway servers.

## Tech Stack

- **Language:** Go 1.25
- **CLI:** Cobra (commands) + Viper-style config
- **Transport:** HTTP REST + WebSocket RPC (gorilla/websocket)
- **Config:** `~/.goclaw/config.yaml` + env vars + flags

## Build & Test

```bash
go build ./...           # Compile check
go vet ./...             # Static analysis
go test ./...            # Run all tests
go test -count=1 ./...   # Skip test cache
make build               # Build binary with ldflags
make install             # Install to GOPATH/bin
```

## Project Structure

```
cmd/           # Cobra command files (1 per resource group)
internal/
├── client/    # HTTP + WebSocket + auth clients
│   ├── errors.go    # APIError (matches server ErrorShape)
│   └── follow.go    # FollowStream() with exponential backoff reconnect
├── config/    # Config loader (~/.goclaw/)
├── output/    # Table/JSON/YAML formatters + AI-ergonomics foundation
│   ├── exit.go      # ExitCode constants (0-6) + MapServerCode + Exit
│   ├── error.go     # ErrorDetail/PrintError/ParseHTTPError/FromError
│   └── tty.go       # IsTTY + ResolveFormat (TTY auto-detection)
└── tui/       # Interactive prompts
```

## AI-First Ergonomics (Phase 0 — implemented)

These patterns are **locked** — do not change without updating CHANGELOG.md.

### Output format auto-detection
Precedence: `--output` flag > `GOCLAW_OUTPUT` env > TTY detection
- stdout is TTY → `"table"` (human)
- stdout is piped/redirected → `"json"` (machine)

### Exit codes (automation contract)
| Code | Trigger |
|------|---------|
| 0 | Success |
| 1 | Generic/unknown |
| 2 | Auth (UNAUTHORIZED, NOT_PAIRED, etc.) |
| 3 | Not found |
| 4 | Validation |
| 5 | Server error |
| 6 | Resource/network/rate-limit |

### Error output shape (JSON mode)
```json
{"error": {"code": "UNAUTHORIZED", "message": "..."}}
```

### Central error handler
All command errors bubble via `return err` to `cmd.Execute()` → `output.PrintError` + `output.Exit(output.FromError(err))`. Do NOT print errors in individual commands.

### Flags
- `--quiet` — suppresses banners/tips in non-interactive contexts
- `--output` / `-o` — default empty (triggers auto-detect), not `"table"`

## Conventions

- Go snake_case file naming
- Cobra command pattern: register in `init()`, implement as `RunE`
- Config precedence: flags > env vars > config file
- Token stored in credential store (not config.yaml)
- All destructive ops require `--yes` or interactive confirmation
- Dual mode: interactive (table output) + automation (JSON/YAML)

## Key Patterns

- `newHTTP()` / `newWS()` — create authenticated clients from global config
- `buildBody()` — construct request body from flag values, skip empty
- `readContent()` — read from `@filepath` or literal string
- `unmarshalMap()` / `unmarshalList()` — parse JSON responses
- `printer.Print()` — output in configured format
- `output.ResolveFormat(flagVal)` — resolve format with TTY fallback
- `output.FromError(err)` — map error to exit code
- `output.PrintError(err, format)` — format-aware error output
- `client.FollowStream(ctx, ...)` — persistent WS streaming with reconnect

## Testing

- Unit tests in `*_test.go` alongside source
- Use `httptest.NewServer` for HTTP client tests
- Use gorilla/websocket upgrader for WS tests
- No CGO race detector on Windows (use Linux CI)
