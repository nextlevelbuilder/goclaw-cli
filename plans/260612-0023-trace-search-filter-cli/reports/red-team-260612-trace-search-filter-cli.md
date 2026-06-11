# Red Team: Trace Search Filter CLI

## Verdict

PASS with three accepted guardrails.

## Findings

### Accepted: `--agent` collision can break scripts

Risk: server PR #155 uses query param `agent` for display-name/key search, but CLI already uses `--agent` for `agent_id`.

Resolution: keep `--agent -> agent_id`; add `--agent-query -> agent`.

### Accepted: false and zero filters can be silently dropped

Risk: current code forwards ints only when `> 0`, which would drop explicit filters like `--has-tool-calls=false` and `--min-tool-calls=0`.

Resolution: new numeric and boolean-like filters must use `cmd.Flags().Changed(name)`.

### Accepted: client wildcard escaping can double-escape server behavior

Risk: PR #155 explicitly added wildcard escaping in server stores. Client-side escaping would change search semantics.

Resolution: CLI only URL-encodes query values through `url.Values`; server owns wildcard escaping.

## Rejected Findings

- Add local RFC3339/date validation: rejected. Server already returns structured invalid-request errors for `from` and `to`; adding client validation duplicates logic and can drift.
- Add a new trace search subcommand: rejected. Existing server contract is list filtering; no extra subcommand needed.

## Unresolved Questions

None.
