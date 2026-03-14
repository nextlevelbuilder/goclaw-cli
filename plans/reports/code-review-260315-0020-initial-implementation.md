# Code Review: GoClaw CLI - Initial Implementation

**Date:** 2026-03-15
**Reviewer:** code-reviewer
**Scope:** Full codebase (28 Go files, ~5600 LOC)

---

## Code Review Summary

### Scope
- **Files:** 28 Go source files (21 cmd/, 4 internal/client/, 1 internal/config/, 1 internal/output/, 1 internal/tui/)
- **LOC:** ~5,600
- **Focus:** Full codebase review
- **Dependencies:** cobra, gorilla/websocket, golang.org/x/term, gopkg.in/yaml.v3

### Overall Assessment

Well-structured Cobra CLI with clean separation of concerns. Code is idiomatic Go, consistently formatted, and covers extensive API surface. The architecture is sound for a v1 CLI tool. Several security, reliability, and edge-case issues need attention before production use.

---

## Critical Issues

### C1. Token stored in plaintext config file (Security)
**File:** `internal/config/config.go` lines 24-30, `cmd/auth.go` line 154-161
- `Profile` struct stores `Token` in YAML and writes to `~/.goclaw/config.yaml`
- Token also saved separately in `CredentialStore` (file-based), creating dual storage
- README claims "OS keyring credential storage" but no keyring integration exists -- tokens are plaintext files with 0600 permissions
- **Impact:** Token theft via file read on shared systems; misleading security documentation
- **Fix:** Either integrate `zalando/go-keyring` or `99designs/keyring`, or at minimum remove the misleading claim from README. The `CredentialStore` has correct file permissions (0600/0700) but the config.yaml also stores tokens.

### C2. Token passed via `--token` flag visible in process list (Security)
**File:** `cmd/root.go` line 45
- `pf.String("token", "", "Auth token (env: GOCLAW_TOKEN)")` -- the token value appears in `ps aux` output
- README claims "no secrets in ps" but this is not enforced
- **Impact:** Token leakage on multi-user systems
- **Fix:** Accept token only via env var (`GOCLAW_TOKEN`) or stdin, not as a flag. If flag must remain, document the risk clearly.

### C3. Path traversal in file-related endpoints (Security)
**File:** `cmd/memory.go` line 48, `cmd/storage.go` lines 24-26, `cmd/agents.go` line 376
- User-supplied path arguments are concatenated directly into URL paths without sanitization
- Example: `"/v1/memory/" + args[0] + "/" + args[1]` -- args could contain `../` sequences
- Storage uses `url.PathEscape` in some places but not all
- **Impact:** Potential server-side path traversal (depends on server validation)
- **Fix:** Always use `url.PathEscape()` on user-supplied path segments. Add `path.Clean()` before encoding.

### C4. WebSocket token sent in handshake payload, not headers (Security)
**File:** `internal/client/websocket.go` lines 82-83, 92-98
- HTTP headers are empty: `header := http.Header{}`
- Token sent inside JSON payload of `connect` method
- **Impact:** Token appears in WebSocket message payloads, potentially logged by proxies/servers at message level rather than header level
- **Fix:** Also send token as `Authorization` header in the WebSocket dial, matching HTTP client behavior. Keep payload token for protocol compatibility.

---

## High Priority

### H1. Silently swallowed JSON unmarshal errors throughout codebase
**File:** `cmd/helpers.go` lines 39-49, used everywhere
- `unmarshalList` and `unmarshalMap` both discard errors: `_ = json.Unmarshal(data, &list)`
- When server returns unexpected format, user sees empty output with no indication of failure
- Also in chat streaming: `_ = json.Unmarshal(e.Payload, &chunk)` (chat.go lines 104, 110, 185, 190)
- **Impact:** Silent data loss; debugging becomes impossible
- **Fix:** Return error from helpers, or at minimum log a warning in verbose mode. For streaming callbacks where errors can't be returned, print to stderr.

### H2. WebSocket listener leak in Stream()
**File:** `internal/client/websocket.go` lines 160-177
- `Stream()` calls `ws.Subscribe()` for chunk, tool.call, tool.result, run.started, run.completed
- Listeners are appended to `ws.listeners` map but never removed after stream completes
- Each call to `chatInteractive` loop adds 5 more listeners per message sent
- **Impact:** Memory leak and duplicate event handling in long interactive sessions
- **Fix:** Add `Unsubscribe` method or return a cleanup function from `Subscribe`. Clean up in `Stream` via defer.

### H3. HTTP retry with consumed request body
**File:** `internal/client/http.go` lines 140-157
- On retry (429/5xx), request body is re-created but `req` itself is reused
- `req.GetBody` is not set, so the `http.Client` cannot replay the body on redirects
- On first retry attempt after body consumption, `req.Body` may already be closed/read
- Line 154 re-assigns `req.Body` which is correct for the custom retry loop, but the initial `req` created on line 127 uses a `bytes.NewReader` which is consumed after first `Do()`
- **Impact:** Retries may send empty body on POST/PUT/PATCH requests
- **Fix:** Use `bytes.NewReader` and call `Seek(0, 0)` on retry, or re-create the full request each attempt.

