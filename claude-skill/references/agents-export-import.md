# Agents — export & import

## When to use

User wants to backup agents, migrate agents between systems, port agent configuration, or restore from archives.

## Commands in scope

- `goclaw agents export <id> [-o file]` — export agent as .tar.gz (source: `cmd/agents_export.go`)
- `goclaw agents import <file>` — import with preview (dry-run by default) (source: `cmd/agents_import.go`)
- `goclaw agents import <file> --apply` — perform import (overwrites)
- `goclaw agents import-merge <file>` — merge imported agent into existing agent

## Verified flags

### `agents export`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<id>` | agent ID | Agent to export |
| `-o, --file <path>` | string | Output file (default: stdout) |

### `agents import`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<file>` | path | Archive file (.tar.gz) |
| `--apply` | bool | Perform import (default: preview only) |

### `agents import-merge`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<file>` | path | Archive file (.tar.gz) |
| `--apply` | bool | Perform merge (default: preview only) |

## JSON output

- ✅ `agents export` — binary .tar.gz (use `-o` to file)
- ✅ `agents import/import-merge` preview — JSON (shows what would import)
- ⚠️ `agents import/import-merge --apply` — success text only

## Destructive ops

| Command | Confirm |
| --- | --- |
| `agents import --apply` | YES (overwrites agent) |
| `agents import-merge --apply` | YES (merges into existing) |

## Common patterns

### Export agent for backup
```bash
goclaw agents export <agent-id> -o /tmp/agent-backup.tar.gz
# → agent-backup.tar.gz with config, files, memory, skills
```

### Preview import
```bash
goclaw agents import /tmp/agent-backup.tar.gz --output json
# shows what would be imported
```

### Perform import
```bash
goclaw agents import /tmp/agent-backup.tar.gz --apply --yes
```

### Merge imported agent into existing
```bash
goclaw agents import-merge /tmp/agent-backup.tar.gz --apply --yes
```

## Edge cases & gotchas

- **Archive contents:** includes config, context files, episodic memory, granted skills, and metadata.
- **Archive size:** grows with agent history. Large agents (many messages, big memory) produce multi-GB archives.
- **Export to stdout:** without `-o`, exports to stdout. Can be piped but binary data → be careful.
- **Import overwrites:** replaces entire agent config. Agent ID may change on import.
- **Merge preserves existing:** merges imported config/memory into current agent. Less destructive than import.
- **Portability:** archives are version-tied. Importing into newer/older server version may fail or degrade.

## Cross-refs

- Agent CRUD: [agents-crud.md](agents-crud.md)
- Data movement: [data-movement.md](data-movement.md)
