---
phase: 5
title: Extensions + Remaining (Pair, OAuth, Packages, Send, Quota, Subcommand Extensions)
status: pending
priority: medium
blockedBy: [phase-00]
splitRecommendation: "5a (new groups), 5b (extend existing) if >800 LoC"
---

# Phase 5 — Extensions + Remaining

## Context Links
- Brainstorm §2 G3/G6/G7/G11/G12/G13 + §3 partial-gap subcommands + §5 Phase 5
- Server WS: `../../../goclaw/internal/gateway/methods/{pairing.go,quota_methods.go,send.go}`
- Server HTTP: `../../../goclaw/internal/http/{oauth.go,oauth_quota.go,packages.go,user_search.go,contact_merge_handlers.go,channel_instances.go,pending_messages.go,mcp.go,skills.go,providers.go,provider_verify.go,provider_embedding_validation.go,provider_models_catalog.go,builtin_tools.go,secure_cli.go,secure_cli_user_credentials.go,secure_cli_agent_grants.go}`

## Overview
- **Priority:** Medium
- **Status:** Pending
- **Description:** Last phase, catch remaining gaps. Mix of new groups (pair, oauth, packages, users, quota, send) và subcommand extensions cho existing groups (mcp, skills, providers, tools, admin credentials, channels). Best candidate for splitting nếu LoC vượt 800.

## Key Insights
- `send` generic WS là inter-agent messaging primitive — AI orchestration use case (1-shot send without chat session lifecycle)
- Pair management group khác với `auth login --pair` flow (latter keep as login shortcut; new `pair` group quản lý existing pairings)
- OAuth pool (ChatGPT/OpenAI): có callback flow cần manual URL paste; CLI support `--callback-url=<url>` to post manually
- Packages = skill dependency packages (Python/Node via uv/npm) — important cho skill runtime management
- Existing `admin credentials` cần thêm: update, test (dry-run), presets, check-binary, user-credentials sub-tree, agent-grants sub-tree
- `channels pending` existing command chỉ có `list/retry` — thiếu `groups/messages/delete/compact` (system-wide pending messages, khác channel-specific)

## Requirements

### New Groups
**Pair:**
- `pair list [--status=pending|approved|revoked]`
- `pair request [--purpose=...]`
- `pair approve <pairID>`
- `pair deny <pairID>`
- `pair revoke <pairID>` (--yes)

**OAuth (ChatGPT/OpenAI pool):**
- `oauth status --provider=chatgpt|openai`
- `oauth quota --provider=...`
- `oauth start --provider=...` — returns URL to paste in browser
- `oauth callback --provider=... --code=<code>` — manual paste after browser auth
- `oauth logout --provider=...` (--yes)

**Packages:**
- `packages list`
- `packages install <name>` (admin)
- `packages uninstall <name>` (admin) (--yes)
- `packages runtimes` — available runtimes
- `packages deny-groups` — GET /v1/shell-deny-groups

**Users:**
- `users search --q=<query> [--limit=N] [--peer-kind=...]`

**Quota:**
- `quota usage [--agent=<key>]` — WS `quota.usage`

**Send:**
- `send --to=<agent> --content=... [--channel=...]` — WS `send`

### Extensions to existing groups
**channels (pending subcommand extensions):**
- `channels pending groups`
- `channels pending messages --group=<id>`
- `channels pending delete`
- `channels pending compact`

**contacts (within channels group):**
- `channels contacts merge --source=<id> --target=<id>`
- `channels contacts merged <tenantUserID>`

**mcp:**
- `mcp servers reconnect <id>`
- `mcp servers test-connection --config=<json>` (test before create, khác `test <id>` test existing)

**skills:**
- `skills install-dep <dep>` (single, existing is `install-deps` plural)
- `skills tenant-config get/set/delete <skillID>`

**providers:**
- `providers verify-embedding <id>`
- `providers codex-pool-activity <id>` (provider version)
- `providers embedding-status`
- `providers claude-cli auth-status`

**tools:**
- `tools builtin tenant-config get/set/delete <name>`

