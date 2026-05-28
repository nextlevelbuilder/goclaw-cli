---
phase: 1
title: Repro and root cause
status: completed
priority: P2
effort: 1.5-2h
dependencies: []
---

# Phase 1: Repro and root cause

## Overview

Lock the failure mode with a deterministic repro and a captured live-gateway response. Classify root cause as CLI-side, server-side, or hybrid — that classification drives Phase 2's scope. No production code changes in this phase. Trimmed scope: 3 red tests here; the remaining 7 (rendering, error categories, malformed-id) land in Phase 2 step 1 to avoid testing payload shape before fixture is captured.

## Requirements

**Functional**
- 3 red repro tests in `cmd/traces_get_test.go` that mirror the actual broken behavior, using `httptest.NewServer` + the captured envelope.
- A captured fixture JSON in `cmd/testdata/trace_detail_get.json` preserving the exact server envelope shape, scrubbed of secrets.
- Written classification (CLI-side / server-side / hybrid) in `reports/repro-260528-issue17.md` with file:line evidence.

**Non-functional**
- Test file follows repo conventions (see `cmd/traces_follow_test.go`, `cmd/traces_list_test.go` for reference): `runCmd(t, ...)`, `okJSON(t, w, payload)` envelope helper.
- Smoke probe is reproducible: documented `curl -H` command in the report with `$TOKEN` env var indirection — token MUST NOT appear in committed artifacts or shell history.

## Architecture

```
+--------------------+        +----------------+        +--------------+
| traces get <id>    |  ----> | newHTTP().Get  |  ----> | GET /v1/...  |
| (cobra RunE)       |        |  (envelope)    |        | { ok, payload}|
+--------------------+        +----------------+        +--------------+
         |                            |
         v                            v
  unmarshalMap(data)            (silent err swallow)  <-- confirmed bug
         |
         v
  printer.Print(map)
         |
         v
  table -> JSON fallback    <-- "unusable output" symptom confirmed
```

**Investigation matrix** (each row = a probable failure mode):

| # | Hypothesis | Evidence | Disposition |
|---|---|---|---|
| H1 | Table mode dumps raw JSON, looks "broken" to user | VERIFIED via [internal/output/output.go:30-37](internal/output/output.go:30) — JSON fallback | CLI-side; Phase 2 fixes render |
| H2 | `unmarshalMap` swallows error on malformed payload | VERIFIED via [cmd/helpers.go:48-53](cmd/helpers.go:48) — `_ = json.Unmarshal` | CLI-side; Phase 2 fixes |
| H3 | No `url.PathEscape` and no id-validation in `tracesGetCmd` | VERIFIED via [cmd/traces.go:61-75](cmd/traces.go:61) | CLI-side; Phase 2 adds validation |
| H4 | Server returns trace under unexpected envelope key | UNVERIFIED — capture in smoke probe | Determines Phase-2 unmarshal path |
| H5 | Server returns 404 for tenant-scoped trace IDs (auth filter) | UNVERIFIED — capture in smoke probe | Determines whether AC#3 needs server-side fix |

## Related Code Files

- Read only: `cmd/traces.go`, `cmd/helpers.go`, `internal/output/output.go`, `internal/output/exit.go`, `internal/client/http.go`.
- Reference: `cmd/traces_follow_test.go` (for `runCmd` + `okJSON` patterns), `cmd/testdata/` (will be created — no prior fixtures exist in repo).
- Create: `cmd/traces_get_test.go` — 3 red tests.
- Create: `cmd/testdata/trace_detail_get.json` — captured fixture (scrubbed).
- Create: `plans/260528-1357-fix-trace-details-by-id/reports/repro-260528-issue17.md` — smoke report with curl command, response sample (scrubbed), classification verdict.

## Implementation Steps

### 1. Smoke probe against live gateway

Use `goclaw.zuey.me` or local dev gateway. **Token MUST come from env, never inline:**

```bash
export TOKEN="$(goclaw config get token 2>/dev/null || cat ~/.goclaw/token)"
export SERVER="https://goclaw.zuey.me"   # or local
# pick a trace id
goclaw traces list --limit 5 -o json | jq -r '.payload[0].trace_id' > /tmp/trace_id
TRACE_ID="$(cat /tmp/trace_id)"
# capture wire-level shape
curl -sS -H "Authorization: Bearer $TOKEN" "$SERVER/v1/traces/$TRACE_ID" > /tmp/trace_raw.json
# capture CLI shape (both modes)
goclaw traces get "$TRACE_ID" -o json > /tmp/trace_json.txt
goclaw traces get "$TRACE_ID" > /tmp/trace_tty.txt  # demonstrates the "unusable" output
# capture error envelopes
goclaw traces get bogus-id-12345 -o json > /tmp/trace_404.txt 2>&1
```

**Capture in the repro report:**
- Server response (scrubbed)
- CLI table-mode output (the visible bug)
- CLI JSON-mode output
- 404 error envelope shape — note the exact `error.code` string (e.g. `NOT_FOUND`, `TRACE_NOT_FOUND`, `RESOURCE_NOT_FOUND` — whichever the server actually returns).
- If reachable, also probe `permission denied` by attempting a trace from another tenant (server returns 403 + some code — capture the literal `error.code` string).

**Gateway unreachable escape:** if no gateway is reachable, fall back to a hand-crafted minimal fixture derived from `traces follow` response shape ([cmd/traces_follow_test.go](cmd/traces_follow_test.go)). Mark the fixture and the report with `TODO: refresh after live smoke probe`. Phase 2 must refresh before merge.

### 2. Save fixture (scrubbed)

