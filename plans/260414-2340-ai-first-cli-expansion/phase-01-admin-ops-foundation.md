---
phase: 1
title: Admin/Ops Foundation
status: pending
priority: high
blockedBy: [phase-00]
---

# Phase 1 — Admin/Ops Foundation

## Context Links
- Brainstorm §2 (G1, G2, G9, G10) + §3 (config permissions)
- Server WS: `../../../goclaw/internal/gateway/methods/{tenants.go,heartbeat.go,config_permissions.go}`
- Server HTTP: `../../../goclaw/internal/http/{tenants.go,system_configs.go,edition.go}`
- Protocol methods: `../../../goclaw/pkg/protocol/methods.go`

## Overview
- **Priority:** High (tier 🔥 admin/ops)
- **Status:** Pending
- **Description:** Thêm 5 command groups mới + config permissions subcommand. Foundation cho multi-tenant, monitoring, system-wide config management. AI-value mixed: tenants/heartbeat high (AI monitor & context); system-configs/edition lower.

## Key Insights
- Tenants là resource top-level cho multi-tenancy — ảnh hưởng đến data isolation của mọi tenant-scoped endpoint
- Heartbeat có 8 WS methods khác nhau — gộp vào một group với nested subcommands (`heartbeat checklist get/set`, `heartbeat logs`)
- `system-configs` dùng HTTP PUT/DELETE cho mutation — khác pattern WS của config.get/apply/patch (dual config system)
- `edition` là public endpoint (no auth) — có thể dùng làm healthcheck
- `config.permissions.*` WS methods bổ sung cho existing config group

## Requirements

### Functional
- **Tenants:** list, get, create, update, users list/add/remove, mine (current user tenant)
- **Heartbeat:** get, set, toggle, test, logs (with `--follow`), targets, checklist get/set
- **System-configs:** list, get, set, delete
- **Edition:** show
- **Config permissions:** list, grant, revoke (extend existing `config` group)

### Non-Functional
- All commands support `--output=table|json|yaml`
- Destructive ops (`tenants update`, `users remove`, `system-configs delete`): `--yes` flag + interactive confirm
- `tenants users remove` có thể cần typed confirmation (tenant ID hoặc username)
- Heartbeat logs with `--follow` uses P0 `FollowStream` helper

## Architecture

### Command tree
```
goclaw tenants
├── list
├── get <id>
├── create --name=... --slug=...
├── update <id> [--name=...]
├── mine
└── users
    ├── list <tenantID>
    ├── add <tenantID> --user-id=... [--role=...]
    └── remove <tenantID> <userID>

goclaw heartbeat
├── get [--agent=<key>]
├── set --agent=<key> --interval=<dur>
├── toggle --agent=<key>
├── test --agent=<key>
├── logs [--agent=<key>] [--follow] [--tail=N]
├── targets
└── checklist
    ├── get --agent=<key>
    └── set --agent=<key> --items=<json>

goclaw system-configs
├── list
├── get <key>
├── set <key> <value> [--json]
└── delete <key>

goclaw edition

goclaw config (existing, extend)
└── permissions
    ├── list
    ├── grant --user=<id> --key=<path>
    └── revoke --user=<id> --key=<path>
```

## Related Code Files

### Create
- `cmd/tenants.go` (estimated ~180 LoC)
- `cmd/heartbeat.go` (estimated ~200 LoC — may split `heartbeat_checklist.go` if exceeds)
- `cmd/system_configs.go` (estimated ~90 LoC)
- `cmd/edition.go` (estimated ~40 LoC)

### Modify
- `cmd/config_cmd.go` — append `permissions` subcommand group
- `cmd/root.go` — register new groups
- `docs/codebase-summary.md` — document new groups
- `README.md` — command table update

### Reference (read, don't modify)
- `internal/client/http.go` (use `Get`/`Post`/`Put`/`Delete` helpers)
- `internal/client/websocket.go` (use `ws.Call`)
- `internal/output/printer.go` (Print formatter)

## Implementation Steps

### Step 1: tenants (HTTP-based)
1. Create `cmd/tenants.go` with `tenantsCmd` root
2. Implement `list` → `GET /v1/tenants`
3. Implement `get <id>` → `GET /v1/tenants/{id}`
4. Implement `create` → `POST /v1/tenants` with `--name`/`--slug` flags
5. Implement `update <id>` → `PATCH /v1/tenants/{id}`
6. Implement `mine` → `ws.Call("tenants.mine", nil)` (WS only per server)
7. Implement `users list/add/remove` subcommands
8. `users remove` requires `--yes` + typed confirmation (tenant ID)

