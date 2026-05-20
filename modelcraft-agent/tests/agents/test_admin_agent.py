# modelcraft-agent/tests/agents/test_admin_agent.py
"""Tests for the admin agent graph."""
import pytest
from agents.admin_agent import admin_graph, ADMIN_TOOLS


def test_admin_graph_compiles():
    """Graph must build without errors."""
    assert admin_graph is not None


def test_admin_graph_has_agent_and_tools_nodes():
    """Graph must have agent and tools nodes."""
    node_names = set(admin_graph.nodes.keys())
    assert "agent" in node_names
    assert "tools" in node_names


def test_admin_tools_include_list_projects():
    """Admin agent must have access to list_projects."""
    tool_names = {t.name for t in ADMIN_TOOLS}
    assert "list_projects" in tool_names


def test_admin_tools_do_not_include_enduser_only():
    """Admin tool names must be exactly the expected set."""
    tool_names = {t.name for t in ADMIN_TOOLS}
    assert tool_names == {
        "list_projects",
        "list_models",
        "get_model_fields",
        "query_model",
        "nl2filter",
    }
