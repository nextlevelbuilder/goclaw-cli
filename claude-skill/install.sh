#!/usr/bin/env bash
# install.sh — GoClaw Claude Code Skill installer.
# Copies skill files to ~/.claude/skills/goclaw/ and merges permissions
# into ~/.claude/settings.json. Idempotent. Safe under `curl | bash`.
set -euo pipefail

# ------------- Windows abort -------------
case "$(uname -s)" in
  MINGW*|MSYS*|CYGWIN*)
    echo "ERROR: Windows not supported in v1. Use WSL or install manually." >&2
    exit 1 ;;
esac

# ------------- args -------------
MODE=""
DRY_RUN=0
FORCE=0

usage() {
  cat <<'EOF'
GoClaw Claude Code Skill installer

Usage: install.sh [options]

Options:
  --mode 1|2|3    Permission mode (skips interactive prompt):
                    1 = full wildcard  Bash(goclaw:*)   (trust this machine)
                    2 = readonly verbs (~20 rules, SAFEST default)
                    3 = no patching (print JSON snippet only)
  --dry-run       Print actions, do not modify files
  --force         Overwrite existing skill dir without prompt
  -h, --help      Show this help

Without --mode: interactive via /dev/tty. Under `curl|bash` (no TTY): auto-picks Mode 3.

Env vars:
  CLAUDE_HOME     Override ~/.claude location (default: $HOME/.claude)
EOF
}

while [[ $# -gt 0 ]]; do
  case $1 in
    --mode) MODE="${2:-}"; shift 2 ;;
    --dry-run) DRY_RUN=1; shift ;;
    --force) FORCE=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown arg: $1" >&2; usage >&2; exit 1 ;;
  esac
done

# ------------- locations -------------
CLAUDE_HOME="${CLAUDE_HOME:-$HOME/.claude}"
SETTINGS="$CLAUDE_HOME/settings.json"
SKILL_DIR="$CLAUDE_HOME/skills/goclaw"
SRC_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ------------- prereqs -------------
if ! command -v goclaw >/dev/null 2>&1; then
  echo "WARNING: 'goclaw' not in PATH. Skill will be installed, but commands will fail until you install the binary." >&2
  echo "         Download: https://github.com/nextlevelbuilder/goclaw-cli/releases" >&2
fi

if [[ ! -d "$CLAUDE_HOME" ]]; then
  echo "ERROR: $CLAUDE_HOME not found. Install Claude Code first." >&2
  exit 1
fi

# ------------- python3 detect -------------
PY=""
if [[ -x "$CLAUDE_HOME/skills/.venv/bin/python3" ]]; then
  PY="$CLAUDE_HOME/skills/.venv/bin/python3"
elif command -v python3 >/dev/null 2>&1; then
  PY="$(command -v python3)"
fi
if [[ -n "$PY" ]]; then
  if ! "$PY" -c 'import sys; sys.exit(0 if sys.version_info[0]==3 else 1)' 2>/dev/null; then
    echo "WARNING: python3 at $PY is not Python 3.x — permission patching will be skipped." >&2
    PY=""
  fi
fi

# ------------- interactive mode selection -------------
if [[ -z "$MODE" ]]; then
  if [[ -t 0 || -r /dev/tty ]]; then
    # We have a TTY to read from.
    echo "Choose permission mode:"
    echo "  1) Full wildcard Bash(goclaw:*) — trust this machine fully"
    echo "  2) Readonly verbs only (~20 rules, SAFEST)"
    echo "  3) No patching — print JSON snippet"
    read -r -p "Select [3]: " MODE < /dev/tty || MODE="3"
    MODE="${MODE:-3}"
  else
    # Piped (curl|bash) — default Mode 3.
    echo "No TTY detected (piped install). Defaulting to Mode 3 (no patching)."
    MODE="3"
  fi
fi

case "$MODE" in
  1|2|3) ;;
  *) echo "ERROR: invalid --mode $MODE. Use 1, 2, or 3." >&2; exit 1 ;;
esac

# ------------- confirm Mode 1 -------------
if [[ "$MODE" == "1" ]]; then
  if [[ -r /dev/tty ]]; then
    echo ""
    echo "WARNING: Mode 1 grants Claude Code permission to run ANY goclaw command,"
    echo "         including destructive operations (delete, revoke, clear)."
    read -r -p "Continue? [y/N] " CONFIRM < /dev/tty || CONFIRM=""
    case "$CONFIRM" in y|Y|yes|YES) ;; *) echo "Aborted."; exit 0 ;; esac
  else
    echo "ERROR: Mode 1 requires TTY confirmation. Re-run in terminal or pass --mode 2|3." >&2
    exit 1
  fi
fi

# ------------- skill overwrite gate -------------
if [[ -d "$SKILL_DIR" && "$FORCE" != "1" ]]; then
  if [[ -r /dev/tty ]]; then
    read -r -p "Existing skill at $SKILL_DIR. Overwrite? [y/N] " OW < /dev/tty || OW=""
    case "$OW" in y|Y|yes|YES) ;; *) echo "Aborted."; exit 0 ;; esac
  else
    echo "ERROR: $SKILL_DIR exists. Re-run with --force to overwrite." >&2
    exit 1
  fi
fi

echo ""
echo "Installing GoClaw skill to: $SKILL_DIR"
echo "Permission mode: $MODE  |  dry-run: $DRY_RUN"

