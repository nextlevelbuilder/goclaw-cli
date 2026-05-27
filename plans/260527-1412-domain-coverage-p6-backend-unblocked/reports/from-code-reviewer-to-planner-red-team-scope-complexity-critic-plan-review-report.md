# Red-Team Plan Review — Scope & Complexity Critic

Plan: `plans/260527-1412-domain-coverage-p6-backend-unblocked/`
Lens: Hostile YAGNI enforcer. Findings backed by codebase grep evidence.

---

## Finding 1: `activityCmd` already exists — top-level namespace collision

- **Severity:** Critical
- **Location:** Phase 5, section "Files" + `plan.md` "Existing CLI State" table
- **Flaw:** Plan declares "no existing `cmd/admin_activity.go`" and "create new `activityCmd` parent". Both wrong — `activityCmd` already exists in `cmd/admin.go:133` and is already registered at root in `cmd/admin.go:172` as `goclaw activity` (audit-log viewer). Defining a second `activityCmd` is a Go duplicate-declaration compile error.
- **Failure scenario:** Phase 5 step 2 ("Implement `activityCmd` + `activityAggregateCmd`") produces `cmd redeclared` at `go build`. Plan's claim "no command-name collisions" (plan.md:99) is false at the symbol level even if not at the user-facing `Use:` level (both would still register `Use: "activity"` at root — Cobra would also reject).
- **Evidence:**
  - `cmd/admin.go:133`: `var activityCmd = &cobra.Command{ Use: "activity", Short: "View audit log", ...}`
  - `cmd/admin.go:172`: `rootCmd.AddCommand(approvalsCmd, delegationsCmd, adminCredentialsCmd, activityCmd, ttsCmd, mediaCmd, voicesCmd)`
  - Plan: `plan.md:96` claims `_(no cmd/admin_activity.go)_  | — | new file cmd/activity_aggregate.go`
  - Plan: `phase-05-activity-and-logs-aggregate.md:58`: "no existing activity command group exists"
- **Suggested fix:** Add `aggregate` as a subcommand of the existing `activityCmd` (1 line: `activityCmd.AddCommand(activityAggregateCmd)`). Drops the "namespace collision risk" entry (`phase-05:110`) entirely. Also removes the entire "Decide: top-level vs admin group" decision that phase 5 stages — the answer is already in the tree.

---

## Finding 2: `providers verify` referenced as user fallback — that command does not exist

- **Severity:** High
- **Location:** Phase 2, section "2.2 `goclaw providers reconnect`"
- **Flaw:** Plan justifies omitting `--verify` flag by telling users: "users call `goclaw providers verify <id>` separately if needed". That command does not exist. The only verify-shaped command is `providers verify-embedding <id>`, which hits `/v1/providers/{id}/verify-embedding` — a different backend endpoint, not a generic provider verify.
- **Failure scenario:** PR review / docs writer reads phase 2 rationale, adds README copy saying "use `providers verify`", users get "unknown command" at runtime. Worse, if the rationale survives into commit body, it becomes permanently misleading.
- **Evidence:**
  - `cmd/providers_verify.go:11`: `Use: "verify-embedding <id>"`
  - `grep "Use:.*\"verify\"" cmd/*.go` → 0 hits.
  - Plan: `phase-02:46`: "users call `goclaw providers verify <id>` separately if needed"
- **Suggested fix:** Strike the sentence or rewrite as "Backend handles verify internally on reconnect; no separate CLI verify exists today." The negative test (assert no `verify` key in body) is correct and should stay.

---

## Finding 3: Phase 1 (Scope Lock, 1h) re-does work already done in plan creation

