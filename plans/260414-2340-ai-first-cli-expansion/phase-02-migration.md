---
phase: 2
title: Migration — Backup/Restore + Export/Import
status: pending
priority: high
blockedBy: [phase-00]
---

# Phase 2 — Migration (Backup/Restore + Export/Import)

## Context Links
- Brainstorm §2 G5 + G8, §5 Phase 2
- Server HTTP: `../../../goclaw/internal/http/{backup_handler.go,backup_s3_handler.go,restore_handler.go,tenant_backup_handler.go,tenant_restore_handler.go,agents_export*.go,agents_import*.go,mcp_export.go,mcp_import.go,skills_export.go,skills_import.go}`

## Overview
- **Priority:** High (tier 🔥)
- **Status:** Pending
- **Description:** Backup/restore system + per-tenant + S3, + export/import cho agents/teams/skills/mcp. Critical cho disaster recovery, cloning environments, agent migration. Highest risk phase due to destructive restore ops.

## Key Insights
- Backup download dùng **signed token URL flow**: endpoint trả token, download `/backup/download/{token}` **KHÔNG** auth header
- Restore is the single most dangerous op in CLI — require typed confirmation + `--dry-run`/`--preview` default behavior
- Export/import: server has both "preview" (dry-run) và actual operation — CLI always call preview first trừ khi user pass `--skip-preview`
- Migration is 2 paradigms:
  1. **System-level backup/restore** (group riêng — system-wide + S3)
  2. **Per-domain export/import** (subcommand trong mỗi group: `agents export`, `skills import`, etc.)
- Decision: **per-domain subcommand** (nearer mental model), skip `migrate` group

## Requirements

### Functional
**System Backup:**
- `backup system` — create backup, output signed download URL
- `backup system preflight` — check disk/capacity
- `backup system download <token>` — download signed URL to file
- `backup s3 config get/set` — S3 destination config
- `backup s3 list` — list backups in S3
- `backup s3 upload <file>` — upload local backup to S3
- `backup s3 backup` — create + upload in one shot

**Tenant Backup:**
- `backup tenant [--tenant-id=...]`
- `backup tenant preflight`
- `backup tenant download <token>`

**Restore:**
- `restore system <file>` — require typed confirm
- `restore tenant <file> [--tenant-id=...]`

**Per-domain Export/Import:**
- `agents export <id> [--output-file=...]` / `agents export --all`
- `agents import <file> [--preview]` — preview mặc định true nếu không có `--apply`
- `agents import-merge <id> <file>` — merge vào agent đang có
- `teams export <id>` / `teams import <file>`
- `skills export` (all) / `skills import <file>`
- `mcp export` / `mcp import <file>`

### Non-Functional
- Signed download: no auth header, validate token format
- Streaming download/upload — no buffer full file vào RAM
- Restore typed confirmation: gõ đúng filename hoặc tenant ID
- `--preview` flag là default cho mọi import; user phải `--apply` để actual execute
- Progress output cho long-running ops (use stderr, not stdout to avoid polluting JSON)

## Architecture

### Command tree
```
goclaw backup
├── system [--wait]
├── system-preflight
├── system-download <token> -o <file>
├── tenant [--tenant-id=...]
├── tenant-preflight
├── tenant-download <token> -o <file>
└── s3
    ├── config get
    ├── config set --bucket=... --region=... --access-key=... --secret-key=...
    ├── list
    ├── upload <file>
    └── backup

goclaw restore
├── system <file> --yes --confirm=<filename>
└── tenant <file> --tenant-id=... --yes --confirm=<tenantID>

goclaw agents export <id> [-o file]
goclaw agents export --all -o file
goclaw agents import <file> [--apply]
goclaw agents import-merge <id> <file> [--apply]

goclaw teams export <id> [-o file]
goclaw teams import <file> [--apply]

goclaw skills export [-o file]
goclaw skills import <file> [--apply]

goclaw mcp export [-o file]
goclaw mcp import <file> [--apply]
```

### Signed download flow
```
1. goclaw backup system            → POST /v1/system/backup → { token, expires_at }
2. goclaw backup system-download T → GET /v1/system/backup/download/{T} (no auth, binary stream)
3. CLI writes to -o file with progress
```

## Related Code Files

### Create
- `cmd/backup.go` (~150 LoC)
- `cmd/backup_s3.go` (~100 LoC — split from backup.go)
- `cmd/restore.go` (~120 LoC)
- `internal/client/signed_download.go` (~60 LoC — unauth binary streaming)

### Modify
- `cmd/agents.go` — add export/import subcommands (may push over 500 LoC → split `cmd/agents_export.go`)
- `cmd/teams.go` — add export/import (split `cmd/teams_export.go` if needed)
- `cmd/skills.go` — add export/import (split `cmd/skills_export.go`)
- `cmd/mcp.go` — add export/import (split `cmd/mcp_export.go`)
- `cmd/root.go` — register backup, restore groups
- `internal/client/http.go` — add `GetStreamNoAuth(url)` method for signed download
- `README.md`, `docs/codebase-summary.md`

## Implementation Steps

### Step 1: Signed download helper
1. Create `internal/client/signed_download.go` with `DownloadSigned(url, writer)` — no auth header, progress callback
2. Test với httptest signed token endpoint

### Step 2: System backup
1. Create `cmd/backup.go` with `backupCmd` root
2. Implement `system`, `system-preflight`, `system-download`
3. `system` trả token; flag `--wait` để block + auto-download

