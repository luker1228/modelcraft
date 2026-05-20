# modelcraft-agent/tests/agents/test_tools.py
"""Smoke tests for agent tool functions."""
import json
from unittest.mock import AsyncMock, MagicMock, patch
import pytest

from agents.tools import list_projects, list_models, get_model_fields, query_model, nl2filter


FAKE_STATE = {
    "messages": [],
    "authorization": "Bearer test-token",
    "org_name": "testorg",
    "project_slug": "testproject",
    "layer": "project",
    "current_model": "",
    "current_db": "",
}


@pytest.mark.asyncio
async def test_list_projects_returns_json():
    mock_result = {"data": {"projects": [{"id": "1", "slug": "demo", "title": "Demo"}]}}
    with patch("agents.tools.make_client") as mock_make_client:
        mock_client = AsyncMock()
        mock_client.list_projects.return_value = mock_result
        mock_make_client.return_value = mock_client

        result = await list_projects.ainvoke({"state": FAKE_STATE})

    data = json.loads(result)
    assert data[0]["slug"] == "demo"


@pytest.mark.asyncio
async def test_list_projects_returns_graphql_error():
    mock_result = {"errors": [{"message": "Unauthorized"}]}
    with patch("agents.tools.make_client") as mock_make_client:
        mock_client = AsyncMock()
        mock_client.list_projects.return_value = mock_result
        mock_make_client.return_value = mock_client

        result = await list_projects.ainvoke({"state": FAKE_STATE})

    assert "GraphQL error" in result


@pytest.mark.asyncio
async def test_nl2filter_returns_valid_json():
    fake_llm_response = MagicMock()
    fake_llm_response.content = '{"name": {"contains": "张"}}'
    with patch("agents.tools._get_llm") as mock_get_llm:
        mock_llm = AsyncMock()
        mock_llm.ainvoke.return_value = fake_llm_response
        mock_get_llm.return_value = mock_llm

        result = await nl2filter.ainvoke({
            "natural_language": "名字包含张",
            "field_names": ["name", "age"],
        })

    parsed = json.loads(result)
    assert "name" in parsed
