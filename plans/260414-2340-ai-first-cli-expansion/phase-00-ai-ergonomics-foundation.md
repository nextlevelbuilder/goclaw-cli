---
phase: 0
title: Cross-cutting AI Ergonomics Foundation
status: pending
priority: critical
---

# Phase 0 — AI Ergonomics Foundation

## Context Links
- Brainstorm report: `../reports/brainstorm-260414-2231-missing-commands-gap-analysis.md` §5 Phase 0
- Server error types: `../../../goclaw/pkg/protocol/errors.go`
- Server error shape: `../../../goclaw/pkg/protocol/frames.go` (`ErrorShape`)
- Server HTTP helper: `../../../goclaw/internal/http/response_helpers.go` (`writeError`, `ErrorResponse`)
- Existing CLI output: `internal/output/`
- Existing CLI client: `internal/client/`

## Overview
- **Priority:** Critical (blocks all other phases)
- **Status:** Pending
- **Description:** Thiết lập patterns chuẩn cho error handling, exit codes, streaming, và TTY-aware output BEFORE thêm command mới. AI tools consume JSON + exit codes; without foundation, P1-P5 sẽ phải retrofit.

## Key Insights
- Server đã có `ErrorShape{code, message, details, retryable, retryAfterMs}` — CLI chỉ pass-through, không invent format mới
- Server HTTP error envelope: `{"error": {"code": "...", "message": "..."}}`
- Idempotency-Key ❌ server chưa support → DROP khỏi P0
- Breaking change TTY default acceptable (CLI pre-1.0)
- File size rule 200 LoC: existing `cmd/agents.go` (416), `cmd/teams.go` (462), `cmd/skills.go` (~330) đã vượt ngưỡng → **không refactor split trong P0** (tránh bloat scope); defer to phase mở rộng group tương ứng

## Requirements

### Functional
- Structured JSON error output khi `--output=json` hoặc auto-detected JSON mode
- Exit codes mapping chính xác cho mọi server error
- `--follow` pattern cho commands stream (logs tail, hereafter events/heartbeat logs)
- TTY detection auto-switch output format
- `--quiet` flag universal suppress banner/tips
- Retry-on-retryable với respect `retryAfterMs` (optional, nếu thời gian cho phép)

### Non-Functional
- Zero new external deps (dùng `golang.org/x/term` — stdlib adjacent)
- Backward-compat path: explicit `--output=table` luôn hoạt động
- Documentation: CHANGELOG.md note breaking change
- Testing: unit tests cho error mapping, TTY detect, exit code setter

## Architecture

### New files
```
internal/output/
├── error.go          # ErrorEnvelope, PrintError(err, format), ParseServerError
├── exit.go           # ExitCode constants, MapServerCode(code) int, Exit(code)
└── tty.go            # IsTTY(fd), ResolveFormat(flag) → "table"|"json"|"yaml"

internal/client/
├── follow.go         # FollowStream(ctx, ws, method, params, handler) — JSON lines streaming
└── retry.go          # RetryableCall(c, fn, maxAttempts) — honors retryable+retryAfterMs
```

### Error flow
```
server (HTTP 4xx/5xx or WS error frame)
    ↓
client/http.go or client/websocket.go  — parse ErrorShape
    ↓
internal/output/error.go PrintError()  — format per --output flag
    ↓
stdout (json mode) or stderr (table mode)
    ↓
internal/output/exit.go Exit(code)     — os.Exit with mapped code
```

### Exit code contract (locked)
| Code | Meaning | Trigger |
|---|---|---|
| 0 | Success | Normal completion |
| 1 | Generic error | Unknown/unmapped |
| 2 | Auth | `UNAUTHORIZED`, `NOT_PAIRED`, `TENANT_ACCESS_REVOKED`, HTTP 401/403 |
| 3 | Not found | `NOT_FOUND`, `NOT_LINKED`, HTTP 404 |
| 4 | Validation | `INVALID_REQUEST`, `FAILED_PRECONDITION`, `ALREADY_EXISTS`, HTTP 400/409/422 |
| 5 | Server | `INTERNAL`, `UNAVAILABLE`, `AGENT_TIMEOUT`, HTTP 5xx |
| 6 | Resource/network | `RESOURCE_EXHAUSTED`, HTTP 429, connection timeout, DNS fail |

