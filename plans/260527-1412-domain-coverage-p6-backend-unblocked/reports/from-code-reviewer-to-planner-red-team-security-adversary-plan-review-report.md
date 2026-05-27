# Red-Team Plan Review — Security Adversary Lens

Plan: `plans/260527-1412-domain-coverage-p6-backend-unblocked/`
Reviewer role: Security Adversary + Fact Checker
Date: 2026-05-27

## Finding 1: Top-level `activityCmd` already exists — naming collision missed
- **Severity:** Critical
- **Location:** plan.md "Existing CLI State" table + phase-05 "Files"
- **Flaw:** Plan asserts "no `cmd/admin_activity.go`" and "No command-name collisions" and then proposes creating `activityCmd` as a new top-level Cobra parent. An `activityCmd` variable is already declared and registered at the root in `cmd/admin.go`.
- **Failure scenario:** Phase 5 implementation will produce a Go compile error (duplicate `activityCmd` declaration) at minimum. If renamed to dodge the conflict, the user will get two distinct meanings for `goclaw activity ...` (audit-log list vs aggregate) or a runtime Cobra "command already added" panic. Reviewers seeing "no collisions" will not check this until phase 5 fails.
- **Evidence:**
  - `cmd/admin.go:133` — `var activityCmd = &cobra.Command{ Use: "activity", Short: "View audit log", ... }`
  - `cmd/admin.go:172` — `rootCmd.AddCommand(... activityCmd ...)`
  - Plan quote (plan.md table): `_(no cmd/admin_activity.go)_ | — | new file cmd/activity_aggregate.go`
  - Plan quote (plan.md line 99): "No command-name collisions."
  - Plan quote (phase-05): "Codex prompt expects `goclaw activity aggregate` as top-level → create new `activityCmd` parent."
- **Suggested fix:** Add `aggregate` as subcommand of the **existing** `activityCmd` in `cmd/admin.go` (or split that var). Update plan.md inventory row to reflect the existing parent. Rerun the "naming collision" sweep step in phase 1 properly (it failed silently because the grep in phase-01 step 5 only looks for `Use:.*"follow|reconnect|branch|aggregate|test"`, not `Use: "activity"`).

## Finding 2: Path prefix mismatch — new "sessions" subcommands hit a different backend tree than existing siblings
- **Severity:** High
- **Location:** plan.md "Command Surface Inventory" + phase-03 "Surfaces"
- **Flaw:** Existing `goclaw sessions <verb>` commands all call `/v1/sessions/...`. Plan adds `goclaw sessions branch` / `goclaw sessions follow` that target `/v1/chat/sessions/{key}/branch` and `/v1/chat/sessions/{key}/history/follow`. The two surfaces live under sibling-but-distinct backend resource trees. Plan never reconciles this — operators will reasonably assume `goclaw sessions X` shares the `/v1/sessions/` namespace.
- **Failure scenario:** A privileged operator runs `goclaw sessions delete <key>` expecting to delete the same record they just branched with `goclaw sessions branch <key>`. The two commands talk to different backend stores. Wrong record deleted, branched session orphaned. Also makes the path-escape audit harder — two distinct prefixes share one Cobra group.
- **Evidence:**
  - `cmd/sessions.go:67,88,109,128` — all use `/v1/sessions/...` with `url.PathEscape(args[0])`.
  - Plan quote (phase-03): "Endpoint: `POST /v1/chat/sessions/{key}/branch`" and "`GET /v1/chat/sessions/{key}/history/follow`".
  - `cmd/chat_sessions.go:5` — there is already a separate `chatSessionsCmd` (`Use: "sessions"`) under `chatCmd`, which is where `/v1/chat/sessions/...` semantics naturally belong.
- **Suggested fix:** Attach `branch` and `follow` to `chatSessionsCmd` so the command path is `goclaw chat sessions branch` (matching `/v1/chat/sessions/...`), or explicitly document in plan.md why the same Cobra parent maps to two different backend trees and add a help-text warning. Phase 1 scope-lock must record the path-prefix split.

