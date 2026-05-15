"""
FastAPI entry point for modelcraft-agent.

Exposes POST /copilotkit — the CopilotKit runtime endpoint consumed by the
Next.js BFF at /api/copilotkit.

Authorization, org_name, and project_slug are injected into the LangGraph state
by the Next.js BFF before forwarding the request here. This service does not
perform any authentication — Gateway handles JWT validation.
"""
import uuid
import json
import asyncio
from typing import Optional, List, Any, AsyncGenerator

import uvicorn
from fastapi import FastAPI
from langchain_core.messages import HumanMessage, AIMessage, ToolMessage
from copilotkit import CopilotKitRemoteEndpoint, Agent
from copilotkit.integrations.fastapi import add_fastapi_endpoint
from copilotkit.action import ActionDict
from copilotkit.types import Message, MetaEvent

import config
from agent import modelcraft_graph
from logging_setup import setup_logging
from middleware import ObservabilityMiddleware

app = FastAPI(title="modelcraft-agent", version="0.1.0")

# Observability: must be added before any routes are registered.
# Middleware is applied in reverse order of add_middleware() calls;
# ObservabilityMiddleware is outermost so it wraps all requests.
setup_logging()
app.add_middleware(ObservabilityMiddleware)


# ---------------------------------------------------------------------------
# Health check
# ---------------------------------------------------------------------------

@app.get("/healthz")
async def healthz():
    return {"status": "ok", "service": "modelcraft-agent"}


# ---------------------------------------------------------------------------
# ModelCraft LangGraph Agent wrapper
#
# Wraps modelcraft_graph in a proper copilotkit.Agent subclass so that
# add_fastapi_endpoint can correctly call dict_repr() and execute().
# ---------------------------------------------------------------------------

def _to_langchain_messages(messages: List[Message]) -> list:
    """Convert CopilotKit message dicts to LangChain message objects."""
    result = []
    for msg in messages:
        role = msg.get("role", "")
        content = msg.get("content", "") or ""
        if role == "user":
            result.append(HumanMessage(content=content))
        elif role == "assistant":
            result.append(AIMessage(content=content))
    return result


async def _stream_graph(thread_id: str, state: dict, lc_messages: list) -> AsyncGenerator[str, None]:
    """Run the LangGraph and yield newline-delimited JSON events."""
    initial = {
        **state,
        "messages": lc_messages,
    }
    config_run = {"configurable": {"thread_id": thread_id}}

    async for event in modelcraft_graph.astream_events(initial, config=config_run, version="v2"):
        kind = event.get("event", "")
        # Text tokens from the LLM
        if kind == "on_chat_model_stream":
            chunk = event.get("data", {}).get("chunk", {})
            content = getattr(chunk, "content", "") or ""
            if content:
                yield json.dumps({"type": "text", "content": content}) + "\n"
        # Agent finished
        elif kind == "on_chain_end" and event.get("name") == "LangGraph":
            output = event.get("data", {}).get("output", {})
            messages_out = output.get("messages", [])
            last = messages_out[-1] if messages_out else None
            if last and hasattr(last, "content") and last.content:
                # Emit final state so CopilotKit can update thread
                yield json.dumps({"type": "state_snapshot", "state": {}}) + "\n"


class ModelCraftAgent(Agent):
    """
    Wraps modelcraft_graph as a copilotkit.Agent so add_fastapi_endpoint works.
    """

    def execute(
        self,
        *,
        state: dict,
        config: Optional[dict] = None,
        messages: List[Message],
        thread_id: str,
        actions: Optional[List[ActionDict]] = None,
        meta_events: Optional[List[MetaEvent]] = None,
        **kwargs,
    ) -> Any:
        """Execute the agent and return a streaming response."""
        from fastapi.responses import StreamingResponse

        lc_messages = _to_langchain_messages(messages)
        tid = thread_id or str(uuid.uuid4())

        return StreamingResponse(
            _stream_graph(tid, state, lc_messages),
            media_type="application/json",
        )

    async def get_state(self, *, thread_id: str):
        """Return current thread state."""
        return {
            "threadId": thread_id or "",
            "threadExists": False,
            "state": {},
            "messages": [],
        }


# ---------------------------------------------------------------------------
# CopilotKit endpoint
# ---------------------------------------------------------------------------

sdk = CopilotKitRemoteEndpoint(
    agents=lambda context: [
        ModelCraftAgent(
            name="modelcraft_agent",
            description="ModelCraft AI 助手：数据查询 + 自然语言筛选",
        )
    ]
)

add_fastapi_endpoint(app, sdk, "/copilotkit")


# ---------------------------------------------------------------------------
# Dev server
# ---------------------------------------------------------------------------

if __name__ == "__main__":
    uvicorn.run("main:app", host="0.0.0.0", port=config.PORT, reload=True)
