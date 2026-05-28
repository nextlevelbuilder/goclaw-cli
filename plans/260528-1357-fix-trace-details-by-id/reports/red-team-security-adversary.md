# Red-Team: Security Adversary — Issue #17 Plan

## Verdict: REVISE

Plan is sound on the functional axis but ships with several under-specified security gates around (a) fixture/issue-body scrubbing, (b) path traversal via the trace id parameter, (c) ANSI injection parity claim, and (d) information-disclosure trade-off in the 404-vs-403 message split. None are showstoppers — all are addressable by tightening Phase 1 step 2 and Phase 2's security-considerations block before code lands.

## Findings

### Critical — Path traversal surface via `url.PathEscape` for trace id

- **Threat:** `url.PathEscape` does NOT encode `.`, so `goclaw traces get ..` produces request path `/v1/traces/..`. `url.PathEscape("../../etc/passwd")` does encode the `/` as `%2F`, but `url.PathEscape("..")` returns the literal `".."`. If a future server route/middleware on `digitopvn/goclaw` ever path-normalizes before authz (or proxies through a path-rewriting reverse proxy like nginx with `merge_slashes`/`alias`), the CLI becomes the easiest tool to probe traversal.
- **Status:** HOLE (CLI-side input validation gap).
- **Evidence:**
  - Verified empirically: `url.PathEscape("..") == ".."`, `url.PathEscape("./../x") == ".%2F..%2Fx"`, `url.PathEscape("../../etc/passwd") == "..%2F..%2Fetc%2Fpasswd"`. Run output captured in this review session.
  - Plan phase-02-cli-fix-output-and-error-mapping.md:84 — `url.PathEscape(id)` is the only id sanitization step.
  - Plan phase-02-cli-fix-output-and-error-mapping.md:84 — id validation is `strings.TrimSpace(id) == ""` only.
- **Attack scenario:** Operator who has a valid auth token but limited trace-read scope runs `goclaw traces get ..` or `goclaw traces get .`. The server receives `GET /v1/traces/..` — at minimum a noisy/unhandled route; at worst, on a future server build that adds `path.Clean` ahead of authz, lands on `/v1/traces/` (list endpoint) under the singular's authz check.
- **Suggested mitigation:** In Phase 2 step 3 id validation, after `TrimSpace` add:
  ```
  if id == "." || id == ".." || strings.ContainsAny(id, "/\\") {
      return fmt.Errorf("invalid trace id")
  }
  ```
  And add a Phase 1 test case `TestTracesGet_RejectsTraversalIDs` covering `..`, `.`, `../foo`, `foo/bar`, `\x00`. Also: tighten to a positive regex (`^[A-Za-z0-9_-]{1,128}$`) if Phase 1's live capture shows trace ids are opaque ULIDs/UUIDs (most likely). Note: this collapses to AC #3's "malformed id" category.

---

### Critical — Fixture/issue-body scrubbing protocol is hand-wavy

- **Threat:** Phase 1 step 2 says "Strip secrets" and Phase 1 Security Considerations says "Strip `user_id`, `tenant_id`, message contents, tool-call args. Keep only structural keys + sentinel values." But (a) no `jq` recipe is given, (b) the scrub list is incomplete — span `name` fields can contain user queries (e.g. `name: "search('user@example.com')"` ), tool-call `result` payloads may contain PII, HTTP-request span attributes can contain auth headers, and (c) Phase 3 step 5 may file an upstream issue using a body drafted in Phase 1 — same payload, same scrub requirement, but Phase 3 doesn't re-state it.
- **Status:** HOLE (operational/process gap; high likelihood of accidental commit of PII).
- **Evidence:**
  - Phase 1 step 2 line 70: "Strip secrets." (no how).
  - Phase 1 Security Considerations line 102 enumerates 4 fields but misses span `name`, span `attributes`, event `message`/`error`, headers in HTTP spans, trace metadata bags.
  - Phase 3 step 5 line 52: issue body reuses the Phase-1 drafted body, but Phase 3 Security Considerations line 76 only says "Smoke probe artifacts ... must be scrubbed" without listing the issue-body case explicitly.
