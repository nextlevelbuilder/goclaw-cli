---
title: GoClaw CLI Feature Parity Update
status: completed
created: 2026-03-26
priority: high
blockedBy: []
blocks: []
---

# GoClaw CLI — Feature Parity Update

Bring the CLI up to date with GoClaw server's current API surface. Major focus: multi-tenant support, missing resource commands, enhanced existing commands, and new streaming features.

## Context

The GoClaw server has evolved significantly since the initial CLI implementation (March 15, 2026):
- **Multi-tenant isolation** (commit cd022699, March 23) — 30+ tables with `tenant_id`, all queries scoped
- **18+ new HTTP endpoints** — tenants, system-config, packages, contacts, per-tenant config
- **8+ new WS methods** — heartbeat, teams.tasks enhancements, config.permissions
- **Breaking changes** — custom tools removed, skill/tool visibility per-tenant

## Gap Analysis

### Commands in README but NOT implemented
| Command | Endpoints | Priority |
|---------|-----------|----------|
| `knowledge-graph` | 8 HTTP (entities, extract, traverse) | medium |
| `usage` | 4 HTTP (summary, breakdown, timeseries, costs) | high |
| `activity` | 1 HTTP (audit log) | medium |
| `delegations` | partial in admin | low |
| `approvals` | partial in admin | low |
| `credentials` | 6 HTTP (CLI credential store) | medium |
| `tts` | 6 WS methods | low |
| `media` | 2 HTTP (upload, download) | medium |

### New server features NOT in CLI or README
| Feature | Endpoints | Priority |
|---------|-----------|----------|
| `tenants` | 7 HTTP (CRUD, user mgmt) | critical |
| `system-config` | 4 HTTP (per-tenant KV store) | high |
| `packages` | 4 HTTP (install, uninstall, runtimes) | medium |
| `contacts` | 5 HTTP (merge, unmerge, resolve) | low |
| `pending-messages` | 3 HTTP (list, compact, delete) | low |
| `heartbeat` | 7 WS methods (monitoring) | medium |
| Config permissions | 3 WS methods | medium |
| Channel writers | 3 HTTP + WS | medium |
| Per-tenant skill config | 2 HTTP | high |
| Per-tenant tool config | 2 HTTP | high |
| Skill versions/files | 2+ HTTP | medium |
| Provider embedding status | 1 HTTP | low |
| Teams tasks enhancements | 5+ WS methods (comments, events, approve/reject) | high |

### Existing commands needing updates
| Command | Changes Needed |
|---------|---------------|
| All commands | `--tenant-id` flag for multi-tenant context |
| `skills` | add versions, files, tenant-config subcommands |
| `tools` | add tenant-config subcommands |
| `teams` | add task comments, events, approve/reject/assign |
| `channels` | add writers subcommands |
| `providers` | add embedding-status, claude-cli-auth subcommands |
| `storage` | add download flag support, move operation |

## Phases

| # | Phase | Status | Effort | Priority |
|---|-------|--------|--------|----------|
| 1 | [Multi-Tenant Commands](phase-01-multi-tenant-commands.md) | complete | M | critical |
| 2 | [Missing Resource Commands](phase-02-missing-resource-commands.md) | complete | L | high |
| 3 | [Enhanced Existing Commands](phase-03-enhanced-existing-commands.md) | complete | M | high |
| 4 | [New WS & Streaming Features](phase-04-new-ws-streaming-features.md) | complete | S | medium |
| 5 | [README Sync, Modularization & Tests](phase-05-readme-sync-and-tests.md) | complete | M | high |

## Dual Mode Reminder

All new commands must support:
- Interactive: colored tables, confirmation prompts
- Automation: `--output json/yaml`, `--yes` flag, env vars

## Dependencies

- GoClaw server running with multi-tenant support enabled
- Go 1.25+
- Existing CLI internal/ layer (HTTP, WS, config, output, tui) — no changes needed

## Validation Log

### Session 1 — 2026-03-26
**Trigger:** Pre-implementation validation of feature parity plan
**Questions asked:** 6

#### Questions & Answers

1. **[Architecture]** The plan adds `--tenant-id` as a global persistent flag on root command. For single-tenant servers this is unnecessary noise. How should tenant context be handled?
   - Options: Global --tenant-id flag | Per-command --tenant-id flag | Config-based tenant
   - **Answer:** Global --tenant-id flag
   - **Rationale:** Consistent across all commands. Ignored when not set, so no noise for single-tenant users.

2. **[Scope]** 9 missing command groups identified. Phase 2 implements ALL of them. Should we implement all 9, or defer low-priority ones?
   - Options: All 9 now | 6 high/medium only | 4 critical/high only
   - **Answer:** All 9 commands now
   - **Rationale:** Full feature parity in one pass. Complete coverage.

3. **[Architecture]** teams.go is 500 lines, agents.go is 491 lines. When should modularization happen?
   - Options: Before adding features | After all features | Only if >600 lines
   - **Answer:** Modularize BEFORE adding features
   - **Rationale:** Clean files first, then add new subcommands to properly-sized modules.

4. **[Architecture]** Some resources have both HTTP REST and WS RPC endpoints. Which transport should new commands prefer?
   - Options: HTTP REST preferred | WebSocket RPC preferred | Match existing pattern per command
   - **Answer:** Match existing pattern per command
   - **Rationale:** Consistency within each command group matters more than global uniformity.

5. **[Breaking]** Server removed custom tools. CLI still has `goclaw tools custom` commands. Should we remove them?
   - Options: Remove custom tool commands | Keep but deprecate | Keep as-is
   - **Answer:** Remove custom tool commands
   - **Rationale:** Match server reality. Clean up dead code.

6. **[Scope]** GoClaw has OpenAI-compatible endpoints (`/chat/completions`, `/responses`). Should CLI support these?
   - Options: Skip for now | Hidden commands | Full commands
   - **Answer:** Skip for now
   - **Rationale:** These are for external integrations (Cursor, Continue.dev), not CLI users.

#### Confirmed Decisions
- **Tenant flag:** Global `--tenant-id` persistent flag on root command
- **Scope:** Implement all 9 missing command groups
- **Modularization:** Split oversized files BEFORE adding Phase 3 features
- **Transport:** Match existing transport per command group
- **Custom tools:** Remove dead `tools custom` commands
- **OpenAI compat:** Not in scope

#### Action Items
- [ ] Phase 3: Add modularization step BEFORE feature additions
- [ ] Phase 3: Add step to remove `tools custom` commands
- [ ] Phase 2: Confirm all 9 groups stay in scope

#### Impact on Phases
- Phase 3: Reorder — modularize teams.go, agents.go, admin.go FIRST, then add features. Also remove `tools custom` commands.
- Phase 5: Modularization already done in Phase 3, so Phase 5 only needs README + tests.
