# Phase Implementation Report

## Executed Phase
- Phase: Phase 2 — Missing Resource Commands
- Plan: plans/260326-1350-cli-feature-parity-update/
- Status: completed (with conflict resolution)

## Files Created
| File | Lines | Notes |
|------|-------|-------|
| cmd/knowledge_graph.go | 148 | KG entities CRUD + traverse/graph/stats; extends existing `kgCmd` from memory.go |
| cmd/packages.go | 90 | list, runtimes, install, uninstall |
| cmd/contacts.go | 100 | list, resolve, merge, unmerge, merged |
| cmd/pending_messages.go | 68 | list, compact, delete |

## Commands Covered by Existing Files (no new file needed)
| Requested | Existing file | Resolution |
|-----------|--------------|------------|
| credentials (list/get/create/test/update/delete/presets) | cmd/admin.go | Conflict — `credentialsCmd` already declared; new file deleted |
| activity list | cmd/admin.go | Conflict — `activityCmd` already declared as top-level cmd; new file deleted |
| tts (status/enable/disable/providers/set-provider) + convert | cmd/admin.go | Conflict — all tts vars already declared; new file deleted. `tts.convert` NOT available (would require admin.go edit) |
| media (upload/get) | cmd/admin.go | Conflict — `mediaCmd` already declared; new file deleted |
| usage (summary/costs) + breakdown/timeseries | cmd/traces.go | Conflict — `usageCmd`/`usageSummaryCmd`/`usageCostsCmd` already declared. `breakdown` and `timeseries` NOT available (would require traces.go edit) |

## Tasks Completed
- [x] cmd/knowledge_graph.go — entities list/get/create/delete, traverse, graph, stats
- [x] cmd/packages.go — list, runtimes, install, uninstall
- [x] cmd/contacts.go — list, resolve, merge, unmerge, merged
- [x] cmd/pending_messages.go — list, compact, delete
- [x] Compile: `go build ./...` passes clean
- [x] Vet: `go vet ./...` passes clean
- [x] Tests: `go test ./...` all pass

## Tests Status
- Type check / build: PASS
- go vet: PASS
- Unit tests: PASS (internal/client, internal/config, internal/output)

## Issues Encountered
1. **admin.go** pre-declares: `credentialsCmd`, `activityCmd`, `ttsCmd`, `mediaCmd` — prevented creating separate files for those 4 resources. Cannot modify existing files per task rules.
2. **traces.go** pre-declares: `usageCmd`, `usageSummaryCmd`, `usageCostsCmd` — prevented creating usage.go. `breakdown` and `timeseries` subcommands are not added.
3. `tts convert` subcommand not added (would need admin.go edit).

## Unresolved Questions
- Should `tts convert`, `usage breakdown`, `usage timeseries`, and the enhanced credentials commands (get/test/update/presets) be added by modifying the existing files (`admin.go`, `traces.go`)? Those would require explicit approval to edit existing files.
