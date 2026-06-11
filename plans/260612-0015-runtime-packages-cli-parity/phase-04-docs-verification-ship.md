---
phase: 4
title: "Docs Verification Ship"
status: pending
priority: P1
effort: "3h"
dependencies: [2, 3]
---

# Phase 4: Docs Verification Ship

## Overview

Verify the implementation, update user-facing docs, review the diff, and ship a beta PR to `dev`.

## Requirements

- Functional: docs tell users the actual command syntax.
- Non-functional: all tests/build/vet pass before commit and PR.

## Architecture

No runtime architecture changes. Shipping follows beta flow: branch from `dev`, commit, push, PR to `dev`, review PR with fix loop, merge only after CI success.

## Related Code Files

- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Modify: `docs/project-roadmap.md`
- Modify: `docs/codebase-summary.md`
- Modify: `docs/system-architecture.md` if command inventory needs sync.
- Modify: plan status files.

## Implementation Steps

1. Run focused tests:
   - `/usr/local/go/bin/go test ./cmd -run 'TestPackages|TestCredentials|TestAdminCred' -count=1`
2. Run full verification:
   - `/usr/local/go/bin/go test ./...`
   - `/usr/local/go/bin/go build ./...`
   - `/usr/local/go/bin/go vet ./...`
3. Update README/CHANGELOG/docs with package and credential command changes.
4. Run code review on pending diff:
   - spec compliance against this plan.
   - adversarial review for secret exposure, destructive operations, and output contract drift.
5. Commit with a non-`docs`/`chore` conventional message.
6. Push `codex/runtime-packages-cli-parity`.
7. Create PR to `dev`, review with `review-pr --fix`, then merge after checks are green.

## Success Criteria

- [ ] Focused and full Go verification pass.
- [ ] Docs reflect command syntax and output contracts.
- [ ] Pending diff review has no Critical or Important findings.
- [ ] PR exists against `dev`.
- [ ] PR CI green before merge.
- [ ] PR merged into `dev`.

## Risk Assessment

Risk: local shell PATH lacks Go. Mitigation: use `/usr/local/go/bin/go` verified on this machine.
