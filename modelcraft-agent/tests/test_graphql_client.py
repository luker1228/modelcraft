"""Tests for GraphQLClient observability instrumentation."""
import pytest
import httpx
import respx

from logging_setup import setup_logging
from tests.conftest import capture_logs_with_context
from client.graphql_client import GraphQLClient

MOCK_URL_PREFIX = "http://gateway:8090"
ORG_URL = f"{MOCK_URL_PREFIX}/graphql/org/lukeco"


@pytest.fixture(autouse=True)
def _setup():
    setup_logging()


@pytest.mark.asyncio
@respx.mock
async def test_find_many_logs_graphql_call_start_and_end():
    respx.post(url__startswith=MOCK_URL_PREFIX).mock(
        return_value=httpx.Response(
            200,
            json={"data": {"findMany": {"items": [{"id": "1"}], "totalCount": 1}}},
        )
    )
    client = GraphQLClient(authorization="Bearer test-token")

    with capture_logs_with_context() as cap:
        await client.find_many(
            org_name="org1",
            project_slug="proj1",
            db_name="maindb",
            model_name="User",
            fields=["id"],
        )

    start_log = next((l for l in cap if l["event"] == "graphql.call.start"), None)
    end_log = next((l for l in cap if l["event"] == "graphql.call.end"), None)

    assert start_log is not None, "graphql.call.start not logged"
    assert start_log["operation"] == "findMany"
    assert "url" in start_log

    assert end_log is not None, "graphql.call.end not logged"
    assert end_log["has_errors"] is False
    assert end_log["status_code"] == 200
    assert isinstance(end_log["duration_ms"], float)
    assert end_log["duration_ms"] >= 0


@pytest.mark.asyncio
@respx.mock
async def test_find_many_logs_has_errors_true_when_graphql_errors_present():
    respx.post(url__startswith=MOCK_URL_PREFIX).mock(
        return_value=httpx.Response(
            200,
            json={"data": None, "errors": [{"message": "Not found"}]},
        )
    )
    client = GraphQLClient(authorization="Bearer test-token")

    with capture_logs_with_context() as cap:
        await client.find_many(
            org_name="org1",
            project_slug="proj1",
            db_name="maindb",
            model_name="User",
            fields=["id"],
        )

    end_log = next(l for l in cap if l["event"] == "graphql.call.end")
    assert end_log["has_errors"] is True


@pytest.mark.asyncio
@respx.mock
async def test_find_many_logs_error_on_http_failure():
    respx.post(url__startswith=MOCK_URL_PREFIX).mock(
        return_value=httpx.Response(500, text="Internal Server Error")
    )
    client = GraphQLClient(authorization="Bearer test-token")

    with capture_logs_with_context() as cap:
        with pytest.raises(httpx.HTTPStatusError):
            await client.find_many(
                org_name="org1",
                project_slug="proj1",
                db_name="maindb",
                model_name="User",
                fields=["id"],
            )

    error_log = next((l for l in cap if l["event"] == "error"), None)
    assert error_log is not None, "error event not logged on HTTP 500"

    end_log = next((l for l in cap if l["event"] == "graphql.call.end"), None)
    assert end_log is not None, "graphql.call.end not emitted on error path"
    assert end_log["has_errors"] is True


# ---------------------------------------------------------------------------
# list_projects tests
# ---------------------------------------------------------------------------

@pytest.mark.asyncio
@respx.mock
async def test_list_projects_returns_project_list():
    """list_projects 正常返回项目列表"""
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
    client = GraphQLClient(authorization="Bearer test-token")

    result = await client.list_projects(org_name="lukeco")

    projects = result["data"]["projects"]
    assert len(projects) == 2
    assert projects[0]["slug"] == "abcde"
    assert projects[1]["slug"] == "default"


