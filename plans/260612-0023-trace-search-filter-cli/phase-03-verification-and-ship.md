---
phase: 3
title: Verification and Ship
status: in-progress
priority: P1
effort: 2h
dependencies:
  - 2
---

# Phase 3: Verification and Ship

## Overview

Update user-facing docs, run verification, review the change, ship a beta PR, review the PR with fix loop, merge, and watch CI until success.

## Requirements

- Functional: docs show the new filters and preserve existing trace examples.
- Non-functional: full Go validation passes before PR.
- Non-functional: beta PR targets `dev`, not `main`.

## Architecture

No architecture changes. Documentation reflects the Cobra command surface and the upstream server query contract.

## Related Code Files

- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Modify: `docs/project-roadmap.md`
- Modify: `docs/codebase-summary.md`
- Read: `.github/workflows/*`

## Implementation Steps

1. Update README trace list example with the new advanced filters.
2. Add a concise CHANGELOG entry for trace search/filter CLI support.
3. Update project docs status for the 2026-06-12 trace filter slice.
4. Run verification:
   - `go test -count=1 ./cmd -run 'TestTracesList|TestTracesReplayCommandAbsent'`
   - `go test -count=1 ./cmd`
   - `go test -count=1 ./...`
   - `go vet ./...`
   - `go build ./...`
5. Run code review and address actionable findings.
6. Ship beta PR to `dev`.
7. Run PR review with fix loop.
8. Merge PR and watch/fix checks until success.

## Todo List

- [x] README updated.
- [x] CHANGELOG updated.
- [x] Project docs updated.
- [x] Focused and full verification pass.
- [x] Code review passes.
- [ ] Beta PR merged and CI success confirmed.

## Success Criteria

- [ ] Pull request to `dev` is merged.
- [ ] GitHub checks for the merged PR are successful or explicitly confirmed not present.
- [ ] Final report lists commits, PR URL, verification commands, and unresolved questions.

## Risk Assessment

CI may run a broader release matrix than local validation. If PR checks fail, inspect the exact job logs, patch only relevant failures, push, and re-watch.

## Security Considerations

Do not stage secrets or local config. Verify `git diff --cached` before commit and keep changes to code, tests, docs, and plan artifacts.

## Next Steps

Use the requested `/ck:ship beta` path after local review passes.
