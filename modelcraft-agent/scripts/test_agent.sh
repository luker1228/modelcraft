#!/usr/bin/env bash
# test_agent.sh — 测试 agent 工具调用链路
# 用法:
#   ./scripts/test_agent.sh                          # 直连 agent :8000，fake token
#   ./scripts/test_agent.sh --via-apisix             # 走 APISIX :9080
#   ./scripts/test_agent.sh --token "Bearer eyJ..."  # 指定真实 token
#   ./scripts/test_agent.sh --via-apisix --token "Bearer eyJ..." --msg "帮我跳转到abcde项目"

set -euo pipefail

# ── 默认参数 ─────────────────────────────────────────────────────────────────
BASE_URL="http://127.0.0.1:8000"
TOKEN="Bearer test-token"
ORG="lukeco"
LAYER="org"
MSG="列出我所有的项目"

# ── 解析参数 ─────────────────────────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
  case "$1" in
    --via-apisix) BASE_URL="http://127.0.0.1:9080"; shift ;;
    --token)      TOKEN="$2"; shift 2 ;;
    --org)        ORG="$2";   shift 2 ;;
    --msg)        MSG="$2";   shift 2 ;;
    *) echo "unknown arg: $1"; exit 1 ;;
  esac
done

ENDPOINT="${BASE_URL}/copilotkit/admin"
THREAD_ID=$(python3 -c "import uuid; print(uuid.uuid4())")
RUN_ID=$(python3 -c "import uuid; print(uuid.uuid4())")
MSG_ID=$(python3 -c "import uuid; print(uuid.uuid4())")
CONTEXT_VAL=$(python3 -c "import json; print(json.dumps({'layer': '${LAYER}', 'orgName': '${ORG}'}))")

PAYLOAD=$(python3 -c "
import json
print(json.dumps({
    'threadId':  '${THREAD_ID}',
    'runId':     '${RUN_ID}',
    'state':     {},
    'messages':  [{'id': '${MSG_ID}', 'role': 'user', 'content': '${MSG}'}],
    'tools':     [],
    'forwardedProps': {'orgName': '${ORG}'},
    'context':   [{'description': '当前 AI 上下文', 'value': '${CONTEXT_VAL}'}],
}))
")

echo "=================================================="
echo "  endpoint : ${ENDPOINT}"
echo "  org      : ${ORG}"
echo "  message  : ${MSG}"
echo "  token    : ${TOKEN:0:30}..."
echo "=================================================="
echo ""

# ── 发请求，解析 SSE 事件 ─────────────────────────────────────────────────────
python3 - <<PYEOF
import urllib.request, json

req = urllib.request.Request(
    '${ENDPOINT}',
    data='${PAYLOAD}'.encode(),
    headers={
        'Content-Type':  'application/json',
        'Authorization': '${TOKEN}',
        'X-Org-Name':    '${ORG}',
    },
    method='POST'
)

try:
    with urllib.request.urlopen(req, timeout=60) as r:
        raw = r.read().decode()
except urllib.error.HTTPError as e:
    print(f"[ERROR] HTTP {e.code} {e.reason}")
    print(e.read().decode()[:500])
    raise SystemExit(1)

lines = [l for l in raw.split('\n') if l.startswith('data:')]
print(f"总 SSE 事件: {len(lines)}\n")

text_parts = []
for line in lines:
    try:
        d = json.loads(line[5:])
        t = d.get('type', '')
        if 'TOOL_CALL_START' in t:
            print(f"  🔧 TOOL_CALL_START  name={d.get('toolCallName', '')}")
        elif 'TOOL_CALL_END' in t:
            print(f"  ✅ TOOL_CALL_END    id={d.get('toolCallId', '')[:16]}...")
        elif 'TOOL_CALL_RESULT' in t:
            content = d.get('content', [])
            preview = str(content)[:120] if content else ''
            print(f"  📦 TOOL_CALL_RESULT {preview}")
        elif 'TEXT_MESSAGE_CONTENT' in t:
            text_parts.append(d.get('delta', ''))
        elif t in ('RUN_STARTED', 'RUN_FINISHED', 'RUN_ERROR'):
            print(f"  {'🚀' if 'STARTED' in t else '🏁' if 'FINISHED' in t else '❌'} {t}")
        elif 'STEP_STARTED' in t:
            print(f"  ▶  STEP_STARTED  step={d.get('stepName', '')}")
        elif 'STEP_FINISHED' in t:
            print(f"  ■  STEP_FINISHED step={d.get('stepName', '')}")
    except Exception:
        pass

if text_parts:
    print("\n── AI 回复 ──────────────────────────────────")
    print(''.join(text_parts))
    print("─────────────────────────────────────────────")
PYEOF
