# Monitoring & ops

## When to use

User wants to check server health, view metadata, inspect LLM traces, watch usage/cost, or tail logs.

## Commands in scope

- `goclaw health` — HTTP health check (source: `cmd/status.go:12-26`)
- `goclaw status` — server status + metadata (WS call `status`, source: `cmd/status.go:28-75`)
- `goclaw version` — CLI version info
- `goclaw traces list` / `get <id>` / `export <id>` — LLM traces
- `goclaw usage summary` / `detail` / `trends` / `export` — cost analytics
- `goclaw logs tail` — **STREAMING, skill REFUSES** (source: `cmd/logs.go`)

## Verified flags

### `traces list`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--agent <id>` | string | Filter by agent |
| `--status <s>` | string | `running` / `success` / `error` |
| `--limit <n>` | int | Max results (default 20) |

### `traces export`
| Flag | Type | Purpose |
| --- | --- | --- |
| `-f, --output <file>` | string | Output file (default `<traceID>.json.gz`) |

### `usage summary`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--from <date>` | string | Start date (ISO) |
| `--to <date>` | string | End date (ISO) |

### `usage detail`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--agent <id>` | string | Filter by agent |
| `--provider <p>` | string | Filter by provider |

### `logs tail` (refuse — streaming)
| Flag | Type | Purpose |
| --- | --- | --- |
| `--agent` | string | Filter by agent ID |
| `--level` | string | `info` / `warn` / `error` |
| `--follow` | bool | Follow (default true) |

## JSON output

- ⚠️ `health` — `printer.Success` text only; check exit code (0 = healthy)
- ✅ `status` — JSON map via WS `status` call
- ⚠️ `version` — plaintext multi-line (not JSON)
- ✅ `traces list/get` — JSON
- ⚠️ `traces export` — writes gzip file, success text
- ✅ `usage summary/detail` — JSON
- ⚠️ `logs tail` — streams NDJSON but long-running; skill refuses

## Destructive ops

None — this cluster is read-only.

## Common patterns

### Example 1: health + status combo
```bash
goclaw health         # exit 0 = healthy (ignore stdout)
goclaw status --output json
```

### Example 2: look at recent LLM runs
```bash
goclaw traces list --limit 10 --status error --output json
goclaw traces get <trace-id> --output json
```

### Example 3: weekly cost breakdown
```bash
goclaw usage summary --from 2026-04-10 --to 2026-04-17 --output json
goclaw usage detail --agent <id> --output json
```

### Example 4: export trace for offline review
```bash
goclaw traces export <trace-id> -f /tmp/t.json.gz
# file is gzipped JSON — user unzips locally
```

### Example 5: `logs tail` alternative
Skill refuses streaming. Suggest:
> *"`logs tail` is a real-time stream and I can't run it through Bash. Please run `goclaw logs tail --agent <id>` in your own terminal, or let me query recent errors via `traces list --status error`."*

## Edge cases & gotchas

- **`status` WS fallback:** if WS handshake returns metadata, CLI returns that instead of calling `status` method (source: `status.go:48-52`). Both shapes valid.
- **`traces export` writes gzipped JSON** — users opening in editor see binary garbage unless they `gunzip` first.
- **Usage date format:** ISO-8601 (`YYYY-MM-DD`). Invalid dates → server 400.
- **Health vs status:** `health` is unauthenticated (no token needed); `status` requires auth + WS.
- **Cost figures** in `usage.*` responses are in cents per line item; sum carefully.

## Cross-refs

- Audit log (who did what): [admin-system.md](admin-system.md) → `activity list`
- MCP connection health: [mcp-integration.md](mcp-integration.md)
- Session-level tokens: [chat-sessions.md](chat-sessions.md)
