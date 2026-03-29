#!/usr/bin/env bash
# Batch validate all skills for path issues
# Usage: validate_all_skills.sh [skills-root-dir]

set -euo pipefail

SKILLS_ROOT="${1:-./.agents/skills}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VALIDATOR="$SCRIPT_DIR/validate_skill.sh"

if [ ! -d "$SKILLS_ROOT" ]; then
  echo "Error: Skills directory not found: $SKILLS_ROOT"
  exit 1
fi

if [ ! -x "$VALIDATOR" ]; then
  echo "Error: validator script not executable: $VALIDATOR"
  exit 1
fi

echo "=== Validating all skills in $SKILLS_ROOT ==="
echo

FAILED=0

for skill_dir in "$SKILLS_ROOT"/*/; do
  [ -d "$skill_dir" ] || continue
  [ -f "$skill_dir/SKILL.md" ] || continue

  skill_name="$(basename "$skill_dir")"
  echo "→ $skill_name"

  if "$VALIDATOR" "$skill_dir" >/tmp/skill-validate.out 2>&1; then
    echo "  ✓ Path check passed"
  else
    echo "  ❌ Path check failed"
    sed -n '1,80p' /tmp/skill-validate.out | sed 's/^/    /'
    FAILED=1
  fi
  echo
done

rm -f /tmp/skill-validate.out

if [ "$FAILED" -eq 0 ]; then
  echo "✓ All skills passed validation"
  exit 0
else
  echo "⚠ Some skills have path issues"
  exit 1
fi