- **Attack scenario:** Operator captures a live trace whose span name is `OpenAI.chat({"messages":[{"role":"user","content":"my SSN is 123-45-6789"}]})`. Operator strips only the 4 listed keys, commits fixture to `cmd/testdata/trace_detail_get.json`, and the PII enters git history. Same payload pasted into the `digitopvn/goclaw` issue body via `gh issue create` becomes public on a different repo.
- **Suggested mitigation:** Add to Phase 1 step 2 (and reference from Phase 3 step 5):
  ```
  # Scrub recipe (run once; assume captured response in raw.json):
  jq '
    walk(if type == "object" then
      with_entries(
        if .key | IN("user_id","tenant_id","org_id","email","token",
                     "authorization","auth","api_key","secret",
                     "messages","tool_calls","arguments","result",
                     "input","output","prompt","completion","content",
                     "attributes","metadata","headers","query","body","error")
        then .value = "REDACTED"
        else . end)
    else . end)
    | .payload.trace_id = "trc_XXXXXXXXXXXXXXXX"
    | .payload.agent_id = "agt_XXXXXXXX"
  ' raw.json > cmd/testdata/trace_detail_get.json
  ```
  Then `rg -i 'email|@|bearer|sk-|eyJ' cmd/testdata/trace_detail_get.json` as a post-scrub gate. Reuse the SAME pipeline for the upstream-issue body. Phase 3 step 5 should explicitly include `gh issue create --body-file scrubbed.md`, never inline.

---

### Critical — `curl` command in repro report risks token leakage

- **Threat:** Phase 1 step 1 instructs running `curl -sS -H "Authorization: Bearer $TOKEN" ...` AND Phase 1 step 5 says to include "Curl command + response samples" in `reports/repro-260528-issue17.md`. If the operator copy/pastes the executed command from shell history, the token expands.
- **Status:** HOLE (documentation/operational; medium likelihood — depends on operator hygiene).
- **Evidence:**
  - Phase 1 step 1 line 69: `curl -sS -H "Authorization: Bearer $TOKEN" "$SERVER/v1/traces/<id>"` — uses env var here, good.
  - Phase 1 step 5 line 81: "Curl command + response samples (redacted)" — no explicit guard against pasting the expanded token.
- **Attack scenario:** Operator runs the curl in zsh with history-expansion or pastes from `history` output where the variable was expanded (some shells do this on certain `set` configs). Token enters the report → enters git → gets pushed.
- **Suggested mitigation:** Phase 1 step 5: explicit rule — "Quote curl as literal: `curl -H 'Authorization: Bearer $GOCLAW_TOKEN' ...` (single-quoted, never expanded). Add a pre-commit grep gate: `! rg -n 'Bearer [A-Za-z0-9._\-]{20,}' plans/`." And: set `HISTFILE=/dev/null` or prefix with space-then-command to skip history (`setopt HIST_IGNORE_SPACE`/`HISTCONTROL=ignorespace`) before running the probe.

---

### Important — ANSI escape injection: parity claim is half-true

- **Threat:** Phase 2 Security Considerations line 123 defers ANSI sanitization with "accept that terminal rendering of untrusted text is consistent with `traces list` (existing risk surface)." Verified: `tracesListCmd` does render user-controlled strings (`agent_id`, `status`) into table cells (cmd/traces.go:53). So the parity claim has factual basis — but the surface AREA is different. `traces list` renders ~7 short flat fields (id, agent, status, ms, tokens, tokens, cost) which are server-generated. The new `traces get` will render span `name`, event `message`, possibly tool-call args — fields that are far more likely to contain attacker-controlled content (user prompts, LLM outputs, tool stdout/stderr). LLM output is the canonical ANSI-injection vector in 2026.
- **Status:** HOLE — parity is structurally true (no existing scrubbing) but the risk-weighted exposure is materially higher. The "accept parity" decision should be made by the user, not the planner.
- **Evidence:**
  - cmd/traces.go:53 — `str(t, "agent_id")` etc. unsanitized into table.
  - `rg "safeRune|ANSI" internal/output/ cmd/` returns no hits — no current ANSI sanitization anywhere.
  - Phase 2 line 79 — TreeNode names will include `<span_id> [<duration_ms>ms] <name>` where `<name>` is user/LLM-controlled.
