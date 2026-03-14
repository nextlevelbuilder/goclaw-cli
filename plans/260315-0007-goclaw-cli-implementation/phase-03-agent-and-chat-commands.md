---
phase: 3
title: Agent & Chat Commands
status: completed
priority: critical
effort: L
depends_on: [phase-02]
---

# Phase 3 — Agent & Chat Commands

## Overview
Agent CRUD, agent links, agent instances, and interactive/single-shot chat with streaming support.

## Key Insights
- Agents are the central entity — most features revolve around agents
- Chat uses WebSocket streaming (events: `chunk`, `tool.call`, `tool.result`, `run.started`, `run.completed`)
- Agent types: `open` (per-user context) vs `predefined` (shared context)
- Agent links enable delegation between agents

## Requirements

### Agent Commands (REST API)
```
goclaw agents list [--status active|inactive]
goclaw agents get <id|key>
goclaw agents create --name <name> --provider <provider> --model <model> [flags]
goclaw agents update <id> [flags]
goclaw agents delete <id> [--yes]
goclaw agents share <id> --user <userID> [--role admin|operator|viewer]
goclaw agents unshare <id> --user <userID>
goclaw agents regenerate <id>
goclaw agents resummon <id>
```

### Agent Links (Delegation)
```
goclaw agents links list [--agent <id>]
goclaw agents links create --source <id> --target <id> --direction outbound|inbound|bidirectional
goclaw agents links update <linkID> [flags]
goclaw agents links delete <linkID>
```

### Agent Instances (Per-User)
```
goclaw agents instances list <agentID>
goclaw agents instances get-file <agentID> --user <userID> --file <name>
goclaw agents instances set-file <agentID> --user <userID> --file <name> --content <content|@file>
goclaw agents instances metadata <agentID> --user <userID> [--patch <json>]
```

### Chat (WebSocket)
```
# Interactive mode (TUI chat interface)
goclaw chat <agent>

# Single-shot (automation mode)
goclaw chat <agent> -m "message" [--session <key>]

# Pipe stdin
echo "Analyze this" | goclaw chat <agent>

# With options
goclaw chat <agent> -m "message" --model claude-sonnet-4-6 --no-stream
```

## Architecture

### Chat Flow (Interactive)
1. Resolve agent (by key or ID)
2. Connect WebSocket, authenticate
3. Enter REPL loop:
   - Prompt user for input
   - Send `chat.send` with streaming
   - Render streaming chunks in real-time
   - Show tool calls with spinner
   - Show tool results
   - Handle `run.completed`
4. Support `/` commands: `/sessions`, `/abort`, `/clear`, `/exit`

### Chat Flow (Automation)
1. Connect WebSocket, authenticate
2. Send `chat.send` with message from `-m` flag or stdin
3. Collect streaming events as NDJSON (if `--output json`) or plain text
4. Exit with status code 0 on success, 1 on error

### Streaming Output
- **Interactive:** Render token-by-token with ANSI colors. Tool calls in dimmed box. Thinking blocks collapsible.
- **Automation JSON:** Each event as NDJSON line:
  ```json
  {"event":"chunk","data":{"content":"Hello"}}
  {"event":"tool.call","data":{"name":"read_file","id":"tc_1"}}
  {"event":"tool.result","data":{"id":"tc_1","result":"..."}}
  {"event":"run.completed","data":{"input_tokens":150,"output_tokens":300}}
  ```

## Implementation Steps

1. `cmd/agents.go` — Cobra command group with all CRUD subcommands
2. `cmd/agents_links.go` — Agent links subcommands
3. `cmd/agents_instances.go` — Instance file management
4. `cmd/chat.go` — Chat command with dual-mode support
5. `internal/tui/chat.go` — Interactive chat TUI (readline + streaming renderer)
6. Agent create flags: `--name`, `--provider`, `--model`, `--context-window`, `--workspace`, `--type`, `--budget`
7. Agent update: accept JSON patch or individual flags
8. Content from file: `--content @path/to/file` reads file content

## Related Code Files
- Create: `cmd/agents.go`, `cmd/agents_links.go`, `cmd/agents_instances.go`
- Create: `cmd/chat.go`
- Create: `internal/tui/chat.go`

## Todo
- [x] Agent CRUD commands (list, get, create, update, delete)
- [x] Agent share/unshare commands
- [x] Agent regenerate/resummon commands
- [x] Agent links CRUD
- [x] Agent instance file management
- [x] Interactive chat TUI with streaming
- [x] Single-shot chat for automation
- [x] Stdin pipe support
- [x] NDJSON streaming for automation mode
- [x] Chat session management (--session flag)
- [x] In-chat slash commands (/abort, /sessions, /clear)

## Success Criteria
- `goclaw agents list` shows all agents in table
- `goclaw agents create` creates agent with all config options
- `goclaw chat myagent` opens interactive streaming chat
- `goclaw chat myagent -m "hello" -o json` returns NDJSON stream
- `echo "query" | goclaw chat myagent` works in pipelines
- Tool calls displayed during chat streaming

## Risk Assessment
- Chat streaming requires robust WebSocket handling (reconnect on drop)
- Terminal rendering of markdown/code blocks in streaming — use lipgloss/glamour
- Large tool results may need truncation in interactive display
