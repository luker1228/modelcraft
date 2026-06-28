#!/usr/bin/env bash
# ──────────────────────────────────────────────────────────────────
# 云托管 APISIX 入口脚本
# 1. 用 $PORT 环境变量替换 config.yaml 中的 node_listen 端口
# 2. 将控制权交还给原始 APISIX 入口脚本
# ──────────────────────────────────────────────────────────────────
set -euo pipefail

PORT="${PORT:-9080}"
CONFIG_FILE="/usr/local/apisix/conf/config.yaml"
TEMPLATE_FILE="/usr/local/apisix/conf/config-template.yaml"

echo "[cloudrun-entrypoint] PORT=${PORT}"

if [ -f "${TEMPLATE_FILE}" ]; then
    sed "s/node_listen: 9080/node_listen: ${PORT}/g" "${TEMPLATE_FILE}" > "${CONFIG_FILE}"
    echo "[cloudrun-entrypoint] Generated config.yaml with node_listen: ${PORT}"
else
    # fallback: 原地替换
    if [ -f "${CONFIG_FILE}" ]; then
        sed -i "s/node_listen: 9080/node_listen: ${PORT}/g" "${CONFIG_FILE}"
        echo "[cloudrun-entrypoint] Patched config.yaml node_listen → ${PORT}"
    fi
fi

# Preserve the base image default command when this image overrides ENTRYPOINT.
if [ "$#" -eq 0 ]; then
    set -- docker-start
fi

# 将控制权交还给 APISIX 原始入口脚本
exec /docker-entrypoint.sh "$@"
