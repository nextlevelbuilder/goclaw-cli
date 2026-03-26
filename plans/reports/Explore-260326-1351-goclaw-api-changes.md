# GoClaw API Changes Analysis (Jan 2026 - Mar 2026)

**Period:** 2026-01-01 through 2026-03-26  
**Total commits analyzed:** 150+  
**Scope:** HTTP endpoints, features, breaking changes, tenant isolation

---

## Executive Summary

GoClaw underwent a **major architectural shift toward multi-tenant isolation** (completed March 23, #359) and introduced **system-wide tenant scoping**. The goclaw-cli must be updated to:

1. **Support per-tenant operations** — API keys now have `tenant_id` binding
2. **Propagate tenant context** — All requests need tenant awareness
3. **Support new endpoints** — Tenants, system config, skill/tool tenant config
4. **Support skill versioning & tenant overrides** — Skills have per-tenant visibility toggles
5. **Support system configuration** — New system_configs table for per-tenant settings

---

## Phase 1: Multi-Tenant Architecture (2026-03-23, commit cd022699)

### Breaking Changes

#### Schema Changes (Migrations 000026 & 000027)

**Migration 000026: User Binding + Teams Grants**
- `api_keys.owner_id` — Forces user_id on auth, prevents spoofing
- `api_keys.tenant_id` — Nullable; NULL = system-level, set = tenant-scoped
- Removes: `delegation_history`, `handoff_routes` tables
- Adds: `team_user_grants` table for user-to-team access

**Migration 000027: Tenant Foundation**
- Creates: `tenants` + `tenant_users` tables
- Adds `tenant_id` column to 30+ tables (agents, sessions, cron, skills, providers, etc.)
- Creates: `builtin_tool_tenant_configs`, `skill_tenant_configs` for per-tenant overrides
- Removes: `custom_tools` table (never wired to agent loop)

### New Tenant Concepts

| Concept | Details |
|---------|---------|
| **Master Tenant** | UUID-based, default for all legacy data |
| **Tenant Scope** | Non-master tenants have isolated agents, sessions, memory, teams, providers |
| **Cross-Tenant Key** | `api_key.tenant_id = NULL` → system-level key, requires `X-GoClaw-Tenant-Id` header |
| **Tenant-Bound Key** | `api_key.tenant_id = UUID` → auto-scoped, no header needed |

### Auth Tenant Resolution

```
HTTP (5 paths):
  1. Bearer token (API key or gateway token)
  2. X-GoClaw-User-Id header (required)
  3. X-GoClaw-Tenant-Id header (optional, system keys only)
  4. X-GoClaw-Agent-Id header (optional, alternative to model field)

WebSocket:
  1. Token on connect
  2. Auto-resolves tenant_id
  3. Injects into Client context

Channels (direct webhook):
  1. Tenant resolved from channel_instances DB config
  2. Baked at setup time, no auth needed
```

### Context Propagation

- `WithTenantID(ctx, uuid)` / `TenantIDFromContext(ctx)` — Fail-closed: `uuid.Nil` errors
- `WithCrossTenant(ctx, bool)` / `IsCrossTenant(ctx)` — Owner/system admin flag
- ALL 30+ store queries: `WHERE tenant_id = $N` (non-cross-tenant)
- System skills (`is_system=true`) bypass tenant filter

### Runtime Isolation

- **Event bus:** TenantID field on Event, fail-closed filter in event_filter.go
- **Cron:** Tenant context injected in RunJob handler
- **Subagent:** Tenant validation prevents cross-tenant spawn
- **Workspace:** Resolver computes tenant-scoped `workspace` + `dataDir`

### New WS RPC Methods (Tenant Management)

- `tenants.list` — List all tenants (admin only)
- `tenants.get {id}` — Get tenant details
- `tenants.create {name, slug}` — Create new tenant
- `tenants.update {id, ...}` — Update tenant
- `tenants.users.list {tenant_id}` — List tenant members
- `tenants.users.add {tenant_id, user_id, role}` — Add user to tenant
- `tenants.users.remove {tenant_id, user_id}` — Remove user from tenant

### New HTTP Endpoints (Tenants)

```
GET  /v1/tenants                 — Admin only, list all
GET  /v1/tenants/{id}           — Get tenant
POST /v1/tenants                — Create tenant
PATCH /v1/tenants/{id}          — Update tenant
GET  /v1/tenants/{id}/users     — List tenant users
POST /v1/tenants/{id}/users     — Add user to tenant
```

---

## Phase 2: Skills Tenant Config (2026-03-25, commit c5164255)

### New Endpoints

```
PUT    /v1/skills/{id}/tenant-config      — Set tenant-specific visibility
DELETE /v1/skills/{id}/tenant-config      — Clear tenant-specific visibility
GET    /v1/skills (list)                   — Includes "tenant_enabled" field when tenant-scoped
```

### Schema Changes (Implicit in Migration 000027)

- `skill_tenant_configs` table — Per-tenant skill visibility overrides
- `SkillTenantConfigStore.ListAll()` — New interface method

### CLI Impact

- CLI skill list should show `tenant_enabled` status
- CLI should support enabling/disabling skills per-tenant
- CLI skill upload must inherit tenant_id from auth context

---

## Phase 3: System Configuration (2026-03-24, commit 651072a9)

### New Schema (Migration 000029)

```sql
CREATE TABLE system_configs (
  tenant_id UUID NOT NULL,
  key TEXT NOT NULL,
  value JSONB,
  PRIMARY KEY (tenant_id, key),
  FOREIGN KEY (tenant_id) REFERENCES tenants(id)
)
```

### New Endpoints

```
GET  /v1/system-configs           — List all configs for tenant
GET  /v1/system-configs/{key}     — Get specific config
PUT  /v1/system-configs/{key}     — Set config (admin only)
```

### Configurable Keys

- `embedding` — Provider/model/dimension config
- `tool_status` — Show tool execution status toggle
- `block_reply` — Suppress intermediate text toggle
- `intent_classify` — Enable intent classification
- `pending_compaction` — Provider/model + threshold settings

### CLI Impact

- CLI should support reading/writing system-level configuration
- Admin-only operations; requires admin role

---

## Phase 4: Per-Tenant Builtin Tool Config (2026-03-23, commit cd022699)

### New Endpoints

```
PUT    /v1/tools/builtin/{name}/tenant-config       — Set tenant visibility
DELETE /v1/tools/builtin/{name}/tenant-config       — Clear tenant visibility
GET    /v1/tools/builtin (list)                      — Includes "tenant_enabled" field
```

### Schema

- `builtin_tool_tenant_configs` table (Migration 000027)

### CLI Impact

- CLI tool list should show per-tenant visibility
- CLI should support enabling/disabling builtin tools per-tenant

---

## Feature: Skill Versioning & Grants (2026-03-25, commit 9168e4b4)

### Changes

- **Skill upload** — Enforces tenant isolation
- **Skill versioning** — Per-tenant version control
- **Skill grants** — Per-tenant grant enforcement (agent/user access control)

### New Endpoint

```
GET /v1/skills/{id}/versions      — List skill versions
GET /v1/skills/{id}/files/{path}  — Read skill file (tenant-scoped)
```

### CLI Impact

- Skill upload must respect tenant context
- Skill grants must be per-tenant visible

---

## Feature: API Key Binding (2026-03-23, commit cd022699)

### Breaking Changes

- API keys now have **required** `tenant_id` field
- Gateway token + `X-GoClaw-Tenant-Id` header for cross-tenant ops
- Tenant-bound keys auto-scope all requests (no header needed)

### New CLI Command Group

**Status:** Already added in commit 4fb4179 (2026-03-15)
```
goclaw api-keys list              — List API keys
goclaw api-keys create [flags]    — Create tenant-bound key
goclaw api-keys revoke {id}       — Revoke key
```

### CLI Impact

- `--tenant-id` flag for key creation
- Keys are now tenant-scoped; system keys require explicit cross-tenant flag
- API key CLI fully implemented

---

## Feature: API Docs Command (Already Implemented)

**Status:** Added in commit 4fb4179 (2026-03-15)
```
goclaw api-docs open              — Open Swagger UI
goclaw api-docs spec              — Fetch OpenAPI spec
```

---

## All New HTTP Endpoints (Complete List)

### Tenants
```
GET  /v1/tenants
GET  /v1/tenants/{id}
POST /v1/tenants
PATCH /v1/tenants/{id}
GET  /v1/tenants/{id}/users
POST /v1/tenants/{id}/users
```

### System Configuration
```
GET  /v1/system-configs
GET  /v1/system-configs/{key}
PUT  /v1/system-configs/{key}
```

### Skills Tenant Config
```
PUT  /v1/skills/{id}/tenant-config
DELETE /v1/skills/{id}/tenant-config
```

### Builtin Tools Tenant Config
```
PUT  /v1/tools/builtin/{name}/tenant-config
DELETE /v1/tools/builtin/{name}/tenant-config
```

### Skill Versions & Files
```
GET  /v1/skills/{id}/versions
GET  /v1/skills/{id}/files/{path...}
```

---

## Breaking Changes Summary

| Change | Impact | Migration Path |
|--------|--------|-----------------|
| API keys require `tenant_id` | Keys are tenant-scoped or system-scoped | Add `--tenant-id` to key creation |
| `X-GoClaw-Tenant-Id` header optional for tenant-bound keys | Auto-scoped requests | Update HTTP client to read tenant from key |
| Master tenant default for legacy data | All queries add `tenant_id = ?` filter | No action; auto-applied |
| Custom tools table removed | Custom tools no longer supported | Use skills or builtin tools instead |
| Session/agent isolation by tenant | Can't cross-tenant spawn/query | Tenants must be explicit in CLI |
| Skill/tool visibility per-tenant | Must toggle per-tenant | Add `--tenant-id` to visibility toggles |

---

## Gateway Endpoints Affected

**Auth injection (all handlers):**
- HTTP: `resolveAuthBearer()` sets `TenantID` on all 5 paths
- WS: `handleConnect()` sets `tenantID` on Client
- Event propagation: Filtered by `TenantID`

**Store layer:**
- All SELECT/INSERT/UPDATE/DELETE add `WHERE tenant_id = $N` (fail-closed)
- No fallback to master tenant for cross-tenant keys
- Strict isolation enforced at DB query level

---

## What the CLI Needs to Support

### 1. Tenant Management
- [ ] `goclaw tenants list` — List all tenants (admin)
- [ ] `goclaw tenants get {id}` — Get tenant details
- [ ] `goclaw tenants create {name} --slug {slug}` — Create tenant
- [ ] `goclaw tenants update {id} --name {name}` — Update tenant
- [ ] `goclaw tenants users list {id}` — List tenant members
- [ ] `goclaw tenants users add {id} --user-id {uid} --role {role}` — Add user
- [ ] `goclaw tenants users remove {id} --user-id {uid}` — Remove user

### 2. System Configuration (Admin)
- [ ] `goclaw config system list` — List all system configs
- [ ] `goclaw config system get {key}` — Get config value
- [ ] `goclaw config system set {key} {value}` — Set config
- [ ] Support keys: `embedding`, `tool_status`, `block_reply`, `intent_classify`, `pending_compaction`

### 3. Skill Tenant Config
- [ ] `goclaw skills list --show-tenant-status` — Show per-tenant visibility
- [ ] `goclaw skills enable-tenant {id} --tenant-id {tid}` — Enable for tenant
- [ ] `goclaw skills disable-tenant {id} --tenant-id {tid}` — Disable for tenant
- [ ] `goclaw skills versions {id}` — List versions

### 4. Builtin Tool Tenant Config
- [ ] `goclaw tools list --show-tenant-status` — Show per-tenant visibility
- [ ] `goclaw tools enable-tenant {name} --tenant-id {tid}` — Enable for tenant
- [ ] `goclaw tools disable-tenant {name} --tenant-id {tid}` — Disable for tenant

### 5. API Key Tenant Binding
- [x] `goclaw api-keys create --tenant-id {tid}` — Create tenant-bound key
- [x] Already implemented (commit 4fb4179)

### 6. Context Propagation
- [ ] All CLI commands must accept `--tenant-id` flag
- [ ] Fall back to auth context tenant if not provided
- [ ] Fail cleanly if tenant context is ambiguous or missing

---

## Database Changes Summary

| Migration | Changes | Impact |
|-----------|---------|--------|
| 000026 | `api_keys.owner_id`, `api_keys.tenant_id`, `team_user_grants` | Auth binding, team grants |
| 000027 | `tenants`, `tenant_users`, +30 tenant_id columns | Tenant foundation |
| 000028 | `comment_type` — Team comment classification | Task audit logs |
| 000029 | `system_configs` — Per-tenant key-value config | System settings |

---

## Unresolved Questions

1. Should CLI support **master tenant operations** explicitly (e.g., `--tenant-id master`)?
2. Should CLI support **cross-tenant admin keys** (system keys + `--tenant-id` header)?
3. Should skill/tool tenant visibility be **scoped to current tenant only**, or should admins see/modify all tenants?
4. Should **system config be global or per-tenant**? (Appears to be per-tenant based on schema)
5. Are there **tenant role-based restrictions** the CLI should enforce (viewer/operator/admin)?

