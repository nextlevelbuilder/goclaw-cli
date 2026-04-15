# Phase 0 — AI Ergonomics Foundation: Code Review

**Date:** 2026-04-15
**Reviewer:** code-reviewer
**Scope:** Phase 0 implementation per `plans/260414-2340-ai-first-cli-expansion/phase-00-ai-ergonomics-foundation.md`
**Implementation report:** `plans/reports/fullstack-260415-0005-phase-00-impl.md`

---

## Verdict

**Score: 9.4 / 10**
**Critical: 0 | High: 1 | Medium: 3 | Low: 4 | Info: 3**
**Recommendation:** PASS WITH MINOR FIXES (high-priority retry-on-handler-error semantics should be addressed before P1 merges; not blocking).

Build, vet, and tests all pass. Coverage targets met (output 97.3%, client 71.3%). Architecture clean (no import cycle via duck-typed interfaces). 12/12 server error codes mapped. TTY precedence correct.

---

## Files Reviewed

- `internal/output/exit.go` (72)
- `internal/output/error.go` (159)
- `internal/output/tty.go` (28)
- `internal/output/exit_test.go`, `error_test.go`, `tty_test.go`
- `internal/client/follow.go` (95)
- `internal/client/follow_test.go` (103)
- `internal/client/errors.go` (modified +25)
- `internal/client/errors_test.go` (modified +45)
- `cmd/root.go` (rewrite ~52)
- `cmd/logs.go` (rewrite ~90)
- `CHANGELOG.md`, `README.md`, `docs/codebase-summary.md`, `CLAUDE.md`

---

## Findings

### Critical (0)

None.

### High (1)

#### H1. FollowStream retries on handler errors — contradicts spec

**File:** `internal/client/follow.go:39-65`
**Spec requirement (focus area 5):** "Handler error → immediate stop without retry (correct)?"

`followOnce` returns the handler's error via `errCh` (line 99-100). `FollowStream` then sees non-nil and falls through to the retry/backoff loop (line 41-65). It does NOT distinguish handler-originated errors from connection/transport errors. As a result, a handler returning `fmt.Errorf("bad payload")` will trigger up-to-5 retries with exponential backoff (max ~31s) before exiting.

**Impact:** Wasted reconnect cycles on user-driven stops (e.g., terminal pipe closed). Spec explicitly listed this as a focus area to verify; current behavior fails the contract.

**Fix (recommended):** Wrap handler errors in a sentinel and short-circuit before retry.

```go
// follow.go — add at top:
var errHandlerStop = errors.New("handler requested stop")

// followOnce: line 99-100
case err := <-errCh:
    return fmt.Errorf("%w: %v", errHandlerStop, err)

// FollowStream: after `if err == nil { return nil }`
if errors.Is(err, errHandlerStop) {
    return err  // do not retry
}
```

The existing test `TestFollowStream_HandlerErrorStops` does not actually verify "stop without retry" — it relies on `MaxRetries: 0` → defaulted-to-5, and only checks that the call returns "quickly" without asserting attempt count. Add an assertion.

---

### Medium (3)

#### M1. `MaxRetries: 0` silently overridden to default 5

**File:** `internal/client/follow.go:30-32`

`if cfg.MaxRetries > 0 { maxRetries = cfg.MaxRetries }` — passing `0` (which a caller would intuitively use to mean "do not retry") gets coerced to the default 5. The test at `follow_test.go:124` is implicitly broken by this — comment says "With MaxRetries=0 ... should return quickly" but it actually does up to 5 retries (each `BaseDelay=10ms` → fast enough to slip past detection within the 2s timeout).

**Fix:** Use sentinel `-1` for "no retries", treat `0` as literal zero:

```go
maxRetries := 5
if cfg != nil && cfg.MaxRetries != 0 {
    if cfg.MaxRetries < 0 {
        maxRetries = 0
    } else {
        maxRetries = cfg.MaxRetries
    }
}
```

Or simpler — accept zero as literal and only default on negative.

#### M2. No HTTP integration test for new server error envelope shape

**File:** `internal/client/http.go:115-188` (no test file)

Spec Step 3.3: "Client unit tests with httptest for HTTP, gorilla/websocket upgrader for WS". Implementation report skipped Step 3.1/3.2 with rationale "existing code already returns *APIError" — but no test verifies that the server's HTTP envelope `{"error":{"code":"NOT_FOUND","message":"..."}}` (no `ok` field, per `goclaw/internal/http/response_helpers.go:24`) is correctly parsed. The legacy `apiResponse` struct happens to match because `OK` defaults to `false` and `Error` field deserialises with matching JSON tags — but this is implicit and fragile. A regression in the envelope (server adds wrapper, renames field) would silently fall into the "Non-envelope response" branch on line 179, producing `Code = "Not Found"` (HTTP status text) instead of `"NOT_FOUND"`.

