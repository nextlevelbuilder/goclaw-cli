# Chat — history & replay

## When to use

User wants to view chat message history, replay conversations for testing/audit, or inspect past exchanges.

## Commands in scope

- `goclaw chat history <session-id>` — list message history (source: `cmd/chat_history.go`)
- `goclaw chat replay <session-id>` — replay conversation (source: `cmd/chat_replay.go`)

## Verified flags

### `chat history`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<session-id>` | string | Session ID |
| `--limit <n>` | int | Max messages (default 50) |
| `--offset <n>` | int | Pagination offset |

### `chat replay`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<session-id>` | string | Session ID |
| `--speed <n>` | float | Replay speed (1.0 = real-time) |

## JSON output

- ✅ `chat history`, `chat replay` — JSON

## Destructive ops

None — read-only operations.

## Common patterns

### View message history
```bash
goclaw chat history <session-id> --output json
# → list of messages with timestamps, tokens, etc.
```

### View history with pagination
```bash
goclaw chat history <session-id> --limit 20 --offset 40 --output json
```

### Replay conversation at real-time speed
```bash
goclaw chat replay <session-id> --speed 1.0 --output json
```

### Replay conversation at 2x speed
```bash
goclaw chat replay <session-id> --speed 2.0 --output json
# accelerated for testing
```

## Edge cases & gotchas

- **History pagination:** default limit prevents huge JSON. Use `--limit` + `--offset` for large histories.
- **Replay speed:** 1.0 = real-time (respects original timing); >1 = accelerated (distorts timing); <1 = slowed down.
- **Token accounting:** history includes `input_tokens`/`output_tokens` per message.
- **Message order:** chronological (oldest first).

## Cross-refs

- Basic chat: [chat-basic.md](chat-basic.md)
- Session operations: [chat-sessions.md](chat-sessions.md)
- Message injection: [chat-injection.md](chat-injection.md)