- **Severity:** High
- **Location:** Phase 1 entirety
- **Flaw:** Phase 1's six steps are: re-read backend handlers, re-confirm beta tag, re-run collision grep, write a contracts table. Every one of these was already executed during plan authoring — see `plan.md:33` ("**Verified via `gh api compare`:** `v3.12.0-beta.20` is the earliest tag identical to `43049d3b`") and `plan.md:88-99` (the existing CLI state table is the collision sweep output). Per `~/.claude/rules/review-audit-self-decision.md` #1, "Verified Decisions Are Sticky — Audit Does Not Auto-Reverse" — re-verifying without a triggering reason is wasted time.
- **Failure scenario:** 1h of phase-1 effort produces a redundant `reports/scope-lock-260527-p6.md` that duplicates `plan.md`'s decisions/inventory tables. Phases 2–5 already cite contracts inline; no downstream phase actually consumes phase 1's report.
- **Evidence:**
  - `phase-01:21`: "Produce `reports/scope-lock-260527-p6.md` summarizing contracts" — but contracts are already inline in phases 2–5 (e.g., `phase-02:30-34` traces follow envelope, `phase-03:30-36` branch body/response).
  - `phase-01:19`: "(already verified during plan creation; phase re-asserts)" — admits this is re-verification.
- **Suggested fix:** Delete phase 1. Move the "production-time contract sanity check" (10 minutes max) into phase 2's TDD prep. Renumber phases 2–7 → 1–6.

---

## Finding 4: Phase 6 (Tests & Docs) and Phase 7 (Ship) red-team checks duplicate each other

- **Severity:** Medium
- **Location:** Phase 6 step 3 "Red-team diff sweep" + Phase 7 step 7 "Run review and fix findings"
- **Flaw:** Phase 6 lists 8 explicit red-team confirmations (no replay command, no generic logs aggregate, no WS delta consumer, etc.) and produces `reports/red-team-260527-p6.md`. Phase 7 then runs `claude-review` workflow (PR review) which performs the same checks. Two artifacts, two passes, same questions.
- **Failure scenario:** Reviewer time wasted re-asserting absences. If phase 6's report disagrees with phase 7's PR review (e.g., new finding surfaces), reconciliation overhead is non-zero.
- **Evidence:**
  - `phase-06:27-35`: enumerated 8-item red-team sweep.
  - `phase-07:38`: "Run review and fix findings before merge."
  - `phase-07:61`: "`claude-review` workflow may flag advisory issues."
- **Suggested fix:** Drop phase 6's red-team sweep as a separate artifact. Fold the 8-item checklist into the PR body template (phase 7 step 6 already lists "Explicit out-of-scope list (verbatim from issue #16)"). Saves ~30 min and one stale report.

---

## Finding 5: `cmd/sessions.go` is 171 lines — split contingency will trigger but is conditional

- **Severity:** Medium
- **Location:** Phase 3, section "Files" + Todo list item "If `cmd/sessions.go` exceeds 200 lines, split"
- **Flaw:** Plan defers the modularization decision to runtime: "If file grows past 200 lines, extract." Per `~/.claude/rules/development-rules.md`, file size limit is 200 LOC. Current file is 171 lines. Adding two new commands with positional args, ~5 flags total, RunE bodies with body construction + path escaping + printer dispatch will easily add 60–100 lines — guaranteed to trip the limit. Speculative-conditional planning where the answer is deterministic.
- **Failure scenario:** Implementer writes both commands inline, ends up at ~250 lines, then has to split mid-phase (or punts the split and breaks the repo rule). Either way it's churn vs. deciding upfront.
- **Evidence:**
  - `wc -l cmd/sessions.go` → 171.
  - `phase-03:57`: "If `cmd/sessions.go` grows past 200 lines, split into `cmd/sessions_branch.go` and `cmd/sessions_follow.go`".
  - Same pattern repeats in `phase-04:72` (channels_writers.go currently 89 lines — adding `test` likely safe) and `phase-05:59` (logs.go currently 111 lines — adding `aggregate` may also overflow).
- **Suggested fix:** Pre-commit to `cmd/sessions_branch.go` + `cmd/sessions_follow.go` from the start (matches existing pattern: `chat_sessions.go`, `providers_verify.go`, `providers_reconnect.go` already preferred in phase 2). For phase 4 (writers test), inline is fine. For phase 5 (logs aggregate), pre-commit to `cmd/logs_aggregate.go`.

