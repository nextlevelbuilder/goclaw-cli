# Agents — core lifecycle

## When to use

User wants to list/inspect/create/update/delete agents, manage agent context files, or control per-user agent instances.

## Commands in scope

- `goclaw agents list` / `get <id>` / `create` / `update <id>` / `delete <id>` — base CRUD (source: `cmd/agents.go`)
- `goclaw agents files list/get/create/delete` — agent context files (source: `agents_files.go`, WS-backed)
- `goclaw agents instances list/get/create/delete/trigger/reset` — per-user instances (source: `agents_instances.go`)
- `goclaw agents wake <id>` — wake sleeping agent (source: `agents_wake.go`)

## Verified flags

### `agents create` / `agents update`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--name <n>` | string | Display name |
| `--provider <p>` | string | LLM provider |
| `--model <m>` | string | Model identifier |
| `--type <t>` | string | `open` or `predefined` (default `open`) |
| `--context-window <n>` | int | Context window size |
| `--workspace <path>` | string | Workspace dir |
| `--budget <cents>` | int | Monthly budget in cents |

### `agents files create`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--data <json-or-@file>` | string | File metadata/content |

### `agents instances list`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--user-id <id>` | string | Filter by user |

## JSON output

- ✅ `agents list/get` — JSON (cmd/agents.go:30, :57)
- ✅ `agents instances list/get` — JSON
- ⚠️ `agents create/update/delete/wake` — `printer.Success` text; parse exit code + stdout ID message
- ⚠️ `agents files create/delete` — success text only
- ⚠️ `agents instances create/delete/trigger/reset` — success text

## Destructive ops

| Command | Source | Confirm |
| --- | --- | --- |
| `agents delete` | `agents.go:140` tui.Confirm | YES |
| `agents files delete` | `agents_files.go` tui.Confirm | YES |
| `agents instances delete` | `agents_instances.go` tui.Confirm | YES |
| `agents instances reset` | clears instance state | YES |

## Common patterns

### Example 1: list + get details
```bash
goclaw agents list --output json
goclaw agents get <agent-id> --output json
```

### Example 2: create agent
```bash
goclaw agents create \
  --name "Support Bot" \
  --provider anthropic \
  --model claude-sonnet-4-6 \
  --type open \
  --context-window 200000 \
  --workspace /workspaces/support \
  --budget 5000 \
  --output json
```

### Example 3: update model + budget
```bash
goclaw agents update <agent-id> --model claude-opus-4-7 --budget 10000 --output json
```

### Example 4: full lifecycle — create → upload file → instance → wake
```bash
goclaw agents create --name "DocBot" --provider openai --model gpt-4o --type open --output json
# → AGENT_ID

goclaw agents files create --data '{"path":"guide.md","content":"..."}' --output json

goclaw agents instances create --user-id user-42 --output json
goclaw agents wake <agent-id>
```

### Example 5: delete agent (always confirm)
```bash
# Claude: echo back "Delete agent <id>? This is permanent." before running:
goclaw agents delete <agent-id> --yes
```

## Edge cases & gotchas

- **`--name` sets `display_name`** on create, but update maps flag `--name` to `display_name` field (source: `agents.go:107`). Both work.
- **Agent vs instance:** agent = template, instance = per-user runtime. Delete agent ⇒ cascades instances. List carefully.
- **Files WS-backed:** `agents files` subcommands use WebSocket, not HTTP — same as `teams`. Latency slightly higher but same JSON shape.
- **Budget is cents, not dollars.** $50 = `--budget 5000`.
- **Agent type:** `open` (user-designed prompt) vs `predefined` (built-in role). Check server docs for full list.
- **Wake:** idempotent — waking awake agent is no-op.

## Cross-refs

- Sharing, linking, delegation: [agents-advanced.md](agents-advanced.md)
- Chat with agent: [chat-sessions.md](chat-sessions.md)
- Run tools as agent: [exec-workflow.md](exec-workflow.md)
