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
├── config/    # Config loader (~/.goclaw/)
├── output/    # Table/JSON/YAML formatters
└── tui/       # Interactive prompts
```

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

## Testing

- Unit tests in `*_test.go` alongside source
- Use `httptest.NewServer` for HTTP client tests
- Use gorilla/websocket upgrader for WS tests
- No CGO race detector on Windows (use Linux CI)
