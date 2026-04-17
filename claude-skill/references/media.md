# Media

## When to use

User wants to upload or download media files (images, audio, video, documents) to/from the GoClaw server. These are distinct from workspace `storage` (which is generic files) and from skill uploads.

## Commands in scope

- `goclaw media upload <path>` — upload local file (source: `cmd/admin_media.go`)
- `goclaw media download <id> [-f <file>]` — download by media ID

## Verified flags

### `media upload`
| Flag | Type | Purpose |
| --- | --- | --- |
| `--type <t>` | string | `image`/`audio`/`video`/`document` |
| `--description <d>` | string | Optional label |

### `media download`
| Flag | Type | Purpose |
| --- | --- | --- |
| `-f, --output <file>` | string | Output file (default stdout) |

## JSON output

- ⚠️ `media upload` — success text with ID; parse from stdout or rely on exit code
- ⚠️ `media download` — raw binary; always use `-f`

## Destructive ops

None — upload/download are non-destructive (though upload counts against storage quota).

## Common patterns

### Example 1: upload image
```bash
goclaw media upload ./screenshot.png --type image --description "Bug repro" --output json
# → returns media ID for later download
```

### Example 2: download by ID
```bash
goclaw media download <media-id> -f /tmp/out.png
```

### Example 3: upload + immediately use as context
```bash
# Upload returns ID; feed into agent's context somehow (via memory store or chat message)
MEDIA_ID=$(goclaw media upload ./doc.pdf --type document --output json | jq -r '.id')
goclaw memory store <agent-id> attachments/doc.pdf --content "@./doc.pdf"
```

## Edge cases & gotchas

- **Large files** hit Bash tool timeout (120s). For files > 10 MB, recommend user runs CLI directly in their terminal.
- **MIME type** inferred server-side from extension; `--type` is a semantic bucket (image/audio/etc.), not MIME.
- **Download binary → stdout corruption:** Claude MUST always pass `-f` to redirect. Never `goclaw media download` without output redirection.
- **Upload quota:** counts against tenant storage quota. Check `goclaw storage size --output json` for remaining.
- **Media vs storage:** media = semantic assets referenced by agents/chats; storage = generic file browser. Don't confuse paths.

## Cross-refs

- Workspace file management: [data-movement.md](data-movement.md) — `storage`
- Media tools for agents (read_image, read_document, etc.): [providers-skills-tools.md](providers-skills-tools.md)
