# Exec workflow (hero use case)

## When to use

User wants to run a shell command on a GoClaw server, invoke any built-in tool directly, or approve/deny an agent's pending execution. This is the #1 reason to use the skill.

## Commands in scope

- `goclaw tools invoke <tool-name> --param <k>=<v>` — invoke any built-in tool (source: `cmd/tools.go:111-144`)
- `goclaw tools builtin list` — discover available tool names (source: `cmd/tools.go:18-32`)
- `goclaw approvals list` — pending execution approvals (WS call: `exec.approval.list`, source: `cmd/admin.go:15-41`)
- `goclaw approvals approve <id>` — approve pending exec (WS call: `exec.approval.approve`)
- `goclaw approvals deny <id> --reason "..."` — deny (WS call: `exec.approval.deny`)

## Verified tool: `exec`

**Server registration** (source: `goclaw/cmd/gateway_builtin_tools.go:24`):
```go
{Name: "exec", DisplayName: "Execute Command", Category: "runtime", Enabled: true}
```

**Parameter schema** (source: `goclaw/internal/tools/shell.go:114-128`):

| Param | Type | Required | Purpose |
| --- | --- | --- | --- |
| `command` | string | ✅ yes | Shell command to execute |
| `working_dir` | string | no (default workspace root) | CWD for the command |

**Response:** stdout + stderr + exit_code in `*Result` struct.

**Safety rails on server side:**
- NUL byte rejection (exits with "command contains invalid NUL byte")
- Unicode normalization (NFKC + zero-width strip) before deny-pattern matching
- Shell deny patterns (per-agent or default) block dangerous ops
- Package installs trigger approval gate (not auto-executed)

## Verified flags

### `goclaw tools invoke <name>`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--param key=value` | repeatable string | One param per flag |
| `--params '<json>'` | string | Alternative: all params as JSON |
| `--output json` | global | Mandatory per skill convention |

### `goclaw approvals deny <id>`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--reason "..."` | string | Denial reason (optional) |

## JSON output

- ✅ `tools invoke` — `printer.Print(unmarshalMap(data))` at `cmd/tools.go:141`
- ✅ `tools builtin list` — `printer.Print(unmarshalList(data))` at `cmd/tools.go:29`
- ⚠️ `approvals approve/deny` — returns `printer.Success("Execution approved/denied")` text; check exit code instead of parsing JSON
- ✅ `approvals list` — JSON via `unmarshalList`

## Destructive ops

| Command | Why destructive | Confirm required? |
| --- | --- | --- |
| `tools invoke exec --param command="rm ..."` | runs arbitrary shell | **YES** — always confirm the command string with user before invoking |
| `tools invoke exec` with package-install command | server triggers approval gate, but skill should still confirm | YES |
| `approvals deny <id>` | denies agent's pending action | YES if dense consequences |

**Rule:** Always echo back the exact `command` param to user and ask *"run this command on the server? (y/N)"* before invoking.

## Common patterns

### Example 1: innocuous read-only command

```bash
goclaw tools invoke exec --param command="uname -a" --output json
```

Expected shape:
```json
{"stdout": "Linux ...", "stderr": "", "exit_code": 0, "approval_required": false}
```

### Example 2: command with custom working directory

```bash
goclaw tools invoke exec \
  --param command="npm test" \
  --param working_dir="/workspace/my-app" \
  --output json
```

### Example 3: command that triggers approval gate

```bash
# First attempt — server may return approval_required=true or an approval_id
goclaw tools invoke exec --param command="apt-get install curl" --output json

# Check pending approvals
goclaw approvals list --output json

# Approve the specific request
goclaw approvals approve <approval-id>

# Re-run if needed
goclaw tools invoke exec --param command="apt-get install curl" --output json
```

### Example 4: discover available tools before invoking

```bash
goclaw tools builtin list --output json
# returns: exec, read_file, write_file, web_search, web_fetch, memory_search, ...
```

### Example 5: invoke non-exec builtin tool

```bash
goclaw tools invoke memory_search --param query="Q3 roadmap" --output json
goclaw tools invoke web_fetch --param url="https://example.com" --output json
```

## Edge cases & gotchas

- **Streaming approvals:** `approvals watch` (if present) uses WS subscribe — NOT Bash-friendly. Use `approvals list --output json` polling instead.
- **NUL bytes in command:** server rejects; Claude should reject `\x00` before calling.
- **Approval ID lifetime:** approvals expire; poll quickly after an exec returns `approval_required`.
- **Params format:** both `--param k=v` and `--params '<json>'` work. For values with `=` or spaces, prefer `--params`.
- **Empty stdout:** `exit_code: 0` with empty `stdout` is success — don't retry.
- **No `exec` tool (404):** run `tools builtin list` first — admin may have disabled it.

## Cross-refs

- Tool mgmt (enable/disable/tenant-config): [providers-skills-tools.md](providers-skills-tools.md)
- Auth required first: [auth-and-config.md](auth-and-config.md)
- Session context: [chat-sessions.md](chat-sessions.md)
