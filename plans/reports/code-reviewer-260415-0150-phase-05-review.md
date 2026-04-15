# Phase 5 — Extensions + Remaining Review

**Target:** goclaw-cli Phase 5 (pair, oauth, packages, users, quota, send + many extensions)
**Scope:** 25 new files + 7 modified, ~2973 LoC new code
**Build:** `go build ./...` clean | `go vet ./...` clean | `go test ./...` pass | coverage `cmd` 34.9%

## Overall Assessment

Solid terminal phase. All compile/vet/test green. Modularization successfully trims `admin.go` (236→173), `channels.go` (196→90), `tools.go` (196→77), `skills.go` (stayed 205 — borderline), `mcp.go` (stayed 185). New Phase 5 files **all ≤195 LoC** (hard requirement met). `send` command — marked AI-critical — has 6 validation tests and good JSON-schema docs in `Long`/`Short`. Security discipline is consistent: every new destructive op wraps `tui.Confirm(..., cfg.Yes)` before making the HTTP/WS call.

## Critical Issues

None. No auth-bypass, no data leaks, no injection vectors.

## High Priority

### H1. `channels instances list --type=<v>` — unescaped query value [pre-existing, not Phase 5 regression]
File: `cmd/channels_instances.go:26`
```go
path += "?channel_type=" + v
```
`v` is appended without `url.QueryEscape`. If value contains `&`, `=`, `#`, or space, the URL breaks or smuggles extra params. Pattern already existed in the older `channels.go` before extraction, so Phase 5 only moved it. Still worth fixing; same pattern at `cmd/tools_custom.go:27` (extracted this phase).

**Fix:** Mirror the `q := url.Values{}` pattern already used elsewhere in this file.

### H2. `oauth --provider=<arbitrary>` — no whitelist enforcement
File: `cmd/oauth.go:17-25`
`oauthPath` defaults any provider that is not `"openai"` into the `chatgpt/<provider>/...` path. Users can pass `--provider=../../escape` and the URL will interpolate it. The server will almost certainly reject, but path segment injection is a trust-boundary smell and makes error messages confusing.

**Fix:** Add explicit whitelist at start of each RunE (or validate inside `oauthPath`):
```go
if provider != "chatgpt" && provider != "openai" {
    return fmt.Errorf("--provider must be chatgpt or openai")
}
```
Plus `url.PathEscape(provider)` as defense-in-depth.

### H3. `send.go` — WS connection lifecycle correctness
File: `cmd/send.go:77-90`
Validation happens before `newWS` / `Connect`, and `defer ws.Close()` follows `Connect`. Correct. One subtle concern: `ws.Call("send", ...)` does not appear to enforce a timeout in the client helper. A misbehaving server could hang `goclaw send` indefinitely. Not a new issue — inherited from `internal/client/websocket.go`. **Not blocking**, but worth a follow-up ticket: add `--timeout=30s` flag on WS-based commands and propagate to `Call`.

## Medium Priority

### M1. `skills tenant-config get <id>` missing
Phase spec requires `skills tenant-config get/set/delete`. File `cmd/skills_tenant_config.go` defines only `set` and `delete`. No `get` subcommand. `toolsBuiltinTenantConfigGetCmd` exists for the parallel tools tree, so this is likely an oversight.

**Fix:** Add `skillsTenantConfigGetCmd` that GETs `/v1/skills/{id}/tenant-config`.

### M2. `users search --limit` not forwarded when zero
File: `cmd/users.go:28-30` skips `limit` when `<= 0`. Default flag value is 30, so in practice always forwarded. Fine, but the guard `if limit > 0` silently disables `--limit=0`, which a user might plausibly use to mean "unlimited". Either document "limit=0 = server default" or remove the guard.

### M3. `mcp servers test-connection` — no raw-body passthrough option
File: `cmd/mcp_servers.go:44-64`
The command requires `--config=<JSON>`. For shell-unfriendly JSON (nested quotes, multi-line), there's no `@filepath` support despite `readContent` being available in `helpers.go`. Same gap in:
- `cmd/skills_tenant_config.go` (`set` requires `--config=<JSON>`)
- `cmd/tools_builtin_tenant.go` (`set` requires `--config=<JSON>`)
- `cmd/admin_credentials_users.go` (`set` requires `--body=<JSON>`)
- `cmd/admin_credentials_grants.go` (`create`/`update` require `--body=<JSON>`)
- `cmd/admin_credentials.go` (`update` requires `--body=<JSON>`)

