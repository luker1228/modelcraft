#!/usr/bin/env bash
# Validate paths in a skill directory
# Usage: validate_skill.sh <skill-dir>

set -euo pipefail

SKILL_DIR="${1:-.}"
SKILL_MD="$SKILL_DIR/SKILL.md"
ISSUES=0

if [ ! -f "$SKILL_MD" ]; then
  echo "Error: No SKILL.md found in $SKILL_DIR"
  exit 1
fi

echo "=== Validating skill: $(basename "$SKILL_DIR") ==="
echo

TMP_NO_CODE="$(mktemp)"
trap 'rm -f "$TMP_NO_CODE"' EXIT

# Remove fenced code blocks and inline code so examples won't be treated as real issues.
awk 'BEGIN{in_code=0} /^```/{in_code=!in_code; next} !in_code{print}' "$SKILL_MD" \
  | sed -E 's/`[^`]*`//g' > "$TMP_NO_CODE"

# 1) Absolute paths in natural text

echo "[1] Checking for absolute paths in narrative text..."
ABS_MATCHES="$(rg -n '(/Users/|/home/|/tmp/|/root/|/var/)' "$TMP_NO_CODE" || true)"
if [ -n "$ABS_MATCHES" ]; then
  echo "❌ Found absolute-path text outside code examples:"
  echo "$ABS_MATCHES"
  ISSUES=1
else
  echo "✓ No absolute-path text issues"
fi

# 2) Broken markdown file references

echo
echo "[2] Checking markdown file references..."
LINKS="$(rg -o '\]\([^)]+\)' "$SKILL_MD" | sed -E 's/^\]\((.*)\)$/\1/' | sort -u || true)"
if [ -z "$LINKS" ]; then
  echo "✓ No local markdown references found"
else
  while IFS= read -r ref; do
    [ -z "$ref" ] && continue

    # Skip external urls and anchors
    case "$ref" in
      http://*|https://*|mailto:*|\#*)
        continue
        ;;
    esac

    # Skip explicit glob/pattern references
    if [[ "$ref" == *"*"* || "$ref" == *"{"* || "$ref" == *"}"* ]]; then
      continue
    fi

    full_path="$SKILL_DIR/$ref"
    if [ -e "$full_path" ]; then
      echo "  ✓ $ref exists"
    else
      echo "  ❌ $ref not found"
      ISSUES=1
    fi
  done <<< "$LINKS"
fi

# 3) Standard directory hints (informational)

echo
echo "[3] Directory structure (informational)..."
[ -d "$SKILL_DIR/scripts" ] && echo "  ✓ scripts/ exists"
[ -d "$SKILL_DIR/references" ] && echo "  ✓ references/ exists"
[ -d "$SKILL_DIR/assets" ] && echo "  ✓ assets/ exists"


echo
if [ "$ISSUES" -eq 0 ]; then
  echo "=== Validation complete: PASS ==="
  exit 0
else
  echo "=== Validation complete: FAIL ==="
  exit 1
fi