---

## Finding 6: Phase 6 CHANGELOG step contradicts itself

- **Severity:** Medium
- **Location:** Phase 6 step 2 and Todo list
- **Flaw:** Step 2 says: "`CHANGELOG.md` — add a `## Unreleased` entry (or follow existing semantic-release commit convention; do not manually edit if release notes are commit-driven)." Todo says: "CHANGELOG entry (or confirmed commit-driven)." The repo IS commit-driven (`go-semantic-release` runs from `.github/workflows/release.yaml` on push to `dev`/`main`). So the answer is already "do not manually edit." Leaving it as a conditional todo invites accidental hand-editing.
- **Failure scenario:** Implementer hand-edits `CHANGELOG.md`, conflicts with the next automated release commit, or worse, the duplicate entry sticks and pollutes the file.
- **Evidence:**
  - `.github/workflows/release.yaml:51-53`: `go install github.com/go-semantic-release/semantic-release/v2/cmd/semantic-release@v2.31.0` invoked on every dev/main push.
  - `CHANGELOG.md:8`: existing `## [Unreleased] — Domain Coverage Expansion (P0–P5)` is hand-curated — suggests prior plans also drifted on this question.
  - `phase-06:26`: the self-contradicting line.
- **Suggested fix:** Replace the line with a single instruction: "Skip CHANGELOG.md — semantic-release handles it from the `feat:` commit message." Remove the todo item.

---

## Finding 7: Phase 7 ship checklist contains items that belong to `/ck:ship`

- **Severity:** Medium
- **Location:** Phase 7 entire "Implementation Steps" + Todo list
- **Flaw:** Phase 7 enumerates 9 procedural items (git status clean, secret scan, push, gh pr create, PR body template, review-and-fix, watch CI, watch beta release publish). Per project workflow, `/ck:ship` is the dedicated skill for this. The plan duplicates that runbook inline. Worse, "Watch CI + Release until beta release publishes" (step 8) is not gated by anything in the plan — it's open-ended async waiting that should not live in a plan phase.
- **Failure scenario:** Phase status stays "in_progress" for hours/days while implementer waits for beta release publish. Plan progress reports become misleading.
- **Evidence:**
  - `phase-07:25-39`: 9 numbered steps, only steps 3–4 (commit composition + push) are plan-specific.
  - `~/.claude/rules/skill-workflow-routing.md`: `/ck:ship — run full shipping pipeline (tests, review, version, PR)`.
- **Suggested fix:** Collapse phase 7 to: "Invoke `/ck:ship` once phases 1–5 are green. Provide it the PR body template (the 4-line backend evidence block)." Drop the post-merge "watch beta release" step — it is not part of plan completion.

---

## Finding 8: Test file name `channels_writers_test_test.go` is a smell — and the contingency note proves doubt

- **Severity:** Medium
- **Location:** Phase 4, section "Files"
- **Flaw:** Plan proposes `cmd/channels_writers_test_test.go` for the writers-`test` subcommand test file, then adds "Alternative if Go tooling balks: `cmd/channels_writers_probe_test.go` — but verify Go accepts `_test_test.go` first (it does)." If the planner felt the need to add a contingency, the name is hostile to readers. No existing file in `cmd/` uses `_test_test` (verified `ls cmd/ | grep test_test` → empty). The kebab-case readability principle from development-rules.md ("self-documenting for LLM tools") favors clarity over cleverness.
- **Failure scenario:** Future grep for "writers_test" matches both source (when added) and test file ambiguously; LLM scanners get confused on filename intent; reviewers waste a beat parsing the double-suffix.
- **Evidence:**
  - `ls cmd/ | grep test_test` → no matches.
  - Existing convention: `cmd/agents_lifecycle_test.go`, `cmd/p5_fillers_test.go`, `cmd/heartbeat_test.go` — single `_test` suffix only.
  - `phase-04:39-41`: the contingency comment.
- **Suggested fix:** Use `cmd/channels_writers_probe_test.go`. Drop the contingency note. One-line decision, no future ambiguity.