### TTY resolution logic
```
if --output explicitly set → use it
else if os.Stdout is TTY → "table"
else → "json"
```

Override via env `GOCLAW_OUTPUT` (optional, lower priority than flag).

## Related Code Files

### Create
- `internal/output/error.go` — error envelope + pretty printer
- `internal/output/exit.go` — exit code constants + mapper
- `internal/output/tty.go` — TTY detection + format resolution
- `internal/client/follow.go` — WS streaming follow helper
- `internal/client/retry.go` — retryable call wrapper
- `CHANGELOG.md` — document breaking changes

### Modify
- `internal/client/http.go` — return typed `APIError` (wrap `ErrorShape`) instead of raw error
- `internal/client/websocket.go` — parse `ErrorShape` from response frames, return typed error
- `internal/output/printer.go` (if exists) or create — route error to proper formatter
- `cmd/root.go` — persistent `--quiet` flag, wire TTY detect vào output resolution
- `cmd/logs.go` — migrate to new `FollowStream` helper
- All `cmd/*.go` — replace `return fmt.Errorf(...)` of server errors với typed error handling; set exit codes via `os.Exit(exit.FromError(err))` in root-level error handler

### Delete
- None (additive)

## Implementation Steps

### Step 1: Error envelope & exit code types
1. Create `internal/output/exit.go` with constants (Success, Generic, Auth, NotFound, Validation, Server, Resource) + `MapServerCode(code string) int`
2. Create `internal/output/error.go` with `APIError` struct (wraps `ErrorShape`), `PrintError(err, format)`, `ParseHTTPError(body []byte, status int) *APIError`
3. Write unit tests for code mapping (all 12 server codes + HTTP status fallback)

### Step 2: TTY detection
1. Add `golang.org/x/term` dep: `go get golang.org/x/term`
2. Create `internal/output/tty.go` with `IsTTY(fd int) bool`, `ResolveFormat(flag string) string`
3. Add env var `GOCLAW_OUTPUT` precedence check
4. Unit tests with stdin fd mock

### Step 3: Refactor client error parsing
1. Modify `internal/client/http.go`: decode response body on 4xx/5xx, try parse `{"error":{"code","message"}}`, fall back to plain text, return `*output.APIError`
2. Modify `internal/client/websocket.go`: parse `ResponseFrame.Error` on `ok=false`, return `*output.APIError`
3. Unit tests với httptest for HTTP, gorilla/websocket upgrader for WS

### Step 4: Root-level error handling
1. Modify `cmd/root.go`:
   - Add persistent `--quiet` flag
   - Wire TTY detect into `outputFormat` resolution
   - Central error handler in `main()`: call `output.PrintError` + `os.Exit(exit.FromError(err))`
2. Remove banner/tips output in non-TTY mode (gate behind `IsTTY(stdout)`)

### Step 5: Follow stream helper
1. Create `internal/client/follow.go`:
   - `FollowStream(ctx, ws, method, params, handler func([]byte)) error`
   - Handles reconnect-on-error with backoff
   - Emits JSON lines on stdout when format=json
2. Migrate `cmd/logs.go:tail` to use new helper — validate no regression

### Step 6: Retry helper (optional, time-permitting)
1. Create `internal/client/retry.go` with `RetryableCall` wrapper respecting `ErrorShape.Retryable` + `RetryAfterMs`
2. Apply to known-retryable ops (chat send, tools invoke) — opt-in per command

### Step 7: Documentation
1. Create `CHANGELOG.md` với breaking change entry
2. Update `README.md` với output format behavior table
3. Update `docs/codebase-summary.md` với new internal packages
4. Update `CLAUDE.md` với AI-first ergonomics note

### Step 8: Backfill refactor
1. Audit all `cmd/*.go` for `fmt.Errorf("%w", err)` patterns that swallow server error details
2. Replace with typed error bubble-up to root handler
3. Ensure no double-printing (errors printed once at root, not per-command)

