# Teams — collaboration

## When to use

User wants to inspect/manage agent teams, add/remove members, create/approve team tasks, or access a team's shared workspace. **All teams ops are WebSocket-backed** (source: `cmd/teams*.go`); no HTTP fallback.

## Commands in scope

- `goclaw teams list/get/create/update/delete` — team CRUD (source: `teams.go`, `teams_extra.go`)
- `goclaw teams members list/add/remove/reassign` — membership (source: `teams_members.go`)
- `goclaw teams events <team-id>` — **STREAMING WS subscribe, skill REFUSES**
- `goclaw teams tasks list/get/create/delete/approve/reject/reassign` — task board (source: `teams_tasks*.go`)
- `goclaw teams workspace list/get/put/delete` — team file workspace (source: `teams_workspace.go`)

## Verified flags (typical — verify per subcommand via `goclaw teams <sub> --help`)

### `teams create`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--name <n>` | string | Team name |
| `--agents <ids>` | CSV | Initial agent list |

### `teams members add` / `remove`
| Flag | Type | Purpose |
| --- | --- | --- |
| (team-id arg) | positional | Parent team (verify syntax with `--help`) |
| `--user <id>` | string | User to add/remove |
| `--role <r>` | string | Role (for add) |

### `teams tasks create` (verify with --help)
Flag set varies. Run `goclaw teams tasks create --help` for current syntax.

## JSON output

- ✅ `teams list/get`, `teams members list`, `teams tasks list/get`, `teams workspace list/get` — JSON
- ⚠️ all `create/update/delete/add/remove/reassign/approve/reject/put` — WS call returns success-ish payload; parse `unmarshalMap` but most fields empty

## Destructive ops

| Command | Confirm |
| --- | --- |
| `teams delete` | YES — cascades members + tasks + workspace |
| `teams members remove` | YES |
| `teams tasks delete` | YES (`teams_tasks_delete.go`) |
| `teams tasks reject` | YES |
| `teams workspace delete` | YES |

## Common patterns

### Example 1: list teams + get one
```bash
goclaw teams list --output json
goclaw teams get <team-id> --output json
```

### Example 2: create team with 2 agents
```bash
goclaw teams create --name "Support Squad" --agents "agent-a,agent-b" --output json
```

### Example 3: add member + create task
```bash
# Run --help first — team-id may be positional:
goclaw teams members add --help
goclaw teams tasks create --help
```

### Example 4: approve task
```bash
goclaw teams tasks approve <task-id> --output json
```

### Example 5: workspace file upload (verify syntax first)
```bash
goclaw teams workspace --help
goclaw teams workspace list --output json
```

## Edge cases & gotchas

- **All WS — no HTTP.** If WS connect fails, nothing works. Advise user to check `goclaw status` first.
- **`teams events`** = WS subscribe streaming live team activity. Skill REFUSES (same as `logs tail`).
- **Workspace path** slash-separated, no leading `/`. Same convention as `memory`.
- **Task state machine:** pending → assigned → in_progress → (approve|reject) → completed. `reassign` works in pending/assigned only.
- **`teams.tasks` vs `teams tasks`:** CLI uses space (subcommand). WS method is `teams.tasks.<verb>`.
- **Cascade delete warning:** `teams delete` removes members + tasks + workspace. Triple-confirm if team has active tasks.

## Cross-refs

- Delegation between agents (different from teams): [agents-advanced.md](agents-advanced.md)
- Chat sessions per user: [chat-sessions.md](chat-sessions.md)
- Individual agent files: [agents-core.md](agents-core.md)
