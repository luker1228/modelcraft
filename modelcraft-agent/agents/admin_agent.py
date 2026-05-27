# modelcraft-agent/agents/admin_agent.py
"""Admin agent — serves tenant administrators."""
from typing import Any

from langchain_core.messages import AIMessage, SystemMessage
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


def _copilotkit_context_messages(state: AgentState) -> list[SystemMessage]:
    """Convert state['copilotkit']['context'] entries to SystemMessages.

    CopilotKit stores useCopilotReadable data (routeCatalog, aiTargets, page info…)
    in state['copilotkit']['context'] as [{'description': ..., 'value': ...}, …].
    Without this conversion the LLM never sees the injected context.
    """
    entries = state.get("copilotkit", {}).get("context", [])  # type: ignore[union-attr]
    msgs = []
    for entry in entries:
        if not isinstance(entry, dict):
            continue
        desc  = entry.get("description", "")
        value = entry.get("value", "")
        if desc or value:
            msgs.append(SystemMessage(content=f"{desc}\n{value}" if desc else value))
    return msgs


_NAV_INTENT_HINTS = (
    "帮我去",
    "带我去",
    "跳转",
    "导航",
    "怎么进入",
    "在哪里",
    "在哪",
    "打开",
    "进入",
    "配置权限",
    "模型管理页面",
    "数据模型管理",
)

_LIST_NAV_INTENT_HINTS = (
    "当前有哪些项目",
    "有哪些项目",
    "列出项目",
    "项目列表",
    "有哪些模型",
    "列出模型",
)

_PROJECT_REQUIRED_INTENT_HINTS = (
    "模型",
    "字段",
    "数据管理",
    "数据记录",
    "记录",
    "枚举",
    "enum",
    "RBAC",
    "权限",
    "角色",
    "权限包",
    "权限点",
    "项目设置",
    "数据库",
    "集群",
)


def _message_text(message: Any) -> str:
    """Best-effort extraction of textual content from dict/LC message."""
    content: Any = ""
    if isinstance(message, dict):
        content = message.get("content", "")
    else:
        content = getattr(message, "content", "")

    if isinstance(content, str):
        return content
    if isinstance(content, list):
        parts: list[str] = []
        for item in content:
            if isinstance(item, dict):
                text = item.get("text")
                if isinstance(text, str):
                    parts.append(text)
        return "".join(parts)
    return ""


def _is_user_message(message: Any) -> bool:
    """Return whether a message is a user/human message."""
    if isinstance(message, dict):
        return str(message.get("role", "")).lower() == "user"

    msg_type = str(getattr(message, "type", "")).lower()
    role = str(getattr(message, "role", "")).lower()
    return msg_type == "human" or role == "user"


def _latest_user_text(state: AgentState) -> str:
    """Return latest user/human message text from state messages."""
    for msg in reversed(state.get("messages", [])):
        if _is_user_message(msg):
            return _message_text(msg)
    return ""


def _has_tool_call(message: Any, tool_name: str) -> bool:
    """Return whether message.tool_calls contains the given tool name."""
    tool_calls = message.get("tool_calls") if isinstance(message, dict) else getattr(message, "tool_calls", None)
    if not tool_calls:
        return False
    for tc in tool_calls:
        if isinstance(tc, dict) and tc.get("name") == tool_name:
            return True
    return False


def _history_has_tool_call(state: AgentState, tool_names: set[str]) -> bool:
    """Return whether any historical message contains tool calls in tool_names."""
    for msg in state.get("messages", []):
        tool_calls = msg.get("tool_calls") if isinstance(msg, dict) else getattr(msg, "tool_calls", None)
        if not tool_calls:
            continue
        for tc in tool_calls:
            if isinstance(tc, dict) and tc.get("name") in tool_names:
                return True
    return False


