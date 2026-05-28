# Repro & root cause — issue #17 (cannot read trace details by id)

Date: 2026-05-28
Phase: 1
Worktree: `worktree-fix-trace-details-by-id`

## Smoke probe

**Intended target:** `goclaw.zuey.me` (production).
**Result:** auth blocked. No `~/.goclaw/config.yaml`, no token in env. The plan-documented escape applies — use a stub fixture derived from the existing `traces follow` payload shape ([cmd/traces_follow_test.go:129-130](cmd/traces_follow_test.go:129)), marked `_TODO_refresh` for the Phase 3 reviewer gate to enforce a real-gateway refresh before merge.

**Curl command for future smoke probe (env-var token, no inline secret):**

```bash
export TOKEN="$(goclaw config get token 2>/dev/null || cat ~/.goclaw/token)"
export SERVER="https://goclaw.zuey.me"
TRACE_ID="$(goclaw traces list --limit 1 -o json | jq -r '.payload[0].trace_id')"
curl -sS -H "Authorization: Bearer $TOKEN" "$SERVER/v1/traces/$TRACE_ID" > /tmp/trace_raw.json
goclaw traces get "$TRACE_ID" > /tmp/trace_tty.txt
goclaw traces get "$TRACE_ID" -o json > /tmp/trace_json.txt
goclaw traces get bogus-id-12345 -o json > /tmp/trace_404.txt 2>&1
```

After refresh, run the `jq walk` scrub recipe from phase-01 step 2, then `grep -i 'eyJ\|Bearer\|sk-\|token=' cmd/testdata/trace_detail_get.json` must return 0 lines.

## Fixture (scrubbed stub — sample)

```json
{
  "_TODO_refresh": "stub fixture ...; refresh against goclaw.zuey.me before merge per phase-03 reviewer gate",
  "trace_id": "trace_FIXTURE_001",
  "agent_id": "agent_FIXTURE_001",
  "session_key": "session_FIXTURE_001",
  "user_id": "user_REDACTED",
  "tenant_id": "tenant_REDACTED",
  "status": "success",
  "duration_ms": 2000,
  "spans": [
    {"span_id": "span_001", "parent_span_id": null, "name": "agent.run", "kind": "agent", "status": "success"},
    {"span_id": "span_002", "parent_span_id": "span_001", "name": "llm.call", "kind": "llm", "status": "success"},
    {"span_id": "span_003", "parent_span_id": "span_001", "name": "tool.call", "kind": "tool", "status": "success"}
  ],
  "events": [
    {"event_id": "ev_001", "type": "llm.prompt"},
    {"event_id": "ev_002", "type": "llm.completion"},
    {"event_id": "ev_003", "type": "tool.invoke"}
  ]
}
```

Secret-scan post-check: `grep -i 'eyJ\|Bearer\|sk-\|token=' cmd/testdata/trace_detail_get.json` → 0 lines. ✅

## Test result summary

```
go test -count=1 ./cmd/... -run TestTracesGet -v
```

| Test | Result | Meaning |
|------|--------|---------|
| `TestTracesGet_PathAndMethod` | PASS | Wire contract (`GET /v1/traces/{id}`) already correct. |
| `TestTracesGet_HappyPath_JSON_LocksFixture` | PASS | JSON-mode envelope round-trips correctly. |
| `TestTracesGet_TableMode_HumanReadable_RED` | **FAIL** | Table mode emits raw JSON beginning with `{` — the reported "unusable output" bug, locked. |

Failure assertion: `table mode rendered raw JSON (starts with '{')`. Exactly as predicted by the scout findings.

## Classification: **CLI-side**

All three verified defects are in the CLI; no server-side root cause is required for AC#1–#3.

| # | Defect | Evidence | Fix lane |
|---|--------|----------|----------|
| 1 | Table mode falls back to JSON dump for `map[string]any` payload | [cmd/traces.go:72](cmd/traces.go:72) calls `printer.Print(unmarshalMap(data))`; `output.Printer.Print` only formats `*TableData` in table mode, otherwise JSON-fallbacks ([internal/output/output.go:30-37](internal/output/output.go:30)) | Phase 2 — inline render (header + span tree via `output.PrintTree` + flat events list) |
| 2 | `unmarshalMap` silently swallows `json.Unmarshal` errors | [cmd/helpers.go:48-53](cmd/helpers.go:48) — literal `_ = json.Unmarshal(data, &m)` | Phase 2 — inline `json.Unmarshal` with `return fmt.Errorf("decode trace payload: %w", err)` |
| 3 | No id validation; raw `args[0]` concatenated into URL with no `url.PathEscape` | [cmd/traces.go:68](cmd/traces.go:68) — `"/v1/traces/" + args[0]` | Phase 2 — strict allowlist regex `^[A-Za-z0-9._-]+$` + reject `.` / `..` / empty / whitespace, then `url.PathEscape` |
| 4 | No client-side categorization between 404 / 403 / malformed-id / 5xx | `cmd/traces.go` `tracesGetCmd` has no error mapping | Phase 2 tests + existing `apiErrorCodeForStatus` ([cmd/helpers.go:152-172](cmd/helpers.go:152)) already maps every required HTTP status — no `MapServerCode` extension. |

## Server error-code findings

Not observed live (auth-blocked). Plan does not speculate server-code strings — relies on `apiErrorCodeForStatus` HTTP-status mapping which covers every AC#3 category:

| HTTP | Canonical code | Exit code |
|------|----------------|-----------|
| 400 / 422 | `INVALID_REQUEST` | 4 |
| 401 | `UNAUTHORIZED` | 2 |
| 403 | `TENANT_ACCESS_REVOKED` | 2 |
| 404 | `NOT_FOUND` | 3 |
| 429 | `RESOURCE_EXHAUSTED` | 6 |
| 5xx | `INTERNAL` | 5 |

This locks AC#3 behavior independent of upstream server-code strings.

## Upstream issue body

**Not filed.** Root cause is CLI-side; no upstream issue required for `digitopvn/goclaw`. Phase 3 step 5 short-circuits.

## Acceptance criteria mapping (preview)

| AC | Phase | Status after Phase 1 |
|----|-------|----------------------|
| 1. Read trace details by id | Phase 2 (render path + decode-error surfacing) | Pending |
| 2. JSON & human-readable output | Phase 2 (red test 3 + new green tests) | Red test locked |
| 3. Distinct errors: not-found / perm / malformed / server | Phase 2 (HTTP-status mapping; existing `apiErrorCodeForStatus`) | Pending tests |
| 4. Link upstream issue if API root cause | n/a — classified CLI-side | Resolved |
| 5. Regression test | Phase 1 (3 tests) + Phase 2 (7 more) | 3 of 10 landed |

## Unresolved questions

- **Real-gateway fixture refresh.** Stub fixture is shaped from `traces follow` payload conventions, not the real `GET /v1/traces/{id}` wire envelope. If the real shape differs materially (e.g. spans nested under `tree`, events under `event_log`), Phase 2 render assertions may need light shape adjustments. Phase 3 code-reviewer gate enforces refresh + re-run before merge.
