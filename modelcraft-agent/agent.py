"""
ModelCraft LangGraph Agent.

Tools receive per-request state via LangGraph InjectedState annotation,
avoiding shared mutable closures and concurrent request cross-contamination.

Available tools:
  - list_projects       列出当前 org 下的所有项目
  - list_models         列出项目下某个数据库的所有模型
  - get_model_fields    获取模型的字段定义
  - query_model         查询模型数据（findMany）
  - nl2filter           自然语言转 filter JSON
"""
import json
from functools import lru_cache
from typing import Annotated, Any

from langchain_openai import ChatOpenAI
from langgraph.graph import StateGraph, END
from langgraph.graph.message import add_messages
from langgraph.prebuilt import ToolNode, InjectedState
from langgraph.checkpoint.memory import MemorySaver
from langchain_core.tools import tool
from typing_extensions import TypedDict

import time
from logging_setup import get_logger

import config
from client.graphql_client import GraphQLClient


# ---------------------------------------------------------------------------
# Agent State
# ---------------------------------------------------------------------------

class AgentState(TypedDict):
    messages: Annotated[list, add_messages]
    authorization: str
    org_name: str
    project_slug: str
    layer: str          # "org" | "project" | ""
    current_model: str  # current model name from route (project layer)
    current_db: str     # current database name from route


# ---------------------------------------------------------------------------
# LLM factory (cached — one instance per process)
# ---------------------------------------------------------------------------

@lru_cache(maxsize=1)
def _get_llm() -> ChatOpenAI:
    return ChatOpenAI(
        model=config.LLM_MODEL,
        api_key=config.LLM_API_KEY,
        base_url=config.LLM_BASE_URL if config.LLM_BASE_URL else None,
        temperature=0,
    )


# ---------------------------------------------------------------------------
# Tool helper
# ---------------------------------------------------------------------------

def _client(state: AgentState) -> GraphQLClient:
    return GraphQLClient(authorization=state["authorization"])


def _log_tool(name: str, args_summary: str):
    log = get_logger()
    log.info("tool.call.start", tool_name=name, args_summary=args_summary[:200])
    return log, time.perf_counter()


def _log_tool_end(log, name: str, start: float, success: bool):
    log.info("tool.call.end", tool_name=name, duration_ms=round((time.perf_counter() - start) * 1000, 2), success=success)


# ---------------------------------------------------------------------------
# Tools
# ---------------------------------------------------------------------------

@tool
async def list_projects(
    state: Annotated[AgentState, InjectedState()],
) -> str:
    """
    List all projects in the current organization.

    Returns:
        JSON array of projects with id, slug, title, description, status.
    """
    log, start = _log_tool("list_projects", f"org={state['org_name']}")
    try:
        result = await _client(state).list_projects(org_name=state["org_name"])
        if "errors" in result and result["errors"]:
            return f"GraphQL error: {result['errors']}"
        projects = result.get("data", {}).get("projects", [])
        _log_tool_end(log, "list_projects", start, True)
        return json.dumps(projects, ensure_ascii=False)
    except Exception:
        log.exception("error", tool_name="list_projects")
        _log_tool_end(log, "list_projects", start, False)
        raise


@tool
async def list_models(
    database_name: str,
    state: Annotated[AgentState, InjectedState()],
) -> str:
    """
    List all data models in a database of the current project.

    Args:
        database_name: Database name, e.g. "maindb"

    Returns:
        JSON array of models with id, name, title, description, databaseName, displayField.
    """
    log, start = _log_tool("list_models", f"db={database_name}")
    try:
        result = await _client(state).list_models(
            org_name=state["org_name"],
            project_slug=state["project_slug"],
            database_name=database_name,
        )
        if "errors" in result and result["errors"]:
            return f"GraphQL error: {result['errors']}"
        data = result.get("data", {}).get("models", {})
        _log_tool_end(log, "list_models", start, True)
        return json.dumps(data, ensure_ascii=False)
    except Exception:
        log.exception("error", tool_name="list_models")
        _log_tool_end(log, "list_models", start, False)
        raise


@tool
async def get_model_fields(
    model_id: str,
    state: Annotated[AgentState, InjectedState()],
) -> str:
    """
    Get the field definitions of a model by its ID.
    Use this before query_model to know what fields are available.

    Args:
        model_id: Model ID (from list_models)

    Returns:
        JSON array of fields with name, title, schemaType, format, isPrimary, isUnique, etc.
    """
    log, start = _log_tool("get_model_fields", f"model_id={model_id}")
    try:
        result = await _client(state).get_model_fields(
            org_name=state["org_name"],
            project_slug=state["project_slug"],
            model_id=model_id,
        )
        if "errors" in result and result["errors"]:
            return f"GraphQL error: {result['errors']}"
        fields = result.get("data", {}).get("fields", [])
        _log_tool_end(log, "get_model_fields", start, True)
        return json.dumps(fields, ensure_ascii=False)
    except Exception:
        log.exception("error", tool_name="get_model_fields")
        _log_tool_end(log, "get_model_fields", start, False)
        raise


