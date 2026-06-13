# Event hooks (internal routing)

## When to use

User wants to route internal system events (agent starts, chat completes, etc.) to handlers — scripts, tool invocations, or external destinations. Different from `webhooks` (HTTP callbacks). Use hooks for event-driven automation within GoClaw itself.

## Commands in scope

- `goclaw hooks list` — list all hooks (source: `cmd/hooks.go`)
- `goclaw hooks get <id>` — get hook details
- `goclaw hooks create` — create hook from JSON config
- `goclaw hooks update <id>` — patch hook fields
- `goclaw hooks delete <id>` — delete a hook (requires `--yes`)
- `goclaw hooks toggle <id>` — enable/disable hook without deleting
- `goclaw hooks test <id>` — dry-run hook against sample event (no side-effects)
- `goclaw hooks history <id>` — show execution history + success/failure logs

## Verified flags

### `hooks create`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--config <json>` | string | Hook config JSON (`@filepath` or literal) |

**Required config fields (JSON):**
- `event` (string) — event type (e.g., `agent.started`, `chat.completed`)
- `scope` (string) — scope (e.g., `system`, `tenant:abc`, `agent:xyz`)
- `handler_type` (string) — handler type (`script`, `tool`, `log`, etc.)
- `name` (string) — display name

**Optional config fields:**
- `matcher` (object) — filter condition (e.g., `{"agent_type": "coding"}`)
- `if_expr` (string) — CEL expression (if handler supports it)
- `timeout_ms` (int) — execution timeout
- `on_timeout` (string) — action if timeout (`log`, `skip`, `alert`)
- `priority` (int) — execution order
- `agent_ids` (array) — limit to specific agents
- `config` (object) — handler-specific config
- `metadata` (object) — custom metadata

### `hooks update`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--config <json>` | string | Patch config fields (partial) |

### `hooks test`
| Flag | Type | Purpose |
| --- | --- | --- |
| (none) | — | Dry-run against sample event (no side-effects) |

### `hooks toggle`
| Flag | Type | Purpose |
| --- | --- | --- |
| (none) | — | Flips enabled/disabled state |

## JSON output

- ✅ `hooks list/get/history` — JSON
- ⚠️ `hooks test` — JSON response (success/failure, handler output)
- ⚠️ `hooks create/update/delete/toggle` — success text only

## Destructive ops

| Command | Confirm |
| --- | --- |
| `hooks delete` | YES (hook stops processing events) |

## Common patterns

### Example 1: list all hooks
```bash
goclaw hooks list --output json
```

### Example 2: create hook from JSON file
```bash
cat > /tmp/hook.json << 'EOF'
{
  "name": "log-agent-starts",
  "event": "agent.started",
  "scope": "system",
  "handler_type": "log",
  "config": {"level": "info"}
}
EOF
goclaw hooks create --config=@/tmp/hook.json --output json
```

### Example 3: create hook with conditional matcher
```bash
goclaw hooks create --config=@- << 'EOF'
{
  "name": "alert-on-chat-error",
  "event": "chat.completed",
  "scope": "tenant:acme",
  "handler_type": "script",
  "matcher": {"status": "error"},
  "config": {
    "script": "curl -X POST https://alerting.example.com/hook -d @-"
  }
}
EOF
```

### Example 4: dry-run hook before enabling
```bash
goclaw hooks test <hook-id> --output json
# → {"success": true, "output": "..."}
```

### Example 5: view hook execution history
```bash
goclaw hooks history <hook-id> --output json
# → list of recent executions + results
```

### Example 6: toggle hook without deleting
```bash
goclaw hooks toggle <hook-id>
# switch between enabled/disabled
```

## Edge cases & gotchas

- **Event types:** varies by server. Run `goclaw hooks list --help` or check server docs for available event types.
- **Scope format:** `system` (all tenants), `tenant:abc` (specific tenant), `agent:xyz` (specific agent). Scopes are hierarchical.
- **Handler types:** `log`, `script`, `tool`, `webhook` (to external). Each type has different `config` schema.
- **CEL expressions:** `if_expr` allows conditional evaluation. Syntax varies by handler.
- **Timeouts:** if handler runs > `timeout_ms`, `on_timeout` action kicks in (log warning, skip, or escalate).
- **Hook execution order:** controlled by `priority` (lower = earlier). Hooks at same priority run in creation order.
- **Dry-run (`test`):** uses sample event, not real data. Results are indicative, not guaranteed.
- **Hook history:** retention server-dependent. Check server config for retention window.

## Cross-refs

- Inbound webhooks (HTTP callbacks): [webhooks.md](webhooks.md)
- Event-driven tool invocation: [exec-workflow.md](exec-workflow.md)
- System configuration: [admin-system.md](admin-system.md)
