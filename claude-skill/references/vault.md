# Knowledge Vault

## When to use

User wants to manage knowledge base for agents — upload/search documents, manage links/relationships, inspect knowledge graph, or trigger indexing pipeline. Vault enables agents to retrieve relevant context during reasoning.

## Commands in scope

- `goclaw vault list` / `goclaw vault documents list` — list vault documents (source: `cmd/vault.go`)
- `goclaw vault documents get <id>` — get document content + metadata
- `goclaw vault documents create` — create document (plain text)
- `goclaw vault documents update <id>` — update document fields
- `goclaw vault documents delete <id>` — delete document
- `goclaw vault upload <file>` — upload file (multipart streaming; supports .pdf, .txt, .md, etc.)
- `goclaw vault search <query>` — full-text + vector search (FTS + embeddings)
- `goclaw vault links list <doc-id>` — show related documents (graph)
- `goclaw vault tree` — show document tree (hierarchy by folder/tag)
- `goclaw vault graph` — get knowledge graph (JSON edges/nodes)
- `goclaw vault enrichment` — manage enrichment pipeline (summarize, extract entities, etc.)
- `goclaw vault rescan` — re-index all documents (admin-only; async, long-running)

## Verified flags

### `vault documents create/update`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--title <t>` | string | Document title |
| `--content <c>` | string | Plain text content (`@filepath` or literal) |
| `--tags <t>` | CSV | Comma-separated tags for filtering |
| `--metadata <m>` | JSON | Custom metadata object |

### `vault upload`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<file>` | path | File to upload (.pdf, .txt, .md, .docx, etc.) |
| `--title <t>` | string | Override detected title |
| `--tags <t>` | CSV | Tags for the uploaded document |

### `vault search`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<query>` | string | Search query (free-text or semantic) |
| `--limit <n>` | int | Max results (default 20) |
| `--offset <n>` | int | Pagination offset |

### `vault links list`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<doc-id>` | string | Document ID |

### `vault tree`
No flags (shows full hierarchy).

### `vault graph`
No flags (returns full graph as JSON).

### `vault enrichment`
See subcommands with `--help`.

### `vault rescan`
Async; returns task ID. Admin-only.

## JSON output

- ✅ `vault documents list/get`, `vault search`, `vault links list`, `vault tree`, `vault graph` — JSON
- ✅ `vault upload` — JSON (document ID + metadata)
- ⚠️ `vault documents create/update/delete` — success text (check exit code)
- ✅ `vault enrichment` — varies by subcommand
- ✅ `vault rescan` — JSON task ID

## Destructive ops

| Command | Confirm |
| --- | --- |
| `vault documents delete` | YES (permanent removal) |

## Common patterns

### Example 1: upload document to vault
```bash
goclaw vault upload /tmp/design.pdf --title "System Design" --tags "architecture,v2" --output json
```

### Example 2: search vault
```bash
goclaw vault search "microservices scalability" --limit 10 --output json
```

### Example 3: list documents
```bash
goclaw vault documents list --output json
```

### Example 4: get document + metadata
```bash
goclaw vault documents get <doc-id> --output json
```

### Example 5: create plain-text document
```bash
goclaw vault documents create \
  --title "API Guidelines" \
  --content "Endpoints should use REST conventions..." \
  --tags "api,standards"
```

### Example 6: inspect knowledge graph
```bash
goclaw vault graph --output json
# → {"nodes": [...], "edges": [...]}
```

### Example 7: show related documents
```bash
goclaw vault links list <doc-id> --output json
# → documents linked (via title/tag/semantic similarity)
```

### Example 8: re-index vault (admin)
```bash
goclaw vault rescan --output json
# → {"task_id": "..."}
```

## Edge cases & gotchas

- **File uploads:** server supports .pdf, .txt, .md, .docx, .pptx. Binary formats extracted via OCR/text-extraction; large files (>100MB) may be rejected.
- **Vector embeddings:** uploaded documents auto-vectorized for semantic search. Depends on embedding provider (OpenAI, Anthropic, local).
- **Search:** FTS (full-text) + vector search. Queries match keywords + semantic intent.
- **Knowledge graph:** auto-linked via title overlap, shared tags, semantic similarity. Manual link creation not exposed via CLI yet.
- **Rescan:** async + long-running (can take minutes for large vaults). Returns task ID; poll or check server logs.
- **Metadata:** custom fields stored but not indexed. Use tags for searchable attributes.
- **Enrichment pipeline:** optional. Extracts summaries, entities, tables. Admin can enable/configure.

## Cross-refs

- Memory (agent-specific, not vault): [knowledge-memory.md](knowledge-memory.md)
- Document export/import: [data-movement.md](data-movement.md)
