# Phase 2 Code Review — Migration (Backup/Restore + Export/Import)

**Date:** 2026-04-15
**Reviewer:** code-reviewer
**Scope:** Phase 2 impl (`internal/client/{signed_download,multipart_upload}.go`, `cmd/{backup,backup_s3,restore,agents_export,teams_export,skills_export,mcp_export,io_helpers}.go` + tests)
**Impl report:** `plans/reports/fullstack-260415-0040-phase-02-impl.md`
**Phase spec:** `plans/260414-2340-ai-first-cli-expansion/phase-02-migration.md`

---

## Overall Assessment

Solid phase. Streaming correctness is good (pipe-based multipart, io.Copy throughout), destructive-op safety is correctly two-gated (--yes + --confirm mismatch checks happen before any network call), preview-default for imports is implemented consistently across all 4 domains, and `DownloadSigned` correctly omits the Authorization header. `tui.Confirm` systemic P0 fix from Phase 0 is inherited — restore paths do not rely on `tui.Confirm` so they are unaffected. Tests cover the main safety properties (no-yes refusal, wrong-confirm refusal, preview-by-default, no-auth-header).

**One critical security bug:** S3 secret masking targets the wrong field name — will leak `secret_access_key` in production output. One other high-severity: query params for tenant IDs are not URL-escaped. Everything else is medium or minor.

---

## Critical Issues

### C1. S3 secret masking targets wrong field — `secret_access_key` leaks in production

`cmd/backup_s3.go:37-39`

```go
if _, ok := m["secret_key"]; ok {
    m["secret_key"] = "***"
}
```

But the `set` command (`cmd/backup_s3.go:75`) writes the secret to field **`secret_access_key`** (matching AWS SDK naming):

```go
"secret_access_key", secretKey,
```

If the server echoes the S3 config back under the same field name it accepts (`secret_access_key`, which is the conventional AWS field), the mask never matches → `goclaw backup s3 config get` prints the secret in cleartext.

**Why the test misses it:** `backup_test.go:123` synthesizes a fake server response using `secret_key` (not `secret_access_key`), so the test passes despite the bug. The test is validating the mask implementation against itself, not against the real server contract.

**Impact:** P0 — credentials leak in shell history, logs, CI artifacts. Exactly the risk the feature was meant to prevent.

**Fix:** Mask both possible field names (defensive) and update the test fixture to match the real server payload. Consider also masking `access_key_id` suffix or last-4 convention.

```go
if !showSecret {
    for _, k := range []string{"secret_key", "secret_access_key", "aws_secret_access_key"} {
        if _, ok := m[k]; ok {
            m[k] = "***"
        }
    }
}
```

And update `backup_test.go` to include both field names or match what server actually returns (check `goclaw/internal/http/backup_s3_handler.go` payload shape).

---

## High Priority

### H1. Tenant ID not URL-escaped in query string — breaks on special chars, potential injection

Two spots use raw string concatenation into URL query params:

- `cmd/restore.go:101`: `fmt.Sprintf("/v1/tenant/restore?tenant_id=%s", tenantID)`
- `cmd/backup.go:132`: `path += "?tenant_id=" + tenantID`

If `tenantID` contains `&`, `#`, `?`, `=`, whitespace, or binary junk, the URL is malformed or injects additional query params. Tenant IDs in GoClaw are usually UUIDs so this is unlikely to break in normal usage, but:
1. No validation of tenant-id format in CLI — trusting caller input.
2. Every other path in cmd/ uses `url.PathEscape` for path segments (storage.go, memory.go, api_keys.go) — this is inconsistent.
3. If tenant IDs ever adopt slugs or human-readable names, this silently breaks.

**Fix:** use `url.QueryEscape(tenantID)` in both spots, or build with `url.Values{}` and `.Encode()`.

### H2. Upload response body discarded — server error details lost on failed restore/import

`cmd/restore.go:59-61`:
```go
if resp.StatusCode >= 400 {
    return fmt.Errorf("restore request failed [%d]", resp.StatusCode)
}
```

Same pattern in `agents_export.go:85`, `teams_export.go:82`, `skills_export.go:81`, `mcp_export.go:81`.

For a restore failure (corrupt archive, schema mismatch, disk full), the user sees only `restore request failed [500]` with no actionable detail. Server typically returns `{"ok":false,"error":{"message":"..."}}` which is exactly what we need.

