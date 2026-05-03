# Phase 4 — UX Polish Batch 1

**Priority:** 🟡 medium
**Status:** not-started
**Estimated LoC:** ~400 (excl. tests)
**Estimated PR size:** ≤ 500 LoC incl. tests
**Depends on:** P3 (multi-profile stable)

## Context Links

- Gap analysis: `plans/reports/brainstorm-260503-1907-gap-analysis-round2.md` § 5 (P4)

## Overview

Composite/UX commands wrapping endpoints already shipped server-side. No server FRs needed. Most are 1-file each.

## Scope

| # | Command | Source | Type |
|---|---|---|---|
| P2 | `codex-pool activity --agent=… \| --provider=…` | unify `agents codex-pool-activity` + `providers codex-pool-activity` | umbrella alias |
| D3 | `api-keys rotate <id>` | composite: create-new + emit raw + revoke-old | composite |
| D6 | `config defaults` | WS `config.defaults` (`pkg/protocol/methods.go` ConfigDefaults) | direct |
| E3 | `chat replay <session-key>` | composite: `sessions preview` + `chat history` | composite |
| E1 | `chat sessions resume <key>` | UX wrapper around `chat send --session-key=<key>` | alias |
| X1 | `agents prompt-preview <id>` | `GET /v1/agents/{id}/system-prompt-preview` | direct |
| X6 | `tools invoke <name> [--args=…]` | `internal/http/tools_invoke.go` | direct |
| X7 | `storage size` | `GET /v1/storage/size` | direct |

## Related Code Files

### Modify

- `cmd/api_keys.go` — add `rotate`
- `cmd/chat.go` — add `replay` + `sessions resume`
- `cmd/agents.go` or `cmd/agents_misc.go` — add `prompt-preview`
- `cmd/tools.go` — add `invoke` (read-write surface, not config-only)
- `CHANGELOG.md`, `docs/codebase-summary.md`

### Create

- `cmd/codex_pool.go` — umbrella group
- `cmd/config_defaults.go`
- `cmd/storage.go`
- companion `_test.go` per file

### Delete

- none (keep `agents codex-pool-activity` + `providers codex-pool-activity` as deprecated aliases)

## Implementation Steps

1. `cmd/codex_pool.go`: register top-level `codex-pool activity` + flag dispatch (`--agent` vs `--provider`). Mark old commands deprecated in Long help.
2. `cmd/api_keys.go::rotate`: orchestrate create → emit raw → revoke-old in single command, JSON output of new key.
3. `cmd/config_defaults.go`: WS `config.defaults`, raw passthrough.
4. `cmd/storage.go`: HTTP GET `/v1/storage/size`, table + JSON.
5. Extend `cmd/chat.go` with `replay <key>` (composite preview+history, stream JSONL to stdout).
6. Add `cmd/chat.go::sessions resume <key>` as alias for `chat send --session-key=<key>` (read stdin/--message body).
7. Extend agents group with `prompt-preview <id>` — GET endpoint, `--format=raw|markdown`.
8. Extend tools group with `invoke <name>` — POST body via `--args=@file.json` or literal JSON.
9. Tests for each. Composites: assert sequence of HTTP calls via httptest.
10. Docs sync.

## Todo List

- [ ] cmd/codex_pool.go umbrella + alias deprecation
- [ ] cmd/api_keys.go rotate composite
- [ ] cmd/config_defaults.go
- [ ] cmd/storage.go (size)
- [ ] cmd/chat.go replay composite
- [ ] cmd/chat.go sessions resume alias
- [ ] cmd/agents prompt-preview
- [ ] cmd/tools invoke (with @file + literal JSON args)
- [ ] tests per command
- [ ] CHANGELOG + docs

## Success Criteria

- `goclaw codex-pool activity` works for both --agent and --provider; legacy commands print deprecation notice (stderr only, doesn't break JSON pipe on stdout).
- `goclaw api-keys rotate <id>` returns new key once + revokes old atomically; if revoke fails, output indicates partial state with exit 5.
- `goclaw chat replay <key>` outputs JSONL transcript suitable for `| jq` pipeline.
- `goclaw tools invoke <name> --args=@payload.json` returns server response, exit 0 on success.
- All new commands respect `--output`, `--quiet`, `--yes` contracts.
- ≥ 60% line coverage on new code.

## Risk Assessment

| Risk | Mitigation |
|---|---|
| `api-keys rotate` partial failure (new created, old revoke failed) | Output structured JSON `{"new_key":..., "old_revoke_status":"failed", "old_key_id":...}` + exit 5; user can manually revoke |
| `tools invoke` argument injection | Server enforces auth + validates payload; CLI only passes through |
| `chat replay` large transcript blows memory | Stream JSONL line-by-line, do not buffer |
| Deprecated codex-pool aliases confuse users | Help text + CHANGELOG note + sunset version (open question — see plan.md Q2) |

## Security Considerations

- `api-keys rotate` raw key visible only once via stdout; suggest `--output=json` and `jq -r .key` for capture.
- `tools invoke` requires authenticated session; CLI does not bypass.

## Next Steps

P5 = verify sweep + final fillers.
