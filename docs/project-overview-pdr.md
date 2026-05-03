# GoClaw CLI - Product Development Requirements

## Project Overview

GoClaw CLI is a production-ready command-line interface for managing GoClaw AI agent gateway servers. Built with Cobra framework and Go, it provides full API coverage for the GoClaw dashboard accessible through both interactive (human) and automation (AI agent/CI) modes.

**Repository:** https://github.com/nextlevelbuilder/goclaw-cli
**Status:** Production Ready (Phases 1-9 + P0-P4 Complete)
**Last Updated:** 2026-04-15

---

## Product Vision

Enable developers and AI agents to seamlessly manage GoClaw servers through a unified CLI, supporting enterprise workflows with security-first principles and multiple output formats.

---

## Core Requirements

### Functional Requirements

| Requirement | Status | Details |
|-------------|--------|---------|
| **28 Command Groups** | Complete | Auth, agents, chat, sessions, skills, MCP, providers, tools, cron, teams, channels, traces, memory, knowledge-graph, usage, config, logs, storage, approvals, delegations, credentials, TTS, media, activity |
| **Full API Coverage** | Complete | Every dashboard feature accessible via CLI |
| **Dual Mode** | Complete | Interactive (human-friendly TUI) + Automation (flags/env vars) |
| **Multiple Output Formats** | Complete | Table (human), JSON (machines), YAML (configuration) |
| **Real-time Streaming** | Complete | WebSocket support for chat, logs, and event streaming |
| **Multi-profile Support** | Complete | Manage multiple server connections with profiles |
| **Configuration Management** | Complete | File (~/.goclaw/config.yaml), env vars, CLI flags |
| **Profile Switching** | Complete | `goclaw auth use-context <profile>` |

### Non-Functional Requirements

| Requirement | Status | Details |
|-------------|--------|---------|
| **Security** | Complete | OS keyring integration, TLS by default, no secrets in `ps` output |
| **Performance** | Complete | <2s response time for HTTP requests, instant WebSocket connection |
| **Compatibility** | Complete | macOS, Linux, Windows (amd64, arm64) via GoReleaser |
| **Error Handling** | Complete | Wrapped errors with context, user-friendly messages |
| **Code Quality** | In Progress | Table-driven tests, comprehensive coverage |
| **Documentation** | In Progress | API reference, user guides, examples |

---

## Functional Specifications

### Authentication & Profiles

```
goclaw auth login --server <url> [--token <token>|--pair]
goclaw auth logout [--profile <name>]
goclaw auth use-context <profile-name>
goclaw credentials get <key>
goclaw credentials set <key> <value>
```

- Credentials stored in OS keyring (not disk)
- Config file stores server URL and profile metadata only
- Default profile from config file

### Agent Management (CRUD)

```
goclaw agents list [-o json|yaml]
goclaw agents get <id>
goclaw agents create --name <name> --provider <provider> --model <model>
goclaw agents update <id> --field <value>
goclaw agents delete <id> [-y]
goclaw agents share <id> --user <email>
goclaw agents delegation-link <id>
```

### Chat Operations

```
goclaw chat <agent-id|agent-key>           # Interactive mode
goclaw chat <agent-id> -m "message"        # Single-shot (automation)
echo "input" | goclaw chat <agent-id>      # Pipe support
goclaw chat <agent-id> --file file.txt     # File input
goclaw chat <agent-id> -o json             # JSON output
```

### Session Management

```
goclaw sessions list
goclaw sessions get <session-id>
goclaw sessions delete <session-id> [-y]
goclaw sessions reset <session-id> [-y]
goclaw sessions label <session-id> --label "name"
```

### Streaming Operations

```
goclaw logs [-f]                           # Real-time log tailing
goclaw chat <agent-id> [interactive]       # Streaming chat
goclaw traces <trace-id> [--stream]        # Trace streaming
```

### Configuration Management

```
goclaw config get [--field <path>]
goclaw config apply -f config.yaml
goclaw config patch --field value
goclaw status                               # Server health check
```

---

## Design Constraints

### Architecture
- **HTTP Client:** REST API for CRUD operations (GET, POST, PUT, PATCH, DELETE)
- **WebSocket Client:** Streaming and bidirectional communication
- **No ORM:** Raw HTTP calls with JSON marshaling
- **No External CLI Dependencies:** Single binary (cobra, viper, gorilla/websocket only)

### Configuration Hierarchy
1. **CLI Flags** (highest priority) — `goclaw agents list -o json`
2. **Environment Variables** — `GOCLAW_SERVER`, `GOCLAW_TOKEN`, `GOCLAW_OUTPUT`
3. **Config File** — `~/.goclaw/config.yaml`
4. **Defaults** — Built-in defaults

### Command Structure
- Root command: `goclaw`
- Subcommands: `goclaw <command> <subcommand> [flags] [args]`
- Global flags: `--server`, `--token`, `--output`, `--yes`, `--verbose`, `--insecure`, `--profile`
- Command-specific flags defined in each command group

### Output Formats
- **table:** Human-readable tables with column alignment
- **json:** Compact JSON for programmatic consumption
- **yaml:** YAML for configuration files and readability

---

## Technical Stack

### Core Dependencies
- **Cobra:** Command-line framework
- **Gorilla WebSocket:** WebSocket client
- **golang.org/x/term:** Terminal utilities (TUI, raw mode)
- **gopkg.in/yaml.v3:** YAML parsing and serialization

### Build & Release
- **Go 1.25.3+:** Minimum version
- **Make:** Build automation
- **GoReleaser:** Multi-platform binary distribution
- **GitHub Actions:** CI/CD (lint, test, build, release)

### Testing
- **Table-driven tests:** Parameterized test cases
- **Race detector:** `go test -race ./...`
- **Linting:** `go vet ./...`

