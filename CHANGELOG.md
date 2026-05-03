# Changelog

All notable changes to goclaw-cli are documented here.
Format: [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## [Unreleased] — Domain Coverage Expansion (P0–P2)

### Added

**P0 — Critical**
- `goclaw hooks` (list, create, update, delete, toggle, test, history) — manage event hooks via WS RPC `hooks.*`. Closes the entire hooks domain that was previously unreachable from CLI.
- `goclaw agents files` (list, get, set) — edit global agent context files (AGENTS.md, SOUL.md, IDENTITY.md, USER.md, USER_PREDEFINED.md, CAPABILITIES.md, BOOTSTRAP.md, MEMORY.json, HEARTBEAT) via WS RPC `agents.files.*`. `--propagate` pushes change to all existing user instances.

**P1 — Lifecycle & analytics**
- `goclaw agents cancel-summon <id>` — cancel an in-progress summon (`POST /v1/agents/{id}/cancel-summon`).
- `goclaw agents skills list <id>` — list skills granted to an agent (`GET /v1/agents/{id}/skills`).
- `goclaw usage timeseries` — bucketed usage over time (`GET /v1/usage/timeseries`).
- `goclaw usage breakdown` — usage broken down by agent/user/tenant (`GET /v1/usage/breakdown`).

**P2 — Coverage completion**
- `goclaw tts test-connection` — test a TTS provider end-to-end (`POST /v1/tts/test-connection`).
- `goclaw voices list` and `goclaw voices refresh` — voice catalog.
- `goclaw memory kg extract` — switched to new endpoint `POST /v1/agents/{id}/kg/extract` (legacy `/v1/knowledge-graph` path retired).
- `goclaw files sign` — generate signed URL for server-side file (`POST /v1/files/sign`).
- `goclaw teams workspace upload` and `goclaw teams workspace move` — multipart upload + rename for team workspace.
- `goclaw packages github-releases` — list GitHub releases for tracked packages.

### Notes
- All new commands honor the AI-first ergonomics contract: `--output=json` envelope, central error handler, `--yes` for destructive ops, `--quiet` for CI.
- Out of scope: OpenAI-compatible `/chat/completions` and `/v1/responses` endpoints (client APIs, not admin CLI surface).

---

## [Unreleased] — AI Ergonomics Foundation (Phase 0)

### Breaking Changes

#### Output format default changed when stdout is piped

**Before:** `goclaw agents list` always defaulted to `table` format regardless of context.

**After:** When stdout is not a terminal (piped, redirected, CI), the default format is now `json`.

**Migration:** Scripts relying on table output must add `--output=table` or `GOCLAW_OUTPUT=table`.

```bash
# Before (broke silently in CI)
goclaw agents list | grep "my-agent"

# After — explicit table for text parsing
goclaw agents list --output=table | grep "my-agent"

# Or use JSON (recommended for automation)
goclaw agents list | jq '.[] | select(.display_name == "my-agent")'
```

**Rationale:** AI tools, CI pipelines, and shell scripts consuming CLI output require
structured JSON. Table format is human-optimised and breaks piped parsing. TTY detection
ensures human operators still get tables by default.

### Added

- **`internal/output/exit.go`** — Exit code constants (0-6) + `MapServerCode(code)` + `MapHTTPStatus(status)` + `Exit(code)`
- **`internal/output/error.go`** — `ErrorDetail` / `ErrorEnvelope` types matching server `ErrorShape`; `ParseHTTPError(body, status)`; `PrintError(err, format)`; `FromError(err) int`
- **`internal/output/tty.go`** — `IsTTY(fd)` via `golang.org/x/term`; `ResolveFormat(flagVal)` with flag > `GOCLAW_OUTPUT` env > TTY precedence
- **`internal/client/follow.go`** — `FollowStream(ctx, ...)` with exponential backoff reconnect (max 5 retries) for `--follow` streaming commands
- **`--quiet` flag** — persistent flag on root command; suppresses banners and informational messages in non-TTY contexts
- **Exit code contract** — all server error codes now map deterministically to exit codes 0-6 for AI/automation consumers

### Changed

- `cmd/root.go` — output format resolved via TTY detection in `PersistentPreRunE`; central error handler in `Execute()` calls `output.PrintError` + `output.Exit(output.FromError(err))`
- `cmd/logs.go` — `logs tail --follow` migrated to `client.FollowStream` with auto-reconnect; banner gated behind `--quiet` and TTY check
- `--output` flag default changed from `"table"` to `""` (empty triggers auto-detect)
- `internal/client/errors.go` — `APIError` extended with `Details`, `Retryable`, `RetryAfterMs` fields matching server `ErrorShape`; added interface methods (`ErrorCode`, `ErrorMessage`, `ErrorDetails`, `IsRetryable`, `RetryAfter`, `HTTPStatus`) for duck-typed error handling in `output` package without import cycle

### Fixed

- Piped invocations no longer silently produce unparseable table output; they emit valid JSON
- Error details from server (`code`, `message`, `retryable`) are now fully preserved and passed through to the caller

---

## Previous releases

See git log for changes prior to this changelog.
