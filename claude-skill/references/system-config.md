# System configuration & upgrades

## When to use

Admin wants to view/manage server-level configuration (schema, defaults, patches), apply bulk config changes, or upgrade the server.

## Commands in scope

### Configuration (source: `cmd/config.go`)
- `goclaw config schema` — get full config schema (JSON Schema)
- `goclaw config defaults` — list default config values
- `goclaw config get <key>` — read config value
- `goclaw config patch <config.yaml>` — apply partial config change (strategic merge)
- `goclaw config apply <config.yaml>` — apply full config (replace entire structure)

### System Upgrades (source: `cmd/system_upgrade.go`)
- `goclaw system upgrade` — upgrade server to latest version (source: `system_upgrade.go`)

## Verified flags

### `config schema`
No flags; returns full JSON Schema.

### `config defaults`
No flags; returns current defaults.

### `config get`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<key>` | string | Config key path (e.g., `gateway.listen_port`) |

### `config patch` / `config apply`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<file>` | path | YAML/JSON config file (@filepath or stdin) |
| `--dry-run` | bool | Preview changes without applying |

### `system upgrade`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--version <v>` | string | Target version (default: latest) |
| `--dry-run` | bool | Preview upgrade steps |
| `--yes` | bool | Skip confirmation |

## JSON output

- ✅ `config schema/defaults/get` — JSON
- ✅ `config patch/apply` (with `--dry-run`) — JSON preview
- ⚠️ `config patch/apply` (apply) — success text only
- ✅ `system upgrade` (with `--dry-run`) — JSON preview
- ⚠️ `system upgrade` (apply) — success text + progress

## Destructive ops

| Command | Confirm |
| --- | --- |
| `config patch` | Recommended (modifies config) |
| `config apply` | **YES** (full replacement) |
| `system upgrade` | **YES** (server restart) |

## Common patterns

### Example 1: get config schema
```bash
goclaw config schema --output json | jq '.properties | keys'
```

### Example 2: read single config value
```bash
goclaw config get gateway.listen_port --output json
```

### Example 3: preview config patch
```bash
cat > /tmp/patch.yaml << 'EOF'
gateway:
  max_connections: 1000
EOF
goclaw config patch /tmp/patch.yaml --dry-run --output json
```

### Example 4: apply config patch (merge)
```bash
goclaw config patch /tmp/patch.yaml --output json
# merges patch into current config
```

### Example 5: apply full config (replace)
```bash
# Get current, edit, apply
goclaw config get "/" --output json > current.json
# edit current.json
goclaw config apply current.json --dry-run --output json
goclaw config apply current.json --yes
```

### Example 6: preview server upgrade
```bash
goclaw system upgrade --version 1.2.0 --dry-run --output json
# → {"steps": [...], "duration_minutes": 5}
```

### Example 7: perform server upgrade
```bash
goclaw system upgrade --version 1.2.0 --yes
# → server restarts automatically after upgrade
```

## Edge cases & gotchas

- **Config schema:** shows structure + validation rules. Essential reading before patching.
- **Patch vs apply:** patch = strategic merge (only specified keys updated); apply = full replacement (unspecified keys revert to file content).
- **Dry-run:** previews what *would* happen without writing. Useful for validation.
- **YAML path keys:** use `.` notation for nesting (e.g., `gateway.listen_port`).
- **Server upgrade:** automatic restart. All agents/sessions gracefully paused. Typically 2-5 min downtime.
- **Backup before upgrade:** server auto-backs up config before upgrade. But manual backup recommended.

## Cross-refs

- Per-tenant config: [admin-system.md](admin-system.md) — `system-config`
- Server health/status: [monitoring-ops.md](monitoring-ops.md) — `status`