- **Attack scenario:** Adversarial LLM (jailbroken or prompt-injected via tool output) emits a span name containing `\x1b[2J\x1b[H` (clear screen) + `\x1b]0;rm -rf ~\x07` (set terminal title) + a fake "operation succeeded" message. Operator runs `goclaw traces get <id>` in a terminal, sees a forged success line, may run further commands assuming the trace is benign. With OSC 52 (`\x1b]52;c;<base64>\x07`), the span name can quietly overwrite the operator's clipboard with arbitrary text.
- **Suggested mitigation:** Phase 2 — implement the `safeRune` helper now (10 lines, no extra dependency). Strip `\x1b`, `\x07`, all `C0` controls except `\t`. Apply in `renderTraceDetail` and `buildSpanTree`. Apply the same to `tracesListCmd` while you're in the area (one-line wrap around `str(...)`). Document the helper in `internal/output/text_sanitize.go`. Cost: ~20 LOC; benefit: closes the entire family.

---

### Important — 404 vs 403 message split is an existence oracle (intentional?)

- **Threat:** Plan deliberately splits "permission denied for trace `<id>`" (403/PERMISSION_DENIED) from "trace `<id>` not found" (404/TRACE_NOT_FOUND). That distinction reveals whether a trace id exists across tenants. AC #3 from issue #17 demands the distinction, so this is a documented trade-off — but the plan does not name it as a trade-off.
- **Status:** HOLE (information disclosure) but intentional per AC; needs explicit acknowledgment.
- **Evidence:**
  - Phase 2 Architecture line 42-45: distinct messages by code.
  - plan.md AC #3 line 33: "Errors clearly distinguish: not found, permission denied, malformed id, and server/API failure."
- **Attack scenario:** Attacker with valid token for tenant A iterates plausible trace-id prefixes. `403 permission denied for trace X` confirms X exists under tenant B; `404 not found` rules X out. Over time the attacker enumerates cross-tenant trace ids — useful for crafting follow-on social-engineering or for confirming activity against a known target.
- **Suggested mitigation:** This is primarily a server-side concern (the server is choosing to return 403 vs 404 distinct codes), and the CLI faithfully surfaces what the server returns. Two options:
  1. Defer to server: add a Phase 2 note "the 403/404 distinction is the server's choice; CLI surfaces it per AC #3. If the server later merges to a single 404 for security, CLI behavior follows automatically because `MapHTTPStatus(404) -> 3` covers both."
  2. Add a `--paranoid` flag (out of scope; just flag in plan).
  Recommend option 1: add 3-line "Information disclosure trade-off" subsection to Phase 2 Security Considerations naming this explicitly so the next reviewer doesn't re-litigate it.

---

### Important — Missing server-code entries for `TRACE_NOT_FOUND` / `PERMISSION_DENIED`

- **Threat:** Plan Phase 2 step 4 line 91-92 says "Confirm `MapServerCode('TRACE_NOT_FOUND')` returns `ExitNotFound`. If not, extend the switch." Verified: `internal/output/exit.go:17-39` does NOT contain `TRACE_NOT_FOUND` or `PERMISSION_DENIED` / `FORBIDDEN`. So `MapServerCode` returns `ExitGeneric` (1) for both, and the code falls through to `MapHTTPStatus` via `FromError` (internal/output/error.go:148-152). That path works iff `APIError.HTTPStatus()` is populated — and that's a non-trivial invariant to depend on.
- **Status:** HOLE (latent fragility, not exploit-grade).
- **Evidence:**
  - `internal/output/exit.go:17-39` — no TRACE_NOT_FOUND, no PERMISSION_DENIED, no FORBIDDEN.
  - `internal/output/error.go:148-152` — fallback only fires when `errors.As(err, &aws)` succeeds AND `HTTPStatus() > 0`. If the server returns a JSON error envelope WITHOUT a status code surface in `APIError`, exit code is 1 instead of 2/3.
- **Attack scenario:** Not exploit; degraded UX. Automation script checking `$? == 2` for auth failures sees `$? == 1` and treats it as unknown error — bad escalation behavior.
- **Suggested mitigation:** Phase 2 step 4 — make the extension non-conditional. Add `TRACE_NOT_FOUND -> ExitNotFound`, `PERMISSION_DENIED -> ExitAuth`, `FORBIDDEN -> ExitAuth` to `serverCodeMap` unconditionally. Verify with `TestMapServerCode_TraceCodes`. Cost: 4 lines + 1 test.

---

### Important — JSON full-payload preservation may leak internal fields

