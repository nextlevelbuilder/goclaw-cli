# OAuth providers (ChatGPT, OpenAI)

## When to use

User wants to authenticate with OAuth providers (ChatGPT, OpenAI) to delegate LLM inference or share account access. CLI guides the browser OAuth flow, persists tokens, checks status, or revokes access.

## Commands in scope

- `goclaw oauth start` — start browser OAuth flow (prints auth URL + returns auth_id) (source: `cmd/oauth.go`)
- `goclaw oauth callback <code>` — complete OAuth flow (submit auth code from browser redirect)
- `goclaw oauth status` — check current OAuth provider authentication status
- `goclaw oauth quota` — show OAuth provider quota/rate-limit usage
- `goclaw oauth logout` — revoke OAuth provider token (requires `--yes`)

## Verified flags

### `oauth start`
No flags (flow is automatic).

### `oauth callback`
| Flag | Type | Purpose |
| --- | --- | --- |
| `<code>` | string | Authorization code from browser redirect |

### `oauth status`
No flags.

### `oauth quota`
No flags.

### `oauth logout`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--yes` | bool | **REQUIRED** — confirm revocation |

## JSON output

- ✅ `oauth status/quota` — JSON
- ⚠️ `oauth start` — URL to stderr, `auth_id` to stdout (non-JSON by design)
- ⚠️ `oauth callback/logout` — success text

## Destructive ops

| Command | Confirm |
| --- | --- |
| `oauth logout` | YES (revokes token; agents lose access) |

## Common patterns

### Example 1: start OAuth flow
```bash
goclaw oauth start
# Output to stderr (user must open in browser):
#   https://accounts.openai.com/authorize?code=...&state=abc123
# Output to stdout (use in callback):
#   abc123
```

### Example 2: complete OAuth flow (in the same script)
```bash
AUTH_ID=$(goclaw oauth start 2>/dev/null)
# user visits URL from stderr, authorizes, browser redirects with code=XYZ
CODE="XYZ"  # extracted from redirect URI
goclaw oauth callback $CODE
```

### Example 3: check OAuth status
```bash
goclaw oauth status --output json
# → {
#   "authenticated": true,
#   "provider": "openai",
#   "email": "user@example.com",
#   "scopes": ["chat.write"],
#   "token_expiry": "2026-12-31T23:59:59Z"
# }
```

### Example 4: check quota
```bash
goclaw oauth quota --output json
# → {
#   "tokens_used_this_period": 500000,
#   "rate_limit_per_minute": 100,
#   "next_reset": "2026-06-13T00:00:00Z"
# }
```

### Example 5: revoke OAuth token
```bash
# Claude: confirm with user — agents lose access to provider
goclaw oauth logout --yes
```

## Edge cases & gotchas

- **Browser redirect:** `oauth start` prints URL to stderr (not stdout) so Claude can capture `auth_id`. User must open URL in browser manually.
- **Auth code TTL:** code from browser redirect expires within minutes. Complete callback quickly.
- **Token refresh:** server auto-refreshes expired tokens. Callback may fail if provider returns new token.
- **Provider scope:** agents access only permitted scopes (read chats, write completions, etc.). Mismatch = 403 errors.
- **Multiple accounts:** CLI tracks one OAuth session per provider. Logging out revokes that session; starting new flow for different user.
- **Quota decay:** quota resets per provider's billing period (not per calendar month).

## Cross-refs

- Provider management (LLM providers): [providers-skills-tools.md](providers-skills-tools.md)
- Agent LLM configuration: [agents-core.md](agents-core.md)
- Usage analytics: [monitoring-ops.md](monitoring-ops.md)
