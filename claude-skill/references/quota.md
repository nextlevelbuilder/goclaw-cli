# Quota inspection

## When to use

User wants to check quota consumption (token usage, API calls, storage, etc.) across all users/agents or for a specific agent.

## Commands in scope

- `goclaw quota usage` — show quota consumption (source: `cmd/quota.go`)

## Verified flags

### `quota usage`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--agent <id>` | string | Filter by specific agent (default: all) |

## JSON output

- ✅ `quota usage` — JSON object with quota metrics

## Destructive ops

None — read-only.

## Common patterns

### Example 1: check system-wide quota
```bash
goclaw quota usage --output json
# → {
#   "period": "2026-06-01 to 2026-06-30",
#   "total_tokens": 1000000,
#   "used_tokens": 750000,
#   "remaining": 250000,
#   "agents": [
#     {"agent_id": "abc", "tokens_used": 500000},
#     {"agent_id": "xyz", "tokens_used": 250000}
#   ]
# }
```

### Example 2: check quota for specific agent
```bash
goclaw quota usage --agent <agent-id> --output json
# → {"agent_id": "...", "tokens_used": 500000, "percentage": 50}
```

## Edge cases & gotchas

- **Quota period:** typically monthly, reset on the 1st. Response includes period dates.
- **Token counting:** may vary by model (gpt-4 more expensive than gpt-3.5). Check server docs for counting rules.
- **Soft vs hard limits:** some tenants have soft warnings (80% alerts) vs hard caps (reject requests at 100%). Check `system-config` for limits.
- **Real-time lag:** quota updates may lag 1-5 minutes behind actual usage. Use for trend analysis, not immediate tracking.

## Cross-refs

- Usage analytics (cost): [monitoring-ops.md](monitoring-ops.md) — `usage summary/detail`
- Storage quota: [data-movement.md](data-movement.md) — `storage size`
- System configuration: [admin-system.md](admin-system.md) — `system-config`
