---
title: "Domain Coverage P6 — Backend-Unblocked CLI Surfaces"
description: "Implement the 7 CLI commands now unblocked by digitopvn/goclaw PR #37 and PR #44. TDD-driven, 4 grouped implementation phases."
status: pending
priority: P2
branch: "feat/p6-backend-unblocked-cli"
base: "dev"
tags: [domain-coverage, p6, cli, tdd, backend-unblocked]
blockedBy: []
blocks: []
parentBacklog: "../260503-1907-domain-coverage-p3-plus/plan.md"
relatedIssue: "https://github.com/nextlevelbuilder/goclaw-cli/issues/16"
created: "2026-05-27T07:12:00Z"
createdBy: "ck:plan --tdd"
source: skill
---

# Domain Coverage P6 — Backend-Unblocked CLI Surfaces

## Overview

Implement the 7 CLI command surfaces unblocked by two backend PRs in `digitopvn/goclaw`:

- **PR #37** (merged, beta `v3.12.0-beta.16`+) → `GET /v1/traces/follow`, `POST /v1/providers/{id}/reconnect`
- **PR #44** (merged, beta `v3.12.0-beta.20`+ verified) → `POST /v1/chat/sessions/{key}/branch`, `GET /v1/chat/sessions/{key}/history/follow`, `POST /v1/channels/instances/{id}/writers/test`, `GET /v1/activity/aggregate`, `GET /v1/logs/runtime/aggregate`

