---
phase: 4
status: complete
priority: medium
effort: S
---

# Phase 4: New WebSocket & Streaming Features

## Overview

Add remaining WS-only features: heartbeat monitoring and enhanced chat capabilities.

## New Features

### 1. Heartbeat Monitoring

```
goclaw heartbeat get
goclaw heartbeat set --interval <seconds> --url <endpoint>
goclaw heartbeat toggle --enabled <bool>
goclaw heartbeat test
goclaw heartbeat logs [--limit 20]
goclaw heartbeat checklist get
goclaw heartbeat checklist set --data @checklist.json
goclaw heartbeat targets
```

WS Methods:
```
heartbeat.get, heartbeat.set, heartbeat.toggle, heartbeat.test
heartbeat.logs, heartbeat.checklist.get, heartbeat.checklist.set
heartbeat.targets
```

### 2. Chat Enhancements

Add to existing `chat` command:
```
goclaw chat inject <agent> --text "..." --session <sid>  # Inject mid-turn
goclaw chat status <agent> --session <sid>               # Session/run status
goclaw chat abort <agent> --session <sid>                # Cancel running agent
```

WS Methods: `chat.inject`, `chat.session.status`, `chat.abort`

### 3. Agent Wait

```
goclaw agents wait <agent-id> --session <sid> [--timeout 60]
```

WS Method: `agent.wait` — Block until agent completes, with timeout.

## Related Code Files

### Files to Create
- `cmd/heartbeat.go`

### Files to Modify
- `cmd/chat.go` — Add inject, status, abort subcommands
- `cmd/agents.go` — Add wait subcommand

## Implementation Steps

1. Create `cmd/heartbeat.go` with 8 subcommands using `newWS()` + `Call()`
2. Add `inject`, `status`, `abort` subcommands to `cmd/chat.go`
3. Add `wait` subcommand to `cmd/agents.go`
4. `go build ./...`

## Todo List

- [ ] Implement heartbeat commands
- [ ] Add chat inject/status/abort
- [ ] Add agents wait
- [ ] Compile check

## Success Criteria

- `goclaw heartbeat get` returns monitoring config
- `goclaw chat abort` cancels running agent
- `goclaw agents wait` blocks until completion or timeout

## Risk Assessment

- **Low:** All WS `Call()` pattern, no streaming needed
- **Low:** Chat abort already has WS method support in client layer
