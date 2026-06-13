# Server edition & features

## When to use

User wants to check server edition (community, pro, enterprise), available feature set, or license info. **No authentication required** — unauthenticated users can query this.

## Commands in scope

- `goclaw edition` — display server edition, features, license (source: `cmd/system.go`, `edition` endpoint)

## Verified flags

None — simple read-only command.

## JSON output

- ✅ `edition` — JSON object with edition, features, limits

## Destructive ops

None — read-only.

## Common patterns

### Example 1: check server edition
```bash
goclaw edition --output json
# → {
#   "edition": "enterprise",
#   "license": "valid until 2026-12-31",
#   "features": {
#     "multi_tenant": true,
#     "sso": true,
#     "audit_log": true,
#     "backup_restore": true,
#     "workstations": true,
#     ...
#   }
# }
```

### Example 2: check specific feature availability
```bash
goclaw edition --output json | jq '.features.workstations'
# → true (feature available in this edition)
```

## Edge cases & gotchas

- **No auth required:** unlike most commands, `goclaw edition` works even without token. Useful for bootstrapping.
- **License expiry:** included in response. Feature access may degrade if license expires.
- **Features vary by edition:**
  - **Community:** agents, chat, basic webhooks
  - **Pro:** multi-tenant, SSO, API keys, vault
  - **Enterprise:** all + audit, workstations, backup/restore, custom MCP, priority support
- **Response format:** may include `trial_days_remaining` if on trial license.

## Cross-refs

- License + system health: [admin-system.md](admin-system.md) — `status`
- Backup/restore (enterprise): [backup-restore.md](backup-restore.md)
- Workstations (enterprise): [workstations.md](workstations.md)
