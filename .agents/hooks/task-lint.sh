#!/usr/bin/env bash
# PostToolUse hook: 文件编辑/写入后运行 just lint。
#
# 从 stdin 读取 JSON 输入（PostToolUse 事件），检查编辑的文件是否为 Go 文件，
# 如果是则运行 just lint 并将结果通过 additionalContext 返回给 Agent。
#
# 退出码: 0 = 继续（lint 结果不影响工具执行）

set -euo pipefail

INPUT=$(cat)

FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')

# 仅对 Go 文件执行 lint
if [[ "$FILE_PATH" == *.go ]]; then
    LINT_OUTPUT=$(just lint 2>&1 || true)

    if [ -n "$LINT_OUTPUT" ]; then
        # 将 lint 结果通过 additionalContext 反馈给 Agent
        echo "$LINT_OUTPUT" | jq -R --slurp '{hookSpecificOutput: {hookEventName: "PostToolUse", additionalContext: .}}'
    fi
fi

exit 0
