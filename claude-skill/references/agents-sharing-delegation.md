# Agents — sharing & delegation

## When to use

User wants to share agents with other users, create delegation links between agents, manage user access roles, view delegation history, or wait for agent completion.

## Commands in scope

- `goclaw agents share <id> --user <uid>` — grant user access (source: `cmd/agents_ops.go`)
- `goclaw agents unshare <id> --user <uid>` — revoke user access
- `goclaw agents links list/create/update/delete` — agent-to-agent delegation links (source: `cmd/agents_links.go`)
- `goclaw agents wait <key>` — block until agent reaches target state (source: `cmd/agents_ops.go`)
- `goclaw agents regenerate <id>` — regenerate agent config (source: `cmd/agents_ops.go`)
- `goclaw agents resummon <id>` — re-trigger agent setup
- `goclaw delegations list/get` — delegation history (read-only) (source: `cmd/admin.go`)

## Verified flags

### `agents share`
| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--user <id>` | string | (REQUIRED) | User ID |
| `--role <r>` | string | `operator` | `admin`, `operator`, `viewer` |

### `agents unshare`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--user <id>` | string | User to revoke |

### `agents links create/update`
| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--source <id>` | string | — | Source agent |
| `--target <id>` | string | — | Target agent |
| `--direction <d>` | string | `outbound` | `outbound`, `inbound`, `bidirectional` |
| `--max-concurrent <n>` | int | 3 | Concurrent delegation cap |

### `agents links delete`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<link-id>` | string | Link ID |

### `agents wait`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<key>` | string | Agent key to wait for |
| `--state <s>` | string | Target: `online`, `running`, `idle`, `offline` |
| `--timeout <d>` | duration | Wait timeout (default 5m) |

### `agents regenerate/resummon`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<id>` | agent ID | Agent ID |

### `delegations list`
| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--agent <id>` | string | — | Filter by agent |
| `--limit <n>` | int | 20 | Max results |

## JSON output

- ✅ `agents links list`, `delegations list/get` — JSON
- ✅ `agents wait` — JSON final state (blocks until complete/timeout)
- ⚠️ `agents share/unshare/regenerate/resummon`, `agents links create/update/delete` — success text only

## Destructive ops

| Command | Confirm |
| --- | --- |
| `agents unshare` | YES (revokes access) |
| `agents links delete` | YES |
| `agents regenerate` | YES (rebuilds config) |

## Common patterns

### Share with user
```bash
goclaw agents share <agent-id> --user user-42 --role operator
```

### Create delegation link
```bash
goclaw agents links create \
  --source agent-alpha \
  --target agent-beta \
  --direction bidirectional \
  --max-concurrent 5 \
  --output json
```

### View delegation links
```bash
goclaw agents links list --output json
```

### View delegation history
```bash
goclaw delegations list --agent <agent-id> --limit 50 --output json
goclaw delegations get <delegation-id> --output json
```

### Wait for agent to reach state
```bash
goclaw agents wait <agent-key> --state idle --timeout 30s --output json
```

### Regenerate after model update
```bash
goclaw agents update <id> --model claude-opus-4-7
goclaw agents regenerate <id>
```

### Revoke user access
```bash
goclaw agents unshare <agent-id> --user user-42 --yes
```

## Edge cases & gotchas

- **Role hierarchy:** `admin` > `operator` > `viewer`. Downgrade requires `unshare` + `share` with new role.
- **Direction:** `inbound` = target receives delegations from source; `outbound` = source delegates to target.
- **Max concurrent:** caps number of active delegations at once. Excess requests queue.
- **`agents wait` timeout:** blocks until state reached OR timeout expires. Always set `--timeout < 110s` for Bash tool (default 120s timeout).
- **Regenerate vs resummon:** regenerate = rebuild config from current settings; resummon = re-trigger setup wizard.
- **Delegations read-only:** history view only; links drive the delegation. No direct delegation CRUD.

## Cross-refs

- Agent CRUD: [agents-crud.md](agents-crud.md)
- Export/import: [agents-export-import.md](agents-export-import.md)
- Team collaboration: [teams-collaboration.md](teams-collaboration.md)
