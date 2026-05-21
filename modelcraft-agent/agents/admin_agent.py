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


def _frontend_tool_names(state: AgentState) -> set[str]:
    """Return the set of frontend tool names from CopilotKit state."""
    return {
        t.get("name", "")
        for t in state.get("copilotkit", {}).get("actions", [])  # type: ignore[union-attr]
        if isinstance(t, dict) and t.get("name")
    }


def _build_admin_graph() -> Any:
    # Base LLM — backend tools only. Frontend tools are added per-request in agent_node.
    _base_llm = ChatOpenAI(
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
        current_route = state.get("current_route", "")
        # Merge frontend tools (registered by useCopilotAction) into the LLM for this turn.
        # CopilotKit puts them in state["copilotkit"]["actions"] as plain dicts.
        frontend_actions = state.get("copilotkit", {}).get("actions", [])  # type: ignore[union-attr]
        from logging_setup import get_logger as _gl
        _gl().info("agent_node.debug",
                   copilotkit_keys=list((state.get("copilotkit") or {}).keys()),
                   frontend_action_count=len(frontend_actions),
                   frontend_action_names=[t.get("name") for t in frontend_actions if isinstance(t, dict)])
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

        if layer == "org":
            context = (
                f"当前在 Org 页面（组织：{org}）。\n"
                "UI 导航工具（只用 show_navigation_proposal）：\n"
                "  routeCatalog 和 aiTargets 已通过上下文注入，选取对应条目生成 candidates。\n"
                + (f"\n当前会话项目上下文：**{project}**（用户未明确指定时默认使用此项目）。" if project else "")
            )
        elif layer == "project":
            model_ctx = f"，当前模型：{current_model}（数据库：{current_db}）" if current_model else ""
            context = (
                f"当前在 Project 页面（组织：{org}，项目：**{project}**{model_ctx}）。\n"
                "UI 导航工具（只用 show_navigation_proposal）：\n"
                "  routeCatalog 和 aiTargets 已通过上下文注入，选取对应条目生成 candidates。\n\n"
                "数据查询工具：\n"
                "  list_databases、list_models、get_model_fields、query_model、nl2filter\n\n"
                "通知工具：show_toast"
            )
        else:
            project_ctx = f"当前会话项目上下文：**{project}**。" if project else "当前无项目上下文。"
            context = (
                f"当前组织：{org}。{project_ctx}\n"
                "UI 导航工具（只用 show_navigation_proposal）：\n"
                "  routeCatalog 和 aiTargets 已通过上下文注入。"
            )

        system_msg = {
            "role": "system",
            "content": (
                "你是 ModelCraft AI 助手（管理员版），帮助租户管理员通过对话完成所有操作。\n\n"
                f"{context}\n\n"
                "【核心规则：UI 导航必须通过 show_navigation_proposal】\n"
                "当用户询问'去哪里'、'在哪里配置'、'怎么操作'、'帮我跳转'时，\n"
                "必须调用 show_navigation_proposal 工具，不能只在文字里描述操作步骤。\n\n"
                "show_navigation_proposal 使用规范：\n"
                "1. response.candidates 每项必须有 type 字段：\n"
                "   - 'action_candidate'：能确定跳转目标时使用，actions 包含 ui.navigate 和/或 ui.highlight\n"
                "   - 'clarification_candidate'：意图不明确时使用，payload 描述用户意图\n"
                "2. ui.navigate 的 route 必须从注入的 routeCatalog 中选取，替换 :orgName/:projectSlug 为实际值\n"
                "3. ui.highlight 的 targetId 必须从注入的 aiTargets 中选取\n"
                "4. 即使只有一个候选项也要包装成 candidates 数组返回，不得直接执行\n\n"
                "数据查询工具调用规则：\n"
                "- 调用任何 project 级工具时，回复中必须明确说明「在项目 **{project}** 中...」\n"
                "- 如需 project 级操作但当前无项目上下文，先调用 list_projects 展示列表\n"
                "- 操作数据前先用 list_models 和 get_model_fields 确认模型和字段存在\n"
                "- 删除操作禁止自动执行，必须引导用户在界面手动确认\n"
                "- 完成操作后用 show_toast 通知用户结果"
            ).replace("{project}", project or "（未知项目）"),
        }
        messages = [system_msg] + sanitize_messages(state["messages"])
        response = await llm.ainvoke(messages)
        return {"messages": [response]}

    def should_continue(state: AgentState):
        last = state["messages"][-1]
        if not (hasattr(last, "tool_calls") and last.tool_calls):
            return END
        # Frontend tools must not go to ToolNode — the graph returns END so
        # ag-ui can stream TOOL_CALL_* events and the frontend executes them.
        frontend = _frontend_tool_names(state)
        backend_calls = [tc for tc in last.tool_calls if tc.get("name") not in frontend]
        return "tools" if backend_calls else END

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