- **Threat:** Phase 2 line 20-21 promises "JSON/YAML mode emits the **complete** server payload — no field dropping, no schema reshaping." That's the AI-ergonomics contract, but it converts a previous bug (silent JSON dump in table mode showed everything anyway) into an intentional, documented passthrough. If the server payload includes internal-only fields (`_internal_user_id`, debug counters, `__raw_sql`, etc.) the CLI now reliably exposes them. Previously, a user in TTY who saw a JSON blob might not parse it; now they get a JSON-structured output that is easier to grep/extract.
- **Status:** MITIGATED (within CLI scope) / HOLE (cross-boundary, server's responsibility).
- **Evidence:**
  - Phase 2 line 20-21 — explicit full-passthrough decision.
  - Phase 1 Architecture line 36 — current code already does `printer.Print(unmarshalMap(data))` which in JSON mode passes through fully. So the surface is NOT new — but the *test* in Phase 2 line 70 (`TestTracesGet_JSONMode_PreservesNestedFields`) freezes the contract as "every field round-trips" — which becomes a forcing function against future server-side scrubbing.
- **Attack scenario:** Server team adds an `_internal_debug_sql` field for observability; CLI test fails because the field is dropped/scrubbed; pressure pushes back on the server scrubbing.
- **Suggested mitigation:** Phase 2 — relax the test from "every input field present in output" to "every input field UNDER `payload` present in output, where input field name does not start with `_`" or "explicit whitelist of known-public fields from the captured fixture." Document the rule in `internal/output/error.go` or a new `docs/trace-payload-contract.md`. Cost: 3-line test refinement, 5-line doc.

---

### Minor — Fixture contains real `agent_id` / `trace_id` even after scrub

- **Threat:** Plan Phase 1 says "Keep only structural keys + sentinel values" but does not specify the sentinel format. Real trace ids and agent ids — even without other PII — can be cross-referenced against logs by anyone with operational access to the gateway. They are not secret but they ARE indirect identifiers (tying CLI test commits to specific live traces).
- **Status:** MITIGATED-IF-FOLLOWED (depends on operator using sentinels). Covered by the `jq` recipe in the Critical finding above.
- **Evidence:** Phase 1 Security Considerations line 102 mentions sentinel values but doesn't define them.
- **Attack scenario:** Marginal — researcher cross-references `trace_id` in fixture against gateway logs to identify a real user session. Low-likelihood, low-impact, but free to mitigate.
- **Suggested mitigation:** Sentinel pattern (already in the `jq` recipe above): `trc_XXXXXXXXXXXXXXXX`, `agt_XXXXXXXX`, `usr_XXXXXXXX`. Add a `TestFixtureContainsOnlySentinels` test that asserts the fixture matches `^[a-z]{3}_X+$` on id fields, blocking accidental real-id regression.

---

### Minor — Timing oracle: not present in current plan

- **Threat:** "Does a test for exit code 3 measure HTTP latency that leaks whether trace existed pre-auth-check?"
- **Status:** NOT_APPLICABLE.
- **Evidence:** Phase 1 and Phase 2 tests assert only exit code and message string, no latency assertions. Tests use `httptest.NewServer` which has near-zero latency anyway.
- **Attack scenario:** None.
- **Suggested mitigation:** None. Flag if any future PR adds `time.Since(start)` assertions on the error path.

---

## Severity Counts

- Critical: 3 (Path traversal, Fixture scrubbing recipe, Curl token leakage)
- Important: 4 (ANSI parity claim, 404/403 oracle, Missing server-code entries, JSON passthrough contract test)
- Minor: 2 (Sentinel format, Timing oracle [N/A])

## Open Questions

- Is the trace id format actually `^[A-Za-z0-9_-]{N}$` (typical ULID/UUID/nanoid)? If yes, regex validation closes path traversal cleanly. If trace ids can contain arbitrary characters (e.g. user-supplied custom ids), validation must be looser and traversal mitigation falls back to explicit `..`/`.` rejection. Phase 1's live capture will resolve this.
- Does `digitopvn/goclaw` server treat `PERMISSION_DENIED` and `FORBIDDEN` interchangeably, or is one the convention? Phase 2 step 4 should grep the server repo (if accessible) to confirm.
- Is there a workspace-wide pre-commit hook for secret scanning (e.g. `gitleaks`, `trufflehog`) that would catch a leaked token in the fixture? If yes, the curl-token mitigation can lean on it. If no, this should become a `.git/hooks/pre-commit` addition in Phase 3.

Status: DONE
