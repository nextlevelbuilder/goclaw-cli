# Auth & config

## When to use

User wants to log in, switch profile/tenant, manage credentials, or configure the server. **Required prerequisite** for any other skill operation.

## Commands in scope

- `goclaw auth login [--pair] [--token ...]` — authenticate (source: `cmd/auth.go:20-31`)
- `goclaw auth logout` — remove stored token for active profile
- `goclaw auth whoami` — confirm auth + server reachability
- `goclaw auth use-context <profile>` — switch active profile
- `goclaw auth list-contexts` — list all profiles
- `goclaw auth login --pair` — device pairing (**streaming — skill REFUSES**)
- `goclaw config get/set` — server configuration KV
- `goclaw config permissions list/update/grant/revoke` — per-config ACL (WS)
- `goclaw credentials list/create/delete/rotate` — CLI credential store
- `goclaw api-keys list/create/reveal/revoke/extend` — scoped API keys

## Global flags (from `cmd/root.go`)

| Flag | Env var | Purpose |
| --- | --- | --- |
| `--server <url>` | `GOCLAW_SERVER` | Gateway URL |
| `--token <t>` | `GOCLAW_TOKEN` | Bearer token (prefer login flow) |
| `--output json\|yaml\|table` | — | Always `json` in skill |
| `--yes` | — | Bypass destructive confirm |
| `--insecure` | — | Skip TLS verify (dev only) |
| `--profile <name>` | — | Select profile (default: `default`) |
| `--tenant-id <id>` | — | Multi-tenant scope |
| `--verbose` | — | Debug logs |

## Verified flags (per subcommand)

### `auth login`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--pair` | bool | Device pairing (streaming — refuse) |
| `--token` | global | Pass token directly |
| `--profile` | global | Name for profile |

### `config set`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--key <k>` | string | Config key |
| `--value <v>` | string | New value |

### `api-keys create`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--name <n>` | string | Label |
| `--scopes <s>` | CSV | Permitted scopes |
| `--expires-in <d>` | duration | TTL e.g. `30d` |

## JSON output

- ✅ `auth list-contexts` — JSON list with `--output json`
- ⚠️ `auth login/logout` — `printer.Success` text only; check exit code
- ⚠️ `auth whoami` — table-focused, JSON may be limited
- ✅ `config get` — JSON map
- ⚠️ `config set/permissions grant|revoke` — success text only
- ✅ `credentials list` — JSON list
- ✅ `api-keys list/create` — JSON (create shows raw key ONCE)

## Destructive ops

| Command | Why destructive |
| --- | --- |
| `auth logout` | removes credential |
| `config set` | modifies server state |
| `credentials delete` | permanent |
| `credentials rotate` | old key invalidated |
| `api-keys revoke` | permanent |
| `config permissions revoke` | revokes access |

## Common patterns

### Example 1: check auth before operations
```bash
goclaw auth whoami --output json
# non-zero exit → tell user to run `goclaw auth login` manually
```

### Example 2: one-shot tenant override
```bash
goclaw --tenant-id acme-corp agents list --output json
```

### Example 3: switch profile
```bash
goclaw auth list-contexts --output json
goclaw auth use-context staging
```

### Example 4: scoped API key for CI
```bash
goclaw api-keys create --name "ci-runner" --scopes "agents:read,sessions:read" --expires-in 90d --output json
# output includes raw key ONCE — user must save it immediately
```

## Edge cases & gotchas

- **Token expiry:** commands exit non-zero — suggest `goclaw auth login`, do NOT retry-loop.
- **`auth pair`:** polls 60× × 2s (120s total) = hits Bash timeout. Skill REFUSES; user runs in own shell.
- **Credentials:** stored in `~/.goclaw/config.yaml` + OS keychain. Never suggest pasting token in chat.
- **Raw API key:** `api-keys create` reveals key ONCE. `api-keys reveal` may only show prefix.
- **Profile precedence:** flags > env > active profile > `default`.

## Cross-refs

- After auth: [agents-core.md](agents-core.md), [exec-workflow.md](exec-workflow.md)
- Multi-tenant admin: [admin-system.md](admin-system.md)