**admin credentials:**
- `admin credentials update <id>`
- `admin credentials test <id>` (dry-run)
- `admin credentials presets`
- `admin credentials check-binary`
- `admin credentials user-credentials list/get/set/delete <credID> [userID]`
- `admin credentials agent-grants list/create/get/update/delete <credID> [grantID]`

## Architecture

### Command tree (new)
```
goclaw pair
├── list
├── request
├── approve <id>
├── deny <id>
└── revoke <id>

goclaw oauth
├── status --provider=chatgpt|openai
├── quota --provider=...
├── start --provider=...
├── callback --provider=... --code=<code>
└── logout --provider=...

goclaw packages
├── list
├── install <name>
├── uninstall <name>
├── runtimes
└── deny-groups

goclaw users search --q=...

goclaw quota usage [--agent=...]

goclaw send --to=<agent> --content=...
```

### File structure (split per modularization rule)
```
cmd/
├── pair.go                     # ~120 LoC
├── oauth.go                    # ~150 LoC
├── packages.go                 # ~100 LoC
├── users.go                    # ~40 LoC
├── quota.go                    # ~50 LoC
├── send.go                     # ~50 LoC
├── channels_pending.go         # extract existing + add groups/messages/delete/compact (~120 LoC)
├── channels_contacts.go        # extract existing + add merge/merged (~100 LoC)
├── mcp_servers.go              # existing mcp.go may already split; add reconnect/test-connection
├── skills_deps.go              # extract from skills.go + install-dep single
├── skills_tenant_config.go     # get/set/delete
├── providers_verify.go         # verify-embedding + embedding-status
├── providers_claude_cli.go     # claude-cli auth-status
├── providers_codex_pool.go     # codex-pool-activity (provider version)
├── tools_builtin_tenant.go     # tenant-config
├── admin_credentials.go        # extract from admin.go + update/test/presets/check-binary
├── admin_credentials_users.go  # user-credentials subcommand
└── admin_credentials_grants.go # agent-grants subcommand
```

Note: extracting from `admin.go` (402 LoC) and `channels.go` (could be ~270 LoC after pending/contacts) necessary per project modularization rule.

## Related Code Files

### Create
- `cmd/pair.go`
- `cmd/oauth.go`
- `cmd/packages.go`
- `cmd/users.go`
- `cmd/quota.go`
- `cmd/send.go`
- `cmd/channels_pending.go`, `cmd/channels_contacts.go` (extract + extend)
- `cmd/skills_deps.go`, `cmd/skills_tenant_config.go` (extract + extend)
- `cmd/providers_verify.go`, `cmd/providers_claude_cli.go`, `cmd/providers_codex_pool.go`
- `cmd/tools_builtin_tenant.go`
- `cmd/admin_credentials.go`, `cmd/admin_credentials_users.go`, `cmd/admin_credentials_grants.go`
- `cmd/mcp_servers.go` (if mcp.go split; else extend existing)

### Modify
- `cmd/admin.go` — trim to <200 LoC (extract credentials subtree)
- `cmd/channels.go` — trim (extract pending + contacts)
- `cmd/mcp.go` — add reconnect/test-connection
- `cmd/skills.go` — remove extracted deps/tenant-config
- `cmd/providers.go` — remove extracted verify/claude-cli/codex
- `cmd/tools.go` — add builtin tenant-config wiring
- `cmd/root.go` — register new groups
- Existing `auth.go` — no change (keep `auth login --pair` shortcut)

### Reference
- `internal/client/http.go`
- `internal/client/websocket.go`
- P0 helpers

## Implementation Steps

### Phase 5a — New Groups

#### Step 1: pair
1. `cmd/pair.go` với pairCmd
2. WS: list/request/approve/deny/revoke → `device.pair.*`
3. Revoke `--yes`

#### Step 2: oauth
1. `cmd/oauth.go` với providers `chatgpt|openai`
2. `start` trả về URL; print to stderr để user browser mở
3. `callback` accept code/URL sau khi browser redirect; handle manual paste
4. `logout --yes`