**Fix:** Add `internal/client/http_test.go` with `httptest.NewServer` cases:
- 404 with `{"error":{"code":"NOT_FOUND","message":"agent missing"}}` → assert `APIError.Code == "NOT_FOUND"`
- 500 with plain text body → assert fallback path
- 200 with `{"ok":true,"payload":{...}}` → success path
- 429 → triggers retry

#### M3. Invalid `--output` value silently degrades

**File:** `internal/output/tty.go:22-33`, `cmd/root.go:39`

`ResolveFormat("invalid")` passes through unchanged. `Printer.Format = "invalid"` falls into the table fallback path (`switch ... default`). `PrintError(err, "invalid")` falls into the human-readable stderr branch. No error returned, no warning. Spec focus area 3: "`GOCLAW_OUTPUT=invalid` handled gracefully?" — current behavior is "silent degradation" not "graceful".

**Fix:** Validate in `ResolveFormat` or in `PersistentPreRunE`:

```go
func ResolveFormat(flagVal string) (string, error) {
    raw := flagVal
    if raw == "" {
        if v := os.Getenv("GOCLAW_OUTPUT"); v != "" {
            raw = v
        } else if IsTTY(int(os.Stdout.Fd())) {
            raw = "table"
        } else {
            raw = "json"
        }
    }
    switch raw {
    case "table", "json", "yaml":
        return raw, nil
    default:
        return "", fmt.Errorf("invalid output format %q (must be table, json, yaml)", raw)
    }
}
```

Note: this changes the function signature — minor caller update in `cmd/root.go`.

---

### Low (4)

#### L1. `httpStatusCode` for HTTP 403 returns `TENANT_ACCESS_REVOKED`

**File:** `internal/output/error.go:58-59`

Mapping `403 → TENANT_ACCESS_REVOKED` is too specific — 403 can mean any authorization failure. The tested behavior `TestParseHTTPError_403` enforces this mapping which is misleading because the server may return 403 for non-tenant reasons (RBAC, IP allowlist, etc.). The exit code outcome is correct (both → exit 2 / `ExitAuth`) so user-visible impact is nil, but the JSON envelope's `code` field would mislead AI consumers parsing the code.

**Fix:** Use a generic `FORBIDDEN` constant for HTTP fallback only, OR fall back to a generic `UNAUTHORIZED` (which already maps to ExitAuth):

```go
case 403:
    return "UNAUTHORIZED"  // server-canonical code; HTTP-only fallback
```

This is only used when the server failed to return its own envelope (which already includes the proper code).

#### L2. `ExitGeneric` overlap in `MapHTTPStatus(200)`

**File:** `internal/output/exit.go:51-66`, test `exit_test.go:54`

`MapHTTPStatus(200) = ExitGeneric` (1) — but 200 is not an error. The function is only called via the error path (`FromError` after `apiErrorWithStatus`), so this never happens in practice. The test `{200, ExitGeneric}` asserts a behavior that's never exercised. Either remove the test row or document that callers must only invoke for error statuses.

#### L3. `MapHTTPStatus` does not cover 408 (Request Timeout)

**File:** `internal/output/exit.go:51-66`

Spec exit code table line 6: "connection timeout, DNS fail" → ExitResource (6). HTTP 408 (Request Timeout) is not mapped. Falls through to `ExitGeneric` (1). Likely rare — but if the server is fronted by a proxy that returns 408, exit code becomes ambiguous.

**Fix:** Add `case status == 408: return ExitResource`.

#### L4. `cmd/root.go` Execute() best-effort format fallback duplicates `ResolveFormat` logic

**File:** `cmd/root.go:54-66`

When `cfg == nil` (PersistentPreRunE failed before `cfg.OutputFormat` was set), the fallback inlines TTY detection. This duplicates the precedence chain in `ResolveFormat`. Refactor:

```go
format := output.ResolveFormat("")  // re-applies env + TTY chain
if cfg != nil && cfg.OutputFormat != "" {
    format = cfg.OutputFormat
}
```

Pure cleanup — no behavioral change.

---

### Info (3)

#### I1. CHANGELOG migration guide is good but missing exit-code consumer note

**File:** `CHANGELOG.md`

Migration section covers piped output well. Consider adding a note for consumers that previously relied on `$?` returning the same code (always 1 from `cobra` errors before): now exit codes 2-6 are emitted. This is technically additive (more granular) but scripts checking `if [ $? -eq 1 ]` may miss new codes.

#### I2. `chat.go` inline error prints are correct (verified)

