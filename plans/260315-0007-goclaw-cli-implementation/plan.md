---
title: GoClaw CLI Implementation
status: planned
created: 2026-03-15
priority: high
blockedBy: []
blocks: []
---

# GoClaw CLI — Implementation Plan

A production-ready CLI for managing GoClaw servers. Supports interactive (human) and automation (AI agent/CI) modes. Security-first. Covers ALL dashboard features.

## Tech Stack

- **Language:** Go 1.25 (single binary, matches server)
- **CLI Framework:** Cobra + Viper (config/env management)
- **HTTP Client:** net/http (REST API)
- **WebSocket:** gorilla/websocket (streaming chat, logs, events)
- **Output:** Table (human), JSON/YAML (automation)
- **Auth:** Bearer token + device pairing flow
- **Config:** `~/.goclaw/config.yaml` + env vars + flags
- **Repo:** Private on `nextlevelbuilder` org

## Architecture

```
goclaw-cli/
├── cmd/                    # Cobra commands (1 file per resource)
│   ├── root.go             # Global flags, config loading
│   ├── auth.go             # login, logout, pairing, whoami
│   ├── agents.go           # CRUD, shares, links, instances
│   ├── chat.go             # Interactive chat + send message
│   ├── sessions.go         # List, preview, delete, reset
│   ├── skills.go           # CRUD, upload, grants, toggle
│   ├── mcp.go              # Servers, grants, requests
│   ├── providers.go        # LLM provider management
│   ├── tools.go            # Custom + builtin tools
│   ├── cron.go             # Jobs, runs, toggle, trigger
│   ├── config.go           # Server config get/apply/patch
│   ├── teams.go            # Teams, members, tasks
│   ├── channels.go         # Channel instances, contacts
│   ├── traces.go           # Trace viewer, export
│   ├── memory.go           # Memory documents, search
│   ├── logs.go             # Log tailing (WebSocket)
│   ├── usage.go            # Usage stats, cost summary
│   ├── storage.go          # Workspace file browser
│   ├── approvals.go        # Execution approvals
│   ├── delegations.go      # Delegation history
│   ├── credentials.go      # CLI credential store
│   ├── status.go           # Health, server info
│   └── version.go          # CLI version
├── internal/
│   ├── client/             # HTTP + WebSocket client
│   │   ├── http.go         # REST API client
│   │   ├── websocket.go    # WS RPC client
│   │   └── auth.go         # Token management
│   ├── config/             # Config loading (~/.goclaw/)
│   ├── output/             # Table, JSON, YAML formatters
│   └── tui/                # Interactive prompts (survey/huh)
├── go.mod
├── go.sum
├── main.go
├── Makefile
├── .goreleaser.yaml        # Cross-platform releases
└── README.md
```

## Phases

| # | Phase | Status | Effort | Priority |
|---|-------|--------|--------|----------|
| 1 | [Project Bootstrap](phase-01-project-bootstrap.md) | planned | S | critical |
| 2 | [Core Client & Auth](phase-02-core-client-and-auth.md) | planned | M | critical |
| 3 | [Agent & Chat Commands](phase-03-agent-and-chat-commands.md) | planned | L | critical |
| 4 | [Session & Skill Commands](phase-04-session-and-skill-commands.md) | planned | M | high |
| 5 | [MCP, Provider & Tool Commands](phase-05-mcp-provider-tool-commands.md) | planned | M | high |
| 6 | [Team, Channel & Cron Commands](phase-06-team-channel-cron-commands.md) | planned | M | high |
| 7 | [Trace, Memory, Usage & Utility Commands](phase-07-trace-memory-usage-utility-commands.md) | planned | M | medium |
| 8 | [Config, Logs, Storage & Admin Commands](phase-08-config-logs-storage-admin-commands.md) | planned | M | medium |
| 9 | [Testing, CI/CD & Release](phase-09-testing-cicd-release.md) | planned | M | high |

## Dual Mode Strategy

| Aspect | Interactive (Human) | Automation (Agent/CI) |
|--------|--------------------|-----------------------|
| Output | Colored tables, spinners | JSON/YAML (`--output json`) |
| Auth | `goclaw login` (interactive prompt) | `--token` flag or `GOCLAW_TOKEN` env |
| Errors | Human-readable messages | Structured error JSON with codes |
| Prompts | Confirmation dialogs | `--yes` flag to skip prompts |
| Streaming | Live terminal output | Newline-delimited JSON events |
| Config | Interactive wizard | Flags + env vars only |

## Security

- Token stored encrypted in `~/.goclaw/credentials` (OS keyring preferred, file fallback)
- No secrets in CLI flags visible in `ps` output — use env vars or stdin
- TLS verification on by default (`--insecure` to disable)
- HMAC signature for MCP bridge calls
- Credential rotation via `goclaw auth rotate`

## Key Dependencies

- `github.com/spf13/cobra` — CLI framework
- `github.com/spf13/viper` — Config management
- `github.com/gorilla/websocket` — WebSocket client
- `github.com/charmbracelet/huh` — Interactive forms
- `github.com/charmbracelet/lipgloss` — Terminal styling
- `github.com/olekukonez/tablewriter` — Table output
- `github.com/zalando/go-keyring` — OS keyring

## Command Overview (Complete Feature Parity)

```
goclaw auth login|logout|whoami|pair|rotate
goclaw status|health|version
goclaw agents list|get|create|update|delete|share|unshare|regenerate|resummon
goclaw agents links list|create|update|delete
goclaw agents instances list|get-file|set-file|metadata
goclaw chat [agent] [message]           # Interactive or single-shot
goclaw sessions list|preview|delete|reset|label
goclaw skills list|get|create|update|delete|upload|toggle|grant|revoke|versions|runtimes
goclaw mcp servers list|get|create|update|delete|test|tools
goclaw mcp grants list|grant|revoke
goclaw mcp requests list|create|review
goclaw providers list|get|create|update|delete|models|verify
goclaw tools custom list|get|create|update|delete
goclaw tools builtin list|get|update
goclaw tools invoke [name] [params]
goclaw cron list|get|create|update|delete|toggle|run|status|runs
goclaw teams list|get|create|update|delete|members|tasks|workspace
goclaw channels instances list|get|create|update|delete
goclaw channels contacts list|resolve
goclaw channels pending list|retry
goclaw channels writers list|add|remove
goclaw traces list|get|export
goclaw memory list|get|store|delete|search
goclaw knowledge-graph query|extract|link
goclaw usage summary|detail
goclaw config get|apply|patch|schema
goclaw logs tail [--agent] [--level]
goclaw storage list|get|delete|size
goclaw approvals list|approve|deny
goclaw delegations list|get
goclaw credentials list|create|delete
goclaw tts status|enable|disable|convert|providers
```
