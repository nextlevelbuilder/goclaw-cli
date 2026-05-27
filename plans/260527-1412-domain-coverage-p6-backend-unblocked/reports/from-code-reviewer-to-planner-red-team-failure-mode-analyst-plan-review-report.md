# Red-Team Plan Review — Failure Mode Analyst

Plan: `plans/260527-1412-domain-coverage-p6-backend-unblocked/`
Lens: Failure Mode Analyst — Murphy's Law
Reviewer: code-reviewer (hostile)
Date: 2026-05-27

## Finding 1: `buildBody` silently drops `--up-to-index 0` — branching plan is broken at the zero boundary

- **Severity:** Critical
- **Location:** Phase 3, section "3.1 `goclaw sessions branch`" — "Required: ... `--up-to-index` (int, >= 0)"
- **Flaw:** The plan declares `--up-to-index` valid at zero ("`>= 0`") and emits the body via the existing `buildBody` helper (Plan §"Key Patterns" / phase 3 body shape). `buildBody` (cmd/helpers.go:86–89) deliberately drops `case int: if v != 0 { body[key] = v }`. So `--up-to-index 0` (branch at the very first message — a real and required use case) silently disappears from the request body. The server either rejects (`INVALID_REQUEST`) or worse, defaults `up_to_index` to something non-zero and the user gets a branch they didn't ask for.
- **Failure scenario:** User runs `goclaw sessions branch sess-1 --up-to-index 0` expecting an empty new session keyed off message #0. The CLI POSTs `{}` (no `up_to_index`). Server either 400s with a confusing error or branches at default index. Data integrity violation; either way reproducible.
- **Evidence:**
  - `cmd/helpers.go:86–89`: `case int: if v != 0 { body[key] = v }`
  - `plans/.../phase-03-sessions-branch-and-follow.md:25–31`: "Required: ... `--up-to-index` (int, >= 0)" and body example shows `"up_to_index":12`.
  - Same hazard for `--cursor 0` in `sessions follow` if it ever migrates to body; today it's a query string but plan §3.2 line 84 says "Default cursor=0, limit=50 appear in query string" — assertion requires special handling to emit zero values.
- **Suggested fix:** Add a phase-3 implementation note: do NOT use `buildBody` for `up_to_index`; build the body map directly so `0` is preserved. Add an explicit unit test `--up-to-index 0` ⇒ body contains `"up_to_index":0`. Similarly test `--cursor 0` appears in query, not omitted.

## Finding 2: `logs aggregate` epoch-millis `last_seen` will render as `1.76e+12` in tables

- **Severity:** High
- **Location:** Phase 5, section "5.2 `goclaw logs aggregate`" — "Table tolerates numeric `last_seen` (epoch millis) without panicking"
- **Flaw:** Plan acknowledges the type mismatch (RFC3339 string vs epoch millis int) but only tests "without panicking". The `str()` helper at `cmd/helpers.go:56–61` formats with `fmt.Sprintf("%v", v)`. Because `encoding/json` decodes JSON numbers into `float64` by default, `1760000000000` round-trips to `%v` as `1.76e+12`. No panic — but the LAST_SEEN column is unreadable garbage in human mode. The plan's own success criterion ("`last_seen` numeric type handled without runtime panic") is satisfied while the actual UX is broken. Murphy ships.
- **Failure scenario:** Operator runs `goclaw logs aggregate --group-by level` after an incident, scans the LAST_SEEN column, sees `1.76e+12` for every row, can't tell which bucket is fresh. Reaches for `--output json` to recover.
- **Evidence:**
  - `cmd/helpers.go:56`: `func str(m map[string]any, key string) string { ... return fmt.Sprintf("%v", v) }`
  - `cmd/helpers.go:49–52`: `unmarshalMap` uses `json.Unmarshal` without `UseNumber()`, so all JSON numbers ⇒ `float64`.
  - `plans/.../phase-05-activity-and-logs-aggregate.md:49–52` and 89: notes "epoch millis number" but test only asserts no panic.
