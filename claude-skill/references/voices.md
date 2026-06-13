# Voice catalog

## When to use

User wants to list available voices for TTS (text-to-speech) agents, or refresh the voice catalog from providers (to sync new voices, models). Not for *invoking* TTS (see [exec-workflow.md](exec-workflow.md) for `tools invoke tts`).

## Commands in scope

- `goclaw voices list` — list available voices from all configured TTS providers (source: `cmd/voices.go`)
- `goclaw voices refresh` — refresh catalog from providers (admin-only, async)

## Verified flags

### `voices list`
No flags beyond globals.

### `voices refresh`
No flags; admin-only.

## JSON output

- ✅ `voices list` — JSON array of voice objects
- ✅ `voices refresh` — JSON response with refresh status/timestamp

## Destructive ops

None — read-only or admin refresh.

## Common patterns

### Example 1: list all voices
```bash
goclaw voices list --output json
# → [
#   {"id": "en-US-Neural2-A", "name": "en-US Neural 2A", "lang": "en-US", "provider": "google"},
#   {"id": "elevenlabs_bella", "name": "Bella", "lang": "multi", "provider": "elevenlabs"},
#   ...
# ]
```

### Example 2: filter voices (client-side)
```bash
goclaw voices list --output json | jq '.[] | select(.lang | startswith("en-US"))'
```

### Example 3: refresh voice catalog (admin)
```bash
goclaw voices refresh --output json
# → {"refreshed_at": "2026-06-12T14:30:00Z", "voices_count": 250}
```

## Edge cases & gotchas

- **Voice ID:** used in TTS tool invocation. Format varies by provider (Google: `en-US-Neural2-A`, ElevenLabs: `en_us_bella`).
- **Languages:** most voices are language-specific. Multi-lingual voices rare.
- **Voice speed/pitch:** not exposed in CLI — set in TTS tool config or agent settings.
- **Refresh latency:** providers may take seconds to respond. Refresh is async; new voices appear in `list` after refresh completes.
- **Provider status:** if a TTS provider is down, `voices list` may return stale data. Check `goclaw status` for provider health.
- **Voice deprecation:** providers retire old voices. `refresh` updates the catalog; old voice IDs may fail in TTS invocations.

## Cross-refs

- TTS invocation (actually generate speech): [exec-workflow.md](exec-workflow.md) — `tools invoke tts`
- TTS status check: [admin-system.md](admin-system.md) — `tts status`
- Provider management: [providers-skills-tools.md](providers-skills-tools.md)
