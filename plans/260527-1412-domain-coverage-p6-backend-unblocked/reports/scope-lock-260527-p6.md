# Scope Lock Report — P6 Backend-Unblocked CLI

Date: 2026-05-27
Status: locked

## Backend contracts

| Surface | Method + Path | Body / Query | Response | Source |
|---------|---------------|--------------|----------|--------|
| traces follow | GET `/v1/traces/follow` | query: `session_key|agent_id`, `since`, `limit`, `status`, `channel`, `include_spans` | `{traces, spans_by_trace_id, server_time, next_since, limit}` | PR #37 commit `56e227c4` |
| providers reconnect | POST `/v1/providers/{id}/reconnect` | empty body | `{status, provider, registry_updated, cache_invalidated}` | PR #37 commit `56e227c4` |
| sessions branch | POST `/v1/chat/sessions/{key}/branch` | `{new_session_key?, up_to_index, label?, metadata?}` | `{ok, source_key, session_key, copied_messages, total_messages, label}` | PR #44 commit `43049d3b` |
| sessions follow | GET `/v1/chat/sessions/{key}/history/follow` | query: `cursor`, `limit` | `{session_key, cursor, next_cursor, total, messages, reset, updated}` | PR #44 commit `43049d3b` |
| channels writers test | POST `/v1/channels/instances/{id}/writers/test` | `{group_id, user_id}` | `{allowed, reason, instance_id, agent_id, group_id, user_id, writer_count}` | PR #44 commit `43049d3b` |
| activity aggregate | GET `/v1/activity/aggregate` | query: `group_by`, `from`, `to`, `limit`, `actor_type`, `actor_id`, `action`, `entity_type`, `entity_id` | `{source, group_by, total, limit, from, to, buckets:[{key,count,last_seen}]}` (last_seen RFC3339 string) | PR #44 |
| logs aggregate | GET `/v1/logs/runtime/aggregate` | query: `group_by`, `level`, `source`, `from` | `{source, retention, capacity, sample_size, group_by, buckets:[{key,count,last_seen}]}` (last_seen epoch millis number) | PR #44 |

## Beta tag

`v3.12.0-beta.20` verified earliest tag containing `43049d3b` (red-team finding F-RT-beta confirmed during plan creation). Latest beta `v3.12.0-beta.35`.

## Naming collision sweep

```
grep 'Use: "(reconnect|branch|follow|aggregate|test)"' cmd/*.go
```

Result:
- `cmd/heartbeat.go:109` declares `Use: "test"` — under `heartbeatCmd` parent. No conflict (different parent).
- `cmd/hooks_test_runner.go:14` declares `Use: "test"` — under `hooksCmd` parent. No conflict.
- `activityCmd` already exists at `cmd/admin.go:133` — phase 5 attaches `aggregate` as subcommand (not new top-level).
- No other clashes for `follow`, `reconnect`, `branch`, `aggregate` under their respective parents.

## Drift table

| Surface | Drift | Resolution |
|---------|-------|------------|
| logs aggregate `last_seen` | epoch-millis number (not string) | phase 5 implements `formatLastSeen` helper |
| `buildBody` int-zero drop | `--up-to-index 0` / `--cursor 0` silently lost | phase 3 builds body/query maps directly |
| `providers verify` | does not exist (only `verify-embedding`) | phase 2 Long-help removed reference |

## Out-of-scope guarantee

These remain unimplemented (verbatim from issue #16):

- `POST /v1/traces/{id}/replay`
- generic `GET /v1/logs/aggregate`
- WebSocket `chat.history.delta`
- SSE chat history follow
- watch loops for traces/sessions follow

## Result

Zero unresolved contract questions. Proceed to phase 2.
