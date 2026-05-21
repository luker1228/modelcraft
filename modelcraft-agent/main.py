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
from fastapi.responses import JSONResponse, StreamingResponse
from ag_ui_langgraph.endpoint import RunAgentInput, EventEncoder
from ag_ui.core import EventType
from copilotkit import LangGraphAGUIAgent
from structlog.contextvars import bind_contextvars

import config
from agents.admin_agent import admin_graph
from agents.enduser_agent import enduser_graph
from logging_setup import get_logger, setup_logging
from middleware import ObservabilityMiddleware

class ObservableAgent(LangGraphAGUIAgent):
    """LangGraphAGUIAgent subclass that logs tool errors before make_json_safe
    converts Exception objects to {}.

    Problem: ag_ui_langgraph's make_json_safe calls vars(exception) which returns {}
    for standard Python exceptions, losing the error message completely.
    This subclass intercepts on_tool_error events before serialization and logs
    the real exception type and message.
    """

    def __init__(self, *, name: str, graph, description=None, config=None):
        super().__init__(name=name, graph=graph, description=description, config=config)

    def _dispatch_event(self, event):
        # Intercept RAW on_tool_error events BEFORE super()._dispatch_event
        # calls make_json_safe, which would lose the exception message.
        if getattr(event, "type", None) == EventType.RAW:
            inner = getattr(event, "event", None)
            if isinstance(inner, dict) and inner.get("event") == "on_tool_error":
                err = inner.get("data", {}).get("error")
                tool_name = inner.get("name", "unknown")
                get_logger().error(
                    "tool.error",
                    tool_name=tool_name,
                    error_type=type(err).__name__ if err is not None else "unknown",
                    error_msg=str(err) if err is not None else "no error object",
                )
        return super()._dispatch_event(event)


app = FastAPI(title="modelcraft-agent", version="0.2.0")

setup_logging()
app.add_middleware(ObservabilityMiddleware)

_admin_agent = ObservableAgent(
    name="modelcraft_admin_agent",
    description="ModelCraft AI 助手（管理员版）：项目管理、建模、数据查询",
    graph=admin_graph,
)

_enduser_agent = ObservableAgent(
    name="modelcraft_enduser_agent",
    description="ModelCraft AI 助手（用户版）：数据查询与自然语言筛选",
    graph=enduser_graph,
)


def _safe_dict(obj) -> dict:
    if isinstance(obj, dict):
        return obj
    if hasattr(obj, "model_dump"):
        try:
            return obj.model_dump()
        except Exception:
            return {}
    return {}


def _extract_ai_text(event) -> str:
    data = _safe_dict(event)
    # Common AG-UI style payload: {type, content}
    content = data.get("content")
    if isinstance(content, str) and content.strip():
        return content.strip()

    # Some events nest data under "message"
    message = data.get("message")
    if isinstance(message, dict):
        text = message.get("content")
        if isinstance(text, str) and text.strip():
            return text.strip()

    return ""


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

    # Parse CopilotKit context for layer, projectSlug, availableActions, and page UI actions.
    layer: str = ""
    project_slug_hint: str = ""
    available_actions: list = []
    page_ui_actions: list = []
    current_route: str = ""
    for ctx_item in (input_data.context or []):
        desc = ctx_item.get("description", "") if isinstance(ctx_item, dict) else getattr(ctx_item, "description", "")
        val  = ctx_item.get("value", "")      if isinstance(ctx_item, dict) else getattr(ctx_item, "value", "")
        if desc == "当前 AI 上下文" and val:
            try:
                parsed = json.loads(val)
                layer = parsed.get("layer", "")
                project_slug_hint = parsed.get("projectSlug", "")
                available_actions = parsed.get("availableActions", [])
            except Exception:
                pass
        elif desc == "当前页面可用的 UI 操作（点击 [ACTION:id] chip 可高亮对应元素）" and val:
            try:
                page_ui_actions = json.loads(val)
            except Exception:
                pass
        elif desc == "当前页面信息" and val:
            try:
                parsed = json.loads(val)
                current_route = parsed.get("route", "")
            except Exception:
                pass

    # Also try forwardedProps for currentRoute
    if not current_route:
        props = getattr(input_data, "forwardedProps", None) or {}
        if isinstance(props, dict):
            current_route = props.get("currentRoute", "")

    # layer always reflects current UI page.
    current_state["layer"] = layer

    # available_actions tells the agent which tools are registered on the current page.
    current_state["available_actions"] = available_actions

    # page_ui_actions: UI elements registered for [ACTION:id] chip highlighting.
    current_state["page_ui_actions"] = page_ui_actions

    # current_route for page-aware context.
    current_state["current_route"] = current_route

    # project_slug is sticky: keep existing session value, only use hint as fallback.
    if not current_state.get("project_slug"):
        current_state["project_slug"] = project_slug_hint

    return input_data.model_copy(update={"state": current_state})


def _make_handler(agent: LangGraphAGUIAgent):
    async def handler(input_data: RunAgentInput, request: Request):
        authorization = request.headers.get("Authorization", "")
        if not authorization:
            get_logger().warning("auth.missing_authorization_header", path=str(request.url.path))
            return JSONResponse(
                status_code=401,
                content={"error": "UNAUTHORIZED", "message": "Missing Authorization header"},
            )

        # Expose thread_id/run_id on request.state so ObservabilityMiddleware
        # can include them in request.end without re-parsing the body.
        request.state.thread_id = input_data.thread_id or ""
        request.state.run_id    = input_data.run_id    or ""

        # Also bind to structlog contextvars so tool/graphql log lines
        # emitted inside the async generator carry these IDs automatically.
        bind_contextvars(
            thread_id=request.state.thread_id,
            run_id=request.state.run_id,
        )
        accept_header = request.headers.get("accept")
        encoder = EventEncoder(accept=accept_header)
        input_data = _inject_state(input_data, request)
        request_agent = agent.clone()
        log = get_logger()

        # Events worth logging — everything else is high-frequency noise.
        _LOG_EVENT_TYPES = {
            "RUN_STARTED", "RUN_FINISHED", "RUN_ERROR",
            "STEP_STARTED", "STEP_FINISHED",
            "TOOL_CALL_START", "TOOL_CALL_END", "TOOL_CALL_RESULT",
            "STATE_SNAPSHOT",
        }

        async def event_generator():
            async for event in request_agent.run(input_data):
                data = _safe_dict(event)
                event_type = data.get("type", type(event).__name__)

                if event_type in _LOG_EVENT_TYPES:
                    log.info("agent.event", event_type=event_type)

                ai_text = _extract_ai_text(event)
                if ai_text:
                    log.info("agent.output", content_preview=ai_text[:500])

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
