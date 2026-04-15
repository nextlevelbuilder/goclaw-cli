---
phase: 4
title: Agent Lifecycle + Chat + Teams + Memory KG (AI-Critical Max Polish)
status: pending
priority: high
blockedBy: [phase-00]
splitRecommendation: "4a (agents+chat), 4b (teams+memory) if >1000 LoC"
---

# Phase 4 — Agent Lifecycle + Chat + Teams + Memory KG

## Context Links
- Brainstorm §3 (chat, teams, memory partial gaps), §2 (G14-G24 agent lifecycle), §5 Phase 4
- **AI-criticality matrix (brainstorm §5.1): chat history/inject, agents wait/identity, memory KG full = MAXIMUM POLISH**
- Server WS: `../../../goclaw/internal/gateway/methods/{chat.go,teams.go,teams_tasks.go,teams_workspace.go,agents.go,agents_identity.go}`
- Server HTTP: `../../../goclaw/internal/http/{agents.go,agents_sharing.go,agents_instances.go,agents_prompt_preview.go,agents_export.go,memory.go,memory_handlers.go,knowledge_graph.go,knowledge_graph_handlers.go,wake.go,evolution_handlers.go,episodic_handlers.go,v3_flags_handlers.go,orchestration_handlers.go,agents_codex_pool.go,team_events.go}`
- Methods: `../../../goclaw/pkg/protocol/methods.go`

## Overview
- **Priority:** HIGH — highest AI-value phase
- **Status:** Pending
- **Description:** Fill gaps trong 4 existing groups. AI orchestration core: chat history/inject, agents wait/identity, memory KG search. Plus lifecycle extensions (wake/evolution/episodic/v3-flags/orchestration), teams tasks advanced ops, global memory.
- **Split recommendation:** If total LoC >1000, split thành PR 4a (agents+chat) và 4b (teams+memory). Mặc định keep thành 1 PR nếu được.

## Key Insights
- `chat.history` là CRITICAL cho AI orchestration — render history as JSON array of messages (structured)
- `chat.inject` inserts message vào context **without** agent processing — useful cho AI tool truyền context
- `agents wait` là blocking call — support `--timeout` + `--state=<target>` filter
- Memory KG có 12 endpoints — full CRUD + traverse + dedup + graph. Refactor existing `cmd/memory.go` to split
- `teams tasks delete-bulk` khác `delete` — accepts multiple IDs; expose `--ids=<csv>` flag
- File size: existing `cmd/agents.go` 416 LoC + `cmd/teams.go` 462 LoC + `cmd/memory.go` ~175 LoC → **MUST split as part of this phase** per project rule

## Requirements

### Functional — Agent Lifecycle (extensions to existing `agents` group)
- `agents wake <id>` — POST /v1/agents/{id}/wake
- `agents wait <key> [--state=online|running|idle] [--timeout=<dur>]` — WS `agent.wait`
- `agents identity <key>` — WS `agent.identity.get`
- `agents sync-workspace` — POST /v1/agents/sync-workspace (admin)
- `agents prompt-preview <id>` — GET /v1/agents/{id}/system-prompt-preview
- `agents evolution metrics <id>`, `agents evolution suggestions <id>`, `agents evolution update <id> <suggestionID> --action=<accept|reject>`
- `agents episodic list <id>`, `agents episodic search <id> <query>`
- `agents v3-flags get <id>`, `agents v3-flags toggle <id> --flag=<name>`
- `agents orchestration <id>` — GET orchestration mode
- `agents instances set-file <id> --user=<userID> --file=<name> --content=... | --file=...`
- `agents instances update-metadata <id> --user=<userID> --metadata=<json>`
- `agents codex-pool-activity <id>` — agent version
  (provider version trong P5)

### Functional — Chat
- `chat history <agent> [--limit=N] [--before=<ts>]` → WS `chat.history`
- `chat inject <agent> --role=user|assistant|system --content=...` → WS `chat.inject`
- `chat session-status <agent>` → WS `chat.session.status`