### Step 2: heartbeat (WS-based)
1. Create `cmd/heartbeat.go` with `heartbeatCmd` root
2. Implement get/set/toggle/test → `ws.Call("heartbeat.{method}", ...)`
3. Implement `targets` → `ws.Call("heartbeat.targets", nil)`
4. Implement `logs` with `--follow` using P0 `FollowStream` (subscribe to heartbeat.logs stream)
5. Implement `checklist get/set` as nested group
6. If file >200 LoC, split `heartbeat_checklist.go`

### Step 3: system-configs (HTTP)
1. Create `cmd/system_configs.go`
2. Implement list/get/set/delete → `/v1/system-configs/*`
3. `set` supports `--json` flag to parse value as JSON
4. `delete` requires `--yes`

### Step 4: edition (HTTP)
1. Create `cmd/edition.go` with single command
2. `GET /v1/edition` (no auth)
3. Pretty-print name/version/features

### Step 5: config permissions
1. Extend `cmd/config_cmd.go` với `configPermissionsCmd` subcommand
2. list → `ws.Call("config.permissions.list", nil)`
3. grant/revoke với `--user`/`--key` flags
4. grant/revoke là admin-only — expose `--yes` for revoke

### Step 6: Tests
1. `cmd/tenants_test.go` — httptest for HTTP endpoints
2. `cmd/heartbeat_test.go` — WS upgrader for WS methods
3. `cmd/system_configs_test.go` — httptest
4. `cmd/edition_test.go` — httptest (no auth flow)
5. Coverage target: ≥60%

### Step 7: Docs
1. Update `docs/codebase-summary.md` sections
2. Update `README.md` command table
3. Add examples to each command's `--help` Long description

## Todo List

- [ ] 1.1: `cmd/tenants.go` with list/get/create/update/mine
- [ ] 1.2: `cmd/tenants.go` users subcommand (list/add/remove)
- [ ] 1.3: Typed confirmation for `tenants users remove`
- [ ] 2.1: `cmd/heartbeat.go` get/set/toggle/test/targets
- [ ] 2.2: `cmd/heartbeat.go` logs with `--follow` via FollowStream
- [ ] 2.3: `cmd/heartbeat.go` checklist get/set (split file if >200 LoC)
- [ ] 3.1: `cmd/system_configs.go` CRUD
- [ ] 3.2: `--json` flag for `set`
- [ ] 4.1: `cmd/edition.go` show command
- [ ] 5.1: Extend `cmd/config_cmd.go` with permissions subcommand
- [ ] 5.2: Wire `--yes` on revoke
- [ ] 6.1: Tests for tenants (httptest)
- [ ] 6.2: Tests for heartbeat (WS upgrader)
- [ ] 6.3: Tests for system-configs + edition + config permissions
- [ ] 6.4: `go test ./...` pass, coverage ≥60% for new code
- [ ] 7.1: `docs/codebase-summary.md` update
- [ ] 7.2: `README.md` command table + 2-3 usage examples
- [ ] 7.3: `--help` Long descriptions with JSON examples

## Success Criteria
- [ ] `goclaw tenants list --output=json` returns tenant array
- [ ] `goclaw tenants mine` returns current user's tenant
- [ ] `goclaw heartbeat logs --follow` streams events until ctrl-c
- [ ] `goclaw heartbeat checklist get --agent=<key>` returns checklist JSON
- [ ] `goclaw system-configs set key value` persists, `get key` retrieves same value
- [ ] `goclaw edition` works without auth
- [ ] `goclaw config permissions list` returns permission entries
- [ ] All commands return P0 structured errors + correct exit codes
- [ ] Build + vet + test all pass on Windows/Linux/macOS CI

## Risk Assessment

| Risk | Mitigation |
|---|---|
| `tenants users remove` delete wrong user | Require typed confirmation (user ID match) |
| `system-configs delete` break server operation | `--yes` required; server side should validate protected keys |
| Heartbeat `--follow` reconnect during outage | Inherit P0 FollowStream backoff |
| Config permissions overlap with existing config commands | Clear `--help` distinguishing config data vs access permissions |
| File size: heartbeat.go có 8 methods có thể vượt 200 LoC | Split `heartbeat_checklist.go` nếu cần |

## Security Considerations
- Tenants mgmt requires admin scope — verify server-side; CLI không bypass
- System-configs có thể chứa secrets — mask output in table mode (show `***` cho sensitive keys)
- Heartbeat targets có thể reveal internal infrastructure — error pass-through only
- Edition endpoint public — no sensitive data leak risk

## Next Steps
- Dependencies: Phase 0 (error/exit/follow patterns)
- Unblocks: None directly (P2-P5 independent of P1)
- Follow-up: Nothing specific
