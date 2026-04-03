#!/bin/bash
# Claude Code Notification hook - 企微通知（需要用户确认时）
WEBHOOK_URL="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=f2a8321e-835f-46e4-9318-95c4524df863"

INPUT=$(cat)

SESSION_ID=$(echo "$INPUT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('session_id','unknown'))" 2>/dev/null || echo "unknown")
CWD=$(echo "$INPUT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('cwd','unknown'))" 2>/dev/null || echo "unknown")
NOTIFICATION_TYPE=$(echo "$INPUT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('notification_type','unknown'))" 2>/dev/null || echo "unknown")
MESSAGE_TEXT=$(echo "$INPUT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('message',''))" 2>/dev/null || echo "")

case "$NOTIFICATION_TYPE" in
  "permission_prompt") TYPE_DESC="Claude Code 请求执行权限，需要你确认" ;;
  *)                   exit 0 ;;
esac

TIMESTAMP=$(date "+%Y-%m-%d %H:%M:%S")

MESSAGE=$(python3 -c "
import json
notification_type = '$NOTIFICATION_TYPE'
type_desc = '$TYPE_DESC'
message_text = '''$MESSAGE_TEXT'''
timestamp = '$TIMESTAMP'
cwd = '$CWD'
session_id = '$SESSION_ID'

if notification_type == 'permission_prompt':
    title = '## Claude Code 需要你的授权'
elif notification_type == 'idle_prompt':
    title = '## Claude Code 正在等待你'
else:
    title = '## Claude Code 通知'

lines = [
    title,
    f'> **时间**：{timestamp}',
    f'> **提示**：{type_desc}',
]
if message_text:
    lines.append(f'> **消息**：{message_text}')
lines += [
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
