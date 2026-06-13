# Agents — CRUD & lifecycle

## When to use

User wants to create, list, inspect, update, or delete agents; manage agent context files; control per-user instances; or wake agents.

## Commands in scope

- `goclaw agents list/get/create/update/delete` — agent CRUD (source: `cmd/agents.go`)
- `goclaw agents files list/get/create/delete` — agent context files (source: `agents_files.go`, WS-backed)
- `goclaw agents instances list/get/create/delete/trigger/reset` — per-user instances (source: `agents_instances.go`)
- `goclaw agents wake <id>` — wake sleeping agent (source: `agents_wake.go`)

## Verified flags

### `agents create/update`
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

- ✅ `agents list/get` — JSON
- ✅ `agents instances list/get` — JSON
- ⚠️ `agents create/update/delete/wake` — success text; parse exit code + stdout ID
- ⚠️ `agents files create/delete` — success text only
- ⚠️ `agents instances create/delete/trigger/reset` — success text

## Destructive ops

| Command | Confirm |
| --- | --- |
| `agents delete` | YES |
| `agents files delete` | YES |
| `agents instances delete` | YES |
| `agents instances reset` | YES (clears state) |

## Common patterns

### Create agent
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

### List agents
```bash
goclaw agents list --output json
```

### Get agent details
```bash
goclaw agents get <agent-id> --output json
```

### Update agent
```bash
goclaw agents update <agent-id> --model claude-opus-4-7 --budget 10000 --output json
```

### Manage agent files
```bash
goclaw agents files create --data '{"path":"guide.md","content":"..."}' --output json
goclaw agents files list --output json
goclaw agents files delete <file-id> --yes
```

### Create per-user instance
```bash
goclaw agents instances create --user-id user-42 --output json
goclaw agents instances list --user-id user-42 --output json
```

### Wake agent
```bash
goclaw agents wake <agent-id>
```

## Edge cases & gotchas

- **`--name` sets `display_name`** on create; update also accepts `--name` for `display_name` field.
- **Agent vs instance:** agent = template, instance = per-user runtime. Delete agent cascades instances.
- **Files WS-backed:** higher latency than HTTP, same JSON shape.
- **Budget in cents:** $50 = `--budget 5000`; not dollars.
- **Agent type:** `open` (user-designed) vs `predefined` (built-in role).
- **Wake:** idempotent — waking awake agent is no-op.
- **Instance reset:** clears message history + state, preserves agent/user bindings.

## Cross-refs

- Sharing + delegation: [agents-sharing-delegation.md](agents-sharing-delegation.md)
- Export/import: [agents-export-import.md](agents-export-import.md)
- Memory + evolution: [agents-memory.md](agents-memory.md)
- Orchestration + identity: [agents-orchestration.md](agents-orchestration.md)
- Chat with agent: [chat-basic.md](chat-basic.md)
