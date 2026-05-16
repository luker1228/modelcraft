"""
FastAPI entry point for modelcraft-agent.

Exposes a LangGraph AG-UI compatible endpoint at /copilotkit/
consumed by the Next.js CopilotRuntime via LangGraphHttpAgent.

Authorization, org_name, and project_slug are injected into the LangGraph
state by the Next.js BFF before forwarding. Gateway handles JWT validation.
"""
import uvicorn
from fastapi import FastAPI
from ag_ui_langgraph import add_langgraph_fastapi_endpoint
from copilotkit import CopilotKitMiddleware, LangGraphAGUIAgent

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
# ---------------------------------------------------------------------------

add_langgraph_fastapi_endpoint(
    app=app,
    agent=LangGraphAGUIAgent(
        name="modelcraft_agent",
        description="ModelCraft AI 助手：数据查询 + 自然语言筛选",
        graph=modelcraft_graph,
    ),
    path="/copilotkit",
)


# ---------------------------------------------------------------------------
# Dev server
# ---------------------------------------------------------------------------

if __name__ == "__main__":
    uvicorn.run("main:app", host="0.0.0.0", port=config.PORT, reload=True)
