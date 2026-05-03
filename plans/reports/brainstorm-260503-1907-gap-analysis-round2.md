# Brainstorm: Gap Analysis Round 2 (post-P0..P5 + P0..P2 expansion)

**Date:** 2026-05-03
**Scope:** Re-audit goclaw-cli vs goclaw server after R1 (P0–P5) + R2 (P0–P2 expansion). Identify residual gaps, cluster into P3+ batches.
**Reference:** R1 report `brainstorm-260414-2231-missing-commands-gap-analysis.md`. Latest CLI commit: `4435fff`.

---

## 1. Counts (math shown)

| Surface | R1 | R2 (now) | Delta |
|---|---|---|---|
| WS method constants (`pkg/protocol/methods.go`) | 95 | **127** | +32 (heartbeat 8, hooks 7, agents.files 3, config.permissions 3, channel-instances 5, agents.links 4, voices 2, …) |
| WS registrations (`router.Register` calls) | ~95 | **102** unique | +7 net |
| HTTP routes (unique `verb path`) | "50+" | **258** | full enumeration this round |
| CLI source files in `cmd/` | 23 | ~70 source (+ ~30 test) | +47 source |
| CLI subcommands (unique `Use:"…"`) | ~120 | **346** | +226 |

CLI now covers nearly all CRUD surface. Remaining gaps concentrated, not domain-wide.

---

## 2. Coverage Delta — what landed since R1

**Closed groups:** tenants, heartbeat, system-configs, edition, vault (full), backup/restore (+S3), migrate (4 domains), pair, oauth, packages (+github-releases), users search, quota, contacts (incl merge/merged), config permissions, hooks, agents.files, agents lifecycle (wake/wait/identity/cancel-summon/skills/sync-workspace/orchestration/codex-pool-activity/v3-flags/evolution/episodic/instances), chat (history/inject/session-status), teams (tasks delete/delete-bulk/events/get-light/active/scopes, events list, members, workspace upload/move), memory KG (full + dedup + graph), MCP (servers test-connection/reconnect/grants/requests/user-credentials), skills (install-dep/tenant-config), providers (verify-embedding/claude-cli auth-status/embedding-status/codex-pool-activity), tools builtin tenant-config, admin credentials full, channels (writers list/add/remove, pending groups/messages/delete/compact), traces (list/get/export + costs summary), usage (timeseries/breakdown), tts (test-connection), voices (list/refresh), files (sign), send, hooks-test-runner.

**Estimated coverage:** ~95% server endpoints have CLI equivalent (vs ~70% R1).

---

## 3. Residual Gap Inventory

### 3.A Observability deep — P3a

| # | Missing | Server source | Tier | Notes |
|---|---|---|---|---|
| O1 | `traces follow` (live tail) | SSE infra `internal/http/sse.go:37` but **no `/v1/traces/follow` route** (`internal/http/traces.go:32-35`) | 🟢 | Server FR first. Skip. |
| O2 | `traces replay` | no endpoint | 🟢 | CLI synth. Defer. |
| O3 | `logs aggregate` | no endpoint; `logs.tail` live only (`internal/gateway/methods/logs.go:24`) | 🟢 | Defer to server. |
| O4 | `audit list/query` | no query route. Audit = fire-emit (`internal/http/audit.go:11`). `/v1/activity` covered as `goclaw activity` (`cmd/admin.go:133`) | OK | Already covered. |
| O5 | `health` probe | no `/v1/health`. WS `MethodHealth = "health"` (`methods.go:46`) unmapped to HTTP | 🟡 | Add `goclaw health` → WS `health`. Trivial. |
| O6 | `traces` filter polish (`--since`, `--agent`, `--status`, `--root-only`) | server `GET /v1/traces` accepts query params (`internal/http/traces.go:32`) | 🔥 | Polish, not gap. AI tools want clean filter flags. |
| O7 | `costs` detail | `/v1/costs/summary` covered as `usage costs` | OK | |

**Net P3a:** O5 health + O6 traces filter polish.

### 3.B Provider/runtime polish — P3b

| # | Missing | Server source | Tier | Notes |
|---|---|---|---|---|
| P1 | `providers reconnect <id>` | no server endpoint (grep returns 0). MCP has reconnect (covered) | 🟢 | Server FR. Defer. |
| P2 | Unified `codex-pool` group | `agents codex-pool-activity` + `providers codex-pool-activity` exist as 2 separate commands. R1 Q6 unresolved | 🟡 | Add `goclaw codex-pool activity --agent=… \| --provider=…` umbrella, keep aliases. |
| P3 | `providers refresh-models` | `GET /v1/providers/{id}/models` covered. No "refresh" server-side | OK | |
| P4-P6 | verify-embedding / claude-cli auth-status / embedding-status | covered | OK | |
| P7 | OAuth state listing | no endpoint | 🟢 | Skip. |