#### Step 3: packages
1. `cmd/packages.go`
2. list/install/uninstall/runtimes → `/v1/packages/*`
3. `deny-groups` → `/v1/shell-deny-groups`
4. `uninstall --yes`

#### Step 4: users + quota + send
1. `cmd/users.go` — search with `--q`/`--limit`/`--peer-kind`
2. `cmd/quota.go` — WS `quota.usage`
3. `cmd/send.go` — WS `send` with `--to`/`--content`/`--channel`

### Phase 5b — Extensions

#### Step 5: channels modularization + extensions
1. Extract `cmd/channels_pending.go` from `channels.go`
2. Add pending groups/messages/delete/compact subcommands
3. Extract `cmd/channels_contacts.go`
4. Add contacts merge/merged

#### Step 6: mcp extensions
1. Add `servers reconnect <id>` to mcp.go (or mcp_servers.go if split)
2. Add `servers test-connection --config=<json>` — test trước create

#### Step 7: skills extensions
1. Extract `cmd/skills_deps.go` (install-deps existing + new install-dep single)
2. Create `cmd/skills_tenant_config.go` with get/set/delete

#### Step 8: providers extensions
1. Create `cmd/providers_verify.go` with verify-embedding + embedding-status
2. Create `cmd/providers_claude_cli.go` với auth-status
3. Create `cmd/providers_codex_pool.go` with activity

#### Step 9: tools builtin tenant-config
1. Create `cmd/tools_builtin_tenant.go`
2. get/set/delete tenant-config for builtin tools

#### Step 10: admin credentials modularization + extensions
1. Extract `cmd/admin_credentials.go` from `admin.go`
2. Add update/test/presets/check-binary
3. Create `cmd/admin_credentials_users.go` — user-credentials sub-tree
4. Create `cmd/admin_credentials_grants.go` — agent-grants sub-tree
5. Verify `admin.go` <200 LoC

### Shared

#### Step 11: Tests
1. Unit tests cho each new file
2. Integration test cho pair flow (request → approve → revoke)
3. OAuth flow test (start → callback)
4. Send test (WS ack)

#### Step 12: Docs
1. README update
2. docs/codebase-summary.md complete refresh
3. Verify `--help` examples

## Todo List

### 5a — New groups
- [ ] 1.1: `cmd/pair.go` CRUD
- [ ] 1.2: Revoke with `--yes`
- [ ] 2.1: `cmd/oauth.go` status/quota
- [ ] 2.2: oauth start (print URL to stderr)
- [ ] 2.3: oauth callback (accept code)
- [ ] 2.4: oauth logout
- [ ] 3.1: `cmd/packages.go` list/install/uninstall
- [ ] 3.2: runtimes + deny-groups
- [ ] 4.1: `cmd/users.go` search
- [ ] 4.2: `cmd/quota.go` usage
- [ ] 4.3: `cmd/send.go` inter-agent send

### 5b — Extensions
- [ ] 5.1: Extract `cmd/channels_pending.go`
- [ ] 5.2: Add pending groups/messages/delete/compact
- [ ] 5.3: Extract `cmd/channels_contacts.go` + merge/merged
- [ ] 6.1: mcp servers reconnect
- [ ] 6.2: mcp servers test-connection
- [ ] 7.1: Extract `cmd/skills_deps.go` + install-dep single
- [ ] 7.2: `cmd/skills_tenant_config.go`
- [ ] 8.1: `cmd/providers_verify.go` verify-embedding + embedding-status
- [ ] 8.2: `cmd/providers_claude_cli.go` auth-status
- [ ] 8.3: `cmd/providers_codex_pool.go` activity
- [ ] 9.1: `cmd/tools_builtin_tenant.go` tenant-config
- [ ] 10.1: Extract `cmd/admin_credentials.go` + update/test/presets/check-binary
- [ ] 10.2: `cmd/admin_credentials_users.go`
- [ ] 10.3: `cmd/admin_credentials_grants.go`
- [ ] 10.4: Verify `admin.go` <200 LoC

