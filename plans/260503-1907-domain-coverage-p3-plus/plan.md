# Plan: Domain Coverage P3+ (Round 2 Fillers)

**Date:** 2026-05-03
**Branch:** feat/ai-first-cli-expansion
**Reference report:** `plans/reports/brainstorm-260503-1907-gap-analysis-round2.md`
**Status:** P3 complete — P4/P5 sweep complete; residual implementation not-started.

## Summary

Sau R1 (P0–P5) + R2 expansion (P0–P2), CLI đạt ~95% server coverage. R2 round phân chia 4 phase nhỏ:

| Phase | Scope | LoC | Tier | Status |
|---|---|---|---|---|
| P3 | AI-critical fillers (multi-profile, sessions compact, health, traces filter polish) | ~250 | 🔥 | complete |
| P4 | UX polish batch 1 residuals (codex-pool umbrella, api-keys rotate, config defaults, chat replay/resume, tools invoke `--args` alias) | ~250 | 🟡 | not-started |
| P5 | Fillers residuals after sweep (team attachments download, evolution skill apply) | ~150 | 🟡 | not-started |
| P6 | Deferred — blocked on server FRs (traces follow, logs aggregate, providers reconnect, …) | n/a | 🟢 | server-blocked |

## Phase Files

- `phase-03-ai-critical-fillers.md`
- `phase-04-ux-polish-batch-1.md`
- `phase-05-fillers-verification-batch-2.md`
- `phase-06-server-fr-backlog.md`

## Key Dependencies

- Superseded/blocked by `plans/260518-1936-super-admin-api-parity/` for the next implementation slice. The newer plan narrows the backlog to super-admin operational parity after server `v3.12.0-beta.5`.
- P3 multi-profile may refactor `internal/config` singleton — finish before P4/P5.
- P5 verify sweep completed 2026-05-19; most suspected gaps already exist under current command paths.
- P6 = upstream goclaw issues, not CLI work.

## Success Criteria

- Coverage ≥98% server routes (script).
- All new commands JSON envelope + exit-code compliant.
- Each phase ≤500 LoC PR; ≥60% line coverage on new code.
- CHANGELOG + README updates per phase.

## Open Questions (consolidated from report)

1. Profile naming: resolved as `profile`; legacy `auth use-context` remains.
2. Codex-pool alias sunset version?
3. Chat replay output: default stdout JSONL; add `--file` only if user workflow demands it.
4. Health schema: raw passthrough vs normalized?
5. Server FR ownership for P6 items?
6. `tts synthesize` AI use case: already covered by `cmd/tts_http.go`; not a server FR.
7. `--profile` vs `GOCLAW_OUTPUT` precedence: `--profile`/`GOCLAW_PROFILE` selects config profile; `GOCLAW_OUTPUT` still wins output format.
