# Phase 4 Review — Agent Lifecycle + Chat + Teams + Memory KG

**Date:** 2026-04-15
**Reviewer:** code-reviewer
**Scope:** 26 new + 4 modified files, cmd/agents_*, cmd/chat_*, cmd/teams_*, cmd/memory_*

---

## Summary

Build clean, vet clean, `go test ./...` passes (146/146 cmd tests + internal/*). Formerly flaky `internal/client` test is now stable — `WSClient.Close()` already uses `sync.Once`, contradicting the impl report's concern #1 (stale). Coverage for AI-critical commands meets the ≥80% bar in the important paths (chat extensions 85.3%, kg graph 84.6%, evolution 80.6%). A few below-80% files are non-blocking (see Informational).

Modularization executed well: `agents.go` 196, `teams.go` 150, `memory.go` 147 LoC — all under the 200-line rule. Two files sit at 214 (`chat.go`, `chat_ai_commands.go`), entirely from `--help` docstrings; documented as intentional MAX POLISH trade-off.

---

## Critical Issues (blocking)

**None.** All P0 ergonomics (tui.Confirm non-interactive block, central error bubble, SIGINT handling, FollowStream use, exit-code 6 on timeout) are correctly applied.

---

## High Priority

### H1 — `memory kg entities upsert` silently accepts malformed JSON
File: `cmd/memory_kg.go:91-116`

```go
list := unmarshalList([]byte(content))
if len(list) > 0 {
    body = list
} else {
    body = unmarshalMap([]byte(content))
}
```

`unmarshalList` / `unmarshalMap` discard errors via `_ = json.Unmarshal(...)`. If the file is malformed JSON, both return empty values and the command POSTs `nil` or `{}` without error. User sees no feedback that their file failed to parse — they see a server response that doesn't match their intent.

**Fix:** Validate with a strict decode first:
```go
var body any
if err := json.Unmarshal([]byte(content), &body); err != nil {
    return fmt.Errorf("parse entity file: %w", err)
}
```

Impact: AI-critical surface (KG upsert). MAX POLISH requirement demands explicit error cases.

### H2 — Goroutine cleanup on `agents wait` timeout path relies on OS process exit
File: `cmd/agents_lifecycle.go:104-115`

On timeout, `output.Exit(output.ExitResource)` calls `os.Exit(6)` — this skips the deferred `ws.Close()`. The spawned goroutine at line 99 keeps the WS call pending. In practice the OS reaps everything, so no user-visible leak; but on the **SIGINT** path (line 107-110) the function returns normally and defers run → goroutine unblocks cleanly via `ws.done`. Inconsistent cleanup paths; pattern-wise one Exit should mirror the other.

**Fix:** Call `ws.Close()` explicitly before `output.Exit` on the timeout branch (or drop the Exit and `return fmt.Errorf(...)` so the root error path owns the exit code via `FromError` — but that loses the deterministic 6-on-timeout contract).

Impact: Low in practice, but matters for consistency and testability (impl report flagged untestable timeout path — owning cleanup explicitly makes subprocess testing unnecessary if return-error path is taken).

---

## Medium Priority

### M1 — `teams members remove` lacks confirmation
File: `cmd/teams_members.go:59-82`

Removes a member without `tui.Confirm`. Analogous to `agents unshare` which also skips confirm (both are reversible by re-adding). Consider adding confirm for parity with `teams delete` and destructive-op consistency. Not in phase spec's required `--yes` list, so non-blocking.

### M2 — `memory kg dedup dismiss` issues no idempotency warning
File: `cmd/memory_kg_dedup.go:99-122`

Dismiss repeated with same candidate-ID — CLI returns `"Candidate X dismissed"` regardless of server's actual idempotency. Phase spec unresolved Q#5 applies. Document server behavior in `--help` once known.

### M3 — Shared-global flag mutation in tests may bleed state
Files: `cmd/agents_lifecycle_test.go`, `cmd/chat_extensions_test.go`, etc.

Tests call `cmd.Flags().Set(...)` on package-level `*Cmd` globals. Go test runs cmd package tests sequentially by default (no `t.Parallel()`), so this works today. If anyone adds `t.Parallel()` later, flag state races will flake. Add a `resetFlags(t, cmd)` helper using `t.Cleanup` or switch to a harness that constructs fresh Commands per test. Non-blocking for now.

### M4 — `chat inject` `--content` uses `MarkFlagRequired` but empty string passes
File: `cmd/chat_ai_commands.go:119-120, 206-207`

`MarkFlagRequired("content")` only checks flag presence, not value. Good defensive check `content == ""` after `readContent` covers literal empty + empty file. Code is correct; just note that `MarkFlagRequired` isn't load-bearing here — the explicit check is. Minor polish: document in help that empty files are rejected.

---

## Low Priority / Informational

- **Redundant comments.** Several files carry two back-to-back header comment blocks (e.g., `agents.go:11-25`, `memory_kg.go:10-14`) — residual from the split. Single block would be cleaner.
- **`agents instances metadata` legacy alias** (`agents_instances.go:124-150`) overlaps with `update-metadata` and has a confusing `Success` message when no flags given. Consider deprecating.
- **Coverage gaps** below 80% target (non-AI-critical):
  - `agents_lifecycle.go` 76.4% — timeout path is the uncovered portion; OK per impl report concern #3
  - `memory_kg.go` 51.1% — entities upsert path not tested (related to H1)
  - `teams_tasks_advanced.go` 67.5% — `--follow` branch untested in cmd package
  - `agents_instances.go` 56.1% — set-file, legacy metadata paths
  - `agents_episodic.go` 62.5% — add a malformed-query edge-case test
- **Timeout magic number.** `WSClient.Call` has a 30s hardcoded deadline (`websocket.go:149`) — unrelated to Phase 4 but worth a config knob for `agents wait` which can block much longer than 30s (default `--timeout=5m`). The goroutine workaround in `agents_lifecycle.go` sidesteps this, but at the cost of the double-select pattern.

---

## Focus-Area Checklist Results

| # | Check | Status |
|---|-------|--------|
| 1 | AI-critical `--help` JSON schemas | PASS — chat history/inject/session-status, agents wait/identity, kg traverse/graph all document response schema in Long text |
| 2 | File size ≤200 LoC | PASS — only chat.go + chat_ai_commands.go at 214 (docstring overage, documented intentional) |
| 3 | `agents wait` timeout | PASS — `signal.NotifyContext` + `timeoutCtx` + `output.Exit(ExitResource)` on timeout, graceful nil-return on SIGINT |
| 4 | `--follow` SIGINT | PASS — `teams_tasks_advanced.go:119-135` and `teams_events.go:46-62` both use `signal.NotifyContext` + `FollowStream` + `ctx.Err() != nil → return nil` |
| 5 | Destructive ops require `--yes` | PASS — `kg entities delete`, `kg merge`, `teams tasks delete`, `teams tasks delete-bulk`, `teams workspace delete`, `memory delete` all gated through `tui.Confirm(..., cfg.Yes)` |
| 6 | `tui.Confirm` non-interactive safety | PASS — `internal/tui/prompt.go:27-30` refuses without `--yes` in non-TTY mode, prints clear stderr message |
| 7 | Extract preserves behavior | PASS — extracted agents_sharing/lifecycle/instances/links, teams_members/tasks/workspace/events/scopes, memory_kg/dedup/graph/legacy/index all register cleanly into parent groups; all one-shot paths exercised by tests |
| 8 | Central error bubble, no double-print | PASS — `root.go:45 SilenceErrors: true`, `Execute()` uses `output.PrintError` + `output.Exit(FromError(err))`; RunE funcs return errors without calling Print themselves (except `agents wait` timeout which intentionally exits with code 6) |

---

## Positive Observations

- Consistent idiom: all WS callers use `newWS("cli")` + `defer ws.Close()`.
- MAX POLISH help text on `chat history/inject/session-status`, `agents wait/identity` is genuinely useful — JSON schemas + examples + security notes on `chat inject`.
- `agents wait` combines signal context + timeout context correctly and distinguishes SIGINT (return nil) from timeout (exit 6).
- Exit-code contract (`internal/output/exit.go`) is a clean, documented map; `FromError` prefers structured server codes over HTTP status fallback.
- `tui.Confirm` non-interactive behavior (refuses without `--yes` in non-TTY) is a systemic P0 fix that Phase 4 commands inherit transparently.
- Dedup workflow (`scan → list → merge|dismiss`) is complete and coherent.
- `--follow` commands pass `nil` for `FollowConfig` — correct default (5 retries, 1s base delay).

---

## Recommended Actions

1. **H1 (high):** Fix silent-parse in `memory kg entities upsert` — strict `json.Unmarshal` with error return.
2. **H2 (medium-high):** Close WS before `os.Exit` on `agents wait` timeout path; or return an error and let root's `FromError` map it.
3. **M1-M4:** Optional polish — confirmation on `teams members remove`, deprecate legacy `agents instances metadata` alias, raise coverage on `memory_kg.go` (upsert) and `teams_tasks_advanced.go` (follow).
4. **Doc:** Note in impl report that `WSClient.Close` is already `sync.Once`-guarded — the flaky-race concern is resolved (tests pass under `go test ./...`).

---

## Metrics

- Build: PASS (`go build ./...`)
- Vet: PASS (`go vet ./...`)
- Tests: PASS (146 cmd + all internal)
- Coverage (cmd pkg overall): 33.4% — normal because many commands have no tests; AI-critical surface is what matters
- AI-critical coverage (target ≥80%):
  - chat_ai_commands: 85.3% — PASS
  - memory_kg_graph: 84.6% — PASS
  - agents_evolution: 80.6% — PASS
  - agents_lifecycle: 76.4% — near target, timeout path intentionally uncovered
  - memory_kg (entities): 51.1% — below target, upsert path untested
  - teams_tasks_advanced: 67.5% — below target, follow path untested

---

## Unresolved Questions

1. Should `teams members remove` gain `--yes` confirmation for parity? (phase spec did not require it)
2. Server behavior on `kg dedup dismiss` re-run with same candidate-ID — silent success or 409? (spec unresolved Q#5)
3. Should `agents wait` timeout path close the WS before `os.Exit` or convert to an error-return so the root error path owns exit code mapping?

---

**Status:** DONE_WITH_CONCERNS
**Score:** 8.6/10
**Critical count:** 0 (2 high-priority, 4 medium, ~6 informational)