## Finding 3: Sibling files used as path-escape reference are themselves unescaped — RT-02 regression vector is wider than plan admits
- **Severity:** High
- **Location:** plan.md "Decisions" (path-escape claim) + phase-04 file modification list
- **Flaw:** Plan claims "Path-escape all path params via the same pattern already used in `cmd/api_keys_rotate.go`, `cmd/storage.go` (after P5 RT-02 fix)." but the files that phase 4 will modify (`cmd/channels_writers.go`) and the analogous neighbor (`cmd/channels_instances.go`) do NOT path-escape `args[0]`. The new `writers test` subcommand sits next to four unescaped siblings. Reviewers may copy-paste the wrong pattern. Also `cmd/mcp_servers.go:23` POSTs to `/v1/mcp/servers/{id}/reconnect` unescaped — exact same shape as the new `providers reconnect` plan in phase 2 ("reference existing pattern" risks copying the wrong one).
- **Failure scenario:** `goclaw channels writers test 'inst/../../admin' --group-id g --user-id u` collapses path traversal segments before the request leaves the client (URL parsing in net/http or the gateway may interpret them); attacker probing one instance triggers writer-test on an unrelated resource. Even if backend rejects, the CLI's enumeration footprint widens.
- **Evidence:**
  - `cmd/channels_writers.go:19` — `c.Get("/v1/channels/instances/" + args[0] + "/writers")` (no escape)
  - `cmd/channels_writers.go:37,55,73` — same pattern, no escape
  - `cmd/channels_instances.go:54,99,118` — `args[0]` unescaped in path
  - `cmd/mcp_servers.go:23` — `c.Post("/v1/mcp/servers/"+args[0]+"/reconnect", nil)` (no escape)
  - `cmd/agents_instances.go:114,140` — `fmt.Sprintf("/v1/agents/%s/instances/%s/metadata", args[0], user)` (no escape)
  - Plan quote (plan.md): "Path-escape all path params via the same pattern already used in `cmd/api_keys_rotate.go`, `cmd/storage.go`"
- **Suggested fix:** Phase 1 scope-lock must enumerate every existing unescaped path concatenation (use `grep -rn '"/v1/.*"+args' cmd/`) and either (a) fix them in scope or (b) explicitly defer with a tracking issue. Otherwise the "P5 RT-02 lesson" claim in the plan is rhetoric, not protection. Add a new validation gate: lint rule or test that fails when `cmd/*.go` introduces `args[0]` directly into a path literal.

## Finding 4: Logs runtime aggregate output is not redacted — sensitive payload leakage
- **Severity:** High
- **Location:** phase-05 "5.2 `goclaw logs aggregate`"
- **Flaw:** The runtime ring buffer surfaces server log content (level, source, last_seen, possibly message keys). Plan describes the JSON response shape and prints it through `printer.Print(unmarshalMap(...))` with no redaction. No existing output sanitizer exists for the new commands. CLI prints whatever the server returned verbatim to TTY, including possibly tokens/secrets logged at warn/error level on the server.
- **Failure scenario:** Operator running `goclaw logs aggregate --level warn` over SSH (output gets captured in shell history / scrollback / CI logs) inadvertently exfiltrates DB connection strings, API keys, or PII that the server logged in a warn message. Compare with `cmd/backup_s3.go:22,36` which explicitly masks secrets defense-in-depth — the new logs aggregate command has no equivalent.
- **Evidence:**
  - `cmd/backup_s3.go:22` — `Short: "Get current S3 backup configuration (secret_key masked by default)"`
  - `cmd/backup_test.go:155` — `"secret_key should be masked in output"`
  - `grep -rn "RedactSecret|redact|mask" internal/output/` returns no hits — no shared redaction helper.
  - Plan quote (phase-05): "Source = runtime ring buffer, not durable audit log." — acknowledges sensitivity but adds no mitigation.
