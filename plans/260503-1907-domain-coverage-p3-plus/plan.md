# Plan: Domain Coverage P3+ (Round 2 Fillers)

**Date:** 2026-05-03
**Branch:** feat/ai-first-cli-expansion
**Reference report:** `plans/reports/brainstorm-260503-1907-gap-analysis-round2.md`
**Status:** P3/P4 complete — P5 implemented pending ship; P6 remains server-blocked.

## Summary

Sau R1 (P0–P5) + R2 expansion (P0–P2), CLI đạt ~95% server coverage. R2 round phân chia 4 phase nhỏ:

| Phase | Scope | LoC | Tier | Status |
|---|---|---|---|---|
| P3 | AI-critical fillers (multi-profile, sessions compact, health, traces filter polish) | ~250 | 🔥 | complete |
| P4 | UX polish batch 1 residuals (codex-pool umbrella, api-keys rotate, config defaults, chat replay convenience, tools invoke `--args` alias) | ~250 | 🟡 | complete |
| P5 | Fillers residuals after sweep (team attachments download, evolution skill apply) | ~150 | 🟡 | implemented pending ship |
| P6 | Deferred — blocked on server FRs (traces follow, logs aggregate, providers reconnect, …) | n/a | 🟢 | server-blocked |

## Phase Files

- `phase-03-ai-critical-fillers.md`
- `phase-04-ux-polish-batch-1.md`
- `phase-05-fillers-verification-batch-2.md`
- `phase-06-server-fr-backlog.md`

## Key Dependencies

- Super-admin API parity is already merged; P4 should proceed from current `dev`.
- P3 multi-profile is complete; P4 can build on stable profile/default output behavior.
- P5 verify sweep completed 2026-05-19; most suspected gaps already exist under current command paths.
- P5 detailed execution plan: `../260520-1050-domain-coverage-p5-fillers/plan.md`.
- P6 = upstream goclaw issues, not CLI work.
- P4 validation/red-team evidence: `reports/validation-red-team-260519-p4.md`; implementation validated with `go build ./...`, `go test ./...`, and `go vet ./...`.

## Success Criteria

- Coverage ≥98% server routes (script).
- All new commands JSON envelope + exit-code compliant.
- Each phase ≤500 LoC PR; ≥60% line coverage on new code.
- CHANGELOG + README updates per phase.

## Open Questions (consolidated from report)

1. Profile naming: resolved as `profile`; legacy `auth use-context` remains.
2. Codex-pool alias sunset version?
3. Chat replay output: resolved as existing printer output from `chat.history` (JSON array in JSON mode); add JSONL/streaming only if a real workflow demands it.
4. Health schema: raw passthrough vs normalized?
5. Server FR ownership for P6 items?
6. `tts synthesize` AI use case: already covered by `cmd/tts_http.go`; not a server FR.
7. `--profile` vs `GOCLAW_OUTPUT` precedence: `--profile`/`GOCLAW_PROFILE` selects config profile; `GOCLAW_OUTPUT` still wins output format.
