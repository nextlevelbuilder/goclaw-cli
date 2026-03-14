# GoClaw CLI

A production-ready CLI for managing [GoClaw](https://github.com/nextlevelbuilder/goclaw) AI agent gateway servers.

## Features

- **Full API coverage** — Every dashboard feature accessible via CLI
- **Dual mode** — Interactive (humans) + Automation (AI agents / CI)
- **Security-first** — OS keyring credential storage, TLS by default, no secrets in `ps`
- **Multiple output formats** — Table, JSON, YAML
- **Streaming** — Real-time chat, log tailing via WebSocket
- **Multi-profile** — Manage multiple server connections

## Installation

### From Source

```bash
go install github.com/nextlevelbuilder/goclaw-cli@latest
```

### From Release

Download the latest binary from [Releases](https://github.com/nextlevelbuilder/goclaw-cli/releases).

## Quick Start

```bash
# Login with token
goclaw auth login --server https://goclaw.example.com --token your-token

# Or use device pairing
goclaw auth login --server https://goclaw.example.com --pair

# Check server health
goclaw health

# List agents
goclaw agents list

# Chat with an agent
goclaw chat myagent

# Single-shot message (automation)
goclaw chat myagent -m "What is the status?" -o json

# Pipe input
echo "Analyze this log" | goclaw chat myagent
```

## Commands

| Command | Description |
|---------|-------------|
| `auth` | Login, logout, device pairing, profile management |
| `agents` | CRUD, shares, delegation links, per-user instances |
| `chat` | Interactive or single-shot messaging with streaming |
| `sessions` | List, preview, delete, reset, label |
| `skills` | Upload, manage, grant/revoke access |
| `mcp` | MCP server management, grants, access requests |
| `providers` | LLM provider CRUD, model listing, verification |
| `tools` | Custom + built-in tool management, invocation |
| `cron` | Scheduled jobs CRUD, trigger, run history |
| `teams` | Team management, task board, workspace |
| `channels` | Channel instances, contacts, pending messages |
| `traces` | LLM trace viewer, export |
| `memory` | Memory documents, semantic search |
| `knowledge-graph` | Entity extraction, linking, querying |
| `usage` | Usage analytics and cost breakdown |
| `config` | Server configuration get/apply/patch |
| `logs` | Real-time log streaming |
| `storage` | Workspace file browser |
| `approvals` | Execution approval management |
| `delegations` | Delegation history |
| `credentials` | CLI credential store |
| `tts` | Text-to-speech operations |
| `media` | Media upload/download |
| `activity` | Audit log |

## Automation Mode

All commands support automation via flags:

```bash
# JSON output
goclaw agents list -o json

# Skip confirmations
goclaw agents delete abc123 -y

# Environment variables
export GOCLAW_SERVER=https://goclaw.example.com
export GOCLAW_TOKEN=your-token
goclaw agents list
```

## Configuration

Config stored in `~/.goclaw/config.yaml`:

```yaml
active_profile: production
profiles:
  - name: production
    server: https://goclaw.example.com
    token: your-token
  - name: staging
    server: https://staging.goclaw.example.com
    token: staging-token
```

Switch profiles:

```bash
goclaw auth use-context staging
```

## Development

```bash
make build    # Build binary
make test     # Run tests
make lint     # Run go vet
make install  # Install to GOPATH/bin
```

## License

MIT