- **Suggested fix:** Add (a) explicit `--quiet`/redaction default for known sensitive keys (`api_key`, `token`, `secret`, `password`), (b) a one-line warning banner in TTY mode reminding the operator this stream may contain secrets, and (c) phase-05 success criteria for "no raw secret-shaped strings in default table output". Reference `cmd/backup_s3.go` masking as the model.

## Finding 5: Backend-supplied strings printed verbatim — ANSI/escape injection from compromised server
- **Severity:** Medium
- **Location:** All implementation phases (2–5), output sections
- **Flaw:** Every new command pipes server-returned `reason`, `label`, log messages, bucket keys, etc. into table output without any control-character stripping. The CLI explicitly auto-detects TTY (`output.IsTTY`) and switches to human-readable table mode in TTY contexts — exactly the contexts where ANSI escape injection (e.g. cursor manipulation, fake prompts, clipboard hijacks via OSC 52) is effective.
- **Failure scenario:** Backend (or attacker who tampered a single field) returns `"reason":"writer]52;c;cGF5bG9hZA=="` in `channels writers test`. Operator sees the table; OSC 52 silently writes attacker-chosen content to the operator's clipboard. Or a fake "[y/N]" prompt is drawn to social-engineer an interactive confirmation. Plan threat-models server contracts but not server-as-adversary.
- **Evidence:**
  - `grep -rn "sanitize\|strip\|escape" internal/output/` returns no hits.
  - All new commands route through `printer.Print(unmarshalMap(...))` per plan's own pattern (plan.md line 42).
  - `cmd/root.go:59`, `cmd/logs.go:36,72`, `cmd/heartbeat.go:184`, `cmd/teams_events.go:50` — many TTY-aware sites; no escape stripping.
- **Suggested fix:** Add a shared `internal/output/sanitize.go` that strips non-printable control bytes (except `\t`, `\n`) from any string field before table rendering. Phase 6 red-team sweep must add an explicit check: "no command path renders raw server strings without sanitization in TTY mode." Cite OWASP "Log injection" / "Terminal escape injection".

## Finding 6: Misleading guidance directing users to a non-existent command (`goclaw providers verify`)
- **Severity:** Medium
- **Location:** phase-02 "2.2 `goclaw providers reconnect`"
- **Flaw:** Plan says: "Do NOT add `--verify` flag; users call `goclaw providers verify <id>` separately if needed." But the existing subcommand is `goclaw providers verify-embedding`, not `goclaw providers verify`. Users following the docs will get "unknown command" errors and may then suspect the reconnect itself didn't work, retry it (it's a state-changing admin op), and double-toggle a production provider.
- **Failure scenario:** Operator reconnects then runs the (non-existent) `verify`. Confused, they re-run `reconnect` (because docs implied verify is a follow-up). The second reconnect may flip cache_invalidated/registry_updated state again and cause an in-flight request to fail. Plan also doesn't document that `reconnect` is admin-only on the backend; CLI gives no pre-flight warning.
- **Evidence:**
  - `cmd/providers_verify.go:11` — `Use: "verify-embedding <id>"`
  - Plan quote (phase-02): "users call `goclaw providers verify <id>` separately if needed"
  - Plan quote (phase-02): "Admin-only on backend; client sends no body." — admin-only stated but no `--quiet` banner / pre-warn proposed.
- **Suggested fix:** Update plan and command Long-help to reference `goclaw providers verify-embedding` (or whatever the actual sibling is). Add a TTY banner: "Reconnect requires admin token; failure with `UNAUTHORIZED` exit code 2 is expected for non-admin tokens." Don't make users guess.