@tool
async def query_model(
    db_name: str,
    model_name: str,
    fields: list[str],
    take: int,
    state: Annotated[AgentState, InjectedState()],
    where: dict | None = None,
    skip: int = 0,
) -> str:
    """
    Query records from a ModelCraft data model.

    Args:
        db_name: Database name, e.g. "maindb"
        model_name: Model name, e.g. "users"
        fields: Field names to return, e.g. ["id", "name", "createdAt"]
        take: Max records to return (1-200, default 20)
        where: Optional filter JSON, e.g. {"name": {"contains": "张"}}
        skip: Records to skip for pagination (default 0)

    Returns:
        JSON with items array and totalCount.
    """
    log, start = _log_tool("query_model", str({"db": db_name, "model": model_name, "take": take}))
    try:
        take = max(1, min(take, 200))
        result = await _client(state).find_many(
            org_name=state["org_name"],
            project_slug=state["project_slug"],
            db_name=db_name,
            model_name=model_name,
            fields=fields,
            where=where,
            take=take,
            skip=skip,
        )
        if "errors" in result and result["errors"]:
            return f"GraphQL error: {result['errors']}"
        data = result.get("data", {}).get("findMany", {})
        _log_tool_end(log, "query_model", start, True)
        return json.dumps(data, ensure_ascii=False)
    except Exception:
        log.exception("error", tool_name="query_model")
        _log_tool_end(log, "query_model", start, False)
        raise


@tool
async def nl2filter(
    natural_language: str,
    field_names: list[str],
) -> str:
    """
    Convert a natural language filter description into a ModelCraft where JSON.

    Args:
        natural_language: User's filter intent, e.g. "名字包含张的且年龄大于18"
        field_names: Available field names in the model, e.g. ["id", "name", "age"]

    Returns:
        A JSON string representing the ModelCraft where clause,
        e.g. {"AND": [{"name": {"contains": "张"}}, {"age": {"gt": 18}}]}
    """
    log, start = _log_tool("nl2filter", natural_language[:100])
    try:
        llm = _get_llm()
        system_prompt = f"""You are a filter JSON generator for ModelCraft.
Convert the user's natural language filter description into a valid ModelCraft where JSON.

Available fields: {field_names}

ModelCraft where JSON rules:
- Top level: {{"AND": [...]}}, {{"OR": [...]}}, or a single field condition
- String operators: contains, startsWith, endsWith, equals, not
- Number operators: equals, not, gt, gte, lt, lte
- Boolean: {{"active": {{"equals": true}}}}
- Combined: {{"AND": [{{"name": {{"contains": "张"}}}}, {{"age": {{"gte": 18}}}}]}}

Return ONLY the raw JSON object, no explanation, no markdown."""

        response = await llm.ainvoke([
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": natural_language},
        ])
        raw = response.content.strip()
        json.loads(raw)  # validate
        _log_tool_end(log, "nl2filter", start, True)
        return raw
    except Exception:
        log.exception("error", tool_name="nl2filter")
        _log_tool_end(log, "nl2filter", start, False)
        raise


# ---------------------------------------------------------------------------
# Graph
# ---------------------------------------------------------------------------

_tools = [
    list_projects,
    list_models,
    get_model_fields,
    query_model,
    nl2filter,
]
_tool_node = ToolNode(_tools)


def _build_graph() -> Any:
    llm = _get_llm().bind_tools(_tools)

    async def agent_node(state: AgentState):
        org = state.get("org_name", "")
        project = state.get("project_slug", "")
        layer = state.get("layer", "")
        current_model = state.get("current_model", "")
        current_db = state.get("current_db", "")

        if layer == "org":
            context = (
                f"当前在 Org 页面（组织：{org}）。\n"
                "可用工具：navigate_to_project、navigate_to_settings、open_create_project、highlight_project、list_projects、nl2filter。\n"
                "注意：不可直接调用 list_models、query_model 等 project 级工具。\n"
                "如需操作项目数据，先调用 navigate_to_project(slug) 跳转到对应项目。"
            )
        elif layer == "project":
            model_ctx = f"，当前模型：{current_model}（数据库：{current_db}）" if current_model else ""
            context = (
                f"当前在 Project 页面（组织：{org}，项目：{project}{model_ctx}）。\n"
                "可用工具：navigate_to_org、navigate_to_model、navigate_to_data、"
                "open_create_model、open_create_record、open_edit_record、highlight_records、"
                "list_models、get_model_fields、query_model、nl2filter。\n"
                "写操作规则：open_create_record 和 open_edit_record 只预填表单，用户点 Save 才真正保存。\n"
                "操作前先用 get_model_fields 确认字段名，避免预填错误字段。\n"
                "如需返回 org 级操作，调用 navigate_to_org。"
            )
        else:
            context = (
                f"当前组织：{org}{'，项目：' + project if project else ''}。\n"
                "可用工具取决于当前页面，请先询问用户当前在哪个页面。"
            )

        system_msg = {
            "role": "system",
            "content": (
                "你是 ModelCraft AI 助手，帮助用户通过对话完成所有操作。\n\n"
                f"{context}\n\n"
                "通用原则：\n"
                "- 操作数据前先用 list_models 和 get_model_fields 确认模型和字段存在\n"
                "- 删除操作禁止自动执行，必须引导用户在界面手动确认\n"
                "- 如果用户说筛选或过滤，先用 nl2filter 生成 filter JSON，再告知前端已应用"
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
    graph.add_node("tools", _tool_node)
    graph.set_entry_point("agent")
    graph.add_conditional_edges("agent", should_continue, {"tools": "tools", END: END})
    graph.add_edge("tools", "agent")
    return graph.compile(checkpointer=MemorySaver())


class _LazyGraph:
    _instance = None

    def __getattr__(self, name):
        if self._instance is None:
            type(self)._instance = _build_graph()
        return getattr(self._instance, name)


modelcraft_graph = _LazyGraph()
