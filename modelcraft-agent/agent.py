"""
ModelCraft LangGraph Agent.

Tools receive per-request state via LangGraph InjectedState annotation,
avoiding shared mutable closures and concurrent request cross-contamination.
"""
import json
from functools import lru_cache
from typing import Annotated, Any

from langchain_openai import ChatOpenAI
from langgraph.graph import StateGraph, END
from langgraph.graph.message import add_messages
from langgraph.prebuilt import ToolNode, InjectedState
from langchain_core.tools import tool
from typing_extensions import TypedDict

import config
from client.graphql_client import GraphQLClient

# ---------------------------------------------------------------------------
# Agent State
# ---------------------------------------------------------------------------

class AgentState(TypedDict):
    messages: Annotated[list, add_messages]
    # JWT from the incoming HTTP request — forwarded to Gateway on every tool call.
    authorization: str
    # Runtime context injected by CopilotKit from CopilotProvider properties.
    org_name: str
    project_slug: str


# ---------------------------------------------------------------------------
# LLM factory (cached — one instance per process)
# ---------------------------------------------------------------------------

@lru_cache(maxsize=1)
def _get_llm() -> ChatOpenAI:
    """Return a cached ChatOpenAI instance (lazy-initialized on first call)."""
    return ChatOpenAI(
        model=config.LLM_MODEL,
        api_key=config.LLM_API_KEY,
        base_url=config.LLM_BASE_URL if config.LLM_BASE_URL else None,
        temperature=0,
    )


# ---------------------------------------------------------------------------
# Tools (use InjectedState to safely read per-request context)
# ---------------------------------------------------------------------------

@tool
async def query_model(
    db_name: str,
    model_name: str,
    fields: list[str],
    take: int,
    state: Annotated[AgentState, InjectedState()],
    where: dict | None = None,
) -> str:
    """
    Query records from a ModelCraft data model.

    Args:
        db_name: Database name, e.g. "maindb"
        model_name: Model name, e.g. "users"
        fields: List of field names to return, e.g. ["id", "name", "createdAt"]
        take: Max records to return (default 20, max 200)
        where: Optional ModelCraft filter JSON, e.g. {"name": {"contains": "张"}}

    Returns:
        JSON string with items array and totalCount.
    """
    take = max(1, min(take, 200))
    client = GraphQLClient(authorization=state["authorization"])
    result = await client.find_many(
        org_name=state["org_name"],
        project_slug=state["project_slug"],
        db_name=db_name,
        model_name=model_name,
        fields=fields,
        where=where,
        take=take,
    )
    if "errors" in result and result["errors"]:
        return f"GraphQL error: {result['errors']}"
    data = result.get("data", {}).get("findMany", {})
    return json.dumps(data, ensure_ascii=False)


@tool
async def nl2filter(
    natural_language: str,
    field_names: list[str],
) -> str:
    """
    Convert a natural language filter description into a ModelCraft where JSON.

    Args:
        natural_language: User's filter intent, e.g. "名字包含张的"
        field_names: Available field names in the model, e.g. ["id", "name", "age"]

    Returns:
        A JSON string representing the ModelCraft where clause,
        e.g. {"AND": [{"name": {"contains": "张"}}]}
    """
    llm = _get_llm()
    system_prompt = f"""You are a filter JSON generator for ModelCraft.
Convert the user's natural language filter description into a valid ModelCraft where JSON.

Available fields: {field_names}

ModelCraft where JSON rules:
- Top level: {{"AND": [...]}}, {{"OR": [...]}}, or a single field condition
- String operators: contains, startsWith, endsWith, equals, not
- Number operators: equals, not, gt, gte, lt, lte
- Combined: {{"AND": [{{"name": {{"contains": "张"}}}}, {{"age": {{"gte": 18}}}}]}}

Return ONLY the raw JSON object, no explanation, no markdown code fences."""

    response = await llm.ainvoke([
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": natural_language},
    ])
    raw = response.content.strip()
    # Validate it parses as JSON before returning
    json.loads(raw)
    return raw


# ---------------------------------------------------------------------------
# Graph
# ---------------------------------------------------------------------------

_tools = [query_model, nl2filter]
_tool_node = ToolNode(_tools)


def _build_graph() -> Any:
    """Build and compile the LangGraph StateGraph (called lazily)."""
    llm = _get_llm().bind_tools(_tools)

    async def agent_node(state: AgentState):
        system_msg = {
            "role": "system",
            "content": (
                "你是 ModelCraft AI 助手。你可以帮用户查询数据（query_model）"
                "或将自然语言筛选条件转换为 filter JSON（nl2filter）。"
                f"当前项目：{state.get('org_name', '')}/{state.get('project_slug', '')}。"
                "如果用户说'筛选'、'过滤'类需求，先用 nl2filter 生成 filter JSON，"
                "然后告知用户 filter 已生成，前端会自动应用。"
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
    return graph.compile()


# Lazy initialization — graph is compiled on first access, not at import time
class _LazyGraph:
    _instance = None

    def __getattr__(self, name):
        if self._instance is None:
            type(self)._instance = _build_graph()
        return getattr(self._instance, name)


modelcraft_graph = _LazyGraph()
