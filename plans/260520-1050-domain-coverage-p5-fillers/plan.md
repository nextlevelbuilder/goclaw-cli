---
title: "Domain Coverage P5 Fillers"
description: "Finish the final CLI-only P5 residuals: team attachment download and evolution skill apply."
status: in-progress
priority: P2
branch: "dev"
tags: [domain-coverage, p5, cli]
blockedBy: []
blocks: [260503-1907-domain-coverage-p3-plus]
created: "2026-05-20T03:50:47.956Z"
createdBy: "ck:plan"
source: skill
---

# Domain Coverage P5 Fillers

## Overview

Implement the final post-sweep CLI residuals from Domain Coverage P5 in one PR from `dev`.
Scope is intentionally small: add a direct team attachment download command and a clear evolution skill-apply wrapper over the existing server approval contract.

## Decisions

- Scope: one PR with both P5 commands.
- Attachment download: `--output/-o` is required. No implicit filename guessing in this round.
- Evolution skill apply: wrapper only for approving `skill_add` suggestions; optional `--skill-draft` override.
- Do not add new server routes. Server already exposes the required contracts.
- Repair existing evolution update payload: current CLI sends stale `action`, while server requires `status`.

## Phases

| Phase | Name | Status |
|-------|------|--------|
| 1 | [Scope Lock](./phase-01-scope-lock.md) | Complete |
| 2 | [Team Attachment Download](./phase-02-team-attachment-download.md) | Complete |
| 3 | [Evolution Skill Apply](./phase-03-evolution-skill-apply.md) | Complete |
| 4 | [Tests and Docs](./phase-04-tests-and-docs.md) | Complete |
| 5 | [Ship Readiness](./phase-05-ship-readiness.md) | Pending |

## Dependencies

- Parent backlog: `../260503-1907-domain-coverage-p3-plus/plan.md`
- P5 legacy phase: `../260503-1907-domain-coverage-p3-plus/phase-05-fillers-verification-batch-2.md`
- Server evidence:
  - `/Volumes/GOON/www/nlb/goclaw/internal/http/team_attachments.go`
  - `/Volumes/GOON/www/nlb/goclaw/internal/http/evolution_handlers.go`
  - `/Volumes/GOON/www/nlb/goclaw/internal/http/evolution_skill_apply.go`

## Command Contract

### Team Attachment Download

```bash
goclaw teams attachments download <team-id> <attachment-id> --output ./artifact.bin
goclaw teams attachments download <team-id> <attachment-id> -o ./artifact.bin --force
```

Behavior:
- Calls `GET /v1/teams/{teamId}/attachments/{attachmentId}/download` using normal Bearer auth.
- Requires `--output/-o`; refuses missing output before network call.
- Refuses overwrite unless `--force` is set.
- Creates parent directory when needed.
- Streams binary response to disk.

### Evolution Skill Apply

```bash
goclaw agents evolution skill apply <agent-id> <suggestion-id>
goclaw agents evolution skill apply <agent-id> <suggestion-id> --skill-draft @./SKILL.md
```

Behavior:
- Calls `PATCH /v1/agents/{agentID}/evolution/suggestions/{suggestionID}`.
- Preflights pending suggestions and refuses non-`skill_add` suggestion IDs.
- Sends `{"status":"approved"}` by default.
- Adds `skill_draft` when `--skill-draft` is provided; supports literal content or `@file` via existing `readContent()`.
- Prints structured server response with `printer.Print(unmarshalMap(data))`.
- Uses escaped path segments for agent and suggestion IDs.

## Scope Boundary

In scope:
- Two new CLI surfaces above.
- Tests for route, body, file writing, required output, overwrite guard, and draft override.
- Docs sync in `CHANGELOG.md`, `README.md`, `docs/codebase-summary.md`, `docs/project-roadmap.md`, and parent P5 phase.

Out of scope:
- New server routes.
- Generic "apply any evolution suggestion" command.
- Signed URL discovery from team task detail.
- Auto filename extraction from `Content-Disposition`.
- Large refactors of teams or evolution command groups.

## Validation Gates

- `/usr/local/go/bin/go build ./...` — pass 2026-05-20
- `/usr/local/go/bin/go test ./...` — pass 2026-05-20
- `/usr/local/go/bin/go vet ./...` — pass 2026-05-20
- Manual PR review against server source contracts.

