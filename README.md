# GoClaw CLI

A production-ready CLI for managing [GoClaw](https://github.com/nextlevelbuilder/goclaw) AI agent gateway servers.

## Features

- **Broad operational API coverage** — Dashboard and super-admin gateway workflows accessible via CLI
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
| `profile` | List, create, switch, inspect, and delete CLI profiles |
| `agents` | CRUD, shares, delegation links, per-user instances, evolution |
| `chat` | Interactive or single-shot messaging with streaming |
| `sessions` | List, preview, delete, reset, label, compact |
| `codex-pool` | Unified Codex pool activity lookup for agents/providers |
| `skills` | Upload, manage, grant/revoke access |
| `mcp` | MCP server management, grants, access requests |
| `providers` | LLM provider CRUD, model listing, verification |
| `tools` | Custom + built-in tool management, invocation |
| `cron` | Scheduled jobs CRUD, trigger, run history |
| `teams` | Team management, task board, workspace, attachments |
| `channels` | Channel instances, contacts, pending messages |
| `traces` | LLM trace viewer, filters, export |
| `memory` | Memory documents, semantic search |
| `knowledge-graph` | Entity extraction, linking, querying |
| `usage` | Usage analytics and cost breakdown |
| `tenants` | Tenant CRUD, user membership management |
| `heartbeat` | Agent heartbeat monitoring, checklist, logs |
| `system-configs` | System-level key-value configuration |
| `edition` | Show server edition info (no auth required) |
| `config` | Server configuration get/apply/patch/permissions |
| `logs` | Real-time log streaming |
| `storage` | Workspace file browser |
| `approvals` | Execution approval management |
| `delegations` | Delegation history |
| `credentials` | CLI credential store |
| `tts` | Text-to-speech operations |
| `media` | Media upload/download |
| `activity` | Audit log |
| `api-keys` | API key management (create, list, revoke, rotate) |
| `system upgrade` | Gateway release upgrade status and trigger controls |
| `workstations` | Coding-agent workstation CRUD, permissions, activity, and agent links |
| `webhooks` | Webhook admin CRUD, secret rotation, and deletion |
| `api-docs` | API documentation (Swagger UI, OpenAPI spec) |
| `backup` | System/tenant backup, signed download, S3 integration |
| `restore` | System/tenant restore from backup archive |
| `vault` | Knowledge Vault — documents, links, search, graph, enrichment |

### Backend-Unblocked Surfaces (P6)

Seven one-shot subcommands wired to backend PRs `#37` and `#44`:

```bash
# Incremental trace polling (one shot; rerun with returned cursor)
goclaw traces follow --session-key <key> [--since <RFC3339>] [--limit <n>]
goclaw traces follow --agent <id> [--since <RFC3339>] [--limit <n>]

# Provider hot-reconnect (bumps registry without recreating credentials)
goclaw providers reconnect <provider-id>

# Branch a chat session at a message index
goclaw sessions branch <session-key> --up-to-index <N> [--new-session-key <k>] \
  [--label <l>] [--metadata k=v ...]

# One-shot session-history poll (cursor-based; not a stream)
goclaw sessions follow <session-key> [--cursor <n>] [--limit <n>]

# Probe a (group, user) pair against a channel's writer policy
goclaw channels writers test <instance-id> --group-id <g> --user-id <u>

# Aggregate audit-log activity by dimension
goclaw activity aggregate --group-by <action|actor_type|entity_type|actor_id> \
  [--from <RFC3339>] [--to <RFC3339>] [--limit <n>] \
  [--actor-type <v>] [--actor-id <v>] [--action <v>] [--entity-type <v>] [--entity-id <v>]

# Summarize the runtime log ring buffer (NOT a stream — see 'logs tail' for that)
goclaw logs aggregate [--group-by <level|source>] [--level <l>] [--source <s>] [--from <RFC3339>]
```

All are one-shot HTTP — no watch loops or WS streams. `logs aggregate` is admin-only on the server; `activity aggregate --group-by actor_id` is also admin-only (server-enforced).

### Reading a Trace by ID

```bash
# Human-readable: header + span tree + events
goclaw traces get <trace-id>

# Machine-readable JSON (also auto-selected when stdout is piped)
goclaw traces get <trace-id> -o json
```

Exit codes for `traces get`: `0` on success, `2` on permission denied, `3` on not-found, `4` on malformed id (rejected before any HTTP call — allowlist `^[A-Za-z0-9._-]+$`), `5` on upstream server failure, `6` on rate-limit / network-resource exhaustion.

## Backup & Restore

### System Backup

```bash
# Check readiness (disk space, pg_dump availability)
goclaw backup system-preflight

# Create backup (returns signed download token)
goclaw backup system

# Create backup and download immediately
goclaw backup system --wait --file backup.tar.gz

# Download an existing backup by token
goclaw backup system-download <token> --file backup.tar.gz
```

### Tenant Backup

```bash
goclaw backup tenant --tenant-id=<id>
goclaw backup tenant --tenant-id=<id> --wait --file tenant-backup.tar.gz
goclaw backup tenant-download <token> --file tenant-backup.tar.gz
```

### S3 Integration

```bash
# Configure S3 destination
goclaw backup s3 config set --bucket my-bucket --access-key AKID --secret-key secret

# List backups in S3
goclaw backup s3 list

# One-shot: create and upload to S3
goclaw backup s3 backup
```

### Restore

> **CAUTION: Restore is a DESTRUCTIVE operation. It overwrites all existing data.**
> All active connections must be stopped before restoring.
> This operation is logged server-side for audit purposes.

Restore requires **both** `--yes` and a typed confirmation to prevent accidental execution:

```bash
# System restore — must type the exact filename
goclaw restore system backup-20240101.tar.gz --yes --confirm=backup-20240101.tar.gz

# Tenant restore — must type the exact tenant ID
goclaw restore tenant tenant-backup.tar.gz --tenant-id=abc123 --yes --confirm=abc123
```

If `--yes` is omitted or `--confirm` does not match, the command **refuses** immediately
without making any server request.

### Export / Import (Per-Domain)

```bash
# Export a single agent to file
goclaw agents export agent-123 --file agent-123.tar.gz

# Import agent (preview by default — no changes made)
goclaw agents import agent-123.tar.gz

# Import agent (apply — actually creates agent)
goclaw agents import agent-123.tar.gz --apply

# Merge into existing agent
goclaw agents import-merge agent-123 updates.tar.gz --apply

# Teams
goclaw teams export team-123 --file team-123.tar.gz
goclaw teams import team-123.tar.gz --apply

# Skills
goclaw skills export --file skills.tar.gz
goclaw skills import skills.tar.gz --apply

# MCP servers
goclaw mcp export --file mcp.tar.gz
goclaw mcp import mcp.tar.gz --apply
```

## Knowledge Vault

The Knowledge Vault is GoClaw's RAG (Retrieval-Augmented Generation) document store. It supports semantic + full-text search, document relationship graphs, and background enrichment.

### Search (RAG)

```bash
# Semantic search — returns top matches with relevance scores
goclaw vault search "authentication flow" | jq '.[].title'

# Limit results
goclaw vault search "API design" --limit=5

# Pipe to jq for AI agent use
goclaw vault search "database schema" -o json | jq '.[0].id'
```

### Document Management

```bash
# List all documents
goclaw vault documents list

# Filter by query
goclaw vault documents list --q=kubernetes

# Get a specific document
goclaw vault documents get <docID>

# Create a document (register path in vault)
goclaw vault documents create --title="API Guide" --path=docs/api-guide.md

# Update metadata
goclaw vault documents update <docID> --title="Updated API Guide"

# Delete (requires --yes)
goclaw vault documents delete <docID> --yes

# Show document links (outlinks + backlinks)
goclaw vault documents links <docID>
```

### File Upload

```bash
# Upload a file to the vault (streaming multipart, no RAM buffering)
goclaw vault upload ./docs/architecture.md

# Upload with title and tags
goclaw vault upload ./report.md --title="Q1 Report" --tags=finance,quarterly
```

### Document Links (Knowledge Graph Edges)

```bash
# Create a link between two documents
goclaw vault links create --from=<docID> --to=<docID> --type=reference

# Delete a link (requires --yes)
goclaw vault links delete <linkID> --yes

# Batch-fetch links for multiple documents
goclaw vault links batch-get doc-1 doc-2 doc-3
```

### Tree View

```bash
# Browse vault directory tree (TTY: ASCII tree, pipe: JSON)
goclaw vault tree

# Filter by path prefix
goclaw vault tree --path=agents/

# Pipe to jq for programmatic access
goclaw vault tree -o json | jq '.entries[].name'
```

### Graph Visualization

```bash
# Get full vault graph as JSON
goclaw vault graph

# Export as Graphviz DOT (pipe to dot for PNG)
goclaw vault graph --format=dot

# Generate a PNG visualization
goclaw vault graph --format=dot | dot -Tpng -o vault-graph.png
```

### Enrichment Pipeline

```bash
# Check enrichment status (background AI processing)
goclaw vault enrichment status

# Stop enrichment (admin, requires --yes)
goclaw vault enrichment stop --yes
```

### Workspace Rescan (Admin)

```bash
# Rescan workspace for new/changed files (requires --yes)
goclaw vault rescan --yes
```

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

# Rotate a key by creating a replacement and revoking the old key
goclaw api-keys rotate <key-id> --name "ci-deploy-v2" --scopes "operator.read,operator.write" --yes
```

Available scopes: `operator.admin`, `operator.read`, `operator.write`, `operator.approvals`, `operator.pairing`

## UX Convenience Commands

```bash
# Unified Codex pool activity
goclaw codex-pool activity --agent=agent-123
goclaw codex-pool activity --provider=provider-123

# Resolved server defaults
goclaw config defaults -o json

# Replay or resume a known chat session
goclaw chat replay myagent --session=sess-123 -o json
goclaw chat sessions resume myagent --session=sess-123 -m "Continue" --no-stream

# Invoke a custom tool with JSON args from file
goclaw tools invoke weather --args=@payload.json

# Download a team task attachment to an explicit file
goclaw teams attachments download team-123 attachment-456 --output ./artifact.bin

# Approve a skill_add evolution suggestion, optionally overriding the draft
goclaw agents evolution skill apply agent-123 suggestion-456 --skill-draft @./SKILL.md
```

## API Docs

```bash
# Open Swagger UI in browser
goclaw api-docs open

# Fetch OpenAPI 3.0 spec as JSON
goclaw api-docs spec -o json
```

## Output Format Behavior

GoClaw CLI auto-detects the appropriate output format:

| Condition | Default Format |
|-----------|---------------|
| stdout is a terminal (TTY) | `table` — human-readable aligned columns |
| stdout is piped / redirected | `json` — machine-readable, one object per response |
| active profile has an output default | configured profile default — unless env/flag overrides it |
| `GOCLAW_OUTPUT=yaml` env set | `yaml` — regardless of TTY |
| `--output` / `-o` flag set | exact value — overrides everything |

**Precedence (highest → lowest):** `--output` flag > `GOCLAW_OUTPUT` env > active profile output default > TTY detection

```bash
# Explicit format (always wins)
goclaw agents list -o json
goclaw agents list -o yaml
goclaw agents list -o table

# Auto-detect: piped → JSON
goclaw agents list | jq '.[0].id'

# Env override
GOCLAW_OUTPUT=yaml goclaw agents list
```

### Exit Codes

| Code | Meaning | Triggers |
|------|---------|---------|
| 0 | Success | Normal completion |
| 1 | Generic error | Unknown/unmapped |
| 2 | Auth failure | `UNAUTHORIZED`, `NOT_PAIRED`, `TENANT_ACCESS_REVOKED`, HTTP 401/403 |
| 3 | Not found | `NOT_FOUND`, `NOT_LINKED`, HTTP 404 |
| 4 | Validation | `INVALID_REQUEST`, `FAILED_PRECONDITION`, `ALREADY_EXISTS`, HTTP 400/409/422 |
| 5 | Server error | `INTERNAL`, `UNAVAILABLE`, `AGENT_TIMEOUT`, HTTP 5xx |
| 6 | Resource/network | `RESOURCE_EXHAUSTED`, HTTP 429, connection timeout |

Errors in JSON mode are emitted as `{"error":{"code":"...","message":"..."}}` to stdout for machine parsing.

## Automation Mode

All commands support automation via flags:

```bash
# JSON output
goclaw agents list -o json

# Skip confirmations
goclaw agents delete abc123 -y

# Suppress banners/tips
goclaw agents list --quiet

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
goclaw profile list
goclaw profile use staging
goclaw auth use-context staging
```

One-shot profile override:

```bash
goclaw --profile staging agents list
GOCLAW_PROFILE=staging goclaw traces list --since=1h --root-only -o json
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
