# Phase 5 — Fillers & Verification Batch 2

**Priority:** 🟡 medium
**Status:** implemented pending ship
**Estimated LoC:** ~150 (excl. tests)
**Depends on:** P3 + P4 merged

## Context Links

- Gap analysis: `plans/reports/brainstorm-260503-1907-gap-analysis-round2.md` § 3.C, 3.F, § 5 (P5)
- Detailed execution plan: `../260520-1050-domain-coverage-p5-fillers/plan.md`

## Overview

Final fillers. Verification sweep completed 2026-05-19; most "missing" items were already present under different command paths.

## Verification Sweep (do FIRST)

For each item below, grep both repos and confirm gap before scoping LoC:

| ID | Server source | Verify CLI | Likely outcome |
|---|---|---|---|
| C1 | `GET /v1/channels/instances/{id}/writers/groups` | `cmd/channels_writers.go` | covered: `channels writers groups` |
| C4 | `POST /v1/contacts/unmerge` | `cmd/channels_contacts.go`, `cmd/contacts.go` | covered: `channels contacts unmerge`, `contacts unmerge` |
| X3 | `GET /v1/agents/{id}/instances` + files | `cmd/agents_instances.go` | covered: list/get-file/set-file/metadata |
| X4 | `GET /v1/mcp/servers/{id}/tools` | `cmd/mcp.go` | covered: `mcp servers tools` |
| X8 | `PATCH /v1/agents/{id}/evolution/suggestions/{sid}` | `cmd/agents_evolution.go` | covered as command surface, but payload compatibility fix required in P5 |
| X11 | `GET /v1/teams/{teamId}/attachments/{aid}/download` | `cmd/teams.go` / `cmd/teams_*.go` | residual |
| X12 | `internal/http/evolution_skill_apply.go` | `cmd/agents_evolution.go` | residual |

After sweep, drop covered items, finalize scope. Report sweep results in PR description.

## Scope (post-sweep, tentative)

| # | Command | Server route | File |
|---|---|---|---|
| X11 | `teams attachments download <team-id> <att-id> --output <file>` | `GET …/attachments/{aid}/download` | new `cmd/teams_attachments.go` |
| X12 | `agents evolution skill apply <id> <sid> [--skill-draft @file]` | `PATCH /v1/agents/{id}/evolution/suggestions/{sid}` with `status=approved` | `cmd/agents_evolution.go` |

## Implementation Steps

1. Follow detailed plan in `../260520-1050-domain-coverage-p5-fillers/`.
2. `teams attachments download` — authenticated binary file save via `GetRaw`; require `--output/-o`; add `--force` for overwrite.
3. `agents evolution skill apply` — approve `skill_add` suggestions via existing PATCH route and expose structured output.
4. Fix existing `agents evolution update` payload mapping: `accept -> approved`, `reject -> rejected`.
5. Add `_test.go` cases for each.
6. CHANGELOG + docs sync.

## Todo List

- [x] verify sweep on 7 items (grep CLI command surface)
- [x] drop confirmed-covered items, finalize scope
- [x] cmd/channels_writers.go: groups subcommand
- [x] cmd/channels_contacts.go: unmerge
- [x] cmd/agents_instances.go: list + files
- [x] cmd/mcp_servers.go: tools subcommand
- [x] cmd/agents_evolution.go: update payload compatibility
- [x] cmd/agents_evolution.go: skill apply
- [x] teams attachments download (use authenticated GetRaw)
- [x] tests per command
- [x] CHANGELOG + docs

## Success Criteria

- 100% of post-sweep scope shipped.
- Coverage script reports ≥98% server routes wrapped.
- ≥ 60% line coverage on new code.
- Build + vet + test clean.

## Risk Assessment

| Risk | Mitigation |
|---|---|
| Sweep reveals all items already covered | Phase becomes verification-only — close as docs PR |
| Binary download path collisions | Require `--output/-o`; refuse existing file unless `--force` |
| `unmerge` destructive without --yes | Require --yes for unmerge |
| Evolution suggestion updates race | Server handles concurrency; CLI passes ETag if available |

## Security Considerations

- Attachment download — write file mode 0644; no binary stdout mode.
- Unmerge requires --yes.

## Next Steps

After P5: coverage ≥98% achieved. P6 (server FRs) tracked as upstream issues only — no CLI work.
