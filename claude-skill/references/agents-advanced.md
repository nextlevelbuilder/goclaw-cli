# Agents — advanced (links, ops, delegations)

## When to use

User wants to share agents across users, create delegation links between agents, regenerate/resummon an agent, or wait for an agent to finish.

## Commands in scope

- `goclaw agents links list/create/update/delete` — delegation links (source: `cmd/agents_links.go`)
- `goclaw agents share/unshare <id>` — user sharing (source: `cmd/agents_ops.go:11-49`)
- `goclaw agents regenerate/resummon <id>` — reset agent setup
- `goclaw agents wait <id>` — **WS-based blocking wait** (source: `cmd/agents_ops.go:87`)
- `goclaw delegations list/get` — delegation history (source: `cmd/admin.go:86-129`)

## Verified flags

### `agents links create`
| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--source <id>` | string | — | Source agent |
| `--target <id>` | string | — | Target agent |
| `--direction <d>` | string | `outbound` | `outbound`/`inbound`/`bidirectional` |
| `--max-concurrent <n>` | int | 3 | Concurrent delegation cap |

### `agents share`
| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--user <id>` | string (REQUIRED) | — | User to share with |
| `--role <r>` | string | `operator` | `admin`/`operator`/`viewer` |

### `agents unshare`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--user <id>` | string (REQUIRED) | User to revoke |

### `agents wait`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--session <key>` | string | Session to wait on |
| `--timeout <sec>` | int (0 = no timeout) | Bash-safe max ~100 |

### `delegations list`
| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--agent <id>` | string | — | Filter by agent |
| `--limit <n>` | int | 20 | Max results |

## JSON output

- ✅ `agents links list` — JSON
- ✅ `delegations list/get` — JSON
- ✅ `agents wait` — JSON final state (blocks until complete or timeout)
- ⚠️ `share/unshare/regenerate/resummon/links create|update|delete` — success text only

## Destructive ops

| Command | Source | Confirm |
| --- | --- | --- |
| `agents links delete` | `agents_links.go:99` tui.Confirm | YES |
| `agents unshare` | removes user access | YES |

## Common patterns

### Example 1: share agent with user as operator
```bash
goclaw agents share <agent-id> --user user-42 --role operator
```

### Example 2: create bidirectional delegation link
```bash
goclaw agents links create \
  --source agent-alpha --target agent-beta \
  --direction bidirectional --max-concurrent 5 \
  --output json
```

### Example 3: wait for agent to finish (with timeout)
```bash
# Bash tool default timeout = 120s, set --timeout lower
goclaw agents wait <agent-id> --session <key> --timeout 90 --output json
```

### Example 4: delegation audit
```bash
goclaw delegations list --agent <agent-id> --limit 50 --output json
goclaw delegations get <delegation-id> --output json
```

### Example 5: regenerate after model swap
```bash
goclaw agents update <id> --model claude-opus-4-7
goclaw agents regenerate <id>
```

## Edge cases & gotchas

- **`agents wait` WS subscribe:** blocks until agent idle OR timeout. With `--timeout 0` it blocks indefinitely → Bash tool kills at 120s. Always set a `--timeout < 110`.
- **Direction `inbound`:** target receives delegations from source, not the other way. Mental model: "direction of the data flow".
- **Role hierarchy:** `admin` > `operator` > `viewer`. Downgrade requires `unshare` + `share` with new role.
- **`regenerate` vs `resummon`:** regenerate rebuilds agent config from current settings; resummon reruns initial setup wizard (may prompt on server side).
- **Delegations view is read-only:** no create/delete — links drive the delegation; history is automatic.

## Cross-refs

- Agent base CRUD: [agents-core.md](agents-core.md)
- Team tasks (different concept): [teams-collaboration.md](teams-collaboration.md)
