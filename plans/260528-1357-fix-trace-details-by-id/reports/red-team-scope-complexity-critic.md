# Red-Team: Scope & Complexity Critic — Issue #17 Plan

## Verdict: TRIM

The plan attacks a real bug with the right instincts (TDD, structured rendering, central error mapping) but inflates work by ~40-50%. Phase 1 over-tests for a fix-before-known-shape; Phase 2 carries a deferred decision (`unmarshalMapStrict` vs inline) into implementation; presentation polish (event truncation, defensive helpers) is masquerading as bug-fix scope. Exit-code mapping is also mostly free — `MapServerCode` + `MapHTTPStatus` already cover all 4 categories via the existing fallback (verified at `internal/output/exit.go:18-44`). Trim, don't reject.

---

## Findings

### Critical — Pre-fixture test count inflates Phase 1

- **Item:** Phase 1 step 3 — 7 red test cases written before the fixture is captured from a live gateway.
- **Disposition:** TRIM
- **Rationale:** Three tests (`HappyPath_Table`, `MalformedID`, `MalformedResponse`) assert behavior of code that doesn't exist yet. They're Phase 2 acceptance tests squatting in Phase 1, which is fine for TDD but inflates Phase 1's "this is the repro" mission. The actual repro need is: (a) one test proving the current JSON-dump-in-TTY symptom, and (b) the captured fixture. The other 5 can be added at the top of Phase 2 (one commit) without breaking TDD discipline.
- **Leaner version:** Phase 1 writes 2 tests (`TestTracesGet_PathAndMethod`, `TestTracesGet_TableMode_DumpsRawJSON_RED`) + the fixture + the classification verdict. Move the other 5 to Phase 2 step 1.

### Critical — `unmarshalMapStrict` decision deferred to implementation

- **Item:** Phase 2 line 63: "add `unmarshalMapStrict(data) (map[string]any, error)`... OR replace usage at the call site only. Decision in step 1."
- **Disposition:** CUT the helper, KEEP the inline approach.
- **Rationale:** `unmarshalMap` is called in 6 places across `cmd/` (traces, usage summary/costs/timeseries/breakdown). Adding `unmarshalMapStrict` is either a one-off (only `tracesGetCmd` uses it — confusing) or a project-wide refactor (out of scope). The bug only requires checking the unmarshal error in this one command. Inline is 4 lines. Leaving the decision to "step 1" forces re-planning at implementation time, which is exactly the anti-pattern that planning is supposed to prevent.
- **Leaner version:** Phase 2 mandates: replace `printer.Print(unmarshalMap(data))` with an inline `json.Unmarshal` + error check. No new helper. Three lines diff.

### Important — Separate `cmd/traces_render.go` file at 80-120 LOC

- **Item:** Phase 2 — `cmd/traces_render.go` with `renderTraceDetail` + `buildSpanTree` + defensive `str/safeFloat` helpers.
- **Disposition:** TRIM
- **Rationale:** `str()` already exists at `cmd/helpers.go:56`. `tracesListCmd` (12 LOC inline render via `output.NewTable`) is the established pattern; this plan would break that pattern for one command. If span-tree rendering stays in scope (see next finding), it justifies a helper, but 80-120 LOC is double what's needed. The header card is 3-5 `printer.Print(output.NewTable(...))` calls. Span tree is `PrintTreeRoot(buildSpanTree(spans), os.Stdout)` — that's the only real helper that earns its keep.
- **Leaner version:** Inline the header in `tracesGetCmd.RunE` (~15 LOC). Extract only `buildSpanTree(spans []any) output.TreeNode` to `cmd/traces.go` as a package-level func, not a new file. Target: total addition ~40 LOC across one file.

### Important — Span-tree rendering vs issue #17 acceptance criteria

- **Item:** Phase 2 architecture — span tree via `output.PrintTree`.
- **Disposition:** KEEP (with constraint)
- **Rationale:** Re-read of issue #17 verbatim: "including relevant metadata, root/child events, messages, tool calls, status, timing". "Root/child events" reads as hierarchical structure — span tree IS a reasonable interpretation, not scope creep. Also the existing `tracesGetCmd.Short` is literally "Get trace with span tree" (`cmd/traces.go:62`), so this is a documented promise the code never delivered. KEEP, but cap aggressively: ~20 LOC for `buildSpanTree`, degrade gracefully if no spans/no parent_id field.
- **Leaner version:** Build adjacency only if `spans` field present and has >1 entry. If flat or absent, skip the tree block and print events list directly. No "deep nesting / cyclic refs" defensive code — the server controls this payload, and if it ever sends a cycle, that's a server bug worth crashing on (or trivially guard with depth limit ≤32, ~3 LOC).

### Important — Events truncation logic

- **Item:** Phase 2 step 2 — "first 5 + last 1 with `... (N-6 more) ...` separator if N > 10".
- **Disposition:** CUT
- **Rationale:** This is presentation polish, not bug-fix scope. The bug is "unusable output". A simple `EVENTS (N=12):` header + full list, one per line, is human-readable; if the user wants pagination they pipe to `less`. The truncation logic adds branching, off-by-ones, and a test surface that doesn't earn its weight for a P2 bugfix. If a real trace has 5,000 events, that's a separate UX issue worth filing.
- **Leaner version:** Print `EVENTS (N=<count>):` header then list each event on its own line as `<timestamp> <type> <one-line-summary>`. ~8 LOC. Done.

### Important — 4 distinct exit codes — already free, "verify" step is over-scoped