### H4. `select {}` blocks forever with no signal handling (logs tail)
**File:** `cmd/logs.go` line 67
- `select {}` blocks the goroutine indefinitely; Ctrl+C will kill the process without cleanup
- WebSocket `defer ws.Close()` on line 24 will never execute
- **Impact:** Unclean shutdown, potential resource leak
- **Fix:** Use `signal.Notify` with `os.Interrupt` and `syscall.SIGTERM`, then clean up.

### H5. Concurrent WebSocket write without mutex
**File:** `internal/client/websocket.go` line 133
- `ws.conn.WriteJSON(req)` is called from `Call()` which can be invoked from multiple goroutines (e.g., interactive chat + abort)
- gorilla/websocket docs state: "Connections support one concurrent reader and one concurrent writer"
- No write mutex protects `WriteJSON`
- **Impact:** Panic or corrupted WebSocket frames under concurrent writes
- **Fix:** Add a write mutex or use a dedicated write goroutine with a channel.

### H6. `run.completed` channel close panic on multiple events
**File:** `internal/client/websocket.go` line 176
- `close(done)` inside a listener callback -- if `run.completed` fires more than once (e.g., server bug or reconnect), this panics
- **Impact:** CLI crash on unexpected server behavior
- **Fix:** Use `sync.Once` for the close operation.

---

## Medium Priority

### M1. `buildBody` silently drops zero-value integers
**File:** `cmd/helpers.go` lines 72-95
- `case int: if v != 0 { body[key] = v }` -- intentionally setting a value to 0 (e.g., budget=0, timeout=0) is silently dropped
- Same for empty string case -- cannot explicitly set a field to empty
- **Impact:** Cannot reset numeric fields to zero via CLI
- **Fix:** Use `cmd.Flags().Changed()` pattern (already used in update commands) instead of zero-value filtering. Or accept pointer types.

### M2. `FindProfile` returns pointer to copy
**File:** `internal/config/config.go` lines 154-161
- `for _, p := range fc.Profiles` creates a copy; `return &p` returns pointer to loop variable
- In Go, this is safe (loop variable address is stable per iteration in modern Go), but it returns a pointer to a copy -- mutating the returned profile won't affect the original slice
- **Impact:** Confusing semantics, though currently only used read-only

### M3. Duplicated WS connect boilerplate
**File:** `cmd/cron.go`, `cmd/teams.go`, `cmd/config_cmd.go`, `cmd/admin.go`
- Every WS-based command repeats: `newWS("cli") -> Connect() -> defer Close()`
- 30+ occurrences of identical 8-line boilerplate
- **Impact:** Maintenance burden, inconsistency risk
- **Fix:** Create `newConnectedWS()` helper that returns a connected client + error.

### M4. Media upload is unimplemented
**File:** `cmd/admin.go` lines 333-346
- `mediaUploadCmd` prints a message but does not actually upload
- `_ = c` silences unused variable warning -- dead code
- **Impact:** Incomplete feature advertised in README
- **Fix:** Implement multipart upload (pattern already exists in `cmd/skills.go` `skillsUploadCmd`).

### M5. `cronListCmd` creates HTTP client but never uses it
**File:** `cmd/cron.go` lines 18-21
- Creates both HTTP and WS client; `_ = c` suppresses unused warning
- HTTP client creation may fail if token is missing, masking the real WS error
- **Impact:** Confusing error messages, wasted resources
- **Fix:** Remove unused HTTP client creation.

### M6. Config file race condition on concurrent CLI invocations
**File:** `internal/config/config.go` lines 99-122, 125-143
- `Save` and `RemoveProfile` read-then-write without file locking
- Running `goclaw auth login` concurrently from two terminals can corrupt config
- **Impact:** Config file corruption in CI/automation scenarios
- **Fix:** Use file locking (`flock` on Unix, `LockFileEx` on Windows) or atomic write with rename.

### M7. No input validation on server URL format
**File:** `cmd/auth.go` lines 118-135
- Server URL accepted without validation -- no check for scheme (http/https), no URL parsing
- Could lead to confusing errors downstream
- **Fix:** Parse with `url.Parse()` and validate scheme is http/https.

### M8. `insecure` flag not gated by `Changed()`
**File:** `internal/config/config.go` lines 91-93
- `cfg.Insecure, _ = cmd.Flags().GetBool("insecure")` -- always reads the flag default
- Unlike server/token/output which respect `Changed()`, insecure always overwrites any config-file value
- Same issue with `verbose` and `yes` flags
- **Impact:** Cannot set `insecure: true` in config file -- always overridden by flag default `false`
- **Fix:** Use `if cmd.Flags().Changed("insecure")` pattern.

---

## Low Priority

### L1. `jsonToMap` duplicates `unmarshalMap`
**File:** `cmd/status.go` lines 79-85 vs `cmd/helpers.go` lines 45-49
- Identical logic with different function name
- **Fix:** Use `unmarshalMap` instead.

