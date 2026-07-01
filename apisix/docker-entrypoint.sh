#!/usr/bin/env bash
set -euo pipefail

PORT="${PORT:-9080}"
APP_APISIX_DIR="/app/apisix"
CONFIG_FILE="/usr/local/apisix/conf/config.yaml"
TEMPLATE_FILE="${APP_APISIX_DIR}/config.template.yaml"
APISIX_TEMPLATE_FILE="${APP_APISIX_DIR}/apisix.template.yaml"
APISIX_FILE="/usr/local/apisix/conf/apisix.yaml"
FRONTEND_URL="${FRONTEND_URL:-http://localhost:3100}"
BACKEND_UPSTREAM="${BACKEND_UPSTREAM:-backend:8080}"
AGENT_UPSTREAM="${AGENT_UPSTREAM:-agent:8000}"
JWT_PUBLIC_KEY="${JWT_PUBLIC_KEY:-}"

echo "[cloudrun-entrypoint] PORT=${PORT}"

if [ -f "${TEMPLATE_FILE}" ]; then
    sed "s/node_listen: 9080/node_listen: ${PORT}/g" "${TEMPLATE_FILE}" > "${CONFIG_FILE}"
    echo "[cloudrun-entrypoint] Generated config.yaml with node_listen: ${PORT}"
else
    if [ -f "${CONFIG_FILE}" ]; then
        sed -i "s/node_listen: 9080/node_listen: ${PORT}/g" "${CONFIG_FILE}"
        echo "[cloudrun-entrypoint] Patched config.yaml node_listen -> ${PORT}"
    fi
fi

if [ -f "${APISIX_TEMPLATE_FILE}" ]; then
    python3 - <<'PY'
import os
from pathlib import Path

template = Path("/app/apisix/apisix.template.yaml").read_text()
rendered = (
    template
    .replace("__FRONTEND_URL__", os.environ["FRONTEND_URL"])
    .replace("__BACKEND_UPSTREAM__", os.environ["BACKEND_UPSTREAM"])
    .replace("__AGENT_UPSTREAM__", os.environ["AGENT_UPSTREAM"])
    .replace("__JWT_PUBLIC_KEY__", os.environ["JWT_PUBLIC_KEY"])
)
Path("/usr/local/apisix/conf/apisix.yaml").write_text(rendered)
PY
    echo "[cloudrun-entrypoint] Rendered apisix.yaml for ${BACKEND_UPSTREAM} and ${AGENT_UPSTREAM}"
fi

if [ "$#" -eq 0 ]; then
    set -- docker-start
fi

exec /docker-entrypoint.sh "$@"
