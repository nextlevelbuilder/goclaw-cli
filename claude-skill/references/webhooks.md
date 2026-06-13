# Webhooks (inbound)

## When to use

User wants to create, manage, or configure **inbound webhooks** — HTTP callbacks triggered by specific events (LLM runs, messages) that POST to external endpoints. Different from `hooks` (internal event routing).

## Commands in scope

- `goclaw webhooks list` — list all webhooks (source: `cmd/webhooks.go`)
- `goclaw webhooks get <id>` — get webhook details (including masked secret)
- `goclaw webhooks create` — create a new webhook
- `goclaw webhooks update <id>` — update webhook fields (URL, scopes, etc.)
- `goclaw webhooks delete <id>` — delete a webhook
- `goclaw webhooks rotate <id>` — rotate webhook signing secret

## Verified flags

### `webhooks create/update`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--name <n>` | string | Webhook display name |
| `--kind <k>` | string | `llm` or `message` |
| `--agent-id <id>` | string | Agent UUID (for LLM webhooks) |
| `--channel-id <id>` | string | Channel UUID (for message webhooks) |
| `--body <json>` | string | Webhook payload template (JSON object) |
| `--scopes <s>` | CSV | Comma-separated webhook scopes (e.g., `run.started,run.completed`) |
| `--require-hmac` | bool | Require HMAC-SHA256 signatures on POST |
| `--ip-allowlist <ips>` | CSV | Comma-separated allowed IP/CIDR (e.g., `10.0.0.0/8,203.0.113.5`) |
| `--localhost-only` | bool | Restrict to localhost/127.0.0.1 callers |
| `--rate-limit-per-min <n>` | int | Per-webhook rate limit (invokes per minute) |

### `webhooks delete`
| Flag | Type | Purpose |
| --- | --- | --- |
| (none — use `--yes` to skip confirm) | — | — |

### `webhooks rotate`
| Flag | Type | Purpose |
| --- | --- | --- |
| (none) | — | Rotates secret; old secret invalidates |

## JSON output

- ✅ `webhooks list/get` — JSON
- ⚠️ `webhooks create/update/delete/rotate` — success text only

## Destructive ops

| Command | Confirm |
| --- | --- |
| `webhooks delete` | YES (webhook stops receiving events) |
| `webhooks rotate` | YES (old secret invalidates immediately; callers must update) |

## Common patterns

### Example 1: list webhooks
```bash
goclaw webhooks list --output json
```

### Example 2: create LLM webhook
```bash
goclaw webhooks create \
  --name "run-logger" \
  --kind llm \
  --agent-id abc123 \
  --scopes "run.started,run.completed" \
  --body '{"agent_id":"${{AGENT_ID}}","run_id":"${{RUN_ID}}"}' \
  --require-hmac \
  --output json
```

### Example 3: create message webhook with IP allowlist
```bash
goclaw webhooks create \
  --name "slack-notifier" \
  --kind message \
  --channel-id chan456 \
  --ip-allowlist "10.0.0.0/8,192.168.0.0/16" \
  --rate-limit-per-min 10
```

### Example 4: update webhook
```bash
goclaw webhooks update <webhook-id> \
  --scopes "run.started,run.failed,run.completed"
```

### Example 5: rotate secret
```bash
# Claude: confirm with user — old secret becomes invalid
goclaw webhooks rotate <webhook-id> --yes
```

## Edge cases & gotchas

- **Webhook secret:** returned only once on create. `webhooks get` shows masked prefix. Save it immediately or rotate.
- **Rate limiting:** per-webhook counter. Bursts exceeding limit get 429 responses; retry logic is caller's responsibility.
- **HMAC signature:** if `--require-hmac`, POST header includes `X-Webhook-Signature: sha256=...`. Caller must verify.
- **IP allowlist vs localhost-only:** mutually exclusive. Use localhost-only for testing; IP allowlist for prod.
- **Scopes vary by kind:** LLM webhooks have `run.*` scopes; message webhooks have `message.*` scopes. See `--help` for current scope list.
- **Payload templates:** `${{AGENT_ID}}`, `${{RUN_ID}}`, etc. are placeholders server substitutes at webhook trigger time.
- **Webhook retry:** if POST fails, server retries up to N times (exponential backoff). Check logs if webhook isn't firing.

## Cross-refs

- Event hook routing (internal): [hooks.md](hooks.md)
- LLM traces + run details: [monitoring-ops.md](monitoring-ops.md)
- Channel management: [channels-messaging.md](channels-messaging.md)
