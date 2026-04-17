# Docs & API

## When to use

User asks about server REST API endpoints, OpenAPI spec, or how to call the server directly.

## Commands in scope

- `goclaw api-docs spec` — fetch OpenAPI/JSON spec (source: `cmd/api_docs.go`)
- `goclaw api-docs open` — **launches system browser — skill REFUSES** (not useful from Claude)

## Verified flags

### `api-docs spec`
No specific flags beyond globals. Use `--output json` to get spec as parseable JSON.

### `api-docs open`
No flags. Attempts to run `open` (macOS), `xdg-open` (Linux), or `rundll32` (Windows).

## JSON output

- ✅ `api-docs spec --output json` — full OpenAPI JSON spec
- ⚠️ `api-docs open` — no output, just spawns browser

## Destructive ops

None.

## Common patterns

### Example 1: fetch spec for analysis
```bash
goclaw api-docs spec --output json > /tmp/openapi.json
# then use jq/yq to explore endpoints
jq '.paths | keys' /tmp/openapi.json
```

### Example 2: user asks "what endpoints exist for agents?"
```bash
goclaw api-docs spec --output json | jq '.paths | to_entries[] | select(.key | contains("agents"))'
```

## Edge cases & gotchas

- **`api-docs open`:** skill MUST refuse. Browser launch is useless in Claude Code context. Suggest `api-docs spec` instead.
- **Spec format:** OpenAPI 3.x JSON. Large (often >500KB). Don't dump whole spec into chat — pipe through jq with a filter.
- **Auth:** spec endpoint may require auth on some deployments. If 401, run `goclaw auth login`.

## Cross-refs

- Direct HTTP if spec doesn't cover: see server repo documentation
- CLI commands mirror most of API: [agents-core.md](agents-core.md), [chat-sessions.md](chat-sessions.md), etc.
