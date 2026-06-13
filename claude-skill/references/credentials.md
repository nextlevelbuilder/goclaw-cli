# Credentials store

## When to use

User wants to manage CLI credentials — list stored credentials, create/delete/rotate credentials, check binary credentials, manage user credential grants, or preset credential options.

## Commands in scope

- `goclaw credentials list` — list all stored credentials (source: `cmd/credentials.go`)
- `goclaw credentials create` — store new credential
- `goclaw credentials delete <id>` — remove credential
- `goclaw credentials update <id>` — update credential fields
- `goclaw credentials check-binary` — verify credential binary integrity
- `goclaw credentials test <id>` — test credential (connectivity check)
- `goclaw credentials user-credentials` — manage per-user credential grants
- `goclaw credentials agent-grants` — view/manage agent credential grants
- `goclaw credentials presets` — list credential preset templates

## Verified flags

### `credentials create`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--name <n>` | string | Credential name |
| `--type <t>` | string | `api-key`, `bearer`, `oauth`, `ssh`, etc. |
| `--value <v>` | string | Credential secret (masked in responses) |
| `--expires-in <d>` | duration | TTL (e.g., `90d`) |

### `credentials delete` / `credentials test`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<id>` | credential ID | — |

### `credentials user-credentials`
Subcommands: `list`, `grant`, `revoke` (verify with `--help`)

### `credentials agent-grants`
Subcommands: `list`, `grant`, `revoke` (verify with `--help`)

## JSON output

- ✅ `credentials list`, `credentials presets`, `credentials user-credentials list`, `credentials agent-grants list` — JSON
- ⚠️ `credentials create/update/delete/test` — success text only
- ✅ `credentials check-binary` — JSON status

## Destructive ops

| Command | Confirm |
| --- | --- |
| `credentials delete` | YES (revokes access) |
| `credentials user-credentials revoke` | YES |
| `credentials agent-grants revoke` | YES |

## Common patterns

### Example 1: list stored credentials
```bash
goclaw credentials list --output json
# → [
#   {"id": "cred-1", "name": "prod-api-key", "type": "api-key", "created_at": "..."},
#   ...
# ]
```

### Example 2: create credential
```bash
goclaw credentials create \
  --name "github-token" \
  --type api-key \
  --value ghp_xxxxxxxxxxxx \
  --expires-in 90d \
  --output json
```

### Example 3: test credential
```bash
goclaw credentials test <cred-id> --output json
# → {"valid": true, "checked_at": "..."}
```

### Example 4: grant credential to user
```bash
goclaw credentials user-credentials grant \
  --credential-id <cred-id> \
  --user-id <user-id> \
  --output json
```

### Example 5: grant credential to agent
```bash
goclaw credentials agent-grants grant \
  --credential-id <cred-id> \
  --agent-id <agent-id> \
  --output json
```

### Example 6: check binary integrity
```bash
goclaw credentials check-binary --output json
# → {"valid": true, "sha256": "..."}
```

### Example 7: list credential presets
```bash
goclaw credentials presets --output json
# → credential template options (GitHub, AWS, etc.)
```

## Edge cases & gotchas

- **Secret masking:** `credentials list` returns ID + metadata only; never the actual secret. `create` response shows secret ONCE.
- **Expiry:** expired credentials fail on use. `test` checks before expiry. Pre-rotate before expiry date.
- **User/agent grants:** separate namespaces. User grants control CLI-level access; agent grants control agent runtime access.
- **Binary check:** verifies CLI binary hasn't been tampered with. Useful in untrusted environments.
- **Presets:** templates for common providers (GitHub, AWS, etc.). Reduces manual entry errors.

## Cross-refs

- API keys (scoped access): [auth-and-config.md](auth-and-config.md) — `api-keys`
- CLI profiles: [profile.md](profile.md)
- Auth + login: [auth-and-config.md](auth-and-config.md)
