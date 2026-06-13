# Agents — memory & evolution

## When to use

User wants to inspect agent episodic memory, search learning history, view evolution metrics, apply AI-suggested improvements, or manage agent learning.

## Commands in scope

- `goclaw agents episodic list <id>` — list episodic memory entries (source: `cmd/agents_episodic.go`)
- `goclaw agents episodic search <id> <query>` — semantic search episodic memory
- `goclaw agents evolution metrics <id>` — get evolution metrics (source: `cmd/agents_evolution.go`)
- `goclaw agents evolution suggestions <id>` — list AI-generated improvements
- `goclaw agents evolution update <id>` — accept/reject evolution suggestion
- `goclaw agents evolution skill <id>` — apply skill evolution

## Verified flags

### `agents episodic list`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<id>` | agent ID | Agent ID |
| `--limit <n>` | int | Max entries |
| `--offset <n>` | int | Pagination offset |

### `agents episodic search`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<id>` | agent ID | Agent ID |
| `<query>` | string | Search query |
| `--limit <n>` | int | Max results |

### `agents evolution update`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<id>` | agent ID | Agent ID |
| `--suggestion-id <id>` | string | Suggestion ID |
| `--accept` | bool | Accept suggestion |
| `--reject` | bool | Reject suggestion |

## JSON output

- ✅ `agents episodic list/search`, `agents evolution metrics/suggestions` — JSON
- ⚠️ `agents evolution update/skill` — success text only

## Destructive ops

| Command | Confirm |
| --- | --- |
| `agents evolution update --reject` | YES (rejects suggestion) |

## Common patterns

### List episodic memory entries
```bash
goclaw agents episodic list <agent-id> --limit 20 --output json
```

### Search episodic memory
```bash
goclaw agents episodic search <agent-id> "user asked about pricing" --output json
```

### Get evolution metrics
```bash
goclaw agents evolution metrics <agent-id> --output json
```

### List evolution suggestions
```bash
goclaw agents evolution suggestions <agent-id> --output json
```

### Accept evolution suggestion
```bash
goclaw agents evolution update <agent-id> --suggestion-id <id> --accept --output json
```

### Reject evolution suggestion
```bash
goclaw agents evolution update <agent-id> --suggestion-id <id> --reject --yes
```

### Apply skill evolution
```bash
goclaw agents evolution skill <agent-id> --output json
```

## Edge cases & gotchas

- **Episodic memory:** semantic search depends on embedding provider. Server must have embeddings enabled.
- **Evolution metrics:** AI-generated suggestions. Accept/reject tracked server-side for learning.
- **Search queries:** free-text; server finds semantically similar entries, not just keyword matches.
- **Skill evolution:** applies evolved skills to agent capabilities. May require agent restart.
- **History retention:** episodic memory retention depends on server config. Check `system-config` for retention window.

## Cross-refs

- Agent CRUD: [agents-crud.md](agents-crud.md)
- Agent identity: [agents-orchestration.md](agents-orchestration.md)
