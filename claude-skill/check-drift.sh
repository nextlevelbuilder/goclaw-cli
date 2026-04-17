#!/usr/bin/env bash
# check-drift.sh — validate that every --flag mentioned in references/*.md
# exists in goclaw-cli/cmd/*.go source. Exit non-zero on mismatch.
# Intended for CI: catches skill documentation going stale relative to CLI.
set -euo pipefail

SKILL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SKILL_DIR}/.." && pwd)"
CMD_DIR="${REPO_ROOT}/cmd"

if [[ ! -d "$CMD_DIR" ]]; then
  echo "ERROR: cmd/ not found at $CMD_DIR. Run from goclaw-cli repo or mount it."
  exit 1
fi

# Cobra built-ins we never expect to find in goclaw cmd/*.go
ALLOWLIST=(help verbose output server token insecure yes profile tenant-id version)

is_allowed() {
  local f="$1"
  for a in "${ALLOWLIST[@]}"; do
    [[ "$f" == "$a" ]] && return 0
  done
  return 1
}

TOTAL=0
MISSING=0
declare -a MISSING_FLAGS=()

# Extract flag names (--foo, --foo-bar) from reference md files.
# Match --flag-name (letters/digits/hyphens), excluding trailing punctuation.
while IFS= read -r line; do
  [[ -z "$line" ]] && continue
  flag="$(echo "$line" | sed -E 's/^.*--([a-zA-Z][a-zA-Z0-9-]*).*$/\1/')"
  [[ -z "$flag" ]] && continue
  TOTAL=$((TOTAL + 1))
  if is_allowed "$flag"; then
    continue
  fi
  # Grep the cmd/*.go source for Flags().<Type>("$flag" ...).
  if ! grep -rEq "Flags\(\)\.(String|Int|Bool|StringSlice|StringP|IntP|BoolP)\([^)]*\"$flag\"" "$CMD_DIR" 2>/dev/null; then
    MISSING=$((MISSING + 1))
    MISSING_FLAGS+=("$flag")
  fi
done < <(grep -hoE -- '--[a-zA-Z][a-zA-Z0-9-]*' "$SKILL_DIR/references/"*.md 2>/dev/null | sort -u)

echo "Checked $(grep -hoE -- '--[a-zA-Z][a-zA-Z0-9-]*' "$SKILL_DIR/references/"*.md 2>/dev/null | sort -u | wc -l | tr -d ' ') unique flag mentions across references/."
echo "Unresolved: $MISSING"

if (( MISSING > 0 )); then
  printf '%s\n' "Missing from cmd/*.go (possible drift or false positive):"
  for f in "${MISSING_FLAGS[@]}"; do
    printf '  --%s\n' "$f"
  done
  exit 1
fi

echo "OK — no drift detected."
exit 0
