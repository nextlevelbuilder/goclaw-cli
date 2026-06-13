# CLI profiles (connection config)

## When to use

User wants to create, switch, or manage CLI connection profiles (different servers, tenants, or credentials per profile).

## Commands in scope

- `goclaw profile list` — list all configured profiles (source: `cmd/profile.go`)
- `goclaw profile current` — print active profile name
- `goclaw profile create` — create new profile
- `goclaw profile use <name>` — switch active profile (alias for `auth use-context`)
- `goclaw profile delete <name>` — delete profile

## Verified flags

### `profile create`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--name <n>` | string | Profile name |
| `--server <url>` | string | GoClaw server URL |
| `--token <t>` | string | Auth token (optional; prompt if omitted) |

### `profile use` / `profile delete`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<name>` | string | Profile name |

## JSON output

- ✅ `profile list/current` — JSON
- ⚠️ `profile create/use/delete` — success text only

## Destructive ops

| Command | Confirm |
| --- | --- |
| `profile delete` | YES (removes local config) |

## Common patterns

### Example 1: list profiles
```bash
goclaw profile list --output json
```

### Example 2: create new profile
```bash
goclaw profile create \
  --name staging \
  --server https://staging-goclaw.example.com \
  --token sk-... \
  --output json
```

### Example 3: switch to profile
```bash
goclaw profile use staging
# subsequent commands use staging server
```

### Example 4: check active profile
```bash
goclaw profile current
```

### Example 5: delete profile
```bash
goclaw profile delete staging --yes
```

## Edge cases & gotchas

- **vs auth contexts:** `profile` and `auth use-context` are equivalent. Both switch active config.
- **Token storage:** stored in credential store (OS keychain or ~/.goclaw/config.yaml). Never passed in shell history.
- **Default profile:** first profile created is active. Switch explicitly if needed.
- **Server URL:** must be valid HTTPS endpoint. TLS cert must be valid (use `--insecure` global flag to skip, dev only).

## Cross-refs

- Auth + login: [auth-and-config.md](auth-and-config.md)
- Credentials store: [credentials.md](credentials.md)
