# GoClaw CLI Skill for Claude Code

Lets the Claude Code agent drive [goclaw](https://github.com/nextlevelbuilder/goclaw-cli) — the CLI for GoClaw AI agent gateway servers — via natural-language prompts. Execute shell commands on remote servers, manage AI agents, approve execution requests, inspect chat sessions, and administer multi-tenant infrastructure, all without leaving Claude Code.

## Requirements

- `goclaw` binary ≥ 0.3.0 in `$PATH` ([install](https://github.com/nextlevelbuilder/goclaw-cli/releases))
- Claude Code CLI installed (`~/.claude/` directory)
- Authenticated: run `goclaw auth login` before using the skill
- macOS or Linux (Windows via WSL)

## Install

### Recommended: release tarball + SHA256

```bash
RELEASE_URL="https://github.com/nextlevelbuilder/goclaw-cli/releases/download/skill-v0.1.0"
curl -fsSL "$RELEASE_URL/goclaw-skill.tar.gz" -o /tmp/goclaw-skill.tar.gz
curl -fsSL "$RELEASE_URL/goclaw-skill.sha256" -o /tmp/goclaw-skill.sha256
(cd /tmp && sha256sum -c goclaw-skill.sha256)  # aborts on mismatch
tar xzf /tmp/goclaw-skill.tar.gz -C /tmp
/tmp/claude-skill/install.sh
```

### From git clone

```bash
git clone https://github.com/nextlevelbuilder/goclaw-cli.git
cd goclaw-cli/claude-skill
./install.sh
```

## Usage examples

Once installed, just ask Claude:

- *"list agents on goclaw"* → runs `goclaw agents list --output json`
- *"run `uname -a` on goclaw server"* → runs `goclaw tools invoke exec --param command="uname -a"`
- *"what's my usage this week"* → runs `goclaw usage summary --output json`

Destructive operations always prompt you first. Claude will ask *"Confirm delete agent xyz?"* before running `goclaw agents delete xyz --yes`.

## Permission modes

The installer offers three modes:

| Mode | Rule | When to use |
|------|------|-------------|
| 1 | `Bash(goclaw:*)` | Trusted personal machine, full-auto |
| 2 | ~20 readonly rules | Shared / production machine (default) |
| 3 | Manual JSON snippet | You want to edit settings.json yourself |

Default in pipe mode (`curl | bash`): Mode 3 (safest).

## What's NOT supported

- Interactive chat REPL (`goclaw chat` without `-m`) — use single-shot
- `logs tail` streaming — Claude will suggest polling instead
- `auth pair` device pairing — run manually, skill refuses

## Uninstall

```bash
rm -rf ~/.claude/skills/goclaw
# Then manually remove goclaw permissions from ~/.claude/settings.json
```

## Compatibility

Tested against `goclaw` CLI ≥ 0.3.0. `check-drift.sh` catches flag drift between skill and CLI.

## License

MIT — see [LICENSE](LICENSE).