`client.DrainResponse()` in `multipart_upload.go:47` already does the right thing (reads body into error message on 4xx/5xx) — **but no caller uses it.** Dead code.

**Fix:** replace the inline `if resp.StatusCode >= 400` blocks with `client.DrainResponse(resp)` in all 6 call sites (and drop the duplicate manual drain). That is why it was added.

### H3. `writeToFile` does not create parent directories

`cmd/io_helpers.go:31-41`: uses `os.Create(path)` directly. If user runs `goclaw agents export abc --file ./out/agents/abc.tar.gz` and `./out/agents` does not exist, `os.Create` fails with `The system cannot find the path specified.`

`downloadBackup` in `cmd/backup.go:162` DOES call `os.MkdirAll(filepath.Dir(outFile), 0o755)` — but the export commands use `writeToFile` which doesn't. Inconsistent behavior between `backup system-download -o nested/x.tgz` (works) and `agents export abc --file nested/x.tgz` (fails).

**Fix:** add `os.MkdirAll(filepath.Dir(path), 0o755)` at the top of `writeToFile`, then delete the duplicate call from `downloadBackup`.

---

## Medium Priority

### M1. `copyProgress` does not actually report progress — misleading name

`cmd/io_helpers.go:12-18` is literally just `io.Copy` with error wrapping. The function name suggests it streams progress updates (the phase spec required "Progress output cho long-running ops (use stderr…)") but it does not. Compare with `client.copyWithProgress` in signed_download.go which actually takes a progress callback.

Either implement per-chunk stderr progress (e.g. every MB) or rename to `copyBody`. Current name will mislead future maintainers.

### M2. `openFileForUpload` is dead code

`cmd/io_helpers.go:22-28` defines `openFileForUpload` but no caller. `client.UploadFile` opens the file internally. Delete.

### M3. `backup system --wait` auto-download deviation from spec is undocumented in help text

Impl report deviation #2 notes the `/v1/system/backup/download/{token}` endpoint uses standard auth (not signed URL flow). That is fine, but the Long help text for `backupSystemCmd` (cmd/backup.go:22-26) says "server returns a signed download token" — which is misleading. It's a token used by an authenticated endpoint, not a signed URL. Either:

- Update Long text to clarify ("token for authenticated download endpoint"), OR
- Keep the signed-URL flow in mind for future use and wire up `DownloadSigned` — currently exposed but unused in phase 2 (only used by tests).

### M4. `restore system` missing-flag error UX inconsistent with its own message

`restoreSystemCmd.MarkFlagRequired("confirm")` (cmd/restore.go:122) makes cobra validate `--confirm` BEFORE `RunE` executes. So:

- `goclaw restore system foo.tgz` → cobra error: `required flag(s) "confirm" not set`
- `goclaw restore system foo.tgz --confirm=foo.tgz` → our nicer error: `restore system is DESTRUCTIVE — add --yes…`

The second path is what tests cover. The first is the more likely user mistake and produces a less helpful error. Consider dropping `MarkFlagRequired` on `--confirm` and handling missing confirm in the unified "DESTRUCTIVE" error message. Same pattern for `restore tenant`.

### M5. `DrainResponse` unused, tests don't exercise error path from server 5xx on upload

`internal/client/multipart_upload.go:47-55` has `DrainResponse` but no caller. Restore tests (`restore_test.go:39-63`) only exercise happy path (server returns 200). No test verifies that a 500 response surfaces to user. Add:

```go
func TestRestoreSystem_ServerError_Propagates(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(500)
        _, _ = w.Write([]byte(`{"ok":false,"error":{"message":"disk full"}}`))
    }))
    ...
    err := runRestoreArgs(t, "system", archivePath, "--yes", "--confirm="+baseName)
    if err == nil || !strings.Contains(err.Error(), "disk full") {
        t.Errorf("expected server error to bubble, got: %v", err)
    }
}
```

This also validates the H2 fix.

### M6. `backup_test.go` stdout capture leaks on error paths

`backup_test.go:135-150` does manual stdout redirection without `defer` restoring. If `runBackupArgs` panics, subsequent tests see a closed stdout. Prefer:

```go
old := os.Stdout
defer func() { os.Stdout = old }()
```

Also the 4096-byte fixed read buffer would silently truncate larger outputs. Non-issue for this test but brittle.