**Net P3b:** P2 codex-pool umbrella alias only.

### 3.C Channels / contacts mở rộng — P3c

| # | Missing | Server source | Tier | Notes |
|---|---|---|---|---|
| C1 | `channels writers groups <id>` | `GET /v1/channels/instances/{id}/writers/groups` (`internal/http/channel_instances.go:70`). CLI has list/add/remove only | 🟡 | Real gap. Quick add. |
| C2 | `channels writers test` | no endpoint | 🟢 | Server FR. Defer. |
| C3 | `contacts merge/merged` | `internal/http/contact_merge_handlers.go`. CLI: `cmd/channels_contacts.go` has merge + merged | OK | |
| C4 | `contacts unmerge` | `POST /v1/contacts/unmerge` — verify CLI | 🟡 | If missing add ~10 LoC. |
| C5 | `contacts resolve` | covered | OK | |
| C6 | `channels pending` (4 actions) | covered (`cmd/channels_pending.go`) | OK | |

**Net P3c:** C1 writers groups, C4 contacts unmerge (verify).

### 3.D Config permissions / profiles / key rotation — P3d

| # | Missing | Server source | Tier | Notes |
|---|---|---|---|---|
| D1 | `config permissions` | `internal/gateway/methods/config_permissions.go:40-42`. Covered (`cmd/config_cmd.go`) | OK | |
| D2 | Multi-profile (`~/.goclaw/profiles/*`, `--profile=foo`) | CLI-side only, no server dep | 🔥 | AI-critical gap. AI agents juggle 2+ tenants/envs. Pure CLI: `goclaw profile use/list/create/delete` + `--profile` flag. |
| D3 | `api-keys rotate <id>` | server has revoke + create. No atomic rotate | 🟡 | CLI composite: create-new → emit raw → revoke-old. |
| D4 | Token refresh / pair re-auth | `auth login --pair` + pair group exist | OK | |
| D5 | `config schema` | covered | OK | |
| D6 | `config defaults` | `MethodConfigDefaults` (`config.go:46`). NOT in CLI Use list | 🟡 | Add `config defaults` (read-only). |

**Net P3d:** D2 multi-profile (🔥), D3 api-keys rotate (🟡), D6 config defaults (🟡).

### 3.E Chat lifecycle — P3e

| # | Missing | Server source | Tier | Notes |
|---|---|---|---|---|
| E1 | `chat sessions resume <key>` | no endpoint. UX wrapper around `chat send --session-key=<key>` | 🟢 | Add as alias. |
| E2 | `chat sessions branch <key>` | no fork-session method | 🟢 | Server FR. Defer. |
| E3 | `chat replay <key>` | composite: `sessions preview` + `chat history` | 🟡 | Compose. |
| E4 | `sessions compact <key>` | `MethodSessionsCompact` (`sessions.go:33`) exists. NOT in CLI (`cmd/sessions.go` has list/preview/delete/reset/label only) | 🔥 | Real gap. AI critical — context window pressure. |
| E5 | `chat history --tail/--follow` | one-shot WS, no delta push | 🟢 | Defer. |
| E6-E8 | inject / session-status / abort | covered | OK | |

**Net P3e:** E1 resume alias (trivial), E3 replay composite (🟡), E4 sessions compact (🔥).

### 3.F Other found gaps (outside the 5 buckets)

| # | Missing | Server source | Tier |
|---|---|---|---|
| X1 | `agents prompt-preview <id>` | `GET /v1/agents/{id}/system-prompt-preview` (`internal/http/agents.go:155`). Not in CLI Use list | 🟡 |
| X2 | `agents export preview` standalone | `GET /v1/agents/{id}/export/preview` — verify | 🟡 |
| X3 | `agents instances list/files` | `GET /v1/agents/{id}/instances`, `…/instances/{userID}/files`. CLI has set-file/metadata; list/get may be partial | 🟡 |
| X4 | `mcp servers tools <id>` | `GET /v1/mcp/servers/{id}/tools` — verify | 🟡 |
| X5 | `tts synthesize` | `POST /v1/tts/synthesize` — audio handling. R1 deferred | 🟢 |
| X6 | `tools invoke <name>` | `internal/http/tools_invoke.go`. CLI tools group is config-only | 🟡 |
| X7 | `storage size` | `GET /v1/storage/size` — verify | 🟡 |
| X8 | `evolution suggestions patch` | `PATCH /v1/agents/{id}/evolution/suggestions/{sid}` — verify | 🟡 |
| X11 | `team attachments download` | `GET /v1/teams/{teamId}/attachments/{aid}/download` — verify | 🟡 |
| X12 | `evolution skill apply` | `internal/http/evolution_skill_apply.go` — verify | 🟡 |

---

## 4. AI-Priority Tier Summary

