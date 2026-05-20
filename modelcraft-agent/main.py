"""
FastAPI entry point for modelcraft-agent.

Exposes a LangGraph AG-UI compatible endpoint at /copilotkit
consumed by the Next.js CopilotRuntime via LangGraphHttpAgent.

Authorization header is extracted from the HTTP request and injected
into the LangGraph state so tools can authenticate GraphQL calls via gateway.
"""
import uvicorn
from fastapi import FastAPI, Request
from fastapi.responses import StreamingResponse
from ag_ui_langgraph.endpoint import RunAgentInput, EventEncoder
from copilotkit import LangGraphAGUIAgent

import config
from agent import modelcraft_graph
from logging_setup import setup_logging
from middleware import ObservabilityMiddleware

app = FastAPI(title="modelcraft-agent", version="0.1.0")

setup_logging()
app.add_middleware(ObservabilityMiddleware)


# ---------------------------------------------------------------------------
# Health check
# ---------------------------------------------------------------------------

@app.get("/healthz")
async def healthz():
    return {"status": "ok", "service": "modelcraft-agent"}


# ---------------------------------------------------------------------------
# LangGraph AG-UI endpoint
# Consumed by Next.js CopilotRuntime via LangGraphHttpAgent(url=.../copilotkit)
#
# We register the endpoint manually (instead of add_langgraph_fastapi_endpoint)
# so we can inject the Authorization header into the LangGraph state before
# the graph runs — tools use state["authorization"] to call backend via gateway.
# ---------------------------------------------------------------------------

_agent = LangGraphAGUIAgent(
    name="modelcraft_agent",
    description="ModelCraft AI 助手：数据查询 + 自然语言筛选",
    graph=modelcraft_graph,
)


@app.post("/copilotkit")
async def copilotkit_endpoint(input_data: RunAgentInput, request: Request):
    accept_header = request.headers.get("accept")
    encoder = EventEncoder(accept=accept_header)

    # Extract Authorization from HTTP header and inject into graph state
    # so tools can authenticate their GraphQL calls through the gateway.
    # Always set the key (even empty) so AgentState["authorization"] never raises KeyError.
    authorization = request.headers.get("Authorization", "")
    current_state = dict(input_data.state) if input_data.state else {}
    current_state["authorization"] = authorization
    input_data = input_data.model_copy(update={"state": current_state})

    request_agent = _agent.clone()

    async def event_generator():
        async for event in request_agent.run(input_data):
            yield encoder.encode(event)

    return StreamingResponse(
        event_generator(),
        media_type=encoder.get_content_type(),
    )


@app.get("/copilotkit/health")
def copilotkit_health():
    return {"status": "ok", "agent": {"name": _agent.name}}


# ---------------------------------------------------------------------------
# Dev server
# ---------------------------------------------------------------------------

if __name__ == "__main__":
    uvicorn.run("main:app", host="0.0.0.0", port=config.PORT, reload=True)