## Todo List

- [x] Step 1.1: `internal/output/exit.go` constants + mapper
- [x] Step 1.2: `internal/output/error.go` APIError + PrintError
- [x] Step 1.3: Unit tests error mapping (12 server codes + HTTP)
- [x] Step 2.1: Add `golang.org/x/term` dep
- [x] Step 2.2: `internal/output/tty.go` IsTTY + ResolveFormat
- [x] Step 2.3: GOCLAW_OUTPUT env precedence
- [x] Step 2.4: TTY unit tests
- [x] Step 3.1: Refactor `internal/client/http.go` error parsing
- [x] Step 3.2: Refactor `internal/client/websocket.go` error parsing
- [x] Step 3.3: Client unit tests
- [x] Step 4.1: `cmd/root.go` persistent `--quiet` + TTY wire
- [x] Step 4.2: Root-level error handler in main()
- [x] Step 4.3: Remove non-TTY banners
- [x] Step 5.1: `internal/client/follow.go` FollowStream
- [x] Step 5.2: Migrate `cmd/logs.go` to FollowStream
- [x] Step 5.3: FollowStream integration test
- [ ] Step 6: Retry helper — **INTENTIONALLY SKIPPED** (time constraint per spec; existing http.go retries 3× on 429/5xx; full RetryableCall deferred to follow-up phase)
- [x] Step 7.1: CHANGELOG.md
- [x] Step 7.2: README.md output format section
- [x] Step 7.3: docs/codebase-summary.md update
- [x] Step 8.1: Audit + refactor `cmd/*.go` error handling
- [x] Step 8.2: Verify `go build ./... && go vet ./... && go test ./...` all pass
- [x] Step 8.3: Manual smoke test: `goclaw agents list` with bad token → exit 2 + JSON error

## Success Criteria
- [ ] `goclaw agents list --output=json` returns JSON array, exit 0
- [ ] `goclaw agents list` in TTY returns table
- [ ] `goclaw agents list` piped to `cat` returns JSON (auto-detect)
- [ ] Invalid token → exit 2 + `{"error":{"code":"UNAUTHORIZED","message":"..."}}`
- [ ] Missing agent → exit 3 + NOT_FOUND error
- [ ] Server 500 → exit 5 + INTERNAL error
- [ ] `goclaw logs tail --follow` streams JSON lines (non-TTY) or table lines (TTY)
- [ ] `go test ./...` passes, new coverage ≥70% for `internal/output` and `internal/client`
- [ ] CHANGELOG documents breaking change

## Risk Assessment

| Risk | Mitigation |
|---|---|
| Breaking existing scripts parsing table stdout | CHANGELOG entry + README migration guide. Users set `--output=table` explicit. |
| Double error printing (handler + cmd) | Centralize at root handler; commands only `return err` |
| WS streaming follow disconnect/reconnect edge cases | Exponential backoff + max 5 retries; escape on ctx cancel |
| Env `GOCLAW_OUTPUT` collision với config file | Flag > env > config > default precedence, documented |
| Retry helper causing duplicate create ops | Gate retry behind `ErrorShape.Retryable=true` only; skip POST by default |

## Security Considerations
- Error messages có thể leak server internals (paths, user IDs). Server đã i18n-translate, CLI pass-through → trust server sanitization
- `--quiet` không được suppress security-critical messages (auth failures, permission denials)
- Exit codes are observable by wrapping processes — do NOT differentiate 401 vs 403 with separate codes (both 2) per principle of least info

## Next Steps
- Dependencies: None (P0 is foundation)
- Unblocks: All P1-P5 phases
- Follow-up: After P5 merged, revisit idempotency-key when server adds support

## Unresolved Questions
1. Có cần `GOCLAW_NO_COLOR` support không? (Cobra auto-handle via `NO_COLOR` env → check default)
2. Retry helper có nên gate behind `--retry` flag thay vì auto? (YAGNI nghiêng về auto cho retryable=true)
3. Log format trong `--follow`: plain text hay JSON lines? (Recommendation: JSON lines khi json mode, text khi table mode)
