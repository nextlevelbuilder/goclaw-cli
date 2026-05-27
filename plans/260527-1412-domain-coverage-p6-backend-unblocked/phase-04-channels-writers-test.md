---
phase: 4
title: "Channels Writers Test (PR #44)"
status: pending
priority: P2
effort: "1.5h"
dependencies: [3]
---

# Phase 4: Channels Writers Test

## Overview

Add the writer-permission probe under the existing `channels writers` command group. Small, isolated, TDD.

## Surface

```bash
goclaw channels writers test <instance-id> --group-id <group-scope> --user-id <user-id> [-o json|yaml|table]
```

- Endpoint: `POST /v1/channels/instances/{id}/writers/test`
- Required positional: `<instance-id>`.
- Required flags: `--group-id`, `--user-id`.
- Body **only** has `group_id` and `user_id`. Reject any extra fields client-side.
  ```json
  {"group_id":"group:telegram:-100123","user_id":"386246614"}
  ```
- Response:
  ```json
  {"allowed":true,"reason":"writer","instance_id":"...","agent_id":"...","group_id":"...","user_id":"...","writer_count":3}
  ```
- Known `reason` values: `writer`, `not_writer`, `no_writers_configured`, `invalid_group`.
- Table: `ALLOWED`, `REASON`, `WRITER_COUNT`, `GROUP_ID`, `USER_ID`.

## Files

- Modify: `cmd/channels_writers.go` — append `channelsWritersTestCmd`, register on `channelsWritersCmd`.
- New: `cmd/channels_writers_test_test.go` (file naming kept descriptive; `_test_test.go` is intentional — the command name is `test` and Go test file suffix is `_test.go`).
  - Alternative if Go tooling balks: `cmd/channels_writers_probe_test.go` — but verify Go accepts `_test_test.go` first (it does).

## TDD Sequence

1. Red: write the probe test file with the cases below.
2. Implement `channelsWritersTestCmd`; green.
3. `go vet ./... && go build ./...` clean.

## Tests

- Missing `--group-id` rejected before HTTP call.
- Missing `--user-id` rejected before HTTP call.
- POST path is `/v1/channels/instances/{escaped-id}/writers/test`.
- Request body has exactly `group_id` and `user_id` keys, no extras.
- Instance-id with special chars is path-escaped.
- JSON output preserves `allowed`, `reason`, `writer_count`.
- Table output includes the five required columns.

## Todo List

- [ ] Red probe tests.
- [ ] `channelsWritersTestCmd` implemented + green.
- [ ] `go vet` + `go build` clean.
- [ ] Phase status flipped to Complete.

## Success Criteria

- Subcommand callable as `goclaw channels writers test ...`.
- Body shape verified to contain only the two required keys.

## Risks

- Existing `cmd/channels_writers.go` already has add/remove/list/groups. Adding `test` might push file over 200 lines — split per repo rule if so.

## Next Steps

Phase 5 (Activity + Logs Runtime Aggregate).
