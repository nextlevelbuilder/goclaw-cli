# Plan: Domain Coverage P3+ (Round 2 Fillers)

**Date:** 2026-05-03
**Branch:** feat/ai-first-cli-expansion
**Reference report:** `plans/reports/brainstorm-260503-1907-gap-analysis-round2.md`
**Status:** Drafted — awaiting impl approval.

## Summary

Sau R1 (P0–P5) + R2 expansion (P0–P2), CLI đạt ~95% server coverage. R2 round phân chia 4 phase nhỏ:

| Phase | Scope | LoC | Tier | Status |
|---|---|---|---|---|
| P3 | AI-critical fillers (multi-profile, sessions compact, health, traces filter polish) | ~250 | 🔥 | not-started |
| P4 | UX polish batch 1 (codex-pool umbrella, api-keys rotate, config defaults, chat replay/resume, agents prompt-preview, tools invoke, storage size) | ~400 | 🟡 | not-started |
| P5 | Fillers + verification batch 2 (writers groups, contacts unmerge, agents instances, mcp tools, evolution patch/apply, team attachments dl) | ~250 | 🟡 | not-started |
| P6 | Deferred — blocked on server FRs (traces follow, logs aggregate, providers reconnect, …) | n/a | 🟢 | server-blocked |

## Phase Files

- `phase-03-ai-critical-fillers.md`
- `phase-04-ux-polish-batch-1.md`
- `phase-05-fillers-verification-batch-2.md`
- `phase-06-server-fr-backlog.md`

## Key Dependencies

- P3 multi-profile may refactor `internal/config` singleton — finish before P4/P5.
- P5 begins with **30-min verify sweep** (grep CLI for X1..X12 items) — likely shrinks scope.
- P6 = upstream goclaw issues, not CLI work.

## Success Criteria

- Coverage ≥98% server routes (script).
- All new commands JSON envelope + exit-code compliant.
- Each phase ≤500 LoC PR; ≥60% line coverage on new code.
- CHANGELOG + README updates per phase.

## Open Questions (consolidated from report)

1. Profile naming: `profile` vs `context`?
2. Codex-pool alias sunset version?
3. Chat replay output: stdout JSONL vs file?
4. Health schema: raw passthrough vs normalized?
5. Server FR ownership for P6 items?
6. `tts synthesize` AI use case?
7. `--profile` vs `GOCLAW_OUTPUT` precedence?
