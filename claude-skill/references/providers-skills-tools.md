# Providers, skills & tools registry

## When to use

User wants to manage LLM providers (register API keys for OpenAI/Anthropic/Gemini/etc.), publish/install/grant skills, enable/disable built-in tools, or inspect runtime packages. For *invoking* tools (including `exec`), see [exec-workflow.md](exec-workflow.md).

## Commands in scope

### Providers (source: `cmd/providers.go`, `providers_crud.go`)
- `goclaw providers list/get/create/update/delete` — provider CRUD
- `goclaw providers models <id>` — list models for a provider
- `goclaw providers verify-embedding <id>` — test embedding endpoint

### Skills (source: `cmd/skills.go`, `skills_*.go`)
- `goclaw skills list/get/update/delete/toggle` — skill CRUD
- `goclaw skills upload <path>` — upload skill bundle (multipart)
- `goclaw skills publish/unpublish <id>` — visibility toggle
- `goclaw skills versions <id>` — version history
- `goclaw skills grant/revoke` — per-agent skill access (source: `skills_grants.go`)
- `goclaw skills tenant-config set/delete` — tenant-level skill config
- `goclaw skills files list/get/create/delete` — skill content files

### Tools registry (source: `cmd/tools.go`)
- `goclaw tools builtin list/get/update <name>` — list/configure builtins
- `goclaw tools builtin tenant-config set/delete <name>` — per-tenant tool enable
- `goclaw tools invoke <name>` — see [exec-workflow.md](exec-workflow.md)

### Packages (source: `cmd/packages.go`)
- `goclaw packages list` — runtime packages (read-only)

## Verified flags

### `providers create/update` (verify with `goclaw providers create --help`)
Common fields passed via `buildBody` — names depend on CLI release. Typical:
- `--name <n>` — provider slug
- `--display-name <d>` — UI label
- `--api-key <k>` — API key
- `--type <t>` — provider type (openai/anthropic/gemini/...)

Run `goclaw providers create --help` for current flag list; endpoint override flag name varies.

### `skills list`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--search <q>` | string | Full-text search |

### `skills upload`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--name <n>` | string | Override skill name |
| `--visibility <v>` | string | `private`/`shared` (default `private`) |

### `skills update`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--name <n>` | string | Rename |
| `--visibility <v>` | string | Change visibility |

### `tools builtin update <name>`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--enabled <bool>` | bool | Enable/disable tool globally |

### `tools builtin tenant-config set <name>`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--enabled <bool>` | bool | Per-tenant enable override |

## JSON output

- ✅ `providers list/get/models`, `skills list/get/versions`, `tools builtin list/get`, `packages list` — JSON
- ⚠️ `providers create/update/delete/verify-embedding`, `skills upload/update/delete/toggle/publish/unpublish/grant/revoke`, `tools builtin update/tenant-config *` — success text

## Destructive ops

| Command | Confirm |
| --- | --- |
| `providers delete` | YES (agents using provider break) |
| `skills delete` | YES (agents lose skill) |
| `skills unpublish` | YES (hides from catalog) |
| `skills revoke` | YES |
| `skills tenant-config delete` | YES |
| `tools builtin update --enabled=false` | YES (agents lose tool) |
| `tools builtin tenant-config delete` | YES |

## Common patterns

### Example 1: register OpenAI provider
```bash
# Always --help first for current flag names
goclaw providers create --help
goclaw providers create \
  --name openai \
  --display-name "OpenAI" \
  --type openai \
  --api-key sk-... \
  --output json
```

### Example 2: verify provider works
```bash
goclaw providers verify-embedding <provider-id> --output json
# → {"ok": true, "latency_ms": 120, ...}
```

### Example 3: list + grant skill to agent
```bash
goclaw skills list --search "docs" --output json
# verify exact grant flags (may be --skill-id/--agent or similar):
goclaw skills grant --help
```

### Example 4: disable dangerous builtin tool globally
```bash
# Claude: confirm with user first — disabling exec breaks agent workflows
goclaw tools builtin update exec --enabled=false
```

### Example 5: tenant-specific tool override
```bash
# Disable web_search for tenant "acme" only
goclaw --tenant-id acme tools builtin tenant-config set web_search --enabled=false
```

## Edge cases & gotchas

- **Skills upload** is multipart form; path can be dir or .zip. Server unpacks. File size limits set server-side — check server docs.
- **`skills toggle`** flips enabled-state but doesn't distinguish from `update --enabled` — both work.
- **Providers `api-key`** stored encrypted server-side. `providers get` masks the key in response.
- **Tool `enabled=false` cascades:** agents using the tool get "tool not found" on invocation. Coordinate with agent owners before disabling.
- **Tenant-config vs global update:** `tools builtin update` affects all tenants; `tenant-config set` overrides for one tenant. Use tenant-config when possible.
- **Packages is read-only.** No install/uninstall via CLI — managed server-side.
- **Skills versioning:** `skills versions <id>` lists all published versions; agents pin a specific version.

## Cross-refs

- Invoke tools (exec, etc.): [exec-workflow.md](exec-workflow.md)
- Agent-scoped skill grants: [agents-core.md](agents-core.md)
- MCP tools (separate concept): [mcp-integration.md](mcp-integration.md)
