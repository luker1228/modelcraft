# modelcraft-agent/agents/admin_agent.py
"""Admin agent — serves tenant administrators."""
from typing import Any

from langchain_openai import ChatOpenAI
from langgraph.checkpoint.memory import MemorySaver
from langgraph.graph import END, StateGraph
from langgraph.prebuilt import ToolNode

import config
from agents.shared import AgentState, sanitize_messages
from agents.tools import (
    get_model_fields,
    list_databases,
    list_models,
    list_projects,
    nl2filter,
    query_model,
)

ADMIN_TOOLS = [
    list_projects,
    list_databases,
    list_models,
    get_model_fields,
    query_model,
    nl2filter,
]

_ADMIN_TOOL_NODE = ToolNode(
    ADMIN_TOOLS,
    # Catch ALL exceptions (not just ToolInvocationError) and convert them to
    # ToolMessages.  This guarantees MemorySaver always stores a complete
    # AIMessage(tool_calls) + ToolMessage pair, preventing broken message
    # sequences that cause DeepSeek 400 errors on subsequent requests.
    # Default (_default_handle_tool_errors) only catches ToolInvocationError;
    # handle_tool_errors=True uses handled_types=(Exception,) — catches everything.
    handle_tool_errors=True,
)


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
                + (f"\n当前会话项目上下文：**{project}**（用户未明确指定时默认使用此项目）。" if project else "")
            )
        elif layer == "project":
            model_ctx = f"，当前模型：{current_model}（数据库：{current_db}）" if current_model else ""
            context = (
                f"当前在 Project 页面（组织：{org}，项目：**{project}**{model_ctx}）。\n"
                "可用工具：navigate_to_org、navigate_to_model、navigate_to_data、"
                "open_create_model、open_create_record、open_edit_record、highlight_records、"
                "navigate_to_enums、navigate_to_cluster、navigate_to_rbac、navigate_to_end_users、"
                "set_filter、clear_filter、list_databases、list_models、get_model_fields、query_model、"
                "nl2filter、show_toast。\n"
                "数据操作顺序：先 list_databases → 再 list_models(database_name) → 再 get_model_fields(model_id)。\n"
                "写操作规则：open_create_record 和 open_edit_record 只预填表单，用户点 Save 才真正保存。\n"
                "操作前先用 get_model_fields 确认字段名，避免预填错误字段。"
            )
        else:
            project_ctx = f"当前会话项目上下文：**{project}**。" if project else "当前无项目上下文。"
            context = (
                f"当前组织：{org}。{project_ctx}\n"
                "可用工具取决于当前页面，请先询问用户当前在哪个页面。"
            )

        system_msg = {
            "role": "system",
            "content": (
                "你是 ModelCraft AI 助手（管理员版），帮助租户管理员通过对话完成所有操作。\n\n"
                f"{context}\n\n"
                "通用原则：\n"
                "- 调用任何 project 级工具（list_models、get_model_fields、query_model 等）时，\n"
                "  回复中必须明确说明「在项目 **{project}** 中...」，防止用户误会操作了错误的项目。\n"
                "- 如需 project 级操作但当前无项目上下文，先调用 list_projects 展示列表，\n"
                "  让用户选择后再继续，不得自行假设项目。\n"
                "- 操作数据前先用 list_models 和 get_model_fields 确认模型和字段存在。\n"
                "- 删除操作禁止自动执行，必须引导用户在界面手动确认。\n"
                "- 如果用户说筛选或过滤，先用 nl2filter 生成 filter JSON，再告知前端已应用。\n"
                "- 完成操作后用 show_toast 通知用户结果。"
            ).replace("{project}", project or "（未知项目）"),
        }
        messages = [system_msg] + sanitize_messages(state["messages"])
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
