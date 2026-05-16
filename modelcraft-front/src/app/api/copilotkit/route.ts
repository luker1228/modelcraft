/**
 * CopilotKit Runtime endpoint — Next.js App Router
 *
 * Uses @copilotkit/runtime CopilotRuntime with LangGraphHttpAgent pointing
 * to the Python modelcraft-agent's AG-UI compatible endpoint.
 *
 * Architecture:
 *   Browser → /api/copilotkit (CopilotRuntime) → AGENT_SERVICE_URL/copilotkit (LangGraph AG-UI)
 */
import {
  CopilotRuntime,
  ExperimentalEmptyAdapter,
  copilotRuntimeNextJSAppRouterEndpoint,
} from "@copilotkit/runtime"
import { LangGraphHttpAgent } from "@copilotkit/runtime/langgraph"
import { NextRequest } from "next/server"

export const maxDuration = 60

const AGENT_SERVICE_URL = process.env.AGENT_SERVICE_URL ?? "http://localhost:8000"

const runtime = new CopilotRuntime({
  agents: {
    modelcraft_agent: new LangGraphHttpAgent({
      url: `${AGENT_SERVICE_URL}/copilotkit`,
    }),
  },
})

const serviceAdapter = new ExperimentalEmptyAdapter()

export const POST = async (req: NextRequest) => {
  const { handleRequest } = copilotRuntimeNextJSAppRouterEndpoint({
    runtime,
    serviceAdapter,
    endpoint: "/api/copilotkit",
  })

  return handleRequest(req)
}
