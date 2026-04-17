# Chat & sessions

## When to use

User wants to send a message to an agent, inject input mid-run, abort a running agent, or inspect/clean up chat sessions.

## Commands in scope

- `goclaw chat <agent> -m "<msg>" --no-stream` — single-shot (source: `cmd/chat.go:44-47`)
- `goclaw chat inject <agent> --text "..."` — inject into running session
- `goclaw chat status <agent>` — session run status (WS call `chat.session.status`)
- `goclaw chat abort <agent>` — abort running agent (WS call `chat.abort`)
- `goclaw sessions list` / `get` / `preview <key>` / `delete` / `reset` / `label` — session CRUD
- `goclaw chat <agent>` *without* `-m` → **interactive REPL, skill REFUSES**

## Verified flags

### `chat <agent>`
| Flag | Type | Purpose |
| --- | --- | --- |
| `-m, --message` | string | Single-shot message (source: `cmd/chat.go:280`) |
| `--session <key>` | string | Continue existing session |
| `--no-stream` | bool | Wait for full response instead of streaming |

### `chat inject`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--text` | string (required) | Text to inject |
| `--session` | string | Session key |

### `sessions list`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--agent` | string | Filter by agent ID |
| `--user` | string | Filter by user ID |
| `--limit` | int | Max results |

### `sessions label`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--label` | string (required) | New label |

## JSON output

- ✅ `chat -m ... --output json` — NDJSON event stream (`chat.go:92-98`)
- ✅ `chat -m ... --no-stream --output json` — single JSON map
- ✅ `sessions list/preview` — JSON
- ⚠️ `sessions delete/reset/label` — `printer.Success` text only

## Destructive ops

| Command | Why destructive |
| --- | --- |
| `sessions delete` | drops session + messages (tui.Confirm at `sessions.go:80`) |
| `sessions reset` | clears messages, keeps key (tui.Confirm at `sessions.go:101`) |
| `chat abort` | force-stops running agent mid-tool-call — **no `--yes` flag**, still warn user |

## Common patterns

### Example 1: single-shot question
```bash
goclaw chat my-agent -m "Summarize today's activity" --no-stream --output json
```

### Example 2: continue previous session
```bash
goclaw chat my-agent -m "Based on that, what next?" --session sess_abc --output json
```

### Example 3: pipe stdin
```bash
echo "Analyze this log" | goclaw chat my-agent --output json
```

### Example 4: list sessions for agent + preview one
```bash
goclaw sessions list --agent <agent-id> --limit 10 --output json
goclaw sessions preview <session-key> --output json
```

### Example 5: abort runaway agent
```bash
# Always confirm with user first
goclaw chat abort my-agent --session <key>
```

## Edge cases & gotchas

- **Interactive mode detection:** `chat <agent>` with no `-m` and no stdin → TUI REPL with `/exit`, `/abort`, `/sessions`, `/clear` slash commands (`chat.go:148`). Skill MUST always pass `-m` or pipe stdin.
- **Streaming NDJSON:** when `--output json` without `--no-stream`, each event is a JSON line: `{"event":"chunk|tool.call|tool.result|run.completed","data":{...}}`. Parse line-by-line, stop at `run.completed`.
- **`chat abort` is not streaming.** Safe to call. But recovers no work done before abort.
- **Session key vs session ID:** API uses `session_key` (human-readable, stable across resets); list shows both.
- **Token accounting:** `sessions list --output json` includes `input_tokens`/`output_tokens` per session — useful for cost tracing.

## Cross-refs

- Approvals from tool calls: [exec-workflow.md](exec-workflow.md) (canonical home)
- Agent lifecycle: [agents-core.md](agents-core.md)
- Traces per LLM call: [monitoring-ops.md](monitoring-ops.md)