### L2. Version command uses `Run` instead of `RunE`
**File:** `cmd/version.go` line 20
- All other commands use `RunE` for consistency. Not a bug since version never errors.

### L3. No shell completion generation
- Cobra supports `completion` command generation. Adding this would improve UX significantly.
- **Fix:** `rootCmd.AddCommand(cobra.GenBashCompletionV2Cmd()...)` or custom completion commands.

### L4. `readContent` has no size limit
**File:** `cmd/helpers.go` lines 60-69
- `os.ReadFile(val[1:])` reads entire file into memory with no size guard
- **Impact:** OOM on very large files
- **Fix:** Add reasonable size limit or stream large files.

### L5. No user-agent header on HTTP requests
**File:** `internal/client/http.go`
- No `User-Agent` header set. Server cannot distinguish CLI traffic from other clients.
- **Fix:** Set `User-Agent: goclaw-cli/<version>`.

### L6. Contacts resolve does not URL-encode IDs
**File:** `cmd/channels.go` line 154
- `"/v1/contacts/resolve?ids=" + args[0]` -- if IDs contain special chars, query breaks
- **Fix:** Use `url.Values{}.Encode()`.

---

## Edge Cases Found by Scout

1. **Interactive session scanner overflow:** `bufio.NewScanner` in chat.go uses default 64KB buffer. Very long pastes will silently truncate input.
2. **Pairing timeout not configurable:** Hard-coded 60 iterations x 2s = 2 min max. Slow networks or manual approval workflows may exceed this.
3. **Ctrl+C during password prompt:** `term.ReadPassword` may leave terminal in raw mode if interrupted.
4. **Empty profile name edge case:** If user passes `--profile ""`, code falls through to `"default"` in some places but not all.
5. **WebSocket reconnection:** No automatic reconnection logic. If connection drops during interactive chat, user gets a cryptic error and must restart.
6. **`io.Copy` error ignored:** `cmd/admin.go` line 369: `n, _ := io.Copy(f, resp.Body)` -- write errors silently ignored for media download.

---

## Positive Observations

1. **Clean Cobra structure** -- Command hierarchy is well-organized with logical grouping
2. **Config precedence** -- flags > env > file is correct and well-implemented
3. **Consistent patterns** -- CRUD commands follow uniform structure across all resources
4. **TLS by default** -- `--insecure` must be explicitly opted into
5. **Confirmation prompts** -- Destructive operations require confirmation (or `--yes`)
6. **Retry logic** -- HTTP client retries on 429/5xx with exponential backoff
7. **Build metadata** -- Version/commit/date injected via ldflags
8. **Minimal dependencies** -- Only 4 direct dependencies, all well-maintained
9. **File permissions** -- Credential files use 0600, directories use 0700
10. **Dual output modes** -- Table for humans, JSON/YAML for automation

---

## Recommended Actions (Priority Order)

1. **[Critical]** Remove token from config.yaml storage; use only CredentialStore. Load token from store during config.Load().
2. **[Critical]** Add `url.PathEscape()` to all user-supplied URL path segments.
3. **[Critical]** Add write mutex to WebSocket client for `WriteJSON`.
4. **[High]** Fix listener leak in `Stream()` with cleanup mechanism.
5. **[High]** Add signal handling to `logs tail` command.
6. **[High]** Fix HTTP retry body consumption issue.
7. **[High]** Use `sync.Once` for `run.completed` channel close.
8. **[High]** Surface JSON unmarshal errors (at least in verbose mode).
9. **[Medium]** Extract `newConnectedWS()` helper to reduce boilerplate.
10. **[Medium]** Implement media upload command.
11. **[Medium]** Fix `insecure`/`verbose`/`yes` flag precedence with `Changed()` guard.
12. **[Low]** Add shell completion support.
13. **[Low]** Add `User-Agent` header.

---

## Metrics

| Metric | Value |
|--------|-------|
| Type Coverage | N/A (Go is statically typed) |
| Test Coverage | 0% (no test files found) |
| Linting Issues | Not run (no `go vet` output) |
| Files > 200 LOC | 5 (agents.go:492, teams.go:501, admin.go:404, mcp.go:340, channels.go:281) |
| Security Issues | 4 critical, pattern-level |
| Missing Features | media upload, shell completion, keyring integration |

---

## Unresolved Questions

1. Does the GoClaw server validate path segments server-side, or is client-side sanitization the only defense against path traversal?
2. Is the WebSocket `connect` handshake the only auth mechanism, or does the server also check HTTP upgrade headers?
3. README lists commands (knowledge-graph, usage, approvals, delegations, credentials, tts, media, activity) that are implemented but some are incomplete (media upload) -- is this intentional for v1?
4. The `--pair` flow uses WebSocket for pairing but saves no token -- how does paired auth work for subsequent HTTP-based commands?
5. Are there plans for test coverage? The `make test` target exists but there are zero test files.
