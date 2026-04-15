# Phase 2 Implementation Report — Migration (Backup/Restore + Export/Import)

**Date:** 2026-04-14
**Plan:** `plans/260414-2340-ai-first-cli-expansion/phase-02-migration.md`
**Status:** DONE

---

## Files Created

| File | LOC | Purpose |
|------|-----|---------|
| `internal/client/signed_download.go` | 65 | `DownloadSigned()` — GET without auth header (signed token flow) |
| `internal/client/multipart_upload.go` | 50 | `UploadFile()` streaming pipe-based multipart POST, `DrainResponse()` |
| `cmd/backup.go` | 210 | `backup system/tenant` — create, preflight, download |
| `cmd/backup_s3.go` | 155 | `backup s3 config get/set, list, upload, backup` |
| `cmd/restore.go` | 115 | `restore system/tenant` with typed confirmation safety |
| `cmd/agents_export.go` | 100 | `agents export/import/import-merge` |
| `cmd/teams_export.go` | 75 | `teams export/import` |
| `cmd/skills_export.go` | 75 | `skills export/import` |
| `cmd/mcp_export.go` | 75 | `mcp export/import` |
| `cmd/io_helpers.go` | 45 | Shared: `copyProgress`, `writeToFile`, `printProgress` |

## Test Files Created

| File | Tests | Focus |
|------|-------|-------|
| `internal/client/signed_download_test.go` | 3 | No auth header, 4xx error, progress callback |
| `cmd/backup_test.go` | 5 | Preflight, create, download-to-file, missing flag, S3 secret masking |
| `cmd/restore_test.go` | 5 | No-yes refusal, wrong confirm, correct confirm, missing tenant-id |
| `cmd/agents_export_test.go` | 4 | Preview default, --apply endpoint, file write, merge preview |

## Docs Updated

- `README.md` — added Backup & Restore, Export/Import section with CAUTION block, `backup`/`restore` command rows
- `docs/codebase-summary.md` — added all new files to inventory, migration system section, updated command hierarchy

---

## Tasks Completed

- [x] 1.1: `internal/client/signed_download.go`
- [x] 1.2: Signed download unit test
- [x] 2.1: `cmd/backup.go` system + system-preflight
- [x] 2.2: `cmd/backup.go` system-download + `--wait` flag
- [x] 3.1: tenant backup subcommands
- [x] 4.1: `cmd/backup_s3.go` config get/set
- [x] 4.2: S3 list/upload/backup
- [x] 5.1: `cmd/restore.go` system with typed confirmation
- [x] 5.2: `cmd/restore.go` tenant with typed confirmation
- [x] 5.3: Destructive warning in help text
- [x] 6.1: `cmd/agents_export.go` export single
- [x] 6.2: `cmd/agents_export.go` import with preview default
- [x] 6.3: import-merge subcommand
- [x] 7.1: teams export/import
- [x] 7.2: skills export/import
- [x] 7.3: mcp export/import
- [x] 8.1: backup tests
- [x] 8.2: restore tests (confirmation enforcement)
- [x] 8.3: export/import tests (preview vs apply)
- [x] 9.1: README restore CAUTION section
- [x] 9.2: docs/codebase-summary.md migration section

---

## Safety Implementation Notes

- **Restore no-`--yes`:** returns error containing "DESTRUCTIVE" before any network call
- **Restore wrong confirm:** returns error containing "mismatch" before any network call
- **Flag check order:** `--file` validation in download commands happens BEFORE `newHTTP()` to fail fast
- **S3 secret masking:** `secret_key` field replaced with `***` in `backup s3 config get` output unless `--show-secret` passed; masking happens CLI-side regardless of what server returns
- **Import preview default:** `agents/teams/skills/mcp import` hit `*/import/preview` endpoint by default; `--apply` switches to `*/import`
- **Streaming:** `io.Copy` used throughout — no full-file RAM buffering

## Known Deviations from Spec

1. **`-o` shorthand removed** from all export/download `--file` flags. Root command owns `-o` as shorthand for `--output`; cobra panics on shorthand conflict. Using `--file` (long form only) instead.
2. **`backup system --wait`** auto-download uses `c.GetRaw` with auth (not `DownloadSigned`). Server's backup download endpoint at `/v1/system/backup/download/{token}` uses standard auth, not unauthenticated signed URL. `DownloadSigned` is exposed for use when server returns a fully signed external URL (e.g. S3 presigned URL in future).
3. **`agents export --all`** not implemented — spec says "iterate or use teams export — check server semantics." Server has no `/v1/agents/export-all` endpoint; omitted per YAGNI.

---

## Test Results

- `go build ./...` — PASS
- `go vet ./...` — PASS
- `go test ./...` — PASS (all packages)
- `internal/client` coverage: **66.7%** (≥60% requirement met)
- `signed_download.go` coverage: **77.8%**

## Unresolved Questions

1. Signed token TTL: server tokens expire in ~15 min. CLI has no retry on expired token — user must re-run `backup system` to get a fresh token.
2. `agents export --all`: omitted (no server endpoint). If needed, could loop over `agents list` and call export for each ID serially.
3. Import conflict rendering: server preview response returned as-is (raw JSON). No pretty-print of conflict details yet.
