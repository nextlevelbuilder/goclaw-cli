# Channels & messaging

## When to use

User wants to connect an agent to a messaging platform (Telegram, Discord, Slack, Zalo, Feishu, WhatsApp), manage contacts/writers, or inspect pending messages.

## Commands in scope

- `goclaw channels instances list/get/create/update/delete` — channel instance CRUD (source: `channels.go`)
- `goclaw channels contacts list/get/create/delete` — contacts per channel (source: `channels_contacts.go`)
- `goclaw channels pending list` — pending messages on channel (source: `channels_pending.go`)
- `goclaw channels writers list/add/remove` — agent authors for a channel (source: `channels_writers.go`)
- `goclaw contacts list/get/create/update/delete/verify` — global contact book (source: `contacts.go`)
- `goclaw pending-messages list/create/send/delete` — pending message queue (source: `pending_messages.go`)

## Verified flags

### `channels instances list`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--type <t>` | string | Filter: `telegram`/`discord`/`slack`/`zalo-oa`/`zalo-personal`/`feishu`/`whatsapp` |

### `channels instances create`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--name` (REQUIRED) | string | Instance name |
| `--type` (REQUIRED) | string | Channel type |
| `--agent` (REQUIRED) | string | Agent ID |

### `channels instances update`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--name <n>` | string | Rename |
| `--enabled <bool>` | bool | Enable/disable |

### `channels writers add/remove`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--agent <id>` | string | Writer agent ID |
| `--user <id>` | string | Or writer user ID |

### `contacts` lookup helper
| Flag | Type | Purpose |
| --- | --- | --- |
| `--identifier <v>` | string | Phone number or email (for `contacts resolve`) |

**Note:** contact create/update field names (name/email/phone) pass through `buildBody` — verify with `goclaw contacts create --help` for current flag names.

## JSON output

- ✅ `channels instances list/get`, `channels contacts list/get`, `channels pending list`, `channels writers list`, `contacts list/get`, `pending-messages list` — JSON
- ⚠️ Everything else — success text

## Destructive ops

| Command | Confirm |
| --- | --- |
| `channels instances delete` | YES (disconnects platform) |
| `channels contacts delete` | YES |
| `channels writers remove` | YES |
| `contacts delete` | YES |
| `pending-messages delete` | YES |
| `pending-messages send` | YES (IRREVERSIBLE user-facing message send) |

## Common patterns

### Example 1: list all Telegram channels
```bash
goclaw channels instances list --type telegram --output json
```

### Example 2: connect agent to Discord
```bash
goclaw channels instances create \
  --name "support-discord" \
  --type discord \
  --agent <agent-id> \
  --output json
```

### Example 3: add writer to channel
```bash
goclaw channels writers add --agent <agent-id> # add agent as writer
# or user:
goclaw channels writers add --user <user-id>
```

### Example 4: send pending message (always confirm)
```bash
# Claude: echo message content back to user for confirmation
goclaw pending-messages send <message-id> --yes
```

### Example 5: contact verify
```bash
goclaw contacts verify <contact-id> --output json
```

## Edge cases & gotchas

- **Channel credentials** (bot tokens for Telegram/Discord) stored server-side, not in CLI. `create` may require follow-up `update` via dashboard to set API keys.
- **Channel instance vs contact:** instance = platform link (one per channel+tenant); contact = person on the platform. Writers = agents/users authorized to send from the instance.
- **`pending-messages send` is USER-facing** — the message is delivered to external user once sent. No undo. ALWAYS confirm content with user first.
- **Zalo variants:** `zalo-oa` = official account (business), `zalo-personal` = personal user flow. Different auth model.
- **`contacts verify`** triggers OTP/email flow — user must complete verification out-of-band.

## Cross-refs

- Configure TTS for channels: [admin-system.md](admin-system.md) — `tts status`
- Agent that owns the channel: [agents-core.md](agents-core.md)
