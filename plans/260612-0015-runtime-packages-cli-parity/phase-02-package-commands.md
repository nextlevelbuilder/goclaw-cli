---
phase: 2
title: "Package Commands"
status: pending
priority: P1
effort: "4h"
dependencies: [1]
---

# Phase 2: Package Commands

## Overview

Fix `goclaw packages` command contracts while preserving existing automation ergonomics.

## Requirements

- Functional: align request bodies, query params, and response renderers for package management.
- Non-functional: preserve JSON/YAML full payloads and central error handling.

## Architecture

Keep `cmd/packages.go` as the command owner if it stays under the file-size guideline after edits; otherwise extract small render/query helpers into a package-specific helper file. Do not alter `internal/client`.

## Related Code Files

- Modify: `cmd/packages.go`
- Modify: `cmd/packages_updates.go`
- Modify: `cmd/packages_contract_test.go`

## Implementation Steps

1. Add helpers:
   - `normalizePackageSpec(name, runtime)` maps `--runtime python` to `pip:name`, `--runtime node` to `npm:name`, and leaves explicit `source:name` unchanged.
   - `queryPath(path, url.Values)` for simple query building if no existing helper fits.
2. `packages list`:
   - non-table: `printer.Print(unmarshalMap(data))`.
   - table: flatten `system`, `pip`, `npm`, and `github` groups with a `SOURCE` column.
3. `packages install/uninstall`:
   - send `{"package": normalizedSpec}`.
   - keep uninstall confirmation.
4. `packages runtimes`:
   - non-table: print map.
   - table: render `READY`, runtime `NAME`, `AVAILABLE`, `VERSION`.
5. `packages deny-groups`:
   - accept both current `{groups:[...]}` and older array fixture for backward compatibility.
6. `packages github-releases`:
   - add required `--repo owner/repo`, optional `--limit`.
   - call `/v1/packages/github-releases?repo=...&limit=...`.
   - render `{releases}` in table mode without dropping JSON fields in machine modes.
7. `packages updates apply-all`:
   - keep `--packages` CSV.
   - optionally accept positional specs too if low-risk; when both are present, merge.
   - preserve non-zero on non-empty `failed[]` unless `--allow-partial`.

## Success Criteria

- [ ] New package contract tests pass.
- [ ] Existing package update partial-failure test still passes.
- [ ] Commands stay Cobra-pattern consistent.
- [ ] No new server assumptions beyond source-verified contracts.

## Risk Assessment

Risk: changing output table breaks human muscle memory. Mitigation: only table changes; JSON/YAML payloads remain server-shaped for automation.
