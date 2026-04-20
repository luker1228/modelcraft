#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_DIR="$ROOT_DIR/internal/app"

failed=0

echo "[check-withtx-repo-usage] scanning $APP_DIR"

# Rule 1: forbid ignored tx querier parameter in WithTx callback.
underscore_hits="$(rg -n -P 'WithTx\(ctx,\s*func\(ctx\s+context\.Context,\s*_\s+dbgen\.Querier\)\s+error' "$APP_DIR" || true)"
if [[ -n "$underscore_hits" ]]; then
  echo ""
  echo "[FAIL] Forbidden pattern: WithTx callback ignores dbgen.Querier with '_'"
  echo "$underscore_hits"
  failed=1
fi

# Rule 2: forbid direct service repo usage inside WithTx callback body.
# Heuristic: inside callback block, disallow `s.<name>Repo.` calls.
while IFS=: read -r file _; do
  [[ -z "$file" ]] && continue
  awk '
    BEGIN { in_cb=0; depth=0; start=0 }
    /WithTx\(ctx, *func\(ctx context\.Context, *[A-Za-z_][A-Za-z0-9_]* dbgen\.Querier\) error/ {
      in_cb=1
      start=NR
      depth=0
    }
    {
      if (in_cb==1) {
        line=$0
        opens=gsub(/\{/, "{", line)
        closes=gsub(/\}/, "}", line)
        depth += opens - closes

        if ($0 ~ /s\.[A-Za-z0-9_]*Repo\./ && $0 !~ /s\.deployRepo\./) {
          printf("%s:%d:%s\n", FILENAME, NR, $0)
          found=1
        }

        if (depth<=0 && NR>start) {
          in_cb=0
        }
      }
    }
    END { if (found==1) exit 10 }
  ' "$file" || {
    if [[ $? -eq 10 ]]; then
      failed=1
    else
      echo "[ERROR] parser failed for $file"
      failed=1
    fi
  }
done < <(rg -n "WithTx\(ctx,\s*func\(ctx context\.Context,\s*[A-Za-z_][A-Za-z0-9_]* dbgen\.Querier\) error" "$APP_DIR" --no-filename -l | sed 's/:.*//')

if [[ "$failed" -ne 0 ]]; then
  echo ""
  echo "[FAIL] WithTx repo usage checks failed"
  echo "Rules:"
  echo "  1) Do not use '_' for tx querier parameter"
  echo "  2) Do not call s.*Repo.* inside WithTx callback; build tx-scoped repos from q"
  exit 1
fi

echo "[PASS] WithTx repo usage checks passed"
