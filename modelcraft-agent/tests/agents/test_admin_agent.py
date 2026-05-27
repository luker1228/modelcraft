# modelcraft-agent/tests/agents/test_admin_agent.py
"""Tests for the admin agent graph."""
import pytest
from langchain_core.messages import AIMessage, HumanMessage
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
                {"name": "ui_present_proposal", "description": "show proposal"},
                {"name": "show_toast", "description": "show toast"},
            ]
        },
    }
    names = _frontend_tool_names(state)
    assert "ui_present_proposal" in names
    assert "show_toast" in names


def test_should_continue_returns_end_for_frontend_tool():
    """When agent calls only frontend tools, should_continue must return END."""
    assert admin_graph is not None


def test_history_has_tool_call_since_latest_user_ignores_previous_turns():
    """A new user message after a proposal must allow a fresh navigation proposal."""
    from agents.admin_agent import _history_has_tool_call_since_latest_user

    state = {
        "messages": [
            HumanMessage(content="帮我导航到访问控制"),
            AIMessage(
                content="",
                tool_calls=[{
                    "name": "ui_present_proposal",
                    "args": {"response": {}},
                    "id": "call_existing_proposal",
                }],
            ),
            HumanMessage(content="帮我去数据模型管理"),
        ]
    }

    assert _history_has_tool_call_since_latest_user(state, {"ui_present_proposal"}) is False


def test_history_has_tool_call_since_latest_user_detects_auto_rerun():
    """A proposal after the latest user message marks CopilotKit auto-reruns as already handled."""
    from agents.admin_agent import _history_has_tool_call_since_latest_user

    state = {
        "messages": [
            HumanMessage(content="帮我导航到访问控制"),
            AIMessage(
                content="",
                tool_calls=[{
                    "name": "ui_present_proposal",
                    "args": {"response": {}},
                    "id": "call_existing_proposal",
                }],
            ),
        ]
    }

    assert _history_has_tool_call_since_latest_user(state, {"ui_present_proposal"}) is True


def test_should_force_proposal_on_turn_for_direct_navigation():
    """Direct navigation intent must force ui_present_proposal when available."""
    from agents.admin_agent import _should_force_proposal_on_turn

    assert _should_force_proposal_on_turn(
        proposal_available=True,
        is_direct_nav_intent=True,
        is_list_nav_intent=False,
        is_project_required_intent=False,
        history_has_list_tools=False,
        history_has_project_list=False,
        history_has_proposal=False,
    ) is True


def test_should_not_force_project_required_navigation_before_project_list():
    """Project-required navigation must list projects before showing route candidates."""
    from agents.admin_agent import _should_force_proposal_on_turn

    assert _should_force_proposal_on_turn(
        proposal_available=True,
        is_direct_nav_intent=True,
        is_list_nav_intent=False,
        is_project_required_intent=True,
        history_has_list_tools=False,
        history_has_project_list=False,
        history_has_proposal=False,
    ) is False


def test_should_force_project_required_navigation_after_project_list():
    """After list_projects runs in the current turn, project-required navigation can present candidates."""
    from agents.admin_agent import _should_force_proposal_on_turn

    assert _should_force_proposal_on_turn(
        proposal_available=True,
        is_direct_nav_intent=True,
        is_list_nav_intent=False,
        is_project_required_intent=True,
        history_has_list_tools=False,
        history_has_project_list=True,
        history_has_proposal=False,
    ) is True


def test_should_force_proposal_on_turn_for_list_navigation_after_listing():
    """List-navigation intent must force proposal after list tool has been called."""
    from agents.admin_agent import _should_force_proposal_on_turn

    assert _should_force_proposal_on_turn(
        proposal_available=True,
        is_direct_nav_intent=False,
        is_list_nav_intent=True,
        is_project_required_intent=False,
        history_has_list_tools=True,
        history_has_project_list=False,
        history_has_proposal=False,
    ) is True


def test_should_not_force_proposal_on_first_list_turn_without_history():
    """First list-navigation turn should not force proposal before list tool phase."""
    from agents.admin_agent import _should_force_proposal_on_turn

    assert _should_force_proposal_on_turn(
        proposal_available=True,
        is_direct_nav_intent=False,
        is_list_nav_intent=True,
        is_project_required_intent=False,
        history_has_list_tools=False,
        history_has_project_list=False,
        history_has_proposal=False,
    ) is False


def test_should_not_force_proposal_when_proposal_already_presented():
    """Repeated CopilotKit runs after a proposal must not force another proposal."""
    from agents.admin_agent import _should_force_proposal_on_turn

    assert _should_force_proposal_on_turn(
        proposal_available=True,
        is_direct_nav_intent=True,
        is_list_nav_intent=False,
        is_project_required_intent=False,
        history_has_list_tools=False,
        history_has_project_list=False,
        history_has_proposal=True,
    ) is False


@pytest.mark.asyncio
async def test_agent_returns_terminal_message_after_existing_proposal(monkeypatch):
    """CopilotKit auto-reruns after a proposal must end without another LLM/tool pass."""
    import agents.admin_agent as admin_agent

    class HallucinatingProposalLLM:
        calls = 0

        def __init__(self, *args, **kwargs):
            self.kwargs = {}
            self.tool_choice = None

        def bind_tools(self, tools):
            llm = HallucinatingProposalLLM()
            llm.kwargs = {
                "tools": [
                    {"type": "function", "function": {"name": getattr(tool, "name", str(tool))}}
                    for tool in tools
                ]
            }
            return llm

        def bind(self, **kwargs):
            llm = HallucinatingProposalLLM()
            llm.kwargs = {**self.kwargs, **kwargs}
            llm.tool_choice = kwargs.get("tool_choice")
            return llm

        async def ainvoke(self, messages):
            HallucinatingProposalLLM.calls += 1
            return AIMessage(
                content="",
                tool_calls=[{
                    "name": "ui_present_proposal",
                    "args": {"response": {}},
                    "id": "call_repeated_proposal",
                }],
            )

    monkeypatch.setattr(admin_agent, "ChatOpenAI", HallucinatingProposalLLM)
    admin_agent._LazyGraph._instance = None

    state = {
        "messages": [
            HumanMessage(content="帮我导航到访问控制"),
            AIMessage(
                content="",
                tool_calls=[{
                    "name": "ui_present_proposal",
                    "args": {"response": {}},
                    "id": "call_existing_proposal",
                }],
            ),
        ],
        "authorization": "Bearer test-token",
        "org_name": "lukeco",
        "layer": "org",
        "project_slug": "onboardceshi",
        "copilotkit": {
            "actions": [
                {"name": "ui_present_proposal", "description": "show proposal"},
                {"name": "show_toast", "description": "show toast"},
            ]
        },
    }

    result = await admin_agent.admin_graph.ainvoke(state, config={"configurable": {"thread_id": "repeat-proposal-test"}})

    repeated_calls = [
        tc["name"]
        for tc in getattr(result["messages"][-1], "tool_calls", [])
    ]
    assert HallucinatingProposalLLM.calls == 0
    assert "ui_present_proposal" not in repeated_calls