### Shared
- [ ] 11.1: Tests per new file
- [ ] 11.2: pair flow integration test
- [ ] 11.3: oauth callback flow test
- [ ] 12.1: README with `send` example for AI orchestration
- [ ] 12.2: docs/codebase-summary.md final refresh
- [ ] 12.3: Verify no cmd/*.go >200 LoC

## Success Criteria

### New groups
- [ ] `goclaw pair list` shows pairings; approve/deny/revoke flow works
- [ ] `goclaw oauth start --provider=chatgpt` prints URL to stderr, auth_id to stdout
- [ ] `goclaw oauth callback --provider=chatgpt --code=...` completes auth
- [ ] `goclaw packages install <name>` installs; `uninstall --yes` removes
- [ ] `goclaw users search --q=duy` returns user array
- [ ] `goclaw quota usage --agent=<key>` returns quota JSON
- [ ] `goclaw send --to=agentA --content="msg"` delivers, returns ack ID

### Extensions
- [ ] `goclaw channels pending groups` lists groups; `messages --group=X` lists msgs
- [ ] `goclaw channels contacts merge --source=X --target=Y` merges
- [ ] `goclaw mcp servers reconnect <id>` triggers reconnection
- [ ] `goclaw mcp servers test-connection --config=<json>` tests before create
- [ ] `goclaw skills install-dep numpy` installs single
- [ ] `goclaw skills tenant-config get <id>` returns config
- [ ] `goclaw providers verify-embedding <id>` tests embedding
- [ ] `goclaw providers claude-cli auth-status` returns status
- [ ] `goclaw tools builtin tenant-config set Bash '{...}'` updates
- [ ] `goclaw admin credentials test <id>` dry-run
- [ ] `goclaw admin credentials user-credentials list <credID>` works
- [ ] `goclaw admin credentials agent-grants list <credID>` works

### Quality
- [ ] All cmd/*.go <200 LoC (enforce via CI check or manual review)
- [ ] `go build ./... && go vet ./... && go test ./...` pass
- [ ] Coverage: ≥60% new code, ≥70% for send (AI orchestration critical)
- [ ] README + docs updated
- [ ] ≥95% server endpoint coverage reached (overall goal)

## Risk Assessment

| Risk | Mitigation |
|---|---|
| OAuth callback URL format mismatch between providers | Document each provider's flow separately in `--help` |
| `pair` group naming conflict with `auth login --pair` | Keep auth shortcut; pair is separate management, document both |
| `send` misuse flooding agent with spam | Rate-limited server-side; CLI no retry on SEND success |
| File extraction breaks imports | Test compile after each extraction step |
| Admin credentials user/agent grants nested 3 levels deep — UX cumbersome | Support `--cred-id` flag at top-level to avoid repetition |
| OAuth pool tokens leak in quota output | Server masks; CLI pass-through |
| Large LoC 5a+5b merged PR | Split proactively if >1000 LoC |

## Security Considerations
- `pair approve` elevates device access — admin-only verify
- `oauth logout` invalidates token — require `--yes`
- `send` bypasses chat session logging — verify server audit logs; document in `--help`
- `admin credentials test` may execute binary — warn in docs, ensure safe env
- `user-credentials` stores per-user API keys — masked server-side; CLI pass-through
- `packages uninstall` affects shared runtimes — admin scope, `--yes` required

## Next Steps
- Dependencies: Phase 0
- Unblocks: None (terminal phase)
- Follow-up post-merge:
  1. Journal entry for full project
  2. Verify overall coverage ≥95% via script
  3. Open issue tracking: idempotency-key (server side), schema dump command (future)
  4. CHANGELOG finalization with migration guide

## Unresolved Questions
1. `send` method params exact shape — check server `methods/send.go` signature
2. OAuth callback: có cần CLI spin up local HTTP server để catch redirect không? (Auth flow complexity vs manual paste UX)
3. `admin credentials agent-grants` create có đủ fields qua flags hay cần `--body=<json>`?
4. Split 5a/5b: final LoC check trước khi open PR
5. `send --channel=<id>`: channel optional hay required? Check server default behavior
