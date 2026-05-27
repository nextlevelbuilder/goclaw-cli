---
phase: 7
title: "Ship Readiness"
status: pending
priority: P2
effort: "15m"
dependencies: [6]
---

# Phase 7: Ship Readiness

## Overview

Single-step phase. Delegate the entire ship pipeline to the `/ck:ship` skill once phases 1–6 are green. This phase contributes the backend-evidence block and the PR body content; everything else (status check, secret scan, commit, push, PR creation, review wait, merge) is owned by `/ck:ship`.

**Red Team F15 resolution:** the prior 9-step checklist duplicated `/ck:ship`. "Watch beta release publish" was open-ended async waiting unsuitable as a phase gate. Both removed. Manual `CHANGELOG.md` edits removed — `semantic-release` is commit-driven (see `.github/workflows/release.yaml`).

## Single Step

Invoke:

```
/ck:ship official
```

Provide this PR-body block when prompted (or paste into the gh-generated body):

```markdown
## Backend evidence

- PR #37 (digitopvn/goclaw) — commit `56e227c4030e85163cd882b29ab472f8ce3e1a27` — surfaces `traces/follow` and `providers/{id}/reconnect`.
- PR #44 (digitopvn/goclaw) — commit `43049d3b3fbb5f457477118252d1f21fdc0480de` — surfaces `chat/sessions/{key}/branch`, `chat/sessions/{key}/history/follow`, `channels/instances/{id}/writers/test`, `activity/aggregate`, `logs/runtime/aggregate`.
- Beta tag: `v3.12.0-beta.20` is the earliest tag containing PR #44 (verified via `gh api repos/digitopvn/goclaw/compare/v3.12.0-beta.20...43049d3b` → `identical`). Latest beta as of 2026-05-27: `v3.12.0-beta.35`.

## Out of scope (verbatim from issue #16)

- POST /v1/traces/{id}/replay
- generic GET /v1/logs/aggregate
- WebSocket chat.history.delta
- SSE chat history follow
- long-running watch loops for traces follow or sessions follow
```

Conventional-commit subject for `/ck:ship` to use:

```
feat(cli): add P6 backend-unblocked CLI surfaces (issue #16)
```

## Todo List

- [ ] `/ck:ship official` invoked with backend-evidence block.
- [ ] PR opened against `dev` with backend-evidence + out-of-scope sections.
- [ ] Phase status flipped to Complete once PR is open (not waiting for merge — `/ck:ship` owns merge cadence).

## Success Criteria

- One PR open against `dev` with the conventional `feat:` subject and the backend-evidence + out-of-scope blocks in the body.

## Out of Scope

- Promoting `dev` to `main` (separate ship cycle, like the one that produced PR #18 on 2026-05-27).
- Manual `CHANGELOG.md` edits (semantic-release commit-driven).
- Watching beta release publish (out of phase 7 scope — beta confirmation happens whenever it happens).

## Next Steps

Close issue #16 with merge reference after `/ck:ship` reports PR merged.
