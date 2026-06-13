# Backup & restore

## When to use

Admin needs to create backups of system or tenant data, or restore from a previously created backup archive. **Requires admin role** — check `goclaw whoami` first.

## Commands in scope

### Backup (source: `cmd/backup.go`)
- `goclaw backup system` — create full system backup (returns download token)
- `goclaw backup system-download <token>` — download backup by token
- `goclaw backup system-preflight` — check disk space, pg_dump readiness
- `goclaw backup tenant` — create single tenant backup
- `goclaw backup tenant-download <token>` — download tenant backup by token
- `goclaw backup tenant-preflight` — check tenant backup readiness
- `goclaw backup s3 [subcommands]` — manage S3 backup integration

### Restore (source: `cmd/restore.go`)
- `goclaw restore system <file>` — restore entire system from .tar.gz archive **DESTRUCTIVE**
- `goclaw restore tenant <file>` — restore tenant from .tar.gz archive **DESTRUCTIVE**

## Verified flags

### `backup system`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--exclude-db` | bool | Skip database from backup |
| `--exclude-files` | bool | Skip files from backup |
| `--file <path>` | string | Output file for auto-download (requires `--wait`) |
| `--wait` | bool | Wait + auto-download (used with `--file`) |

### `backup tenant`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--tenant-id <id>` | string | Tenant ID to backup |
| `--file <path>` | string | Output file for auto-download |
| `--wait` | bool | Wait + auto-download |

### `backup system-preflight` / `backup tenant-preflight`
No flags; read-only checks.

### `restore system`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--confirm <basename>` | string | **REQUIRED** — type archive filename to confirm |
| `--yes` | bool | **REQUIRED** — confirm destructive operation |

### `restore tenant`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--tenant-id <id>` | string | **REQUIRED** — tenant ID to restore |
| `--confirm <tenantid>` | string | **REQUIRED** — type tenant ID to confirm |
| `--yes` | bool | **REQUIRED** — confirm destructive operation |

## JSON output

- ✅ `backup system/tenant` — returns JSON with token (e.g., `{"token": "...",...}`)
- ✅ `backup system-preflight/tenant-preflight` — JSON readiness status
- ⚠️ `backup s3` subcommands — check `--help` per subcommand
- ⚠️ `restore system/tenant` — success text only (check exit code 0)

## Destructive ops

| Command | Why dangerous |
| --- | --- |
| `restore system <file>` | Overwrites ALL system data (db + files). No undo. |
| `restore tenant <file>` | Overwrites ALL tenant data (agents, sessions, memory). No undo. |

## Common patterns

### Example 1: check backup readiness
```bash
goclaw backup system-preflight --output json
goclaw backup tenant-preflight --tenant-id acme --output json
```

### Example 2: create system backup with auto-download
```bash
goclaw backup system --wait -o /tmp/backup.tar.gz --output json
# response includes token for resuming download separately
```

### Example 3: create tenant backup
```bash
goclaw backup tenant --tenant-id acme --wait -o /tmp/tenant-backup.tar.gz
```

### Example 4: restore system (VERY DESTRUCTIVE)
```bash
# Claude: confirm with user FIRST — this overwrites everything
goclaw restore system /tmp/backup-20260601.tar.gz \
  --confirm=backup-20260601.tar.gz \
  --yes \
  --output json
```

### Example 5: restore tenant (DESTRUCTIVE)
```bash
# Claude: confirm with user FIRST
goclaw restore tenant /tmp/tenant-backup.tar.gz \
  --tenant-id acme \
  --confirm=acme \
  --yes
```

## Edge cases & gotchas

- **Download token TTL:** returned token expires after ~24h. Re-create backup if token expires.
- **Backup file size:** large systems (big databases, many files) may take minutes and produce multi-GB archives.
- **Restore requires literal file paths**, not stdin. Archive must exist locally; decompress if needed.
- **Restore `--confirm` flag is TYPO-PROOF:** you must type the exact archive basename (for system) or tenant ID (for tenant). Mismatch = rejected.
- **All connections must stop before restore:** agents, webhooks, WS streams must be idle. Server won't let you restore with active connections.
- **Restore is logged:** full audit trail server-side (who, when, what restored).
- **S3 integration:** separate `backup s3` commands for configuring automated S3 offloads. See `goclaw backup s3 --help`.

## Cross-refs

- Admin role check: [auth-and-config.md](auth-and-config.md) — `whoami`
- Tenant management: [admin-system.md](admin-system.md) — `tenants list/delete`
