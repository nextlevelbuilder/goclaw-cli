# Red-Team: Assumption Destroyer — Issue #17 Plan

## Verdict: REVISE

Most structural claims verify, but two findings materially impact the plan:
- The `MapServerCode` table does NOT contain `TRACE_NOT_FOUND` or `PERMISSION_DENIED` today (plan says "may need adding" — it's not "may", it's required).
- The trace detail payload shape (`spans`/`events`/`messages` keys, hierarchical-vs-flat span tree) is purely speculated. No fixture exists in repo. Phase 1 must capture it before Phase 2 can implement `renderTraceDetail`/`buildSpanTree`.

## Findings

### Important — `tracesGetCmd` line number is off by 1
- **Assumption:** "`cmd/traces.go` has `tracesGetCmd` at line 61"
- **Status:** PASS (close enough)
- **Evidence:** `cmd/traces.go:61` reads `var tracesGetCmd = &cobra.Command{`
- **Impact:** None. Plan citation is accurate.

### Important — `unmarshalMap` silently swallows errors
- **Assumption:** "`unmarshalMap` at `cmd/helpers.go:48-53` silently swallows errors"
- **Status:** PASS
- **Evidence:** `cmd/helpers.go:48-53`:
  ```go
  func unmarshalMap(data json.RawMessage) map[string]any {
      var m map[string]any
      _ = json.Unmarshal(data, &m)
      return m
  }
  ```
  The `_ =` discards the error. Same flaw at `unmarshalList` (lines 41-46).
- **Impact:** Confirms Phase 2 step 3 ("Replace silent `unmarshalMap` call with checked unmarshal") is correctly motivated. Note: the same anti-pattern is used across `tracesListCmd`, `usageSummaryCmd`, etc. — fixing only at the `traces get` call site is the minimum fix per the issue scope; the broader code smell is out of scope.

### Important — Printer table fallback dumps JSON
- **Assumption:** "`printer.Print` falls back to JSON dump in table mode at `internal/output/output.go:30-37`"
- **Status:** PASS
- **Evidence:** `internal/output/output.go:31-37`:
  ```go
  default:
      // Table format requires TableData
      if td, ok := data.(*TableData); ok {
          p.printTable(td)
      } else {
          p.printJSON(data) // fallback for non-table data
      }
  ```
  A `map[string]any` is NOT `*TableData`, so it hits `printJSON`. In a TTY this dumps formatted JSON — useless as a "table".
- **Impact:** Confirms the root-cause hypothesis H1 in Phase 1.

### Important — `output.PrintTree` / `TreeNode` exist
- **Assumption:** "`output.PrintTree` and `TreeNode` exist at `internal/output/tree.go:7-36` for reuse"
- **Status:** PASS
- **Evidence:** `internal/output/tree.go:8-12` defines `TreeNode{Name, Children}`; `tree.go:16-27` defines `PrintTree(node, w, prefix, isLast)`; `tree.go:30-35` defines `PrintTreeRoot`.
- **Impact:** Plan's reuse plan in Phase 2 step 2 is valid. **However**: `PrintTree` takes an `io.Writer` — plan must remember to pass `os.Stdout` (the existing API doesn't write to stdout by default).

