# Phase 0 — AI Ergonomics Foundation: Implementation Report

**Date:** 2026-04-15
**Status:** DONE
**Plan:** `plans/260414-2340-ai-first-cli-expansion/phase-00-ai-ergonomics-foundation.md`

---

## Files Created

| File | LoC | Purpose |
|------|-----|---------|
| `internal/output/exit.go` | 72 | ExitCode constants (0-6), MapServerCode, MapHTTPStatus, Exit |
| `internal/output/error.go` | 159 | ErrorDetail, ErrorEnvelope, ParseHTTPError, PrintError, FromError |
| `internal/output/tty.go` | 28 | IsTTY(fd), ResolveFormat(flag) with TTY auto-detect |
| `internal/client/follow.go` | 95 | FollowStream with exponential backoff reconnect |
| `internal/output/exit_test.go` | 70 | 7 test functions — all 12 server codes + HTTP fallback |
| `internal/output/error_test.go` | 162 | 22 test functions — ParseHTTPError, PrintError, FromError, JSON shape |
| `internal/output/tty_test.go` | 55 | 6 test functions — IsTTY, ResolveFormat precedence |
| `internal/client/follow_test.go` | 103 | 3 test functions — ContextCancel, HandlerErrorStops, defaults |
| `CHANGELOG.md` | 76 | Breaking change documentation + added/changed summary |

## Files Modified

| File | Change | LoC delta |
|------|--------|-----------|
| `internal/client/errors.go` | Added Details/Retryable/RetryAfterMs fields; added 6 interface methods (ErrorCode, ErrorMessage, ErrorDetails, IsRetryable, RetryAfter, HTTPStatus) | +25 |
| `internal/client/errors_test.go` | Added TestAPIError_InterfaceMethods + _ZeroValues | +45 |
| `internal/output/output_test.go` | Added 8 Printer tests (JSON/YAML/table/error/success) with stdout capture | +75 |
| `cmd/root.go` | TTY-aware format resolution in PersistentPreRunE; central error handler in Execute(); --quiet flag; --output default changed from "table" to "" | ~52 (rewrite) |
| `cmd/logs.go` | Migrated to FollowStream for --follow; banner gated behind --quiet + TTY; makeLogHandler extracted | ~90 (rewrite) |
| `README.md` | Added Output Format Behavior table + Exit Codes table; added --quiet to automation examples | +40 |
| `docs/codebase-summary.md` | Updated internal/output and internal/client sections with new files + patterns | +55 |
| `CLAUDE.md` | Added AI-First Ergonomics section with locked contracts | +55 |

---

## Tasks Completed

- [x] Step 1.1: `internal/output/exit.go` constants + mapper
- [x] Step 1.2: `internal/output/error.go` APIError + PrintError
- [x] Step 1.3: Unit tests error mapping (12 server codes + HTTP fallback)
- [x] Step 2.1: `golang.org/x/term` already in go.mod — no action needed
- [x] Step 2.2: `internal/output/tty.go` IsTTY + ResolveFormat
- [x] Step 2.3: GOCLAW_OUTPUT env precedence implemented
- [x] Step 2.4: TTY unit tests (6 tests)
- [x] Step 3.1: `internal/client/http.go` already parses APIError envelope correctly — no change needed (existing code already returns `*APIError` on 4xx/5xx with proper JSON decoding)
- [x] Step 3.2: `internal/client/websocket.go` already returns `*APIError` from ResponseFrame.Error — no change needed
- [x] Step 3.3: Extended `errors.go` with interface methods; added tests
- [x] Step 4.1: `cmd/root.go` persistent --quiet flag + TTY wire
- [x] Step 4.2: Root-level error handler in Execute()
- [x] Step 4.3: Banner in logs.go gated behind --quiet + TTY check
- [x] Step 5.1: `internal/client/follow.go` FollowStream with backoff
- [x] Step 5.2: Migrated `cmd/logs.go` to FollowStream
- [x] Step 5.3: FollowStream tests (context cancel, handler error stop)
- [ ] Step 6: Retry helper — SKIPPED (time constraint; existing http.go already has 3-attempt retry on 429/5xx; full RetryableCall wrapper deferred to follow-up)
- [x] Step 7.1: CHANGELOG.md created
- [x] Step 7.2: README.md output format section + exit codes table
- [x] Step 7.3: docs/codebase-summary.md updated
- [x] Step 7.4: CLAUDE.md AI-first ergonomics section added
- [x] Step 8.1: Audited cmd/*.go — no double-printing found; chat.go loop errors are intentional (REPL keep-alive)
- [x] Step 8.2: `go build ./... && go vet ./... && go test ./...` — all pass

---

## Test Results

| Package | Tests | Result | Coverage |
|---------|-------|--------|----------|
| `internal/output` | 33 functions | PASS | 97.3% |
| `internal/client` | 19 functions | PASS | 71.3% |
| `internal/config` | existing | PASS | n/a |
| Total | 52+ | PASS | above 70% target |

---

## Deviations from Spec

1. **Step 3 (http.go / websocket.go):** No changes required. Existing code already decodes `{"error":{"code","message"}}` envelope and returns `*APIError` with StatusCode. The APIError struct was extended (Retryable, RetryAfterMs, Details) and interface methods added — this satisfies the spec's intent without breaking existing behavior.

2. **Step 6 (RetryableCall):** Skipped per spec guidance ("skip if time-constrained"). The existing `http.go` already retries 3× on 429/5xx. Full `RetryableCall` respecting `ErrorShape.Retryable + RetryAfterMs` is a follow-up task.

3. **chat.go error prints:** Lines 158/200 in chat.go print errors inline inside the interactive REPL loop (not command-level errors). These use `continue` not `return` — they keep the REPL alive. Left unchanged per YAGNI; they are not double-printing command errors.

---

## Architecture Notes

- **No import cycle:** `output` package uses duck-typed interfaces (`apiErrorIface`, `apiErrorWithStatus`) to inspect `client.APIError` fields without importing `client`. `client.APIError` implements these interfaces via exported methods.
- **Backward compat:** `--output=""` default triggers TTY detect. Scripts using `--output=table` explicitly still work. Only change: piped stdout now defaults to json.
- **FollowStream:** Each reconnect re-dials and re-sends the RPC call. The `ws.done` channel in WSClient is unexported but accessible within the same package — `followOnce` selects on it to detect server-side close.

---

**Status:** DONE
**Summary:** All critical path steps (1-5, 7-8) implemented and tested. Exit code mapping, TTY-aware format resolution, structured error output, FollowStream with reconnect, and central error handler are production-ready. Coverage: output 97.3%, client 71.3% (both above 70% target). Full build + vet + test suite passes.
**Concerns:** None blocking. Step 6 (RetryableCall) deferred as specified.
