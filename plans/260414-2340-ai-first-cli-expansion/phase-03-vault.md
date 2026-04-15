---
phase: 3
title: Vault — Documents, Links, Search, Graph
status: pending
priority: high
blockedBy: [phase-00]
---

# Phase 3 — Vault

## Context Links
- Brainstorm §2 G4, §5 Phase 3
- Server HTTP: `../../../goclaw/internal/http/vault_handlers.go`, `vault_handler_documents.go`, `vault_handler_links.go`, `vault_handler_tree.go`, `vault_handler_upload.go`, `vault_graph_handler.go`

## Overview
- **Priority:** High (🔥, user explicit request; search is AI-critical)
- **Status:** Pending
- **Description:** Vault = knowledge document graph (credential/document management). 15+ server endpoints. AI-critical: `vault search` + `vault documents get` for RAG use cases. Mutation ops (create/update/delete/upload) medium priority.

## Key Insights
- 3 sub-domains: documents, links (relationships between docs), graph (rendered)
- `upload` là multipart HTTP, khác pattern JSON body
- `search` có thể return large result — default pagination + JSON streaming
- `enrichment status/stop` là async background task control
- `tree` returns hierarchical structure — render as ASCII tree in table mode, JSON otherwise
- Graph có 2 endpoints: `/v1/vault/graph` (system) và `/v1/agents/{agentID}/kg/graph/compact` (per-agent KG compact) — latter thuộc memory KG group, không phải vault

## Requirements

### Functional
- `vault documents list [--q=<query>] [--limit=N] [--offset=N]`
- `vault documents get <docID>`
- `vault documents create --title=... --content=... | --file=...`
- `vault documents update <docID> --title=... --content=...`
- `vault documents delete <docID>` (--yes required)
- `vault documents links <docID>` — show links for a doc
- `vault links create --from=<docID> --to=<docID> --type=<relType>`
- `vault links delete <linkID>`
- `vault links batch-get <docIDs...>` — batch fetch links for multiple docs
- `vault upload <file> [--title=...] [--tags=...]`
- `vault rescan` — trigger server rescan
- `vault tree [--depth=N]` — hierarchy
- `vault search <query> [--limit=N]`
- `vault enrichment status`
- `vault enrichment stop`
- `vault graph [--format=dot|json]` — full vault graph

### Non-Functional
- `delete` requires `--yes`
- Upload streaming multipart (no buffer full file)
- `search` support `--limit`/`--offset` pagination
- `tree` in JSON mode: nested objects; in table: indented tree with prefix chars
- Graph output: JSON default; `--format=dot` for Graphviz

## Architecture

### Command tree
```
goclaw vault
├── documents
│   ├── list
│   ├── get <docID>
│   ├── create --title --content|--file
│   ├── update <docID>
│   ├── delete <docID>
│   └── links <docID>
├── links
│   ├── create --from --to [--type]
│   ├── delete <linkID>
│   └── batch-get <docIDs...>
├── upload <file>
├── rescan
├── tree
├── search <query>
├── enrichment
│   ├── status
│   └── stop
└── graph [--format=json|dot]
```

### File structure (split to respect 200 LoC rule)
```
cmd/
├── vault.go                    # root cmd + simple subcommands (rescan, tree, search, graph) — ~120 LoC
├── vault_documents.go          # documents CRUD + links listing — ~180 LoC
├── vault_links.go              # standalone links group — ~100 LoC
├── vault_upload.go             # multipart upload — ~80 LoC
└── vault_enrichment.go         # status/stop — ~60 LoC
```

## Related Code Files

### Create
- `cmd/vault.go`
- `cmd/vault_documents.go`
- `cmd/vault_links.go`
- `cmd/vault_upload.go`
- `cmd/vault_enrichment.go`

### Modify
- `cmd/root.go` — register `vaultCmd`
- `internal/client/http.go` — add `PostMultipart` if not exists
- `internal/output/tree.go` (new helper) — render tree in table mode
- `docs/codebase-summary.md`
- `README.md`

### Reference
- `internal/client/http.go` Get/Post/Put/Delete helpers
- `internal/output/printer.go`

## Implementation Steps

### Step 1: vault root + simple queries
1. `cmd/vault.go` với `vaultCmd`, register tree/search/rescan/graph
2. `tree` → `GET /v1/vault/tree?depth=N`
3. `search <q>` → `POST /v1/vault/search {query, limit, offset}`
4. `rescan` → `POST /v1/vault/rescan` (may be admin-only — check)
5. `graph --format=json|dot` → `GET /v1/vault/graph`, transform to DOT if requested

