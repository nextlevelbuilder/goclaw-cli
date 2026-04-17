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
| `agents` | CRUD, shares, delegation links, per-user instances, wait |
| `chat` | Interactive/single-shot messaging, inject, status, abort |
| `sessions` | List, preview, delete, reset, label |
| `skills` | Upload, manage, grant/revoke, versions, files, tenant-config, deps, runtimes |
| `mcp` | MCP server management, grants, access requests |
| `providers` | LLM provider CRUD, model listing, verification, embedding status |
| `tools` | Builtin tool management, tenant-config |
| `cron` | Scheduled jobs CRUD, trigger, run history |
| `teams` | Team management, task board, task approval, workspace, events |
| `channels` | Channel instances, contacts, pending messages, writers |
| `traces` | LLM trace viewer, export |
| `memory` | Memory documents, semantic search |
| `knowledge-graph` | Entity extraction, linking, querying, traversal |
| `usage` | Usage analytics, cost breakdown, timeseries |
| `config` | Server configuration get/apply/patch, permissions |
| `logs` | Real-time log streaming |
| `storage` | Workspace file browser, download, move |
| `approvals` | Execution approval management |
| `delegations` | Delegation history |
| `credentials` | CLI credential store, presets, testing |
| `tts` | Text-to-speech operations, convert |
| `media` | Media upload/download |
| `activity` | Audit log |
| `api-keys` | API key management (create, list, revoke) |
| `api-docs` | API documentation (Swagger UI, OpenAPI spec) |
| `tenants` | Tenant CRUD, user management (admin) |
| `system-config` | Per-tenant key-value configuration |
| `packages` | Package management, runtimes |
| `contacts` | Contact resolution, merge/unmerge |
| `pending-messages` | Pending message management |
| `heartbeat` | Health monitoring, checklist, targets |

## API Keys

Create scoped, revocable API keys for CI/CD and integrations:

```bash
# Create a key with read+write scopes
goclaw api-keys create --name "ci-deploy" --scopes "operator.read,operator.write"

# Create a key with 30-day expiry
goclaw api-keys create --name "temp-access" --scopes "operator.read" --expires-in 2592000

# List all keys (raw key is only shown at creation)
goclaw api-keys list

# Revoke a key
goclaw api-keys revoke <key-id>
```

Available scopes: `operator.admin`, `operator.read`, `operator.write`, `operator.approvals`, `operator.pairing`

## API Docs

```bash
# Open Swagger UI in browser
goclaw api-docs open

# Fetch OpenAPI 3.0 spec as JSON
goclaw api-docs spec -o json
```

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

## Multi-Tenant

All commands support tenant context via the `--tenant-id` flag:

```bash
# Set tenant context for all operations
goclaw agents list --tenant-id my-tenant

# Or via environment variable
export GOCLAW_TENANT_ID=my-tenant
goclaw agents list

# Manage tenants (admin only)
goclaw tenants list
goclaw tenants create --name "My Tenant"
goclaw tenants users list <tenant-id>
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

Environment variables:

| Variable | Description |
|----------|-------------|
| `GOCLAW_SERVER` | Server URL |
| `GOCLAW_TOKEN` | Auth token or API key |
| `GOCLAW_TENANT_ID` | Tenant ID for multi-tenant operations |

Switch profiles:

```bash
goclaw auth use-context staging
```

## Claude Code Skill

A Claude Code skill wrapping this CLI lives in [`claude-skill/`](./claude-skill/).
It lets Claude Code autonomously invoke `goclaw` to manage your GoClaw server —
list agents, run `exec`-style shell commands on the server, inspect traces, etc.

Install (once the binary is in `PATH`):

```bash
RELEASE_URL="https://github.com/nextlevelbuilder/goclaw-cli/releases/download/skill-v0.1.0"
curl -fsSL "$RELEASE_URL/goclaw-skill.tar.gz" -o /tmp/goclaw-skill.tar.gz
curl -fsSL "$RELEASE_URL/goclaw-skill.sha256" -o /tmp/goclaw-skill.sha256
(cd /tmp && shasum -a 256 -c goclaw-skill.sha256)
tar xzf /tmp/goclaw-skill.tar.gz -C /tmp
/tmp/claude-skill/install.sh
```

See [`claude-skill/README.md`](./claude-skill/README.md) for permission modes,
example prompts, and uninstall instructions.

## Development

```bash
make build    # Build binary
make test     # Run tests
make lint     # Run go vet
make install  # Install to GOPATH/bin
```

## License

MIT
