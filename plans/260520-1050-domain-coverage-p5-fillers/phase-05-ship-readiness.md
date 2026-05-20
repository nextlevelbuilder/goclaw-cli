---
phase: 5
title: "Ship Readiness"
status: pending
priority: P2
effort: "1h"
dependencies: [4]
---

# Phase 5: Ship Readiness

## Overview

Prepare the P5 implementation branch for beta shipping. This phase is verification and PR hygiene only.

## Requirements

- Functional: final diff is one cohesive P5 PR.
- Functional: PR description lists sweep results and validation commands.
- Non-functional: no direct push to `dev`; use PR to `dev`.
- Non-functional: no untracked generated artifacts in final status.

## Architecture

Expected flow:

```text
dev synced
  -> feature branch/worktree
  -> implement phases 2-4
  -> local validation
  -> ck:git cp
  -> ck:ship beta
  -> ck:review-pr
```

Release expectation:
- Merge to `dev` triggers CI + Release workflows.
- Dev release should publish prerelease through existing semantic-release setup.

## Implementation Steps

1. Confirm `git status --short --branch` is clean.
2. Review diff for accidental generated files.
3. Run secret scan on staged diff.
4. Commit with conventional message, recommended:
   - `feat(cli): add domain coverage P5 fillers`
5. Push branch and open PR to `dev`.
6. PR body must include:
   - Commands added.
   - Server contracts used.
   - P5 sweep summary.
   - Validation commands and results.
7. Run review and fix findings before merge.
8. After merge, watch CI + Release until complete.
9. Confirm beta release appears.

## Success Criteria

- [ ] Branch pushed.
- [ ] PR to `dev` opened.
- [ ] Review has no critical/important findings.
- [ ] CI pass.
- [ ] Release pass after merge.
- [ ] Beta prerelease published.

## Risk Assessment

| Risk | Mitigation |
|---|---|
| Release workflow fails despite local pass | Watch CI/Release and fix in follow-up branch if needed. |
| PR contains parent-plan-only noise | Keep docs updates scoped to P5 status and command examples. |
| Generated local docs/journals appear ignored | Check `git status --ignored` only if needed; do not force-add ignored journals. |
