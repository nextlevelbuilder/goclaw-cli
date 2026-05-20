---
phase: 4
title: "Tests and Docs"
status: complete
priority: P2
effort: "1.5h"
dependencies: [2, 3]
---

# Phase 4: Tests and Docs

## Overview

Add focused regression tests and sync documentation so P5 is discoverable and the parent roadmap reflects the final residual closure.

## Requirements

- Functional: tests cover new commands and failure modes.
- Functional: docs list exact command shapes.
- Non-functional: no broad docs rewrite.
- Non-functional: tests must use real httptest routes, no fake command-only assertions for HTTP contracts.

## Architecture

Test strategy:
- HTTP `httptest.NewServer` for binary download and evolution PATCH route.
- Environment-driven auth via existing test helpers where possible.
- File output tests use `t.TempDir()`.
- Avoid touching real filesystem outside temp dirs.

Docs strategy:
- `README.md`: add concise examples under UX/API/team/evolution sections.
- `CHANGELOG.md`: add P5 entries under Unreleased.
- `docs/codebase-summary.md`: add new files/commands.
- `docs/project-roadmap.md`: mark P5 as planned/in progress only after implementation starts; plan creation can note "P5 execution plan created".
- Parent phase: link this plan and update stale output decisions.
- Parent phase: document verified `agents evolution update` payload mismatch and that P5 fixes it.

## Related Code Files

- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Modify: `docs/codebase-summary.md`
- Modify: `docs/project-roadmap.md`
- Modify: `plans/260503-1907-domain-coverage-p3-plus/plan.md`
- Modify: `plans/260503-1907-domain-coverage-p3-plus/phase-05-fillers-verification-batch-2.md`
- Test: `cmd/p5_fillers_test.go` or existing adjacent test files.

## Implementation Steps

1. Add tests immediately after code changes for each command.
2. Run targeted tests first:
   - `/usr/local/go/bin/go test ./cmd -run 'P5|Attachment|Evolution' -count=1`
3. Run full validation:
   - `/usr/local/go/bin/go build ./...`
   - `/usr/local/go/bin/go test ./...`
   - `/usr/local/go/bin/go vet ./...`
4. Update README and CHANGELOG with exact examples.
5. Update `docs/codebase-summary.md` command table.
6. Update `docs/project-roadmap.md` after implementation status is known.
7. Update parent P5 phase to point at this dedicated execution plan.
8. Document red-team/validation outcomes in plan reports.

## Success Criteria

- [x] Tests prove route, method, request body, file output, and overwrite behavior.
- [x] Full build/test/vet pass.
- [x] Docs mention both P5 commands.
- [x] Parent P5 plan no longer contradicts required `--output`.
- [x] Parent P5 plan notes the required evolution update compatibility fix.

## Risk Assessment

| Risk | Mitigation |
|---|---|
| Tests pass without checking body | Decode request body and assert exact fields. |
| Docs drift from command shape | Copy command examples from tests/help strings. |
| Parent plan stale | Update parent phase in same PR. |
