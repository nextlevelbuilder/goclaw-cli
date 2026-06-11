---
title: Trace Search Filter CLI
description: >-
  Add goclaw-cli support for the GET /v1/traces search and advanced filter query
  parameters merged in digitopvn/goclaw#155.
status: in-progress
priority: P1
branch: codex/trace-search-filter-traces
tags:
  - traces
  - cli
  - api
  - tdd
  - beta
blockedBy: []
blocks: []
created: '2026-06-11T17:23:28.191Z'
createdBy: 'ck:plan'
source: skill
---

# Trace Search Filter CLI

## Overview

`digitopvn/goclaw#155` is merged into server `dev` and extends `GET /v1/traces` with search and advanced filters. `goclaw-cli` already supports basic trace list filters, server-shaped list envelopes, detail, export, follow, and timeline. This plan adds only the missing list filter flags and documentation, using tests first.

## Expected Output

Users can run `goclaw traces list` with these additional flags and the CLI forwards them to `GET /v1/traces`:

```bash
goclaw traces list --q "error refund" --agent-query helper --channel-query telegram \
  --from 2026-06-10T00:00:00Z --to 2026-06-12T00:00:00Z \
  --min-input-tokens 10 --max-input-tokens 1000 \
  --min-output-tokens 5 --max-output-tokens 500 \
  --min-tool-calls 1 --max-tool-calls 3 \
  --tool-name web_search --has-tool-calls true
```

## Acceptance Criteria

- `traces list` forwards exact server query names:
  `q`, `agent`, `channel_query`, `from`, `to`, `min_input_tokens`, `max_input_tokens`, `min_output_tokens`, `max_output_tokens`, `min_tool_calls`, `max_tool_calls`, `tool_name`, `has_tool_calls`.
- Existing filters remain unchanged:
  `agent_id`, `user_id`, `session_key`, `status`, `channel`, `limit`, `offset`.
- JSON/YAML output still prints the server list envelope; table output still renders rows from `traces`.
- CLI does not client-escape wildcard characters inside search strings; server handles wildcard escaping.
- `traces follow`, `traces get`, `traces export`, and `traces timeline` behavior stays unchanged.
- No `traces replay` command is added.
- README, CHANGELOG, and project docs mention the new list filters without implying watch loops or server-side replay.

## Scope Boundary

In scope:
- Add `traces list` flags for the merged server PR #155 list filters.
- Add focused `httptest` contract coverage for query forwarding.
- Update docs for the new CLI flags.

Out of scope:
- Server changes in `digitopvn/goclaw`.
- New trace UI work.
- `POST /v1/traces/{id}/replay`.
- New streaming/watch behavior.
- Changing output payload shapes or trace table columns.

## Touchpoints

- Modify: `cmd/traces.go`
- Modify: `cmd/traces_contract_test.go`
- Modify: `cmd/traces_list_test.go`
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Modify: `docs/project-roadmap.md`
- Modify: `docs/codebase-summary.md`
- Read-only contract sources:
  `digitopvn/goclaw#155`, `internal/http/traces.go`, `internal/store/tracing_store.go`

## Phases

| Phase | Name | Status |
|-------|------|--------|
| 1 | [Contract Lock](./phase-01-contract-lock.md) | Completed |
| 2 | [TDD Implementation](./phase-02-tdd-implementation.md) | Completed |
| 3 | [Verification and Ship](./phase-03-verification-and-ship.md) | In Progress |

## Dependencies

| Relationship | Plan | Reason |
|--------------|------|--------|
| Related | `../260611-2303-complete-traces-cli-contract/plan.md` | Existing completed traces contract work provides the list envelope and replay exclusion baseline. |
| Supersedes detail only | `../260528-1357-fix-trace-details-by-id/plan.md` | Older trace detail plan remains out of scope for this list-filter slice. |

## Validation Gates

- Focused red/green: `go test -count=1 ./cmd -run 'TestTracesList|TestTracesReplayCommandAbsent'`
- Command compile: `go test -count=1 ./cmd`
- Full verification: `go test -count=1 ./...`
- Static check: `go vet ./...`
- Compile: `go build ./...`

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| Flag naming could confuse server `agent_id` with server `agent` label query | Keep CLI `--agent` for existing ID behavior and use `--agent-query` for the server `agent` query; document mapping in tests. |
| Boolean flag cannot express server false filter if modeled as Cobra bool with default false | Use a string flag `--has-tool-calls` and forward only when explicitly changed; accept `true` or `false` as server parses bool. |
| Numeric range flags with default zero could accidentally filter | Forward int range flags only when the flag changed. |
| Search wildcard characters might be escaped twice | Forward raw values; server PR #155 owns wildcard escaping. |
| `cmd/traces.go` is already over 200 lines | Keep edits local and minimal; do not refactor unrelated trace commands in this slice. |

## Unresolved Questions

None.