### Step 2: documents
1. `cmd/vault_documents.go`
2. list → `GET /v1/vault/documents` (with q/limit/offset)
3. get → `GET /v1/vault/documents/{docID}`
4. create → `POST /v1/vault/documents` (read `--file` or `--content`)
5. update → `PUT /v1/vault/documents/{docID}`
6. delete → `DELETE /v1/vault/documents/{docID}` (+ `--yes`)
7. links subcommand → `GET /v1/vault/documents/{docID}/links`

### Step 3: links
1. `cmd/vault_links.go`
2. create → `POST /v1/vault/links`
3. delete → `DELETE /v1/vault/links/{linkID}`
4. batch-get → `POST /v1/vault/links/batch`

### Step 4: upload
1. `cmd/vault_upload.go`
2. Multipart POST `/v1/vault/upload`
3. Support `--title`, `--tags=tag1,tag2`

### Step 5: enrichment
1. `cmd/vault_enrichment.go`
2. status → `GET /v1/vault/enrichment/status`
3. stop → `POST /v1/vault/enrichment/stop`

### Step 6: Tree rendering
1. Create `internal/output/tree.go` with `PrintTree(root TreeNode, out io.Writer)` — indented tree with `├─`/`└─` prefixes for TTY
2. JSON mode: pass nested structure through
3. Unit tests for tree render

### Step 7: DOT format
1. In `graph` command, if `--format=dot` then transform JSON graph → DOT string
2. Helper function in `cmd/vault.go` or `internal/output/dot.go`

### Step 8: Tests
1. Per-file tests, httptest for all HTTP endpoints
2. Multipart upload test với form boundary
3. Tree rendering unit tests

### Step 9: Docs
1. README vault section với RAG search example
2. `docs/codebase-summary.md` vault subsystem
3. Help examples showing search JSON output piped into jq

## Todo List

- [ ] 1.1: `cmd/vault.go` skeleton + root cmd
- [ ] 1.2: tree/search/rescan
- [ ] 1.3: graph + DOT format
- [ ] 2.1: `cmd/vault_documents.go` list/get
- [ ] 2.2: documents create (support `--file` read)
- [ ] 2.3: documents update/delete (+ `--yes`)
- [ ] 2.4: documents links subcommand
- [ ] 3.1: `cmd/vault_links.go` create/delete/batch-get
- [ ] 4.1: `cmd/vault_upload.go` multipart upload
- [ ] 5.1: `cmd/vault_enrichment.go` status/stop
- [ ] 6.1: `internal/output/tree.go` TreeNode + PrintTree
- [ ] 6.2: Tree unit tests
- [ ] 7.1: DOT format transform helper
- [ ] 8.1: Vault root tests (tree/search/graph)
- [ ] 8.2: Documents tests
- [ ] 8.3: Links + upload + enrichment tests
- [ ] 9.1: README vault section with examples
- [ ] 9.2: docs/codebase-summary.md

## Success Criteria
- [ ] `goclaw vault search "authentication"` returns JSON array of matches
- [ ] `goclaw vault tree --depth=2` renders tree (TTY) or JSON (piped)
- [ ] `goclaw vault documents create --title=X --file=doc.md` creates doc, returns docID
- [ ] `goclaw vault upload file.pdf --tags=a,b` uploads multipart, returns docID
- [ ] `goclaw vault documents delete <id>` refuses without `--yes`
- [ ] `goclaw vault graph --format=dot | dot -Tpng > graph.png` produces valid visualization
- [ ] `goclaw vault enrichment status` returns JSON with progress fields
- [ ] Exit codes P0-compliant
- [ ] Tests pass, ≥60% coverage

## Risk Assessment

| Risk | Mitigation |
|---|---|
| Large upload OOM | Multipart streaming, never buffer full |
| `rescan` triggers long-running job, CLI hangs | Return immediately with job ID; `enrichment status` to poll |
| Search result flood terminal | Default `--limit=20`, configurable |
| Graph huge for large vault | `--format=dot` supports Graphviz; JSON default paginable |
| Typed delete confirmation missing for docs with many links | Server-side cascade warning in error response; CLI pass-through |
| DOT format implementation bug | Fallback to JSON on error; log warning to stderr |

## Security Considerations
- Vault documents có thể chứa credentials/PII — CLI output preserves server-side masking (trust server)
- Upload: validate file exists + readable before opening stream
- Links between sensitive docs: batch-get result may expose relationships — user-scoped by server
- Search query strings: not logged in CLI debug output (privacy)
- `enrichment stop` is admin-level — check server permission

## Next Steps
- Dependencies: Phase 0
- Unblocks: None
- Follow-up: If vault becomes heavy-use, consider cache layer in CLI (unlikely YAGNI)

## Unresolved Questions
1. DOT format: có cần CLI render graph ASCII hay chỉ output DOT?
2. Search: server có support fuzzy + semantic? CLI cần flag `--fuzzy`?
3. Upload: có file size limit server-side? CLI warn nếu >100MB?
