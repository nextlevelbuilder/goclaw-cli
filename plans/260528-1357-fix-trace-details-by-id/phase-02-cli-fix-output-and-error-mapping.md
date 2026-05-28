---
phase: 2
title: 'CLI fix: output and error mapping'
status: completed
priority: P2
effort: 2.5-3h
dependencies:
  - 1
---

# Phase 2: CLI fix: output and error mapping

## Overview

Turn the red test from Phase 1 green. Add 7 more tests (rendering, error categories, malformed-id validation). Then implement: human-readable render path inline in `tracesGetCmd.RunE`, checked unmarshal, strict id validation (blocks path traversal), and reliance on existing HTTP-status fallback for the four AC#3 error categories.

## Locked decisions inherited from plan.md

- **No `MapServerCode` extension.** 404→ExitNotFound, 403→ExitAuth, 5xx→ExitServerError already resolve via [cmd/helpers.go:152-172](cmd/helpers.go:152) → [internal/output/exit.go:46+](internal/output/exit.go:46) HTTP-status fallback.
- **Inline render in `cmd/traces.go`**, no `cmd/traces_render.go`. Cap render at ~50 LOC.
- **Inline `json.Unmarshal` + error check.** No new helper.
- **Strict id allowlist:** `^[A-Za-z0-9._-]+$` AND non-empty AND not `.` AND not `..`. Then `url.PathEscape` on top.
- **5xx test allows `calls >= 1`** because `internal/client/http.go` retries 3x.

## Requirements

**Functional**
- Table/TTY mode renders a structured trace summary: header card (trace_id, agent_id, status, duration, tokens, cost — whichever Phase-1 fixture confirms exists) + span tree (via `output.PrintTree`) + flat events list (`EVENTS (n=N):` header + one line per event, no truncation).
- JSON/YAML mode emits the decoded payload via the existing printer path — no field dropping by us. (Note: `json.NewEncoder` re-encodes with HTML escape + key reorder; "preserves nested fields" means structural preservation, not byte-for-byte.)
- Empty/malformed server response returns a wrapped error, not a silent empty `{}`.
- Malformed id (empty, whitespace, `..`, `/`, `\\`, control char, non-allowlist) returns validation error BEFORE the HTTP call. Exit 4.
- 4 error categories produce distinguishable messages and exit codes per AC#3 via existing HTTP-status fallback.

**Non-functional**
- All render logic stays inside `tracesGetCmd.RunE` (~50 LOC). No new file. Pattern matches `tracesListCmd` ([cmd/traces.go:30-59](cmd/traces.go:30)).
- Reuse `output.PrintTree` / `output.TreeNode` ([internal/output/tree.go:7](internal/output/tree.go:7)); no parallel tree renderer.
- No new dependencies.

## Architecture

```
goclaw traces get <id>
  |
  v
[validate id: regex allowlist + reject ./../ /]
  |--invalid--> return validation error -> exit 4
  v
GET /v1/traces/{url.PathEscape(id)}
  |
  +-- APIError? ---> central handler via output.FromError
  |                     |
  |                     +-- StatusCode 404 -> exit 3 (via MapHTTPStatus)
  |                     +-- StatusCode 403 -> exit 2 (via MapHTTPStatus)
  |                     +-- StatusCode 5xx -> exit 5 (via MapHTTPStatus, after 3 retries)
  |                     +-- StatusCode 400 -> exit 4 (server validation, if ever returned)
  |
  v
json.Unmarshal(data, &trace) — checked, NOT silent
  |--err--> wrap "decode trace payload" -> exit 5
  v
switch cfg.OutputFormat:
  json|yaml -> printer.Print(trace)
  table     -> inline render: header lines + PrintTree(spans) + events list
```

## Related Code Files

- Modify: `cmd/traces.go` — `tracesGetCmd.RunE` rewritten with validation, checked unmarshal, inline render. Net add ~50-70 LOC.
- Modify: `cmd/traces_get_test.go` — add 7 new tests on top of Phase 1's 3.
- No new files.
- No changes to `internal/output/exit.go` or `internal/output/error.go`.

## Implementation Steps

### 1. Extend test file (red first)

Add these 7 tests to `cmd/traces_get_test.go`:

```go
// 4. Table render must include header card + span markers.
func TestTracesGet_TableMode_HasHeaderAndSpanMarkers(t *testing.T) {
    // serve fixture; runCmd with --output table
    // assert stdout contains "TRACE_ID" header AND tree marker ("├─" or "└─" or "└" or "├")
    // assert stdout does NOT start with '{'
}

// 5. JSON mode preserves structural keys observed in fixture.
func TestTracesGet_JSONMode_PreservesStructure(t *testing.T) {
    // serve fixture; runCmd with --output json
    // parse stdout as JSON; assert every top-level key present in fixture is present in output
    // assert nested key (e.g. spans[0].name) reachable
}

// 6. 404 → exit 3 + message contains "not found".
func TestTracesGet_NotFound_ExitCode3(t *testing.T) {
    // server returns 404 + {"ok":false, "error":{"code":"<observed code>", "message":"trace not found"}}
    // assert err non-nil; output.FromError(err) == output.ExitNotFound (3)
    // assert err.Error() contains "not found" (case-insensitive)
}

// 7. 403 → exit 2.
func TestTracesGet_PermissionDenied_ExitCode2(t *testing.T) {
    // server returns 403; assert output.FromError(err) == output.ExitAuth (2)
}

// 8. Malformed id rejected client-side; no HTTP call.
func TestTracesGet_MalformedID_NoHTTPCall(t *testing.T) {
    // record calls counter on httptest server
    // try ids: "", "  ", "..", "../etc/passwd", "a/b", "a\\b", "a\x00b"
    // each should: return err, exit code 4, AND not increment calls
}

// 9. 5xx → exit 5 (allow retries; do not assert call count).
func TestTracesGet_ServerError_ExitCode5(t *testing.T) {
    // server returns 500 always
    // assert err non-nil; output.FromError(err) == output.ExitServerError (5)
    // assert calls counter >= 1 (NOT == 1; retry adds calls)
}

// 10. Malformed JSON response surfaces decode error, not silent empty.
func TestTracesGet_MalformedResponse_SurfacesError(t *testing.T) {
    // server returns 200 + body "this is not json"
    // assert err non-nil; err.Error() contains "decode" or "unmarshal"
    // assert stdout is empty or contains the error, NOT a literal "{}"
}
```

**Test harness rules:**
- Every table-mode test passes `--output table` explicitly.
- Every JSON-mode test passes `--output json` explicitly.
- Use `httptest.NewServer` + `okJSON(t, w, payload)` envelope helper.
- Use `t.Setenv` to point CLI at the test server.

### 2. Implement `tracesGetCmd.RunE` rewrite

Replace [cmd/traces.go:61-75](cmd/traces.go:61) with (sketch):

```go
var tracesGetCmd = &cobra.Command{
    Use:   "get <traceID>",
    Short: "Get trace with span tree",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        id := strings.TrimSpace(args[0])
        if err := validateTraceID(id); err != nil {
            return err  // returns ExitValidation via output.FromError
        }
        c, err := newHTTP()
        if err != nil {
            return err
        }
        data, err := c.Get("/v1/traces/" + url.PathEscape(id))
        if err != nil {
            return err  // APIError -> central handler maps via HTTP status
        }
        var trace map[string]any
        if err := json.Unmarshal(data, &trace); err != nil {
            return fmt.Errorf("decode trace payload: %w", err)
        }
        format := output.ResolveFormat(cfg.OutputFormat)
        if format == "json" || format == "yaml" {
            printer.Print(trace)
            return nil
        }
        return renderTraceTable(trace)  // inline below in same file
    },
}

// validateTraceID enforces the strict allowlist (rejects path-traversal etc.)
func validateTraceID(id string) error {
    if id == "" || id == "." || id == ".." {
        return validationErr("trace id is empty or invalid")
    }
    if !traceIDRegex.MatchString(id) {
        return validationErr("trace id contains invalid characters (allowed: A-Z a-z 0-9 . _ -)")
    }
    return nil
}
var traceIDRegex = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// renderTraceTable prints header + span tree + events. ~30-40 LOC.
func renderTraceTable(t map[string]any) error {
    // defensive helpers — define inline or in cmd/helpers.go
    s := func(k string) string { v, _ := t[k].(string); return v }
    // header lines
    fmt.Printf("TRACE_ID: %s\nAGENT_ID: %s\nSTATUS: %s\n", s("trace_id"), s("agent_id"), s("status"))
    // span tree via output.PrintTree — adapt spans field per Phase-1 fixture shape
    if spans, ok := t["spans"].([]any); ok && len(spans) > 0 {
        root := buildSpanTree(spans)  // helper, ~15 LOC
        fmt.Println()
        output.PrintTree(root, "")
    } else {
        fmt.Println("(no spans)")
    }
    // events flat list
    if evs, ok := t["events"].([]any); ok {
        fmt.Printf("\nEVENTS (n=%d):\n", len(evs))
        for _, e := range evs {
            if m, ok := e.(map[string]any); ok {
                fmt.Printf("  - %v\n", m["type"])  // or whatever key Phase-1 fixture confirms
            }
        }
    }
    return nil
}
```

