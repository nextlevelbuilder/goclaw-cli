# Brainstorm: Missing Commands Gap Analysis

**Date:** 2026-04-14
**Scope:** So sánh `goclaw-cli/cmd/` với `../goclaw/internal/` để tìm command còn thiếu
**Reference commit:** [c742f5d](https://github.com/nextlevelbuilder/goclaw-cli/commit/c742f5d36df9c1406b79c17efe022eb68cacb3ea) — added `api-keys` + `api-docs` groups

## Primary use case (REVISED 2026-04-14)

CLI này là **interface để AI tools (Claude Code, LangChain, custom agents) interact với GoClaw Gateway**. Consumer split: **~90% AI tools, ~10% human ops**. Không build MCP server song song — CLI là form duy nhất.

**Implications:**
- JSON output là default khi non-TTY; table cho TTY
- Structured error object khi `--output=json`
- Exit codes chuẩn hóa để AI dispatch action
- Không interactive prompt trên core path — mọi destructive op phải `--yes`-able
- Streaming/follow cho tail/events để AI poll progress
- Idempotency key cho create ops (AI retry safe)

**Decisions locked:**
- Scope: **Fill all core gaps** (High + Medium priority)
- Persona: **90% AI tools, 10% human** (dual but AI-weighted)
- Skip: Zalo Personal, WhatsApp (YAGNI)
- Include: **Full vault** group
- AI ergonomics: **structured errors + exit codes, `--follow`, idempotency keys** (opt-in per user)
- Skip: `goclaw schema/describe` meta-command (not selected — AI tools discover via `--help` + docs)
- **Phase ordering: KEEP** 5-phase as is (user preserved admin/ops priority)
- Add: **Phase 0** cross-cutting ergonomics refactor

---

## 1. Gap Landscape

### Data sources audited
- WS RPC: `../goclaw/pkg/protocol/methods.go` — 95 method constants
- WS registrations: `../goclaw/internal/gateway/methods/*.go` — 24 method files
- HTTP routes: `../goclaw/internal/http/*.go` — 50+ handler files
- CLI current surface: `D:/www/nextlevelbuilder/goclaw-cli/cmd/*.go` — 23 command files

### Coverage estimate
- **Covered:** ~70% (core CRUD for agents, teams, channels, skills, cron, mcp, sessions, providers, tools, traces, api-keys, memory/KG partial)
- **Missing:** ~30% — concentrated in admin/ops, migration, vault, multi-tenant, agent lifecycle extensions

---

## 2. Missing Command Groups (HOÀN TOÀN)

| # | Group | Server refs | Key operations | Tier |
|---|---|---|---|---|
| G1 | `tenants` | `tenants.*` WS + `/v1/tenants/*` HTTP | list/get/create/update + users add/remove/list + mine | 🔥 |
| G2 | `heartbeat` | `heartbeat.*` WS (8 methods) | get/set/toggle/test/logs/checklist/targets | 🔥 |
| G3 | `pair` | `device.pair.*` WS | request/approve/deny/list/revoke (mở rộng flow đăng ký device) | 🟡 |
| G4 | `vault` | `/v1/vault/*` HTTP (15 routes) | documents CRUD, links, upload, search, tree, rescan, enrichment, graph | 🔥 |
| G5 | `backup` + `restore` | `/v1/system/backup/*`, `/v1/tenant/backup/*`, `/restore` | system/tenant backup, S3, download, preflight | 🔥 |
| G6 | `oauth` | `/v1/auth/chatgpt/*`, `/v1/auth/openai/*` | start/callback/status/quota/logout cho provider pool | 🟡 |
| G7 | `packages` | `/v1/packages*`, `/v1/shell-deny-groups` | list/install/uninstall/runtimes + deny groups | 🟡 |
| G8 | `migrate` | `/v1/agents/export+import`, `/v1/teams/export+import`, `/v1/skills/export+import`, `/v1/mcp/export+import` | Export/import cho 4 domain chính | 🔥 |
| G9 | `system-configs` | `/v1/system-configs/*` | list/get/set/delete | 🟡 |
| G10 | `edition` | `GET /v1/edition` | show (public info — chỉ 1 subcommand) | 🟢 |
| G11 | `users` | `/v1/users/search` | search | 🟢 |
| G12 | `contacts` (expand) | `/v1/contacts/merge`, `/contacts/merged/{id}` | merge, list-merged (mở rộng `channels contacts`) | 🟡 |
| G13 | `quota` | `quota.usage` WS | show | 🟡 |

### Nhóm agent lifecycle (sub-group của `agents`)
| # | Subcommand | Endpoint | Tier |
|---|---|---|---|
| G14 | `agents wake <id>` | `POST /v1/agents/{id}/wake` | 🟡 |
| G15 | `agents evolution metrics/suggestions` | `/v1/agents/{id}/evolution/*` | 🟢 |
| G16 | `agents episodic list/search` | `/v1/agents/{id}/episodic/*` | 🟢 |
| G17 | `agents v3-flags get/set` | `/v1/agents/{id}/v3-flags` | 🟢 |
| G18 | `agents orchestration` | `/v1/agents/{id}/orchestration` | 🟢 |
| G19 | `agents identity <key>` | `agent.identity.get` WS | 🟡 |
| G20 | `agents wait <key>` | `agent.wait` WS | 🟢 |
| G21 | `agents sync-workspace` | `POST /v1/agents/sync-workspace` | 🟢 |
| G22 | `agents prompt-preview <id>` | `/v1/agents/{id}/system-prompt-preview` | 🟢 |
| G23 | `agents instances set-file/metadata` | `PUT/PATCH /v1/agents/{id}/instances/{user}/*` | 🟢 |
| G24 | `agents codex-pool-activity` | `/v1/agents/{id}/codex-pool-activity` | 🟢 |

---

## 3. Missing Subcommands trong Group hiện có

### `chat` (đang chỉ có send/abort)
- `chat history <agent>` → `chat.history` WS
- `chat inject <agent>` → `chat.inject` WS (inject message vào context)
- `chat session-status <agent>` → `chat.session.status` WS

### `teams`
- `teams tasks delete <teamID> <taskID>` → `teams.tasks.delete`
- `teams tasks delete-bulk <teamID>` → `teams.tasks.delete-bulk`
- `teams tasks events <teamID> <taskID>` → `teams.tasks.events`
- `teams tasks get-light <teamID> <taskID>` → `teams.tasks.get-light`
- `teams tasks active <teamID> --session=<key>` → `teams.tasks.active-by-session`
- `teams scopes <teamID>` → `teams.scopes`
- `teams events list <teamID>` → `teams.events.list` WS hoặc `GET /v1/teams/{id}/events`

### `memory` (KG chỉ có query/extract/link, thiếu full CRUD)
- `memory kg entities list/get/upsert/delete <agent>`
- `memory kg traverse <agent>`
- `memory kg stats <agent>`
- `memory kg graph <agent>` (và `--compact`)
- `memory kg dedup scan/list/merge/dismiss <agent>`
- `memory chunks <agent>` — list chunks đã index
- `memory index <agent> <path>` — trigger index
- `memory index-all <agent>` — trigger reindex
- `memory documents` (global, không agent scope) → `/v1/memory/documents`

### `mcp`
- `mcp servers reconnect <id>` → `POST /v1/mcp/servers/{id}/reconnect`
- `mcp servers test-connection` — test **trước khi** create (khác với `test <id>` hiện có test sau khi tạo)

### `skills`
- `skills install-dep <dep>` → `/v1/skills/install-dep` (single dep thay vì all)
- `skills tenant-config set/delete <id>` → `PUT/DELETE /v1/skills/{id}/tenant-config`

### `providers`
- `providers verify-embedding <id>` → `/v1/providers/{id}/verify-embedding`
- `providers codex-pool-activity <id>` → `/v1/providers/{id}/codex-pool-activity`
- `providers embedding-status` → `/v1/embedding/status`
- `providers claude-cli auth-status` → `/v1/providers/claude-cli/auth-status`

### `tools builtin`
- `tools builtin tenant-config get/set/delete <name>` → `/v1/tools/builtin/{name}/tenant-config` GET/PUT/DELETE

### `admin credentials` (đang list/create/delete) — thiếu PUT/test/user-creds/agent-grants
- `admin credentials update <id>`
- `admin credentials test <id>` (dry-run)
- `admin credentials presets`
- `admin credentials check-binary`
- `admin credentials user-credentials list/get/set/delete <credID> [userID]`
- `admin credentials agent-grants list/create/get/update/delete <credID>`

### `config`
- `config permissions list/grant/revoke` → `config.permissions.*` WS

### `channels`
- `channels pending compact` → `/v1/pending-messages/compact`
- `channels pending delete` → `DELETE /v1/pending-messages`
- `channels pending messages --group=<id>` → `/v1/pending-messages/messages`
- `channels pending groups` → `GET /v1/pending-messages`

---

## 4. Không đưa vào scope

| Item | Lý do |
|---|---|
| Zalo Personal (`zalo.personal.qr.start`, `contacts`) | User quyết định SKIP — kênh đặc thù, QR UX browser-first |
| WhatsApp (`whatsapp.qr.start`) | Tương tự Zalo |
| `browser.act/snapshot/screenshot` | CLI không phù hợp cho browser automation interactive — dùng UI riêng |
| `send` (generic method) | Unclear use case — cần confirm trước khi thêm |
| TTS (`tts.convert`) | CLI admin.go đã có status/enable/disable/providers/set-provider — convert cần audio handling phức tạp, skip hoặc tách riêng |

---

## 5. Phase Strategy (PR batching)

**Không thể đưa 20+ command group vào 1 PR.** Đề xuất chia **5 phase + Phase 0**, mỗi phase là 1 PR độc lập có thể review/merge riêng.

### Phase 0 — Cross-cutting AI Ergonomics (NEW, prerequisite)
**Rationale:** AI tools là consumer chính (~90%). Thiết lập pattern CHUẨN trước khi thêm command mới để không phải retrofit sau.

**Scope:**
1. **Structured error format** — refactor `internal/output` để khi `--output=json` và lỗi xảy ra, in JSON chuẩn:
   ```json
   {"error": {"code": "auth_expired", "message": "token expired", "details": {...}}}
   ```
2. **Exit code convention** — định nghĩa constants trong `internal/` và apply toàn bộ commands:
   - `0` success
   - `1` generic error
   - `2` authentication/authorization
   - `3` not found
   - `4` validation/bad-request
   - `5` server error (5xx)
   - `6` timeout/network
3. **`--follow` pattern helper** — helper function trong `internal/client/` cho streaming JSON lines. Apply vào `logs tail` hiện có + chuẩn bị cho phase sau.
4. ~~`--idempotency-key` flag helper~~ **DROPPED** — server chưa support (zero matches trong codebase). Track as server feature request, revisit khi server thêm.
5. **`--quiet` + output auto-detect** — default `json` khi stdout không phải TTY, `table` khi là TTY. Remove banner/tip output trong non-TTY mode.
6. **Error wrapping guidelines** — doc `docs/code-standards.md` về cách wrap server errors để giữ category info.

**Files:** `internal/output/error.go` (new), `internal/output/exit.go` (new), `internal/client/follow.go` (new), `internal/client/idempotency.go` (new). Update toàn bộ `cmd/*.go` để dùng patterns mới.

**Risk:** Medium — refactor rộng nhưng additive. Có thể break JSON schema của error nếu consumers hiện đang parse stderr text — cần verify.

**Size:** ~400 LoC new + ~300 LoC refactor. **PR này phải merge TRƯỚC Phase 1** để các phase sau kế thừa.

**Success criteria:**
- Mọi command hiện có return exit code đúng theo convention
- `goclaw agents list --output=json` và `goclaw agents list` có behavior output khác nhau theo TTY detect
- Error từ server 401 → CLI exit 2 với JSON error object
- `goclaw agents create --idempotency-key=foo ...` gửi header đúng

### Phase 1 — Admin/Ops Foundation (tier 🔥)
**Rationale:** ROI cao nhất cho ops workflow (cron, CI, monitoring).
- G1 `tenants` (list/get/create/update + users)
- G2 `heartbeat` (get/set/toggle/test/logs/checklist/targets)
- G9 `system-configs` (list/get/set/delete)
- G10 `edition` (show)
- Thêm `config permissions` subcommand

**Files mới:** `cmd/tenants.go`, `cmd/heartbeat.go`, `cmd/system_configs.go`, `cmd/edition.go`. Mở rộng `cmd/config_cmd.go`.
**Risk:** Medium — tenants có destructive ops (delete users), cần `--yes` + interactive confirm.
**Size:** ~600 LoC ước tính.

### Phase 2 — Migration (tier 🔥)
**Rationale:** Critical cho disaster recovery + môi trường migration.
- G5 `backup` (system + tenant + S3) và `restore`
- G8 `migrate` (agents/teams/skills/mcp export+import với `--preview` dry-run)

**Files mới:** `cmd/backup.go`, `cmd/restore.go`, `cmd/migrate.go`.
**Risk:** **High** — destructive restore, file streaming, token signed download. Cần:
- Backup download dùng signed token URL flow (`/backup/download/{token}`)
- Restore MUST require `--yes` + explicit confirmation string
- Preview mode trước mọi import

**Size:** ~800 LoC, heavy test coverage cần thiết.

### Phase 3 — Vault (tier 🔥)
**Rationale:** User request explicit.
- G4 `vault` đầy đủ

**Files mới:** `cmd/vault.go` (có thể split thành `cmd/vault_documents.go`, `cmd/vault_graph.go` nếu >200 LoC theo rule).

**Subcommands:**
- `vault documents list/get/create/update/delete [docID]`
- `vault links list/create/delete/batch <docID>`
- `vault upload <file>`
- `vault rescan`
- `vault tree`
- `vault search <query>`
- `vault enrichment status/stop`
- `vault graph` (render graph JSON — optional ASCII tree output)

**Risk:** Medium — CRUD phức tạp, cần schema chuẩn.
**Size:** ~700 LoC.

### Phase 4 — Agent Lifecycle + Chat + Teams extensions (tier 🟡)
**Rationale:** Fill dev persona gaps.
- G14 wake, G15 evolution, G16 episodic, G17 v3-flags, G18 orchestration, G19 identity, G20 wait, G21 sync-workspace, G22 prompt-preview, G23 instances set-file/metadata, G24 codex-pool-activity
- `chat` history/inject/session-status
- `teams` tasks delete/delete-bulk/events/get-light/active, scopes, events list
- `memory` KG full + chunks/index

**Files mở rộng:** `cmd/agents.go` (split thành `cmd/agents_lifecycle.go` nếu vượt 200 LoC), `cmd/chat.go`, `cmd/teams.go` (split: `cmd/teams_tasks.go` đã nên tách), `cmd/memory.go` (tách `cmd/memory_kg.go`).

**Risk:** Low-Medium — mostly additive, không touch core auth/session.
**Size:** ~1000 LoC. **Sẽ split thành 2 PR nếu quá lớn** (4a: agents+chat, 4b: teams+memory).

### Phase 5 — Packages + Pair + OAuth Pool + MCP/Skills/Providers extensions (tier 🟡)
- G3 `pair` (device pair management — khác với auth login flow)
- G6 `oauth` (ChatGPT/OpenAI pool)
- G7 `packages` (+ shell-deny-groups)
- G11 `users search`
- G12 `contacts merge`
- G13 `quota usage`
- `channels pending` extensions
- `mcp servers reconnect`
- `skills install-dep`, `skills tenant-config`
- `providers verify-embedding`, `codex-pool-activity`, `embedding-status`, `claude-cli auth-status`
- `tools builtin tenant-config`
- `admin credentials` extensions (update/test/user-credentials/agent-grants)

**Files:** `cmd/pair.go`, `cmd/oauth.go`, `cmd/packages.go`, `cmd/users.go`, `cmd/quota.go`. Mở rộng nhiều file hiện có.

**Risk:** Medium — pair flow đã một phần trong `auth.go`, cần tránh duplicate logic. OAuth pool có callback flow phức tạp (manual callback input).
**Size:** ~800 LoC, split thành 2 PR (5a: pair+oauth+packages+quota, 5b: extensions).

---

## 5.1 AI-value re-mapping (không đảo phase nhưng flag mức độ)

User quyết định giữ phase order hiện tại nhưng consumer primary là AI. Bảng dưới đánh dấu AI-criticality để guide **implementation depth** (AI-critical commands cần polish hơn: full JSON schema, examples trong `--help`, test coverage cao hơn).

| Phase | Command | AI-critical? | Polish level |
|---|---|---|---|
| P1 | tenants list/mine | 🔥 Yes — AI need tenant context | High |
| P1 | tenants users mgmt | 🟢 Low | Standard |
| P1 | heartbeat get/logs | 🔥 Yes — AI monitor liveness | High |
| P1 | heartbeat set/toggle/test | 🟡 Medium | Standard |
| P1 | system-configs | 🟡 Medium | Standard |
| P1 | edition | 🟢 Trivial | Minimal |
| P1 | config permissions | 🟢 Low | Standard |
| P2 | backup/restore | 🟢 Low — human ops | Standard |
| P2 | migrate (export/import) | 🔥 Yes — AI snapshot/clone agents | **High** |
| P3 | vault search/read | 🔥 Yes — AI RAG core | **High** |
| P3 | vault documents CRUD | 🟡 Medium — AI write memory | Standard |
| P3 | vault graph/links | 🟡 Medium | Standard |
| P4 | **chat history/inject/session-status** | 🔥 **CRITICAL** | **Maximum** |
| P4 | **agents wait/identity** | 🔥 **CRITICAL** | **Maximum** |
| P4 | agents evolution/episodic/v3-flags | 🟡 Medium | Standard |
| P4 | agents wake/sync-workspace | 🟡 Medium | Standard |
| P4 | teams tasks events/delete | 🔥 Yes — AI orchestrate tasks | High |
| P4 | teams scopes/active-by-session | 🔥 Yes — AI query team state | High |
| P4 | **memory KG full + chunks/index/search global** | 🔥 **CRITICAL** | **Maximum** |
| P5 | quota usage | 🔥 Yes — AI budget check | High |
| P5 | packages install/list | 🟡 Medium — AI install deps for agent | Standard |
| P5 | channels pending | 🟡 Medium | Standard |
| P5 | pair mgmt | 🟢 Low — human flow | Standard |
| P5 | oauth pool | 🟡 Medium (quota check only) | Standard |
| P5 | users search | 🟡 Medium | Standard |
| P5 | contacts merge | 🟢 Low | Standard |
| P5 | mcp reconnect | 🔥 Yes — AI self-heal | High |
| P5 | skills install-dep/tenant-config | 🟡 Medium | Standard |
| P5 | providers embedding/verify-embedding | 🟡 Medium | Standard |
| P5 | admin credentials extensions | 🟡 Medium | Standard |
| P5 | send (generic) | 🔥 Yes — AI inter-agent msg | **High** |

**"Maximum polish" nghĩa là:** full JSON schema doc trong `--help`, ≥80% line coverage, JSON examples trong `docs/`, tên flag ngắn gọn (`--agent` không phải `--agent-key-identifier`), không có stdout noise.

### `send` generic method — giờ makes sense
Trong AI-first context, `send` là primitive để AI tool push message tới agent **KHÔNG CẦN** mở chat session lifecycle (không cần track session, không cần history). Tốt cho:
- Fire-and-forget notification
- Cross-agent handoff
- Webhook-style trigger

→ Add vào **Phase 5** (hoặc bump lên P4 nếu AI orchestration use case thường xuyên).

## 6. Cross-cutting Design Considerations

### 6.1 HTTP vs WS lựa chọn
- Nhiều endpoint có **cả** WS và HTTP (vd tenants, teams, heartbeat có thể qua WS). CLI hiện tại **mix both** — chọn theo helper sẵn có:
  - WS: `ws.Call(method, params)` — ngắn, không cần path build
  - HTTP: `c.Get/Post(path, body)` — cần path construct
- **Quyết định:** Dùng WS khi method tồn tại, HTTP khi chỉ có HTTP. Nhất quán cùng 1 group.

### 6.2 Destructive operations
- **All new destructive commands** phải follow pattern hiện có: `--yes` flag + interactive confirm fallback (xem `cmd/agents.go:140` delete pattern).
- Restore/delete tenant/wipe vault: yêu cầu **typed confirmation** (vd gõ tên tenant) — KHÔNG chỉ `y/N`.

### 6.3 Output modes
- Mọi command phải support `--output table|json|yaml` theo pattern hiện có (`printer.Print()`).
- Commands trả về bulk data (vault search, episodic) nên default table, JSON khi pipe detect.

### 6.4 File handling
- Backup download, vault upload, agents export: dùng `GetRaw`/`PostRaw` với streaming, không load full vào memory.
- Signed token URL flow: download `/backup/download/{token}` KHÔNG dùng auth header — cần helper mới `GetRawNoAuth`.

### 6.5 Modularization
- Per project rule: file >200 LoC phải split. Hiện `cmd/agents.go` (416 LoC), `cmd/teams.go` (462 LoC), `cmd/skills.go` (~330 LoC) đã **vượt ngưỡng**. Phase 4 nên kèm **refactor split**:
  - `cmd/agents.go` → `agents.go` (CRUD) + `agents_links.go` + `agents_instances.go` + `agents_lifecycle.go`
  - `cmd/teams.go` → `teams.go` + `teams_members.go` + `teams_tasks.go` + `teams_workspace.go`
  - `cmd/skills.go` → `skills.go` + `skills_grants.go` + `skills_deps.go` + `skills_versions.go`

### 6.6 Test coverage
- Mỗi new command group cần `*_test.go` theo pattern hiện có (httptest server cho HTTP, WS upgrader cho WS).
- Đặc biệt Phase 2 (restore) cần integration test với fixture backup file.

### 6.7 Documentation
- `docs/codebase-summary.md` update sau mỗi phase.
- README.md command table cần cập nhật liên tục.
- API-docs command đã có — không cần duplicate.

---

## 7. Risk Assessment

| Risk | Mitigation |
|---|---|
| PR quá lớn, reviewer burn out | Strict 5-phase split, mỗi phase standalone mergeable |
| Duplicate pair logic giữa `auth login --pair` và `pair` group | Phase 5: refactor shared helper, auth.go dùng lại |
| Backup/restore phá data prod | Require typed confirmation + `--dry-run` mặc định + tài liệu rõ ràng |
| Vault schema drift server-side | Dùng JSON passthrough cho body, KHÔNG định nghĩa struct rigid |
| Endpoint method changes (WS → HTTP deprecation) | Dùng protocol constants file thay vì hard-code string |
| CLI binary size tăng đáng kể | Không đáng lo — Cobra commands share code paths, không import lib mới |
| Test coverage giảm | Bắt buộc test new commands, CI gate 60%+ coverage |

---

## 8. Success Metrics

1. **Coverage:** ≥95% server endpoints có CLI equivalent (đo bằng script count HandleFunc vs command registrations).
2. **Build:** `go build ./...` + `go vet ./...` + `go test ./...` pass trên cả 3 OS (Windows/Linux/macOS CI).
3. **Test coverage:** Mỗi phase giữ hoặc tăng coverage; new commands ≥60% line coverage.
4. **UX consistency:** Mọi command mới follow pattern output/auth/flag hiện có — pass smoke test manual checklist.
5. **Docs:** README + CLAUDE.md + codebase-summary reflect đầy đủ commands mới.

---

## 9. Next Steps

1. User **approve phase sequencing** hoặc adjust priority.
2. Tạo implementation plan chi tiết bằng `/ck:plan` cho Phase 1 (Admin/Ops Foundation).
3. Optional: bổ sung endpoint compatibility matrix (CLI version vs server version) nếu server tiếp tục drift.

---

## 10. Server Audit Findings (2026-04-14)

### ✅ Error format: SERVER-SIDE CHUẨN HÓA
- HTTP: `{"error": {"code": "...", "message": "..."}}` via `writeError` helper (`internal/http/response_helpers.go`)
- WS: richer `ErrorShape` trong `pkg/protocol/frames.go`: `code, message, details, retryable, retryAfterMs`
- Error codes (`pkg/protocol/errors.go`): `UNAUTHORIZED`, `NOT_FOUND`, `INVALID_REQUEST`, `ALREADY_EXISTS`, `RESOURCE_EXHAUSTED`, `FAILED_PRECONDITION`, `INTERNAL`, `UNAVAILABLE`, `AGENT_TIMEOUT`, `NOT_LINKED`, `NOT_PAIRED`, `TENANT_ACCESS_REVOKED`

**→ CLI CHỈ cần MAP codes → exit codes + pass-through error JSON. KHÔNG invent format mới.**

### Exit code mapping (locked)
| Server code | Exit |
|---|---|
| (success) | 0 |
| (generic) | 1 |
| `UNAUTHORIZED`, `NOT_PAIRED`, `TENANT_ACCESS_REVOKED` | 2 |
| `NOT_FOUND`, `NOT_LINKED` | 3 |
| `INVALID_REQUEST`, `FAILED_PRECONDITION`, `ALREADY_EXISTS` | 4 |
| `INTERNAL`, `UNAVAILABLE`, `AGENT_TIMEOUT` | 5 |
| `RESOURCE_EXHAUSTED`, network/timeout | 6 |

### ❌ Idempotency-Key: SERVER CHƯA SUPPORT
- Zero matches trong codebase server
- **Decision:** DROP `--idempotency-key` flag khỏi Phase 0 scope. CLI-side only implementation là noop. Track as server feature request.

### ✅ TTY-aware JSON default: BREAKING CHANGE, APPROVED
- CLI pre-1.0 → acceptable
- Default behavior sau change:
  - `--output` explicit → respect exact
  - Unset + stdout is TTY → `table`
  - Unset + stdout piped → `json`
- Document trong CHANGELOG + README notice

---

## 11. Remaining Open Questions

1. ~~`send` generic~~ RESOLVED — inter-agent primitive, P5.
2. **TTS `tts.convert`:** default skip, revisit khi có request cụ thể.
3. **Browser automation:** skip hoàn toàn.
4. ~~Phase ordering~~ RESOLVED.
5. **Migration structure:** `skills export` / `teams export` subcommand vs `migrate` group. Chốt ở P2 planning.
6. **Codex pool activity:** merge thành `codex-pool activity --agent|--provider`? Chốt ở P5.
7. **Pair naming:** `goclaw pair` primary group, `auth login --pair` giữ nguyên. No conflict.
8. ~~JSON default breaking~~ RESOLVED — breaking change OK, document.
9. ~~Idempotency server support~~ RESOLVED — dropped.
10. ~~Error format coordination~~ RESOLVED — server đã chuẩn.
