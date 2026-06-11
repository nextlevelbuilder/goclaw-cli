---
phase: 3
title: "CLI Credentials Commands"
status: pending
priority: P1
effort: "4h"
dependencies: [1]
---

# Phase 3: CLI Credentials Commands

## Overview

Align server-side CLI credential commands with the Runtime & Packages CLI Credentials tab.

## Requirements

- Functional: parse current envelopes and expose missing agent credential endpoints.
- Non-functional: avoid printing decrypted secret values unless an existing command explicitly requires `--show-secrets`.

## Architecture

Keep top-level `credentials` command assembled by existing `admin_credentials*.go` files. Add a focused `cmd/admin_credentials_agents.go` if needed to keep files small.

## Related Code Files

- Modify: `cmd/admin_credentials.go`
- Modify: `cmd/admin_credentials_grants.go`
- Modify: `cmd/admin_credentials_users.go`
- Create: `cmd/admin_credentials_agents.go`
- Modify: `cmd/credentials_contract_test.go`

## Implementation Steps

1. Fix envelope parsing:
   - `credentials list` reads `{items}`.
   - `credentials presets` reads `{presets}`.
   - `credentials agent-grants list` reads `{grants}`.
   - `credentials user-credentials list` reads `{user_credentials}`.
2. Add `credentials agent-credentials` subtree:
   - `list <credID>` -> `GET /v1/cli-credentials/{id}/agent-credentials`
   - `get <credID> <agentID>` -> `GET /v1/cli-credentials/{id}/agent-credentials/{agentID}`
   - `set <credID> <agentID> --body JSON` -> `PUT /v1/cli-credentials/{id}/agent-credentials/{agentID}`
   - `delete <credID> <agentID> --yes` -> `DELETE /v1/cli-credentials/{id}/agent-credentials/{agentID}`
3. Preserve raw map output for credential get/test/check-binary.
4. Confirm no command prints secret env values except existing `agent-grants env-reveal` requiring both `--yes` and `--show-secrets`.

## Success Criteria

- [ ] New credential contract tests pass.
- [ ] Missing `agent-credentials` route family is exposed.
- [ ] No broad credential UX rewrite.
- [ ] Secret-reveal guard remains explicit.

## Risk Assessment

Risk: confusing agent grants vs agent credentials. Mitigation: keep separate subtrees matching server endpoint names exactly.
