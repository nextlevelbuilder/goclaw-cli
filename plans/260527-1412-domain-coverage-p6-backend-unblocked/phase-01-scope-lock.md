---
phase: 1
title: "Scope Lock"
status: pending
priority: P2
effort: "1h"
dependencies: []
---

# Phase 1: Scope Lock

## Overview

Lock the 7-surface scope, re-verify backend contracts against `digitopvn/goclaw` `dev` tip, and produce the contract evidence file that subsequent phases reference. No CLI code changes.

## Requirements

- Re-fetch and read each backend handler to confirm path, method, query params, body shape, status enum, and error semantics.
- Confirm beta tag `v3.12.0-beta.20` is the minimum that contains PR #44 commit `43049d3b` (already verified during plan creation; phase re-asserts).
- Confirm no command-name collisions in current `cmd/` tree.
- Produce `reports/scope-lock-260527-p6.md` summarizing contracts and any drift discovered.

## Implementation Steps

1. From the `digitopvn/goclaw` checkout (or via `gh api`), read:
   - `internal/http/traces.go` — locate the `traces/follow` handler.
   - `internal/http/providers.go` — locate the `providers/{id}/reconnect` handler.
   - `internal/http/sessions.go` — locate the branch and history/follow handlers.
   - `internal/http/channel_instances.go` — locate the writers/test handler.
   - `internal/http/activity.go` — locate the aggregate handler.
   - `internal/http/logs.go` — locate the runtime/aggregate handler.
   - `internal/http/openapi_spec.json` for any documented schema differences.
2. For each handler record: route, method, required vs optional inputs, response shape, status enum, error codes, admin-only flag.
3. Compare against `plans/reports/codex-prompt-260522-p6-pr44-backend-unblocked-cli.md` (lives on `feat/claude-skill-v0.1` worktree) and the plan overview. Flag any drift.
4. Re-verify beta tag: `gh api repos/digitopvn/goclaw/compare/v3.12.0-beta.20...43049d3b --jq '.status'` must return `identical`.
5. Re-verify no naming collisions: `grep -in 'Use:.*"follow\|reconnect\|branch\|aggregate\|test"' cmd/*.go`.
6. Write `reports/scope-lock-260527-p6.md` with the table of confirmed contracts.

## Todo List

- [ ] Backend contract re-read for all 7 endpoints.
- [ ] Drift table populated (or zero-drift confirmed).
- [ ] Beta tag re-confirmed.
- [ ] Naming collision sweep run with output captured.
- [ ] `reports/scope-lock-260527-p6.md` written.
- [ ] Phase status flipped to Complete.

## Success Criteria

- Zero unresolved contract questions before phase 2 starts.
- Scope file under `reports/` exists and is referenced from each implementation phase's "Evidence" section.

## Out of Scope

- Any CLI code change.
- Any test scaffolding (phase 2+).

## Next Steps

Proceed to Phase 2 (Traces Follow + Providers Reconnect).
