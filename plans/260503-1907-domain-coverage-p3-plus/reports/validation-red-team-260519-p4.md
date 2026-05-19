# P4 Validation + Red-Team Report

**Date:** 2026-05-19
**Scope:** `phase-04-ux-polish-batch-1.md`
**Verdict:** ready for implementation after tightening scope below.

## Validation Findings

| ID | Severity | Finding | Evidence | Decision |
|---|---:|---|---|---|
| V-01 | High | `chat replay/resume` must not duplicate existing session support. CLI already supports `chat history <agent> --session=<key>` and `chat <agent> --session=<key>`. | `cmd/chat_ai_commands.go:13-70`, `cmd/chat.go:18-52`, `cmd/chat.go:208-211` | Narrow to convenience wrappers only. |
| V-02 | High | `api-keys rotate` cannot be truly atomic with current HTTP contracts; it is a composite create + revoke. Plan must define partial-failure output and tests. | `cmd/api_keys.go:51-103`, `cmd/api_keys.go:106-122` | Keep feature, call it composite, not atomic. |
| V-03 | Medium | `tools invoke` already supports structured JSON through `--params`; residual is only `--args` alias and `@file` support for JSON. | `cmd/tools_custom.go:147-175`, `cmd/tools_custom.go:191-192`, `cmd/helpers.go:63-72` | Keep tiny alias/file change. |
| V-04 | Medium | `config.defaults` exists in server and is read-only/secret-free, but CLI has no command. | `/Volumes/GOON/www/digitop/goclaw/pkg/protocol/methods.go:29-34`, `/Volumes/GOON/www/digitop/goclaw/internal/gateway/methods/config.go:38-47`, `cmd/config_cmd.go:12-113` | Add CLI passthrough. |
| V-05 | Low | `codex-pool` umbrella is an alias/UX improvement; existing agent/provider commands work. | `cmd/agents_misc.go:37-59`, `cmd/providers_codex_pool.go:10-25` | Implement top-level group, keep aliases. |

## Red-Team Notes

1. `api-keys rotate` failure mode is riskiest. If create succeeds, stdout must include the raw new key exactly once before the command attempts old-key revoke. If revoke fails, stderr/central error output should explain the old key remains active and the command should exit with code 5.
2. Avoid deprecation text on stdout for legacy `agents codex-pool-activity` and `providers codex-pool-activity`; it would break JSON pipelines. Use command `Deprecated` help metadata or stderr-only notices.
3. `chat replay` should be explicit about the agent key requirement. A session key alone is not enough for current `chat.history`, which requires `agent_key`.
4. `chat sessions resume <key>` should be framed as discoverability around existing `chat <agent> --session <key>`, not a new stateful launcher unless the command accepts/resolves an agent.
5. `config defaults` is master-scope gated server-side. Tests should assert method dispatch, not assume owner-only auth.

## Required Plan Adjustments

- Update phase text from `chat replay/resume` to `chat replay/resume convenience wrappers`.
- State that `api-keys rotate` is non-atomic composite with emit-before-revoke partial failure handling.
- State `tools invoke --args` is an alias for `--params` and supports `@file`.
- Add tests for exact command registration and JSON/stdout safety.

## Unresolved Questions

1. Should `chat replay <agent> --session <key>` be the command shape instead of `chat replay <session-key>`?
2. Should legacy codex-pool commands show deprecation notices now, or only mark `Deprecated` in Cobra help without runtime stderr?
