---
phase: 5
title: "Activity Aggregate + Logs Runtime Aggregate (PR #44)"
status: pending
priority: P2
effort: "3h"
dependencies: [4]
---

# Phase 5: Activity Aggregate + Logs Runtime Aggregate

## Overview

Two aggregation surfaces with similar structure (group_by + filters → buckets). Implement together for shared validation/output helpers. TDD.

## Surfaces

### 5.1 `goclaw activity aggregate`

```bash
goclaw activity aggregate --group-by <action|actor_type|entity_type|actor_id> \
  [--from <RFC3339>] [--to <RFC3339>] [--limit <n>] \
  [--actor-type <v>] [--actor-id <v>] [--action <v>] [--entity-type <v>] [--entity-id <v>] \
  [-o json|yaml|table]
```

- Endpoint: `GET /v1/activity/aggregate`
- Required: `--group-by` ∈ {`action`, `actor_type`, `entity_type`, `actor_id`}.
- Backend restricts `group_by=actor_id` to admin (server enforces; CLI does not pre-check).
- Optional filters: `from`, `to` (both RFC3339 if provided), `limit`, plus actor/action/entity scope filters.
- Response:
  ```json
  {"source":"activity","group_by":"action","total":10,"limit":50,"from":"...","to":"...","buckets":[{"key":"session.branch","count":7,"last_seen":"..."}]}
  ```
- Table: `KEY`, `COUNT`, `LAST_SEEN`.

### 5.2 `goclaw logs aggregate`

```bash
goclaw logs aggregate [--group-by <level|source>] [--level <debug|info|warn|error>] [--source <source>] [--from <RFC3339>] [-o json|yaml|table]
```

- Endpoint: `GET /v1/logs/runtime/aggregate`
- Admin-only. Source = runtime ring buffer, not durable audit log.
- `--group-by` default `level`; valid: {`level`, `source`}.
- Optional filters: `level`, `source`, `from`.
- Response:
  ```json
  {"source":"runtime","retention":"ring_buffer","capacity":100,"sample_size":25,"group_by":"level","buckets":[{"key":"warn","count":3,"last_seen":1760000000000}]}
  ```
  Note: `last_seen` in this response is an **epoch millis number**, not a string. Output helpers must handle.
- Table: `KEY`, `COUNT`, `LAST_SEEN`, plus a header summary line with `SOURCE`, `RETENTION`, `CAPACITY`, `SAMPLE_SIZE` if existing output helpers support it; otherwise drop to JSON-only summary.
- **Do not confuse with `goclaw logs tail` (WS streaming).**

## Files

- New: `cmd/activity_aggregate.go` (no existing activity command group exists).
  - Decide: place under a new top-level `activity` Cobra group, or attach to existing `admin` group? Codex prompt expects `goclaw activity aggregate` as top-level → create new `activityCmd` parent.
- Modify: `cmd/logs.go` — append `logsAggregateCmd`, register on `logsCmd`. If file grows past 200 lines, extract to `cmd/logs_aggregate.go`.
- New: `cmd/activity_aggregate_test.go`
- New: `cmd/logs_aggregate_test.go`

## TDD Sequence

1. Red: activity aggregate test cases.
2. Implement `activityCmd` + `activityAggregateCmd`; green.
3. Red: logs aggregate test cases.
4. Implement `logsAggregateCmd`; green.
5. `go vet ./... && go build ./...` clean.

## Tests

### `cmd/activity_aggregate_test.go`

- Missing `--group-by` rejected before HTTP call.
- Invalid `--group-by=foo` rejected before HTTP call.
- Valid values (`action`, `actor_type`, `entity_type`, `actor_id`) all accepted.
- `--from` / `--to` parse RFC3339; non-RFC3339 rejected before HTTP call.
- All filter flags appear in query string with snake_case keys.
- JSON output preserves `source`, `group_by`, `total`, `buckets`.
- Table renders bucket rows with `KEY`, `COUNT`, `LAST_SEEN`.

### `cmd/logs_aggregate_test.go`

- Default `--group-by` is `level` (or omitted from query — match existing CLI default-omission style; phase 1 scope-lock confirms).
- Invalid `--group-by=foo` rejected before HTTP call.
- `--level`, `--source`, `--from` build correct query.
- JSON output preserves `retention`, `capacity`, `sample_size`.
- Table tolerates numeric `last_seen` (epoch millis) without panicking.

## Todo List

- [ ] Red tests for activity aggregate.
- [ ] `activityCmd` + `activityAggregateCmd` implemented + green.
- [ ] Red tests for logs aggregate.
- [ ] `logsAggregateCmd` implemented + green.
- [ ] `go vet` + `go build` clean.
- [ ] Confirm `logs aggregate` is clearly distinct from `logs tail` in `--help`.
- [ ] Phase status flipped to Complete.

## Success Criteria

- Both aggregate commands callable.
- `last_seen` numeric type handled without runtime panic.
- No confusion with `logs tail` — distinct help text.

## Risks

- `last_seen` type mismatch between activity (RFC3339 string) and logs runtime (epoch millis int). Tests assert both shapes.
- New top-level `activity` command might collide with future scope. Document the namespace decision in PR body.

## Next Steps

Phase 6 (Tests and Docs Sweep).
