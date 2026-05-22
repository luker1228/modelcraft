"""
notebooks/nb_utils.py — Notebook 公共工具包

提供：
  - AGENT_ROOT / setup_env()      — 环境初始化
  - get_token(username, password)  — 登录获取真实 JWT
  - run_agent(endpoint, token, ...) — 发请求并解析 SSE 事件
  - print_events(events)           — 格式化打印 SSE 事件
"""
from __future__ import annotations

import json
import os
import sys
import uuid
from typing import Any

import httpx

# ── 1. 环境初始化 ─────────────────────────────────────────────────────────────

AGENT_ROOT: str = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))

def setup_env() -> None:
    """把 AGENT_ROOT 加入 sys.path 并加载 .env，幂等安全。"""
    if AGENT_ROOT not in sys.path:
        sys.path.insert(0, AGENT_ROOT)
    from dotenv import load_dotenv
    load_dotenv(os.path.join(AGENT_ROOT, ".env"))


# ── 2. 获取 Token ─────────────────────────────────────────────────────────────

async def get_token(
    username: str,
    password: str,
    gateway: str | None = None,
) -> tuple[str, str]:
    """登录并返回 (bearer_token, org_name)。

    Args:
        username: 登录用户名
        password: 登录密码
        gateway:  Gateway URL，默认读 config.GATEWAY_URL

    Returns:
        (token, org_name)  — token 格式 'Bearer eyJ...'
    """
    setup_env()
    import config  # noqa: PLC0415 — 延迟导入，保证 .env 已加载

    url = f"{gateway or config.GATEWAY_URL}/api/tenant/auth/login"
    async with httpx.AsyncClient(timeout=10) as c:
        resp = await c.post(
            url,
            json={
                "identifier": username,
                "identifierType": "USERNAME",
                "password": password,
            },
        )
    if resp.status_code != 200:
        raise RuntimeError(f"Login failed {resp.status_code}: {resp.text[:300]}")

    data = resp.json()
    token    = f"Bearer {data['accessToken']}"
    org_name = data.get("orgName", "")
    expires  = data.get("expiresIn", 0)
    print(f"✅ 登录成功  org={org_name}  token 有效期={expires}s")
    return token, org_name


# ── 3. 发送 Agent 请求 ────────────────────────────────────────────────────────

def make_payload(
    msg: str,
    *,
    org: str,
    layer: str = "org",
    project_slug: str = "",
    tools: list[dict] | None = None,
    thread_id: str | None = None,
    extra_context: list[dict] | None = None,
) -> dict[str, Any]:
    """构造 RunAgentInput payload。

    Args:
        extra_context: 额外注入的 context 列表（如 routeCatalog、aiTargets）。
                       会追加在默认「当前 AI 上下文」之后。
    """
    context = [
        {
            "description": "当前 AI 上下文",
            "value": json.dumps({
                "layer":       layer,
                "orgName":     org,
                "projectSlug": project_slug,
            }),
        }
    ]
    if extra_context:
        context.extend(extra_context)

    return {
        "threadId":  thread_id or str(uuid.uuid4()),
        "runId":     str(uuid.uuid4()),
        "state":     {},
        "messages":  [{"id": str(uuid.uuid4()), "role": "user", "content": msg}],
        "tools":     tools or [],
        "forwardedProps": {"orgName": org, "projectSlug": project_slug},
        "context":   context,
    }


# ── 5. 导航提案工具定义 ────────────────────────────────────────────────────────

NAV_TOOLS: list[dict] = [
    {
        "name": "show_toast",
        "description": "向用户显示一条临时通知消息（不需要用户在聊天框内查看）",
        "parameters": {
            "type": "object",
            "properties": {
                "message": {"type": "string", "description": "通知内容"},
                "type":    {"type": "string", "description": "success | error | info | warning"},
            },
            "required": ["message"],
        },
    },
    {
        "name": "ui_present_proposal",
        "description": "向用户展示 AI 导航提案卡片，用户可点击候选项执行页面跳转、元素高亮或发送澄清消息",
        "parameters": {
            "type": "object",
            "properties": {
                "response": {"type": "object", "properties": {}},
            },
            "required": ["response"],
        },
    },
]