- **Suggested fix:** Add a dedicated table-rendering helper that detects numeric `last_seen` and (a) formats via `time.UnixMilli(int64(v)).Format(time.RFC3339)` or (b) at minimum uses `%d` after type-switching. Update the phase-5 test to assert the rendered cell contains a parseable RFC3339 string OR a non-scientific integer — NOT just "no panic".

## Finding 3: Plan claims `cmd/providers_crud.go` exists; it does not

- **Severity:** High
- **Location:** Plan overview, "Existing CLI State" table (line 93)
- **Flaw:** Row reads `cmd/providers.go + providers_crud.go + providers_verify.go | CRUD + verify`. `providers_crud.go` does not exist in this worktree. The code comment at `cmd/providers.go:69` even says "create/update/delete/verify/status registered from providers_crud.go" — that's a stale comment but the plan replicates the stale fact. There are NO `providersCreateCmd / UpdateCmd / DeleteCmd / StatusCmd` registered anywhere. Adding `reconnect` implies a working CRUD surface to integrate with; the plan never noticed CRUD isn't implemented and the existing comment is wrong. Result: phase 2 dev will discover halfway through that the "register on `providersCmd`" surface is barely populated.
- **Failure scenario:** Phase 2 dev opens `cmd/providers.go`, sees only `list/get/models`, no `update/delete`, can't find `providers_crud.go`, loses 30 minutes debugging plan vs reality, has to escalate or improvise scope.
- **Evidence:**
  - `find . -name "providers_*.go"` returns only `providers_claude_cli.go`, `providers_codex_pool.go`, `providers_verify.go`. No `providers_crud.go`.
  - `cmd/providers.go:68–73`: comment claims registration "from providers_crud.go" but `init()` only adds `providersListCmd, providersGetCmd, providersModelsCmd`.
  - Plan `plan.md:93` perpetuates the falsehood.
- **Suggested fix:** Phase 1 (Scope Lock) must grep `cmd/providers_*.go` and update both the plan table AND the stale comment in `cmd/providers.go:69` (file or commit message). Either delete the lying comment or implement the missing CRUD before phase 2.

## Finding 4: "One polling request only" assertion has no test that actually counts requests in the right way

- **Severity:** High
- **Location:** Phase 3 test list, "Only one HTTP request is issued (no implicit loop)"; Phase 2 test list, "**One request only. No watch loop.**"
- **Flaw:** The phase-3 file lists "Only one HTTP request is issued" as a test for `sessions follow`, but the phase-2 (`traces follow`) test list contains NO equivalent count-the-requests test. The plan's strongest guarantee — no watch loop — is unverified for two of three follow surfaces (`traces follow` + `providers reconnect` could in theory accidentally retry on transient error). Worse: even the listed test for `sessions follow` doesn't specify HOW: does the test wait for a timeout to confirm no second call, or just count `r.URL` hits after `RunE` returns? If the implementation uses `client.FollowStream` (which retries with exponential backoff per `internal/client/follow.go`) by mistake, a simple "expect one call" assertion against an `httptest` server may still pass on the first iteration before reconnect.
- **Failure scenario:** Implementation accidentally wires `traces follow` through `client.FollowStream` (the WS streamer with reconnect). On a happy server it makes one call and returns; on an EOF mid-response or 502, it retries forever. Single-call test passes locally; prod loops on the bad path.
- **Evidence:**
  - `cmd/logs.go:43–47`: `logs tail` calls `client.FollowStream(...)` — pattern is contagious and a developer cooking phase 2 may copy-paste.
  - `internal/client/follow.go` and `internal/client/follow_test.go` confirm built-in reconnect with `MaxRetries`.
  - Phase 2 test list (lines 73–80 of `phase-02-traces-follow-and-providers-reconnect.md`) has no request-count assertion.
- **Suggested fix:** Mandate, for every "follow" command added: (a) a test that wraps `httptest.NewServer` with an atomic counter and asserts `count == 1` after `RunE` returns, AND (b) a test where the server returns 502 once — assert the CLI fails fast (no retry), not that it succeeds on retry. Add to phase 2 AND phase 3.

## Finding 5: New top-level `activity` Cobra command breaks `TestAllCommandsRegistered` invariant philosophy and risks namespace squat

