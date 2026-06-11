---
title: "Runtime Packages CLI Parity"
description: "Align goclaw-cli Runtime & Packages commands with digitopvn/goclaw dev contracts, including grouped package payloads and CLI credential tab routes."
status: pending
priority: P1
effort: 1d
branch: "codex/runtime-packages-cli-parity"
base: "dev"
tags: [runtime-packages, cli, contract, tdd]
blockedBy: []
blocks: []
created: "2026-06-11T17:15:32.053Z"
createdBy: "ck:plan --tdd"
source: skill
---

# Runtime Packages CLI Parity

## Overview

Bring `goclaw packages` and server-side `goclaw credentials` command surfaces back in sync with `/Volumes/GOON/www/digitop/goclaw` `dev`. The current CLI has the command names, but several commands still assume older response/request shapes.

## Scope

In scope:
- Align package list/install/uninstall/runtimes/deny-groups/github-releases with current server contracts.
- Keep existing `packages install <name> --runtime python|node` UX by translating to `pip:<name>` or `npm:<name>` before sending `{"package": ...}`.
- Preserve raw server payloads for JSON/YAML output; only table mode flattens grouped payloads for humans.
- Add missing `credentials agent-credentials` subtree for `/v1/cli-credentials/{id}/agent-credentials`.
- Fix existing CLI credential envelope handling where server returns `{items}`, `{presets}`, `{grants}`, `{user_credentials}`.
- Add focused httptest contract tests before implementation changes.
- Update README/CHANGELOG/docs summaries after verification.

Out of scope:
- Server changes in `/Volumes/GOON/www/digitop/goclaw`.
- New web UI work.
- Runtime package rollback/version-history features.
- Live gateway smoke requiring credentials; local httptest contract coverage is enough for this PR.

## Contract Sources

- Server packages handler: `/Volumes/GOON/www/digitop/goclaw/internal/http/packages.go`
- Server package updates handler: `/Volumes/GOON/www/digitop/goclaw/internal/http/packages_updates.go`
- Server installed package shape: `/Volumes/GOON/www/digitop/goclaw/internal/skills/package_lister.go`
- Server runtime shape: `/Volumes/GOON/www/digitop/goclaw/internal/skills/runtime_check.go`
- Web packages page: `/Volumes/GOON/www/digitop/goclaw/ui/web/src/pages/packages/packages-page.tsx`
- Server CLI credentials handler: `/Volumes/GOON/www/digitop/goclaw/internal/http/secure_cli.go`
- Web CLI credential hooks: `/Volumes/GOON/www/digitop/goclaw/ui/web/src/pages/cli-credentials/hooks/`

## Phase Overview

| Phase | Name | Status |
|-------|------|--------|
| 1 | [Contract Lock](./phase-01-contract-lock.md) | Pending |
| 2 | [Package Commands](./phase-02-package-commands.md) | Pending |
| 3 | [CLI Credentials Commands](./phase-03-cli-credentials-commands.md) | Pending |
| 4 | [Docs Verification Ship](./phase-04-docs-verification-ship.md) | Pending |

## TDD Strategy

1. Replace stale package/credential tests with server-shaped fixtures that fail on current CLI.
2. Implement the smallest command changes to satisfy those fixtures.
3. Run focused package tests, then full `go test ./...`, `go build ./...`, and `go vet ./...`.
4. Review diff against server source contracts before beta PR.

## Cross-Plan Notes

- Older package/credential plans are completed or stale broad-backlog artifacts.
- This plan does not block trace issue plans.
- Relevant stale in-progress plans (`P5`, `P6`) already have their implementation landed in current `dev`; do not mutate them except optional status cleanup after ship.

## Acceptance Criteria

- `goclaw packages list` prints grouped server payload correctly in table mode and preserves full map in JSON/YAML mode.
- `goclaw packages install/uninstall` send `package` body and support `github:`, `pip:`, `npm:`, `apk:`/bare system names.
- `goclaw packages runtimes` renders `ready` plus each runtime from `{runtimes,ready}`.
- `goclaw packages github-releases --repo owner/repo [--limit N]` calls the query contract and prints `{releases}` correctly.
- `goclaw packages updates apply/apply-all` still supports `github|pip|npm|apk:<name>` specs and preserves partial-failure non-zero behavior.
- `goclaw credentials list/presets/agent-grants/user-credentials` parse current server envelopes.
- New `goclaw credentials agent-credentials list|get|set|delete` maps to current server routes.
- All destructive commands keep `--yes` or interactive confirmation behavior where existing convention requires it.
- Full verification passes or blocker is reported with exact command output.

## Unresolved Questions

None.
