#!/usr/bin/env bash
# PostToolUse hook: 文件编辑/写入后按类型执行后端/前端 lint。
#
# 从 stdin 读取 JSON 输入（PostToolUse 事件），检查编辑文件路径：
# - Go 文件: 在后端目录运行 just lint
# - 前端代码文件: 在前端目录运行 npm run lint
# 并将 lint 结果通过 additionalContext 返回给 Agent。
#
# 退出码: 0 = 继续（lint 结果不影响工具执行）

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
BACKEND_DIR="$ROOT_DIR/modelcraft-backend"
FRONTEND_DIR="$ROOT_DIR/modelcraft-front"

INPUT=$(cat)

FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')

# Go 文件执行后端 lint
if [[ "$FILE_PATH" == *.go ]]; then
    LINT_OUTPUT=$(cd "$BACKEND_DIR" && just lint 2>&1 || true)

    if [ -n "$LINT_OUTPUT" ]; then
        # 将 lint 结果通过 additionalContext 反馈给 Agent
        echo "$LINT_OUTPUT" | jq -R --slurp '{hookSpecificOutput: {hookEventName: "PostToolUse", additionalContext: .}}'
    fi
fi

# 前端代码文件执行前端 lint
if [[ "$FILE_PATH" == modelcraft-front/* ]] && [[ "$FILE_PATH" =~ \.(js|jsx|ts|tsx)$ ]]; then
    FRONT_LINT_OUTPUT=$(cd "$FRONTEND_DIR" && npm run lint 2>&1 || true)

    if [ -n "$FRONT_LINT_OUTPUT" ]; then
        echo "$FRONT_LINT_OUTPUT" | jq -R --slurp '{hookSpecificOutput: {hookEventName: "PostToolUse", additionalContext: .}}'
    fi
fi

exit 0
