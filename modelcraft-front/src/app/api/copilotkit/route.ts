/**
 * CopilotKit Runtime endpoint — Next.js App Router
 *
 * Two agents, each pointing to its own Python endpoint via APISIX:
 *   modelcraft_admin_agent   → APISIX /copilotkit/admin
 *   modelcraft_enduser_agent → APISIX /copilotkit/enduser
 *
 * NOTE: Do NOT pass Authorization in LangGraphHttpAgent.headers.
 * CopilotKit's handle-run already calls extractForwardableHeaders() which
 * forwards the browser's `authorization` header automatically. Adding it
 * here too causes two separate JS object keys (Authorization vs authorization)
 * to be sent as duplicate HTTP headers, which FastAPI joins as
 * "Bearer token, Bearer token".
 *
 * NOTE: Route through APISIX (BACKEND_URL), not directly to the agent service.
 * APISIX handles service discovery, CORS, and request tracing.
 */
import {
  CopilotRuntime,
  ExperimentalEmptyAdapter,
  copilotRuntimeNextJSAppRouterEndpoint,
} from "@copilotkit/runtime"
import { LangGraphHttpAgent } from "@copilotkit/runtime/langgraph"
import { NextRequest } from "next/server"

export const maxDuration = 60

// Route through APISIX — never call the agent service directly.
const APISIX_URL = process.env.BACKEND_URL ?? "http://localhost:9080"

const serviceAdapter = new ExperimentalEmptyAdapter()

export const POST = async (req: NextRequest) => {
  const runtime = new CopilotRuntime({
    agents: {
      modelcraft_admin_agent: new LangGraphHttpAgent({
        url: `${APISIX_URL}/copilotkit/admin`,
        // No headers — CopilotKit's handle-run forwards Authorization automatically.
      }),
      modelcraft_enduser_agent: new LangGraphHttpAgent({
        url: `${APISIX_URL}/copilotkit/enduser`,
        // No headers — CopilotKit's handle-run forwards Authorization automatically.
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
