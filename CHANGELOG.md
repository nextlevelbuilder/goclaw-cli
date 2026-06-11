# Changelog

All notable changes to goclaw-cli are documented here.
Format: [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## [Unreleased] — Domain Coverage Expansion (P0–P6)

### Added

**P0 — Critical**
- `goclaw hooks` (list, create, update, delete, toggle, test, history) — manage event hooks via WS RPC `hooks.*`. Closes the entire hooks domain that was previously unreachable from CLI.
- `goclaw agents files` (list, get, set) — edit global agent context files (AGENTS.md, SOUL.md, IDENTITY.md, USER.md, USER_PREDEFINED.md, CAPABILITIES.md, BOOTSTRAP.md, MEMORY.json, HEARTBEAT) via WS RPC `agents.files.*`. `--propagate` pushes change to all existing user instances.

**P1 — Lifecycle & analytics**
- `goclaw agents cancel-summon <id>` — cancel an in-progress summon (`POST /v1/agents/{id}/cancel-summon`).
- `goclaw agents skills list <id>` — list skills granted to an agent (`GET /v1/agents/{id}/skills`).
- `goclaw usage timeseries` — bucketed usage over time (`GET /v1/usage/timeseries`).
- `goclaw usage breakdown` — usage broken down by agent/user/tenant (`GET /v1/usage/breakdown`).

**P2 — Coverage completion**
- `goclaw tts test-connection` — test a TTS provider end-to-end (`POST /v1/tts/test-connection`).
- `goclaw voices list` and `goclaw voices refresh` — voice catalog.
- `goclaw memory kg extract` — switched to new endpoint `POST /v1/agents/{id}/kg/extract` (legacy `/v1/knowledge-graph` path retired).
- `goclaw files sign` — generate signed URL for server-side file (`POST /v1/files/sign`).
- `goclaw teams workspace upload` and `goclaw teams workspace move` — multipart upload + rename for team workspace.
- `goclaw packages github-releases` — list GitHub releases for tracked packages.

**P3 — AI-critical fillers**
- `goclaw profile` (list, current, create, use, delete) — first-class CLI profile management with safe profile names.
- `GOCLAW_PROFILE` — per-command profile selection precedence between `--profile` and active config.
- `goclaw sessions compact <key>` — invokes WS RPC `sessions.compact` behind destructive confirmation.
- `goclaw health` — uses WS RPC `health` when authenticated, retaining unauthenticated HTTP `/health` fallback.
- `goclaw traces list --agent --user --session-key --status --channel --limit --offset` — server-aligned filters for paginated trace listing.

**P4 — UX polish**
- `goclaw codex-pool activity --agent=<id>|--provider=<id>` — unified Codex pool activity lookup; legacy agent/provider commands remain as deprecated aliases.
- `goclaw api-keys rotate <id>` — create replacement key, show raw key once, then revoke old key with structured partial-failure reporting.
- `goclaw config defaults` — read-only WS passthrough for server default config values.
- `goclaw chat replay <agent> --session=<key>` and `goclaw chat sessions resume <agent> --session=<key>` — discoverability wrappers over existing chat session contracts.
- `goclaw tools invoke <name> --args=<json|@file>` — alias for `--params` with file-backed JSON support.

**P5 — Residual command fillers**
- `goclaw teams attachments download <team-id> <attachment-id> --output <file>` — authenticated attachment download with required output path and no-overwrite default.
- `goclaw agents evolution skill apply <agent-id> <suggestion-id> [--skill-draft @file]` — explicit wrapper for approving `skill_add` suggestions through the server evolution approval route.
- `goclaw agents evolution update` now maps `--action=accept|reject` to the server-compatible `status=approved|rejected` payload.

**P6 — Backend-unblocked surfaces (gateway `v3.12.0-beta.20`+)**
- `goclaw traces list --q --agent-query --channel-query --from --to --min-input-tokens --max-input-tokens --min-output-tokens --max-output-tokens --min-tool-calls --max-tool-calls --tool-name --has-tool-calls` — forwards the trace search and advanced filters added by server PR #155 while preserving existing `--agent` as `agent_id`.
- `goclaw traces follow --session-key|--agent [--since RFC3339] [--limit N]` — one-shot incremental trace polling (`GET /v1/traces/follow`). Re-invoke with returned cursor to advance; no WS stream, no watch loop.
- `goclaw traces timeline <run-id> [--session-key K] [--limit N] [--offset N]` — read archived run timeline items (`GET /v1/runs/{runID}/timeline`) without replaying or mutating a run.
- `goclaw providers reconnect <provider-id>` — hot-reconnect a provider, bumping the registry without touching credentials (`POST /v1/providers/{id}/reconnect`).
- `goclaw sessions branch <session-key> --up-to-index N [--new-session-key K] [--label L] [--metadata k=v ...]` — branch a chat session at a 1-based message index into a new session (`POST /v1/chat/sessions/{key}/branch`). `--up-to-index=0` is preserved on the wire.
- `goclaw sessions follow <session-key> [--cursor N] [--limit N]` — one-shot cursor-based history poll (`GET /v1/chat/sessions/{key}/history/follow`). Not a stream; `--cursor=0` is preserved literally in the query string.
- `goclaw channels writers test <instance-id> --group-id G --user-id U` — probe a (group, user) pair against a channel's writer policy without mutating state (`POST /v1/channels/instances/{id}/writers/test`). Request body has exactly two keys.
- `goclaw activity aggregate --group-by {action|actor_type|entity_type|actor_id} [--from --to --limit --actor-type --actor-id --action --entity-type --entity-id]` — group audit-log activity by dimension with bucket counts (`GET /v1/activity/aggregate`). Attached as subcommand of existing `activity` parent.
- `goclaw logs aggregate [--group-by {level|source}] [--level --source --from]` — summarize the runtime log ring buffer (`GET /v1/logs/runtime/aggregate`, admin-only). Distinct from `logs tail`. Epoch-millis `last_seen` rendered as RFC3339, never scientific notation.

**Runtime & Packages parity**
- `goclaw credentials agent-credentials` — list/get/set/delete per-agent credential material for secure CLI credentials.
- `goclaw packages updates apply-all [packages...]` — accepts positional package specs in addition to `--packages`.

### Fixed

- `goclaw packages list` now decodes current server grouped payloads `{system,pip,npm,github}` in table mode while preserving raw object payloads for JSON/YAML.
- `goclaw packages install` and `goclaw packages uninstall` now send the server-compatible `package` key; legacy `--runtime python|node` translates to `pip:`/`npm:` specs.
- `goclaw packages runtimes`, `packages deny-groups`, and `packages github-releases --repo --limit` now match current server envelopes and required query parameters.
- `goclaw credentials list`, `credentials presets`, `credentials agent-grants list`, and `credentials user-credentials list` now decode current server envelope payloads.
- `goclaw traces list` now decodes the current server payload `{traces,total,limit,offset}`. JSON/YAML mode preserves that envelope; table mode renders rows from `traces` using `id`, `total_input_tokens`, `total_output_tokens`, and `total_cost`.
- `goclaw traces get <id>` — TTY mode now renders a human-readable summary (header card + span tree) instead of dumping raw JSON. JSON-mode payload unchanged. Decode failures surface as wrapped errors instead of an empty `{}`. Trace ids are validated against `^[A-Za-z0-9._-]+$` and reserved tokens (`.`, `..`) are rejected before any HTTP call. Distinct exit codes per failure: not-found → 3, permission-denied → 2, malformed-id → 4, server-failure → 5. Latent retry-body bug in `internal/client/http.go` fixed: the final 5xx/429 response body is now preserved so the typed `APIError` reaches the caller (previously collapsed to exit 1). Closes #17.
- `goclaw traces get <id>` now handles the current server detail payload `{trace,spans}` while preserving the server envelope in JSON/YAML mode.
- `goclaw traces export <id>` now validates trace IDs and path-escapes the export route before making the HTTP request.

### Notes
- All new commands honor the AI-first ergonomics contract: `--output=json` envelope, central error handler, `--yes` for destructive ops, `--quiet` for CI.
- P4/P5 backlog was re-swept against the current CLI surface; already-covered items were removed from residual scope before implementation.
- Out of scope: OpenAI-compatible `/chat/completions` and `/v1/responses` endpoints (client APIs, not admin CLI surface).

---

## [Unreleased] — AI Ergonomics Foundation (Phase 0)

### Breaking Changes

#### Output format default changed when stdout is piped

**Before:** `goclaw agents list` always defaulted to `table` format regardless of context.

**After:** When stdout is not a terminal (piped, redirected, CI), the default format is now `json`.

**Migration:** Scripts relying on table output must add `--output=table` or `GOCLAW_OUTPUT=table`.

```bash
# Before (broke silently in CI)
goclaw agents list | grep "my-agent"

# After — explicit table for text parsing
goclaw agents list --output=table | grep "my-agent"

# Or use JSON (recommended for automation)
goclaw agents list | jq '.[] | select(.display_name == "my-agent")'
```

**Rationale:** AI tools, CI pipelines, and shell scripts consuming CLI output require
structured JSON. Table format is human-optimised and breaks piped parsing. TTY detection
ensures human operators still get tables by default.

### Added

- **`internal/output/exit.go`** — Exit code constants (0-6) + `MapServerCode(code)` + `MapHTTPStatus(status)` + `Exit(code)`
- **`internal/output/error.go`** — `ErrorDetail` / `ErrorEnvelope` types matching server `ErrorShape`; `ParseHTTPError(body, status)`; `PrintError(err, format)`; `FromError(err) int`
- **`internal/output/tty.go`** — `IsTTY(fd)` via `golang.org/x/term`; `ResolveFormat(flagVal)` with flag > `GOCLAW_OUTPUT` env > TTY precedence
- **`internal/client/follow.go`** — `FollowStream(ctx, ...)` with exponential backoff reconnect (max 5 retries) for `--follow` streaming commands
- **`--quiet` flag** — persistent flag on root command; suppresses banners and informational messages in non-TTY contexts
- **Exit code contract** — all server error codes now map deterministically to exit codes 0-6 for AI/automation consumers

### Changed

- `cmd/root.go` — output format resolved via TTY detection in `PersistentPreRunE`; central error handler in `Execute()` calls `output.PrintError` + `output.Exit(output.FromError(err))`
- `cmd/logs.go` — `logs tail --follow` migrated to `client.FollowStream` with auto-reconnect; banner gated behind `--quiet` and TTY check
- `--output` flag default changed from `"table"` to `""` (empty triggers auto-detect)
- `internal/client/errors.go` — `APIError` extended with `Details`, `Retryable`, `RetryAfterMs` fields matching server `ErrorShape`; added interface methods (`ErrorCode`, `ErrorMessage`, `ErrorDetails`, `IsRetryable`, `RetryAfter`, `HTTPStatus`) for duck-typed error handling in `output` package without import cycle

### Fixed

- Piped invocations no longer silently produce unparseable table output; they emit valid JSON
- Error details from server (`code`, `message`, `retryable`) are now fully preserved and passed through to the caller

---

## Previous releases

See git log for changes prior to this changelog.
