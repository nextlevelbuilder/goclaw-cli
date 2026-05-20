---
phase: 1
title: "Scope Lock"
status: complete
priority: P1
effort: "45m"
dependencies: []
---

# Phase 1: Scope Lock

## Overview

Lock exact P5 contract before implementation. This phase prevents reintroducing already-covered backlog items or implementing against stale server assumptions.

## Requirements

- Functional: confirm only two residual commands are implemented.
- Functional: verify server route, HTTP method, request body, and auth mode for each command.
- Non-functional: keep implementation CLI-only and small.
- Non-functional: preserve existing command behavior unless a contract mismatch is verified.

## Architecture

Inputs:
- Parent plan: `../260503-1907-domain-coverage-p3-plus/phase-05-fillers-verification-batch-2.md`
- Server route: `/Volumes/GOON/www/nlb/goclaw/internal/http/team_attachments.go`
- Server route: `/Volumes/GOON/www/nlb/goclaw/internal/http/evolution_handlers.go`
- Server helper: `/Volumes/GOON/www/nlb/goclaw/internal/http/evolution_skill_apply.go`

Decision points:
- Use authenticated `GetRaw()` for attachment binary download because the route accepts Bearer auth.
- Use `PATCH` with `status=approved` for skill apply because server dispatches skill creation from suggestion approval.
- Fix existing `agents evolution update` payload. It currently sends stale `action=accept|reject`, while the server requires `status=approved|rejected`.

## Related Code Files

- Read: `cmd/agents_evolution.go`
- Read: `cmd/teams.go`
- Read: `cmd/teams_workspace.go`
- Read: `cmd/agents_lifecycle_test.go`
- Read: `cmd/storage.go`
- Read: `cmd/media.go`
- Read: `internal/client/http.go`

## Implementation Steps

1. Confirm `dev` is clean and synced.
2. Re-run grep on current CLI for the seven P5 sweep items.
3. Reconfirm final residual list is only X11 + X12.
4. Verify server route details:
   - `GET /v1/teams/{teamId}/attachments/{attachmentId}/download`
   - `PATCH /v1/agents/{agentID}/evolution/suggestions/{suggestionID}`
5. Decide file boundaries:
   - Create `cmd/teams_attachments.go`; `cmd/teams_workspace.go` is already near 200 LoC.
   - Extend `cmd/agents_evolution.go`; expected to remain under 200 LoC.
6. Record the verified evolution update mismatch in parent P5 phase before implementation.

## Success Criteria

- [x] Final scope is exactly two new commands plus required evolution update compatibility fix.
- [x] Parent P5 phase links to this dedicated plan.
- [x] No server route changes are planned.
- [x] File ownership is clear before code edits.

## Risk Assessment

| Risk | Mitigation |
|---|---|
| Duplicate planning sources | Make this plan the execution plan and link it from the parent P5 phase. |
| Stale server assumptions | Use server source as source of truth before coding. |
| Hidden scope creep | Reject covered backlog items unless current code proves broken. |
