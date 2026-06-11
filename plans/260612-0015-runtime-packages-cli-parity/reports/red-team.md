# Plan Red-Team Report

## Verdict

PASS after accepted safeguards below.

## Findings

| Severity | Finding | Decision |
|---|---|---|
| High | CLI credential work can leak env values if list/get renders decrypted fields blindly. | Accept: tests/review must verify only existing explicit `env-reveal --yes --show-secrets` reveals secrets. |
| High | `packages install/uninstall` could silently keep sending stale `name/runtime` fields if tests only check status code. | Accept: body assertions must verify exact `package` key. |
| Medium | `packages github-releases` without `--repo` is unusable and can waste GitHub quota with invalid requests. | Accept: require `--repo` client-side before HTTP call. |
| Medium | Flattened table render can drop important fields and mislead automation if output auto-detection changes. | Accept: only flatten when `cfg.OutputFormat == "table"`; JSON/YAML prints raw map. |
| Medium | Agent grants and agent credentials names are easy to confuse. | Accept: expose server endpoint names exactly; no aliases in this PR. |

## Required Plan Edits

- Include explicit no-secret-leak review gate.
- Include exact request-body tests.
- Include `--repo` client validation for GitHub releases.

## Unresolved Questions

None.
