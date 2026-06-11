---
phase: 2
title: TDD Implementation
status: completed
priority: P1
effort: 2h
dependencies:
  - 1
---

# Phase 2: TDD Implementation

## Overview

Add focused failing tests for the new trace list filters, then implement the smallest CLI change in `cmd/traces.go`.

## Requirements

- Functional: `traces list` forwards all PR #155 filter flags when supplied.
- Functional: zero-valued numeric filters can be forwarded when the user explicitly sets them.
- Functional: `--has-tool-calls=false` is forwarded, not dropped.
- Non-functional: no output shape or command behavior changes outside list query construction.

## Architecture

The command keeps using `url.Values`. A small helper may copy changed string/int/bool-like string flags to query params if it reduces repetitive flag handling without creating a generic framework.

## Related Code Files

- Modify: `cmd/traces.go`
- Modify: `cmd/traces_list_test.go`
- Modify: `cmd/traces_contract_test.go`

## Implementation Steps

1. Extend `TestTracesList_ServerFilters` to include all new flags and expected query names.
2. Add a focused test for false/zero explicit forwarding:
   `--has-tool-calls=false`, `--min-tool-calls=0`, `--min-input-tokens=0`.
3. Run focused tests and confirm failure before implementation.
4. Add flags to `tracesListCmd`:
   `--q`, `--agent-query`, `--channel-query`, `--from`, `--to`,
   `--min-input-tokens`, `--max-input-tokens`,
   `--min-output-tokens`, `--max-output-tokens`,
   `--min-tool-calls`, `--max-tool-calls`,
   `--tool-name`, `--has-tool-calls`.
5. Map flags to server params:
   `q -> q`, `agent-query -> agent`, `channel-query -> channel_query`,
   ranges to snake_case names, `has-tool-calls -> has_tool_calls`.
6. Run focused tests again and fix only failures in this scope.

## Todo List

- [x] Add red tests for new filter forwarding.
- [x] Add explicit false/zero forwarding test.
- [x] Implement flag registration and query mapping.
- [x] Keep existing basic filter test green.

## Success Criteria

- [x] Focused trace list tests pass.
- [x] No changes to `traces follow`, `get`, `export`, or `timeline`.
- [x] `cmd/traces.go` remains compilable and has no duplicated large blocks.

## Risk Assessment

The main regression risk is accidentally changing `--agent` from agent ID to label search. Keep `--agent` as `agent_id` for backward compatibility and add `--agent-query` for the server `agent` query.

## Security Considerations

User-provided filter strings are only URL query values encoded through `url.Values`. No shell execution, path usage, or local file reads are introduced.

## Next Steps

After focused tests pass, update README and docs in Phase 3.
