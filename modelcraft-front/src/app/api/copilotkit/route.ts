/**
 * CopilotKit Runtime endpoint — Next.js App Router
 *
 * Two agents, each pointing to its own Python endpoint:
 *   modelcraft_admin_agent  → AGENT_SERVICE_URL/copilotkit/admin
 *   modelcraft_enduser_agent → AGENT_SERVICE_URL/copilotkit/enduser
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
  const authorization = req.headers.get("Authorization") ?? ""
  const authHeaders: Record<string, string> = authorization ? { Authorization: authorization } : {}

  const runtime = new CopilotRuntime({
    agents: {
      modelcraft_admin_agent: new LangGraphHttpAgent({
        url: `${AGENT_SERVICE_URL}/copilotkit/admin`,
        headers: authHeaders,
      }),
      modelcraft_enduser_agent: new LangGraphHttpAgent({
        url: `${AGENT_SERVICE_URL}/copilotkit/enduser`,
        headers: authHeaders,
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