## Finding 7: `--up-to-index` bounds validation specifies `>= 0` only — no upper bound, no overflow guard
- **Severity:** Medium
- **Location:** phase-03 "3.1 `goclaw sessions branch`"
- **Flaw:** Plan: "Required: `<session-key>` positional, `--up-to-index` (int, >= 0)." No upper bound, no defense against `--up-to-index=9223372036854775807` (or huge ints causing server-side allocation). Combined with `--metadata` repeated flags and `--new-session-key` accepting arbitrary strings, the body shape allows attacker-shaped requests that pass client validation trivially.
- **Failure scenario:** Operator script accidentally passes `--up-to-index $(expr 2 ** 62)` from a buggy upstream. CLI forwards. Server allocates / iterates. Or, in a chained scripted abuse: a malicious actor with valid CLI credentials but limited intent uses the lack of client-side bound checks to amplify load (each branch op = `copied_messages` proportional to index). The "validation-before-HTTP" claim in plan (multiple phases) becomes shallow theater.
- **Evidence:**
  - Plan quote (phase-03): "Required: `<session-key>` positional, `--up-to-index` (int, >= 0)."
  - Plan quote (phase-03 tests): "Negative `--up-to-index` returns validation error before HTTP call." — only negatives, not absurd positives.
  - Phase-02 also: "Server default `limit=50`, max `200`. Don't enforce max client-side; let server respond." — plan takes a position to defer to server, but doesn't apply it consistently with branch / cursor limits.
- **Suggested fix:** Either (a) consistently let server validate all integer bounds (and document in plan) or (b) add reasonable client-side ceilings (e.g. `--up-to-index <= 1_000_000`, `--limit <= 200`, `--cursor <= 10_000_000`). Pick one and apply across phases 2–5. Add tests for huge / overflowing values.

## Finding 8: `--metadata key=value` parser allows clobber + empty keys; no key validation
- **Severity:** Medium
- **Location:** phase-03 tests + phase-03 surface description
- **Flaw:** Plan reuses the existing `parseToolInvokeParams` pattern (`strings.SplitN(pair, "=", 2)` — see `cmd/tools_invoke_args.go:39`). That parser silently allows `--metadata =value` (empty key, becomes object key `""`), `--metadata foo=` (empty string value — fine), and `--metadata foo=bar --metadata foo=baz` (silent last-wins clobber, no warning). Plan only specifies rejection of "no `=`" (`--metadata foobar`), missing the rest.
- **Failure scenario:** Operator pipeline produces `--metadata source= --metadata source=cli` — they expect "cli" to win, but if order is reversed by environment, the server sees `{"source":""}` and a downstream consumer keyed on `metadata.source != ""` mis-routes the branch. Empty-key `""` may produce server-side errors that the CLI surfaces as opaque 400s. No way to forward `=` literal in values that contain `=` (works for SplitN("=",2), so OK — but key cannot contain `=`, undocumented).
- **Evidence:**
  - `cmd/tools_invoke_args.go:38-43` — current parser, no validation.
  - Plan quote (phase-03 tests): "Malformed `--metadata foobar` (no `=`) rejected before HTTP call." — only one negative case.
- **Suggested fix:** Plan must specify: reject empty-key, warn on duplicate-key (or fail), allow values to contain `=`. Add tests for each. Document key-character constraint in `--metadata` flag help.

