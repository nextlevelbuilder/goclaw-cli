---
phase: 3
status: complete
priority: high
effort: M
---

# Phase 3: Enhanced Existing Commands

## Overview

Update existing commands with new subcommands and features added to the GoClaw server since initial implementation.

<!-- Updated: Validation Session 1 - Modularize BEFORE adding features; Remove custom tools -->

**IMPORTANT:** Before adding any new subcommands, modularize oversized files first:
1. Split `cmd/teams.go` (500 lines) → `cmd/teams.go` + `cmd/teams_tasks.go` + `cmd/teams_events.go`
2. Split `cmd/agents.go` (491 lines) → `cmd/agents.go` + `cmd/agents_instances.go` + `cmd/agents_links.go`
3. Split `cmd/admin.go` (403 lines) → `cmd/admin.go` + `cmd/admin_users.go` + `cmd/admin_audit.go`
4. Remove `tools custom` commands (server removed custom tools — only builtin remain)

Then proceed with feature additions below.

## Changes by Command

### 1. `skills` — Add tenant config + versions + files

New subcommands:
```
goclaw skills versions <skill-id>
goclaw skills files <skill-id> [path]
goclaw skills tenant-config set <skill-id> --enabled <bool>
goclaw skills tenant-config delete <skill-id>
goclaw skills install-dep --skill-id <id> --dep <name>
goclaw skills rescan-deps --skill-id <id>
goclaw skills runtimes
```

HTTP Endpoints:
```
GET  /v1/skills/{id}/versions
GET  /v1/skills/{id}/files
GET  /v1/skills/{id}/files/{path...}
PUT  /v1/skills/{id}/tenant-config
DELETE /v1/skills/{id}/tenant-config
POST /v1/skills/install-dep
POST /v1/skills/install-deps
POST /v1/skills/rescan-deps
GET  /v1/skills/runtimes
```

### 2. `tools` — Add tenant config

New subcommands:
```
goclaw tools builtin tenant-config set <name> --enabled <bool>
goclaw tools builtin tenant-config delete <name>
```

HTTP Endpoints:
```
PUT    /v1/tools/builtin/{name}/tenant-config
DELETE /v1/tools/builtin/{name}/tenant-config
```

### 3. `teams` — Add task comments, events, approve/reject

New subcommands:
```
goclaw teams tasks approve <team-id> <task-id>
goclaw teams tasks reject <team-id> <task-id> [--reason <msg>]
goclaw teams tasks assign <team-id> <task-id> --user <uid>
goclaw teams tasks comment <team-id> <task-id> --text "..."
goclaw teams tasks comments <team-id> <task-id>
goclaw teams tasks events <team-id> <task-id>
goclaw teams events <team-id>
goclaw teams scopes <team-id>
goclaw teams known-users <team-id>
```

WS Methods:
```
teams.tasks.approve, teams.tasks.reject, teams.tasks.assign
teams.tasks.comment, teams.tasks.comments, teams.tasks.events
teams.events.list, teams.scopes, teams.known_users
```

### 4. `channels` — Add writers management

New subcommands:
```
goclaw channels writers list <instance-id>
goclaw channels writers add <instance-id> --user-id <uid>
goclaw channels writers remove <instance-id> <user-id>
goclaw channels writers groups <instance-id>
```

HTTP Endpoints:
```
GET    /v1/channels/instances/{id}/writers
POST   /v1/channels/instances/{id}/writers
DELETE /v1/channels/instances/{id}/writers/{userId}
GET    /v1/channels/instances/{id}/writers/groups
```

### 5. `providers` — Add embedding & Claude CLI auth status

New subcommands:
```
goclaw providers embedding-status
goclaw providers claude-auth-status
```

HTTP: `GET /v1/embedding/status`, `GET /v1/providers/claude-cli/auth-status`

### 6. `storage` — Add download flag + move

Update existing + new:
```
goclaw storage download <path>   # GET with ?download=true
goclaw storage move --from <src> --to <dst>  # PUT /v1/storage/move
```

### 7. `config` — Add permissions subcommands

New subcommands:
```
goclaw config permissions list
goclaw config permissions grant --user-id <uid> --key <key>
goclaw config permissions revoke --user-id <uid> --key <key>
```

WS Methods: `config.permissions.list`, `config.permissions.grant`, `config.permissions.revoke`

## Related Code Files

### Files to Modify
- `cmd/skills.go` — Add versions, files, tenant-config, deps subcommands
- `cmd/tools.go` — Add builtin tenant-config subcommands
- `cmd/teams.go` — Add task approve/reject/assign/comment/events, team events/scopes
- `cmd/channels.go` — Add writers subcommands
- `cmd/providers.go` — Add embedding-status, claude-auth-status
- `cmd/storage.go` — Add download flag, move operation
- `cmd/config.go` — Add permissions subcommands

## Implementation Steps

1. Update `cmd/skills.go` — add 7 new subcommands
2. Update `cmd/tools.go` — add 2 tenant-config subcommands
3. Update `cmd/teams.go` — add 9 new subcommands (consider splitting if >200 lines)
4. Update `cmd/channels.go` — add 4 writers subcommands
5. Update `cmd/providers.go` — add 2 status subcommands
6. Update `cmd/storage.go` — add download + move
7. Update `cmd/config.go` — add 3 permissions subcommands
8. `go build ./...` after each file

## Todo List

- [ ] **FIRST:** Split teams.go → teams.go + teams_tasks.go + teams_events.go
- [ ] **FIRST:** Split agents.go → agents.go + agents_instances.go + agents_links.go
- [ ] **FIRST:** Split admin.go → admin.go + admin_users.go + admin_audit.go
- [ ] **FIRST:** Remove `tools custom` commands (dead code)
- [ ] Enhance skills with versions/files/tenant-config
- [ ] Enhance tools builtin with tenant-config
- [ ] Enhance teams with task workflow + events
- [ ] Enhance channels with writers
- [ ] Enhance providers with embedding/claude status
- [ ] Enhance storage with download/move
- [ ] Enhance config with permissions
- [ ] Compile check all

## Success Criteria

- All new subcommands appear in `goclaw <cmd> --help`
- Tenant-config toggle works for skills and tools
- Teams task approve/reject workflow functional
- Storage download saves file to disk

## Risk Assessment

- **Medium:** `teams.go` already 500 lines; adding 9 subcommands will exceed 200-line limit → needs modularization into `cmd/teams_tasks.go` and `cmd/teams_events.go`
- **Low:** All follow established patterns