def _history_has_tool_call_since_latest_user(state: AgentState, tool_names: set[str]) -> bool:
    """Return whether a tool call exists after the latest user message."""
    for msg in reversed(state.get("messages", [])):
        if _is_user_message(msg):
            return False

        tool_calls = msg.get("tool_calls") if isinstance(msg, dict) else getattr(msg, "tool_calls", None)
        if not tool_calls:
            continue
        for tc in tool_calls:
            if isinstance(tc, dict) and tc.get("name") in tool_names:
                return True
    return False


def _is_direct_navigation_intent(user_text: str) -> bool:
    return any(k in user_text for k in _NAV_INTENT_HINTS)


def _is_list_navigation_intent(user_text: str) -> bool:
    return any(k in user_text for k in _LIST_NAV_INTENT_HINTS)


def _is_project_required_intent(user_text: str) -> bool:
    return any(k in user_text for k in _PROJECT_REQUIRED_INTENT_HINTS)


def _should_force_proposal_on_turn(
    *,
    proposal_available: bool,
    is_direct_nav_intent: bool,
    is_list_nav_intent: bool,
    is_project_required_intent: bool,
    history_has_list_tools: bool,
    history_has_project_list: bool,
    history_has_proposal: bool,
) -> bool:
    """Return whether this turn must force a ui_present_proposal tool call.

    Prevents re-forcing when a proposal has already been presented in the
    conversation — this avoids the infinite loop where the graph re-invokes
    agent_node after frontend tool results and the latest user text still
    matches navigation intent keywords.
    """
    if history_has_proposal or not proposal_available:
        return False
    if is_project_required_intent:
        return history_has_project_list
    return is_direct_nav_intent or (is_list_nav_intent and history_has_list_tools)


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
        latest_user_text = _latest_user_text(state)
        is_direct_nav_intent = _is_direct_navigation_intent(latest_user_text)
        is_list_nav_intent = _is_list_navigation_intent(latest_user_text)
        is_project_required_intent = _is_project_required_intent(latest_user_text)

        # Merge frontend tools (registered by useCopilotAction) into the LLM for this turn.
        # CopilotKit puts them in state["copilotkit"]["actions"] as plain dicts.
        frontend_actions = state.get("copilotkit", {}).get("actions", [])  # type: ignore[union-attr]
        frontend_tool_names = {
            t.get("name", "")
            for t in frontend_actions
            if isinstance(t, dict) and t.get("name")
        }
        proposal_available = "ui_present_proposal" in frontend_tool_names
        history_has_list_tools = _history_has_tool_call_since_latest_user(state, {"list_projects", "list_models"})
        history_has_project_list = _history_has_tool_call_since_latest_user(state, {"list_projects"})
        history_has_proposal = _history_has_tool_call_since_latest_user(state, {"ui_present_proposal"})
        force_proposal_on_this_turn = _should_force_proposal_on_turn(
            proposal_available=proposal_available,
            is_direct_nav_intent=is_direct_nav_intent,
            is_list_nav_intent=is_list_nav_intent,
            is_project_required_intent=is_project_required_intent,
            history_has_list_tools=history_has_list_tools,
            history_has_project_list=history_has_project_list,
            history_has_proposal=history_has_proposal,
        )

        from logging_setup import get_logger as _gl
        _gl().info("agent_node.debug",
                   copilotkit_keys=list((state.get("copilotkit") or {}).keys()),
                   frontend_action_count=len(frontend_actions),
                   frontend_action_names=[t.get("name") for t in frontend_actions if isinstance(t, dict)],
                   latest_user_text=latest_user_text[:120],
                   is_direct_nav_intent=is_direct_nav_intent,
                   is_list_nav_intent=is_list_nav_intent,
                   is_project_required_intent=is_project_required_intent,
                   history_has_list_tools=history_has_list_tools,
                   history_has_project_list=history_has_project_list,
                   history_has_proposal=history_has_proposal,
                   force_proposal_on_this_turn=force_proposal_on_this_turn)

        if history_has_proposal:
            return {"messages": [AIMessage(content="已展示导航候选卡片，请在卡片中选择目标页面。")]}

        forced_proposal_llm = None
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

            proposal_tools = [
                tool
                for tool in extra
                if tool.get("function", {}).get("name") == "ui_present_proposal"
            ]
            if proposal_tools:
                forced_proposal_llm = _base_llm.bind(
                    tools=proposal_tools,
                    tool_choice={"type": "function", "function": {"name": "ui_present_proposal"}},
                )

            if force_proposal_on_this_turn and forced_proposal_llm is not None:
                llm = forced_proposal_llm
            else:
                llm = _base_llm.bind(tools=[*existing, *extra])
        else:
            llm = _base_llm

        if layer == "org":
            context = (
                f"当前在 Org 页面（组织：{org}）。\n"
                "UI 导航工具（只用 ui_present_proposal）：\n"
                "  routeCatalog 和 aiTargets 已通过上下文注入，选取对应条目生成 candidates。\n"
                + (f"\n当前会话项目上下文：**{project}**（用户未明确指定时默认使用此项目）。" if project else "")
            )
        elif layer == "project":
            model_ctx = f"，当前模型：{current_model}（数据库：{current_db}）" if current_model else ""
            context = (
                f"当前在 Project 页面（组织：{org}，项目：**{project}**{model_ctx}）。\n"
                "UI 导航工具（只用 ui_present_proposal）：\n"
                "  routeCatalog 和 aiTargets 已通过上下文注入，选取对应条目生成 candidates。\n\n"
                "数据查询工具：\n"
                "  list_databases、list_models、get_model_fields、query_model、nl2filter\n\n"
                "通知工具：show_toast"
            )
        else:
            project_ctx = f"当前会话项目上下文：**{project}**。" if project else "当前无项目上下文。"
            context = (
                f"当前组织：{org}。{project_ctx}\n"
                "UI 导航工具（只用 ui_present_proposal）：\n"
                "  routeCatalog 和 aiTargets 已通过上下文注入。"
            )

        system_msg = {
            "role": "system",
            "content": (
                "你是 ModelCraft AI 助手（管理员版），帮助租户管理员通过对话完成所有操作。\n\n"
                f"{context}\n\n"
                "═══════════════════════════════════════\n"
                "【第一步：判断用户意图类型，再决定调用什么工具】\n"
                "═══════════════════════════════════════\n\n"
                "▌ 类型 A — 纯导航意图\n"
                "  触发词：去哪里 / 怎么进入 / 帮我跳转 / 在哪里配置 / 找到 X 页面 / 怎么操作 X\n"
                "  ✅ 正确做法：若目标 routeCatalog.requiresProject=false，直接调用 ui_present_proposal\n"
                "  ✅ 若目标 routeCatalog.requiresProject=true，必须先调用 list_projects：\n"
                "     - 有项目：用 ui_present_proposal 让用户选择项目，再跳到该项目下目标页面\n"
                "     - 无项目：用 ui_present_proposal 推荐先去 /org/{org}/workspace 创建项目\n"
                "  ❌ 禁止：没有 list_projects 结果时直接推荐或进入 requiresProject=true 的页面\n\n"
                "▌ 类型 B — 列举可导航资源（项目/模型）\n"
                "  触发词：有哪些项目 / 列出项目 / 当前项目 / 有哪些模型\n"
                "  ✅ 问项目：先调 list_projects，再调 ui_present_proposal，每个项目一个 action_candidate\n"
                "  ✅ 问模型/字段/数据等 project 资源：先调 list_projects，让用户选择项目；选中项目后再查 list_models\n"
                "  ❌ 禁止：只用文字列出，不调 ui_present_proposal\n\n"
                "▌ 类型 C — 纯数据查询（字段值、记录内容）\n"
                "  触发词：查询 / 显示数据 / X 字段的值 / 有多少条记录\n"
                "  ✅ 正确做法：调后端工具查询，直接在对话中返回结果，不需要导航\n\n"
                "═══════════════════════════════════════\n"
                "【Layer Skills 强约束】\n"
                "═══════════════════════════════════════\n"
                "ORG_SKILL / PROJECT_SKILL：\n"
                "1) 任何导航意图必须调用 ui_present_proposal，禁止只输出文字指路\n"
                "2) routeCatalog.requiresProject=true 或后端工具 list_databases/list_models/get_model_fields/query_model 都依赖项目存在\n"
                "3) 执行上述 project 依赖动作前，本轮必须先调用 list_projects；不能依赖历史 sticky projectSlug\n"
                "4) list_projects 为空时，不调用 project 依赖工具；改为推荐用户先创建项目\n"
                "5) 问“有哪些项目/模型”时，若调用了 list_projects/list_models，后续必须再调用 ui_present_proposal 返回可点击候选\n"
                "6) ui.navigate/ui.highlight/ui.guide 只能由前端在用户点击后执行，Agent 不直接执行页面动作\n\n"
                "═══════════════════════════════════════\n"
                "【ui_present_proposal 调用规范】\n"
                "═══════════════════════════════════════\n"
                "⚠️  response 参数必须是 JSON 对象，不能是字符串。结构：\n"
                "{\n"
                "  \"kind\": \"proposal\",\n"
                "  \"proposalId\": \"nav-001\",\n"
                "  \"proposalType\": \"navigation\",\n"
                "  \"message\": \"找到以下页面：\",\n"
                "  \"candidates\": [\n"
                "    {\n"
                "      \"id\": \"c1\",\n"
                "      \"type\": \"action_candidate\",\n"
                "      \"title\": \"项目名称或页面名称\",\n"
                "      \"isPrimary\": true,\n"
                "      \"actions\": [{\"type\": \"ui.navigate\", \"args\": {\"route\": \"/org/ORGNAME/project/SLUG/model-editor\"}}]\n"
                "    }\n"
                "  ]\n"
                "}\n"
                "规则：\n"
                "1. ui.navigate 的 route 从 routeCatalog 的 routeTemplate 派生，\n"
                "   将 :orgName/:projectSlug 替换为实际值\n"
                "2. ui.highlight 的 targetId 必须从 aiTargets 中选取已注册的 id\n"
                "3. 即使只有一个候选项也要包装成 candidates 数组\n\n"
                "═══════════════════════════════════════\n"
                "【其他规则】\n"
                "═══════════════════════════════════════\n"
                "- 删除操作禁止自动执行，必须引导用户在界面手动确认\n"
                "- 完成操作后用 show_toast 通知用户结果"
            ).replace("{project}", project or "（未知项目）"),
        }
        messages = [system_msg] + _copilotkit_context_messages(state) + sanitize_messages(state["messages"])

        response = await llm.ainvoke(messages)

        # Hard guard: retry once with stricter forcing if proposal tool was not called.
        has_ui_proposal_call = _has_tool_call(response, "ui_present_proposal")
        should_retry_for_direct_nav = (
            proposal_available
            and is_direct_nav_intent
            and not is_project_required_intent
            and not has_ui_proposal_call
            and not history_has_proposal
        )
        should_retry_for_project_required_nav = (
            proposal_available
            and is_project_required_intent
            and history_has_project_list
            and not has_ui_proposal_call
            and not history_has_proposal
        )
        should_retry_for_list_nav = (
            proposal_available
            and is_list_nav_intent
            and history_has_list_tools
            and not has_ui_proposal_call
            and not history_has_proposal
        )

        if should_retry_for_direct_nav or should_retry_for_project_required_nav or should_retry_for_list_nav:
            retry_hint = SystemMessage(content=(
                "【SKILL_GUARD_RETRY】你刚才没有按强约束调用 ui_present_proposal。\n"
                "现在必须改为调用 ui_present_proposal 并返回结构化 candidates。\n"
                "禁止仅返回文字解释。"
            ))
            retry_llm = forced_proposal_llm if forced_proposal_llm is not None else llm
            response = await retry_llm.ainvoke([*messages, retry_hint])

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
