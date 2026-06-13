# Agents — orchestration & identity

## When to use

User wants to check agent orchestration config, view agent identity (persona, traits, goals), preview system prompts, inspect granted skills, manage lifecycle (summon, cancel, sync), or toggle experimental flags.

## Commands in scope

- `goclaw agents orchestration <id>` — get orchestration mode + delegation config (source: `cmd/agents_orchestration.go`)
- `goclaw agents identity <key>` — get agent identity (persona, traits, goals, constraints)
- `goclaw agents prompt-preview <id>` — get fully rendered system prompt
- `goclaw agents skills list <id>` — list skills granted to agent
- `goclaw agents cancel-summon <id>` — cancel pending/running summon (source: `cmd/agents_summon.go`)
- `goclaw agents sync-workspace` — admin: sync workspace for all agents
- `goclaw agents v3-flags get <id>` — get v3 feature flags (experimental)
- `goclaw agents v3-flags toggle <id>` — toggle v3 feature flag

## Verified flags

### `agents orchestration`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<id>` | agent ID | Agent ID |

### `agents identity`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<key>` | string | Agent key |

### `agents prompt-preview`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<id>` | agent ID | Agent ID |

### `agents skills list`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<id>` | agent ID | Agent ID |

### `agents cancel-summon`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<id>` | agent ID | Agent to cancel |

### `agents sync-workspace`
Admin-only; no flags (affects all agents).

### `agents v3-flags toggle`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<id>` | agent ID | Agent ID |
| `--flag <name>` | string | Flag name to toggle |

## JSON output

- ✅ `agents orchestration`, `agents identity`, `agents prompt-preview`, `agents skills list` — JSON
- ⚠️ `agents cancel-summon`, `agents sync-workspace`, `agents v3-flags get/toggle` — success text only

## Destructive ops

| Command | Confirm |
| --- | --- |
| `agents v3-flags toggle` | YES (experimental feature) |

## Common patterns

### Get orchestration config
```bash
goclaw agents orchestration <agent-id> --output json
# → delegation mode, concurrent limits, etc.
```

### Get agent identity
```bash
goclaw agents identity <agent-key> --output json
# → persona, traits, goals, constraints
```

### Preview system prompt
```bash
goclaw agents prompt-preview <agent-id> --output json | jq '.prompt' -r
# shows final prompt after variable substitution + file inclusion
```

### List granted skills
```bash
goclaw agents skills list <agent-id> --output json
```

### Cancel summon in progress
```bash
goclaw agents cancel-summon <agent-id>
```

### Sync workspace (admin)
```bash
goclaw agents sync-workspace
```

### Get v3 feature flags
```bash
goclaw agents v3-flags get <agent-id> --output json
```

### Toggle experimental flag
```bash
goclaw agents v3-flags toggle <agent-id> --flag streaming-responses --yes
```

## Edge cases & gotchas

- **Orchestration:** read-only endpoint showing delegation mode (independent, collaborative, hierarchical).
- **Identity:** WS-backed; includes persona, traits, goals, constraints. Read-only.
- **Prompt preview:** shows final prompt after variable substitution, context file inclusion, skill injection, etc. Useful for debugging.
- **Skills:** list only; grant/revoke via provider's `skills` commands.
- **Cancel-summon:** stops pending/running agent setup. Agent may require `resummon` to retry.
- **Sync-workspace:** admin-only. Re-syncs all agent workspaces. Long-running.
- **v3-flags:** experimental toggles. Not all agents support all flags. May require restart.

## Cross-refs

- Agent CRUD: [agents-crud.md](agents-crud.md)
- Sharing & delegation: [agents-sharing-delegation.md](agents-sharing-delegation.md)
- Memory & evolution: [agents-memory.md](agents-memory.md)
- Skills management: [providers-skills-tools.md](providers-skills-tools.md)