Mirror issue [#16](https://github.com/nextlevelbuilder/goclaw-cli/issues/16) scope exactly. 4 implementation phases grouped by backend PR + functional cluster. TDD: failing tests before each implementation slice.

## Resolved Pre-Plan Question

Issue #16 asked: *"Which beta tag ultimately contains backend commit `43049d3b`?"*

**Verified via `gh api compare`:** `v3.12.0-beta.20` is the earliest tag identical to `43049d3b`. Live smoke MUST target ≥ `v3.12.0-beta.20`. Latest beta as of 2026-05-27 is `v3.12.0-beta.35`.

## Decisions

- Scope: exactly the 7 surfaces in issue #16, no extras.
- Grouping: 4 implementation phases, not 7 (shared client helpers + test fixtures).
- TDD: failing command tests land before Cobra command code in every phase.
- No watch loops. `traces follow` and `sessions follow` are one-shot polling requests; SSE/WS deferred.
- No commands or stubs for explicitly out-of-scope APIs (see below).
- Reuse `internal/client.HTTPClient`, `internal/output`, existing `printer.Print(unmarshalMap/List(...))` patterns.
- Path-escape all path params via the same pattern already used in `cmd/api_keys_rotate.go`, `cmd/storage.go` (after P5 RT-02 fix).
- Branch off `dev`, not `main`. One PR back to `dev`.

## Phases

| Phase | Name | Surfaces | Status |
|-------|------|----------|--------|
| 1 | [Scope Lock](./phase-01-scope-lock.md) | n/a | Pending |
| 2 | [Traces Follow + Providers Reconnect](./phase-02-traces-follow-and-providers-reconnect.md) | 2 (PR #37) | Pending |
| 3 | [Sessions Branch + Follow](./phase-03-sessions-branch-and-follow.md) | 2 (PR #44 chat) | Pending |
| 4 | [Channels Writers Test](./phase-04-channels-writers-test.md) | 1 (PR #44 channels) | Pending |
| 5 | [Activity + Logs Runtime Aggregate](./phase-05-activity-and-logs-aggregate.md) | 2 (PR #44 aggregation) | Pending |
| 6 | [Tests and Docs Sweep](./phase-06-tests-and-docs.md) | n/a | Pending |
| 7 | [Ship Readiness](./phase-07-ship-readiness.md) | n/a | Pending |

## Command Surface Inventory

```bash
goclaw traces follow --session-key <key> [--since <RFC3339>] [--limit N] [--include-spans] [--status <status>] [--channel <channel>]
goclaw traces follow --agent <uuid-or-key> [same flags]

goclaw providers reconnect <provider-id>

goclaw sessions branch <session-key> --up-to-index <n> [--new-session-key <key>] [--label <label>] [--metadata k=v]...
goclaw sessions follow <session-key> [--cursor <n>] [--limit <n>]

goclaw channels writers test <instance-id> --group-id <group-scope> --user-id <user-id>

goclaw activity aggregate --group-by <action|actor_type|entity_type|actor_id> [filters]

goclaw logs aggregate [--group-by <level|source>] [--level <level>] [--source <source>] [--from <RFC3339>]
```

All commands support `--output {table,json,yaml}` per existing TTY auto-detection.

## Explicitly Out of Scope (verbatim from issue #16)

Do not add commands, stubs, hidden flags, or docs for APIs that still do not exist:

- `POST /v1/traces/{id}/replay`
- generic `GET /v1/logs/aggregate`
- WebSocket `chat.history.delta`
- SSE chat history follow
- long-running watch loops for `traces follow` or `sessions follow`

## Existing CLI State (verified 2026-05-27 on `dev` tip `143624d`)

| File | Existing surfaces | New work |
|------|------------------|----------|
| `cmd/traces.go` | `list`, `get`, `export` | add `follow` subcommand |
| `cmd/providers.go` + `providers_crud.go` + `providers_verify.go` | CRUD + verify | add `reconnect` subcommand |
| `cmd/sessions.go` + `chat_sessions.go` | `list`, `preview`, `delete`, `reset`, `label`, `compact` | add `branch`, `follow` |
| `cmd/channels_writers.go` | `list`, `groups`, `add`, `remove` | add `test` |
| _(no `cmd/admin_activity.go`)_ | — | new file `cmd/activity_aggregate.go` |
| `cmd/logs.go` | `tail` (WS streaming) | add `aggregate` (HTTP, distinct subcommand) |

No command-name collisions.

## Dependencies

- Parent backlog: `../260503-1907-domain-coverage-p3-plus/plan.md` (P6 listed as server-blocked; this plan unblocks)
- Codex prompt evidence: `plans/reports/codex-prompt-260522-p6-pr44-backend-unblocked-cli.md` (lives on `feat/claude-skill-v0.1` worktree)
- Backend contracts:
  - `digitopvn/goclaw` `internal/http/traces.go` (PR #37)
  - `digitopvn/goclaw` `internal/http/providers.go` (PR #37)
  - `digitopvn/goclaw` `internal/http/sessions.go` (PR #44)
  - `digitopvn/goclaw` `internal/http/channel_instances.go` (PR #44)
  - `digitopvn/goclaw` `internal/http/activity.go` (PR #44)
  - `digitopvn/goclaw` `internal/http/logs.go` (PR #44)
- OpenAPI: `digitopvn/goclaw` `internal/http/openapi_spec.json`

## Validation Gates (per phase + final)

- `/usr/local/go/bin/go test ./...` (TDD: red → green per phase)
- `/usr/local/go/bin/go vet ./...`
- `/usr/local/go/bin/go build ./...`
- Live smoke against `v3.12.0-beta.20`+ backend (optional, after merge)
- Red-team diff before PR (path/body/output regressions, accidental scope creep into out-of-scope list)

## Success Criteria

- [ ] All 7 surfaces implemented per contracts in respective phases.
- [ ] Focused tests cover path, query, body, validation-before-HTTP, output format preservation.
- [ ] No CLI command exists for `traces/{id}/replay` or generic `/v1/logs/aggregate`.
- [ ] No watch loops added.
- [ ] `go test ./... && go vet ./... && go build ./...` all green.
- [ ] README/help text reflects new commands (deferred to phase 6 if needed).
- [ ] PR body lists backend PR/tag evidence and notes ≥ `v3.12.0-beta.20` requirement.

## Risk Assessment

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Backend response shape drift between dev and beta tag | Low | Phase 1 re-verifies contracts against `digitopvn/goclaw` `dev` HEAD before tests written |
| Path-escape regression (RT-02 from P5) | Medium | TDD test for escaped session keys with `:` and `/` in phase 3 |
| Scope creep into watch loops or replay | Medium | Explicit out-of-scope checklist in phase 6 red-team diff |
| Untracked files in `feat/claude-skill-v0.1` accidentally staged | Low | This plan uses a clean worktree (`.claude/worktrees/elated-galileo-2c7cfd`), separate branch |
| Naming collision with existing subcommands | Low | Verified inventory above; no collisions found |

## Handoff

Recommended next: invoke `/ck:cook` to execute phase 1 first (scope lock + contract re-verification), then proceed phase-by-phase with TDD.
