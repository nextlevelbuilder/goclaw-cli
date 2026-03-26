---
phase: 2
status: complete
priority: high
effort: L
---

# Phase 2: Missing Resource Commands

## Overview

Implement all commands listed in README that are not yet built, plus new server features that need CLI commands.

## Commands to Implement

### 1. `knowledge-graph` (8 HTTP endpoints)

```
goclaw knowledge-graph entities list <agent-id>
goclaw knowledge-graph entities get <agent-id> <entity-id>
goclaw knowledge-graph entities create <agent-id> --data @file.json
goclaw knowledge-graph entities delete <agent-id> <entity-id>
goclaw knowledge-graph extract <agent-id> --text "..."
goclaw knowledge-graph traverse <agent-id> --from <entity-id>
goclaw knowledge-graph graph <agent-id>
goclaw knowledge-graph stats <agent-id>
```

HTTP Endpoints:
```
GET    /v1/agents/{agentID}/kg/entities
GET    /v1/agents/{agentID}/kg/entities/{entityID}
POST   /v1/agents/{agentID}/kg/entities
DELETE /v1/agents/{agentID}/kg/entities/{entityID}
POST   /v1/agents/{agentID}/kg/extract
POST   /v1/agents/{agentID}/kg/traverse
GET    /v1/agents/{agentID}/kg/graph
GET    /v1/agents/{agentID}/kg/stats
```

### 2. `usage` (4 HTTP endpoints)

```
goclaw usage summary [--agent-id <id>] [--from <date>] [--to <date>]
goclaw usage breakdown [--agent-id <id>] [--group-by model|agent|day]
goclaw usage timeseries [--agent-id <id>] [--interval hour|day|week]
goclaw usage costs [--agent-id <id>]
```

HTTP Endpoints:
```
GET /v1/usage/summary
GET /v1/usage/breakdown
GET /v1/usage/timeseries
GET /v1/costs/summary
```

### 3. `activity` (1 HTTP endpoint)

```
goclaw activity list [--agent-id <id>] [--action <type>] [--limit 50]
```

HTTP: `GET /v1/activity`

### 4. `credentials` (6 HTTP endpoints)

```
goclaw credentials list
goclaw credentials get <id>
goclaw credentials create --name <name> --type <type> --data @file.json
goclaw credentials test <id>
goclaw credentials update <id> --name <name>
goclaw credentials delete <id>
goclaw credentials presets
```

HTTP Endpoints:
```
GET    /v1/cli-credentials
GET    /v1/cli-credentials/{id}
POST   /v1/cli-credentials
POST   /v1/cli-credentials/{id}/test
PUT    /v1/cli-credentials/{id}
DELETE /v1/cli-credentials/{id}
GET    /v1/cli-credentials/presets
```

### 5. `media` (2 HTTP endpoints)

```
goclaw media upload <file-path> [--agent-id <id>]
goclaw media get <media-id>
```

HTTP: `POST /v1/media/upload`, `GET /v1/media/{id}`

### 6. `packages` (4 HTTP endpoints)

```
goclaw packages list
goclaw packages runtimes
goclaw packages install --name <pkg> [--runtime <rt>]
goclaw packages uninstall --name <pkg>
```

HTTP Endpoints:
```
GET  /v1/packages
GET  /v1/packages/runtimes
POST /v1/packages/install
POST /v1/packages/uninstall
```

### 7. `tts` (6 WS methods)

```
goclaw tts status
goclaw tts enable
goclaw tts disable
goclaw tts convert --text "..." [--provider <name>]
goclaw tts providers
goclaw tts set-provider <name>
```

WS Methods: `tts.status`, `tts.enable`, `tts.disable`, `tts.convert`, `tts.providers`, `tts.setProvider`

### 8. `contacts` (5 HTTP endpoints)

```
goclaw contacts list
goclaw contacts resolve --identifier <phone/email>
goclaw contacts merge <contact-id-1> <contact-id-2>
goclaw contacts unmerge <tenant-user-id>
goclaw contacts merged <tenant-user-id>
```

HTTP Endpoints:
```
GET  /v1/contacts
GET  /v1/contacts/resolve
POST /v1/contacts/merge
POST /v1/contacts/unmerge
GET  /v1/contacts/merged/{tenantUserId}
```

### 9. `pending-messages` (3 HTTP endpoints)

```
goclaw pending-messages list
goclaw pending-messages compact
goclaw pending-messages delete
```

HTTP: `GET /v1/pending-messages`, `POST /v1/pending-messages/compact`, `DELETE /v1/pending-messages`

## Related Code Files

### Files to Create
- `cmd/knowledge_graph.go`
- `cmd/usage.go`
- `cmd/activity.go`
- `cmd/credentials.go`
- `cmd/media.go`
- `cmd/packages.go`
- `cmd/tts.go`
- `cmd/contacts.go`
- `cmd/pending_messages.go`

### Files to Modify
- `cmd/root.go` — Register all 9 new command groups

## Implementation Steps

1. Create each command file following existing Cobra patterns
2. Wire HTTP endpoints using `newHTTP()` + `buildBody()`
3. Wire WS methods using `newWS()` + `Call()`
4. Add table columns for each list command
5. Register in `root.go`
6. `go build ./...` after each command group

## Todo List

- [ ] Implement knowledge-graph commands
- [ ] Implement usage commands
- [ ] Implement activity command
- [ ] Implement credentials commands
- [ ] Implement media commands
- [ ] Implement packages commands
- [ ] Implement tts commands
- [ ] Implement contacts commands
- [ ] Implement pending-messages commands
- [ ] Register all in root.go
- [ ] Compile check

## Success Criteria

- All 9 command groups compile and show help text
- List commands return proper table/JSON output
- Upload commands handle multipart properly (media, skills)

## Risk Assessment

- **Low:** All follow established CRUD patterns
- **Medium:** `tts.convert` may need streaming support for long audio
- **Medium:** `media upload` needs multipart file handling (already pattern in skills)