### Functional — Teams (extend existing)
- `teams tasks delete <teamID> <taskID>` → `teams.tasks.delete`
- `teams tasks delete-bulk <teamID> --ids=<csv>` → `teams.tasks.delete-bulk`
- `teams tasks events <teamID> <taskID> [--follow]` → `teams.tasks.events` with P0 FollowStream
- `teams tasks get-light <teamID> <taskID>` → `teams.tasks.get-light`
- `teams tasks active --session=<key>` → `teams.tasks.active-by-session`
- `teams scopes <teamID>` → `teams.scopes`
- `teams events list <teamID> [--follow]` → `teams.events.list` hoặc `GET /v1/teams/{id}/events`

### Functional — Memory KG (full)
- `memory kg entities list <agent>`, `get <agent> <entityID>`, `upsert <agent> --file=...`, `delete <agent> <entityID>`
- `memory kg traverse <agent> --from=<entityID> --depth=N`
- `memory kg stats <agent>`
- `memory kg graph <agent> [--compact]`
- `memory kg dedup scan <agent>`, `list <agent>`, `merge <agent> <entityA> <entityB>`, `dismiss <agent> <candidateID>`
- `memory chunks <agent>` — list chunks
- `memory index <agent> <path>` — trigger reindex
- `memory index-all <agent>`
- `memory documents-global` — GET /v1/memory/documents (no agent scope)

### Non-Functional
- **Maximum polish** cho AI-critical commands:
  - `chat history/inject/session-status`, `agents wait/identity`, `memory KG full`
  - JSON examples trong `--help` Long
  - ≥80% test coverage
  - Explicit error cases documented
- `--follow` cho events commands use P0 helper
- Destructive: `kg entities delete`, `kg merge`, `teams tasks delete-bulk` require `--yes`

## Architecture

### Modularization (mandatory per project rules)
```
cmd/
├── agents.go                   # root + list/get/create/update/delete/share/unshare (trim to <200 LoC)
├── agents_links.go             # existing, unchanged
├── agents_instances.go         # list/get-file/set-file/metadata (extracted)
├── agents_lifecycle.go         # wake/wait/identity/sync-workspace/prompt-preview
├── agents_evolution.go         # evolution metrics/suggestions
├── agents_episodic.go          # episodic list/search
├── agents_v3_flags.go          # v3-flags get/toggle
├── agents_misc.go              # orchestration/codex-pool-activity (if small)
├── chat.go                     # send/abort + new history/inject/session-status (existing + extend)
├── teams.go                    # root + list/get/create/update/delete (trim)
├── teams_members.go            # members add/remove/list (extract)
├── teams_tasks.go              # tasks CRUD + delete-bulk + events + get-light + active + assign
├── teams_workspace.go          # workspace list/read/delete (extract)
├── teams_events.go             # events list
├── teams_scopes.go             # scopes
├── memory.go                   # root + documents + search (trim)
├── memory_kg.go                # KG entities CRUD + traverse + stats + graph
├── memory_kg_dedup.go          # dedup scan/list/merge/dismiss
└── memory_index.go             # chunks/index/index-all/documents-global
```

Total new files: ~15, rename/refactor existing 4. Use `git mv` equivalent carefully.

## Related Code Files

### Create
- `cmd/agents_instances.go` (extract from agents.go)
- `cmd/agents_lifecycle.go`
- `cmd/agents_evolution.go`
- `cmd/agents_episodic.go`
- `cmd/agents_v3_flags.go`
- `cmd/agents_misc.go`
- `cmd/teams_members.go` (extract)
- `cmd/teams_tasks.go` (extract + extend)
- `cmd/teams_workspace.go` (extract)
- `cmd/teams_events.go`
- `cmd/teams_scopes.go`
- `cmd/memory_kg.go`
- `cmd/memory_kg_dedup.go`
- `cmd/memory_index.go`

### Modify
- `cmd/agents.go` — remove extracted code, keep CRUD only
- `cmd/chat.go` — add history/inject/session-status subcommands
- `cmd/teams.go` — remove extracted code, keep CRUD only
- `cmd/memory.go` — remove extracted, keep root + documents + search
- `cmd/root.go` — register new subcommand trees (most already registered via parent groups)
- Test files: corresponding `*_test.go` for each new file

