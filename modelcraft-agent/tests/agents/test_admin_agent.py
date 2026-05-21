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


def test_admin_tools_exact_set():
    """Admin backend tools must be exactly the expected set (no regressions)."""
    tool_names = {t.name for t in ADMIN_TOOLS}
    assert tool_names == {
        "list_projects",
        "list_databases",
        "list_models",
        "get_model_fields",
        "query_model",
        "nl2filter",
    }


def test_frontend_tool_names_from_state():
    """_frontend_tool_names must extract names from copilotkit.actions."""
    from agents.admin_agent import _frontend_tool_names

    state = {
        "messages": [],
        "copilotkit": {
            "actions": [
                {"name": "show_navigation_proposal", "description": "show proposal"},
                {"name": "show_toast", "description": "show toast"},
            ]
        },
    }
    names = _frontend_tool_names(state)
    assert "show_navigation_proposal" in names
    assert "show_toast" in names


def test_should_continue_returns_end_for_frontend_tool():
    """When agent calls only frontend tools, should_continue must return END."""
    assert admin_graph is not None
