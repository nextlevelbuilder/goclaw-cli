# Phase 1 Code Review — Admin/Ops Foundation

**Reviewer:** code-reviewer
**Date:** 2026-04-15
**Scope:** cmd/tenants.go, cmd/heartbeat.go, cmd/heartbeat_checklist.go, cmd/system_configs.go, cmd/edition.go, cmd/config_cmd.go (+ test files)
**Phase spec:** plans/260414-2340-ai-first-cli-expansion/phase-01-admin-ops-foundation.md

## Summary
- Build: `go build ./...` clean, `go vet ./...` clean
- Tests: all P1 tests pass (`go test ./cmd/... -run ...` OK, 0.37s)
- LoC: most new files comply; 2 files marginally over limit
- P0 pattern compliance (newHTTP, newWS, FollowStream, APIError) — good
- WS vs HTTP routing — matches server contracts
- **1 critical destructive-op bug** (config permissions revoke unguarded)
- **1 high-priority UX bug** (heartbeat logs --follow no signal handling)

---

## Critical Issues

### C1. `config permissions revoke` missing confirmation guard
**File:** cmd/config_cmd.go:192-226
**Severity:** Critical (destructive op unguarded + help text lies to user)

Phase spec TODO §5.2 required "Wire `--yes` on revoke". The `Long` help text advertises "Requires --yes to confirm", but `RunE` calls `ws.Call("config.permissions.revoke", ...)` directly with no `tui.Confirm` gate. `tui` is not even imported in config_cmd.go.

**Impact:** Admin running `goclaw config permissions revoke --agent=X --user=Y` will revoke immediately without prompt, contradicting documentation. Silently passes in CI without `--yes` — inconsistent with every other destructive command in the codebase.

**Fix:**
```go
import "github.com/nextlevelbuilder/goclaw-cli/internal/tui"

// inside RunE, before ws := newWS:
if !tui.Confirm(fmt.Sprintf("Revoke %s permission for user %s on agent %s?", configType, user, agent), cfg.Yes) {
    return nil
}
```

Also add a test `TestConfigPermissionsRevoke_RequiresConfirm` that runs with `cfg.Yes=false` in interactive-less mode and verifies the call was NOT made (or, preferably, change `tui.Confirm` to gate on `autoYes` only — see Informational I1).

---

## High Priority

### H1. `heartbeat logs --follow` ignores SIGINT/SIGTERM
**File:** cmd/heartbeat.go:175-190
**Severity:** High (UX/resource leak)

Uses `context.Background()` — Ctrl+C will not cancel the stream cleanly. The existing `logs.go:43-49` already demonstrates the correct pattern with `signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)`. User must kill the process; in-flight WS connection leaks until server disconnects.

**Fix:** Mirror `cmd/logs.go`:
```go
if follow {
    ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer cancel()
    return client.FollowStream(ctx, cfg.Server, cfg.Token, "cli", cfg.Insecure,
        "heartbeat.logs", map[string]any{"agentId": agent, "limit": tail},
        handler, nil)
}
```

### H2. `system-configs delete` relies on `tui.Confirm` in non-interactive mode
**File:** cmd/system_configs.go:111-114
**Severity:** High (silent auto-approve in CI)

`tui.Confirm(msg, autoYes)` returns `true` when `autoYes || !IsInteractive()`. In CI/automation mode (non-TTY stdin), delete proceeds **without** `--yes`. Phase spec §3.1 requires "`delete` requires `--yes`". Help text also says "Requires --yes to confirm", but with piped input the gate is bypassed.

This is actually a **project-wide pattern issue** (same behavior in agents, cron, channels, mcp, etc.) inherited from Phase 0. Flagging here because system-configs delete can break server startup (admin keys, feature flags) — higher blast radius than most deletes.

**Options:**
- (preferred) Change `tui.Confirm` to require `autoYes=true` when non-interactive: treat missing TTY as implicit "no" unless `--yes` explicit.
- (local fix) In `systemConfigsDeleteCmd.RunE`, check `!cfg.Yes && !tui.IsInteractive()` and return an error demanding `--yes`.

Since this is systemic, recommend filing a separate Phase-0 follow-up to harden `tui.Confirm` rather than patch each command.

### H3. `tenants update` has no confirmation
**File:** cmd/tenants.go:82-105
**Severity:** High (phase spec says destructive; status=suspended can lock out users)

