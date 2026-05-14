"""
FastAPI entry point for modelcraft-agent.

Exposes POST /copilotkit — the CopilotKit runtime endpoint consumed by the
Next.js BFF at /api/copilotkit.

Authorization, org_name, and project_slug are injected into the LangGraph state
by the Next.js BFF before forwarding the request here. This service does not
perform any authentication — Gateway handles JWT validation.
"""
import uvicorn
from fastapi import FastAPI
from copilotkit import CopilotKitRemoteEndpoint, LangGraphAGUIAgent
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
#
# agents is a lambda so a fresh agent instance is created per request,
# allowing future context-aware configuration if needed.
# ---------------------------------------------------------------------------

sdk = CopilotKitRemoteEndpoint(
    agents=lambda context: [
        LangGraphAGUIAgent(
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