## Finding 9: Phase 1 scope-lock grep is too narrow to detect the real collisions
- **Severity:** Medium
- **Location:** phase-01 step 5
- **Flaw:** Phase 1's collision sweep command is: `grep -in 'Use:.*"follow\|reconnect\|branch\|aggregate\|test"' cmd/*.go`. This:
  1. Misses parent-command collisions (e.g. `Use: "activity"` — Finding 1).
  2. Misses positional/word-boundary matches (`Use: "test-connection"`, `Use: "test"` standalone won't match the quoted form).
  3. Misses Cobra `AddCommand(...)` registrations that determine the actual collision surface.
- **Failure scenario:** Phase 1 returns "zero collisions" (rubber-stamped), phase 5 fails to compile. The plan's risk register lists "Naming collision with existing subcommands | Low" — mitigation circular because the sweep itself was insufficient.
- **Evidence:**
  - Plan quote (phase-01 step 5): `grep -in 'Use:.*"follow\|reconnect\|branch\|aggregate\|test"' cmd/*.go`
  - Plan quote (plan.md risk table): "Naming collision with existing subcommands | Low | Verified inventory above; no collisions found"
  - Actual collision: `cmd/admin.go:133` `activityCmd` — would not have been found by the planned grep because plan only checks `Use:` strings matching the verb list, not parent group names.
- **Suggested fix:** Replace the grep with: enumerate every `cobra.Command{ Use: "..." }` in `cmd/`, then check that no proposed `Use:` token (`activity`, `aggregate`, `branch`, `follow`, `reconnect`, `test`) conflicts at the level it will be registered. Better: actually attempt to register the new commands in a quick prototype and let Cobra detect duplicates.

## Finding 10: `cmd/sessions.go` already at 171 LOC — adding 2 commands will breach 200-line repo rule, plan's "split if" hedge is too soft
- **Severity:** Medium
- **Location:** phase-03 "Files" + repo CLAUDE.md modularization rule
- **Flaw:** Repo rule (CLAUDE.md): "If a code file exceeds 200 lines of code, consider modularizing". `cmd/sessions.go` is already 171 lines. Phase 3 adds two new commands (each ~30–50 lines including flags + RunE + register), plus validation helpers — will land ~280 lines. Plan says: "If `cmd/sessions.go` grows past 200 lines, split into `cmd/sessions_branch.go` and `cmd/sessions_follow.go`" — but only as conditional, not a default. Same risk noted in phase 4 for `cmd/channels_writers.go` (89 LOC, less acute).
- **Failure scenario:** Implementer follows the path of least resistance, appends to `cmd/sessions.go`, lands 280 LOC. Code-reviewer agent flags it in PR. Reviewer pushes back, files get split mid-PR, conflicts the `feat:` commit semantic-release expects (Finding implicit: rework risks slipping the `feat:` prefix on a renamed commit).
- **Evidence:**
  - `wc -l cmd/sessions.go` → 171
  - CLAUDE.md (project): "Keep individual code files under 200 lines for optimal context management"
  - Plan quote (phase-03): "If `cmd/sessions.go` grows past 200 lines, split into `cmd/sessions_branch.go` and `cmd/sessions_follow.go` per repo modularization rule."
- **Suggested fix:** Make split the default in plan: new files `cmd/sessions_branch.go` and `cmd/sessions_follow.go` from the start. Same for `cmd/channels_writers.go` → `cmd/channels_writers_probe.go` (also avoids the awkward `_test_test.go` file name from Finding-adjacent issue in phase-04). Update phase 3 and phase 4 "Files" sections to remove the conditional.

## Summary of impact

| # | Finding | Severity | Phase affected |
|---|---------|----------|----------------|
| 1 | activityCmd already exists | Critical | 1, 5 |
| 2 | Sessions path-prefix mismatch | High | 1, 3 |
| 3 | Path-escape claim is selective | High | 1, 2, 4 |
| 4 | Logs runtime not redacted | High | 5 |
| 5 | ANSI escape injection from server | Medium | 2, 3, 4, 5 |
| 6 | `providers verify` doesn't exist | Medium | 2 |
| 7 | Integer bounds inconsistent | Medium | 2, 3 |
| 8 | metadata parser too permissive | Medium | 3 |
| 9 | phase-01 collision sweep too narrow | Medium | 1 |
| 10 | sessions.go LOC budget | Medium | 3 |

## Unresolved questions

1. Backend authoritative path: is `/v1/chat/sessions/{key}/...` correct, or has PR #44 actually exposed these under `/v1/sessions/...` to align with existing CLI? Phase 1 must verify against handler source, not inferred from plan.
2. Does the existing `activityCmd` use `/v1/activity` (audit log) vs the new `/v1/activity/aggregate`? If same root, can the new command be a clean subcommand under existing parent, or does scope-of-data differ enough to warrant a rename?
3. Is there an existing redaction helper anywhere in the codebase (beyond `backup_s3.go` inline masking) we should reuse, or does this plan need to ship one?
