---
phase: 1
status: complete
priority: critical
effort: M
---

# Phase 1: Multi-Tenant Commands

## Overview

Add tenant management and system configuration commands. This is the foundation for all multi-tenant operations — every subsequent phase depends on tenant context being available.

## Key Insights

- GoClaw server added multi-tenant isolation (March 23, commit cd022699)
- 30+ DB tables now have `tenant_id` columns
- API keys are tenant-scoped
- System keys require `X-GoClaw-Tenant-Id` header
- All existing operations must be tenant-aware

## Requirements

### Functional
- `goclaw tenants list|get|create|update` — Tenant CRUD
- `goclaw tenants users list|add|remove` — Tenant user management
- `goclaw system-config list|get|set|delete` — Per-tenant KV config store
- Global `--tenant-id` flag on root command for tenant context override

### Non-Functional
- Admin-only access for tenant operations
- Proper error messages when tenant context missing

## Architecture

### HTTP Endpoints to Wire

```
# Tenants
GET    /v1/tenants
GET    /v1/tenants/{id}
POST   /v1/tenants
PATCH  /v1/tenants/{id}
GET    /v1/tenants/{id}/users
POST   /v1/tenants/{id}/users
DELETE /v1/tenants/{id}/users/{userId}

# System Config
GET    /v1/system-configs
GET    /v1/system-configs/{key}
PUT    /v1/system-configs/{key}
DELETE /v1/system-configs/{key}
```

### WS Methods to Wire

```
config.permissions.list
config.permissions.grant
config.permissions.revoke
```

## Related Code Files

### Files to Create
- `cmd/tenants.go` — Tenant CRUD + user management
- `cmd/system_config.go` — System configuration commands

### Files to Modify
- `cmd/root.go` — Add `--tenant-id` persistent flag
- `cmd/helpers.go` — Pass tenant-id header in HTTP requests if set
- `internal/client/http.go` — Support `X-GoClaw-Tenant-Id` header

## Implementation Steps

1. Add `--tenant-id` persistent flag to `root.go`
2. Update `internal/client/http.go` to attach `X-GoClaw-Tenant-Id` header when set
3. Create `cmd/tenants.go`:
   - `tenants list` — GET /v1/tenants, table: id, name, created
   - `tenants get <id>` — GET /v1/tenants/{id}
   - `tenants create --name <name>` — POST /v1/tenants
   - `tenants update <id> --name <name>` — PATCH /v1/tenants/{id}
   - `tenants users list <tenant-id>` — GET /v1/tenants/{id}/users
   - `tenants users add <tenant-id> --user-id <uid>` — POST /v1/tenants/{id}/users
   - `tenants users remove <tenant-id> <user-id>` — DELETE /v1/tenants/{id}/users/{userId}
4. Create `cmd/system_config.go`:
   - `system-config list` — GET /v1/system-configs
   - `system-config get <key>` — GET /v1/system-configs/{key}
   - `system-config set <key> --value <val>` — PUT /v1/system-configs/{key}
   - `system-config delete <key>` — DELETE /v1/system-configs/{key}
5. Register both command groups in `root.go`
6. `go build ./...` to verify compilation

## Todo List

- [ ] Add --tenant-id persistent flag to root.go
- [ ] Update HTTP client to pass tenant-id header
- [ ] Implement tenants CRUD commands
- [ ] Implement tenants users subcommands
- [ ] Implement system-config commands
- [ ] Register commands in root.go
- [ ] Compile check

## Success Criteria

- `goclaw tenants list` returns tenant list
- `goclaw system-config list` returns config keys
- `--tenant-id` flag propagates to all HTTP requests
- JSON/table output works for all new commands

## Risk Assessment

- **Low:** Straightforward CRUD, follows existing patterns exactly
- **Medium:** Tenant-id header propagation must not break existing commands when not set
