---
phase: 1
title: "Contract Lock"
status: pending
priority: P1
effort: "2h"
dependencies: []
---

# Phase 1: Contract Lock

## Overview

Lock exact server response/request shapes before changing CLI code. This phase prevents passing tests that still encode old flat-array assumptions.

## Requirements

- Functional: capture package and CLI credential contract facts into failing httptest fixtures.
- Non-functional: no production code changes before failing tests exist.

## Architecture

CLI remains thin Cobra over `internal/client.HTTPClient`. Tests assert path, method, query string, request body, and printed payload handling where practical.

## Related Code Files

- Modify: `cmd/phase5_test.go`
- Modify: `cmd/super_admin_parity_test.go`
- Create: `cmd/packages_contract_test.go`
- Create: `cmd/credentials_contract_test.go`

## Implementation Steps

1. Add server-shaped package fixtures:
   - `GET /v1/packages` returns `{system,pip,npm,github}`.
   - `POST /v1/packages/install` and uninstall require body key `package`.
   - `GET /v1/packages/runtimes` returns `{runtimes,ready}`.
   - `GET /v1/packages/github-releases?repo=cli/cli&limit=10` returns `{releases}`.
2. Add credentials fixtures:
   - list returns `{items}`.
   - presets returns `{presets}`.
   - grants returns `{grants}`.
   - user credentials returns `{user_credentials}`.
   - agent credentials returns `{agent_credentials}`.
3. Verify focused tests fail for the current implementation before production edits.

## Success Criteria

- [ ] Focused new tests fail for the known contract mismatches.
- [ ] Test names describe stable server contracts, not plan IDs.
- [ ] No implementation files changed in this phase except tests.

## Risk Assessment

Risk: overfitting fixtures. Mitigation: fixtures mirror current server source and web hooks, not inferred docs.
