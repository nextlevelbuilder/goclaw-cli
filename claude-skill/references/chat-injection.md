# Chat — injection & advanced operations

## When to use

User wants to inject messages into a session mid-conversation, check session status/metadata, or follow session in real-time (for testing/debugging).

## Commands in scope

- `goclaw chat inject <session-id>` — inject message into session (source: `cmd/chat_inject.go`)
- `goclaw chat session-status <session-id>` — check session state + metadata (source: `cmd/chat_status.go`)
- `goclaw chat follow <session-id>` — **follow session in real-time, STREAMING, skill REFUSES**
- `goclaw chat abort <session-id>` — force-stop running agent

## Verified flags

### `chat inject`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<session-id>` | string | Session ID |
| `--role <r>` | string | `user` or `assistant` |
| `--content <c>` | string | Message content |

### `chat session-status`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<session-id>` | string | Session ID |

### `chat follow`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<session-id>` | string | Session ID |

### `chat abort`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<session-id>` | string | Session ID |

## JSON output

- ✅ `chat session-status` — JSON
- ⚠️ `chat inject` — success text only
- ❌ `chat follow` — streaming, skill refuses

## Destructive ops

| Command | Confirm |
| --- | --- |
| `chat inject` | Recommended (alters session) |
| `chat abort` | Recommended (stops in-progress work) |

## Common patterns

### Inject user message
```bash
goclaw chat inject <session-id> --role user --content "What's the status?" --output json
```

### Inject assistant message (for testing)
```bash
goclaw chat inject <session-id> --role assistant --content "Status is good" --output json
```

### Check session status
```bash
goclaw chat session-status <session-id> --output json
# → {"state": "active|paused|archived", "message_count": 42, "last_activity": "..."}
```

### Abort running agent
```bash
# Always confirm with user first
goclaw chat abort <session-id>
```

## Edge cases & gotchas

- **Injection:** bypasses normal chat flow. Useful for testing/recovery but affects audit trail. No undo.
- **Role:** `user` = from user perspective; `assistant` = from agent perspective. Both valid for testing.
- **Status:** read-only metadata. Session state: `active` (normal), `paused` (temporarily), `archived` (closed).
- **Abort:** force-stops mid-tool-call. No retry. May leave partial state.
- **Follow:** real-time stream; STREAMING operation. Skill refuses. Suggest user terminal.

## Cross-refs

- Basic chat: [chat-basic.md](chat-basic.md)
- Session operations: [chat-sessions.md](chat-sessions.md)
- Chat history: [chat-history.md](chat-history.md)