**`validationErr` helper**: produce an error that `output.FromError` maps to `ExitValidation` (4). If no such helper exists today, add a 2-line one (e.g. wrap in `APIError{Code: "INVALID_REQUEST"}` since `MapServerCode` covers that → exit 4). Verify via Phase 1 reading of `internal/output/exit.go` what the simplest path is.

**`buildSpanTree`**: only define IF Phase 1 fixture confirms spans are hierarchical or have `parent_id`. If flat with no parent relationship, just render a flat list and skip the tree. **Do not guess the shape; let the fixture drive it.**

### 3. Run tests until green

```bash
go test -count=1 ./cmd/... -run TestTracesGet
go vet ./...
go build ./...
```

### 4. Live smoke (manual)

Re-run the four `goclaw traces get` invocations from Phase 1's smoke probe. Confirm:
- TTY: readable header + span tree + events
- `-o json`: round-trips through `jq` with all fields intact
- 404 id: exit 3, message contains "not found"
- 403 id (if reachable): exit 2

## Success Criteria

- [ ] All 3 Phase-1 tests + 7 Phase-2 tests pass under `go test -count=1`.
- [ ] `go vet ./...` clean.
- [ ] `go build ./...` clean.
- [ ] `tracesGetCmd.RunE` + `renderTraceTable` + `validateTraceID` + `buildSpanTree` together ≤ 100 LOC of new code in `cmd/traces.go` (excluding doc comments).
- [ ] Manual smoke: `goclaw traces get <real-id>` in TTY renders summary + span tree + events; `-o json` round-trips through `jq`.
- [ ] `output.FromError` returns 0/2/3/4/5 correctly for the categories per test assertions.
- [ ] Path-traversal attempt (`goclaw traces get ../etc/passwd`) returns exit 4 without making an HTTP request.
- [ ] No plan-artifact references in production or test code (no "Phase 2", finding codes, etc.).

## Risk Assessment

- **Risk:** Phase-1 fixture reveals a span shape that doesn't match the hierarchical-or-flat assumption. **Mitigation:** `renderTraceTable` degrades gracefully (`(no spans)` line) if structure is unrecognized. Tree renderer only invoked if shape is recognizable.
- **Risk:** server returns a 4xx that isn't 400/403/404 (e.g. 422). **Mitigation:** `MapHTTPStatus` likely already covers it via `4xx → ExitGeneric` or similar; verify in Phase 1 step 4 reading of `cmd/helpers.go:152-172`.
- **Risk:** `renderTraceTable` becomes a kitchen-sink. **Mitigation:** ≤100 LOC cap; defer richer rendering (timeline, ASCII chart, color) to a future enhancement.

## Security Considerations

- **Path traversal blocked client-side** via `validateTraceID` allowlist. Even if a downstream proxy mishandles encoded `..`, the CLI refuses to send it.
- **ANSI escape injection via span names.** Span `name` fields may contain LLM/tool output — potentially attacker-controlled. Current `traces list` ([cmd/traces.go:30-59](cmd/traces.go:30)) already renders user-controlled fields (`agent_id`, `status`) without sanitization, so structural parity exists. Span/event payloads are a **materially higher risk surface** because content is less constrained. Accepted-trade-off OR add `strings.Map(safeRune, v)` filter — **decide based on Phase-1 fixture inspection** (does the fixture have any non-printable bytes in span names?). If yes, add the filter; if no, accept and document.
- **404/403 existence oracle.** AC#3 explicitly requires the distinction. Trade-off accepted, documented here. Mitigation belongs to the server (could randomize a small delay on 404 to defeat timing oracle), not the CLI.
- **JSON full-payload exposure.** If server payload includes internal fields, `-o json` now shows them where table mode used to hide them via the broken render. This is correct behavior, not a regression — issue #17 explicitly wants both modes working. Documented.
- **No new auth surface.** Existing `newHTTP()` token handling unchanged.

## Next Steps

Phase 3 updates docs and ships.