## Red Team Review

### Findings

| ID | Severity | Finding | Evidence | Disposition |
|---|---|---|---|---|
| RT-01 | High | Existing `agents evolution update` is not just "maybe stale"; it is verified stale and must be fixed with P5. | `cmd/agents_evolution.go:88-90`; `/Volumes/GOON/www/nlb/goclaw/internal/http/evolution_handlers.go:199-213` | Accepted |
| RT-02 | Medium | Evolution PATCH paths should escape IDs. New code must not copy the current unescaped `fmt.Sprintf` path pattern. | `cmd/agents_evolution.go:88-89`; `cmd/api_keys_rotate.go:44`; `cmd/storage.go:55` | Accepted |
| RT-03 | Medium | Binary download plan needs explicit `resp.Body.Close()` after successful status check, before file-copy returns. | `cmd/storage.go:62`; `cmd/media.go:55`; `cmd/helpers.go:120-126` | Accepted |
| RT-04 | Low | Optional `--force` was not explicitly in the user decision, but it is low-risk and bounded because default remains no-overwrite. | `plans/260520-1050-domain-coverage-p5-fillers/phase-02-team-attachment-download.md:19-20` | Accepted |
| CR-01 | Medium | Local attachment `--output` flag can shadow the root output-format flag unless config resolution reads the persistent root flag. | `cmd/root.go`; `internal/config/config.go`; `cmd/teams_attachments.go:91` | Accepted |
| CR-02 | Medium | Skill-specific wrapper must not approve non-`skill_add` suggestions through the generic server route. | `/Volumes/GOON/www/digitop/goclaw/internal/http/evolution_handlers.go:223-245` | Accepted |

### Whole-Plan Consistency Sweep

- Files reread: `plan.md`, `phase-01-scope-lock.md`, `phase-02-team-attachment-download.md`, `phase-03-evolution-skill-apply.md`, `phase-04-tests-and-docs.md`, `phase-05-ship-readiness.md`.
- Decision deltas checked: stale `action` payload, escaped IDs, response close, `--force` no-overwrite default.
- Reconciled stale references: optional update fix converted to required update fix.
- Unresolved contradictions: 0

## Validation Log

### Verification Results

- **Tier:** Full
- **Claims checked:** 24
- **Verified:** 24 | **Failed:** 0 | **Unverified:** 0

Verified examples:
- Team attachment route exists and accepts Bearer auth: `/Volumes/GOON/www/nlb/goclaw/internal/http/team_attachments.go:25-53`.
- Team attachment route validates team and attachment IDs: `/Volumes/GOON/www/nlb/goclaw/internal/http/team_attachments.go:64-84`.
- Server exposes `Content-Disposition`, but user decision requires explicit `--output`: `/Volumes/GOON/www/nlb/goclaw/internal/http/team_attachments.go:104-112`.
- Evolution PATCH route exists: `/Volumes/GOON/www/nlb/goclaw/internal/http/evolution_handlers.go:65-69`.
- Evolution PATCH body requires `status`, not `action`: `/Volumes/GOON/www/nlb/goclaw/internal/http/evolution_handlers.go:199-213`.
- Skill creation dispatches when `status=approved` and suggestion type is `SuggestSkillAdd`: `/Volumes/GOON/www/nlb/goclaw/internal/http/evolution_handlers.go:223-232`.

### Critical Questions

| Question | Decision |
|---|---|
| Should P5 ship both residual commands in one PR? | Yes, user confirmed both commands in one PR. |
| Should attachment download infer filename from `Content-Disposition`? | No, user confirmed `--output/-o` is required. |
| Should evolution skill apply be generic or a skill-specific wrapper? | Skill-specific wrapper only, user confirmed recommended approach. |
| Should stale `agents evolution update` payload be handled? | Yes, validation found it is a verified contract mismatch, so fix with P5. |

### Whole-Plan Consistency Sweep

- Files reread: `plan.md`, all five phase files.
- Decision deltas checked: required output, skill-specific apply, status payload, path escaping, response close.
- Reconciled stale references: 5.
- Unresolved contradictions: 0

## Handoff

Recommended next gate after implementation: `/ck:git cp` then `/ck:ship beta`.
