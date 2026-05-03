# Phase 6 — Server FR Backlog (Deferred)

**Priority:** 🟢 nice-to-have
**Status:** server-blocked — NO CLI WORK
**Action:** File upstream issues in `goclaw` repo with concrete payload examples.

## Context Links

- Gap analysis: `plans/reports/brainstorm-260503-1907-gap-analysis-round2.md` § 5 (P6)

## Rationale

Per YAGNI: shipping CLI stubs for endpoints that don't exist server-side = broken UX + maintenance debt. Track as upstream FRs.

## Backlog Items

| ID | CLI surface (proposed) | Required server work | Priority |
|---|---|---|---|
| O1 | `goclaw traces follow [--agent=…]` | `GET /v1/traces/follow` SSE stream | 🟡 — popular AI poll pattern |
| O2 | `goclaw traces replay <trace-id>` | `POST /v1/traces/{id}/replay` (synthesize from existing data) | 🟢 |
| O3 | `goclaw logs aggregate --since=…` | `GET /v1/logs/aggregate` with bucket params | 🟢 |
| P1 | `goclaw providers reconnect <id>` | `POST /v1/providers/{id}/reconnect` | 🟡 |
| C2 | `goclaw channels writers test <id>` | `POST /v1/channels/instances/{id}/writers/test` | 🟢 |
| E2 | `goclaw chat sessions branch <key>` | `POST /v1/chat/sessions/{key}/branch` (fork/copy session state) | 🟡 |
| E5 | `goclaw chat history --follow` | WS `chat.history.delta` push or SSE | 🟢 |
| X5 | `goclaw tts synthesize --text=… --voice=…` | already exists `POST /v1/tts/synthesize` — verify; if exists, demote to P5 | 🟢 |

## Recommended Issue Template (per item)

```
**Title:** [CLI Coverage] Add `<endpoint>` for goclaw-cli `<command>`

**Use case:**
AI tools using goclaw-cli need <thing> to <accomplish goal>.

**Proposed endpoint:**
- Method + path
- Request body example (JSON)
- Response shape

**CLI consumer:**
After endpoint ships, CLI will add `goclaw <command>` mapping 1:1.

**Priority justification:**
- AI-tool consumer share: ~90% of CLI users
- Why not synthesizable client-side: <reason>
```

## Todo List

- [ ] File 7 issues in goclaw repo (skip X5 if `tts synthesize` already exists)
- [ ] Track issue numbers in this file
- [ ] When server ships endpoint, demote item to a future CLI phase

## Success Criteria

- All 7-8 items have upstream issue links recorded here.
- Zero CLI stub commands shipped.
- README does NOT advertise unimplemented commands.

## Open Questions

- Q5 (from report): Server FR ownership — CLI maintainer or upstream PM files? Resolve before opening issues.
- Q7 (from report): `tts synthesize` AI use case — confirms whether to file or demote.

## Next Steps

When ≥ 50% of P6 items have server endpoints, plan a P7 CLI catch-up phase.
