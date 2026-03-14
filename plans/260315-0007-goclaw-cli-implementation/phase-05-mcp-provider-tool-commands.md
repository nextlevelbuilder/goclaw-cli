---
phase: 5
title: MCP, Provider & Tool Commands
status: completed
priority: high
effort: M
depends_on: [phase-02]
---

# Phase 5 — MCP, Provider & Tool Commands

## Overview
MCP server management, LLM provider CRUD, custom/builtin tool management, and tool invocation.

## Requirements

### MCP Servers
```
goclaw mcp servers list
goclaw mcp servers get <id>
goclaw mcp servers create --name <n> --transport stdio|sse|streamable-http [flags]
goclaw mcp servers update <id> [flags]
goclaw mcp servers delete <id> [--yes]
goclaw mcp servers test <id>
goclaw mcp servers tools <id>
```

### MCP Grants
```
goclaw mcp grants list --agent <agentID>
goclaw mcp grants grant --server <id> --agent <agentID> [--allow <tools>] [--deny <tools>]
goclaw mcp grants grant --server <id> --user <userID>
goclaw mcp grants revoke --server <id> --agent <agentID>
goclaw mcp grants revoke --server <id> --user <userID>
```

### MCP Access Requests
```
goclaw mcp requests list [--status pending|approved|rejected]
goclaw mcp requests create --server <id> [--reason <text>]
goclaw mcp requests review <id> --action approve|reject
```

### LLM Providers
```
goclaw providers list
goclaw providers get <id>
goclaw providers create --name <n> --type openai_compat --api-base <url> --api-key <key>
goclaw providers update <id> [flags]
goclaw providers delete <id> [--yes]
goclaw providers models <id>
goclaw providers verify <id>
```

### Custom Tools
```
goclaw tools custom list [--agent <id>]
goclaw tools custom get <id>
goclaw tools custom create --name <n> --command <cmd> --description <d> [flags]
goclaw tools custom update <id> [flags]
goclaw tools custom delete <id> [--yes]
```

### Built-in Tools
```
goclaw tools builtin list
goclaw tools builtin get <name>
goclaw tools builtin update <name> [--enabled true|false]
```

### Tool Invocation
```
goclaw tools invoke <name> [--param key=value]...
goclaw tools invoke <name> --params '{"key":"value"}'
```

## Implementation Steps

1. `cmd/mcp.go` — MCP servers + grants + requests subcommands
2. `cmd/providers.go` — LLM provider CRUD + model listing + verify
3. `cmd/tools.go` — Custom + builtin + invoke subcommands
4. Provider create: prompt for API key via masked input (not flag)
5. MCP create: different flags per transport type
6. Tool invoke: parse `key=value` pairs or JSON params

## Related Code Files
- Create: `cmd/mcp.go`, `cmd/providers.go`, `cmd/tools.go`

## Todo
- [x] MCP server CRUD + test + tools listing
- [x] MCP grants management (agent + user level)
- [x] MCP access request workflow
- [x] Provider CRUD with encrypted key input
- [x] Provider model listing and verification
- [x] Custom tool CRUD
- [x] Built-in tool listing and settings
- [x] Tool invocation with parameter parsing

## Success Criteria
- `goclaw mcp servers create --transport stdio --command npx --args ...` registers server
- `goclaw mcp servers test <id>` shows connection status
- `goclaw providers verify <id>` confirms API key works
- `goclaw tools invoke dns_lookup --param domain=example.com` executes tool
- API keys prompted securely (not in `--api-key` flag in history)

## Security Considerations
- API keys: prompt via `Password()` in interactive, `GOCLAW_PROVIDER_API_KEY` env in automation
- MCP server env vars: encrypted before sending to server
