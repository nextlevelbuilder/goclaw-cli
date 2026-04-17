# Knowledge graph & memory

## When to use

User wants to store/retrieve agent memory docs (key-value-ish, with semantic search), or manage the agent's knowledge graph (entities + relations), or run entity deduplication.

## Commands in scope

### Memory (source: `cmd/memory.go`)
- `goclaw memory list <agent-id>` — list documents
- `goclaw memory get <agent-id> <path>` — fetch document
- `goclaw memory store <agent-id> <path> --content <text|@file>` — upsert
- `goclaw memory delete <agent-id> <path>` — delete document
- `goclaw memory search <agent-id> --query "..."` — semantic search

### Knowledge Graph (source: `cmd/knowledge_graph.go`, alias `kg`)
- `goclaw kg entities list/get/create/delete <agent-id>`
- `goclaw kg traverse <agent-id> --from <entity>`
- `goclaw kg graph <agent-id>` — full graph dump
- `goclaw kg stats <agent-id>` — node/edge counts
- `goclaw kg query <agent-id> [--entity <name>]` — legacy query (source: `memory.go:121-141`)
- `goclaw kg extract <agent-id> --text "..."` — extract entities from text
- `goclaw kg link <agent-id> --from <e1> --to <e2> --relation <r>` — create link

### Dedup (source: `cmd/knowledge_graph_dedup.go`)
- `goclaw kg dedup scan <agent-id>` — scan for duplicates
- `goclaw kg dedup merge-candidates <agent-id>` — list candidate pairs
- `goclaw kg dedup execute-merge <agent-id> --pair <id>` — **IRREVERSIBLE merge**

## Verified flags

### `memory list`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--user <id>` | string | Filter by user |

### `memory store` / `kg extract`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--content <v>` (memory store) | string (REQUIRED) | Value or `@filepath` |
| `--text <v>` (kg extract) | string (REQUIRED) | Text or `@filepath` |

### `memory search`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--query <q>` | string (REQUIRED) | Search query |
| `--user <id>` | string | Filter by user |

### `kg entities create`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--data <json>` | string (REQUIRED) | Entity body JSON or `@filepath` |

### `kg traverse` / `kg link`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--from <id>` | string (REQUIRED) | Start entity |
| `--to <id>` | string (link only, REQUIRED) | Target |
| `--relation <r>` | string (link only, REQUIRED) | Relation type |

## JSON output

- ✅ All `list/get/search/query/graph/stats/traverse` — JSON
- ✅ `kg entities create` — JSON
- ⚠️ `memory store/delete`, `kg link`, `kg dedup execute-merge` — success text

## Destructive ops

| Command | Source | Confirm |
| --- | --- | --- |
| `memory delete` | `memory.go:83` tui.Confirm | YES |
| `kg entities delete` | `knowledge_graph.go:85` tui.Confirm | YES |
| `kg dedup execute-merge` | **IRREVERSIBLE** | YES — always confirm pair |
| `memory clear` (if exists) | wipes all memory for agent | YES — high blast radius |

## Common patterns

### Example 1: store + retrieve memory doc
```bash
goclaw memory store <agent-id> projects/q3-roadmap --content "@./roadmap.md"
goclaw memory get <agent-id> projects/q3-roadmap --output json
```

### Example 2: semantic search memory
```bash
goclaw memory search <agent-id> --query "customer churn analysis" --output json
```

### Example 3: extract entities + link
```bash
goclaw kg extract <agent-id> --text "Alice works at Acme Corp" --output json
goclaw kg link <agent-id> --from "Alice" --to "Acme Corp" --relation "works_at"
```

### Example 4: dedup workflow
```bash
goclaw kg dedup scan <agent-id> --output json
goclaw kg dedup merge-candidates <agent-id> --output json
# review pairs, then explicit merge:
goclaw kg dedup execute-merge <agent-id> --pair <pair-id>  # CONFIRM FIRST
```

### Example 5: graph overview
```bash
goclaw kg stats <agent-id> --output json
goclaw kg graph <agent-id> --output json | jq '.nodes | length'
```

## Edge cases & gotchas

- **`--content` / `--text` with `@file`:** `readContent()` dereferences `@path` to file contents — write prompt handles both, no escaping needed.
- **`memory store` overwrites** silently at same path. No version history at CLI level.
- **`kg dedup execute-merge`** merges entities + their relations + their memory refs — cannot split back. Always list candidates first.
- **Memory path namespace:** paths are slash-separated (`projects/q3-roadmap`), can contain slashes but no leading `/`. URL-encoded on the wire.
- **`memory.clear` command:** check binary — may or may not be present. If invoked, clears ENTIRE agent memory — skill must force explicit double-confirm.
- **KG endpoints scoped per-agent:** all routes under `/v1/agents/<id>/kg/...`. No global-tenant graph.

## Cross-refs

- Built-in tools for agent-side memory access: [providers-skills-tools.md](providers-skills-tools.md) — `memory_search`, `knowledge_graph_search` tools
- Agent base: [agents-core.md](agents-core.md)
