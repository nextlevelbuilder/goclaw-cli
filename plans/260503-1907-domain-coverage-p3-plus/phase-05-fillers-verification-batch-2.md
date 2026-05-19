# Phase 5 — Fillers & Verification Batch 2

**Priority:** 🟡 medium
**Status:** verified residuals — not-started
**Estimated LoC:** ~150 (excl. tests)
**Depends on:** P3 + P4 merged

## Context Links

- Gap analysis: `plans/reports/brainstorm-260503-1907-gap-analysis-round2.md` § 3.C, 3.F, § 5 (P5)

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
| X8 | `PATCH /v1/agents/{id}/evolution/suggestions/{sid}` | `cmd/agents_evolution.go` | covered: `agents evolution update` |
| X11 | `GET /v1/teams/{teamId}/attachments/{aid}/download` | `cmd/teams.go` / `cmd/teams_*.go` | residual |
| X12 | `internal/http/evolution_skill_apply.go` | `cmd/agents_evolution.go` | residual |

After sweep, drop covered items, finalize scope. Report sweep results in PR description.

## Scope (post-sweep, tentative)

| # | Command | Server route | File |
|---|---|---|---|
| X11 | `teams attachments download <team-id> <att-id> [--out=…]` | `GET …/attachments/{aid}/download` | `cmd/teams_workspace.go` or new `cmd/teams_attachments.go` |
| X12 | `agents evolution skill apply <id> <sid>` | `internal/http/evolution_skill_apply.go` | `cmd/agents_evolution.go` |

## Implementation Steps

1. Extend the existing module file for each confirmed residual (no new files unless >200 LoC pushes existing over budget).
2. `teams attachments download` — binary file save via signed-URL pattern (`internal/client/signed_download.go`); reuse helper.
3. `agents evolution skill apply` — wire the server route and expose structured output.
4. Add `_test.go` cases for each.
5. CHANGELOG + docs sync.

## Todo List

- [x] verify sweep on 7 items (grep CLI command surface)
- [x] drop confirmed-covered items, finalize scope
- [x] cmd/channels_writers.go: groups subcommand
- [x] cmd/channels_contacts.go: unmerge
- [x] cmd/agents_instances.go: list + files
- [x] cmd/mcp_servers.go: tools subcommand
- [ ] cmd/agents_evolution.go: skill apply
- [ ] teams attachments download (reuse signed_download)
- [ ] tests per command
- [ ] CHANGELOG + docs

## Success Criteria

- 100% of post-sweep scope shipped.
- Coverage script reports ≥98% server routes wrapped.
- ≥ 60% line coverage on new code.
- Build + vet + test clean.

## Risk Assessment

| Risk | Mitigation |
|---|---|
| Sweep reveals all items already covered | Phase becomes verification-only — close as docs PR |
| Binary download path collisions | Default `--out=./<filename>`; respect existing file with --force flag |
| `unmerge` destructive without --yes | Require --yes for unmerge |
| Evolution suggestion updates race | Server handles concurrency; CLI passes ETag if available |

## Security Considerations

- Attachment download — write file mode 0644.
- Unmerge requires --yes.

## Next Steps

After P5: coverage ≥98% achieved. P6 (server FRs) tracked as upstream issues only — no CLI work.
