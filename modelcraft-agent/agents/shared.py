# modelcraft-agent/agents/shared.py
"""Shared AgentState and logging helpers for ModelCraft agents."""
import time
from typing import Annotated

from typing_extensions import TypedDict
from langgraph.graph.message import add_messages

from client.graphql_client import GraphQLClient
from logging_setup import get_logger


class AgentState(TypedDict):
    messages: Annotated[list, add_messages]
    authorization: str
    org_name: str
    project_slug: str
    layer: str          # "org" | "project" | ""
    current_model: str  # current model name from route
    current_db: str     # current database name from route


def make_client(state: AgentState) -> GraphQLClient:
    """Create a GraphQL client authenticated with the state's token."""
    return GraphQLClient(authorization=state["authorization"])


def log_tool_start(name: str, args_summary: str):
    """Start a tool log entry, return (log, start_time)."""
    log = get_logger()
    log.info("tool.call.start", tool_name=name, args_summary=args_summary[:200])
    return log, time.perf_counter()


def log_tool_end(log, name: str, start: float, success: bool) -> None:
    """Finish a tool log entry."""
    log.info(
        "tool.call.end",
        tool_name=name,
        duration_ms=round((time.perf_counter() - start) * 1000, 2),
        success=success,
    )