---

## Acceptance Criteria

### Phase 1-9 Complete
- All 28 command groups implemented
- Full API coverage verified
- Dual mode (interactive + automation) working
- Multi-profile support functional
- WebSocket streaming operational
- Security hardening completed

### Phase 10+ (Future)
- Unit test coverage >80%
- Integration tests for critical paths
- Shell completion scripts (bash, zsh, fish)
- Homebrew tap for easy installation

---

## Success Metrics

| Metric | Target | Current |
|--------|--------|---------|
| Command Coverage | 100% (28/28 groups) | 28/28 ✓ |
| Build Time | <5s | <2s ✓ |
| Binary Size | <15MB | ~8MB ✓ |
| Test Coverage | >80% | In Progress |
| Documentation Coverage | 100% | In Progress |
| Automation Support | All commands | ✓ |

---

## Deployment & Release

### Build
```bash
make build              # Local binary
make install            # Install to GOPATH/bin
go install ./...        # Latest from main
```

### Release (Tagged Versions)
```bash
git tag v1.0.0
git push origin v1.0.0  # Triggers GitHub Actions release
```

### Artifacts
- **goclaw_X.X.X_darwin_amd64.tar.gz** — macOS Intel
- **goclaw_X.X.X_darwin_arm64.tar.gz** — macOS Apple Silicon
- **goclaw_X.X.X_linux_amd64.tar.gz** — Linux Intel
- **goclaw_X.X.X_linux_arm64.tar.gz** — Linux ARM
- **goclaw_X.X.X_windows_amd64.zip** — Windows Intel
- **goclaw_X.X.X_windows_arm64.zip** — Windows ARM

---

## Security Considerations

- **Credentials:** Stored in OS keyring, never in config file
- **TLS:** Required by default, `--insecure` only for testing
- **Token Exposure:** Commands accept `--token` flag but prefer `GOCLAW_TOKEN` env var
- **No Logging:** Sensitive data (tokens, credentials) never logged
- **Process Security:** Credentials not visible in `ps` output

---

## Configuration Example

File: `~/.goclaw/config.yaml`

```yaml
active_profile: production
profiles:
  - name: production
    server: https://goclaw.example.com
    token: {stored-in-keyring}
  - name: staging
    server: https://staging.goclaw.example.com
    token: {stored-in-keyring}
  - name: local
    server: http://localhost:8080
    token: {stored-in-keyring}
```

Environment variables take precedence:
```bash
export GOCLAW_SERVER=https://custom.example.com
export GOCLAW_TOKEN=custom-token
export GOCLAW_OUTPUT=json
goclaw agents list  # Uses custom server/token/output
```

---

## Command Inventory

### Auth & Profiles (6 commands)
`auth`, `credentials`

### Agent Management (~30 subcommands)
`agents` core: list, get, create, update, delete, share, unshare, regenerate, resummon, cancel-summon, prompt-preview
`agents files` (list, get, set) — global context files (AGENTS.md, SOUL.md, IDENTITY.md, ...)
`agents instances` (list, get-file, set-file, metadata, update-metadata)
`agents episodic` (list, search), `agents evolution` (metrics, suggestions, update)
`agents links` (list/create/update/delete), `agents skills list`
`agents v3-flags` (get, toggle), `agents wake/wait/identity`
`agents orchestration`, `agents codex-pool-activity`, `agents export/import/import-merge`

### Chat & Messaging (1 command)
`chat`

### Session Management (5 commands)
`sessions` (list, get, delete, reset, label)

### Skill Management (4 commands)
`skills` (list, upload, delete, grant-access, revoke-access)

### MCP Server Management (5 commands)
`mcp` (list, add, remove, grants, access-requests)

### LLM Providers (4 commands)
`providers` (list, create, update, delete, verify, models)

### Custom Tools (3 commands)
`tools` (list, invoke, delete)

### Scheduled Jobs (4 commands)
`cron` (list, create, update, delete, trigger, history)

### Team Management (5 commands)
`teams` (list, create, members, task-board, events, scopes, export, import)
`teams workspace` (list, read, delete, upload, move)

### Channels (3 commands)
`channels` (list, contacts, pending-messages)

### LLM Traces (2 commands)
`traces` (list, export)

### Memory Documents (3 commands)
`memory` (list, search, upsert)

### Knowledge Graph (2 commands)
`knowledge-graph` (entities, links, query)

### Usage Analytics (5 subcommands)
`usage` (summary, detail, costs, timeseries, breakdown)

### Hooks (7 subcommands) — event interception
`hooks` (list, create, update, delete, toggle, test, history)

### Files & Voices
`files sign` — signed URL helper
`voices` (list, refresh) — voice catalog

### Server Config (3 commands)
`config` (get, apply, patch)

### Logs (1 command)
`logs`

### Workspace Storage (2 commands)
`storage` (list, download)

### Approvals (2 commands)
`approvals` (list, approve, deny)

### Delegations (1 command)
`delegations`

### Text-to-Speech (6 subcommands)
`tts` (status, enable, disable, providers, set-provider, test-connection)

### Media (2 commands)
`media` (upload, download)

### Activity & Audit (1 command)
`activity`

### Utility (2 commands)
`version`, `status`

**Total: ~40 command groups** (after AI-first expansion through P6 — CLI parity with server admin surface)

---

## Version History

| Version | Date | Status | Notes |
|---------|------|--------|-------|
| v1.1.0 | 2026-05-02 | Unreleased | P6 Domain Coverage Expansion (hooks, agents files, usage analytics, voices, etc.) |
| v1.0.0 | 2026-03-15 | Production | All phases complete (1-9) |
| v0.x | Earlier | Dev | Feature development |

---

## Open Questions

None at this time. Full specification complete and implemented.
