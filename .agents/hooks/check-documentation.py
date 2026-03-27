#!/usr/bin/env python3
"""
PreToolUse hook: 检查文件写入前的规范。

从 stdin 读取 JSON 输入（PreToolUse 事件），检查：
- 阻止写入 .env 等敏感文件
- 仅允许在 .plan/、docs/、openspec/ 下新建 .md 文件（已有文件允许编辑）

退出码: 0 = 允许, 2 = 阻止
"""
import json
import sys
import os

# 受保护文件模式（不允许写入）
PROTECTED_PATTERNS = [".env", "package-lock.json", "node_modules/", ".golangci.yml"]

# 允许新建 .md 文件的目录前缀
ALLOWED_MD_DIRS = [".plan/", ".agents/", "ai-metadata/","docs/", "openspec/"]

def main():
    try:
        data = json.load(sys.stdin)
    except (json.JSONDecodeError, EOFError):
        sys.exit(0)

    file_path = data.get("tool_input", {}).get("file_path", "")

    if not file_path:
        sys.exit(0)

    # 检查受保护文件
    for pattern in PROTECTED_PATTERNS:
        if pattern in file_path:
            print(json.dumps({
                "continue": False,
                "stopReason": f"写入受保护文件被阻止: {file_path} 包含 '{pattern}'"
            }))
            sys.exit(2)

    # 检查 .md 文件：仅允许在指定目录下新建
    if file_path.endswith(".md") and not os.path.exists(file_path):
        is_allowed = any(d in file_path for d in ALLOWED_MD_DIRS)
        if not is_allowed:
            print(json.dumps({
                "continue": False,
                "stopReason": (
                    f"新建 .md 文件被阻止: {file_path}\n"
                    f".md 文件只能在以下目录新建: {', '.join(ALLOWED_MD_DIRS)}"
                )
            }))
            sys.exit(2)

    sys.exit(0)

if __name__ == "__main__":
    main()
