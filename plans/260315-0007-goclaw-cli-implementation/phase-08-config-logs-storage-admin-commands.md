---
phase: 8
title: Config, Logs, Storage & Admin Commands
status: completed
priority: medium
effort: M
depends_on: [phase-02]
---

# Phase 8 — Config, Logs, Storage & Admin Commands

## Overview
Server configuration management, real-time log tailing, workspace file browser, CLI credentials, TTS, and media.

## Requirements

### Server Config
```
goclaw config get [--key <path>]
goclaw config apply --file <config.json>
goclaw config patch --key <path> --value <value>
goclaw config schema
```

### Logs (WebSocket streaming)
```
goclaw logs tail [--agent <id>] [--level info|warn|error] [--follow]
```

### Storage
```
goclaw storage list [--path <subdir>]
goclaw storage get <path> [--output <file>]
goclaw storage delete <path> [--yes]
goclaw storage size
```

### CLI Credentials
```
goclaw credentials list
goclaw credentials create --name <n>
goclaw credentials delete <id> [--yes]
```

### TTS (Text-to-Speech)
```
goclaw tts status
goclaw tts enable
goclaw tts disable
goclaw tts providers
goclaw tts set-provider --name <provider>
goclaw tts convert --text <text> [--output <file>]
```

### Activity/Audit
```
goclaw activity list [--limit N] [--type <action>]
```

### Media
```
goclaw media upload <file>
goclaw media get <mediaID> [--output <file>]
```

## Implementation Steps

1. `cmd/config.go` — Config get/apply/patch/schema
2. `cmd/logs.go` — WebSocket log tailing with color-coded levels
3. `cmd/storage.go` — Workspace file browser
4. `cmd/credentials.go` — CLI credential management
5. `cmd/tts.go` — TTS operations
6. `cmd/activity.go` — Audit log viewer
7. `cmd/media.go` — Media upload/download
8. Log tailing: connect WS, call `logs.tail`, render with level colors
9. Config apply: read JSON file, POST to server
10. Storage get: download file, write to local path or stdout

## Related Code Files
- Create: `cmd/config.go`, `cmd/logs.go`, `cmd/storage.go`
- Create: `cmd/credentials.go`, `cmd/tts.go`, `cmd/activity.go`, `cmd/media.go`

## Todo
- [x] Config get/apply/patch/schema
- [x] Real-time log tailing via WebSocket
- [x] Storage file operations
- [x] CLI credential CRUD
- [x] TTS operations
- [x] Activity/audit log viewer
- [x] Media upload/download with progress

## Success Criteria
- `goclaw config get` shows full server config
- `goclaw config patch --key "rate_limit" --value 100` updates config
- `goclaw logs tail --agent mybot --follow` streams logs in real-time
- `goclaw storage list` shows workspace files with sizes
- `goclaw credentials create --name "ci-token"` returns new credential
- `goclaw tts convert --text "Hello" --output hello.mp3` generates audio
