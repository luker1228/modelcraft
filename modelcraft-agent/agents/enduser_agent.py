# modelcraft-agent/agents/enduser_agent.py
"""End-user agent — serves end users querying and filtering data."""
from typing import Any

from langchain_openai import ChatOpenAI
from langgraph.checkpoint.memory import MemorySaver
from langgraph.graph import END, StateGraph
from langgraph.prebuilt import ToolNode

import config
from agents.shared import AgentState, sanitize_messages
from agents.admin_agent import _frontend_tool_names
from agents.tools import (
    get_model_fields,
    list_models,
    nl2filter,
    query_model,
)

ENDUSER_TOOLS = [
    list_models,
    get_model_fields,
    query_model,
    nl2filter,
]

_ENDUSER_TOOL_NODE = ToolNode(ENDUSER_TOOLS, handle_tool_errors=True)


def _build_enduser_graph() -> Any:
    _base_llm = ChatOpenAI(
        model=config.LLM_MODEL,
        api_key=config.LLM_API_KEY,
        base_url=config.LLM_BASE_URL if config.LLM_BASE_URL else None,
        temperature=0,
    ).bind_tools(ENDUSER_TOOLS)

    async def agent_node(state: AgentState):
        org = state.get("org_name", "")
        project = state.get("project_slug", "")

        frontend_actions = state.get("copilotkit", {}).get("actions", [])  # type: ignore[union-attr]
        if frontend_actions:
            extra = [
                {
                    "type": "function",
                    "function": {
                        "name": t["name"],
                        "description": t.get("description", ""),
                        "parameters": t.get("parameters", {"type": "object", "properties": {}}),
                    },
                }
                for t in frontend_actions
                if isinstance(t, dict) and t.get("name")
            ]
            existing = _base_llm.kwargs.get("tools", []) or []
            llm = _base_llm.bind(tools=[*existing, *extra])
        else:
            llm = _base_llm
        org = state.get("org_name", "")
        project = state.get("project_slug", "")

        context = (
            f"当前在数据页面（组织：{org}，项目：{project}）。\n"
            "可用工具：navigate_to_project、navigate_to_workspace、"
            "set_filter、clear_filter、highlight_records、"
            "list_models、get_model_fields、query_model、nl2filter、show_toast。\n"
            "重要原则：只能查询数据，不能修改数据模型或结构。\n"
            "筛选流程：先用 nl2filter 将用户意图转为 filter JSON，再调用前端 set_filter 应用。"
        )

        system_msg = {
            "role": "system",
            "content": (
                "你是 ModelCraft AI 助手（用户版），帮助终端用户探索和查询数据。\n\n"
                f"{context}\n\n"
                "通用原则：\n"
                "- 优先用自然语言解释数据内容，而不是展示原始 JSON\n"
                "- 如果用户不知道有哪些数据，先调用 list_models 介绍\n"
                "- 数据量大时引导用户用自然语言缩小范围\n"
                "- 遇到权限问题时提示用户联系管理员"
            ),
        }
        messages = [system_msg] + sanitize_messages(state["messages"])
        response = await llm.ainvoke(messages)
        return {"messages": [response]}

    def should_continue(state: AgentState):
        last = state["messages"][-1]
        if not (hasattr(last, "tool_calls") and last.tool_calls):
            return END
        frontend = _frontend_tool_names(state)
        backend_calls = [tc for tc in last.tool_calls if tc.get("name") not in frontend]
        return "tools" if backend_calls else END

    graph = StateGraph(AgentState)
    graph.add_node("agent", agent_node)
    graph.add_node("tools", _ENDUSER_TOOL_NODE)
    graph.set_entry_point("agent")
    graph.add_conditional_edges("agent", should_continue, {"tools": "tools", END: END})
    graph.add_edge("tools", "agent")
    return graph.compile(checkpointer=MemorySaver())


class _LazyGraph:
    _instance = None

    def __getattr__(self, name):
        if self._instance is None:
            type(self)._instance = _build_enduser_graph()
        return getattr(self._instance, name)


enduser_graph = _LazyGraph()
