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
) -> dict[str, Any]:
    """构造 RunAgentInput payload。"""
    return {
        "threadId":  thread_id or str(uuid.uuid4()),
        "runId":     str(uuid.uuid4()),
        "state":     {},
        "messages":  [{"id": str(uuid.uuid4()), "role": "user", "content": msg}],
        "tools":     tools or [],
        "forwardedProps": {"orgName": org, "projectSlug": project_slug},
        "context": [
            {
                "description": "当前 AI 上下文",
                "value": json.dumps({
                    "layer":       layer,
                    "orgName":     org,
                    "projectSlug": project_slug,
                }),
            }
        ],
    }


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