### Critical — `MapServerCode` does NOT include `TRACE_NOT_FOUND` or `PERMISSION_DENIED`
- **Assumption:** "extend `MapServerCode` if `TRACE_NOT_FOUND` / `PERMISSION_DENIED` not yet mapped" (Phase 2, line 64)
- **Status:** FAIL (semantically — plan hedges "if not yet mapped" but it's NOT mapped; phrasing implies "may" or "may not", which is misleading)
- **Evidence:** `internal/output/exit.go:17-39` `serverCodeMap` contains: `UNAUTHORIZED`, `NOT_PAIRED`, `TENANT_ACCESS_REVOKED`, `NOT_FOUND`, `NOT_LINKED`, `INVALID_REQUEST`, `FAILED_PRECONDITION`, `ALREADY_EXISTS`, `INTERNAL`, `UNAVAILABLE`, `AGENT_TIMEOUT`, `RESOURCE_EXHAUSTED`. **No** `TRACE_NOT_FOUND`. **No** `PERMISSION_DENIED`. Currently both return `ExitGeneric` (1) via `MapServerCode`.
- **Impact:** Phase 2 step 4 must commit to ADDING these codes, not just "confirming". Otherwise the `TestTracesGet_NotFound_ExitCode3` test will fail because the code-string path returns 1, and only the HTTP-status fallback path saves it (404→3 via `MapHTTPStatus` at `exit.go:54`). The `FromError` logic at `error.go:144-154` does fall through to `MapHTTPStatus` when `MapServerCode == ExitGeneric`, so the test will likely pass via the status-code fallback — but the plan should be explicit that the code-string mapping is currently a no-op, and acknowledge that the fallback is what actually saves the assertion.
- **Suggested fix:** Update Phase 2 step 4 from "Confirm `MapServerCode("TRACE_NOT_FOUND")` returns `ExitNotFound`" to "**ADD** `TRACE_NOT_FOUND → ExitNotFound` and `PERMISSION_DENIED → ExitAuth` to `serverCodeMap` in `internal/output/exit.go:17`." Also rerun `internal/output/exit_test.go` after the change.

### Critical — `TRACE_NOT_FOUND` / `PERMISSION_DENIED` are guessed names, not verified
- **Assumption:** "server returns error codes like `TRACE_NOT_FOUND` and `PERMISSION_DENIED`"
- **Status:** UNVERIFIABLE (zero evidence in repo)
- **Evidence:** `grep -rn "TRACE_NOT_FOUND\|PERMISSION_DENIED"` across `/internal/`, `/cmd/`, `/docs/` returns **only** matches inside the plan documents themselves. Sibling test `cmd/traces_follow_test.go` does not exercise an error envelope at all. `apiErrorCodeForStatus` at `cmd/helpers.go:152-172` shows the **CLI's** notion of canonical codes: 403 → `TENANT_ACCESS_REVOKED`, not `PERMISSION_DENIED`. 404 → `NOT_FOUND`, not `TRACE_NOT_FOUND`.
- **Impact:** If the server actually returns `NOT_FOUND` (already mapped to ExitNotFound at `exit.go:24`) and `TENANT_ACCESS_REVOKED` (already mapped to ExitAuth at `exit.go:21`), then no `MapServerCode` extension is needed — the existing mapping already handles them. The plan's Phase 2 step 4 risks adding dead code if `TRACE_NOT_FOUND` is never emitted by the upstream `digitopvn/goclaw` server.
- **Suggested fix:** Phase 1 step 1 (smoke probe) MUST capture an actual 404 error envelope from the live gateway to lock the real `code` field value. Drop speculative codes from Phase 2 plan until that fixture is captured. Tests should assert on the actual code returned, not a guessed name.

### Important — `okJSON` and `runCmd` test helpers exist
- **Assumption:** "`okJSON(t, w, payload)` and `runCmd(t, args...)` helpers exist"
- **Status:** PASS
- **Evidence:** `cmd/phase5_test.go:13-19` defines `okJSON`; `cmd/phase5_test.go:21-27` defines `runCmd`. Both used across `cmd/traces_follow_test.go`, `cmd/activity_aggregate_test.go`, etc.
- **Impact:** None — helpers available.
- **Caveat:** `okJSON` wraps payload in `{"ok": true, "payload": ...}`. If the live `/v1/traces/{id}` envelope is structurally different (e.g. payload at root, no `ok` wrapper, or under a different key), tests built on `okJSON` will be testing a fake envelope. This matters for "TestTracesGet_HappyPath_*" — verify Phase 1 smoke probe captures the **real** envelope before assuming the `okJSON` shape applies.

### Critical — Trace payload shape (`spans`/`events`/`messages`) is speculation
- **Assumption:** "trace payload has `spans` (hierarchical or flat with parent_id), `events`, `messages`"
- **Status:** UNVERIFIABLE
- **Evidence:** Only on-disk reference to these field names in the context of `/v1/traces/{id}`:
  - `cmd/traces.go:51-54` (`tracesListCmd` columns): `trace_id`, `agent_id`, `status`, `duration_ms`, `input_tokens`, `output_tokens`, `cost`. **No** `spans`, `events`, `messages`.
  - `cmd/traces_follow_test.go:34, 130-134`: envelope contains `traces`, `spans_by_trace_id`, `next_since` — but that's the **list/follow** shape, not the single-trace-detail shape.
  - `cmd/traces.go:62` Short: "Get trace with span tree" — implies spans exist but says nothing about format.
  - No fixture file under `cmd/testdata/` (the directory does not exist).
- **Impact:** Phase 2 plans `renderTraceDetail` and `buildSpanTree` against an unknown shape. If spans are returned as `spans_by_trace_id[trace_id]` (the only known wire format from `traces_follow`), span linkage is **flat** under a map keyed by trace_id — different from "hierarchical with `children` or flat with `parent_id`" as plan speculates. Implementation could be built against the wrong shape.
- **Suggested fix:** Phase 1 is mandatory **before** any Phase 2 implementation. Phase 2 step 2 ("Implement `renderTraceDetail`") must explicitly block on Phase 1's captured fixture. Make this dependency loud in `phase-02-cli-fix-output-and-error-mapping.md` (currently only `dependencies: [1]` in frontmatter).

### Minor — No existing secret-stripped fixture pattern in repo
- **Assumption:** "Strip secrets" / "Do NOT commit fixture with auth header or unredacted user IDs"
- **Status:** UNVERIFIABLE (no precedent to follow)
- **Evidence:** `find . -type d -name testdata` returns no results — no `cmd/testdata/` exists, no fixtures committed today.
- **Impact:** Phase 1 will introduce the first on-disk fixture pattern for this repo. Without a precedent, the reviewer has to design the redaction rules. Risk: under-redaction (PII slips through) or over-redaction (fixture becomes useless for testing real shape edge cases).
- **Suggested fix:** Phase 1 should list explicit fields to scrub: `user_id`, `tenant_id`, `agent_id` content if proprietary, `messages[].content`, `events[].args`, any token-bearing fields. Use sentinel placeholders (`user-redacted`, `tenant-redacted`, `[REDACTED]`). Document the redaction rules in the report so future trace fixtures follow them.

### Minor — `Closes #17` auto-close requires merge to default branch
- **Assumption:** "PR auto-close via `Closes #17` only fires when merged into the default branch (`main`). PR targets `dev` first." (phase-03, risk section)
- **Status:** PASS (correctly self-flagged in plan)
- **Evidence:** Default branch is `main` (per gitStatus header and CLAUDE.md). Plan targets `dev`. GitHub auto-closes referenced issues only when the PR merges to the default branch. The plan already calls this out and proposes manual close after dev→main promotion.
- **Impact:** None — plan is self-consistent. Issue #17 is verified to exist in `nextlevelbuilder/goclaw-cli` (the correct repo per `git remote -v`).

### Minor — `Get` returns `json.RawMessage` (envelope already unwrapped)
- **Observation (bonus):** `HTTPClient.Get` at `internal/client/http.go:47` returns `(json.RawMessage, error)`. The plan's diagrams hint at this but never explicitly state what `data` is when `c.Get(...)` returns. Worth verifying whether `Get` strips the `{ok, payload}` envelope before returning — otherwise `unmarshalMap(data)` operates on the envelope, not the inner payload, and tests using `okJSON` wrap exactly the envelope shape that `Get` would already have unwrapped.
- **Status:** UNVERIFIABLE without reading `internal/client/http.go` in full — but the plan does not document this contract, and the test envelope wrapping in `okJSON` strongly implies `Get` does unwrap. Phase 1's repro test must confirm the layer at which `data` is observed.
- **Suggested fix:** Phase 1 step 1 add a sub-step: "Document what shape `data := c.Get(...)` returns — envelope or payload — by reading `internal/client/http.go`'s Get implementation."

## Open Questions

- Does the live `/v1/traces/{id}` endpoint exist on the current `digitopvn/goclaw` server, or is the path different (e.g. `/v1/traces/get/{id}`, `/v1/traces?id=...`, `/v1/tenants/{tenant}/traces/{id}`)? Plan assumes the path works; H5 in Phase 1 nominally checks this but doesn't allocate a discrete acceptance criterion to the answer.
- Are span IDs/parent linkage fields named `span_id`/`parent_id`, `id`/`parent`, `id`/`parent_span_id`, or something else? `renderTraceDetail`/`buildSpanTree` cannot be designed until the fixture lands.
- If the server returns a 401/403 due to tenant mismatch on `traces get` but allows `traces list`, that's a server bug, not a CLI bug — should Phase 3's upstream-issue path be widened to "auth/tenant filtering" beyond pure not-found regressions?

Status: DONE
Severity counts: 3 Critical, 5 Important, 3 Minor.
