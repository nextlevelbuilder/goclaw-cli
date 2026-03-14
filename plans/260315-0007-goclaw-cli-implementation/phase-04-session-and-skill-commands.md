---
phase: 4
title: Session & Skill Commands
status: planned
priority: high
effort: M
depends_on: [phase-02]
---

# Phase 4 — Session & Skill Commands

## Overview
Session management and skill system commands.

## Requirements

### Sessions (REST API)
```
goclaw sessions list [--agent <id>] [--user <userID>] [--limit N]
goclaw sessions preview <sessionKey>
goclaw sessions delete <sessionKey> [--yes]
goclaw sessions reset <sessionKey>
goclaw sessions label <sessionKey> --label "new label"
```

### Skills (REST API)
```
goclaw skills list [--search <query>]
goclaw skills get <id>
goclaw skills upload <path> [--name <name>] [--visibility private|shared]
goclaw skills update <id> [flags]
goclaw skills delete <id> [--yes]
goclaw skills toggle <id>
goclaw skills grant <id> --agent <agentID> [--version <v>]
goclaw skills grant <id> --user <userID>
goclaw skills revoke <id> --agent <agentID>
goclaw skills revoke <id> --user <userID>
goclaw skills versions <id>
goclaw skills runtimes
goclaw skills files <id> [--path <subpath>]
goclaw skills rescan-deps
goclaw skills install-deps
```

## Implementation Steps

1. `cmd/sessions.go` — Session list/preview/delete/reset/label
2. `cmd/skills.go` — Full skill lifecycle
3. Skill upload: multipart form upload with progress bar
4. Session preview: render messages with role labels + timestamps
5. Skill search: pass query param for BM25/semantic search

## Related Code Files
- Create: `cmd/sessions.go`, `cmd/skills.go`

## Todo
- [ ] Session list with agent/user filters
- [ ] Session preview with message rendering
- [ ] Session delete/reset/label
- [ ] Skill CRUD commands
- [ ] Skill upload with progress
- [ ] Skill grant/revoke for agents and users
- [ ] Skill toggle enable/disable
- [ ] Skill version listing
- [ ] Runtime listing
- [ ] File browser for skill content
- [ ] Dependency management (rescan/install)

## Success Criteria
- `goclaw sessions list` shows sessions with token counts
- `goclaw sessions preview <key>` renders conversation
- `goclaw skills upload ./my-skill/` uploads with progress bar
- `goclaw skills grant <id> --agent <agentID>` works
- `goclaw skills list --search "web scraping"` returns semantic results
