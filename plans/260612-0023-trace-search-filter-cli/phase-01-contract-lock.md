---
phase: 1
title: Contract Lock
status: completed
priority: P1
effort: 1h
dependencies: []
---

# Phase 1: Contract Lock

## Overview

Lock the merged server contract from `digitopvn/goclaw#155` before implementation. The output of this phase is a testable list of query mappings and explicit exclusions.

## Requirements

- Functional: map every new server list filter to one CLI flag.
- Non-functional: preserve existing `traces list` automation contract and avoid speculative server behavior.

## Architecture

`traces list` remains a one-shot `GET /v1/traces` command. The CLI only builds `url.Values`; all matching semantics, wildcard escaping, date parsing, and tenant scoping stay server-side.

## Related Code Files

- Modify: `cmd/traces_contract_test.go`
- Read: `cmd/traces.go`
- Read: `cmd/traces_list_test.go`
- Read: upstream `digitopvn/goclaw#155`

## Implementation Steps

1. Verify the server PR is merged and capture exact query names.
2. Add or update a contract test helper that resets all trace list flags, including new ones.
3. Confirm existing replay-absence coverage stays in place.

## Todo List

- [x] Capture exact query parameter names from server PR #155.
- [x] Update `resetTracesListFlags` for all new flags.
- [x] Keep `TestTracesReplayCommandAbsent` green.

## Success Criteria

- [x] Contract test names all new query params.
- [x] No plan or test implies `traces replay` exists.
- [x] No unresolved scope questions remain.

## Risk Assessment

Misnaming `agent` as `agent_id` would silently change semantics. The test must assert both `--agent` and `--agent-query` mappings to keep ID and query search distinct.

## Security Considerations

No credentials or live traces are used. Tests use synthetic `httptest` requests and server-shaped fixtures only.

## Next Steps

Proceed to Phase 2 only after focused contract tests fail for missing flags.
