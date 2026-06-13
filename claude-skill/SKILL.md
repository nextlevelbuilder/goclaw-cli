---
name: goclaw
description: Manage GoClaw AI agent gateway servers from Claude Code. Use this skill when the user mentions "goclaw", "gateway server", "AI agent platform", or wants to execute shell commands remotely on a server, manage AI agents, approve agent actions, inspect chat sessions, or administer multi-tenant AI infrastructure. Wraps the `goclaw` CLI to call REST + WebSocket APIs of GoClaw Gateway.
when_to_use: goclaw CLI operations, remote shell execution via exec tool, AI agent lifecycle, chat session inspection, tenant administration, skill/provider/MCP management, execution approvals
allowed-tools: Bash(goclaw:*)
disable-model-invocation: false
user-invocable: true
argument-hint: <natural-language-intent>
---

# GoClaw CLI Skill

Lets Claude invoke the `goclaw` binary via Bash to interact with a GoClaw Gateway server. Single source of truth = the CLI; this skill teaches Claude the ergonomics.

## Conventions (always apply)

1. **Always append `--output json`** to every goclaw call. Rely on JSON parsing, not table output.
2. **Read auth from `~/.goclaw/config.yaml`** — do NOT accept token via prompt; user must have run `goclaw auth login` beforehand.
3. **Destructive ops require user confirm.** Any command with `--yes` flag or verbs `delete`, `reset`, `revoke`, `unpublish`, `clear`, `rotate`, `deny`, `execute-merge` → ask user to confirm *before* adding `--yes`.
4. **Streaming commands are NOT Bash-friendly.** Refuse to run: `chat` interactive mode, `logs tail`, `auth pair`, `approvals watch`. Suggest polling or user-run-manually alternative.
5. **Exit code 1 after auth error** → suggest `goclaw auth login`, do not retry-loop.
6. **Never hardcode** server URL, tenant ID, agent ID, user ID — use placeholders in examples, actual values from user context.

## Hero use case — execute shell on server

Tool `exec` registered as builtin on GoClaw Gateway. Claude invokes via:

```bash
goclaw tools invoke exec --param command="ls -la /workspace" --output json
# with working directory:
goclaw tools invoke exec --param command="npm test" --param working_dir="/workspace/app" --output json
```

Approval flow when server deems command sensitive (package installs, deny-patterns): call hits `approvals` queue. Handle via `references/exec-workflow.md`.

## Navigation — load relevant reference on demand

Claude: when user's intent matches, `Read` the listed reference file before constructing commands.

| Intent signal | Reference |
|---------------|-----------|
| exec / run shell / remote command / approvals / tools builtin/custom | [references/exec-workflow.md](references/exec-workflow.md) |
| login / token / api-keys / tenant switch | [references/auth-and-config.md](references/auth-and-config.md) |
| CLI profiles (server config), connections | [references/profile.md](references/profile.md) |
| credential store, grants, presets | [references/credentials.md](references/credentials.md) |
| agents CRUD (create, list, update, delete, files, instances, wake) | [references/agents-crud.md](references/agents-crud.md) |
| agents sharing, delegation links, history, wait, regenerate, resummon | [references/agents-sharing-delegation.md](references/agents-sharing-delegation.md) |
| agents export, import, merge archives | [references/agents-export-import.md](references/agents-export-import.md) |
| agents episodic memory, evolution metrics, suggestions, skill evolution | [references/agents-memory.md](references/agents-memory.md) |
| agents orchestration, identity, skills, prompt-preview, lifecycle (summon, cancel, sync, v3-flags) | [references/agents-orchestration.md](references/agents-orchestration.md) |
| chat single-shot, continue session, connectivity test | [references/chat-basic.md](references/chat-basic.md) |
| chat session operations (list, preview, delete, reset, label, branch, compact) | [references/chat-sessions.md](references/chat-sessions.md) |
| chat message history, replay | [references/chat-history.md](references/chat-history.md) |
| chat message injection, session status, follow | [references/chat-injection.md](references/chat-injection.md) |
| health, status, logs (tail), traces, usage, metrics | [references/monitoring-ops.md](references/monitoring-ops.md) |
| knowledge graph (dedup, entities, stats), agent episodic memory | [references/knowledge-memory.md](references/knowledge-memory.md) |
| knowledge vault (documents, search, graph, enrichment, upload) | [references/vault.md](references/vault.md) |
| teams, members, team tasks, attachments, events, scopes, workspace | [references/teams-collaboration.md](references/teams-collaboration.md) |
| channels, contacts, instances, pending, tenant-users, writers | [references/channels-messaging.md](references/channels-messaging.md) |
| inter-agent messaging (send), pending message management | [references/channels-messaging.md](references/channels-messaging.md) |
| export / import agents/teams/skills/mcp (portable archives) | [references/data-movement.md](references/data-movement.md) |
| storage (list, get, upload, delete, move, size), files (sign) | [references/data-movement.md](references/data-movement.md) |
| providers (create/get/list), embedding status, skills, tools, packages | [references/providers-skills-tools.md](references/providers-skills-tools.md) |
| LLM provider management, OAuth (ChatGPT, OpenAI) | [references/oauth.md](references/oauth.md) |
| cron jobs (create, schedule, run, history) | [references/automation-scheduling.md](references/automation-scheduling.md) |
| heartbeat (agent health monitoring), device pairing, workstations | [references/automation-scheduling.md](references/automation-scheduling.md) |
| MCP servers, grants, requests, integration | [references/mcp-integration.md](references/mcp-integration.md) |
| tenants (CRUD, users), audit activity (activity log), TTS status | [references/admin-system.md](references/admin-system.md) |
| per-tenant system-config KV store | [references/admin-system.md](references/admin-system.md) |
| server-level config schema/defaults/get/patch/apply, system upgrade | [references/system-config.md](references/system-config.md) |
| media upload/download, quota | [references/media.md](references/media.md) |
| API documentation browsing | [references/docs-api.md](references/docs-api.md) |
| backup/restore (system/tenant), preflight, S3 integration | [references/backup-restore.md](references/backup-restore.md) |
| inbound webhooks (HTTP callbacks to external services) | [references/webhooks.md](references/webhooks.md) |
| event hooks (internal system routing, handlers, matchers) | [references/hooks.md](references/hooks.md) |
| voice catalog management, TTS voices | [references/voices.md](references/voices.md) |
| coding-agent workstations (isolated environments, permissions, activity) | [references/workstations.md](references/workstations.md) |
| server edition info, feature availability, license | [references/edition.md](references/edition.md) |
| quota inspection (token/API usage by agent) | [references/quota.md](references/quota.md) |
| user search and discovery (cross-tenant) | [references/users.md](references/users.md) |

## Compatibility

Tested against `goclaw` CLI ≥ 0.3.0. Run `goclaw version` to check. Schema drift caught by `check-drift.sh` in CI.

## Prerequisites

- `goclaw` binary in PATH (download from https://github.com/nextlevelbuilder/goclaw-cli/releases)
- Authenticated: `goclaw auth login` or paired via `goclaw auth pair`
- Permissions granted in `~/.claude/settings.json` (see install.sh)

## Troubleshooting

- Command returns 401 → token expired, run `goclaw auth login`
- Command returns 403 → user lacks role for resource; check `goclaw whoami --output json`
- Command hangs → likely streaming op — Claude should refuse, see convention #4