Phase spec §Non-Functional: "Destructive ops (`tenants update`, …): `--yes` flag + interactive confirm". Code currently calls PATCH unconditionally. Suspending a tenant suspends all its users; should be gated.

**Fix:** add `tui.Confirm(fmt.Sprintf("Update tenant %s?", args[0]), cfg.Yes)` before `c.Patch(...)`, or more targeted: prompt only when `status` flag changed.

---

## Medium Priority

### M1. File size over 200-LoC guideline
- `cmd/config_cmd.go` — 266 LoC
- `cmd/heartbeat.go` — 249 LoC
- `cmd/tenants.go` — 244 LoC

Per CLAUDE.md and development-rules.md: "Keep individual code files under 200 lines for optimal context management". `heartbeat_checklist.go` was correctly split; `config_cmd.go` could split `config_permissions.go` (lines 115-264 are cohesive and would trim it to ~115 LoC).

Not a blocker. Suggest splitting `config_permissions.go` since it's a natural boundary and would mirror the heartbeat_checklist.go pattern.

### M2. Flag state leaks between cobra tests
**Files:** cmd/tenants_test.go, cmd/heartbeat_test.go, cmd/config_cmd_test.go, cmd/system_configs_test.go

Tests call `cmd.Flags().Set(...)` on shared `var *cobra.Command` values. Cobra command flags are package-level singletons, so `Set()` in TestA leaks into TestB. Current tests pass because each test resets the flags it cares about, but tests are order-dependent. Examples:
- `TestTenantsCreate` sets `name=Acme`, `slug=acme` — these persist into `TestTenantsUpdate` which only resets `name`.
- `TestSystemConfigsSet_JSONValue` uses `defer ... Set("json", "false")` — good; but `TestSystemConfigsSet_StringValue` relies on default remaining false.

**Fix:** wrap each test's flag ops in a helper that resets all flags after, or instantiate a fresh `*cobra.Command` per test. Not blocking, but will cause flaky tests under `-shuffle=on`.

### M3. `heartbeat.logs --follow` handler ignores event stream type
**File:** cmd/heartbeat.go:181-187

