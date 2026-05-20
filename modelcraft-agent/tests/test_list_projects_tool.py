"""
Tests for the list_projects agent tool.

Verifies tool-level behavior (above the GraphQL client layer):
  - happy path: returns JSON string of projects
  - GraphQL errors: returns "GraphQL error: ..." string (not raises)
  - HTTP error: raises exception (not silently swallowed)
  - logging: tool.call.start / tool.call.end emitted with correct fields
  - duplicate Bearer token: reproduces the production bug
"""
import json

import httpx
import pytest
import respx

import config
from logging_setup import setup_logging
from tests.conftest import capture_logs_with_context

MOCK_URL_PREFIX = config.GATEWAY_URL
ORG_URL = f"{MOCK_URL_PREFIX}/graphql/org/lukeco"

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _make_state(authorization: str = "Bearer test-token", org_name: str = "lukeco") -> dict:
    """Build a minimal AgentState dict for direct tool invocation."""
    return {
        "messages": [],
        "authorization": authorization,
        "org_name": org_name,
        "project_slug": "",
        "layer": "org",
        "current_model": "",
        "current_db": "",
    }


@pytest.fixture(autouse=True)
def _setup():
    setup_logging()


# ---------------------------------------------------------------------------
# Happy path
# ---------------------------------------------------------------------------

@pytest.mark.asyncio
@respx.mock
async def test_list_projects_tool_returns_json_array():
    """tool 正常调用时应返回 JSON 字符串，可解析为 list"""
    respx.post(ORG_URL).mock(
        return_value=httpx.Response(
            200,
            json={
                "data": {
                    "projects": [
                        {"id": "p1", "slug": "abcde", "title": "ABCDE Project",
                         "description": "", "status": "ACTIVE",
                         "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"},
                        {"id": "p2", "slug": "default", "title": "Default Project",
                         "description": "", "status": "ACTIVE",
                         "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"},
                    ]
                }
            },
        )
    )

    from agents.tools import list_projects

    result = await list_projects.ainvoke({"state": _make_state()})

    projects = json.loads(result)
    assert isinstance(projects, list)
    assert len(projects) == 2
    slugs = {p["slug"] for p in projects}
    assert "abcde" in slugs
    assert "default" in slugs


@pytest.mark.asyncio
@respx.mock
async def test_list_projects_tool_empty_project_list():
    """org 下无项目时，tool 应返回空 JSON 数组 '[]'"""
    respx.post(ORG_URL).mock(
        return_value=httpx.Response(200, json={"data": {"projects": []}})
    )

    from agents.tools import list_projects

    result = await list_projects.ainvoke({"state": _make_state()})

    assert json.loads(result) == []


# ---------------------------------------------------------------------------
# GraphQL-level errors (HTTP 200 + errors field)
# ---------------------------------------------------------------------------

@pytest.mark.asyncio
@respx.mock
async def test_list_projects_tool_returns_error_string_on_graphql_error():
    """GraphQL 返回 errors 时，tool 应返回 'GraphQL error: ...' 字符串而不是抛异常"""
    respx.post(ORG_URL).mock(
        return_value=httpx.Response(
            200,
            json={"data": None, "errors": [{"message": "Unauthorized"}]},
        )
    )

    from agents.tools import list_projects

    result = await list_projects.ainvoke({"state": _make_state(authorization="Bearer bad-token")})

    assert isinstance(result, str)
    assert "GraphQL error" in result
    assert "Unauthorized" in result


# ---------------------------------------------------------------------------
# HTTP-level errors — must NOT be silently swallowed
# ---------------------------------------------------------------------------

@pytest.mark.asyncio
@respx.mock
async def test_list_projects_tool_raises_on_http_401():
    """HTTP 401 时 tool 应 raise，确保错误不被静默吞掉"""
    respx.post(ORG_URL).mock(
        return_value=httpx.Response(401, text="Unauthorized")
    )

    from agents.tools import list_projects

    with pytest.raises(httpx.HTTPStatusError):
        await list_projects.ainvoke({"state": _make_state(authorization="Bearer expired-token")})


@pytest.mark.asyncio
@respx.mock
async def test_list_projects_tool_raises_on_http_500():
    """HTTP 500 时 tool 应 raise，不静默返回空数据"""
    respx.post(ORG_URL).mock(
        return_value=httpx.Response(500, text="Internal Server Error")
    )

    from agents.tools import list_projects

    with pytest.raises(httpx.HTTPStatusError):
        await list_projects.ainvoke({"state": _make_state()})


# ---------------------------------------------------------------------------
# Logging instrumentation
# ---------------------------------------------------------------------------

@pytest.mark.asyncio
@respx.mock
async def test_list_projects_tool_logs_start_and_end_on_success():
    """tool 成功时应记录 tool.call.start 和 tool.call.end（success=True）"""
    respx.post(ORG_URL).mock(
        return_value=httpx.Response(200, json={"data": {"projects": []}})
    )

    from agents.tools import list_projects

    with capture_logs_with_context() as cap:
        await list_projects.ainvoke({"state": _make_state()})

    start_log = next((l for l in cap if l["event"] == "tool.call.start"), None)
    end_log = next((l for l in cap if l["event"] == "tool.call.end"), None)

    assert start_log is not None, "tool.call.start not logged"
    assert start_log["tool_name"] == "list_projects"
    assert "lukeco" in start_log["args_summary"]

    assert end_log is not None, "tool.call.end not logged"
    assert end_log["tool_name"] == "list_projects"
    assert end_log["success"] is True
    assert isinstance(end_log["duration_ms"], float)
    assert end_log["duration_ms"] >= 0


@pytest.mark.asyncio
@respx.mock
async def test_list_projects_tool_logs_failure_on_http_error():
    """HTTP 错误时 tool 应记录 tool.call.end（success=False）"""
    respx.post(ORG_URL).mock(
        return_value=httpx.Response(401, text="Unauthorized")
    )

    from agents.tools import list_projects

    with capture_logs_with_context() as cap:
        with pytest.raises(httpx.HTTPStatusError):
            await list_projects.ainvoke({"state": _make_state()})

    end_log = next((l for l in cap if l["event"] == "tool.call.end"), None)
    assert end_log is not None, "tool.call.end not logged on failure"
    assert end_log["success"] is False


# ---------------------------------------------------------------------------
# Bug reproduction: duplicate Bearer token in authorization
# ---------------------------------------------------------------------------

@pytest.mark.asyncio
@respx.mock
async def test_list_projects_tool_duplicate_bearer_token_raises():
    """
    复现生产 Bug：前端将 authorization 拼成 'Bearer token1, Bearer token2'（两个相同 token）。
    非法 Authorization header 导致 gateway 返回 401，tool 应 raise 而非返回空。

    修复方向：前端应确保只传一个 Bearer token；
    agent 侧可在 _client() 中校验 authorization 格式。
    """
    respx.post(ORG_URL).mock(
        return_value=httpx.Response(401, text="Unauthorized")
    )

    from agents.tools import list_projects

    duplicate_auth = "Bearer eyJtoken1..., Bearer eyJtoken1..."  # 前端 bug 复现
    with pytest.raises(httpx.HTTPStatusError) as exc_info:
        await list_projects.ainvoke({"state": _make_state(authorization=duplicate_auth)})

    assert exc_info.value.response.status_code == 401
