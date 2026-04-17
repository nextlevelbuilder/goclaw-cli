# Data movement — export / import / storage

## When to use

User wants to export agents/teams/skills/mcp configs for backup or porting, import previously exported bundles, or manage workspace files (upload/download/move/delete).

## Commands in scope

### Export (source: `cmd/export_import.go`)
- `goclaw export <resource>-preview <id>` — preview export JSON (safe, no side-effects)
- `goclaw export <resource> <id> [-f <file>]` — export to file
- `<resource>` ∈ `agent` / `team` / `skills` / `mcp`

### Import (source: `cmd/export_import.go`)
- `goclaw import <resource>-preview <file>` — preview import (no writes)
- `goclaw import <resource> <file> --yes` — apply import (OVERWRITES)

### Storage (source: `cmd/storage.go`)
- `goclaw storage list [--path <subdir>]` — list files
- `goclaw storage get <path> [-f <file>]` — download file (or stdout)
- `goclaw storage download <path> [-f <file>]` — forced download with Content-Disposition
- `goclaw storage delete <path>` — remove file
- `goclaw storage move --from <src> --to <dst>` — rename/move
- `goclaw storage size` — usage stats

## Verified flags

### `storage list`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--path <sub>` | string | Sub-directory filter |

### `storage get` / `storage download`
| Flag | Type | Purpose |
| --- | --- | --- |
| `-f, --output <file>` | string | Write to file (default stdout) |

### `storage move`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--from <src>` (REQUIRED) | string | Source path |
| `--to <dst>` (REQUIRED) | string | Destination path |

### `export agent` / `export <resource>`
| Flag | Type | Purpose |
| --- | --- | --- |
| `-f, --output <file>` | string | Output file (defaults to `<resource>-<id>.json`) |

## JSON output

- ✅ `storage list`, `storage size` — JSON
- ✅ All `export *-preview` and `import *-preview` — JSON
- ⚠️ `storage get/download` — raw file bytes (not JSON); Claude must avoid piping large binaries into context
- ⚠️ `storage delete/move/put`, `import *` — success text only
- ⚠️ `export <resource>` (non-preview) — writes file; success text to stdout

## Destructive ops

| Command | Confirm |
| --- | --- |
| `storage delete` | YES (permanent) |
| `storage move` | YES (if overwriting destination) |
| `import <resource>` with `--yes` | YES — OVERWRITES existing resource |

## Common patterns

### Example 1: preview + export agent
```bash
goclaw export agent-preview <agent-id> --output json
goclaw export agent <agent-id> -f /tmp/agent-backup.json
```

### Example 2: import with dry-run preview first
```bash
goclaw import agent-preview /tmp/agent-backup.json --output json
# review, then apply:
goclaw import agent /tmp/agent-backup.json --yes
```

### Example 3: list + download workspace file
```bash
goclaw storage list --path projects/ --output json
goclaw storage get projects/plan.md -f /tmp/plan.md
```

### Example 4: move file
```bash
goclaw storage move --from drafts/post.md --to published/post.md
```

### Example 5: check storage quota
```bash
goclaw storage size --output json
# → {"used_bytes": N, "quota_bytes": M, ...}
```

## Edge cases & gotchas

- **Large exports/imports** may exceed Bash tool 120s timeout. For big agents (many files, long memory), recommend user run manually or use `--output` to file and process offline.
- **Storage paths** are NOT URL-escaped by CLI (source: `storage.go:25-27`, "Don't escape path separators"). Server expects raw. Spaces/unicode may fail — warn user.
- **`storage get` with no `-f`** writes to stdout. Binary files will corrupt Claude's context. Always use `-f` for non-text.
- **Import overwrites by default.** No diff view on apply. Always preview first.
- **Export bundles** include agent/team IDs — re-importing to same server creates duplicates unless server dedupes on key. Check server behavior.
- **Resource types:** only 4 supported — `agent`/`team`/`skills`/`mcp`. Others (channels, memory, sessions) can't be exported this way.

## Cross-refs

- Skills upload (different from export): [providers-skills-tools.md](providers-skills-tools.md)
- Media upload/download (separate from storage): [media.md](media.md)
- Memory docs (agent-scoped, not tenant-storage): [knowledge-memory.md](knowledge-memory.md)
