---
phase: 6
title: Team, Channel & Cron Commands
status: completed
priority: high
effort: M
depends_on: [phase-02]
---

# Phase 6 — Team, Channel & Cron Commands

## Overview
Team management (members, tasks, workspace), channel instance management, and cron job scheduling.

## Requirements

### Teams
```
goclaw teams list
goclaw teams get <id>
goclaw teams create --name <n> --agents <id1,id2,...>
goclaw teams update <id> [flags]
goclaw teams delete <id> [--yes]
goclaw teams members add <teamID> --agent <agentID> [--role lead|member]
goclaw teams members remove <teamID> --agent <agentID>
goclaw teams members list <teamID>
goclaw teams tasks list <teamID> [--status open|assigned|approved|rejected]
goclaw teams tasks get <teamID> <taskID>
goclaw teams tasks create <teamID> --title <t> --description <d> [--assignee <agentID>]
goclaw teams tasks assign <teamID> <taskID> --agent <agentID>
goclaw teams tasks approve <teamID> <taskID>
goclaw teams tasks reject <teamID> <taskID> [--reason <text>]
goclaw teams tasks comment <teamID> <taskID> --body <text>
goclaw teams tasks comments <teamID> <taskID>
goclaw teams workspace list <teamID>
goclaw teams workspace read <teamID> <path>
goclaw teams workspace delete <teamID> <path>
```

### Channels
```
goclaw channels instances list [--type telegram|discord|slack|...]
goclaw channels instances get <id>
goclaw channels instances create --type <type> --agent <agentID> --name <n> [flags]
goclaw channels instances update <id> [flags]
goclaw channels instances delete <id> [--yes]
goclaw channels contacts list [--channel <type>]
goclaw channels contacts resolve <id1,id2,...>
goclaw channels pending list <channelID>
goclaw channels pending retry <channelID> <messageID>
goclaw channels writers list <instanceID>
goclaw channels writers add <instanceID> --user <userID> [--display-name <n>]
goclaw channels writers remove <instanceID> --user <userID>
```

### Cron Jobs
```
goclaw cron list [--agent <id>]
goclaw cron get <id>
goclaw cron create --agent <agentID> --name <n> --schedule <cron-expr|interval> [flags]
goclaw cron update <id> [flags]
goclaw cron delete <id> [--yes]
goclaw cron toggle <id>
goclaw cron run <id>
goclaw cron status <id>
goclaw cron runs <id> [--limit N]
```

## Implementation Steps

1. `cmd/teams.go` — Full team lifecycle + members + tasks + workspace
2. `cmd/channels.go` — Channel instances + contacts + pending + writers
3. `cmd/cron.go` — Cron job CRUD + trigger + run history
4. Channel create: type-specific credential prompts (Telegram bot token, Discord bot token, etc.)
5. Cron create: validate cron expression locally before sending
6. Task comments: render with timestamps and agent attribution

## Related Code Files
- Create: `cmd/teams.go`, `cmd/channels.go`, `cmd/cron.go`

## Todo
- [x] Team CRUD + member management
- [x] Team task board (list, create, assign, approve, reject, comment)
- [x] Team workspace file operations
- [x] Channel instance CRUD with type-specific config
- [x] Contact management (list, resolve)
- [x] Pending message retry
- [x] Group file writers management
- [x] Cron job CRUD
- [x] Cron schedule validation
- [x] Manual cron trigger
- [x] Cron run history display

## Success Criteria
- `goclaw teams create --name "Research" --agents agent1,agent2` creates team
- `goclaw teams tasks list <teamID>` shows task board
- `goclaw channels instances create --type telegram --agent mybot` works
- `goclaw cron create --agent mybot --schedule "0 */6 * * *" --name "Report"` schedules job
- `goclaw cron runs <id>` shows history with duration and status
