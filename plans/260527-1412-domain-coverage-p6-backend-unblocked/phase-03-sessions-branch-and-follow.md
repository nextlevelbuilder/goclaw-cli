---
phase: 3
title: "Sessions Branch + Follow (PR #44 chat surfaces)"
status: pending
priority: P2
effort: "3h"
dependencies: [2]
---

# Phase 3: Sessions Branch + Follow

## Overview

Implement the two chat-session surfaces from backend PR #44 (beta `v3.12.0-beta.20`+): branch a session at a message index, and cursor-based history follow. TDD.

## Surfaces

### 3.1 `goclaw sessions branch`

```bash
goclaw sessions branch <session-key> --up-to-index <n> [--new-session-key <key>] [--label <label>] [--metadata k=v]... [-o json|yaml|table]
```

- Endpoint: `POST /v1/chat/sessions/{key}/branch`
- **Backend path note (Red Team F3):** the new `branch`/`follow` commands sit under top-level `sessionsCmd` for UX continuity, but target `/v1/chat/sessions/...` (matches the domain of `cmd/chat_sessions.go`), while sibling `sessions list/preview/delete/reset/label/compact` target `/v1/sessions/...`. Long-help for both new commands MUST state: `Backend route: POST /v1/chat/sessions/{key}/branch (chat domain).` This is the documented exception, not a bug.
- Required: `<session-key>` positional, `--up-to-index` (int, >= 0 — **including zero**).
- Optional: `--new-session-key`, `--label`, `--metadata key=value` (repeatable).
- `--metadata` parses repeated `key=value`; reject malformed entries before HTTP call.
- Path-escape source session key via `url.PathEscape` (may contain `:` and `/`).
- **buildBody-zero workaround (Red Team F2):** the shared `buildBody` helper at `cmd/helpers.go:86-89` drops `int v == 0`. For `up_to_index`, DO NOT use `buildBody`; construct the body map directly so `{"up_to_index": 0, ...}` is preserved on the wire.
- Request body:
  ```json
  {"new_session_key":"...","up_to_index":12,"label":"...","metadata":{"source":"cli"}}
  ```
- Response:
  ```json
  {"ok":true,"source_key":"...","session_key":"...","copied_messages":12,"total_messages":24,"label":"..."}
  ```
- Table: `SOURCE`, `NEW_KEY`, `COPIED`, `TOTAL`, `LABEL`.

### 3.2 `goclaw sessions follow`

```bash
goclaw sessions follow <session-key> [--cursor <n>] [--limit <n>] [-o json|yaml|table]
```

- Endpoint: `GET /v1/chat/sessions/{key}/history/follow`
- Query: `cursor` (default 0, >= 0 — **including zero**), `limit` (default 50, > 0, server max 200).
- **buildBody-zero workaround (Red Team F2):** build the query string directly with `url.Values`; do NOT use `buildBody` (which would drop `cursor=0` per the int-zero skip rule). `cursor=0` MUST appear in the query string when `--cursor 0` is passed.
- **One polling request only. No SSE/WS watch.** Direct `httpClient.Get` call — must NOT use `client.FollowStream` (`internal/client/follow.go`), which reconnects on EOF.
- Response:
  ```json
  {"session_key":"...","cursor":12,"next_cursor":18,"total":18,"messages":[],"reset":false,"updated":"..."}
  ```
- Table: print summary row (cursor, next_cursor, total, reset, updated) plus compact message rows.

## Files

- Modify: `cmd/sessions.go` — append `sessionsBranchCmd` + `sessionsFollowCmd` and register both.
  - If `cmd/sessions.go` grows past 200 lines, split into `cmd/sessions_branch.go` and `cmd/sessions_follow.go` per repo modularization rule.
- New: `cmd/sessions_branch_test.go`
- New: `cmd/sessions_follow_test.go`

## TDD Sequence

1. Red: `cmd/sessions_branch_test.go`.
2. Implement `sessionsBranchCmd`; green.
3. Red: `cmd/sessions_follow_test.go`.
4. Implement `sessionsFollowCmd`; green.
5. `go vet ./... && go build ./...` clean.

## Tests

### `cmd/sessions_branch_test.go`

- `--up-to-index` missing returns validation error before HTTP call.
- Negative `--up-to-index` returns validation error before HTTP call.
- **Zero-boundary test (Red Team F2):** `--up-to-index 0` produces request body containing `"up_to_index":0` literally (not omitted, not missing). Use `json.Unmarshal` on the captured request body and assert the key is present with value `0`.
- Request body shape exact: `up_to_index` as int, `metadata` as object.
- `--metadata foo=bar --metadata baz=qux` produces `{"foo":"bar","baz":"qux"}`.
- Malformed `--metadata foobar` (no `=`) rejected before HTTP call.
- Session key with `:` and `/` is path-escaped (regression for RT-02).
- HTTP 409 conflict surfaces via existing error handler (use `client.APIError` path).
- JSON output preserves `copied_messages` and `total_messages`.

### `cmd/sessions_follow_test.go`

- Default cursor=0, limit=50 appear in query string.
- **Zero-boundary test (Red Team F2):** `--cursor 0` results in `cursor=0` appearing in raw query string (not omitted by buildBody int-zero skip).
- Custom cursor and limit appear in query string.
- Negative `--cursor` rejected before HTTP call.
- Non-positive `--limit` rejected before HTTP call.
- Session key path-escaped.
- JSON output preserves `reset`, `next_cursor`, `messages`.
- **Atomic-counter test (Red Team F7):** wrap test server handler with `atomic.AddInt64(&calls, 1)`; assert `calls == 1` after `RunE`. Assert command does NOT import / call `client.FollowStream`.

## Todo List

- [ ] Red tests for sessions branch.
- [ ] `sessionsBranchCmd` implemented + green.
- [ ] Red tests for sessions follow.
- [ ] `sessionsFollowCmd` implemented + green.
- [ ] `go vet` + `go build` clean.
- [ ] If `cmd/sessions.go` exceeds 200 lines, split (per repo rule).
- [ ] Phase status flipped to Complete.

## Success Criteria

- Both commands callable.
- Path-escape regression test passes.
- No watch loop introduced.

## Risks

- Drifting metadata shape (object vs array). Mitigated by exact-body test against backend handler signature confirmed in phase 1.
- Session-key path escape forgotten. Mitigated by explicit regression test.

## Next Steps

Phase 4 (Channels Writers Test).