- 🔥 **3 critical**: D2 multi-profile, E4 sessions compact, O6 traces filter polish.
- 🟡 **~12 medium**: codex-pool umbrella, writers groups, contacts unmerge, config defaults, api-keys rotate, agents prompt-preview, instances list, mcp tools, tools invoke, storage size, team attachments, evolution suggestions patch, replay composite.
- 🟢 **~10 deferred** (server gap): traces follow/replay, logs aggregate, providers reconnect, writers test, chat resume/branch/history-follow, tts synthesize.
- ✅ **Corrections from R1 backlog**: O4 audit (= activity), C3 contacts merge, P3-P5 providers — already shipped.

---

## 5. Phase Strategy (P3+) — refined into 4 phases

### P3 — AI-critical fillers (🔥 only) — 1 PR, ~250 LoC
- D2 multi-profile — `~/.goclaw/profiles/<name>/{config.yaml,token}` + `goclaw profile {use|list|create|delete|current}` + `--profile` global flag.
- E4 `sessions compact <key>` — wire WS `sessions.compact` (1-line method call).
- O5 `health` — wire WS `health` method.
- O6 traces filter polish — `--since`, `--agent`, `--status`, `--root-only`, `--limit` flags on `traces list`.

Risk: profile mgmt touches `internal/config` singleton; token store needs profile namespace.

### P4 — UX polish (🟡 batch 1) — 1 PR, ~400 LoC
Composite/UX commands wrapping existing endpoints. No server prereqs.
- P2 codex-pool umbrella alias group.
- D3 `api-keys rotate <id>` (composite).
- D6 `config defaults`.
- E3 `chat replay <key>` (composite).
- E1 `chat sessions resume <key>` (alias).
- X1 `agents prompt-preview`.
- X6 `tools invoke <name>`.
- X7 `storage size`.

### P5 — fillers & verifications (🟡 batch 2) — 1 PR, ~250 LoC
- C1 channels writers groups, C4 contacts unmerge (verify), X3 agents instances list/files, X4 mcp servers tools, X8 evolution suggestions patch, X11 team attachments download, X12 evolution skill apply.

### P6 — deferred (blocked on server FRs)
File issues in `goclaw` repo, do NOT ship CLI stubs:
- O1 traces follow (SSE), O2 traces replay, O3 logs aggregate, P1 providers reconnect, C2 writers test, E2 chat branch, E5 chat history follow, X5 tts synthesize.

YAGNI rationale: stubs without server endpoints = broken UX + maintenance debt.

---

## 6. Cross-cutting

- Verification sweep before P5 (30-min grep pass on X-prefix items) — likely shrinks P5.
- D2 multi-profile is large enough to merit its own PR if config refactor explodes.
- Profile token storage: keyring namespace `goclaw-cli:<profile>:token`.
- Codex-pool umbrella keeps old commands as aliases for 1 minor version, deprecation note in `--help`.
- Modularization: `cmd/sessions.go` and `cmd/traces.go` both small enough to absorb additions without splitting.

---

## 7. Risks

| Risk | Mitigation |
|---|---|
| Multi-profile breaks single-config users | Default profile `default` auto-created; existing `~/.goclaw/config.yaml` migrated transparently |
| Profile token leakage | Distinct keyring keys per profile |
| Codex-pool alias confusion | Keep old commands working, deprecation note |
| Sessions compact destructive | Server-side ownership check + CLI `--yes` flag |
| Server FR queue ignored → P6 stalls | File issues NOW with concrete payload examples |
| `tools invoke` security | Server enforces auth; CLI passes through; `--yes` for destructive tools |
| Health endpoint reveals tenant info | Match server response 1:1, no embellishment |

---

## 8. Success Metrics

1. Coverage ≥98% server routes wrapped (script: `grep HandleFunc` vs CLI route refs).
2. AI ergonomics validated: profile + sessions compact + traces filters via mock CI running 2 profiles parallel.
3. PR size ≤500 LoC (excl. tests) per phase.
4. New commands ≥60% line coverage.
5. README profile section + CHANGELOG entry per PR.

---

## 9. Open Questions

1. Profile vs context naming — recommend `profile` (aws-cli precedent, less collision with kubectl).
2. Codex-pool alias deprecation timeline — keep both indefinitely or sunset in v0.X? User decision.
3. `chat replay` output mode — file (`--output=file.jsonl`) or stdout JSONL? Recommend stdout for pipe-friendliness.
4. `health` schema — match WS response 1:1 or normalize (uptime sec vs human)? Recommend 1:1 raw.
5. Server FR ownership — who files goclaw issues for P6 items, this CLI maintainer or upstream PM? Action before P6.
6. P5 verification sweep — 30-min grep pass to confirm X-prefix true gaps before scoping.
7. `tts synthesize` — any AI use case requiring audio bytes from CLI? If only test/QA, stay deferred.
8. `--profile` vs `GOCLAW_OUTPUT` env precedence — recommend env > profile (operator forces JSON in CI regardless of profile).
