#!/bin/bash
# Claude Code Stop hook - 企微通知（任务完成时）
WEBHOOK_URL="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=f2a8321e-835f-46e4-9318-95c4524df863"

INPUT=$(cat)

SESSION_ID=$(echo "$INPUT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('session_id','unknown'))" 2>/dev/null || echo "unknown")
CWD=$(echo "$INPUT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('cwd','unknown'))" 2>/dev/null || echo "unknown")

TIMESTAMP=$(date "+%Y-%m-%d %H:%M:%S")

MESSAGE=$(python3 -c "
import json
timestamp = '$TIMESTAMP'
cwd = '$CWD'
session_id = '$SESSION_ID'

lines = [
    '## Claude Code 任务已完成',
    f'> **时间**：{timestamp}',
    f'> **工作目录**：{cwd}',
    f'> **会话ID**：{session_id}',
]
content = '\n'.join(lines)
print(json.dumps({'msgtype': 'markdown', 'markdown': {'content': content}}, ensure_ascii=False))
" 2>/dev/null)

curl -s -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -d "$MESSAGE" \
  > /dev/null 2>&1

exit 0