**Fix:** Run `body` values through `readContent()` (`@file` or literal) before `json.Unmarshal`. Consistent with `send --content=@file.txt`. Low effort, high ergonomics win for JSON-heavy admin flows.

### M4. `oauth start` stdout contract may surprise users
File: `cmd/oauth.go:78-87`
When server response contains `auth_id`, it prints the ID to stdout and swallows everything else. When `auth_id` is missing, it falls back to `printer.Print(m)` (full JSON/table). Two different stdout shapes for the same command — breaks automation expecting stable output. Prefer: always print full `m` to stdout, URL to stderr. Or: promote the stdout behavior into doc'd contract with `--format=id-only` explicit flag.

### M5. `channels pending delete` — filters silently no-op'd
File: `cmd/channels_pending.go:66-77`
If neither `--channel` nor `--key` is passed, the DELETE request goes to `/v1/pending-messages` with no filter — plausibly deletes *everything*. `--yes` is gated, but without filters the confirm prompt is opaque: "Delete pending messages?" does not convey scope.

**Fix:** Require at least one of `--channel` or `--key`, OR add `--all` flag to make mass-delete explicit:
```go
if ch == "" && key == "" {
    all, _ := cmd.Flags().GetBool("all")
    if !all {
        return fmt.Errorf("specify --channel, --key, or --all to delete all pending messages")
    }
}
```

### M6. `admin credentials test` — silent destructive potential
File: `cmd/admin_credentials.go:108-119`
The command is called "dry-run" in `Short`, but it POSTs to `/v1/cli-credentials/{id}/test` with no `--yes`. If the server implementation actually invokes the underlying binary with any side effect (network call to auth provider, for instance), user might trigger a real call unknowingly. Not a CLI bug per se — just worth a `Long:` warning mirroring the phase spec's security note: "may execute binary — ensure safe env".

## Low Priority

- **L1.** `cmd/oauth.go:148` — `MarkFlagRequired("provider")` combined with `.String("provider", "chatgpt", ...)` default makes the "required" marker redundant (default value always satisfies required). Harmless but misleading.
- **L2.** `cmd/pair.go:127` — `revokeCmd` uses `--sender-id` + `--channel` but its `Use:` is just `"revoke"` (no visible arg). `Short` says `(--sender-id + --channel required)`; prefer `Use: "revoke --sender-id=<id> --channel=<ch>"` for clarity.
- **L3.** `cmd/packages.go:45-53` — `install` does not use `--yes` even though `uninstall` does. Install affects shared runtime too. Given `packages list` is read-only and install strictly adds, install without confirm is defensible but worth documenting that only uninstall is gated.
- **L4.** `cmd/skills_deps.go:21` — field name is `"dep"` while `skills_misc.go` `install-deps` (plural) uses no body. Server contract check: if server expects `"package"` or `"name"`, the mismatch will silently fail. Worth a quick verification against server `/v1/skills/install-dep` handler.
- **L5.** `cmd/send.go:97` — success message uses `→` (UTF-8 arrow). Windows CP1252 terminals may mangle it. Non-blocking on modern terminals but `->` is safer.
- **L6.** `cmd/channels_contacts.go:47` — `/v1/contacts/merged/` + `args[0]` unescaped path segment. Path-segment injection is largely mitigated because cobra trims the token on whitespace, but `url.PathEscape(args[0])` is still best practice and is used elsewhere in the codebase (`cmd/skills_misc.go:113`).
- **L7.** `cmd/phase5_test.go:282-299` `TestSend_OK` logic is awkward — it asserts *absence* of specific flag errors while expecting WS upgrade to fail. The intent is sound but the test body is hard to read; consider splitting into `TestSend_RejectsConnectionFailure` vs `TestSend_ValidatesFlagsBeforeConnect`.

## Edge Cases Found by Scout

1. **oauth callback** — no retry on 4xx server rejection (e.g., expired code). Current error message just surfaces HTTP status. Users may need to re-run `oauth start` but the CLI gives no hint. Consider detecting `400`/`410` and appending "re-run `oauth start` to get a fresh URL".
2. **pair revoke** — confirmation fires *before* the WS connection attempt. If user types "y" but server is unreachable, they get both "prompted" and "error" with no rollback. Fine for idempotent revoke but creates UX noise; minor.
3. **send** — no de-dup or idempotency key. Server may double-deliver on retry loops. `Long:` text notes "No retry is performed" — good, contract documented. Nothing to fix in CLI.
4. **admin credentials delete** — deletes credential server-side; does CLI also purge local token-cache? Codebase search shows no local cache for credentials (only auth token via keyring). No action needed but worth confirming.
5. **channels pending compact** — triggers LLM tokens; confirmation says "uses LLM tokens" but no cost indication. Admins running this in CI might not see the prompt (TTY gate). Consider requiring `--yes` explicit value (no auto-approve via non-TTY) for this one.

