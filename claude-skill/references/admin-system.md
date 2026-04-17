# Admin & system

## When to use

User wants admin-only operations: multi-tenant management, per-tenant config KV store, view audit log, check TTS status. **Requires admin role** — check `goclaw whoami` first.

## Commands in scope

- `goclaw tenants list/get/create/update/delete/users` — tenant CRUD (source: `cmd/tenants.go`)
- `goclaw system-config list/get/set/delete` — per-tenant KV config (source: `system_config.go`)
- `goclaw activity list` — audit log (source: `admin_activity.go`)
- `goclaw tts status` — TTS service status (source: `admin_tts.go`, WS call)

## Verified flags

### `tenants create/update/users` (verify with --help)
Run `goclaw tenants create --help` / `goclaw tenants users --help` for current flag names.
Common fields: tenant `name`, user-management `action` (add/remove), user ID.

### `system-config set`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--key <k>` | string | Config key |
| `--value <v>` | string | New value |

### `system-config get/delete`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--key <k>` | string | Config key |

### `activity list` (verify with --help)
- `--limit <n>` — max results (typical)
- additional filter flags vary; run `goclaw activity list --help`

## JSON output

- ✅ `tenants list/get/users`, `system-config list/get`, `activity list`, `tts status` — JSON
- ⚠️ `tenants create/update/delete`, `system-config set/delete` — success text

## Destructive ops

| Command | Confirm |
| --- | --- |
| `tenants delete` | **YES — CRITICAL**; cascades ALL tenant data |
| `tenants users --action remove` | YES |
| `system-config delete` | YES |

## Common patterns

### Example 1: list all tenants (admin)
```bash
goclaw tenants list --output json
```

### Example 2 & 3: create tenant + add user (flag names vary)
```bash
goclaw tenants create --help    # always check current flags
goclaw tenants users --help
```

### Example 4: per-tenant config override
```bash
# Read current
goclaw --tenant-id acme system-config list --output json
# Set feature flag
goclaw --tenant-id acme system-config set --key "feature.streaming" --value "true"
```

### Example 5: audit trail
```bash
goclaw activity list --limit 50 --output json
# for filters, check --help for current flag names
```

### Example 6: check TTS service
```bash
goclaw tts status --output json
# → {"enabled": true, "provider": "elevenlabs", "voices_count": 42}
```

## Edge cases & gotchas

- **Admin-only:** commands here require `admin` role in the active tenant. Check `goclaw whoami --output json` — if `role` ≠ `admin`, commands will 403.
- **`tenants delete` blast radius:** deletes ALL agents, sessions, memory, channels for that tenant. Triple-confirm. No undo.
- **`--tenant-id` global flag** (source: root.go) scopes `system-config` and most other commands to a specific tenant even if you're a cross-tenant admin.
- **System-config keys** are free-form strings; conventions (e.g. `feature.*`, `limit.*`) are app-level, not enforced.
- **Activity log retention:** server-dependent. Check with `system-config get --key "activity.retention_days"`.
- **TTS** is read-only here; actual TTS generation is a builtin tool (`tools invoke tts --param text="..."`).

## Cross-refs

- TTS invoked as tool: [exec-workflow.md](exec-workflow.md) — `tools invoke tts`
- User auth/profile (non-admin): [auth-and-config.md](auth-and-config.md)
- Per-config permissions: [auth-and-config.md](auth-and-config.md) — `config permissions`
