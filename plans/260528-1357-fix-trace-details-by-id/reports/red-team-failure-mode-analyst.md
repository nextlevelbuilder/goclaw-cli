# Red-Team: Failure Mode Analyst — Issue #17 Plan

## Verdict: REVISE

Plan is mostly sound but ships with **4 critical correctness gaps** that will cause test flakes or false-positive merges, and **3 important gaps** in test scaffolding/spec language. Two minor items also called out. Recommend revisions to phase-01 and phase-02 before TDD lands.

## Failure Modes

### Critical — Auto-retry inflates request count on 5xx/429 tests
- **Mode:** `internal/client/http.go` `do()` retries up to **3 attempts** on `StatusCode == 429 || StatusCode >= 500`, with 1s + 2s backoff between attempts. Phase-2 test `TestTracesGet_ServerError_ExitCode5` will see **3 hits**, not 1, plus a ~3s sleep cost. Phase-1 test `TestTracesGet_MalformedID_ValidatedClientSide` `calls == 0` assertion is fine, but any future "exactly N calls" assertion on 5xx is wrong.
- **Status:** EXPOSED
- **Evidence:** `internal/client/http.go:138-168`: `for attempt := range 3 { ... if resp.StatusCode != 429 && resp.StatusCode < 500 { break } ... if attempt < 2 { time.Sleep(time.Duration(1<<attempt) * time.Second) }`. The plan's `TestTracesGet_ServerError_5xx` description in phase-01.md:78 and phase-02.md:74 says "server returns 500; assert distinct error category" — silent on call count, but the ~3s sleep will slow the whole suite.
- **Impact:** Test suite gets ~3s slower per 5xx case; if anyone asserts `calls == 1` they get a red. 403 also EXPOSED — `403 < 500` and `!= 429`, so no retry there (good), but worth pinning.
- **Suggested mitigation:** phase-02 §1 — add to `TestTracesGet_ServerError_ExitCode5`: "server handler counts calls; assert `calls == 3` (auto-retry contract) and uses `time.Sleep` budget < 4s or stub clock". phase-02 §3 — note `403 PERMISSION_DENIED` does NOT retry (good). Optionally inject a retry-disabled client for tests via env knob or `HTTPClient.MaxRetries` — but that's a P3 scope decision.

