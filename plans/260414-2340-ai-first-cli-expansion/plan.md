---
title: AI-First CLI Expansion — Fill Server Coverage Gaps
status: completed
created: 2026-04-14
completed: 2026-04-15
priority: high
blockedBy: []
blocks: []
---

# AI-First CLI Expansion

Fill ~30% server coverage gaps trong goclaw-cli. Primary consumer: AI tools (Claude Code, LangChain). Foundation refactor + 5 feature phases, each a standalone PR.

## Source
- Brainstorm: `plans/reports/brainstorm-260414-2231-missing-commands-gap-analysis.md`
- Server audit: `../goclaw/pkg/protocol/{errors.go,frames.go,methods.go}`, `../goclaw/internal/http/response_helpers.go`

## Locked Decisions
- CLI-only (no MCP server); ~90% AI consumer
- Exit codes: 0/1/2/3/4/5/6 mapped to server error codes
- TTY auto-switch output default (breaking change, document)
- DROP idempotency-key (server chưa support)
- Error pass-through using server's `ErrorShape`
- Skip Zalo/WhatsApp/TTS-convert/browser automation
- Include full vault group

## Phase Overview

| # | Phase | Status | Size | Score | Risk | AI-Critical? |
|---|---|---|---|---|---|---|
| 0 | [Ergonomics Foundation](phase-00-ai-ergonomics-foundation.md) | **completed** 2026-04-15 | ~700 LoC | 9.4/10 | Medium (refactor wide) | 🔥 Yes (prerequisite) |
| 1 | [Admin/Ops Foundation](phase-01-admin-ops-foundation.md) | **completed** 2026-04-15 | ~600 LoC | 7.8/10 | Low | 🟡 Mixed |
| 2 | [Migration (Backup/Restore + Export/Import)](phase-02-migration.md) | **completed** 2026-04-15 | ~800 LoC | 7.8/10 | **High** (destructive) | 🟡 Mixed |
| 3 | [Vault](phase-03-vault.md) | **completed** 2026-04-15 | ~700 LoC | 7.2/10 | Medium | 🔥 Yes (search) |
| 4 | [Agent Lifecycle + Chat + Teams + Memory KG](phase-04-agent-lifecycle-chat-teams-memory.md) | **completed** 2026-04-15 | ~3300 LoC | 8.6/10 | Low-Medium | 🔥 **Maximum** |
| 5 | [Extensions + Remaining](phase-05-extensions-and-remaining.md) | **completed** 2026-04-15 | ~1900 LoC (single PR) | pending review | Medium | 🟡 Mixed |

**Coverage metrics (Phase 0):**
- `internal/output`: 97.3% test coverage
- `internal/client`: 71.3% test coverage
- Code review score: 9.4/10
- Builds: `go build ./... && go vet ./... && go test ./...` all pass
- Breaking change: TTY auto-detect (documented in CHANGELOG.md)

## Sequencing

```
P0 (blocks all) ──┐
                  ├─> P1 ─┐
                  ├─> P2 ─┤
                  ├─> P3 ─┼─> Merge sequentially
                  ├─> P4 ─┤
                  └─> P5 ─┘
```

**P0 MUST merge first** — establishes error/exit/follow/TTY patterns that P1-P5 inherit. P1-P5 có thể dev parallel (different file owners) nhưng merge sequential để tránh conflict trong README/docs.

## Key Dependencies
- Go 1.25 (existing)
- Cobra, gorilla/websocket (existing)
- `golang.org/x/term` for TTY detection (likely new import in P0)
- No new external deps required

## Success Criteria (Overall)
- ≥95% server endpoint coverage (measured by script count)
- All 6 phases merged, each standalone test-passing
- AI-critical commands (chat history/inject, memory KG, agents wait/identity, vault search, send) full-polish với JSON examples trong `--help`
- Exit codes chuẩn hóa trên toàn bộ commands
- README + docs/codebase-summary.md reflect final state

## Related Files Summary
- New command files: `cmd/{tenants,heartbeat,system_configs,edition,backup,restore,migrate,vault,pair,oauth,packages,users,quota,send}.go`
- Refactor: `cmd/{agents,teams,memory,chat,skills,mcp,providers,tools,channels,admin,config_cmd}.go`
- New internal: `internal/output/{error.go,exit.go}`, `internal/client/follow.go`
- Docs: `docs/codebase-summary.md`, `README.md`, `CHANGELOG.md` (new)
