---
phase: 3
title: "Evolution Skill Apply"
status: complete
priority: P1
effort: "2h"
dependencies: [1]
---

# Phase 3: Evolution Skill Apply

## Overview

Add a clear CLI wrapper for approving `skill_add` evolution suggestions. The server creates the managed skill when the suggestion is approved.

## Requirements

- Functional: `goclaw agents evolution skill apply <agent-id> <suggestion-id>` sends `status=approved`.
- Functional: command refuses suggestions whose `suggestion_type` is not `skill_add`.
- Functional: optional `--skill-draft <literal|@file>` sends `skill_draft` override.
- Functional: response is printed as structured output.
- Functional: missing/invalid IDs are left to server validation; CLI only enforces arg count.
- Non-functional: no new generic evolution apply surface.
- Non-functional: preserve existing `agents evolution update` UX while fixing stale payload mapping.

## Architecture

Command tree:

```text
agents
  evolution
    metrics <id>
    suggestions <id>
    update <id> <suggestionID> --action accept|reject
    skill
      apply <id> <suggestionID> [--skill-draft <literal|@file>]
```

Skill apply payload:

```json
{
  "status": "approved",
  "skill_draft": "optional override content"
}
```

Existing update compatibility:
- Current CLI command should map `--action=accept` to `status=approved`.
- Current CLI command should map `--action=reject` to `status=rejected`.
- This is required: current CLI sends `{"action": ...}` but server reads `status`.
- Fix it in the same file with tests because it shares the exact server route.
- Use `url.PathEscape` for agent and suggestion ID path segments in both the new skill command and the existing update command.

## Related Code Files

- Modify: `cmd/agents_evolution.go`
- Test: extend `cmd/agents_lifecycle_test.go` or add `cmd/p5_fillers_test.go`
- Read: `/Volumes/GOON/www/nlb/goclaw/internal/http/evolution_handlers.go`
- Read: `/Volumes/GOON/www/nlb/goclaw/internal/http/evolution_skill_apply.go`

## Implementation Steps

1. Add `agentsEvolutionSkillCmd` subgroup.
2. Add `agentsEvolutionSkillApplyCmd`.
3. Add `--skill-draft` flag.
4. Use existing `readContent()` for literal or `@file` draft content.
5. PATCH `/v1/agents/{id}/evolution/suggestions/{suggestionID}` with escaped path segments and `status=approved`.
6. Preflight pending suggestions and verify the selected suggestion has `suggestion_type=skill_add`.
7. Print `unmarshalMap(data)` so JSON/YAML users get server action fields.
8. Change `agentsEvolutionUpdateCmd` body from `{"action": action}` to `{"status": mappedStatus}`.
   - `accept` -> `approved`
   - `reject` -> `rejected`
   - Keep CLI flag as `--action` for backward compatibility.
   - Update success text to avoid `acceptd`; use approved/rejected wording.
9. Add tests:
   - skill apply sends `status=approved`.
   - `--skill-draft=@file` includes exact content.
   - non-`skill_add` suggestion is refused before approval PATCH.
   - update accept maps to `status=approved`.
   - update reject maps to `status=rejected`.
   - evolution PATCH routes escape IDs.
   - invalid update action still rejects before network.

## Success Criteria

- [x] `agents evolution skill apply` appears in command tree.
- [x] PATCH payload matches server contract.
- [x] Non-`skill_add` suggestions are refused before approval.
- [x] Optional draft override works from `@file`.
- [x] Existing update command remains backward-compatible at CLI flag level and sends server-compatible payload.
- [x] Tests cover both new wrapper and any payload mapping fix.

## Risk Assessment

| Risk | Mitigation |
|---|---|
| Applying non-skill suggestions | Server enforces suggestion type; CLI command name makes intent explicit. |
| Stale payload in existing update | Validate and fix mapping in same module if needed. |
| Draft content leaking in process list | `--skill-draft @file` documented as preferred for large/sensitive drafts. |
| Over-abstraction | Keep command wrapper direct; no generic suggestion apply framework. |
