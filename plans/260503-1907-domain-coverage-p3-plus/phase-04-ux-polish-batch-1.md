# Phase 4 — UX Polish Batch 1

**Priority:** 🟡 medium
**Status:** complete
**Estimated LoC:** ~250 (excl. tests)
**Estimated PR size:** ≤ 500 LoC incl. tests
**Depends on:** P3 (multi-profile stable)

## Context Links

- Gap analysis: `plans/reports/brainstorm-260503-1907-gap-analysis-round2.md` § 5 (P4)
- Validation/red-team: `reports/validation-red-team-260519-p4.md`

## Overview

Composite/UX commands wrapping endpoints already shipped server-side. No server FRs needed. Sweep on 2026-05-19 found several planned items already covered.

## Scope

| # | Command | Source | Type |
|---|---|---|---|
| P2 | `codex-pool activity --agent=… \| --provider=…` | unify `agents codex-pool-activity` + `providers codex-pool-activity` | residual |
| D3 | `api-keys rotate <id>` | composite: create-new + emit raw + revoke-old | residual |
| D6 | `config defaults` | WS `config.defaults` (`pkg/protocol/methods.go` ConfigDefaults) | residual |
| E3 | `chat replay <agent> --session=<key>` | convenience wrapper over existing `chat history --session` | residual |
| E1 | `chat sessions resume <agent> --session=<key>` | discoverability wrapper over existing `chat --session` | residual |
| X1 | `agents prompt-preview <id>` | `cmd/agents_admin.go` | covered |
| X6 | `tools invoke <name>` | `cmd/tools_custom.go`; residual is `--args=@file.json` alias only | partial |
| X7 | `storage size` | `cmd/storage.go` | covered |

## Related Code Files

### Modify

- `cmd/api_keys.go` — add `rotate`
- `cmd/chat.go` / `cmd/chat_ai_commands.go` — add replay/resume convenience wrappers without duplicating existing session logic
- `cmd/tools_custom.go` — add `--args` alias / `@file` support for existing `tools invoke`
- `CHANGELOG.md`, `docs/codebase-summary.md`

### Create

- `cmd/codex_pool.go` — umbrella group
- `cmd/config_defaults.go`
- companion `_test.go` per file

### Delete

- none (keep `agents codex-pool-activity` + `providers codex-pool-activity` as deprecated aliases)

## Implementation Steps

1. `cmd/codex_pool.go`: register top-level `codex-pool activity` + flag dispatch (`--agent` vs `--provider`). Mark old commands deprecated in Long help.
2. `cmd/api_keys.go::rotate`: orchestrate create → emit raw → revoke-old in single command, JSON output of new key. Treat as non-atomic composite; preserve machine-readable partial failure output.
3. `cmd/config_defaults.go`: WS `config.defaults`, raw passthrough.
4. Add chat wrappers only: `replay` calls existing `chat.history` with `agent_key` + `session_key`; `resume` dispatches to existing chat flow with `--session`.
5. Extend existing `tools invoke <name>` with `--args=@file.json` or literal JSON alias. Reject simultaneous conflicting `--params` and `--args`.
6. Keep deprecated codex-pool aliases stdout-clean; runtime notices must not break JSON output.
7. Tests for each. Composites: assert sequence of HTTP/WS calls via httptest.
8. Docs sync.

## Todo List

- [x] cmd/codex_pool.go umbrella + alias deprecation
- [x] cmd/api_keys.go rotate composite
- [x] cmd/config_defaults.go
- [x] chat replay convenience wrapper
- [x] chat sessions resume convenience wrapper
- [x] cmd/agents prompt-preview
- [x] cmd/tools invoke `--args` alias + `@file` (existing `--params` + `--param` already covered)
- [x] cmd/storage.go (size)
- [x] tests per command
- [x] CHANGELOG + docs

## Success Criteria

- `goclaw codex-pool activity` works for both --agent and --provider; legacy commands print deprecation notice (stderr only, doesn't break JSON pipe on stdout).
- `goclaw api-keys rotate <id>` returns new key once + attempts old revoke; if revoke fails, output indicates partial state with exit 5.
- `goclaw config defaults` calls WS `config.defaults` and returns existing printer output.
- `goclaw chat replay <agent> --session=<key>` outputs transcript suitable for `| jq` pipeline.
- `goclaw chat sessions resume <agent> --session=<key>` reuses existing chat send/interactive path.
- `goclaw tools invoke <name> --args=@payload.json` returns server response, exit 0 on success.
- All new commands respect `--output`, `--quiet`, `--yes` contracts.
- ≥ 60% line coverage on new code.

## Risk Assessment

| Risk | Mitigation |
|---|---|
| `api-keys rotate` partial failure (new created, old revoke failed) | Emit raw new key immediately after create, then return structured partial-failure error with `new_key_id`, `old_key_id`, `old_revoke_status`, and exit 5 if revoke fails |
| `tools invoke` argument injection | Server enforces auth + validates payload; CLI only passes through |
| `chat replay` large transcript blows memory | Current wrapper reuses `chat.history`; add pagination/streaming later only if transcript size becomes a real workflow issue |
| Deprecated codex-pool aliases confuse users | Help text + CHANGELOG note + sunset version (open question — see plan.md Q2) |
| Runtime deprecation notice breaks JSON pipelines | Prefer Cobra `Deprecated` metadata or stderr-only notices; never write notices to stdout |

## Security Considerations

- `api-keys rotate` raw key visible only once via stdout; suggest `--output=json` and `jq -r .key` for capture.
- `tools invoke` requires authenticated session; CLI does not bypass.

## Validation Log

- 2026-05-19 validate/red-team complete. Evidence and decisions recorded in `reports/validation-red-team-260519-p4.md`.
- `config.defaults` contract verified in sibling server: `protocol.MethodConfigDefaults = "config.defaults"` and handler registered as read-only, master-scope gated.
- Existing CLI session support verified; P4 must add discoverability wrappers, not a second session implementation.
- 2026-05-19 implementation complete. Verified with `/usr/local/go/bin/go build ./...`, `/usr/local/go/bin/go test ./...`, and `/usr/local/go/bin/go vet ./...`.

## Next Steps

P5 = final fillers: team attachment download + evolution skill apply.
