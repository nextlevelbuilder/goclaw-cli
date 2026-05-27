---
phase: 6
title: "Tests and Docs Sweep"
status: pending
priority: P2
effort: "1.5h"
dependencies: [5]
---

# Phase 6: Tests and Docs Sweep

## Overview

Final validation gate before ship. Run full test suite, vet, build. Update README and codebase-summary.md. Run a red-team diff sweep against the explicit out-of-scope list. No new functional code.

## Implementation Steps

1. Run full validation:
   - `go test -count=1 ./...`
   - `go vet ./...`
   - `go build ./...`
   - `make build` (LDFLAGS smoke)
2. Update docs:
   - `README.md` — append the 7 new command surfaces to the command inventory section if such a section exists.
   - `docs/codebase-summary.md` — list new files added and which Cobra command groups gained surfaces.
   - `CHANGELOG.md` — add a `## Unreleased` entry (or follow existing semantic-release commit convention; do not manually edit if release notes are commit-driven).
3. Red-team diff sweep (before opening PR):
   - Confirm no command exists for `POST /v1/traces/{id}/replay`.
   - Confirm no command exists for generic `GET /v1/logs/aggregate` (only `/v1/logs/runtime/aggregate`).
   - Confirm no WebSocket `chat.history.delta` consumer added.
   - Confirm no SSE chat history follow added.
   - Confirm no watch loop in `traces follow` or `sessions follow`.
   - Confirm no `--verify` flag added to `providers reconnect`.
   - Confirm every new POST/PATCH path is path-escaped.
   - Confirm no untracked unrelated files (`.claude/`, `AGENTS.md`) are staged.
4. Write `reports/red-team-260527-p6.md` capturing the sweep outcome.

## Todo List

- [ ] `go test`, `go vet`, `go build` all green.
- [ ] `make build` smoke passes.
- [ ] README updated (or noted as not applicable).
- [ ] `docs/codebase-summary.md` updated.
- [ ] CHANGELOG entry (or confirmed commit-driven).
- [ ] Red-team sweep documented in `reports/red-team-260527-p6.md`.
- [ ] Out-of-scope checklist all-clear.
- [ ] Phase status flipped to Complete.

## Success Criteria

- Zero failed tests.
- Zero vet warnings.
- Red-team sweep finds zero scope leaks.

## Out of Scope (verbatim from issue #16, must remain absent)

- `POST /v1/traces/{id}/replay`
- generic `GET /v1/logs/aggregate`
- WebSocket `chat.history.delta`
- SSE chat history follow
- long-running watch loops for `traces follow` or `sessions follow`

## Next Steps

Phase 7 (Ship Readiness).
