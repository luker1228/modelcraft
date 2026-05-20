"""
FastAPI entry point for modelcraft-agent.

Two agents on one service, each with its own endpoint:
  POST /copilotkit/admin   → modelcraft_admin_agent  (tenant admins)
  POST /copilotkit/enduser → modelcraft_enduser_agent (end users)

CopilotKit runtime (route.ts) maps agent name → URL; routing is done
there, not here. Each endpoint simply injects Authorization and runs
the appropriate graph.
"""
import json
import uvicorn
from fastapi import FastAPI, Request
from fastapi.responses import StreamingResponse
from ag_ui_langgraph.endpoint import RunAgentInput, EventEncoder
from copilotkit import LangGraphAGUIAgent

import config
from agents.admin_agent import admin_graph
from agents.enduser_agent import enduser_graph
from logging_setup import setup_logging
from middleware import ObservabilityMiddleware

app = FastAPI(title="modelcraft-agent", version="0.2.0")

setup_logging()
app.add_middleware(ObservabilityMiddleware)

_admin_agent = LangGraphAGUIAgent(
    name="modelcraft_admin_agent",
    description="ModelCraft AI 助手（管理员版）：项目管理、建模、数据查询",
    graph=admin_graph,
)

_enduser_agent = LangGraphAGUIAgent(
    name="modelcraft_enduser_agent",
    description="ModelCraft AI 助手（用户版）：数据查询与自然语言筛选",
    graph=enduser_graph,
)


@app.get("/healthz")
async def healthz():
    return {"status": "ok", "service": "modelcraft-agent"}


def _inject_state(input_data: RunAgentInput, request: Request) -> RunAgentInput:
    """Inject trusted context into LangGraph state before every run.

    Security model:
      org_name     — ONLY from X-Org-Name header (injected by APISIX from JWT).
                     Client-provided value is NEVER trusted.
      project_slug — Sticky within the session.  On each request:
                       1. Keep the existing state value if already set.
                       2. Otherwise fall back to the UI hint from CopilotKit
                          context (represents the page the user is currently on).
                       3. Empty string when neither is available — agent will ask
                          the user to choose a project when one is required.
      layer        — Always taken from CopilotKit context (reflects current UI page).
      authorization — From HTTP Authorization header (forwarded by CopilotKit).
    """
    authorization = request.headers.get("Authorization", "")

    # org_name: trust ONLY the APISIX-injected header (derived from verified JWT).
    org_name = request.headers.get("X-Org-Name", "")

    current_state = dict(input_data.state) if input_data.state else {}
    current_state["authorization"] = authorization
    current_state["org_name"] = org_name  # always overwrite with trusted value

    # Parse CopilotKit context for layer and projectSlug hint.
    layer: str = ""
    project_slug_hint: str = ""
    for ctx_item in (input_data.context or []):
        desc = ctx_item.get("description", "") if isinstance(ctx_item, dict) else getattr(ctx_item, "description", "")
        val  = ctx_item.get("value", "")      if isinstance(ctx_item, dict) else getattr(ctx_item, "value", "")
        if desc == "当前 AI 上下文" and val:
            try:
                parsed = json.loads(val)
                layer = parsed.get("layer", "")
                project_slug_hint = parsed.get("projectSlug", "")
            except Exception:
                pass

    # layer always reflects current UI page.
    current_state["layer"] = layer

    # project_slug is sticky: keep existing session value, only use hint as fallback.
    if not current_state.get("project_slug"):
        current_state["project_slug"] = project_slug_hint

    return input_data.model_copy(update={"state": current_state})


def _make_handler(agent: LangGraphAGUIAgent):
    async def handler(input_data: RunAgentInput, request: Request):
        accept_header = request.headers.get("accept")
        encoder = EventEncoder(accept=accept_header)
        input_data = _inject_state(input_data, request)
        request_agent = agent.clone()

        async def event_generator():
            async for event in request_agent.run(input_data):
                yield encoder.encode(event)

        return StreamingResponse(event_generator(), media_type=encoder.get_content_type())
    return handler


app.post("/copilotkit/admin")(_make_handler(_admin_agent))
app.post("/copilotkit/enduser")(_make_handler(_enduser_agent))


@app.get("/copilotkit/health")
def copilotkit_health():
    return {
        "status": "ok",
        "agents": [
            {"name": _admin_agent.name, "endpoint": "/copilotkit/admin"},
            {"name": _enduser_agent.name, "endpoint": "/copilotkit/enduser"},
        ],
    }


if __name__ == "__main__":
    uvicorn.run("main:app", host="0.0.0.0", port=config.PORT, reload=True)
