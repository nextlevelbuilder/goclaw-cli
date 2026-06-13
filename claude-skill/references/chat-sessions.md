# Chat & sessions

## When to use

User wants to send a message to an agent, inject input mid-run, abort a running agent, or inspect/clean up chat sessions.

## Commands in scope

- `goclaw sessions list` — list chat sessions (source: `cmd/sessions.go`)
- `goclaw sessions get <id>` — get session details (alias for preview)
- `goclaw sessions preview <id>` — preview session transcript
- `goclaw sessions delete <id>` — delete session
- `goclaw sessions reset <id>` — clear messages, keep session
- `goclaw sessions label <id>` — set/change session label
- `goclaw sessions branch <id>` — branch conversation from point
- `goclaw sessions compact` — batch delete old sessions

## Verified flags

### `sessions list`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--agent <id>` | string | Filter by agent ID |
| `--user <id>` | string | Filter by user ID |
| `--limit <n>` | int | Max results (default 20) |
| `--offset <n>` | int | Pagination offset |

### `sessions preview/delete/reset/label`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<id>` | session ID | Session identifier |
| (label only) `--label <l>` | string | New label name |

### `sessions branch`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<id>` | session ID | Session to branch from |
| `--from-message <n>` | int | Branch point (message index) |
| `--label <l>` | string | Label for new branch |

### `sessions compact`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--older-than <d>` | duration | Delete sessions older than (e.g., `30d`) |

## JSON output

- ✅ `sessions list/preview/get` — JSON
- ⚠️ `sessions delete/reset/label/branch/compact` — success text only

## Destructive ops

| Command | Confirm |
| --- | --- |
| `sessions delete` | YES (permanent) |
| `sessions reset` | YES (clears history) |
| `sessions compact --older-than <d>` | YES (batch delete) |

## Common patterns

### List sessions for agent
```bash
goclaw sessions list --agent <agent-id> --limit 10 --output json
```

### Preview session transcript
```bash
goclaw sessions preview <session-id> --limit 50 --output json
```

### Label session for organization
```bash
goclaw sessions label <session-id> --label "Bug investigation - Q2 2026"
```

### Branch conversation from point
```bash
goclaw sessions branch <session-id> --from-message 15 --label "Alternative path" --output json
# → new session ID (child of original)
```

### Reset session (clear history)
```bash
goclaw sessions reset <session-id> --yes
```

### Delete session
```bash
goclaw sessions delete <session-id> --yes
```

### Batch delete old sessions
```bash
goclaw sessions compact --older-than 90d --yes
# deletes all sessions older than 90 days
```

## Edge cases & gotchas

- **Session key vs ID:** API uses `session_key` (human-readable, stable); list shows both. Use key for referencing.
- **Branching:** creates new session with copied history up to branch point. Original unchanged.
- **Reset:** clears all messages but preserves agent/user bindings and session key.
- **Compact:** batch deletes. `--older-than` uses relative durations (`30d`, `90d`, `6mo`). No undo.
- **Label:** purely metadata for humans; doesn't affect agent behavior.
- **Token accounting:** session responses include `input_tokens`/`output_tokens` per message.
- **Message pagination:** use `--limit` + `--offset` for large histories to avoid huge JSON.

## Cross-refs

- Basic chat operations: [chat-basic.md](chat-basic.md)
- Chat history & replay: [chat-history.md](chat-history.md)
- Message injection: [chat-injection.md](chat-injection.md)
- Agent lifecycle: [agents-crud.md](agents-crud.md)
