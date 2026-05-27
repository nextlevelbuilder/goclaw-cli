---
phase: 7
title: "Ship Readiness"
status: pending
priority: P2
effort: "1h"
dependencies: [6]
---

# Phase 7: Ship Readiness

## Overview

PR hygiene and ship. Single PR from `feat/p6-backend-unblocked-cli` → `dev`. No direct push to `dev` or `main`.

## Requirements

- Branch is `feat/p6-backend-unblocked-cli`, off `dev` tip.
- Working tree clean.
- All phase 1–6 validation gates green.
- No unrelated files staged.

## Implementation Steps

1. `git status --short --branch` clean.
2. Secret scan staged diff.
3. Commit with conventional message:
   - `feat(cli): add P6 backend-unblocked CLI surfaces`
   - Body lists the 7 surfaces, references issue [#16](https://github.com/nextlevelbuilder/goclaw-cli/issues/16), backend PRs #37 / #44, and beta tag `v3.12.0-beta.20`+.
4. Push branch: `git push -u origin feat/p6-backend-unblocked-cli`.
5. Open PR to `dev` via `gh pr create --base dev --head feat/p6-backend-unblocked-cli`.
6. PR body must include:
   - The 7 command surfaces with example invocations.
   - Backend evidence: PR #37 (commit `56e227c4030e85163cd882b29ab472f8ce3e1a27`), PR #44 (commit `43049d3b3fbb5f457477118252d1f21fdc0480de`).
   - Beta tag note: `v3.12.0-beta.20` is the earliest tag containing PR #44; latest is `v3.12.0-beta.35` (2026-05-27).
   - Explicit out-of-scope list (verbatim from issue #16).
   - Validation results.
7. Run review and fix findings before merge.
8. After merge to `dev`, watch CI + Release until beta release publishes.

## Todo List

- [ ] Working tree clean.
- [ ] Conventional commit composed.
- [ ] Branch pushed.
- [ ] PR opened against `dev`.
- [ ] PR body includes backend evidence and out-of-scope list.
- [ ] Review findings addressed.
- [ ] PR merged.
- [ ] Beta release confirmed.
- [ ] Phase status flipped to Complete.

## Success Criteria

- One PR, merged to `dev`.
- Beta release contains the new commands.
- Issue #16 closed with link to merged PR.

## Risks

- `claude-review` workflow may flag advisory issues; address before merge.
- semantic-release may need a `feat:` commit to trigger a beta bump; verify the conventional message header.

## Out of Scope

- Promoting `dev` to `main` (separate ship cycle, like the one that produced PR #18 today).

## Next Steps

Close issue #16 with merge reference. Consider a follow-up P7 plan if backend introduces additional FRs.
