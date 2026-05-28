---
title: 'Fix issue #17: cannot read trace details by trace id'
description: >-
  Restore reliable single-trace-by-id reads in goclaw-cli with explicit error
  categorization, human-readable rendering, and regression tests. TDD mode.
status: pending
priority: P2
branch: worktree-fix-trace-details-by-id
tags:
  - bugfix
  - traces
  - ai-ergonomics
  - tdd
blockedBy: []
blocks: []
created: '2026-05-28T06:58:02.755Z'
createdBy: 'ck:plan'
source: skill
---

# Fix issue #17: cannot read trace details by trace id

## Overview

`goclaw traces get <traceID>` exists ([cmd/traces.go:61](cmd/traces.go:61)) and calls `GET /v1/traces/{traceID}`, but the issue reporter observed that detail lookup "does not work reliably" with "unusable output". Verified CLI-side defects from scout:

1. **No human-readable render path.** `tracesGetCmd` calls `printer.Print(unmarshalMap(data))`. In table mode, `output.Printer.Print` falls back to JSON dump ([internal/output/output.go:30-37](internal/output/output.go:30)) because the trace payload is a `map[string]any`, not a `*TableData`. Users in a TTY see an unindented blob.
2. **Silent unmarshal failures.** `unmarshalMap` discards `json.Unmarshal` errors ([cmd/helpers.go:48-53](cmd/helpers.go:48)) — a malformed/empty server response renders as empty `{}`.
3. **Path-traversal exposure.** `tracesGetCmd` concatenates `args[0]` into the URL with no `url.PathEscape` and no allowlist; `url.PathEscape("..")` returns literal `..`. Bad ids reach the server.
4. **No regression test.** `cmd/traces_get_test.go` does not exist.

Plus one acceptance gap:
- No client-side categorization between not-found / permission-denied / malformed-id / server-failure. The exit-code-mapping infrastructure already exists ([internal/output/exit.go:18-44](internal/output/exit.go:18)): `MapServerCode` covers `NOT_FOUND`→3, `TENANT_ACCESS_REVOKED`→2, `INVALID_REQUEST`→4, `INTERNAL`/5xx→5. `cmd/helpers.go:152-172` derives codes from raw HTTP status as a fallback. **No `MapServerCode` extension needed.**

Server-vs-CLI root cause is most likely CLI-side per the above, but Phase 1 captures a real response shape against a live gateway and confirms the failure mode before Phase 2 implementation.

**Acceptance criteria** (verbatim from [issue #17](https://github.com/nextlevelbuilder/goclaw-cli/issues/17)):
1. `goclaw-cli` can read trace details by id.
2. Output works in JSON and human-readable modes.
3. Errors clearly distinguish: not found, permission denied, malformed id, and server/API failure.
4. If API is the root cause, document/link the server-side issue in `digitopvn/goclaw`.
5. Add regression test or smoke test for trace detail lookup.

**Out of scope (verbatim from invocation):**
- Redesigning the traces domain.
- WS streaming for trace events.
- Anything beyond the single-trace-by-id read path.

## Locked decisions (from red-team review 2026-05-28)

These supersede earlier drafts and are not re-litigated during implementation:

- **No `MapServerCode` extension.** HTTP-status fallback at `cmd/helpers.go:152-172` already covers 403→ExitAuth, 404→ExitNotFound, 5xx→ExitServerError. Server-code names are NOT speculated in tests.
- **Inline render, no `cmd/traces_render.go`.** Follows the established pattern of `tracesListCmd` ([cmd/traces.go:30-59](cmd/traces.go:30)). Cap render code at ~50 LOC inside `tracesGetCmd.RunE`.
- **Inline `json.Unmarshal` with error check.** No new `unmarshalMapStrict` helper. Pattern: `var trace map[string]any; if err := json.Unmarshal(data, &trace); err != nil { return fmt.Errorf("decode trace payload: %w", err) }`.
- **Strict id validation.** Reject empty/whitespace, `..`, leading/trailing `/`, embedded `/`, `\\`, control chars. Use a small allowlist regex (e.g. `^[A-Za-z0-9._-]+$`) AND `len > 0` AND not equal to `.` or `..`. Then `url.PathEscape` on top.
- **Test harness forces table mode explicitly.** Every table-mode test in `cmd/traces_get_test.go` passes `--output table` via `runCmd(t, "traces", "get", id, "--output", "table")` — `go test` stdout is piped, so default would be JSON.
- **Span tree IS in scope.** The command's existing `Short` description says "Get trace with span tree" ([cmd/traces.go:62](cmd/traces.go:62)) and AC #2 requires human-readable mode. Reuse `output.PrintTree` / `output.TreeNode` ([internal/output/tree.go:7](internal/output/tree.go:7)). No new tree renderer.
- **Events: simple flat list, no truncation.** Print `EVENTS (n=N):` header then one line per event. No first-5-last-1 polish.
- **5xx auto-retry awareness.** `internal/client/http.go` retries 3x on 5xx; Phase-2 5xx tests assert `calls >= 1` (not equal) and verify exit-code-5, not call count.

## Phases

| Phase | Name | Status |
|-------|------|--------|
| 1 | [Repro and root cause](./phase-01-repro-and-root-cause.md) | Completed |
| 2 | [CLI fix: output and error mapping](./phase-02-cli-fix-output-and-error-mapping.md) | Completed |
| 3 | [Docs and ship](./phase-03-docs-and-ship.md) | Pending |

## TDD discipline (per `--tdd`)

Each phase opens with red tests describing the desired behavior, then implementation lands until tests are green. Specifically:
- **Phase 1** — write 3 red tests in `cmd/traces_get_test.go` reproducing the bug against a mock `httptest` server with a realistic envelope; capture the exact payload shape from a live gateway smoke probe (or document gateway-unreachable escape).
- **Phase 2** — extend the test file with cases for human-readable rendering + 4 error categories + path-traversal validation before implementation.
- **Phase 3** — no new tests, only verification + ship.

## Dependencies

None. P6 (HEAD = `2801486`) provides all required prerequisite surfaces.

## Risk and rollback

- **Risk:** root cause is purely server-side. Mitigation: Phase 1 explicitly classifies CLI vs server; if server, Phase 2 narrows to error-mapping + smoke-test + upstream-issue filing, code change to traces command itself is minimal.
- **Risk:** trace payload shape (spans/events/messages structure) differs from speculation. Mitigation: Phase 1 captures the real fixture FIRST; Phase 2 render code matches the captured shape, not a guess.
- **Risk:** existence oracle from distinct 404 / 403 messages. AC #3 explicitly requires the distinction — accepted trade-off, documented in Phase 2 Security Considerations.
- **Rollback:** all changes confined to `cmd/traces.go` and `cmd/traces_get_test.go`. Single revert.
