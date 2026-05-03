# Phase 3 — AI-Critical Fillers

**Priority:** 🔥 critical
**Status:** not-started
**Estimated LoC:** ~250 (excl. tests)
**Estimated PR size:** ≤ 500 LoC incl. tests

## Context Links

- Gap analysis: `plans/reports/brainstorm-260503-1907-gap-analysis-round2.md` § 3.D, 3.E, 3.A, § 5 (P3)
- AI ergonomics contract: `CLAUDE.md` § AI-First Ergonomics

## Overview

Tier-🔥 items only. AI-critical because they touch context-window pressure (sessions compact), multi-tenant orchestration (multi-profile), liveness probe (health), and observability filters (traces).

## Key Insights

- **Multi-profile** is CLI-only (no server dep) but largest blast radius — touches config loader, token store, every `newHTTP()` / `newWS()` call.
- **Sessions compact** is a 1-line WS method call (`MethodSessionsCompact = "sessions.compact"` in `goclaw/internal/gateway/methods/sessions.go:33`). Add to `cmd/sessions.go`.
- **Health** wires WS `health` method (`pkg/protocol/methods.go:46`). No HTTP route exists. Trivial.
- **Traces filter polish** = extending existing `traces list` with query-param-backed flags. Server already accepts them.

## Requirements

### Functional

- `goclaw profile use <name>` — switch active profile
- `goclaw profile list` — list profiles + mark active
- `goclaw profile create <name> [--copy-from=…]` — new profile
- `goclaw profile delete <name>` — delete (refuse if active, `--yes` override)
- `goclaw profile current` — print active profile name (machine-friendly)
- Global flag `--profile=<name>` overrides config for one invocation
- `goclaw sessions compact <key> [--yes]` — invoke WS `sessions.compact`
- `goclaw health` — invoke WS `health`, print raw response
- `goclaw traces list` extra flags: `--since`, `--agent`, `--status`, `--root-only`, `--limit`

### Non-functional

- Profile token storage: keyring namespace `goclaw-cli:<profile>:token`
- Default profile auto-named `default`; existing `~/.goclaw/config.yaml` migrates transparently → `~/.goclaw/profiles/default/config.yaml`
- Migration runs once on first profile-aware command; idempotent.
- Precedence: `--profile` flag > `GOCLAW_PROFILE` env > config "active_profile" key > literal `default`
- `GOCLAW_OUTPUT` env still wins over profile-specific output mode (per audit Q8)

## Architecture

```
internal/config/
  loader.go          # extended: ProfileDir(name) + activeName()
  profile.go         # NEW: Profile struct + List/Create/Delete/Use
  migrate.go         # NEW: migrate ~/.goclaw/config.yaml → profiles/default/

cmd/
  profile.go         # NEW: profile group + 5 subcommands
  root.go            # add --profile global flag, hook into config init
  sessions.go        # add compact subcommand
  health.go          # NEW: 1-file
  traces.go          # extend list with new flags
```

## Related Code Files

### Modify

- `cmd/root.go` — `--profile` flag, config init order
- `cmd/sessions.go` — add `compact`
- `cmd/traces.go` — extend `traces list` flags
- `internal/config/loader.go` — profile-aware path resolution
- `internal/client/auth.go` (if separate) — token lookup by profile

### Create

- `cmd/profile.go`
- `cmd/health.go`
- `internal/config/profile.go`
- `internal/config/migrate.go`
- `cmd/profile_test.go`
- `cmd/sessions_compact_test.go`
- `cmd/health_test.go`
- `cmd/traces_filters_test.go`
- `internal/config/profile_test.go`
- `internal/config/migrate_test.go`

### Delete

- none

## Implementation Steps

1. Read `internal/config/loader.go` + every caller of `ConfigDir()` / `LoadConfig()` to map blast radius.
2. Add `internal/config/profile.go` with `Profile{Name, Dir, ConfigPath, TokenKey}` + `List/Create/Delete/Use/Current`.
3. Add `internal/config/migrate.go` — detect legacy single-config, move to `profiles/default/`, set active.
4. Wire `--profile` global flag in `cmd/root.go`; resolve precedence; load active profile during PersistentPreRunE.
5. Implement `cmd/profile.go` — 5 subcommands + table/JSON output.
6. Update token store keyring key to include profile name.
7. Add `cmd/sessions.go` `compact` subcommand — 1 WS call, `--yes` for destructive prompt.
8. Add `cmd/health.go` — 1 WS call, raw JSON passthrough by default.
9. Extend `cmd/traces.go` `list` with new flags; map to query params; update help text.
10. Tests for each new file (httptest server for HTTP, gorilla WS upgrader for WS, temp-dir for profile fs ops).
11. Update `docs/codebase-summary.md` (profile architecture), `CHANGELOG.md` (Unreleased — Phase 3).
12. `go build ./... && go vet ./... && go test ./...`

## Todo List

- [ ] config: profile struct + List/Create/Delete/Use/Current
- [ ] config: legacy migration (config.yaml → profiles/default/config.yaml)
- [ ] root: --profile flag + GOCLAW_PROFILE env + precedence
- [ ] keyring: profile-namespaced token key
- [ ] cmd/profile.go: 5 subcommands
- [ ] cmd/sessions.go: compact subcommand
- [ ] cmd/health.go: WS health probe
- [ ] cmd/traces.go: --since/--agent/--status/--root-only/--limit flags
- [ ] tests: profile, migrate, sessions_compact, health, traces_filters
- [ ] docs sync: codebase-summary + CHANGELOG

## Success Criteria

- Existing single-config user upgrades transparently — no manual steps, token preserved.
- `goclaw --profile=staging agents list` works without writing config first (uses `staging` if exists, errs `4` if not).
- `goclaw sessions compact <key> --yes` returns exit 0 + JSON `{"ok":true}` (or server payload).
- `goclaw health` returns server health JSON without extra wrapping.
- `goclaw traces list --since=1h --agent=abc --status=error` returns filtered traces.
- ≥ 60% line coverage on new code.
- `go build ./... && go vet ./... && go test ./...` clean.

## Risk Assessment

| Risk | Mitigation |
|---|---|
| Profile migration corrupts existing config | Backup to `~/.goclaw/config.yaml.bak.<timestamp>` before move; refuse migration if profiles/default already exists |
| Token leakage between profiles | Distinct keyring keys `goclaw-cli:<profile>:token`; never read default key from non-default profile |
| `--profile=foo` typo silently uses wrong tenant | Resolve profile NAME → fail-fast with exit 3 NOT_FOUND if profile missing |
| `sessions compact` triggers without --yes in CI | Default destructive guard; `--yes` or `GOCLAW_YES=1` required |
| Health response leaks tenant info | Raw passthrough — server controls disclosure; CLI does not enrich |
| Traces filter regression on existing scripts | All new flags optional; default behavior unchanged |

## Security Considerations

- Token files mode 0600.
- Profile dirs mode 0700.
- Keyring namespace prevents cross-profile token reuse on shared dev box.
- `--profile=../foo` path traversal — sanitize name with regex `^[a-zA-Z0-9_-]{1,32}$`.

## Next Steps

After P3 ships:

- P4 (UX polish batch 1) — depends on stable profile mgmt for any future per-profile defaults.
- P5 (verify + filler batch) — independent.
- P6 — file upstream goclaw issues for deferred items.