---

## Finding 9: 4 phases vs 7 surfaces — claimed grouping ("shared client helpers + test fixtures") is unsubstantiated

- **Severity:** Medium
- **Location:** `plan.md` line 38 "Grouping: 4 implementation phases, not 7 (shared client helpers + test fixtures)"
- **Flaw:** Plan justifies 4-phase grouping by claiming shared helpers/fixtures, but no phase actually declares shared code. Phase 2 (traces follow + providers reconnect) — totally different request shapes (GET with query vs POST empty body). Phase 5 (activity aggregate + logs aggregate) — share `KEY/COUNT/LAST_SEEN` table shape, but `last_seen` is a string in one and epoch millis int in the other (phase 5 itself flags this), so the shared output helper has to branch. The grouping is visual tidiness, not engineering reuse.
- **Failure scenario:** Future maintainer reads phase 2 expecting a shared helper, finds two independent commands; or worse, an implementer over-engineers a shared abstraction to satisfy the rationale and ships a flag-driven dispatcher for two unrelated endpoints.
- **Evidence:**
  - `phase-02:55-60` Files section: independent files per surface, no shared helper named.
  - `phase-05:109` Risks: "`last_seen` type mismatch between activity (RFC3339 string) and logs runtime (epoch millis int)" — proves the surfaces don't share output logic cleanly.
  - `plan.md:38`: the unsubstantiated claim.
- **Suggested fix:** Drop the grouping rationale. Either keep 4 phases on grounds of "review batching" (honest reason: similar PR-side scope) or split into 7 micro-phases. Don't lie about shared code that isn't shared.

---

## Finding 10: Validation Gates block in plan.md duplicates per-phase validation steps

- **Severity:** Medium
- **Location:** `plan.md:114-120` "Validation Gates (per phase + final)"
- **Flaw:** plan.md enumerates the validation commands (`go test`, `go vet`, `go build`, live smoke, red-team diff), and every phase 2–6 then repeats `go vet ./... && go build ./...` in its TDD Sequence and todo list. Phase 6 then runs them again as the "final gate." Three layers of the same commands. Per development-rules.md ("Keep individual code files under 200 lines for optimal context management"), plans should reduce redundancy, not multiply it.
- **Failure scenario:** Plan reader stops trusting the validation lists ("they're boilerplate, skip"). When phase 6's gate adds `-count=1` and `make build` that earlier phases don't, the divergence is invisible.
- **Evidence:**
  - `plan.md:116-120`: top-level validation list.
  - `phase-02:67`, `phase-03:67`, `phase-04:46`, `phase-05:69`: each repeats "go vet + go build clean."
  - `phase-06:18-22`: full final gate including `make build`.
- **Suggested fix:** Keep validation list in plan.md only. Per-phase todos reduce to "phase tests pass." Phase 6 is the only phase that explicitly enumerates `go test -count=1` + `make build`.

---

## Summary

10 findings, 1 critical, 2 high, 7 medium. The critical defect (Finding 1: duplicate `activityCmd`) will cause a compile failure in phase 5. The high-severity defects either misinform users (Finding 2) or waste a planning phase (Finding 3). The medium findings cluster around YAGNI violations: speculative conditionals (`if file > 200 lines`), duplicated red-team passes, ship-procedure boilerplate that belongs in `/ck:ship`, and unsubstantiated grouping rationale.

## Unresolved Questions

- Should `goclaw activity aggregate` nest under existing `activityCmd` (audit-log viewer) even though semantics differ (point-in-time aggregation vs list-all)? Decision needed before phase 5 starts.
- Does the existing `activityCmd` use `Use: "activity"` cleanly enough to share, or does adding `aggregate` warrant a refactor to `activity list` + `activity aggregate` (breaking change for current `goclaw activity` users)?
- Is the live-smoke step (`v3.12.0-beta.20`+) blocking or optional? Plan says "optional, after merge" — but no follow-up plan handles a smoke-test failure.
