# Phase 5 — Fillers & Verification Batch 2

**Priority:** 🟡 medium
**Status:** not-started
**Estimated LoC:** ~250 (excl. tests; may shrink after verify sweep)
**Depends on:** P3 + P4 merged

## Context Links

- Gap analysis: `plans/reports/brainstorm-260503-1907-gap-analysis-round2.md` § 3.C, 3.F, § 5 (P5)

## Overview

Final fillers. **Begins with a 30-min verification sweep** — several "missing" items are flagged "verify" because grep matched server file but not CLI registration. Some may already exist under different command paths.

## Verification Sweep (do FIRST)

For each item below, grep both repos and confirm gap before scoping LoC:

| ID | Server source | Verify CLI | Likely outcome |
|---|---|---|---|
| C1 | `GET /v1/channels/instances/{id}/writers/groups` | `cmd/channels_writers.go` Use list | confirmed gap |
| C4 | `POST /v1/contacts/unmerge` | `cmd/channels_contacts.go` | likely gap |
| X3 | `GET /v1/agents/{id}/instances` + files | `cmd/agents_instances.go` | partial — verify sub-routes |
| X4 | `GET /v1/mcp/servers/{id}/tools` | `cmd/mcp_servers.go` | likely covered |
| X8 | `PATCH /v1/agents/{id}/evolution/suggestions/{sid}` | `cmd/agents_evolution.go` | likely missing |
| X11 | `GET /v1/teams/{teamId}/attachments/{aid}/download` | `cmd/teams.go` / `cmd/teams_*.go` | likely missing |
| X12 | `internal/http/evolution_skill_apply.go` | `cmd/agents_evolution.go` | likely missing |

After sweep, drop covered items, finalize scope. Report sweep results in PR description.

## Scope (post-sweep, tentative)

| # | Command | Server route | File |
|---|---|---|---|
| C1 | `channels writers groups <id>` | `GET /v1/channels/instances/{id}/writers/groups` | `cmd/channels_writers.go` |
| C4 | `contacts unmerge <merge-id>` | `POST /v1/contacts/unmerge` | `cmd/channels_contacts.go` |
| X3 | `agents instances list <agent-id>` + `agents instances files <agent-id> <user-id>` | server | `cmd/agents_instances.go` |
| X4 | `mcp servers tools <id>` | `GET /v1/mcp/servers/{id}/tools` | `cmd/mcp_servers.go` |
| X8 | `agents evolution suggestions update <id> <sid>` | `PATCH …/suggestions/{sid}` | `cmd/agents_evolution.go` |
| X11 | `teams attachments download <team-id> <att-id> [--out=…]` | `GET …/attachments/{aid}/download` | `cmd/teams_workspace.go` or new `cmd/teams_attachments.go` |
| X12 | `agents evolution skill apply <id> <sid>` | `internal/http/evolution_skill_apply.go` | `cmd/agents_evolution.go` |

## Implementation Steps

1. **Sweep first.** Use Grep on `cmd/*.go` for each route literal + each `Use:` line. Document results in PR description.
2. For each confirmed gap, extend the existing module file (no new files unless >200 LoC pushes existing over budget).
3. `teams attachments download` — binary file save via signed-URL pattern (`internal/client/signed_download.go`); reuse helper.
4. Add `_test.go` cases for each.
5. CHANGELOG + docs sync.

## Todo List

- [ ] verify sweep on 7 items (grep both repos)
- [ ] drop confirmed-covered items, finalize scope
- [ ] cmd/channels_writers.go: groups subcommand
- [ ] cmd/channels_contacts.go: unmerge
- [ ] cmd/agents_instances.go: list + files (if missing)
- [ ] cmd/mcp_servers.go: tools subcommand
- [ ] cmd/agents_evolution.go: suggestions update + skill apply
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