## Positive Observations

1. **Modularization rule respected 100% on Phase 5 files** — none exceed 200 LoC, even though several came close (tools_custom 195, admin_credentials 188).
2. **`send.go` quality matches the AI-critical mandate** — detailed `Long:` with schema, three examples including `@filepath`, three inline usage examples, and clear security note about audit-log behavior. Sets the bar for other AI-oriented commands.
3. **`phase5_test.go` validation-path coverage** — 6 test cases for `send` alone, covering every MarkFlagRequired branch plus file-not-found. `TestMCPServersTestConnection_InvalidJSON` and `TestSkillsTenantConfigSet_InvalidJSON` properly exercise JSON parse failures without needing a server.
4. **`--yes` enforced on all destructive ops** — pair revoke, packages uninstall, channels pending delete/compact, contacts merge, skills tenant-config delete, tools builtin tenant-config delete, admin credentials delete, admin credentials user-credentials delete, admin credentials agent-grants delete, oauth logout. Verified each.
5. **`oauth start` stderr/stdout split is correct** — browser URL to stderr (human-facing), auth_id to stdout (machine-consumable). Automation can pipe `auth_id` cleanly.
6. **Extraction commit left good breadcrumb comments** — `admin.go:172` explains which file assembles the credentials subtree; `skills.go:199-200` notes where misc commands live. Keeps future readers oriented.

## Recommended Actions

Priority order for follow-up:

1. **M1** — Add `skills tenant-config get` to match spec (10 LoC, trivial)
2. **M5** — Add filter enforcement or `--all` flag to `channels pending delete` (safety)
3. **M3** — Add `@file` support to all `--config` / `--body` JSON flags (DX consistency)
4. **H2** — Whitelist `oauth --provider` to `chatgpt|openai` (trust-boundary)
5. **H1** — Fix `channels instances list --type` and `tools custom list --agent` query escaping
6. **M4** — Stabilize `oauth start` stdout contract
7. **M6** — Extend `admin credentials test` docstring with security note
8. **L1-L7** — Batch in a polish PR (none individually blocking)

## Coverage Assessment

- `cmd` package: 34.9% — below typical 60% target, but acceptable given most `RunE` bodies are thin wrappers around HTTP client that are hard to test without a full server harness.
- `send.go`: ~80% estimated for validation paths (5/7 named branches covered; only the Success path and WS-response unmarshal are uncovered — both require WS harness).
- Destructive-op confirmations: 0% direct test coverage. Relies on manual review. Consider mocking `tui.Confirm` via a package-level override var in a future pass.

## Metrics

- **Phase 5 new files:** 25
- **Modified files:** 7
- **LoC delta:** +2973 new / -1649 removed (net +1324)
- **Files >200 LoC in Phase 5:** 0
- **Files >200 LoC elsewhere (pre-existing):** 12 (not a Phase 5 concern)
- **Destructive ops checked:** 10/10 have `--yes` + `tui.Confirm`
- **Tests added:** 24 (all passing)
- **Compile:** clean
- **Vet:** clean

## Unresolved Questions

1. Is `skills tenant-config get` intentionally omitted or forgotten? Spec line 80 says "get/set/delete".
2. Server contract for `POST /v1/skills/install-dep`: does it expect `{"dep": ...}` or `{"package": ...}` / `{"name": ...}`?
3. Does `admin credentials test` have any side effects (auth-provider calls, token refresh)? If yes, add `--yes`.
4. Is `channels pending delete` with no filters intentionally able to wipe all pending messages? If yes, add `--all` guard.
5. `oauth start` stdout contract — is the current "auth_id-only when present, full payload otherwise" intentional for piping?

---

**Status:** DONE_WITH_CONCERNS
**Score:** 8.2/10
**Critical count:** 0 (0 blockers, 3 high, 6 medium, 7 low + 5 unresolved questions)
**Summary:** Phase 5 ships a clean terminal phase with strong modularization discipline and well-documented AI-critical `send` command. No blockers. Concerns are ergonomics (JSON `@file` support, `oauth --provider` whitelist, `channels pending delete` filter enforcement) and one spec miss (`skills tenant-config get`). Overall pragmatic quality, ready to merge after addressing M1 + M5 as minimum bar.