### Reference
- `internal/client/follow.go` (P0 helper)
- `internal/output/*` (P0 patterns)

## Implementation Steps

### Phase 4a — Agents + Chat (suggested PR split if too large)

#### Step 1: Agents modularization
1. Create `agents_instances.go`, extract instance commands from `agents.go`
2. Create `agents_lifecycle.go` với wake/wait/identity/sync-workspace/prompt-preview
3. Create `agents_evolution.go` với evolution CRUD
4. Create `agents_episodic.go` với list/search
5. Create `agents_v3_flags.go` với get/toggle
6. Create `agents_misc.go` với orchestration + codex-pool-activity
7. Verify `agents.go` now <200 LoC

#### Step 2: Chat extensions
1. Add `chatHistoryCmd` to `cmd/chat.go` — WS `chat.history` with `--limit`/`--before`
2. Add `chatInjectCmd` — WS `chat.inject` with `--role`/`--content`
3. Add `chatSessionStatusCmd` — WS `chat.session.status`

#### Step 3: Agents tests
1. Per-file tests for each new extracted file
2. Full coverage for wait (blocking + timeout), identity, history

### Phase 4b — Teams + Memory KG

#### Step 4: Teams modularization
1. Create `teams_members.go`, extract from `teams.go`
2. Create `teams_tasks.go`, extract + add delete/delete-bulk/events/get-light/active
3. Create `teams_workspace.go`, extract workspace
4. Create `teams_events.go` với events list
5. Create `teams_scopes.go` với scopes
6. Verify `teams.go` <200 LoC

#### Step 5: Memory KG
1. Create `memory_kg.go` với entities list/get/upsert/delete + traverse + stats + graph
2. Create `memory_kg_dedup.go` với dedup scan/list/merge/dismiss
3. Create `memory_index.go` với chunks/index/index-all/documents-global
4. Refactor `memory.go` to keep only root + documents (agent-scoped) + search

#### Step 6: Tests
1. Teams tests (delete, delete-bulk, events follow)
2. Memory KG tests (entities CRUD, traverse, dedup)
3. Index/chunks tests

### Shared Steps

#### Step 7: Docs
1. README update with AI-orchestration examples (chat history → inject workflow)
2. `docs/codebase-summary.md` new file structure
3. `--help` examples for AI-critical commands (JSON output structure)

#### Step 8: Manual polish check
1. For each AI-critical command, verify:
   - JSON schema documented in `--help` Long
   - Error cases tested
   - Exit codes map correctly
   - No noise on stdout

## Todo List

### 4a — Agents + Chat
- [ ] 1.1: Extract `agents_instances.go` (list/get-file/set-file/update-metadata)
- [ ] 1.2: Create `agents_lifecycle.go` wake/wait/identity/sync-workspace/prompt-preview
- [ ] 1.3: Create `agents_evolution.go` metrics/suggestions/update
- [ ] 1.4: Create `agents_episodic.go` list/search
- [ ] 1.5: Create `agents_v3_flags.go` get/toggle
- [ ] 1.6: Create `agents_misc.go` orchestration/codex-pool-activity
- [ ] 1.7: Verify `agents.go` <200 LoC, compile passes
- [ ] 2.1: `chat history` with `--limit`/`--before`
- [ ] 2.2: `chat inject` with `--role`/`--content`
- [ ] 2.3: `chat session-status`
- [ ] 3.1: Tests for all extracted agents files
- [ ] 3.2: Tests for chat history/inject/session-status
- [ ] 3.3: Special attention: `agents wait` blocking + timeout behavior tests