- **Item:** Phase 2 step 4 — "Verify error-code mapping in `internal/output/error.go`... extend `MapServerCode` if `TRACE_NOT_FOUND` / `PERMISSION_DENIED` not yet mapped."
- **Disposition:** KEEP the requirement (it's locked per CLAUDE.md), CUT the "extend" hedging.
- **Rationale:** Verified at `internal/output/exit.go:18-44` — `serverCodeMap` maps `NOT_FOUND→3`, `TENANT_ACCESS_REVOKED→2`, `INVALID_REQUEST→4`, `INTERNAL/UNAVAILABLE→5`. And `apiErrorCodeForStatus` at `cmd/helpers.go:152-172` already translates HTTP 403→`TENANT_ACCESS_REVOKED`, 404→`NOT_FOUND`, 400/422→`INVALID_REQUEST`, 5xx→`INTERNAL`. The server doesn't need to emit `TRACE_NOT_FOUND` — bare HTTP status is already correctly mapped through two layers. The 4-category exit-code requirement is met today by `apiErrorFromRawBody` + `MapServerCode`. The remaining gap is **client-side malformed-id validation** (`MalformedID` → ExitValidation), which has to be added because the server never sees the request.
- **Leaner version:** Phase 2 step 4 becomes: "Add client-side id validation in `tracesGetCmd.RunE`: reject empty/whitespace id with an error that surfaces `INVALID_REQUEST` (or `output.ExitValidation` directly). No `internal/output/*` edits expected; if test fails for a specific server code, add to `serverCodeMap` at that point."

### Important — Phase 1 effort (2-3h) is overstated for the trimmed scope

- **Item:** Phase 1 — effort `"2-3h"` covers smoke probe + fixture + 7 red tests + repro report + classification.
- **Disposition:** TRIM
- **Rationale:** With the test count trim above, Phase 1 = smoke probe (15 min if gateway up) + fixture capture + 2 tests + report. Realistic: 45-75 min. The "2-3h" estimate suggests the planner padded for the bloated test count.
- **Leaner version:** Re-estimate Phase 1 at 1-1.5h after the test trim. Frees attention budget for Phase 2.

### Minor — Upstream issue filing is unforked

- **Item:** Phase 3 step 5 — "if Phase 1 classification was server-side or hybrid, file upstream issue."
- **Disposition:** KEEP, but fork plan branches.
- **Rationale:** Scout of `cmd/traces.go:68` shows the path `/v1/traces/<id>` is straightforward — a server-side root cause would most likely be tenant-filter or payload-shape, both of which still need a CLI render fix to handle the response shape. Likely outcome: CLI-only or hybrid. If pure server-side (gateway returns 404 for valid IDs), the CLI render work is wasted but the test infrastructure isn't. Don't fork phases, but add a branch point at end of Phase 1: "If pure server-side: Phase 2 reduces to error-mapping + smoke test only; render scope deferred."
- **Leaner version:** Add a one-liner at end of Phase 1 Success Criteria: "Classification verdict drives Phase 2 scope reduction — record verdict in plan.md before starting Phase 2."

### Minor — `docs/codebase-summary.md` update trigger undefined

- **Item:** Phase 3 — "add a bullet... if `cmd/traces_render.go` was created" / "if non-trivial".
- **Disposition:** TRIM
- **Rationale:** Vague triggers create skip-the-step inertia. Either the helper is a new public-ish surface worth documenting, or it isn't. After the trim above (no separate file, ~40 LOC inline + one tiny helper), it's not.
- **Leaner version:** Cut the `codebase-summary.md` step entirely. The CHANGELOG entry is enough. If a future planner needs to know about the helper, grep finds it.

### Minor — Phase 3 step 1 hand-edits plan.md status

- **Item:** Phase 3 step 1 — "mark phase 1, 2 as completed in `plan.md`".
- **Disposition:** TRIM
- **Rationale:** The ck-stack has `/ck:plan-check` / `ck plan check` (referenced in checklists) for status sync. Hand-editing plan.md drift is exactly what that tool exists to prevent. Not catastrophic, but inconsistent with project tooling.
- **Leaner version:** Replace with: "Run `/ck:plan-check --sync` (or hand-edit if tooling unavailable) to flip phase 1 and 2 to `completed` in `plan.md`." Cite the tool, fall back to manual.

---

## Summary of Recommended Cuts

| Original | Trimmed |
|---|---|
| 14 tests total | 9 tests total (2 in Phase 1 + 7 in Phase 2) |
| Separate `cmd/traces_render.go` file, 80-120 LOC | Inline + one small `buildSpanTree` helper in `cmd/traces.go`, ~40 LOC |
| `unmarshalMapStrict` helper OR inline (deferred) | Inline only — decision made now |
| Events: "first 5 + last 1 with separator" | Print all events, one per line |
| Phase 2 step 4: "verify and possibly extend `MapServerCode`" | "Add client-side id validation; trust existing mapping" |
| Phase 1 effort 2-3h | Phase 1 effort 1-1.5h |
| Plan.md hand-edit for status | `/ck:plan-check --sync` |
| `codebase-summary.md` update (conditional) | CUT |

**Net expected savings:** ~1.5h Phase 1 + ~1h Phase 2 + cognitive load of one deferred decision. Total ~25-30% reduction. Bug still gets fixed, contract still honored, TDD still enforced.

---

## Open Questions for User

None affect verified user decisions. TDD discipline is preserved (red tests still precede impl in both phases). Exit-code contract is preserved (the trim doesn't reduce categories, just shifts where they're enforced). Span-tree rendering is preserved because it appears to honor the issue text "root/child events" AND the existing command's stated `Short: "Get trace with span tree"` promise.

If the user disagrees on **event truncation** specifically (Finding "Events truncation"), surface that — it's the only judgment call where reasonable people might differ on UX-vs-scope.

---

**Status:** DONE
**Severity counts:** Critical: 2 / Important: 5 / Minor: 3
