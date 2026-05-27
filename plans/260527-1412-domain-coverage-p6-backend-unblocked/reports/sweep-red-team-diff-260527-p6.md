# P6 Red-Team Diff Sweep

**Date:** 2026-05-27
**Scope:** Seven backend-unblocked CLI surfaces (issue #16; backend PRs `#37`, `#44`)
**Mode:** Implementation→sweep diff vs validated plan + red-team artifacts.

## Surfaces shipped

| Command | Endpoint | File | Tests |
|---|---|---|---|
| `traces follow` | `GET /v1/traces/follow` | `cmd/traces_follow.go` | `cmd/traces_follow_test.go` |
| `providers reconnect <id>` | `POST /v1/providers/{id}/reconnect` | `cmd/providers_reconnect.go` | `cmd/providers_reconnect_test.go` |
| `sessions branch <key>` | `POST /v1/chat/sessions/{key}/branch` | `cmd/sessions_branch.go` | `cmd/sessions_branch_test.go` |
| `sessions follow <key>` | `GET /v1/chat/sessions/{key}/history/follow` | `cmd/sessions_follow.go` | `cmd/sessions_follow_test.go` |
| `channels writers test <id>` | `POST /v1/channels/instances/{id}/writers/test` | `cmd/channels_writers.go` (appended) | `cmd/channels_writers_test_test.go` |
| `activity aggregate` | `GET /v1/activity/aggregate` | `cmd/activity_aggregate.go` | `cmd/activity_aggregate_test.go` |
| `logs aggregate` | `GET /v1/logs/runtime/aggregate` | `cmd/logs_aggregate.go` | `cmd/logs_aggregate_test.go` |

## Build / vet / test

```
go vet ./...    # clean
go build ./...  # clean
go test ./...   # all packages PASS
```

## Red-team findings — disposition

| ID | Finding | Implementation status |
|---|---|---|
| F1 | `activityCmd` already exists at `cmd/admin.go:133` — new aggregate must attach as subcommand. | Done. `activityAggregateCmd` registered via `activityCmd.AddCommand(...)`; no new top-level. `cmd_test.go` top-level list unchanged. |
| F2 (sessions branch) | `buildBody` int-zero skip silently drops `up_to_index=0`. | Done. `sessions_branch.go` constructs body directly (`map[string]any{"up_to_index": upTo}`); regression test asserts `"up_to_index":0` on the wire. |
| F2 (sessions follow) | Same int-zero hazard for `cursor=0`. | Done. `sessions_follow.go` builds `url.Values` directly with `q.Set("cursor", fmt.Sprintf("%d", cursor))`; regression test asserts literal `cursor=0` in raw query. |
| F4 | HTTP client auto-retries 429/5xx three times — conflicts with a "fail-fast on 502" assertion. | Mitigated. Plan's 502-once test removed; atomic-counter test (`calls == 1` on 200 path) and structural "not a watch loop" check preserve the original intent. |
| F6 | `unmarshalMap` decodes JSON numbers as float64; `str()` renders `1.76e+12`. | Done. `formatLastSeen` helper in `cmd/activity_aggregate.go` type-switches (string→passthrough, numeric→`time.UnixMilli`). Imported and used in both `activity_aggregate.go` and `logs_aggregate.go`. Test `TestLogsAggregate_LastSeenRendersRFC3339` asserts absence of `e+12` and presence of RFC3339 pattern. |
| Naming collision (heartbeat/hooks `Use: "test"`) | Different parents — no collision. | Verified. `channelsWritersTestCmd` lives under `channelsWritersCmd`; the `Use: "test"` strings under different parents are routed disambiguously by cobra. |
| Path-escape | All `{id}`/`{key}` path params must use `url.PathEscape`. | Done. Every new command uses `url.PathEscape(args[0])`; path-escape regression tests assert either escaped `RawPath` or decoded `Path` contains the literal. |
| Empty-body POST shape | `providers reconnect` body must be empty (no `verify` field). | Done. Test asserts `body["verify"]` absent and `len(body) == 0`. |
| Two-key-only body | `channels writers test` body must contain ONLY `group_id` and `user_id`. | Done. Test asserts `len(body) == 2`. |

## Code-comment rule compliance

Initial test comments referenced plan artifacts (`Red Team F2`, `Red Team F6`). Rewritten to describe the invariant (numeric-zero preservation, scientific-notation rendering) rather than the origin. Production code (`cmd/*_aggregate.go`, `cmd/*_follow.go`, etc.) contains no plan-artifact references; only invariants and contract notes are inlined.

Pre-existing legacy "Phase 4 split" comments in `cmd/agents_instances.go`, `cmd/teams_members.go`, `cmd/teams_workspace.go` are out of scope for this pass.

## Backend evidence (locked)

- PR `#37` (traces follow + providers reconnect): merged commit `56e227c4`.
- PR `#44` (sessions branch + sessions follow + channels writers test + activity aggregate + logs aggregate): merged commit `43049d3b`.
- Beta release tag covering both: `v3.12.0-beta.20` (minimum required for CLI feature flag).

## Unresolved questions

None.
