# Users (search & management)

## When to use

User wants to search for users by email, name, or ID, or check user details. Different from `tenants users` (which is tenant-level user management). Use this for cross-tenant user discovery.

## Commands in scope

- `goclaw users search <query>` — search users by email, name, or ID (source: `cmd/users.go`)

## Verified flags

### `users search`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<query>` | string | Search term (email prefix, name, or user ID) |

## JSON output

- ✅ `users search` — JSON array of user objects

## Destructive ops

None — read-only.

## Common patterns

### Example 1: search by email prefix
```bash
goclaw users search alice@example.com --output json
# → [
#   {"id": "user_abc123", "email": "alice@example.com", "name": "Alice Chen", "created_at": "2025-01-15"},
#   {"id": "user_def456", "email": "alice.backup@example.com", "name": "Alice (Backup)", "created_at": "2025-02-01"}
# ]
```

### Example 2: search by name
```bash
goclaw users search bob --output json
# → [
#   {"id": "user_ghi789", "email": "bob.smith@example.com", "name": "Bob Smith", "created_at": "2025-03-10"},
#   {"id": "user_jkl012", "email": "bob@another.co", "name": "Bob Johnson", "created_at": "2025-04-05"}
# ]
```

### Example 3: search by user ID
```bash
goclaw users search user_abc123 --output json
```

## Edge cases & gotchas

- **Search scope:** typically cross-tenant (returns users from all tenants). Use `--tenant-id` global flag if server supports tenant-scoped search.
- **Query format:** free-text substring match. Results ranked by relevance (exact match > prefix > substring).
- **Privacy:** visible users depend on your role and tenants. Non-admins may see only users in their own tenant.
- **Active vs inactive:** response may include deactivated users. Check `active` field.

## Cross-refs

- Tenant user management: [admin-system.md](admin-system.md) — `tenants users`
- Authentication + profiles: [auth-and-config.md](auth-and-config.md)
- Team collaboration: [teams-collaboration.md](teams-collaboration.md)
