# modelcraft-agent/tests/agents/test_enduser_agent.py
"""Tests for the end-user agent graph."""
import pytest
from agents.enduser_agent import enduser_graph, ENDUSER_TOOLS


def test_enduser_graph_compiles():
    """Graph must build without errors."""
    assert enduser_graph is not None


def test_enduser_graph_has_agent_and_tools_nodes():
    """Graph must have agent and tools nodes."""
    node_names = set(enduser_graph.nodes.keys())
    assert "agent" in node_names
    assert "tools" in node_names


def test_enduser_tools_do_not_include_list_projects():
    """End-user agent must NOT have list_projects (admin-only)."""
    tool_names = {t.name for t in ENDUSER_TOOLS}
    assert "list_projects" not in tool_names


def test_enduser_tools_include_query_tools():
    """End-user tools must be exactly the query/filter set."""
    tool_names = {t.name for t in ENDUSER_TOOLS}
    assert tool_names == {
        "list_models",
        "get_model_fields",
        "query_model",
        "nl2filter",
    }