Strip secrets and PII from `/tmp/trace_raw.json` with this concrete `jq` recipe (extend as needed per real shape):

```bash
jq '
  walk(
    if type == "object" then
      with_entries(
        if .key | test("token|secret|api_key|authorization"; "i") then
          .value = "REDACTED"
        elif .key == "user_id" then .value = "user_REDACTED"
        elif .key == "tenant_id" then .value = "tenant_REDACTED"
        elif .key == "trace_id" then .value = "trace_FIXTURE_001"
        elif .key == "agent_id" then .value = "agent_FIXTURE_001"
        elif .key == "session_key" then .value = "session_FIXTURE_001"
        elif .key | test("email|phone|address"; "i") then .value = "REDACTED"
        elif .key | test("content|message|prompt|response|tool_args|tool_result|query"; "i") then .value = "REDACTED"
        else .
        end
      )
    else .
    end
  )
' /tmp/trace_raw.json > cmd/testdata/trace_detail_get.json
```

After writing, **manually scan** `cmd/testdata/trace_detail_get.json` for any residual:
- bearer tokens (`grep -i 'eyJ\|Bearer\|sk-\|token=' cmd/testdata/trace_detail_get.json` — must return 0 lines)
- numeric ids that look like real user/tenant ids
- span names that embed user input (sanitize manually if so)

### 3. Write red repro tests

In `cmd/traces_get_test.go` — only THREE tests in this phase (the rest moves to Phase 2):

```go
// 1. Path/method correctness — should already pass; documents the contract.
func TestTracesGet_PathAndMethod(t *testing.T) { ... }

// 2. JSON happy path with fixture — should already pass; locks the envelope shape.
func TestTracesGet_HappyPath_JSON_LocksFixture(t *testing.T) { ... }

// 3. Table mode renders raw JSON, not human-readable — RED (this is the bug).
func TestTracesGet_TableMode_HumanReadable_RED(t *testing.T) {
    // serve fixture; runCmd with "--output", "table"
    // assert stdout does NOT start with `{` (would indicate JSON fallback)
    // assert stdout contains human-readable markers like "TRACE_ID" or tree markers
}
```

Test scaffolding pattern (per repo convention from `cmd/traces_follow_test.go`):
- `httptest.NewServer` + `okJSON(t, w, payload)` envelope helper
- `t.Setenv("GOCLAW_API_URL", srv.URL)` and `t.Setenv("GOCLAW_TOKEN", "test-token")`
- `runCmd(t, "traces", "get", traceID, "--output", "table")` — **explicit `--output` flag** because `go test` pipes stdout (default would resolve to JSON, defeating the table-mode test).
- Each test calls `runCmd` directly; cobra persistent-flag state is reset by passing the flag every time. No `resetTracesGetFlags` helper needed.

### 4. Run tests; expect TestTracesGet_TableMode_HumanReadable_RED to fail

```bash
go test -count=1 ./cmd/... -run TestTracesGet
go vet ./...
go build ./...
```

Tests 1 and 2 should pass (path/method/JSON envelope are already correct). Test 3 should fail with a message like `stdout starts with '{', expected human-readable output`.

### 5. Write `reports/repro-260528-issue17.md`

Sections required:
- **Curl command** (with `$TOKEN` env var, no inline secret).
- **Response sample** — quote 10-30 lines of the scrubbed fixture.
- **Test result summary** — which red, which already green.
- **Classification: CLI-side / server-side / hybrid** with `file:line` evidence for each defect.
- **Server error-code findings** — literal strings observed for 404 and (if probed) 403.
- **If server-side root cause found:** drafted upstream issue body for `digitopvn/goclaw` (also scrubbed; reuse same `jq` recipe).

## Success Criteria

- [ ] `cmd/traces_get_test.go` exists with 3 test cases.
- [ ] `cmd/testdata/trace_detail_get.json` captures real server envelope (or, if gateway unreachable, a stand-in marked `TODO: refresh after live smoke probe`).
- [ ] No bearer token, no real user/tenant id, no PII in the committed fixture (`grep` verified).
- [ ] `go test -count=1 ./cmd/...` runs without compile errors; `TestTracesGet_TableMode_HumanReadable_RED` fails with expected assertion message.
- [ ] `reports/repro-260528-issue17.md` exists with: curl command (env-var token), response sample (scrubbed), classification verdict, file:line evidence, observed server error-code strings.
- [ ] No edits to `cmd/traces.go` or any production code.

## Risk Assessment

- **Risk:** no access to live gateway during planning. **Mitigation:** fixture stand-in marked `TODO: refresh`; Phase 2 must refresh before merge.
- **Risk:** secrets leak into fixture despite `jq` recipe. **Mitigation:** post-scrub grep check listed in step 2; fail-loud during code review subagent in Phase 3.
- **Risk:** issue reporter's actual repro depends on a specific id that no longer exists. **Mitigation:** test against any valid id from `traces list`; if every id 404s for the reporter, that's the root cause and gets captured.

## Security Considerations

- **Smoke probe uses real auth token.** Token MUST come from env var, never inlined into report or fixture. Shell history risk: prefix probe commands with leading space (`HISTCONTROL=ignorespace`) or run from a subshell that won't persist history.
- **Scrub recipe** uses `jq walk` to recursively redact by key-name allowlist; `grep` post-check enforces zero residual secrets.
- **Issue body to upstream** (if filed) reuses the same scrubbed fixture — never the raw one.

## Next Steps

Phase 2 picks up the red test (`TestTracesGet_TableMode_HumanReadable_RED`), adds 7 more tests for the remaining acceptance criteria, then implements until green.