Handler prints every `event.Payload` as a map. The server emits both initial backlog (inside the Call response) and continuous push events. The initial `Call` response payload is discarded (`_, err := FollowStream` doesn't plumb it); only subscribed events are printed. For `--follow`, this is generally fine, but `--tail=20` is sent as `limit` — behavior depends on whether server includes backlog in pushed events or response only. If response-only, `--tail` is silently dropped in follow mode.

**Suggestion:** document in help text that `--tail` only applies when `--follow=false`, or modify FollowStream to surface initial response to handler.

### M4. `tui.Confirm` bypass in `tenants users remove` is partially defensive
**File:** cmd/tenants.go:184-202

Typed `--confirm=<userID>` gate (required flag + exact match) is a strong guard. However, the subsequent `tui.Confirm` is decorative in non-interactive mode (returns true). The typed-match is sufficient protection, but help text should clarify which is the real gate. Currently implies both; typed-match is the only one that works in CI.

**Fix (docs):** clarify in `Long` that `--confirm=<userID>` is the mandatory safety check; the `[y/N]` prompt is interactive-only additional confirmation.

---

## Low Priority

### L1. `edition.go` raw-JSON parsing path is fragile
**File:** internal/client/http.go:166-187 (not P1, but exposed by edition)

`edition` endpoint returns non-envelope JSON `{"edition":"pro","version":"1.0.0"}`. `HTTPClient.do` unmarshals as envelope → `OK=false`, `Error=nil`, `Payload=nil`, then falls through. Works, but relies on envelope-ish heuristic. Not a P1 issue; existing behavior.

### L2. Inconsistent use of `newHTTP()` vs direct `client.NewHTTPClient`
**File:** cmd/edition.go:21

`edition` correctly bypasses `newHTTP()` (which enforces token). But `cfg.Server == ""` check is duplicated with helper. Consider extracting `newHTTPNoAuth()` helper for consistency with P2+ no-auth endpoints (health, version).

### L3. `heartbeat.targets` passes `agentId: ""`
**File:** cmd/heartbeat.go:143

Comment says "agentId kept for server backward-compat but targets are tenant-scoped". Fine, but empty string may or may not be handled by server as "all agents" vs validation error. Suggest verifying server behavior or conditionally omitting the field (see `buildBody` which skips empty strings).

### L4. `systemConfigsSetCmd` PUT response always parsed as map
**File:** cmd/system_configs.go:94-99

Response is printed via `printer.Print(unmarshalMap(data))`. For `set` operations, a `printer.Success(...)` line would be more useful. Current output dumps the entire config object back — noisy.

---

## Positive Observations

- **Phase 0 patterns applied correctly**: `newHTTP()` for HTTP commands, `newWS("cli")` for WS commands, `FollowStream` for log streaming. `ResolveFormat` handled globally in root.go PersistentPreRunE.
- **Typed confirmation** on `tenants users remove` (line 187-189) is best-in-class — matches the tenant-destructive-ops risk table in spec.
- **`--json` flag on `system-configs set`** correctly validates JSON before sending, returns a useful error wrap.
- **Tests**: HTTP tests assert method+path+body; WS tests use v3 protocol frames correctly. Error-path tests for `TestTenantsUsersRemove_ConfirmMismatch` and `TestSystemConfigsSet_InvalidJSON` are exactly the kind of edge-case coverage P0 wanted.
- **`tenants.mine` → WS** correctly uses WS (per server), while `tenants list/get/create/update` use HTTP.
- **`edition_test.go`** covers both no-server and no-token cases — good boundary testing.
- **`heartbeat_checklist.go` split** proactively addresses LoC limit (Phase spec §2.3).
- **Error bubbling**: all commands `return err` from helpers; central handler in root.go `Execute()` prints via `output.PrintError` + `output.Exit(output.FromError(err))`. No manual `os.Exit(1)` scattered.
- **SilenceErrors/SilenceUsage**: set on rootCmd, prevents cobra's double-print behavior.

---

## Edge Cases Found by Scout

- **`tui.Confirm` in non-interactive mode auto-confirms** — affects system-configs delete, tenants update, any delete without typed-confirm. Systemic, not P1-specific.
- **Cobra flag singleton leakage** in tests — order-dependent, would fail `go test -shuffle=on`.
- **Ctrl+C handling missing** in heartbeat logs --follow; `logs.go` got it right, heartbeat.go copied the shape but missed signal wiring.
- **`heartbeat set --agent=X --interval=0`** silently omits interval (buildBody drops int==0). Server likely rejects missing interval, but error will be server-side rather than client validation.
- **`config permissions revoke` fabricated help text** — docs promise behavior code doesn't implement.

---

## Recommended Actions (priority order)

1. **[CRITICAL]** Add `tui.Confirm` gate to `configPermissionsRevokeCmd`; import `tui` in config_cmd.go; add test `TestConfigPermissionsRevoke_NoConfirm_NoCall`.
2. **[HIGH]** Wire `signal.NotifyContext` into `heartbeat logs --follow` path.
3. **[HIGH]** Add confirmation to `tenants update` (or explicitly drop it from spec's destructive list).
4. **[HIGH/SYSTEMIC]** Harden `tui.Confirm` to refuse non-interactive execution without explicit `--yes`. Add Phase-0 follow-up task.
5. **[MED]** Split `config_permissions.go` out of config_cmd.go to restore <200 LoC.
6. **[MED]** Add test harness that resets cobra flags between tests (or use `t.Cleanup`).
7. **[LOW]** `system-configs set` → use `printer.Success(...)` instead of dumping full object.
8. **[LOW]** Document `--tail` behavior under `--follow`.

---

## Metrics
- Files reviewed: 11 (6 prod + 5 test)
- Prod LoC: 1017
- Test LoC: 703
- Critical issues: 1
- High issues: 3
- Medium issues: 4
- Low issues: 4
- Build: PASS | Vet: PASS | Tests: PASS (all P1)

---

## Unresolved Questions
- Should `tui.Confirm` fix be a P1 blocker or filed as Phase 0 follow-up? (I recommend follow-up; patching each command now is churn.)
- Is `tenants update` actually destructive enough to warrant confirmation, or was spec overly broad? (Defer to PM; status=suspended is the teeth.)
- Does the server side of `config.permissions.revoke` return a useful error if the grant doesn't exist, or silent success? (Affects whether missing confirmation is latent or immediate risk.)
- Does `heartbeat.logs` WS method include initial backlog in the response payload or only push events? (Affects whether `--tail` under `--follow` is functional or silently ignored.)

---

**Status:** DONE_WITH_CONCERNS
**Score:** 7.8/10
**Critical count:** 1