ROUTE_CATALOG_CONTEXT: dict = {
    "description": (
        "系统所有可导航页面目录（routeCatalog）。"
        "调用 ui_present_proposal 时，ui.navigate 的 route 字段必须从 routeTemplate 派生，"
        "将 :orgName、:projectSlug 等参数替换为当前会话的实际值。"
    ),
    "value": json.dumps([
        {"routeTemplate": "/org/:orgName/workspace",                              "title": "项目列表",      "keywords": ["项目列表", "所有项目"]},
        {"routeTemplate": "/org/:orgName/project/:projectSlug/model-editor",      "title": "数据模型编辑器", "keywords": ["模型", "字段", "数据模型"]},
        {"routeTemplate": "/org/:orgName/project/:projectSlug/enums",             "title": "枚举管理",      "keywords": ["枚举", "enum"]},
        {"routeTemplate": "/org/:orgName/project/:projectSlug/rbac/roles",        "title": "RBAC 角色管理", "keywords": ["权限", "RBAC", "角色"]},
        {"routeTemplate": "/org/:orgName/project/:projectSlug/rbac/users",        "title": "RBAC 用户授权", "keywords": ["用户权限", "授权"]},
        {"routeTemplate": "/org/:orgName/project/:projectSlug/rbac/bundles",      "title": "权限包管理",    "keywords": ["权限包", "bundle"]},
        {"routeTemplate": "/org/:orgName/project/:projectSlug/end-users",         "title": "终端用户管理",  "keywords": ["终端用户", "end user"]},
        {"routeTemplate": "/org/:orgName/project/:projectSlug/settings",          "title": "项目设置",      "keywords": ["项目设置", "集群配置", "数据库连接"]},
        {"routeTemplate": "/org/:orgName/developers",                             "title": "成员管理",      "keywords": ["成员", "开发者"]},
        {"routeTemplate": "/org/:orgName/end-users",                              "title": "终端用户（Org）","keywords": ["终端用户", "org 级用户"]},
        {"routeTemplate": "/org/:orgName/settings",                               "title": "组织设置",      "keywords": ["组织设置"]},
    ], ensure_ascii=False),
}


def make_nav_payload(
    msg: str,
    *,
    org: str,
    layer: str = "org",
    project_slug: str = "",
    ai_targets: list[dict] | None = None,
    thread_id: str | None = None,
) -> dict[str, Any]:
    """构造包含前端导航工具 + routeCatalog 的 payload（用于测试 ui_present_proposal）。

    Args:
        ai_targets: 当前页面的 AiTarget 列表，默认空列表。
    """
    targets = ai_targets or []
    extra = [
        ROUTE_CATALOG_CONTEXT,
        {
            "description": "当前页面已注册的 AI 高亮目标（AiTarget）。调用 ui_present_proposal 时，ui.highlight 的 targetId 必须从这个列表中选取。",
            "value": json.dumps(targets, ensure_ascii=False),
        },
    ]
    return make_payload(
        msg,
        org=org,
        layer=layer,
        project_slug=project_slug,
        tools=NAV_TOOLS,
        thread_id=thread_id,
        extra_context=extra,
    )


# ── 6. 解析导航提案 ────────────────────────────────────────────────────────────

def parse_nav_proposal(events: list[dict]) -> dict | None:
    """从 SSE 事件流中提取 ui_present_proposal 的完整 response 对象。"""
    args_buf: dict[str, list[str]] = {}

    for ev in events:
        t = ev.get("type", "")
        if "TOOL_CALL_START" in t and ev.get("toolCallName") == "ui_present_proposal":
            call_id = ev.get("toolCallId", "__default__")
            args_buf[call_id] = []
            continue

        # 新旧事件格式兼容：
        # - 旧：TOOL_CALL_ARGS_DELTA（通常带 toolCallName）
        # - 新：TOOL_CALL_ARGS（toolCallName 可能为 null，仅携带 toolCallId + delta）
        if "TOOL_CALL_ARGS_DELTA" in t or "TOOL_CALL_ARGS" in t:
            call_id = ev.get("toolCallId", "__default__")
            tool_name = ev.get("toolCallName")
            if call_id in args_buf or tool_name == "ui_present_proposal":
                args_buf.setdefault(call_id, []).append(ev.get("delta", "") or "")

    for call_id, parts in args_buf.items():
        raw = "".join(parts)
        try:
            args = json.loads(raw)
            resp = args.get("response", {})
            if isinstance(resp, str):
                resp = json.loads(resp)
            return resp
        except Exception:
            pass
    return None


