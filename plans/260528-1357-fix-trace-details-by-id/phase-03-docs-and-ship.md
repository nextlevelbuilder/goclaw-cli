---
phase: 3
title: "Docs and ship"
status: pending
priority: P2
effort: "45m-1h"
dependencies: [2]
---

# Phase 3: Docs and ship

## Overview

Update CHANGELOG, file upstream issue ONLY if Phase 1 classified server-side root cause, then ship via `/ck:ship beta` (target: `dev`). Closes issue #17.

## Requirements

**Functional**
- CHANGELOG entry under `[Unreleased]` documenting the fix.
- README `traces get` example updated IF output behavior materially changed (likely yes — TTY mode is new).
- `docs/codebase-summary.md` reflects new id-validation + render path IF non-trivial (likely a one-line bullet).
- Upstream issue filed at `digitopvn/goclaw` IFF Phase 1 classification was server-side or hybrid. PR body links it.
- `gh issue close 17` runs after the dev→main promotion merges (or auto-closes via `Closes #17` in the promotion PR).

**Non-functional**
- Commit message follows conventional commits (`fix(cli): ...`).
- PR body references `Closes #17` (will auto-close only when promoted to default branch), links Phase-1 repro report, lists acceptance-criteria mapping.

## Architecture

No code changes — docs + ship pipeline only.

## Related Code Files

- Modify: `CHANGELOG.md` — add entry under existing `[Unreleased]` block (append to P0–P6 section if present).
- Modify: `README.md` — refresh `traces get` example block.
- Modify: `docs/codebase-summary.md` — one bullet on the new traces-get behavior.
- Modify: `plans/260528-1357-fix-trace-details-by-id/plan.md` — flip phase statuses via `ck plan check` (not hand-edit).
- No source code edits.

## Implementation Steps

1. **Sync plan status via CLI** — `cd plans/260528-1357-fix-trace-details-by-id && ck plan check 1 && ck plan check 2 && ck plan check 3 --start`. (Phase 3 marked in-progress while shipping; flip to completed at the end via `ck plan check 3`.)

2. **CHANGELOG** — add under `[Unreleased]`:
   ```
   ### Fixed
   - `goclaw traces get <id>` — human-readable rendering for TTY mode (header + span tree + events), checked unmarshal (no more silent empty `{}`), strict id validation (rejects path-traversal), and distinct exit codes for not-found (3) / permission-denied (2) / malformed-id (4) / server-failure (5). Closes issue #17.
   ```
   If Phase 1 found a server-side root cause, add a `### Notes` line linking the upstream issue URL.

3. **README** — locate `traces get` example. If absent, add to traces section showing expected human + JSON output samples (use scrubbed Phase-1 fixture as the JSON sample).

4. **codebase-summary.md** — add bullet under traces/output section noting validated id + structured render.

5. **Upstream issue (conditional)** — only if root cause was server-side or hybrid:
   - Check rights first: `gh api repos/digitopvn/goclaw -q .permissions`. If no issue-create permission, instead post a comment on issue #17 with the drafted body for the maintainer to file.
   - If rights present: `gh issue create -R digitopvn/goclaw --title "..." --body "..."` using the scrubbed body drafted in Phase 1's repro report.
   - Add the issue URL to CHANGELOG `### Notes`.

6. **Code review** — spawn `code-reviewer` subagent on the diff before ship. Reviewer must explicitly verify:
   - No bearer tokens or PII in committed fixture/report (`grep -i 'eyJ\|Bearer\|sk-' cmd/testdata/` returns 0).
   - All 10 trace-get tests pass.
   - `cmd/traces.go` net add ≤ 100 LOC.
   - No plan-artifact references ("Phase 2", "F1", etc.) in production/test code.

7. **Ship** — `/ck:ship beta` (target: `dev` branch).

8. **Verify PR** — confirm PR body includes `Closes #17`, links Phase-1 repro report, lists acceptance-criteria mapping (1→fix; 2→table+JSON tests; 3→exit-code tests + path-traversal block; 4→upstream link if applicable; 5→`cmd/traces_get_test.go`).

9. **Issue close after promotion** — PR targets `dev`; auto-close fires only when promoted to default branch. Document this in the issue with a comment: "Fixed in dev via PR #X; will auto-close on next `dev → main` promotion."

## Success Criteria

- [ ] CHANGELOG `[Unreleased]` entry present.
- [ ] README `traces get` example reflects new behavior.
- [ ] `docs/codebase-summary.md` updated.
- [ ] Upstream `digitopvn/goclaw` issue filed/commented-for-filing (conditional on Phase 1 verdict).
- [ ] code-reviewer subagent verdict: DONE, 0 Critical, 0 Important.
- [ ] PR created against `dev`, all CI checks green.
- [ ] Plan phases 1, 2, 3 all marked completed via `ck plan check`.
- [ ] Comment posted on issue #17 explaining the dev → main promotion auto-close behavior.

## Risk Assessment

- **Risk:** PR auto-close via `Closes #17` only fires on default-branch merge; PR targets `dev` first. **Mitigation:** explicit comment on #17 explaining the timing; manual close after `dev → main` promotion if needed.
- **Risk:** ship pipeline blocks on unrelated CI flakiness. **Mitigation:** investigate per failure; do not bypass with `--skip-tests`.
- **Risk:** no rights on `digitopvn/goclaw`. **Mitigation:** fallback to issue-#17 comment with the drafted body (step 5 above).

## Security Considerations

- Fixture/report secret-scrub gate re-checked by code-reviewer subagent (step 6).
- Upstream issue body reuses scrubbed content only.

## Next Steps

After merge to `dev`, the natural follow-up is `/ck:ship official` to promote `dev` → `main`. Out of scope for this plan.