# ------------- copy skill files -------------
if [[ "$DRY_RUN" == "1" ]]; then
  echo "[dry-run] mkdir -p $SKILL_DIR/references"
  echo "[dry-run] cp SKILL.md, README.md, LICENSE, check-drift.sh, references/*.md, .verified-commands.txt"
else
  mkdir -p "$SKILL_DIR/references"
  cp "$SRC_DIR/SKILL.md" "$SKILL_DIR/"
  cp "$SRC_DIR/README.md" "$SKILL_DIR/"
  cp "$SRC_DIR/LICENSE" "$SKILL_DIR/"
  cp "$SRC_DIR/check-drift.sh" "$SKILL_DIR/"
  [[ -f "$SRC_DIR/.verified-commands.txt" ]] && cp "$SRC_DIR/.verified-commands.txt" "$SKILL_DIR/"
  cp "$SRC_DIR/references/"*.md "$SKILL_DIR/references/"
  chmod +x "$SKILL_DIR/check-drift.sh"
fi

# ------------- prepare permission rules -------------
if [[ "$MODE" == "1" ]]; then
  RULES_JSON='["Bash(goclaw:*)"]'
elif [[ "$MODE" == "2" ]]; then
  # Read-only verbs enumerated per known resource group.
  RULES_JSON='[
    "Bash(goclaw agents list:*)",
    "Bash(goclaw agents get:*)",
    "Bash(goclaw sessions list:*)",
    "Bash(goclaw sessions preview:*)",
    "Bash(goclaw traces list:*)",
    "Bash(goclaw traces get:*)",
    "Bash(goclaw usage summary:*)",
    "Bash(goclaw usage detail:*)",
    "Bash(goclaw status:*)",
    "Bash(goclaw health:*)",
    "Bash(goclaw whoami:*)",
    "Bash(goclaw version:*)",
    "Bash(goclaw auth whoami:*)",
    "Bash(goclaw auth list-contexts:*)",
    "Bash(goclaw tools builtin list:*)",
    "Bash(goclaw tools builtin get:*)",
    "Bash(goclaw skills list:*)",
    "Bash(goclaw skills get:*)",
    "Bash(goclaw providers list:*)",
    "Bash(goclaw memory list:*)",
    "Bash(goclaw memory get:*)",
    "Bash(goclaw knowledge-graph entities list:*)",
    "Bash(goclaw kg entities list:*)",
    "Bash(goclaw storage list:*)",
    "Bash(goclaw tenants list:*)",
    "Bash(goclaw system-config list:*)",
    "Bash(goclaw activity list:*)",
    "Bash(goclaw approvals list:*)",
    "Bash(goclaw delegations list:*)",
    "Bash(goclaw packages list:*)"
  ]'
else
  RULES_JSON='[]'
fi

# ------------- Mode 3 snippet output -------------
if [[ "$MODE" == "3" ]]; then
  cat <<EOF

Mode 3: no settings.json patching.

To grant permissions yourself, add to $SETTINGS:

{
  "permissions": {
    "allow": [
      "Bash(goclaw:*)"
    ]
  }
}

Or, for readonly-only, copy the enumerated rules:
  ./install.sh --mode 2 --dry-run | grep Bash

Done. Skill files at: $SKILL_DIR
EOF
  exit 0
fi

# ------------- backup settings.json -------------
if [[ -f "$SETTINGS" ]]; then
  if [[ "$DRY_RUN" == "1" ]]; then
    echo "[dry-run] cp $SETTINGS $SETTINGS.bak.<ts>"
  else
    if ! cp "$SETTINGS" "$SETTINGS.bak.$(date +%s)"; then
      echo "ERROR: failed to back up $SETTINGS — aborting to avoid destructive change." >&2
      exit 1
    fi
  fi
fi

# ------------- patch settings.json via python3 -------------
if [[ -z "$PY" ]]; then
  echo "No Python 3 available. Permissions NOT patched — see Mode 3 snippet above." >&2
  exit 0
fi

if [[ "$DRY_RUN" == "1" ]]; then
  echo "[dry-run] would merge into $SETTINGS permissions.allow: $RULES_JSON"
  exit 0
fi

SETTINGS_PATH="$SETTINGS" RULES="$RULES_JSON" "$PY" <<'PYEOF'
import json, os, sys
p = os.environ["SETTINGS_PATH"]
rules = json.loads(os.environ["RULES"])
data = {}
if os.path.exists(p):
    try:
        with open(p) as f:
            data = json.load(f) or {}
    except Exception as e:
        sys.exit(f"ERROR: {p} is not valid JSON ({e}); aborting to protect existing config.")
perms = data.setdefault("permissions", {})
allow = perms.setdefault("allow", [])
added = 0
for r in rules:
    if r not in allow:
        allow.append(r)
        added += 1
with open(p, "w") as f:
    json.dump(data, f, indent=2)
print(f"Merged {added} new permission rule(s) into {p} (already had {len(allow)-added} matching).")
PYEOF

echo ""
echo "Install complete."
echo "  Skill:    $SKILL_DIR"
echo "  Settings: $SETTINGS  (backup kept as .bak.<timestamp>)"
echo ""
echo "Next: restart Claude Code if running, then ask it \"list agents on goclaw\"."