def print_nav_proposal(events: list[dict]) -> None:
    """打印导航提案候选项（人类可读）。"""
    proposal = parse_nav_proposal(events)
    if not proposal:
        print("  ⚠️  未找到 ui_present_proposal 调用")
        return
    print(f"  🧭 ui_present_proposal")
    print(f"     message     : {proposal.get('message', '')}")
    print(f"     proposalType: {proposal.get('proposalType', '')}")
    candidates = proposal.get("candidates", [])
    print(f"     candidates  : {len(candidates)} 项")
    for c in candidates:
        ctype = c.get("type", "?")
        icon = "✅" if ctype == "action_candidate" else "❓"
        print(f"       {icon} [{ctype}] {c.get('title', '')}")
        for a in c.get("actions", []):
            atype = a.get("type", "")
            args  = a.get("args", {})
            print(f"            {atype} → {json.dumps(args, ensure_ascii=False)}")


async def run_agent(
    endpoint: str,
    token: str,
    payload: dict[str, Any],
    org: str = "",
    timeout: int = 60,
) -> list[dict]:
    """发送 RunAgentInput 请求，返回解析后的 SSE 事件列表。

    Args:
        endpoint: 完整 URL，如 http://127.0.0.1:8000/copilotkit/admin
        token:    Bearer token
        payload:  由 make_payload() 生成的 dict
        org:      用于 X-Org-Name header（走直连时需要；走 APISIX 由网关注入）
        timeout:  超时秒数

    Returns:
        list[dict] — 解析后的 SSE 事件
    """
    headers = {
        "Content-Type":  "application/json",
        "Authorization": token,
    }
    if org:
        headers["X-Org-Name"] = org

    async with httpx.AsyncClient(timeout=timeout) as client:
        resp = await client.post(endpoint, headers=headers, json=payload)

    if resp.status_code != 200:
        raise RuntimeError(f"HTTP {resp.status_code}: {resp.text[:300]}")

    events: list[dict] = []
    for line in resp.text.split("\n"):
        if line.startswith("data:"):
            try:
                events.append(json.loads(line[5:]))
            except json.JSONDecodeError:
                pass
    return events


# ── 4. 格式化打印事件 ─────────────────────────────────────────────────────────

def print_events(events: list[dict], *, show_text: bool = True) -> None:
    """格式化打印 SSE 事件，汇总工具调用和最终文本。"""
    text_parts: list[str] = []
    print(f"共 {len(events)} 个 SSE 事件\n")

    for ev in events:
        t = ev.get("type", "")
        if "RUN_STARTED" in t:
            print("🚀 RUN_STARTED")
        elif "RUN_FINISHED" in t:
            print("🏁 RUN_FINISHED")
        elif "RUN_ERROR" in t:
            print(f"❌ RUN_ERROR: {ev.get('message', '')}")
        elif "STEP_STARTED" in t:
            print(f"▶  STEP  {ev.get('stepName', '')}")
        elif "TOOL_CALL_START" in t:
            print(f"  🔧 TOOL_CALL_START  name={ev.get('toolCallName', '')}")
        elif "TOOL_CALL_END" in t:
            print(f"  ✅ TOOL_CALL_END")
        elif "TOOL_CALL_RESULT" in t:
            content = ev.get("content", [])
            preview = (
                content[0].get("text", "")[:120]
                if content and isinstance(content[0], dict)
                else str(content)[:120]
            )
            print(f"  📦 RESULT  {preview}")
        elif "TEXT_MESSAGE_CONTENT" in t:
            text_parts.append(ev.get("delta", ""))

    if show_text and text_parts:
        print("\n── AI 回复 ─────────────────────────────────────────")
        print("".join(text_parts))
        print("────────────────────────────────────────────────────")
