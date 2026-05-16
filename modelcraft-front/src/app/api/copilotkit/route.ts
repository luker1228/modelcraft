/**
 * CopilotKit Runtime endpoint — Next.js App Router
 *
 * Uses @copilotkit/runtime CopilotRuntime with LangGraphHttpAgent pointing
 * to the Python modelcraft-agent's AG-UI compatible endpoint.
 *
 * Authorization header from the browser is forwarded to the agent so it can
 * authenticate its own GraphQL calls to the backend via the gateway.
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

const serviceAdapter = new ExperimentalEmptyAdapter()

export const POST = async (req: NextRequest) => {
  // Forward the browser's Authorization header to the agent so it can
  // authenticate its GraphQL calls through the gateway.
  const authorization = req.headers.get("Authorization") ?? ""

  const runtime = new CopilotRuntime({
    agents: {
      modelcraft_agent: new LangGraphHttpAgent({
        url: `${AGENT_SERVICE_URL}/copilotkit`,
        headers: authorization ? { Authorization: authorization } : {},
      }),
    },
  })

  const { handleRequest } = copilotRuntimeNextJSAppRouterEndpoint({
    runtime,
    serviceAdapter,
    endpoint: "/api/copilotkit",
  })

  return handleRequest(req)
}
