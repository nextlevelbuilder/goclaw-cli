# Code review — issue #17 fix (pre-ship)

Date: 2026-05-28
Phase: 3 (step 6 — pre-ship gate)
Reviewer: `code-reviewer` subagent

## Verdict

**DONE_WITH_CONCERNS** — 0 Critical, 0 Important, 4 informational. Ship-ready.

## Gate checks (all pass)

| Check | Result |
|------|--------|
| `grep -i 'eyJ\|Bearer\|sk-\|token='  cmd/testdata/` returns 0 lines | ✅ |
| All 10 trace-get tests pass + full repo suite green | ✅ |
| `cmd/traces.go` net add ≤ 100 LOC (advisory) | ⚠️ 107 — accepted trade-off for `buildSpanTree` clarity |
| No plan-artifact refs in source/comments | ✅ |

## Verified behavior

- `buildSpanTree` cycle safety: A↔B cycles silently drop (never appear in `children[""]`, so `build()` is never invoked on them). No infinite recursion possible.
- `validateTraceID` gates: 9 malformed-id cases all rejected pre-HTTP. Confirmed by `TestTracesGet_MalformedID_NoHTTPCall` (server hit count = 0 across the table).
- HTTP retry-body fix: `if attempt == 2 { break }` correctly skips the per-iteration `Close()` so the outer `defer resp.Body.Close()` handles cleanup exactly once. No leak, no double-close.
- Error mapping: 404→3, 403→2, 5xx→5, malformed→4 — all asserted, all pass.

## Informational concerns

1. **107 LOC vs 100 advisory cap.** Eliminating 7 lines would force inlining `parentOf` or removing the orphan-detection branch — clarity loss > LOC gain. Accept.
2. **Fixture `_TODO_refresh` marker.** Intentional per Phase 1 reviewer gate — refresh against `goclaw.zuey.me` before final merge.
3. **`tracesExportCmd` un-validated id (`cmd/traces.go:207`).** Pre-existing. Same path-traversal/control-char inputs that `validateTraceID` blocks flow through `export` unfiltered. Out of scope here — spawn follow-up task.
4. **README exit-code list missing 6.** Fixed inline (added "rate-limit / network-resource exhaustion → 6").

## Recommended actions (all addressed or scheduled)

| Action | Status |
|--------|--------|
| README exit-6 mention | Fixed inline this turn |
| Fixture refresh against live gateway | Plan-gated; auth-blocked smoke probe documented in Phase 1 repro |
| Harden `tracesExportCmd` with `validateTraceID` + `url.PathEscape` | Spawn follow-up task |