- **Severity:** Medium
- **Location:** Phase 5, section "Files" — "Decide: place under a new top-level `activity` Cobra group ... create new `activityCmd` parent"
- **Flaw:** The decision is wrapped in soft language ("Codex prompt expects `goclaw activity aggregate` as top-level") but the plan ships the decision without verifying impact on `cmd/cmd_test.go` (which maintains an `expected` list of root commands). Adding `activity` without updating `expected` won't fail the existing test (it's a "must-contain" check), but it WILL fail any future test that asserts the inverse. More concretely: the codex prompt is a single source; the actual server `/v1/activity/aggregate` lives alongside `/v1/audit/*` and `/v1/admin/*`. A top-level `activity` competes with potential future `goclaw admin activity ...` namespacing. The plan acknowledges this risk in §"Risks" (line 110: "New top-level `activity` command might collide with future scope. Document the namespace decision in PR body") — but documenting in a PR body is not a mitigation; it's a deferral.
- **Failure scenario:** Three months later, backend ships `/v1/admin/activity/*`. CLI now has `goclaw activity aggregate` at top level and would need `goclaw admin activity ...` siblings. Confused users, breaking renames, deprecation churn.
- **Evidence:**
  - `cmd/cmd_test.go:18–57`: hard-coded list of root commands. `activity` not present; plan never asks to add it.
  - `plan.md:96`: "_(no `cmd/admin_activity.go`)_ — new file `cmd/activity_aggregate.go`" — but `cmd/admin.go`, `cmd/admin_credentials.go`, `cmd/admin_tts_media.go` show an established `admin_*` family. Why is activity not `admin_activity_aggregate.go` under `admin activity aggregate`?
- **Suggested fix:** Phase 1 should escalate this namespace decision to the user (or document a verified backend roadmap claim that `/v1/activity/*` is the permanent home, NOT `/v1/admin/activity/*`). If keeping `activity` at top level, phase 5 must also patch the `expected` list in `cmd/cmd_test.go` AND `TestCommandUseFields`.

## Finding 6: Test setup mutates package-level globals (`cfg`, `printer`); plan adds 7 new test files with no parallel-safety mitigation

- **Severity:** Medium
- **Location:** Phase 2/3/4/5 test files (all 7+ new `_test.go` files)
- **Flaw:** The existing `setupP5HTTPTest` (cmd/p5_fillers_test.go:17–20) sets `cfg = &config.Config{...}` and `printer = output.NewPrinter("json")` — these are package-level globals. Tests cannot be `t.Parallel()` safe. The release workflow runs `go test -race -count=1 ./...` (`.github/workflows/release.yaml:30`). If any new test ever adds `t.Parallel()` (a future contributor or a copy-pasted block), the race detector flags `cfg`/`printer` mutations and the release breaks. The plan adds 7 tests sharing this pattern and never mandates a per-test `t.Cleanup(restoreCfg)` to restore prior state — meaning test ordering can leak `cfg` from `traces_follow_test` into `sessions_follow_test` if the latter runs before its own setup.
- **Failure scenario:** Test A sets `cfg.OutputFormat = "json"`. Test B added later forgets to call setup and asserts table output. Race detector run in CI passes (no parallel). Six months later someone adds `t.Parallel()` for speed — boom, intermittent CI fail.
- **Evidence:**
  - `cmd/p5_fillers_test.go:17–20`: globals mutated, no cleanup.
  - `.github/workflows/release.yaml:30`: `go test -race -count=1 ./...`.
- **Suggested fix:** Phase 6 (Tests and Docs Sweep) must add a section "Add `t.Cleanup(func(){ cfg, printer = origCfg, origPrinter })` to every new test that mutates package globals" and add a phase-6 lint pass that greps `t.Parallel()` against `cmd/*_test.go` and fails the gate.

## Finding 7: `--metadata k=v` parser contract is underspecified — duplicate keys, empty values, Unicode, `=` in value