### Step 3: Tenant backup
1. Add `tenant`, `tenant-preflight`, `tenant-download` vào `backup.go`

### Step 4: S3 backup
1. Create `cmd/backup_s3.go` with `backupS3Cmd`
2. Implement config get/set, list, upload, backup

### Step 5: Restore
1. Create `cmd/restore.go`
2. `restore system <file>` — require `--yes --confirm=<basename(file)>`
3. `restore tenant <file> --tenant-id=X` — require `--yes --confirm=X`
4. Stream file upload via multipart POST
5. Destructive warning in `--help` Long description

### Step 6: Agents export/import
1. Create `cmd/agents_export.go`
2. `agents export <id>` → `GET /v1/agents/{id}/export` → writes archive to `-o`
3. `agents export --all` → dùng team export hoặc loop — check server semantics
4. `agents import <file>` → `POST /v1/agents/import/preview` (default), `--apply` switches to `POST /v1/agents/import`
5. `agents import-merge <id> <file>` → `POST /v1/agents/{id}/import`

### Step 7: Teams/Skills/MCP export-import
1. Mirror pattern from agents
2. `teams export <id>` → `GET /v1/teams/{id}/export`
3. `teams import <file>` → `POST /v1/teams/import`
4. `skills export/import` → `/v1/skills/export`, `/v1/skills/import`
5. `mcp export/import` → `/v1/mcp/export`, `/v1/mcp/import`

### Step 8: Tests
1. `cmd/backup_test.go` — httptest signed token flow
2. `cmd/restore_test.go` — verify typed confirmation rejects mismatched input
3. `cmd/agents_export_test.go` — verify preview vs apply
4. Fixture backup file trong `testdata/` for restore integration test

### Step 9: Docs
1. Add CAUTION section to restore in README
2. Document signed token expiry behavior
3. `docs/codebase-summary.md` migration section

## Todo List

- [ ] 1.1: `internal/client/signed_download.go`
- [ ] 1.2: Signed download unit test
- [ ] 2.1: `cmd/backup.go` system + system-preflight
- [ ] 2.2: `cmd/backup.go` system-download + `--wait` flag
- [ ] 3.1: tenant backup subcommands
- [ ] 4.1: `cmd/backup_s3.go` config get/set
- [ ] 4.2: S3 list/upload/backup
- [ ] 5.1: `cmd/restore.go` system with typed confirmation
- [ ] 5.2: `cmd/restore.go` tenant with typed confirmation
- [ ] 5.3: Destructive warning in help text
- [ ] 6.1: `cmd/agents_export.go` export single + all
- [ ] 6.2: `cmd/agents_export.go` import with preview default
- [ ] 6.3: import-merge subcommand
- [ ] 7.1: teams export/import
- [ ] 7.2: skills export/import
- [ ] 7.3: mcp export/import
- [ ] 8.1: backup tests
- [ ] 8.2: restore tests (confirmation enforcement)
- [ ] 8.3: export/import tests (preview vs apply)
- [ ] 9.1: README restore CAUTION section
- [ ] 9.2: docs/codebase-summary.md migration section

## Success Criteria
- [ ] `goclaw backup system --wait -o backup.tgz` creates + downloads backup
- [ ] `goclaw backup s3 backup` creates + uploads to S3
- [ ] `goclaw restore system backup.tgz` without `--yes` refuses
- [ ] `goclaw restore system backup.tgz --yes --confirm=wrong` refuses
- [ ] `goclaw restore system backup.tgz --yes --confirm=backup.tgz` proceeds
- [ ] `goclaw agents export abc > abc.zip` produces valid archive
- [ ] `goclaw agents import abc.zip` shows preview, no mutation
- [ ] `goclaw agents import abc.zip --apply` performs import
- [ ] All types (agents/teams/skills/mcp) export+import round-trip successful
- [ ] Tests pass, coverage ≥60% for new code

## Risk Assessment

| Risk | Mitigation |
|---|---|
| Restore wipes production data | Typed confirmation + destructive warning in help + CHANGELOG note |
| Signed token leak via shell history | Document not to share URLs; server rotates tokens ~15min |
| Large backup OOM | Streaming download/upload; never buffer full file |
| Import preview/apply confusion | Default to preview, require explicit `--apply` |
| S3 credentials in config get leak | Mask secret-key in output |
| Partial restore failure corrupts DB | Server-side transactional; CLI documents rollback not guaranteed |
| File size agents.go + export → >600 LoC | Split `cmd/agents_export.go` |

## Security Considerations
- Signed download URL: treat as secret, don't log, don't echo to stdout in JSON mode (only via -o file)
- S3 secret-key: mask in `s3 config get` output; support `--show-secret` flag for explicit reveal
- Restore file: validate archive signature if server supports (check openapi spec)
- Tenant restore with wrong tenant-id could overwrite — typed confirmation mandatory
- Destructive ops log to audit trail server-side; CLI should state this in `--help`

## Next Steps
- Dependencies: Phase 0
- Unblocks: None
- Follow-up: Consider `goclaw backup schedule` command nếu server thêm cron-backup (future phase)

## Unresolved Questions
1. Signed token TTL: có cần CLI retry nếu token expire trong lúc download không?
2. `--all` export pagination: nếu user có 1000 agents, export serial có OK không hay cần parallel?
3. Import conflict resolution: server trả conflict info qua preview? CLI cần render đẹp?
