# modelcraft-agent/agents/admin_agent.py
"""Admin agent — serves tenant administrators."""
from typing import Any

from langchain_openai import ChatOpenAI
from langgraph.checkpoint.memory import MemorySaver
from langgraph.graph import END, StateGraph
from langgraph.prebuilt import ToolNode

import config
from agents.shared import AgentState
from agents.tools import (
    get_model_fields,
    list_models,
    list_projects,
    nl2filter,
    query_model,
)

ADMIN_TOOLS = [
    list_projects,
    list_models,
    get_model_fields,
    query_model,
    nl2filter,
]

_ADMIN_TOOL_NODE = ToolNode(ADMIN_TOOLS)


def _build_admin_graph() -> Any:
    llm = ChatOpenAI(
        model=config.LLM_MODEL,
        api_key=config.LLM_API_KEY,
        base_url=config.LLM_BASE_URL if config.LLM_BASE_URL else None,
        temperature=0,
    ).bind_tools(ADMIN_TOOLS)

    async def agent_node(state: AgentState):
        org = state.get("org_name", "")
        project = state.get("project_slug", "")
        layer = state.get("layer", "")
        current_model = state.get("current_model", "")
        current_db = state.get("current_db", "")

        if layer == "org":
            context = (
                f"当前在 Org 页面（组织：{org}）。\n"
                "可用工具：navigate_to_project、navigate_to_settings、open_create_project、"
                "highlight_project、list_projects、nl2filter、show_toast。\n"
                "注意：不可直接调用 list_models、query_model 等 project 级工具。\n"
                "如需操作项目数据，先调用 navigate_to_project(slug) 跳转到对应项目。"
            )
        elif layer == "project":
            model_ctx = f"，当前模型：{current_model}（数据库：{current_db}）" if current_model else ""
            context = (
                f"当前在 Project 页面（组织：{org}，项目：{project}{model_ctx}）。\n"
                "可用工具：navigate_to_org、navigate_to_model、navigate_to_data、"
                "open_create_model、open_create_record、open_edit_record、highlight_records、"
                "navigate_to_enums、navigate_to_cluster、navigate_to_rbac、navigate_to_end_users、"
                "set_filter、clear_filter、list_models、get_model_fields、query_model、"
                "nl2filter、show_toast。\n"
                "写操作规则：open_create_record 和 open_edit_record 只预填表单，用户点 Save 才真正保存。\n"
                "操作前先用 get_model_fields 确认字段名，避免预填错误字段。"
            )
        else:
            context = (
                f"当前组织：{org}{'，项目：' + project if project else ''}。\n"
                "可用工具取决于当前页面，请先询问用户当前在哪个页面。"
            )

        system_msg = {
            "role": "system",
            "content": (
                "你是 ModelCraft AI 助手（管理员版），帮助租户管理员通过对话完成所有操作。\n\n"
                f"{context}\n\n"
                "通用原则：\n"
                "- 操作数据前先用 list_models 和 get_model_fields 确认模型和字段存在\n"
                "- 删除操作禁止自动执行，必须引导用户在界面手动确认\n"
                "- 如果用户说筛选或过滤，先用 nl2filter 生成 filter JSON，再告知前端已应用\n"
                "- 完成操作后用 show_toast 通知用户结果"
            ),
        }
        messages = [system_msg] + state["messages"]
        response = await llm.ainvoke(messages)
        return {"messages": [response]}

    def should_continue(state: AgentState):
        last = state["messages"][-1]
        if hasattr(last, "tool_calls") and last.tool_calls:
            return "tools"
        return END

    graph = StateGraph(AgentState)
    graph.add_node("agent", agent_node)
    graph.add_node("tools", _ADMIN_TOOL_NODE)
    graph.set_entry_point("agent")
    graph.add_conditional_edges("agent", should_continue, {"tools": "tools", END: END})
    graph.add_edge("tools", "agent")
    return graph.compile(checkpointer=MemorySaver())


class _LazyGraph:
    _instance = None

    def __getattr__(self, name):
        if self._instance is None:
            type(self)._instance = _build_admin_graph()
        return getattr(self._instance, name)


admin_graph = _LazyGraph()
