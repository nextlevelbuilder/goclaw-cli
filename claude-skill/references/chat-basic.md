# Chat — basic operations

## When to use

User wants to send a single-shot message to an agent, continue a previous chat session, or test agent connectivity.

## Commands in scope

- `goclaw chat <agent> -m "<msg>"` — single-shot message (source: `cmd/chat.go`)
- `goclaw chat <agent> -m "<msg>" --no-stream` — single-shot, wait for full response
- `goclaw chat <agent> -m "<msg>" --session <key>` — continue existing session
- `goclaw chat echo <agent-id>` — connectivity test (source: `cmd/chat_echo.go`)
- `goclaw chat` (no args) — **interactive REPL, skill REFUSES**

## Verified flags

### `chat <agent>`
| Flag | Type | Purpose |
| --- | --- | --- |
| `-m, --message <msg>` | string | Single-shot message (required for non-interactive) |
| `--session <key>` | string | Continue existing session |
| `--no-stream` | bool | Wait for full response (no NDJSON stream) |

### `chat echo`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<agent-id>` | string | Agent to test |
| `--message <m>` | string | Message to send |

## JSON output

- ✅ `chat -m ... --no-stream --output json` — single JSON map
- ✅ `chat -m ... --output json` — NDJSON event stream (parse line-by-line)
- ✅ `chat echo` — JSON
- ❌ `chat` (interactive) — REPL, skill refuses

## Destructive ops

None — read-only operations.

## Common patterns

### Single-shot question (wait for full response)
```bash
goclaw chat my-agent -m "Summarize today's activity" --no-stream --output json
```

### Stream response (NDJSON)
```bash
goclaw chat my-agent -m "Write a poem" --output json
# → parse line-by-line, stop at run.completed event
```

### Continue previous session
```bash
goclaw chat my-agent -m "Based on that, what next?" --session sess_abc --output json
```

### Pipe input from stdin
```bash
echo "Analyze this log" | goclaw chat my-agent --output json
```

### Test connectivity
```bash
goclaw chat echo <agent-id> --message "Hello" --output json
```

## Edge cases & gotchas

- **Interactive mode:** `chat <agent>` with no `-m` and no stdin → TUI REPL with slash commands (`/exit`, `/abort`, `/sessions`, `/clear`). Skill must always pass `-m` or stdin.
- **Streaming NDJSON:** without `--no-stream`, events arrive line-by-line. Parse incrementally; stop at `run.completed` event.
- **Event types:** `chunk` (text), `tool.call`, `tool.result`, `run.completed`. Build response from chunks.
- **Session key vs ID:** API uses `session_key` (human-readable, stable); list shows both. Use key for `--session`.
- **Token accounting:** response includes `input_tokens`/`output_tokens` per message.
- **Echo:** lightweight connectivity test. Does NOT create session or affect agent state.

## Cross-refs

- Session operations: [chat-sessions.md](chat-sessions.md)
- Chat history: [chat-history.md](chat-history.md)
- Message injection: [chat-injection.md](chat-injection.md)
