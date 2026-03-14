---
phase: 2
title: Core Client & Auth
status: planned
priority: critical
effort: M
depends_on: [phase-01]
---

# Phase 2 — Core Client & Auth

## Overview
HTTP REST client, WebSocket RPC client, auth commands, config management, and output formatters.

## Context Links
- [Plan Overview](plan.md)
- GoClaw WebSocket Protocol: `../goclaw/websocket-protocol.md`
- GoClaw API Reference: `../goclaw/api-reference.md`

## Key Insights
- Server uses Bearer token auth (admin) + device pairing (operator)
- WebSocket protocol v3: `req`/`res`/`event` frames, `connect` handshake first
- Config stored in `~/.goclaw/config.yaml`, credentials in OS keyring or encrypted file
- All API calls go through `/v1/` prefix

## Architecture

```
internal/config/config.go     — Load/save ~/.goclaw/config.yaml
internal/client/http.go       — REST client (GET/POST/PUT/DELETE/PATCH)
internal/client/websocket.go  — WS RPC client (connect, call, subscribe)
internal/client/auth.go       — Token store (keyring/file), login flow
internal/output/formatter.go  — Table/JSON/YAML output dispatcher
internal/output/table.go      — Table renderer
internal/output/json.go       — JSON/YAML renderer
internal/tui/prompt.go        — Interactive prompts (confirm, select, input)
```

## Implementation Steps

### Config (`internal/config/`)
1. Config struct: `ServerURL`, `DefaultAgent`, `OutputFormat`, `Locale`
2. Load from `~/.goclaw/config.yaml` → env vars → flags (precedence)
3. `Save()` to persist changes
4. Profile support: `goclaw auth login --profile staging`

### HTTP Client (`internal/client/http.go`)
1. `Client` struct with `baseURL`, `token`, `httpClient`
2. Methods: `Get`, `Post`, `Put`, `Patch`, `Delete`
3. Auto-add `Authorization: Bearer {token}` header
4. Error handling: parse JSON error response, map to CLI errors
5. Retry with backoff on 429/5xx
6. TLS config: verify by default, `--insecure` to skip
7. Timeout: 30s default, configurable

### WebSocket Client (`internal/client/websocket.go`)
1. `WSClient` struct with connection, request ID counter
2. `Connect(serverURL, token, userID)` — handshake
3. `Call(method, params)` — send req, wait for matching res
4. `Subscribe(eventType, handler)` — register event listener
5. `Stream(method, params, onEvent)` — for streaming (chat, logs)
6. Auto-reconnect with exponential backoff
7. Ping/pong keepalive

### Auth (`internal/client/auth.go` + `cmd/auth.go`)
1. Token storage:
   - Primary: OS keyring (`go-keyring`)
   - Fallback: `~/.goclaw/credentials` (0600 permissions, encrypted)
2. `goclaw auth login` — prompt for server URL + token, verify via `/health`, save
3. `goclaw auth login --pair` — device pairing flow:
   - Connect WS without token → get pairing code
   - Display code to user, poll `browser.pairing.status`
   - On approval, save sender_id for reconnection
4. `goclaw auth logout` — clear stored credentials
5. `goclaw auth whoami` — show current user, role, server
6. `goclaw auth rotate` — regenerate token hint

### Output Formatters (`internal/output/`)
1. `Format(data, format)` — dispatch to table/json/yaml
2. Table: colored headers, truncated long values, responsive width
3. JSON: pretty-print by default, `--compact` for minified
4. YAML: for complex nested output
5. Error formatter: human-readable vs structured JSON

### Interactive TUI (`internal/tui/`)
1. Detect automation mode: `--yes` flag or non-TTY stdin
2. `Confirm(msg)` — y/n prompt (auto-yes in automation)
3. `Select(label, options)` — single select
4. `Input(label, default)` — text input
5. `Password(label)` — masked input

## Related Code Files
- Create: `internal/config/config.go`
- Create: `internal/client/http.go`, `websocket.go`, `auth.go`
- Create: `internal/output/formatter.go`, `table.go`, `json.go`
- Create: `internal/tui/prompt.go`
- Create: `cmd/auth.go`, `cmd/status.go`

## Todo
- [ ] Config loader with profile support
- [ ] HTTP REST client with auth, retry, TLS
- [ ] WebSocket RPC client with connect, call, subscribe, stream
- [ ] Token storage (keyring + file fallback)
- [ ] Auth commands: login, logout, whoami, pair
- [ ] Output formatters: table, JSON, YAML
- [ ] TUI prompts with automation mode detection
- [ ] Status/health command

## Success Criteria
- `goclaw auth login` stores credentials securely
- `goclaw auth login --pair` completes device pairing
- `goclaw status` shows server health via REST
- `goclaw status --output json` returns structured JSON
- Auth token auto-injected into all subsequent API calls

## Security Considerations
- Token never in CLI flags visible to `ps` — use env var or prompt
- Credentials file: 0600 permissions, AES-encrypted content
- TLS verification on by default
- No token logging in verbose mode (redact in debug output)

## Risk Assessment
- OS keyring may not work in headless/container — file fallback required
- WebSocket reconnection under flaky networks — exponential backoff with jitter
