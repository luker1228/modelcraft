"""
FastAPI entry point for modelcraft-agent.

Exposes POST /copilotkit — the CopilotKit runtime endpoint consumed by the
Next.js BFF at /api/copilotkit.
"""
import uvicorn
from fastapi import FastAPI, Request
from copilotkit import CopilotKitRemoteEndpoint
from copilotkit.langgraph_agui_agent import LangGraphAgent
from copilotkit.integrations.fastapi import add_fastapi_endpoint

import config
from agent import modelcraft_graph

app = FastAPI(title="modelcraft-agent", version="0.1.0")


# ---------------------------------------------------------------------------
# Health check
# ---------------------------------------------------------------------------

@app.get("/healthz")
async def healthz():
    return {"status": "ok", "service": "modelcraft-agent"}


# ---------------------------------------------------------------------------
# CopilotKit endpoint
# ---------------------------------------------------------------------------

sdk = CopilotKitRemoteEndpoint(
    agents=[
        LangGraphAgent(
            name="modelcraft_agent",
            description="ModelCraft AI 助手：数据查询 + 自然语言筛选",
            graph=modelcraft_graph,
        )
    ]
)

add_fastapi_endpoint(app, sdk, "/copilotkit")


# ---------------------------------------------------------------------------
# Dev server
# ---------------------------------------------------------------------------

if __name__ == "__main__":
    uvicorn.run("main:app", host="0.0.0.0", port=config.PORT, reload=True)