### M7. `cmd/backup.go` at 209 LoC (guideline is 200)

Spec predicted 150, delivered 209. Only 9 lines over, not urgent. If split, the natural line is system vs tenant vs download-helper → `cmd/backup_tenant.go`. Deferrable.

---

## Low Priority

### L1. `backupSystemDownloadCmd` and `backupTenantDownloadCmd` could use `MarkFlagRequired("file")`

Currently they check `if outFile == ""` in RunE. Using `_ = cmd.MarkFlagRequired("file")` gives cobra-native error earlier and matches the `--confirm` pattern. Or keep consistent with agents/teams/skills/mcp export which silently streams to stdout if no --file — but these are binary archives so stdout streaming to a terminal pollutes it. Current explicit-required for download is defensible.

### L2. `agentsImportMergeCmd` uses `?preview=true` query param; others POST to separate `/preview` endpoint

The preview mechanism is inconsistent:
- `agents import`, `teams import`, `skills import`, `mcp import` → POST to `/v1/.../import/preview`
- `agents import-merge` → POST to `/v1/agents/{id}/import?preview=true`

This reflects server-side API shape (verified in impl report — no separate merge-preview endpoint), not a CLI bug. But it should be called out in docs so users understand why `--apply` flag behavior is identical but wire format differs.

### L3. `DownloadSigned` 10-minute timeout may be insufficient for very large backups

`internal/client/signed_download.go:20` sets `Timeout: 10 * time.Minute`. A multi-GB backup over slow link (50 Mbps) can exceed this. Consider `Timeout: 0` (no timeout) with progress-based liveness, or expose flag `--timeout`. Low priority since `DownloadSigned` is not yet used by any command.

### L4. `copyWithProgress` progress callback may fire with stale `total` on Write error

`signed_download.go:57-62`: progress is called after `total += int64(written)` but before checking `writeErr`. On short write with error, progress fires once with partial count, then function returns error. User sees `progress(X)` then `wrote X bytes: error` — correctly consistent. No bug, just noting the contract.

### L5. S3 config `set` has no confirmation despite being sensitive

`backupS3ConfigSetCmd` overwrites credentials without a confirm prompt (unlike delete/revoke paths). Low risk (user is explicitly passing them) but a `tui.Confirm("Overwrite existing S3 config?", cfg.Yes)` guard would match project patterns. Defer.

---

## Edge Cases Found

1. **Signed-URL endpoint assumption**: `DownloadSigned` expects `/v1/system/backup/download/{token}` to accept no auth. Impl report confirmed the server requires auth, so `DownloadSigned` is currently an unused helper. If server behavior changes, wire up. Meanwhile, `downloadBackup` uses `GetRaw` (authenticated) — correct for current server.

2. **Restore upload cancellation**: if user Ctrl-C mid-upload, the pipe writer goroutine in `client.UploadFile` may leak. No `context.Context` plumbing. Low-frequency edge case but worth a follow-up.

