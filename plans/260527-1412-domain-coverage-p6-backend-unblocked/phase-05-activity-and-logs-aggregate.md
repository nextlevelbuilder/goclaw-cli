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
  Note: `last_seen` in this response is an **epoch millis number**, not a string.
- **Renderer requirement (Red Team F6):** `unmarshalMap` at `cmd/helpers.go:49-61` decodes JSON numbers as `float64`; the shared `str()` helper uses `fmt.Sprintf("%v", v)`, which renders `1.76e+12` for large numbers. Implement a local `formatLastSeen(v interface{}) string`:
  - If `v` is `string` → assume RFC3339 and return as-is.
  - If `v` is `float64` or `int64` → treat as epoch millis and return `time.UnixMilli(int64(v)).UTC().Format(time.RFC3339)`.
  - If nil/empty → return `"-"`.
  Apply this helper in BOTH `activity aggregate` and `logs aggregate` table renderers so neither produces `1.76e+12`. Place the helper in `cmd/activity_aggregate.go` (or a small shared file `cmd/aggregate_helpers.go`) and import from `cmd/logs_aggregate.go`.
- Table: `KEY`, `COUNT`, `LAST_SEEN` (via `formatLastSeen`), plus a header summary line with `SOURCE`, `RETENTION`, `CAPACITY`, `SAMPLE_SIZE` if existing output helpers support it; otherwise drop to JSON-only summary.
- **Do not confuse with `goclaw logs tail` (WS streaming).**

## Files

- New: `cmd/activity_aggregate.go` — declares `activityAggregateCmd` ONLY (no new top-level parent).
  - **Red Team F1 resolution:** `activityCmd` already exists at `cmd/admin.go:133` (current behavior: lists audit log via `goclaw activity`). Attach the new aggregate as a subcommand in `init()`: `activityCmd.AddCommand(activityAggregateCmd)`. Do NOT declare a new `var activityCmd`. Do NOT touch `cmd/cmd_test.go` top-level list (no new top-level command added).
  - Command UX: `goclaw activity aggregate --group-by ...` (subcommand under existing parent — natural namespacing, no `cmd_test.go` churn).
- Modify: `cmd/logs.go` — append `logsAggregateCmd`, register on `logsCmd`. `cmd/logs.go` is 111 LOC today; adding aggregate may push past 200 → consider `cmd/logs_aggregate.go` upfront.
- New: `cmd/activity_aggregate_test.go`
- New: `cmd/logs_aggregate_test.go`

## TDD Sequence

1. Red: activity aggregate test cases.
2. Implement `activityAggregateCmd` (no `activityCmd` declared — reuse existing parent from `cmd/admin.go:133`); green.
3. Red: logs aggregate test cases (including the `last_seen` RFC3339 rendering assertion).
4. Implement `logsAggregateCmd` + `formatLastSeen` helper; green.
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
- **Render assertion (Red Team F6):** table cell for `LAST_SEEN` matches RFC3339 regex (`^\d{4}-\d{2}-\d{2}T`), NOT `1.76e+12` scientific-notation form. Assert rendered cell content via captured stdout, not absence of panic.

## Todo List

- [ ] Red tests for activity aggregate.
- [ ] `activityAggregateCmd` implemented as subcommand of existing `activityCmd` (do NOT declare new parent); green.
- [ ] Red tests for logs aggregate (incl. RFC3339 cell-content assertion).
- [ ] `formatLastSeen` helper + `logsAggregateCmd` implemented; green.
- [ ] `go vet` + `go build` clean.
- [ ] Confirm `logs aggregate` is clearly distinct from `logs tail` in `--help`.
- [ ] Phase status flipped to Complete.

## Success Criteria

- Both aggregate commands callable.
- `last_seen` numeric type handled without runtime panic.
- No confusion with `logs tail` — distinct help text.

## Risks

- `last_seen` type mismatch between activity (RFC3339 string) and logs runtime (epoch millis int). Resolved by `formatLastSeen` type-switch helper (Red Team F6).
- ~~New top-level `activity` command might collide~~ — **resolved (Red Team F1):** aggregate attaches as subcommand of existing `activityCmd` at `cmd/admin.go:133`. No new top-level.

## Next Steps

Phase 6 (Tests and Docs Sweep).
