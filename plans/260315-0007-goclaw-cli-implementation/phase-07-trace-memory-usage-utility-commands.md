---
phase: 7
title: Trace, Memory, Usage & Utility Commands
status: planned
priority: medium
effort: M
depends_on: [phase-02]
---

# Phase 7 — Trace, Memory, Usage & Utility Commands

## Overview
LLM trace viewer, memory management, knowledge graph, usage analytics, and delegations.

## Requirements

### Traces
```
goclaw traces list [--agent <id>] [--status running|success|error] [--limit N]
goclaw traces get <traceID>
goclaw traces export <traceID> [--output <file>]
goclaw usage costs [--from <date>] [--to <date>]
```

### Memory
```
goclaw memory list <agentID> [--user <userID>]
goclaw memory get <agentID> <path>
goclaw memory store <agentID> <path> --content <content|@file>
goclaw memory delete <agentID> <path> [--yes]
goclaw memory search <agentID> --query <text> [--user <userID>]
```

### Knowledge Graph
```
goclaw knowledge-graph query <agentID> [--entity <name>]
goclaw knowledge-graph extract <agentID> --text <text>
goclaw knowledge-graph link <agentID> --from <entity> --to <entity> --relation <type>
```

### Usage & Analytics
```
goclaw usage summary [--from <date>] [--to <date>]
goclaw usage detail [--agent <id>] [--provider <name>]
```

### Delegations
```
goclaw delegations list [--agent <id>] [--limit N]
goclaw delegations get <id>
```

### Approvals
```
goclaw approvals list [--status pending]
goclaw approvals approve <id>
goclaw approvals deny <id> [--reason <text>]
```

## Implementation Steps

1. `cmd/traces.go` — Trace list/get/export with span tree rendering
2. `cmd/memory.go` — Memory document CRUD + semantic search
3. `cmd/knowledge_graph.go` — KG query/extract/link
4. `cmd/usage.go` — Usage stats with cost breakdown
5. `cmd/delegations.go` — Delegation history viewer
6. `cmd/approvals.go` — Execution approval management
7. Trace detail: render span tree with indentation, duration, token counts
8. Trace export: download gzipped JSON, save to file
9. Memory store: read content from `@file` syntax
10. Usage summary: show table with agent, provider, tokens, cost

## Related Code Files
- Create: `cmd/traces.go`, `cmd/memory.go`, `cmd/knowledge_graph.go`
- Create: `cmd/usage.go`, `cmd/delegations.go`, `cmd/approvals.go`

## Todo
- [ ] Trace list with filters
- [ ] Trace detail with span tree visualization
- [ ] Trace export to file
- [ ] Cost summary
- [ ] Memory CRUD operations
- [ ] Memory semantic search
- [ ] Knowledge graph operations
- [ ] Usage analytics display
- [ ] Delegation history
- [ ] Approval management

## Success Criteria
- `goclaw traces list` shows traces with token/cost summary
- `goclaw traces get <id>` renders span tree hierarchically
- `goclaw memory search <agent> --query "API design"` returns semantic results
- `goclaw usage summary` shows cost breakdown by agent/provider
- `goclaw approvals list` shows pending approvals, `approve` works
