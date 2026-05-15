"""Tests for GraphQLClient observability instrumentation."""
import pytest
import httpx
import respx

from logging_setup import setup_logging
from tests.conftest import capture_logs_with_context
from client.graphql_client import GraphQLClient

MOCK_URL_PREFIX = "http://gateway:8090"


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