- **Severity:** Medium
- **Location:** Phase 3, section "3.1 `goclaw sessions branch`" — "`--metadata` parses repeated `key=value`; reject malformed entries before HTTP call"
- **Flaw:** The plan defines two cases: well-formed (`foo=bar`) and missing `=` (`foobar`, rejected). It is silent on:
  1. Duplicate keys: `--metadata foo=bar --metadata foo=baz` — second wins, error, or array?
  2. Empty value: `--metadata foo=` — keep as `""`, drop, or reject?
  3. Empty key: `--metadata =bar` — reject before HTTP? Plan doesn't say.
  4. Multiple `=`: `--metadata token=abc=def` — `SplitN(s, "=", 2)` or `Split(s, "=")` (would silently truncate values containing `=` like base64 padding)?
  5. Unicode/whitespace: `--metadata "key with space=val"` — preserved or rejected?
  6. Nested object: `--metadata foo.bar=baz` — flat or nested? `agents instances update-metadata` (cmd/agents_instances.go:101) uses `--metadata='{"tier":"premium"}'` JSON-string style. This new flag invents a divergent contract.
- **Failure scenario:** User stores a base64-encoded session metadata value (`token=AbCd123==`). CLI truncates at first `=`, server gets `{"token":"AbCd123"}` instead of `{"token":"AbCd123=="}`. Silent data corruption. Branch session metadata wrong forever.
- **Evidence:**
  - `cmd/agents_instances.go:98–101`: existing convention is JSON-string `--metadata='{"k":"v"}'`. Plan diverges to `k=v` repeated.
  - Phase 3 lines 26, 76 of `phase-03-sessions-branch-and-follow.md`: spec only addresses the no-`=` case.
- **Suggested fix:** Phase 1 must lock the parsing contract: use `strings.SplitN(s, "=", 2)`; reject empty key; allow empty value; "duplicate keys: last wins"; Unicode allowed verbatim. Add unit tests for all 6 cases. Reconsider whether to match the existing `agents instances update-metadata` JSON convention instead of inventing a new one.

## Finding 8: `cmd/sessions.go` is 171 lines today; phase 3 adds two commands but trigger to split is "if it grows past 200" — split decision happens mid-implementation