**File:** `cmd/chat.go:158, 200`

Both are inside the REPL loop (`for scanner.Scan()`); both use `continue` to keep the session alive. They are NOT command-level errors and do not interact with the central error handler. Implementation report correctly identified this as intentional.

#### I3. Banner gating in `cmd/logs.go` is correct

**File:** `cmd/logs.go:36-38, 72-74`

`if !quiet && IsTTY(stdout)` — both conditions correct. Banners suppressed in `--quiet` OR non-TTY mode. Matches spec security requirement: `--quiet` does not suppress security-critical messages (only banners/tips).

---

## Concurrency & Race Audit

- **FollowStream / readLoop:** Handler invoked sequentially from a single readLoop goroutine (`websocket.go:222-268`) — no concurrent handler calls. Safe.
- **errCh non-blocking send (`follow.go:84-87`):** Buffered chan size 1, `select { case errCh <- err: default: }` — only first handler error is captured; subsequent ones dropped. Acceptable: only one error needed to trigger stop.
- **Context cancel during reconnect delay:** Properly handled at `follow.go:59-63` via `select { case <-ctx.Done(): ... case <-time.After(delay): }`. No leak.
- **WSClient.done channel close:** `Close()` uses `select { case <-ws.done: default: close(ws.done) }` to avoid double-close. Safe.

---

## Security Audit

- **Error message leakage:** Server already i18n-sanitizes error messages. CLI pass-through preserves `code`/`message` only. No stack traces, no file paths leaked. PASS.
- **Exit code information disclosure:** 401 and 403 both → exit 2. Per spec security requirement (principle of least info). PASS. *Note L1*: the JSON envelope's `code` field for 403 returns `TENANT_ACCESS_REVOKED` which IS more specific — minor info leak via JSON code, not via exit code.
- **`--quiet` suppression scope:** Only suppresses banners/tips. Errors still printed via central handler. Auth failures NOT suppressed. PASS.
- **GOCLAW_OUTPUT env injection:** No shell expansion, raw string compared in `switch`. No risk. PASS.

---

## Test Coverage Audit

- `internal/output`: 97.3% — exceeds 70% target.
- `internal/client`: 71.3% — meets target. Note: HTTP error parsing path (`http.go:do()`) untested for new server envelope (M2).
- All 12 server codes covered by `TestFromError_KnownServerCode` and `TestMapServerCode_*`.
- HTTP status fallback covered by `TestMapHTTPStatus`.
- TTY precedence covered.
- FollowStream context cancel + handler stop covered (but H1 + M1 indicate test intent vs behavior mismatch).

---

## Recommended Actions (priority order)

1. **H1 (high):** Add `errHandlerStop` sentinel; do not retry on handler errors. Add assertion to `TestFollowStream_HandlerErrorStops` proving attempt count ≤ 1.
2. **M1 (medium):** Treat `MaxRetries: 0` as literal zero in `FollowConfig`; document explicitly.
3. **M2 (medium):** Add `internal/client/http_test.go` covering server error envelope, plain-text fallback, success envelope, and retry path.
4. **M3 (medium):** Validate `--output` values; reject `invalid` early with clear error.
5. **L1-L4 (low):** Polish — change 403→UNAUTHORIZED in HTTP fallback, drop misleading test row for 200, add 408 mapping, refactor duplicated TTY chain in `Execute()`.

---

## Score Breakdown

| Dimension | Score | Notes |
|---|---|---|
| Correctness | 9.0 | H1 + M1 retry semantics issues |
| Architecture | 10.0 | Clean duck-typing, no import cycle, single source of truth |
| Test coverage | 9.5 | Output excellent; client missing HTTP envelope test |
| Documentation | 9.5 | CHANGELOG + README + CLAUDE.md all updated |
| Security | 10.0 | Exit code privacy, no leakage, --quiet scope correct |
| Spec compliance | 9.5 | Step 6 deferred per spec; Step 3.3 partial (no HTTP test) |
| **Overall** | **9.4** | Auto-approve threshold is 9.5 — flag for user review (1 high finding) |

---

## Unresolved Questions

1. Should `RetryableCall` (Step 6) be implemented in P0.1 hotfix or deferred to P1+? Currently `http.go:do()` retries 3x on 5xx/429 unconditionally without honoring `ErrorShape.Retryable=false` — could retry non-retryable ops.
2. For H1 fix, should handler errors propagate verbatim to caller, or should they always map to a generic "stream interrupted" error? Spec doesn't say.
3. `TestResolveFormat_DefaultNonTTY` accepts both "json" and "table" outcomes (`tty_test.go:43-45`) — should this be tightened with a deterministic stdout pipe?
