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
| exec / run shell / remote command / approvals | [references/exec-workflow.md](references/exec-workflow.md) |
| login / token / profile / tenant switch / credentials / api-keys | [references/auth-and-config.md](references/auth-and-config.md) |
| agent list/get/create/delete, agent files, instances, wake | [references/agents-core.md](references/agents-core.md) |
| agent share, link, delegate, regenerate | [references/agents-advanced.md](references/agents-advanced.md) |
| chat with agent, session list/preview/delete | [references/chat-sessions.md](references/chat-sessions.md) |
| health, status, logs, traces, usage, metrics | [references/monitoring-ops.md](references/monitoring-ops.md) |
| knowledge graph, entity dedup, memory | [references/knowledge-memory.md](references/knowledge-memory.md) |
| teams, members, team tasks, workspace files | [references/teams-collaboration.md](references/teams-collaboration.md) |
| channels, contacts, pending messages, writers | [references/channels-messaging.md](references/channels-messaging.md) |
| export / import / storage (workspace files) | [references/data-movement.md](references/data-movement.md) |
| providers, skills, built-in tools list/config, packages | [references/providers-skills-tools.md](references/providers-skills-tools.md) |
| cron, heartbeat, device pairing | [references/automation-scheduling.md](references/automation-scheduling.md) |
| MCP servers, grants, requests | [references/mcp-integration.md](references/mcp-integration.md) |
| tenants, system-config, audit activity, TTS | [references/admin-system.md](references/admin-system.md) |
| media upload/download | [references/media.md](references/media.md) |
| API documentation browsing | [references/docs-api.md](references/docs-api.md) |

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