- **Severity:** Medium
- **Location:** Phase 3, section "Files" — "If `cmd/sessions.go` grows past 200 lines, split into `cmd/sessions_branch.go` and `cmd/sessions_follow.go`"
- **Flaw:** Current `sessions.go` is 171 LOC; two new subcommands with `RunE`, flags, validation, table renderer will EASILY add 60–120 LOC each = file definitely exceeds 200. The plan defers the split decision until "after" but TDD writes the test file first, sees red, then writes the command. By the time you notice the file is 230 LOC, you've already committed the wrong structure. Same risk for `cmd/channels_writers.go` (89 LOC, phase 4 line 72 acknowledges) and `cmd/logs.go` (111 LOC).
- **Failure scenario:** Developer cooks phase 3, ends up with `cmd/sessions.go` at 280 lines. Pushes. Reviewer flags the modularization rule. Developer extracts to `sessions_branch.go` AFTER tests landed against `cmd/sessions.go`. Renames break test imports if not careful (they shouldn't because Go tests are by package, not file, but the diff is noisy).
- **Evidence:**
  - `wc -l cmd/sessions.go` ⇒ 171; `cmd/channels_writers.go` ⇒ 89; `cmd/logs.go` ⇒ 111.
  - Phase 3 line 57: "If `cmd/sessions.go` grows past 200 lines, split"
- **Suggested fix:** Make the split mandatory and upfront. Phase 3 should say "Create `cmd/sessions_branch.go` and `cmd/sessions_follow.go` from the start." Same for `cmd/channels_writers_test_cmd.go` (rename the file to avoid the `_test_test.go` confusion entirely) and `cmd/logs_aggregate.go`. Removes a mid-phase decision.

## Finding 9: Phase 7 relies on `feat:` to trigger a beta version bump — but no validation step or rollback story if it doesn't

- **Severity:** Medium
- **Location:** Phase 7, "Risks" — "semantic-release may need a `feat:` commit to trigger a beta bump; verify the conventional message header"
- **Flaw:** "Verify the conventional message header" is not an action; it's a wish. The release workflow (`.github/workflows/release.yaml:50–58`) installs `go-semantic-release` and runs it. If the commit subject is `feat(cli): add P6 backend-unblocked CLI surfaces` (exactly as plan §7 step 3 prescribes), conventional commits should produce a minor bump prerelease. But the workflow's "Configure beta release stream" step (line 35–47) calculates next minor against the **latest stable tag**, not the latest beta. Plan doesn't verify that the merge will actually produce a new beta tag distinguishable from `v3.12.0-beta.35`. Worse, if the build/test/race step fails (line 31: `go test -race -count=1 ./...`), the release won't tag. Plan §6 only runs `go test -count=1 ./...` (no `-race`), so the worktree can pass phase 6 yet fail CI release.
- **Failure scenario:** Phase 6 green. Merged to `dev`. Release workflow runs `-race`, hits an unflagged data race in one of the 7 new commands (e.g., shared `cfg` mutation across tests OR a goroutine in a copy-pasted FollowStream pattern). Release job fails, no tag emitted, beta consumers never see the surfaces. Rollback story: nothing — the merge stays, the tag doesn't.
- **Evidence:**
  - `.github/workflows/release.yaml:30`: `go test -race -count=1 ./...` runs in release; plan §6 only `go test -count=1`.
  - Phase 7 line 62: "semantic-release may need a `feat:` commit ... verify the conventional message header" — no concrete check.
- **Suggested fix:** Add to phase 6 validation gate: `go test -race -count=1 ./...` (matches CI exactly). Add to phase 7 a pre-merge check: dry-run semantic-release locally (`semantic-release --dry`) and confirm the predicted next tag. Add a rollback note: if the release tag fails to emit, revert the merge commit on `dev` before the next push.

## Finding 10: Beta-tag verification is a single point in time; backend drift between scope-lock and merge is uncovered

- **Severity:** Medium
- **Location:** Phase 1, "Implementation Steps" step 4 — "`gh api repos/digitopvn/goclaw/compare/v3.12.0-beta.20...43049d3b --jq '.status'` must return `identical`"
- **Flaw:** Phase 1 verifies that **commit** `43049d3b` is in `v3.12.0-beta.20`. It does NOT verify that the **response shapes** in `internal/http/{traces,sessions,channel_instances,activity,logs}.go` haven't drifted on `dev` HEAD between phase 1 (run on day N) and phase 7 (merged on day N+M). Plan §"Risk Assessment" line 137 names this risk ("Backend response shape drift between dev and beta tag") with mitigation "Phase 1 re-verifies contracts" — but only at phase 1. If beta-35 (current latest, 2026-05-27) silently renamed `next_since` to `next_cursor` for traces follow, the CLI shipped against beta-20 contracts will work on beta-20 only.
- **Failure scenario:** Plan executed over 2 weeks. During phase 5, backend lands a refactor in `dev` that renames `spans_by_trace_id` to `spans_by_id`. Phase 1 evidence is stale. Phase 6 live smoke is "optional" (plan.md:119). PR merges. Production users on beta-35 see CLI returning empty spans columns.
- **Evidence:**
  - `plan.md:119`: "Live smoke against `v3.12.0-beta.20`+ backend (optional, after merge)" — "optional" and "after merge" are both fatal.
  - `phase-01-scope-lock.md` does not require a contract re-check at phase 7.
- **Suggested fix:** Phase 6 (Tests and Docs Sweep) MUST add a step: re-run the phase-1 contract verification against `digitopvn/goclaw@dev` HEAD AND against the latest beta tag. If response keys differ from phase-1 evidence, abort and re-spec. Make the live smoke MANDATORY, not optional, run against the latest beta tag at the time of merge.

---

## Unresolved Questions

1. Does the backend actually accept `up_to_index: 0`? If yes (branching at message 0 is meaningful), Finding 1 is Critical. If no (server rejects 0), Finding 1 is High but recoverable.
2. Is `goclaw activity aggregate` the chosen long-term namespace per backend roadmap, or should it sit under `goclaw admin activity aggregate`? Finding 5 resolution depends.
3. What is the canonical `--metadata` syntax across the CLI: `k=v` repeated (this plan) or JSON-string (existing `agents instances update-metadata`)? Finding 7 depends.
4. Should phase 6's `go test` step include `-race` to match the release workflow gate, and should live smoke be promoted from optional to required pre-merge? Finding 9 + 10 depend.
