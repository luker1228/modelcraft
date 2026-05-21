# modelcraft-agent/agents/shared.py
"""Shared AgentState and logging helpers for ModelCraft agents."""
import functools
import time
from typing import Annotated

from langchain_core.messages import AIMessage, ToolMessage
from typing_extensions import TypedDict
from langgraph.graph.message import add_messages

from client.graphql_client import GraphQLClient
from logging_setup import get_logger


class AgentState(TypedDict, total=False):
    messages: Annotated[list, add_messages]
    authorization: str
    org_name: str
    project_slug: str
    layer: str          # "org" | "project" | ""
    current_model: str  # current model name from route
    current_db: str     # current database name from route
    # ag-ui / CopilotKit protocol fields — must be declared so LangGraph
    # persists them across node transitions (undeclared keys are dropped).
    tools: list         # raw frontend tool schemas from RunAgentInput.tools
    copilotkit: dict    # {"actions": [...], "context": [...]} injected by ag-ui


def sanitize_messages(messages: list) -> list:
    """Remove orphaned tool_call / ToolMessage pairs that would cause LLM API 400 errors.

    Two classes of invalid sequences that DeepSeek (and OpenAI-compatible APIs) reject:
      A) AIMessage(tool_calls=[x]) with no matching ToolMessage(tool_call_id=x) following it
         → "insufficient tool messages following tool_calls message"
      B) ToolMessage(tool_call_id=x) with no preceding AIMessage(tool_calls=[x])
         → "Messages with role 'tool' must be a response to a preceding message with 'tool_calls'"

    This can happen when MemorySaver stores incomplete sequences from failed runs.

    Two-pass algorithm:
      Pass 1 — identify which tool_call_ids form complete valid pairs.
      Pass 2 — drop any AIMessage/ToolMessage whose IDs are not in the valid set.
    """
    if not messages:
        return messages

    # Pass 1: find tool_call_ids that have BOTH a tool_calls AI message AND tool result(s)
    valid_tool_call_ids: set = set()
    i = 0
    while i < len(messages):
        msg = messages[i]
        if isinstance(msg, AIMessage) and getattr(msg, "tool_calls", None):
            expected = {tc["id"] for tc in msg.tool_calls if isinstance(tc, dict) and tc.get("id")}
            j = i + 1
            found: set = set()
            while j < len(messages) and isinstance(messages[j], ToolMessage):
                if hasattr(messages[j], "tool_call_id"):
                    found.add(messages[j].tool_call_id)
                j += 1
            if expected and expected.issubset(found):
                valid_tool_call_ids.update(expected)
        i += 1

    # Pass 2: rebuild message list, dropping anything that isn't in a valid pair
    log = get_logger()
    result: list = []
    for msg in messages:
        if isinstance(msg, AIMessage) and getattr(msg, "tool_calls", None):
            ids = {tc["id"] for tc in msg.tool_calls if isinstance(tc, dict) and tc.get("id")}
            if ids and ids.issubset(valid_tool_call_ids):
                result.append(msg)
            else:
                log.warning("sanitize_messages.orphan_ai_removed", missing_ids=list(ids - valid_tool_call_ids))
        elif isinstance(msg, ToolMessage):
            tid = getattr(msg, "tool_call_id", None)
            if tid and tid in valid_tool_call_ids:
                result.append(msg)
            else:
                log.warning("sanitize_messages.orphan_tool_removed", tool_call_id=tid)
        else:
            result.append(msg)

    return result


def make_client(state: "AgentState") -> GraphQLClient:
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


def log_tool_errors(func):
    """Decorator for async @tool functions: handles start/end timing and error logging.

    Replaces the repetitive try/except pattern in every tool. Usage:

        @tool
        @log_tool_errors
        async def my_tool(param: str, state: Annotated[AgentState, InjectedState()]) -> str:
            result = await do_something(param)
            return json.dumps(result)

    The decorator automatically:
    - logs tool.call.start with a summary of non-state args
    - logs tool.call.end with success/failure and duration
    - logs tool.error with error_type + error_msg on exception (before re-raising)

    Note: must be applied AFTER @tool so that @tool sees the original signature
    and InjectedState annotations are preserved via functools.wraps.
    """
    @functools.wraps(func)
    async def wrapper(*args, **kwargs):
        tool_name = func.__name__

        # Build a short summary from non-state keyword args (state is injected, not LLM-provided)
        summary_parts = [
            f"{k}={str(v)[:60]}"
            for k, v in kwargs.items()
            if k != "state"
        ]
        # Also capture positional args that aren't the InjectedState (dicts)
        if not summary_parts and args:
            summary_parts = [str(a)[:60] for a in args if not isinstance(a, dict)]
        summary = ", ".join(summary_parts)

        log, start = log_tool_start(tool_name, summary)
        try:
            result = await func(*args, **kwargs)
            log_tool_end(log, tool_name, start, True)
            return result
        except Exception as exc:
            log.exception(
                "tool.error",
                tool_name=tool_name,
                error_type=type(exc).__name__,
                error_msg=str(exc),
            )
            log_tool_end(log, tool_name, start, False)
            raise

    return wrapper
