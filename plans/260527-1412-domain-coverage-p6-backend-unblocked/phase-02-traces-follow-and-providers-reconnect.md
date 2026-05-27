---
phase: 2
title: "Traces Follow + Providers Reconnect (PR #37 surfaces)"
status: pending
priority: P2
effort: "3h"
dependencies: [1]
---

# Phase 2: Traces Follow + Providers Reconnect

## Overview

Implement two CLI surfaces from backend PR #37 (already in beta `v3.12.0-beta.16`+): polling-friendly trace follow and provider reconnect. Strict TDD — failing tests land before any Cobra command code.

## Surfaces

### 2.1 `goclaw traces follow`

```bash
goclaw traces follow --session-key <key> [--since <RFC3339>] [--limit N] [--include-spans] [--status <status>] [--channel <channel>] [-o json|yaml|table]
goclaw traces follow --agent <uuid-or-key> [same flags]
```

- Endpoint: `GET /v1/traces/follow`
- Require exactly one of `--session-key` or `--agent` (validate before HTTP call).
- Optional: `status`, `channel`, `since`, `limit`, `include_spans`. `since` must be RFC3339.
- Server default `limit=50`, max `200`. Don't enforce max client-side; let server respond.
- Response shape:
  ```json
  {"traces":[],"spans_by_trace_id":{},"server_time":"...","next_since":"...","limit":50}
  ```
- JSON/YAML: print full envelope.
- Table: rows like `traces list` — `TRACE_ID`, `AGENT`, `STATUS`, `DURATION_MS`, `INPUT_TOKENS`, `OUTPUT_TOKENS`, `COST`.
- **One request only. No watch loop.**

### 2.2 `goclaw providers reconnect`

```bash
goclaw providers reconnect <provider-id> [-o json|yaml|table]
```

- Endpoint: `POST /v1/providers/{id}/reconnect`
- Admin-only on backend; client sends no body. Do NOT send `{"verify":true}`.
- Do NOT add `--verify` flag; users call `goclaw providers verify <id>` separately if needed.
- Path-escape `<provider-id>` via existing helper.
- Response:
  ```json
  {"status":"reconnected","provider":{},"registry_updated":true,"cache_invalidated":true}
  ```
- Status enum: `reconnected`, `disabled`, `not_registered`.
- Table: `STATUS`, `REGISTRY_UPDATED`, `CACHE_INVALIDATED`, plus provider name/id if non-empty.

## Files

- Modify: `cmd/traces.go` — append `tracesFollowCmd` + register on `tracesCmd`.
- Modify: `cmd/providers.go` or new `cmd/providers_reconnect.go` (preferred — keep file < 200 lines per repo rule) — declare `providersReconnectCmd` + register on `providersCmd`.
- New: `cmd/traces_follow_test.go`
- New: `cmd/providers_reconnect_test.go`

## TDD Sequence

1. Write `cmd/traces_follow_test.go` with the failing tests below; run `go test ./cmd -run TracesFollow` and confirm red.
2. Implement `tracesFollowCmd` minimally until tests pass.
3. Write `cmd/providers_reconnect_test.go`; confirm red.
4. Implement `providersReconnectCmd`; confirm green.
5. `go vet ./... && go build ./...` clean.

## Tests

### `cmd/traces_follow_test.go`

- Session-key target builds path `/v1/traces/follow?session_key=...&...`.
- Agent target builds path `/v1/traces/follow?agent_id=...&...`.
- Missing both target flags returns validation error before HTTP call.
- Setting both target flags returns validation error before HTTP call.
- Non-RFC3339 `--since` returns validation error before HTTP call.
- JSON output preserves `next_since` and `spans_by_trace_id`.
- Table output includes the seven required columns.

### `cmd/providers_reconnect_test.go`

- POST path is `/v1/providers/{escaped-id}/reconnect`.
- Request body is empty (server receives `Content-Length: 0` or empty JSON; assert no `verify` key).
- JSON output preserves `registry_updated` and `cache_invalidated`.
- Table output renders status + boolean columns.
- Provider ID with `/` or `:` is path-escaped (regression test for RT-02).

## Todo List

- [ ] Red tests for traces follow.
- [ ] `tracesFollowCmd` implemented + green.
- [ ] Red tests for providers reconnect.
- [ ] `providersReconnectCmd` implemented + green.
- [ ] `go vet` + `go build` clean.
- [ ] Phase status flipped to Complete.

## Success Criteria

- Both commands callable via `goclaw --help`.
- All new tests pass without skipping or relying on live server.
- No regression in existing `traces list`, `traces get`, `traces export`, or any `providers` subcommand.

## Risks

- Forgetting path-escape on provider ID (RT-02 lesson). Mitigated by explicit escape test.
- Accidentally adding `--verify` based on familiarity with `providers verify`. Mitigated by codex prompt's explicit prohibition + test asserting empty body.

## Next Steps

Phase 3 (Sessions Branch + Follow).