@pytest.mark.asyncio
@respx.mock
async def test_list_projects_logs_graphql_call_start_and_end():
    """list_projects 应记录 graphql.call.start 和 graphql.call.end 日志"""
    respx.post(ORG_URL).mock(
        return_value=httpx.Response(200, json={"data": {"projects": []}})
    )
    client = GraphQLClient(authorization="Bearer test-token")

    with capture_logs_with_context() as cap:
        await client.list_projects(org_name="lukeco")

    start_log = next((l for l in cap if l["event"] == "graphql.call.start"), None)
    end_log = next((l for l in cap if l["event"] == "graphql.call.end"), None)

    assert start_log is not None, "graphql.call.start not logged"
    assert start_log["operation"] == "listProjects"
    assert "lukeco" in start_log["url"]

    assert end_log is not None, "graphql.call.end not logged"
    assert end_log["has_errors"] is False
    assert end_log["status_code"] == 200
    assert isinstance(end_log["duration_ms"], float)
    assert end_log["duration_ms"] >= 0


@pytest.mark.asyncio
@respx.mock
async def test_list_projects_logs_has_errors_true_on_graphql_error():
    """GraphQL 返回 errors 字段时，has_errors 应为 True，函数仍正常返回（不抛异常）"""
    respx.post(ORG_URL).mock(
        return_value=httpx.Response(
            200,
            json={"data": None, "errors": [{"message": "Unauthorized"}]},
        )
    )
    client = GraphQLClient(authorization="Bearer invalid-token")

    with capture_logs_with_context() as cap:
        result = await client.list_projects(org_name="lukeco")

    assert "errors" in result
    end_log = next(l for l in cap if l["event"] == "graphql.call.end")
    assert end_log["has_errors"] is True


@pytest.mark.asyncio
@respx.mock
async def test_list_projects_http_401_raises_and_logs_error():
    """HTTP 401 时应 raise HTTPStatusError，并记录 error 日志"""
    respx.post(ORG_URL).mock(
        return_value=httpx.Response(401, text="Unauthorized")
    )
    client = GraphQLClient(authorization="Bearer expired-token")

    with capture_logs_with_context() as cap:
        with pytest.raises(httpx.HTTPStatusError):
            await client.list_projects(org_name="lukeco")

    error_log = next((l for l in cap if l["event"] == "error"), None)
    assert error_log is not None, "error event not logged on HTTP 401"

    end_log = next((l for l in cap if l["event"] == "graphql.call.end"), None)
    assert end_log is not None
    assert end_log["has_errors"] is True
    assert end_log["status_code"] == 401


@pytest.mark.asyncio
@respx.mock
async def test_list_projects_uses_org_level_url_not_project_url():
    """list_projects 应调用 org 级 URL（不含 /project/ 路径段）"""
    route = respx.post(ORG_URL).mock(
        return_value=httpx.Response(200, json={"data": {"projects": []}})
    )
    client = GraphQLClient(authorization="Bearer test-token")

    await client.list_projects(org_name="lukeco")

    assert route.called, "org-level URL was not called"
    called_url = str(route.calls[0].request.url)
    assert "/project/" not in called_url, "list_projects must NOT use project-level URL"
    assert "lukeco" in called_url


@pytest.mark.asyncio
@respx.mock
async def test_list_projects_duplicate_bearer_token_causes_http_error():
    """
    复现 Bug：authorization 包含两个重复的 Bearer token（前端 Bug，格式非法）。
    gateway 返回 401，客户端应抛出异常而非静默返回空数据。
    """
    respx.post(ORG_URL).mock(
        return_value=httpx.Response(401, text="Unauthorized")
    )
    # 模拟前端错误地拼了两个相同 token
    duplicate_auth = "Bearer eyJtoken..., Bearer eyJtoken..."
    client = GraphQLClient(authorization=duplicate_auth)

    with pytest.raises(httpx.HTTPStatusError) as exc_info:
        await client.list_projects(org_name="lukeco")

    assert exc_info.value.response.status_code == 401