3. **Multipart upload goroutine error swallowed**: `multipart_upload.go:26-37` — if `CreateFormFile` or file I/O fails, `pw.CloseWithError` surfaces to the reader (PostRaw's request body), which then returns the error via `HTTPClient.Do`. Correctly propagated. Good.

4. **backup system with `--wait -o file` but server returns no token**: `cmd/backup.go:50` — `if wait && outFile != "" && token != ""`. If server returns empty token, falls through to `printer.Print(m)` showing the raw response. Silent downgrade; consider warning user that `--wait` was requested but not honored.

5. **agents export streaming to stdout + terminal**: `agents_export.go:47` writes binary .tar.gz to stdout. If user forgot `-o` and ran in terminal, binary spew corrupts terminal state. Consider detecting `term.IsTerminal(stdout)` and refusing, matching patterns elsewhere.

---

## Positive Observations

- **Streaming is correct throughout.** `multipart_upload.go` uses `io.Pipe` to avoid buffering, `io_helpers.writeToFile` uses `io.Copy`, `copyWithProgress` uses a 32KB buffer. No in-memory `ReadAll` of backup archives.
- **Two-gate destructive confirmation** (--yes AND --confirm match) is implemented correctly and before any network I/O. Tests verify both rejection paths.
- **Preview-by-default** across all 4 domains is clear, consistent, and explicitly tested for agents.
- **No auth header on signed download** verified by test inspecting `r.Header.Get("Authorization")`.
- **Destructive warnings in Long help text** for restore are clear and include example invocations.
- `printProgress` correctly goes to stderr, preserving JSON/YAML stdout for piping.
- `buildBody` skips empty flags — clean.
- `newHTTP()` auth check happens consistently before every network call.

---

## Recommended Actions

**Must fix before merge:**

1. **[CRITICAL]** Fix C1: mask `secret_access_key` (and `aws_secret_access_key` defensively) in `backup_s3.go` — verify against actual server response shape by reading `goclaw/internal/http/backup_s3_handler.go`. Update test fixture.

**Should fix before merge:**

2. **[HIGH]** Fix H1: `url.QueryEscape(tenantID)` in `cmd/restore.go:101` and `cmd/backup.go:132`.
3. **[HIGH]** Fix H2: replace inline `if resp.StatusCode >= 400` blocks with `client.DrainResponse(resp)` — surface server error messages. Add test M5.
4. **[HIGH]** Fix H3: `writeToFile` must `os.MkdirAll(filepath.Dir(path), 0o755)` — unify with `downloadBackup`.

**Follow-up (post-merge OK):**

5. M1 Rename `copyProgress` → `copyBody` or implement real progress.
6. M2 Remove dead `openFileForUpload`.
7. M3 Clarify help text for `backup system --wait` (signed URL vs authenticated token endpoint).
8. M4 Drop `MarkFlagRequired("confirm")` on restore; unify error message.
9. M5 Add server-error propagation test.
10. L5 Consider confirm prompt for `backup s3 config set`.

---

## Test Quality

| Check | Result |
|-------|--------|
| Restore no-yes refusal tested | ✅ |
| Restore wrong-confirm refusal tested | ✅ |
| Restore correct-confirm proceeds | ✅ |
| Tenant restore missing tenant-id refusal | ✅ |
| Signed download no-auth-header tested | ✅ |
| Signed download 4xx error tested | ✅ |
| Download writes to file verified | ✅ |
| S3 secret masking tested | ⚠️ tests wrong field name (see C1) |
| Preview default vs --apply tested (agents) | ✅ |
| Preview default tested (teams/skills/mcp) | ❌ (only agents) — minor, pattern is identical |
| Server 5xx error propagation | ❌ missing |
| Multipart upload streaming (large file doesn't OOM) | ❌ no size test (acceptable — implementation is clearly streaming) |
| Restore integration test with real fixture archive | ❌ (spec §8.1 listed testdata/ fixture — deferred) |

---

## Metrics

- New LoC: ~1142 (9 source + 4 test files, excluding existing modifications)
- Build: PASS (`go build ./...`)
- Vet: PASS (`go vet ./...`)
- Tests: PASS (`go test ./...`, no -race)
- New package coverage (per impl report): `internal/client` 66.7%, `signed_download.go` 77.8% — meets ≥60%.
- Files over 200 LoC: `cmd/backup.go` (209) — 9 over guideline. Many sibling cmd files also exceed.

---

## Unresolved Questions

1. What is the actual field name the server uses for the S3 secret in `GET /v1/system/backup/s3/config` response? Need to grep `goclaw/internal/http/backup_s3_handler.go` to confirm masking target. C1 fix should defensively mask both forms regardless.
2. Does server `POST /v1/agents/{id}/import?preview=true` actually honor the query param for merge preview, or is preview only available for non-merge imports? Impl note says spec was ambiguous.
3. Should `backup system --wait` be wired to `DownloadSigned` if/when server exposes a signed-URL variant, or is `GetRaw` with auth the final design? Currently `DownloadSigned` is an orphan.
4. Should export commands refuse to stream binary to a TTY stdout (see edge case 5)?

---

**Status:** DONE_WITH_CONCERNS
**Score:** 7.8/10
**Critical count:** 1

**Summary:** Solid streaming, safety-gating, and preview-default implementation. One critical bug: S3 secret masking targets `secret_key` but `set` writes `secret_access_key` — likely leaks the secret in production output. Two high-priority fixes (URL escaping for tenant IDs, use `DrainResponse` to surface server errors, `writeToFile` missing MkdirAll). Tests cover the main safety invariants but fixture for S3 masking test self-validates the bug instead of catching it.
