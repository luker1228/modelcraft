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
 *
 * NOTE: Do NOT pass Authorization in LangGraphHttpAgent.headers.
 * CopilotKit's handle-run handler already calls extractForwardableHeaders()
 * which forwards the browser's `authorization` header automatically.
 * Adding it here as well results in two separate header keys (Authorization vs
 * authorization — different JS object keys, same HTTP header name) being sent
 * to the Python agent, which FastAPI joins as "Bearer token, Bearer token".
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
  const runtime = new CopilotRuntime({
    agents: {
      modelcraft_agent: new LangGraphHttpAgent({
        url: `${AGENT_SERVICE_URL}/copilotkit`,
        // No headers here — CopilotKit's handle-run forwards Authorization automatically.
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
