# Scout Report: GoClaw CLI Feature Parity Gap Analysis

**Date:** 2026-03-26
**Scope:** GoClaw server (`/Volumes/GOON/www/nlb/goclaw`) vs CLI (`goclaw-cli`)

## Summary

CLI is ~70% feature-complete. Major gaps: multi-tenant support (critical), 9 missing command groups, 28+ missing subcommands on existing groups.

## Key Findings

### Critical: Multi-Tenant (March 23, commit cd022699)
- 30+ DB tables now have `tenant_id`
- API keys tenant-scoped
- All queries enforce `WHERE tenant_id = $N`
- CLI has ZERO tenant awareness

### Missing Commands (9 groups, ~35 endpoints)
| Command | Endpoints | Priority |
|---------|-----------|----------|
| tenants | 7 HTTP | critical |
| system-config | 4 HTTP | high |
| knowledge-graph | 8 HTTP | medium |
| usage | 4 HTTP | high |
| credentials | 6 HTTP | medium |
| packages | 4 HTTP | medium |
| tts | 6 WS | low |
| contacts | 5 HTTP | low |
| pending-messages | 3 HTTP | low |

### Missing Subcommands on Existing (28+ endpoints)
- skills: +7 (versions, files, tenant-config, deps)
- teams: +9 (task approve/reject/assign/comment, events, scopes)
- channels: +4 (writers CRUD)
- tools: +2 (tenant-config)
- providers: +2 (embedding, claude-auth)
- storage: +2 (download, move)
- config: +3 (permissions)
- chat: +3 (inject, status, abort)
- heartbeat: +8 (new monitoring group)

### Oversized Files
- admin.go: 403 lines
- teams.go: 500 lines
- agents.go: 491 lines

### Server API Surface
- 100+ HTTP REST endpoints
- 47+ WebSocket RPC methods
- CLI covers ~70 of them

## Plan Created
`plans/260326-1350-cli-feature-parity-update/`
5 phases: multi-tenant → missing cmds → enhanced cmds → WS features → README/tests
