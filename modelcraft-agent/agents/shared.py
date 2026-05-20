# modelcraft-agent/agents/shared.py
"""Shared AgentState and logging helpers for ModelCraft agents."""
import time
from typing import Annotated

from langchain_core.messages import AIMessage, ToolMessage
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


def sanitize_messages(messages: list) -> list:
    """Remove orphaned AI tool_call messages that have no corresponding ToolMessage.

    MemorySaver can store incomplete sequences when a tool execution fails mid-run
    (e.g. network error that terminates the graph before the ToolMessage is written).
    DeepSeek (and OpenAI-compatible APIs) reject message sequences where an AI
    message with tool_calls is not immediately followed by tool result messages for
    every tool call ID.

    This function removes any such orphaned AI tool_call messages (and any partial
    ToolMessages that follow them) to produce a sequence the LLM will accept.
    """
    if not messages:
        return messages

    result: list = []
    i = 0
    while i < len(messages):
        msg = messages[i]
        if isinstance(msg, AIMessage) and getattr(msg, "tool_calls", None):
            expected_ids = {tc["id"] for tc in msg.tool_calls if isinstance(tc, dict) and tc.get("id")}
            # Collect consecutive ToolMessages that follow
            j = i + 1
            found_ids: set = set()
            while j < len(messages) and isinstance(messages[j], ToolMessage):
                if hasattr(messages[j], "tool_call_id"):
                    found_ids.add(messages[j].tool_call_id)
                j += 1

            if expected_ids and expected_ids.issubset(found_ids):
                # All tool calls have corresponding results — keep the whole group
                result.extend(messages[i:j])
                i = j
            else:
                # Orphaned tool_call — drop the AI message and any partial ToolMessages
                log = get_logger()
                log.warning(
                    "sanitize_messages.orphan_removed",
                    missing_ids=list(expected_ids - found_ids),
                    message_id=getattr(msg, "id", None),
                )
                i = j  # skip the AI message and any partial tool results
        else:
            result.append(msg)
            i += 1

    return result


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
