# Workstations (coding-agent)

## When to use

User wants to set up isolated coding environments (workstations) for agents, link/unlink agents to workstations, manage file permissions, or inspect workstation activity.

## Commands in scope

- `goclaw workstations list` — list all workstations (source: `cmd/workstations.go`)
- `goclaw workstations get <id>` — get workstation details + status
- `goclaw workstations create` — create new workstation
- `goclaw workstations update <id>` — update workstation config
- `goclaw workstations delete <id>` — delete workstation
- `goclaw workstations link-agent <id>` — link agent to workstation over WS RPC
- `goclaw workstations unlink-agent <id>` — unlink agent from workstation
- `goclaw workstations permissions` — manage path-based file permissions
- `goclaw workstations activity <id>` — list recent activity (file ops, command invocations)

## Verified flags

### `workstations create/update`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--name <n>` | string | Display name |
| `--workstation-key <k>` | string | Stable key (for referencing workstation) |
| `--backend-type <t>` | string | Backend type (e.g., `docker`, `k8s`, `local`) |
| `--default-cwd <d>` | string | Default working directory (e.g., `/workspace`) |
| `--default-env <j>` | JSON | Default environment vars as JSON object |
| `--metadata <m>` | JSON | Custom metadata |

### `workstations link-agent`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<id>` | workstation ID | Workstation to link |
| `--agent-id <id>` | string | Agent ID to link |

### `workstations unlink-agent`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<id>` | workstation ID | — |
| `--agent-id <id>` | string | Agent to unlink |

### `workstations permissions`
See subcommands: `list`, `grant`, `revoke` (flags vary).

### `workstations activity`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<id>` | workstation ID | — |
| `--limit <n>` | int | Max log entries |

## JSON output

- ✅ `workstations list/get/activity` — JSON
- ⚠️ `workstations create/update/delete/link-agent/unlink-agent` — success text
- ✅ `workstations permissions list` — JSON

## Destructive ops

| Command | Confirm |
| --- | --- |
| `workstations delete` | YES (agents lose workspace) |

## Common patterns

### Example 1: list workstations
```bash
goclaw workstations list --output json
```

### Example 2: create workstation with environment
```bash
goclaw workstations create \
  --name "python-ml" \
  --workstation-key "ml-workspace" \
  --backend-type docker \
  --default-cwd "/workspace" \
  --default-env '{"PYTHON_VERSION":"3.11","CONDA_ENV":"ml"}' \
  --output json
```

### Example 3: link agent to workstation
```bash
goclaw workstations link-agent <workstation-id> \
  --agent-id <agent-id> \
  --output json
```

### Example 4: view workstation activity
```bash
goclaw workstations activity <id> --limit 50 --output json
# → recent command invocations, file ops, errors
```

### Example 5: manage file permissions
```bash
goclaw workstations permissions list <workstation-id> --output json
goclaw workstations permissions grant <workstation-id> \
  --path "/workspace/sensitive" \
  --agent-id <agent-id> \
  --access "read-only"
```

## Edge cases & gotchas

- **Backend types:** `docker`, `k8s`, `local` (varies by server). Server must support chosen type.
- **Link/unlink:** WS RPC operations. If workstation unreachable, returns error. Retry or check workstation health.
- **File permissions:** path-based. Deny patterns can exclude sensitive dirs. Agents get 403 if accessing denied paths.
- **Activity log:** retention server-dependent. Check `system-config` for activity retention window.
- **Default env vars:** merged with agent-level env vars. Agent vars override workstation defaults.
- **Metadata:** custom fields stored; not indexed. Useful for tagging (e.g., `{"team": "ml", "region": "us-west"}`).

## Cross-refs

- Agent setup + linking: [agents-core.md](agents-core.md)
- Tool invocation (exec): [exec-workflow.md](exec-workflow.md)
- System administration: [admin-system.md](admin-system.md)
