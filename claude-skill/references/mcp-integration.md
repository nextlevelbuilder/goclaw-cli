# MCP integration

## When to use

User wants to manage Model Context Protocol (MCP) servers the agents connect to — add/remove MCP servers, grant/revoke agent access, review incoming connection requests, or trigger reconnections.

## Commands in scope

- `goclaw mcp servers list/get/create/update/delete` — MCP server CRUD (source: `cmd/mcp.go`)
- `goclaw mcp servers test <id>` — sync test call
- `goclaw mcp grants list/grant/revoke` — per-agent access
- `goclaw mcp requests list/approve/deny` — inbound connection requests
- `goclaw mcp reconnect <id>` — **async** trigger reconnection (source: `mcp_reconnect.go`)

## Verified flags

### `mcp servers create` (verified from cmd/mcp.go)
| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--name <n>` | string | — | Server name |
| `--transport <t>` | string | `stdio` | `stdio`/`sse`/`streamable-http` |
| `--command <cmd>` | string | — | Launch command (stdio) |
| `--args <csv>` | string slice | — | Command args (repeatable) |
| `--url <url>` | string | — | Endpoint (sse/http) |
| `--prefix <p>` | string | — | Tool prefix |
| `--timeout <s>` | int | 60 | Timeout seconds |

**Note:** env vars for stdio MCP set via server-side config, not CLI flag.

### `mcp grants grant/revoke`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--agent <id>` | string | Agent to grant/revoke |
| `--server <id>` | string | MCP server |

### `mcp requests approve/deny`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--reason <r>` | string | (deny only) reason |

## JSON output

- ✅ `mcp servers list/get/test`, `mcp grants list`, `mcp requests list` — JSON
- ⚠️ `mcp servers create/update/delete`, `grants grant/revoke`, `requests approve/deny`, `mcp reconnect` — success text

## Destructive ops

| Command | Confirm |
| --- | --- |
| `mcp servers delete` | YES (agents lose tools) |
| `mcp grants revoke` | YES |
| `mcp requests deny` | YES |

## Common patterns

### Example 1: register stdio MCP server
```bash
goclaw mcp servers create \
  --name github \
  --transport stdio \
  --command "npx" \
  --args "@modelcontextprotocol/server-github" \
  --output json
# env vars (like GITHUB_TOKEN) configure via dashboard/server config after create
```

### Example 2: register HTTP MCP server
```bash
goclaw mcp servers create \
  --name remote-tools \
  --transport streamable-http \
  --url "https://mcp.example.com/sse" \
  --output json
```

### Example 3: test connectivity
```bash
goclaw mcp servers test <server-id> --output json
# → {"ok": true, "tools_count": 12, ...}
```

### Example 4: grant MCP server to agent
```bash
goclaw mcp grants grant --agent <agent-id> --server <server-id>
goclaw mcp grants list --agent <agent-id> --output json
```

### Example 5: approve inbound connection request
```bash
goclaw mcp requests list --output json
goclaw mcp requests approve <request-id>
```

### Example 6: force reconnect
```bash
# After config change or server restart
goclaw mcp reconnect <server-id>
# async — no return value; check with test
goclaw mcp servers test <server-id> --output json
```

## Edge cases & gotchas

- **`mcp reconnect` is async.** Fire-and-forget. Check health via `test`.
- **`mcp servers test` is sync.** Returns tool inventory snapshot.
- **Transport mix:** stdio runs subprocess on server, streamable-http/sse connects out to URL. Security implications differ.
- **Env vars for stdio MCP:** configured server-side (dashboard or `system-config`), NOT via CLI flag.
- **Request approval:** external MCP clients request to connect; admin approves via `mcp requests approve`. Check list regularly.
- **Grant scope:** MCP server grants are per-agent. No tenant-level grant.
- **Reconnect on network blip:** servers with transient failures auto-reconnect; manual reconnect for config changes.

## Cross-refs

- Built-in tools (different mechanism): [providers-skills-tools.md](providers-skills-tools.md) — `tools builtin`
- Invoke tools (built-in or MCP): [exec-workflow.md](exec-workflow.md)