### Critical — `MapServerCode` extension would cross-impact every other command
- **Mode:** Phase-02 §4 says "extend `MapServerCode` if `TRACE_NOT_FOUND` / `PERMISSION_DENIED` not yet mapped." But the existing map (`internal/output/exit.go:17-39`) uses `NOT_FOUND` (generic) and `TENANT_ACCESS_REVOKED` (auth-403). Adding `TRACE_NOT_FOUND` is **safe** (it's net-new). Adding `PERMISSION_DENIED` would be a **new key** that other commands' server responses may already use under another flow, silently re-routing their exit codes from `1` (unmapped/generic) to `2` (auth). No cross-impact audit in the plan.
- **Status:** EXPOSED
- **Evidence:** `internal/output/exit.go:17-39` shows current map; `cmd/helpers.go:152-172` `apiErrorCodeForStatus` returns `TENANT_ACCESS_REVOKED` for 403, never `PERMISSION_DENIED`. But server may return `PERMISSION_DENIED` directly as a body code — `apiErrorFromRawBody` at `cmd/helpers.go:139-143` would pass it through verbatim. Today it'd hit `MapServerCode` and fall through to `MapHTTPStatus(403)` → `ExitAuth` (already correct). So adding it to the map is **a no-op for traces** (HTTP fallback already handles it) but **a behavior change** for any caller whose server returns `PERMISSION_DENIED` with status 200 / no-status / 5xx.
- **Impact:** Risk is low but the plan doesn't make the trade-off explicit. Worse: extending the map silently changes the value of `MapServerCode("PERMISSION_DENIED")` from `ExitGeneric` to `ExitAuth`, which is observable via the existing `MapServerCode` unit tests in `internal/output/exit_test.go:43-46` — those tests verify "unknown returns `ExitGeneric`", not specifically `PERMISSION_DENIED`, so they pass either way. But any future test/caller relying on the current generic mapping breaks silently.
- **Suggested mitigation:** phase-02 §4 — say "Do **not** extend `MapServerCode`; rely on `MapHTTPStatus` 403 fallback. Add a unit test in `internal/output/exit_test.go` pinning `MapHTTPStatus(403) == ExitAuth` and `MapHTTPStatus(404) == ExitNotFound` for the explicit `PERMISSION_DENIED` / `TRACE_NOT_FOUND` case names." If extension is required, add a separate **cross-impact note** listing other call sites that today emit those codes.

### Critical — `runCmd` test harness has zero output-format isolation
- **Mode:** `runCmd(t, args ...string)` in `cmd/phase5_test.go:21-27` just calls `rootCmd.Execute()`. The output format is decided by `PersistentPreRunE` (`cmd/root.go:21-41`) which calls `ResolveFormatWithDefault(flagVal, cfg.ProfileOutputFormat)`. When tests run under `go test`, stdout is **piped** (not TTY), so `IsTTY(...)` returns `false` and `ResolveFormat` returns `"json"` — not `"table"`. Plan's `TestTracesGet_HappyPath_Table` (phase-01.md:74) and `TestTracesGet_TableMode_HumanReadable` (phase-02.md:69) will **never see table mode** unless they explicitly pass `--output table`. The plan does not mention this.
- **Status:** EXPOSED
- **Evidence:** `internal/output/tty.go:31-57` — `ResolveFormatWithDefault` falls through to `IsTTY(int(os.Stdout.Fd()))`, which under `go test` is `false`. Confirmed by `cmd/heartbeat.go:184` which also checks `IsTTY` directly. No mocking of `IsTTY` exists in the codebase.
- **Impact:** Phase-01 and phase-02 table-mode assertions fail for the wrong reason — they fail because format defaulted to JSON, not because rendering is broken. Tester wastes time chasing a phantom render bug.
- **Suggested mitigation:** phase-01 §3 and phase-02 §1 — every table-mode test MUST pass `--output table` explicitly: `runCmd(t, "--output", "table", "traces", "get", id)`. Add an explicit note in `phase-01.md` Non-functional requirements: "All TTY-mode assertions force format via `--output table` flag; do not rely on TTY auto-detect." Equivalently, every JSON-preservation test should pass `--output json` to remove ambiguity.

### Critical — `resetTracesGetFlags` is undefined; flag persistence risks contamination
- **Mode:** Phase-01.md:23 mentions `resetTracesGetFlags(t)` "if `tracesGetCmd` adds flags later". But `runCmd` calls `rootCmd.SetArgs(args)` and `Execute()`. Cobra's persistent flags (`--output`, `--server`, etc.) **stick** across `Execute()` calls because they're registered on `rootCmd.PersistentFlags()` once. Setting `--output table` in test A then NOT setting it in test B leaves `--output=table` set, breaking test B's default auto-detect. The pattern `resetActivityAggregateFlags` (`cmd/activity_aggregate_test.go:11-17`) only resets local flags, not persistent ones.
- **Status:** EXPOSED
- **Evidence:** `cmd/root.go:71-81` registers `output` on `PersistentFlags()`. `cmd/phase5_test.go:21-27` `runCmd` calls `SetArgs` but never resets persistent flag values. `cmd/p4_ux_polish_test.go:408` `resetTestFlag` exists but operates per-flag on a passed command. No global persistent-flag reset helper in the codebase.
- **Impact:** Test order dependency. `go test -count=1 -run ...` may pass; `go test ./...` may flake depending on which test sets `-o json` last.
- **Suggested mitigation:** phase-01 §3 — define `resetTracesGetFlags(t)` that resets `rootCmd.PersistentFlags()` `output`, `server`, `token`, `quiet`, `verbose`, `insecure` AND any local `tracesGetCmd` flags (currently none). Use `t.Cleanup(func() { resetTracesGetFlags(t) })` in every test. Also reset between test runs because cobra-persistent flags carry across.

### Important — JSON-mode "preserves full payload" claim is technically false
- **Mode:** Phase-02 acceptance: "JSON/YAML mode emits the **complete** server payload — no field dropping, no schema reshaping." Implementation path is `printer.Print(decoded)` where `decoded := map[string]any`. `printJSON` uses `json.NewEncoder(os.Stdout).Encode(...)` (`internal/output/output.go:57-61`) which (a) re-orders keys alphabetically per Go's `map[string]any` marshaling, (b) escapes HTML by default (`<`, `>`, `&` → `<`...), (c) does NOT preserve original byte-for-byte. So "preserves nested fields" is true, "preserves full payload" byte-equivalent is false.
- **Status:** EXPOSED
- **Evidence:** `internal/output/output.go:57-61` `enc := json.NewEncoder(os.Stdout); enc.SetIndent(...)` — no `SetEscapeHTML(false)` call. Go stdlib default is HTML-escape ON.
- **Impact:** Trace span content with `<` `>` `&` (e.g. tool calls with XML/HTML in args, or `&&` in shell commands) will have those chars escaped in output. Round-trip `jq` parse still works (escapes are valid JSON), but byte-comparison fails. Test `TestTracesGet_JSONMode_PreservesNestedFields` description (phase-02.md:70-71) says "round-trip parse; assert every field present" — that's the *correct* assertion. But the **plan prose** at phase-02.md:20 overpromises.
- **Suggested mitigation:** phase-02 — soften prose: "JSON/YAML mode emits the complete server payload as a structured object — every field round-trips through `jq`, but bytes are re-encoded (key order normalized, HTML chars escaped per Go stdlib default)." Optionally pass-through `json.RawMessage` instead of `map[string]any` to preserve bytes; but that's a bigger change.

### Important — span shape assumptions are unverified before code lands
- **Mode:** Phase-02 §2 says "if spans hierarchical, recurse on `children`; if flat with `parent_id`, build adjacency map first." But: (a) what if **some** traces have `spans: []` (no spans)? (b) what if `spans` field is absent / `null`? (c) what if shape is **mixed** — top-level flat list with `children: []` already populated by the server? (d) what about cycles in `parent_id` graph (malicious / bug)? (e) what if `events: null` instead of `[]`? (f) what about giant span trees (1000+ nodes) — does `PrintTree` indent prefix grow unboundedly?
- **Status:** EXPOSED
- **Evidence:** Phase-01 fixture-capture is the only place where the actual shape would be locked, but Phase-01 risk-mitigation (phase-01.md:96-97) says "if server unreachable, fixture derived from `traces follow` shape with TODO" — meaning code lands without verified shape if smoke probe fails. `internal/output/tree.go:14-30` `PrintTree` is unguarded against deep recursion.
- **Impact:** Render helper may panic on `nil` `events`, infinite-loop on cyclic `parent_id`, or produce useless output on absent `spans`.
- **Suggested mitigation:** phase-02 §2 — explicit defensive list: "(a) `spans == nil` or `len(spans) == 0` → print `(no spans)` line, skip tree; (b) `events == nil` → treat as `[]`; (c) `parent_id` adjacency map MUST track visited to break cycles; (d) hard-cap tree depth at 50 levels with `...` truncation; (e) hard-cap printed events at 50 even when n <= 10."

### Important — `str`/`safeFloat` helpers — `safeFloat` does not exist
- **Mode:** Phase-02.md:80-81 says "every field access through `str(m, "key")` / `safeFloat(m, "key")` helpers." `str` exists at `cmd/helpers.go:56-61`. `safeFloat` does **not** exist anywhere in `/cmd/` or `/internal/`. Plan does not specify who writes it or where it lives.
- **Status:** EXPOSED
- **Evidence:** `grep -rn "safeFloat" --include="*.go"` returns zero hits in the worktree. `str` defined `cmd/helpers.go:56-61`.
- **Impact:** Phase-02 implementation halts at compile error or scope-creep into `cmd/helpers.go`. Defensive numeric coercion of `map[string]any` (where JSON numbers decode as `float64`, but server may emit them as strings) is non-trivial and untested.
- **Suggested mitigation:** phase-02 Related Code Files — explicitly add `cmd/helpers.go` modification: "Add `safeFloat(m map[string]any, key string) float64` that handles `float64`, `int`, `int64`, `json.Number`, and string-encoded numbers; returns 0 on failure. Add unit tests in `cmd/helpers_test.go`."

### Important — `url.PathEscape` "validate non-empty" is underspecified
- **Mode:** Phase-02 §3 step: "Validate id: `id := args[0]; if strings.TrimSpace(id) == "" { return fmt.Errorf("trace id is required") }`." This only catches whitespace/empty. It does NOT reject path-traversal (`../../etc/passwd`), URL-control chars (`?foo=bar` injecting query, `#frag` injecting fragment), or slash (`a/b` becoming `/v1/traces/a/b` — but `url.PathEscape` does escape `/` to `%2F`, so that's safe). What does the issue reporter consider "malformed"? The plan never says.
- **Status:** EXPOSED
- **Evidence:** `url.PathEscape` Go stdlib docs: escapes `/`, `?`, `#`, etc. — so `goclaw traces get '../../admin'` becomes `GET /v1/traces/..%2F..%2Fadmin` which the server can choose to 404 or 400. Not a CLI security hole (server is the trust boundary). But test `TestTracesGet_MalformedID_ValidatedClientSide` (phase-02.md:73) asserts `calls == 0` — meaning the CLI MUST short-circuit before HTTP, but the plan never lists which inputs count as "malformed client-side".
- **Impact:** Test is ambiguous. Implementer may pick a different set than tester expects.
- **Suggested mitigation:** phase-02 §3 — explicit list: "client-side rejects only (a) empty after trim, (b) length > 256 (DoS guard). All other inputs go through `url.PathEscape` and let the server decide. Test passes `""`, `"   "`, and a 1KB string for `calls == 0` assertions; passes `"../../admin"` for `calls == 1` with the server returning 400."

### Minor — Phase-01 smoke probe has TODO escape; Phase-02 manual smoke has none
- **Mode:** Phase-01 risk mitigation explicitly allows "if server unreachable, mark TODO". Phase-02 §6 says "Live smoke (manual) — re-run the four `goclaw traces get` invocations from Phase 1's smoke probe; confirm now produces useful output" with no fallback.
- **Status:** EXPOSED
- **Evidence:** phase-02.md:102.
- **Impact:** Phase-02 success-criteria "Manual smoke: ... renders a readable summary" (phase-02.md:111) can't be checked off if no gateway is reachable; ship pipeline stalls.
- **Suggested mitigation:** phase-02 §6 — add "If gateway unreachable, mark manual-smoke as `DEFERRED` and gate the PR description to require a smoke-check before merge to `main`."

### Minor — Phase-3 codebase-summary update target verified
- **Mode:** Plan promises a bullet under "output/render section" in `docs/codebase-summary.md`.
- **Status:** MITIGATED
- **Evidence:** `docs/codebase-summary.md` line 149 has `#### output/ — Output Formatting + Error Handling`, line 587 references tree rendering. Section exists.
- **Impact:** None — claim is true.
- **Suggested mitigation:** None.

### Minor — `gh issue create -R digitopvn/goclaw` rights not pre-verified
- **Mode:** Phase-03 §5 conditionally runs `gh issue create -R digitopvn/goclaw`. If the user lacks issue-creation rights on that repo, the command fails mid-ship.
- **Status:** EXPOSED
- **Evidence:** phase-03.md:52. No `gh api repos/digitopvn/goclaw -q .permissions` pre-check.
- **Impact:** Ship pipeline stops with a permission error instead of completing the docs update.
- **Suggested mitigation:** phase-03 §5 — prepend `gh api repos/digitopvn/goclaw -q .permissions.push 2>/dev/null` check; on failure, drop the drafted issue body into `plans/.../reports/upstream-issue-draft.md` and instruct the user to file it manually.

### NOT_APPLICABLE — 403 retry concern
- **Mode:** "Does the client retry on 403?"
- **Status:** NOT_APPLICABLE
- **Evidence:** `internal/client/http.go:161`: `if resp.StatusCode != 429 && resp.StatusCode < 500 { break }`. 403 breaks immediately. No retry. No mitigation needed.

## Open Questions

- Should `HTTPClient` gain a `MaxRetries` knob to disable retries in tests, or is the ~3s 5xx-test cost acceptable (5xx tests are rare)?
- Does the issue reporter's actual repro live trace_id still exist on `goclaw.zuey.me`? If not, what's the smoke-probe fallback gateway?
- Does the server emit `PERMISSION_DENIED` as a body code anywhere, or always `TENANT_ACCESS_REVOKED`? Determines whether `MapServerCode` extension is actually a no-op or a behavior change.
- For deeply nested span trees (>50 levels), is truncation `... (N more levels)` acceptable, or should the tree render still walk but flatten?

Status: DONE
Severity counts: Critical=4, Important=4, Minor=3, NOT_APPLICABLE=1
