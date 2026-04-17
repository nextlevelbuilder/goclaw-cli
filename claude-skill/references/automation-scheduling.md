# Automation & scheduling (cron, heartbeat, devices)

## When to use

User wants to schedule recurring agent runs (cron), configure agent heartbeats/health-checks, or manage paired devices.

## Commands in scope

### Cron (source: `cmd/cron.go`) — WS-backed
- `goclaw cron list/get/create/update/delete/toggle/status`
- `goclaw cron run <id>` — manual trigger
- `goclaw cron runs <id>` — run history

### Heartbeat (source: `heartbeat.go`, `heartbeat_checklist_targets.go`)
- `goclaw heartbeat get/set`
- `goclaw heartbeat checklist get/set`
- `goclaw heartbeat targets list/add/remove`

### Devices (source: `devices.go`)
- `goclaw devices list/delete/approve/reject`

## Verified flags

### `cron create` (all REQUIRED except `message`, `timezone`)
| Flag | Type | Purpose |
| --- | --- | --- |
| `--agent <id>` | REQUIRED | Agent to trigger |
| `--name <n>` | REQUIRED | Job name |
| `--schedule <expr>` | REQUIRED | Cron expression (e.g. `"0 9 * * 1"`) |
| `--message <m>` | string | Payload message |
| `--timezone <tz>` | string | IANA TZ (e.g. `Asia/Saigon`) |

### `cron update`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--name <n>` | string | Rename |
| `--schedule <expr>` | string | Change schedule |

### `cron runs`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--limit <n>` | int | Default 20 |

### `heartbeat set`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--agent <id>` | string | Target agent |
| `--interval <sec>` | int | Heartbeat interval |
| `--enabled <bool>` | bool | Turn on/off |

### `heartbeat targets add/remove` (verify with --help)
Exact flag names depend on CLI version. Run `goclaw heartbeat targets add --help`.

### `devices approve/reject` (verify with --help)
Exact args (positional vs flag) depend on CLI version. Run `goclaw devices approve --help`.

## JSON output

- ✅ `cron list/get/status/runs`, `heartbeat get`, `heartbeat checklist get`, `heartbeat targets list`, `devices list` — JSON
- ⚠️ All `create/update/delete/toggle/run/set/add/remove/approve/reject` — success text

## Destructive ops

| Command | Confirm |
| --- | --- |
| `cron delete` | YES (loses schedule) |
| `cron toggle` (if disabling) | YES |
| `heartbeat set --enabled=false` | YES (health checks stop) |
| `heartbeat targets remove` | YES |
| `devices delete` | YES (unpairs) |
| `devices reject` | YES (rejects pairing) |

## Common patterns

### Example 1: weekday morning agent run
```bash
goclaw cron create \
  --agent <agent-id> \
  --name "Daily triage" \
  --schedule "0 9 * * 1-5" \
  --message "Check overnight tickets" \
  --timezone "Asia/Saigon" \
  --output json
```

### Example 2: trigger cron manually + check status
```bash
goclaw cron run <cron-id>
goclaw cron status <cron-id> --output json
goclaw cron runs <cron-id> --limit 5 --output json
```

### Example 3: disable cron job temporarily
```bash
# Confirm intent, then:
goclaw cron toggle <cron-id>
```

### Example 4: heartbeat configuration
```bash
goclaw heartbeat get --help   # always check current flag names
goclaw heartbeat checklist set --data '{"enabled":true,"items":[]}'
```

### Example 5: approve paired device
```bash
goclaw devices list --output json
# syntax varies — check help:
goclaw devices approve --help
```

## Edge cases & gotchas

- **Cron is WS primary with HTTP fallback** (source: `cron.go:20-32`). If WS unavailable, some commands fall back to HTTP.
- **Cron expression format:** standard 5-field (min hour day month weekday) OR interval format. Check server parser for extensions.
- **Timezone** is server-interpreted. If omitted, defaults to server TZ (check `goclaw status`).
- **`cron run` is idempotent** but creates a run entry each time — avoid spam clicking.
- **Heartbeat interval minimum:** server-enforced (typically ≥ 30s). Very short intervals rejected.
- **Devices `approve`** requires interactive TUI (source: `devices.go`) — pass `--yes` for automation.
- **`devices delete` after approve** = fully unpair (user must re-pair with fresh code).

## Cross-refs

- Device pairing auth flow: [auth-and-config.md](auth-and-config.md) — `auth pair`
- Cron-triggered agent: [agents-core.md](agents-core.md)