### 4b — Teams + Memory
- [ ] 4.1: Extract `teams_members.go`
- [ ] 4.2: Create `teams_tasks.go` with delete/delete-bulk/events(follow)/get-light/active-by-session
- [ ] 4.3: Extract `teams_workspace.go`
- [ ] 4.4: Create `teams_events.go` list (follow)
- [ ] 4.5: Create `teams_scopes.go`
- [ ] 4.6: Verify `teams.go` <200 LoC
- [ ] 5.1: Create `memory_kg.go` entities + traverse + stats + graph
- [ ] 5.2: Create `memory_kg_dedup.go`
- [ ] 5.3: Create `memory_index.go` chunks/index/index-all/documents-global
- [ ] 5.4: Trim `memory.go`
- [ ] 6.1: Teams tests (follow stream tests critical)
- [ ] 6.2: Memory KG tests
- [ ] 6.3: Verify ≥80% coverage on AI-critical commands

### Shared
- [ ] 7.1: README AI orchestration section with workflow example
- [ ] 7.2: docs/codebase-summary.md new structure
- [ ] 7.3: JSON schema in `--help` for AI-critical commands
- [ ] 8.1: Manual smoke test AI-critical commands
- [ ] 8.2: Build/vet/test all pass
- [ ] 8.3: Decide 1 PR or split 4a/4b based on final LoC

## Success Criteria

### AI-Critical (Maximum polish)
- [ ] `goclaw chat history <agent> --limit=50 --output=json` returns structured array
- [ ] `goclaw chat inject <agent> --role=system --content="context"` injects without triggering response
- [ ] `goclaw chat session-status <agent>` returns current session state JSON
- [ ] `goclaw agents wait <key> --state=idle --timeout=30s` blocks then exits 0 on state match, exit 6 on timeout
- [ ] `goclaw agents identity <key>` returns identity JSON
- [ ] `goclaw memory kg entities list <agent>` returns entity array
- [ ] `goclaw memory kg traverse <agent> --from=X --depth=2` returns graph slice
- [ ] `goclaw memory kg graph <agent> --compact` returns compact graph

### Standard
- [ ] All agents lifecycle subcommands working
- [ ] Teams tasks delete/delete-bulk/events/get-light/active working
- [ ] Memory KG dedup full workflow testable
- [ ] `teams tasks events --follow` streams JSON lines
- [ ] All commands respect P0 exit codes + error format

### Quality
- [ ] No file >200 LoC in `cmd/`
- [ ] `go build ./... && go vet ./... && go test ./...` pass
- [ ] Coverage: ≥80% AI-critical, ≥60% others
- [ ] Docs updated

## Risk Assessment

| Risk | Mitigation |
|---|---|
| File split breaks imports / circular deps | Test compile after each extraction; keep helpers in shared `cmd/helpers.go` |
| `agents wait` blocking leaks goroutines on timeout | Use context with cancel; cleanup on ctx.Done() |
| `chat inject` misuse (inject fake assistant msg) | Document security implications in `--help` |
| KG delete cascade unexpected | Server-side cascade warning; CLI passes through |
| `delete-bulk` accepts 1000 IDs, long running | `--batch-size=N` flag, default 100 |
| Memory index triggers expensive server op | Display warning message, support `--no-wait` |
| Events follow reconnect during server restart | Inherit P0 FollowStream backoff |
| Split 4a/4b decision late | Check LoC at end of step 3; decide before PR open |

## Security Considerations
- `chat inject` can inject arbitrary system messages → admin-only scope verify server-side
- `agents sync-workspace` affects all agents — admin-only
- `kg entities upsert` with untrusted JSON — server validates schema; CLI passes through
- `agents v3-flags toggle` may enable experimental features — warn in `--help`
- Memory index-all triggers expensive reindex — rate-limited server-side

## Next Steps
- Dependencies: Phase 0
- Unblocks: None
- Follow-up: If split, merge 4a first, 4b sau. Otherwise single PR.

## Unresolved Questions
1. Decide Phase 4 as **single PR** or **split 4a/4b** — criterion: final LoC (single if <1200, split if >1200)
2. `agents episodic search`: server returns similarity scores? CLI display?
3. `kg merge` outcome: hard delete of merged entity, hay soft marker? Check server
4. Events stream format: structured events from server có `type` field để CLI filter không?
5. `teams tasks delete-bulk` idempotency: re-run với same IDs có lỗi hay silent?
