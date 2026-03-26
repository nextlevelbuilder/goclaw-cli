---
phase: 5
status: complete
priority: high
effort: M
---

# Phase 5: README Sync, Modularization & Tests

## Overview

Sync README with actual implementation, modularize oversized files, and add basic compile/smoke tests.

## Tasks

### 1. README Sync

Update `README.md` command table to reflect ALL implemented commands:
- Add: `tenants`, `system-config`, `knowledge-graph`, `usage`, `activity`, `credentials`, `media`, `packages`, `tts`, `contacts`, `pending-messages`, `heartbeat`
- Update: `skills` (new subcommands), `tools` (tenant-config), `teams` (task workflow), `channels` (writers), `config` (permissions)
- Remove any aspirational commands that are not implemented
- Add examples for new multi-tenant workflow

### 2. Verify Modularization (Done in Phase 3)

<!-- Updated: Validation Session 1 - Modularization moved to Phase 3 -->

Modularization was completed in Phase 3 (before adding new features). Verify no file in `cmd/` exceeds 200 lines. If `skills.go` grew past 200 after Phase 3 additions, split into `skills.go` + `skills_config.go`.

### 3. Basic Tests

Add compile + flag validation tests for new commands:
- Verify all commands register without panic
- Verify required flags produce errors when missing
- Verify `--help` output for all new commands

### 4. Docs Update

Update `docs/codebase-summary.md` and `docs/development-roadmap.md` with new commands and completion status.

## Related Code Files

### Files to Modify
- `README.md`
- `cmd/admin.go` → split
- `cmd/teams.go` → split
- `cmd/agents.go` → split
- `cmd/skills.go` → split
- `docs/codebase-summary.md`
- `docs/development-roadmap.md`

## Implementation Steps

1. Modularize oversized files (mechanical refactor, no logic changes)
2. Update README command table and examples
3. Add basic command registration tests
4. Update docs
5. `go build ./...` && `go test ./...`

## Todo List

- [ ] Split admin.go into subfiles
- [ ] Split teams.go into subfiles
- [ ] Split agents.go into subfiles
- [ ] Split skills.go if needed
- [ ] Update README.md
- [ ] Add command registration tests
- [ ] Update docs/codebase-summary.md
- [ ] Update docs/development-roadmap.md
- [ ] Full build + test pass

## Success Criteria

- No Go file in `cmd/` exceeds 200 lines
- `README.md` matches actual implementation 100%
- `go build ./...` passes
- `go test ./...` passes
- `go vet ./...` clean

## Risk Assessment

- **Low:** Modularization is mechanical — move functions between files, no logic changes
- **Low:** README update is documentation-only
