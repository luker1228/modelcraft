/**
 * BFF proxy for CopilotKit runtime.
 *
 * Forwards requests to the Python modelcraft-agent service.
 *
 * Key responsibilities:
 * 1. Forward Authorization header so Python agent can carry JWT to Gateway
 * 2. Inject authorization, org_name, project_slug into request body state
 *    so LangGraph AgentState receives them (CopilotKit passes input.state to graph)
 * 3. Forward Cookie for refresh token support
 *
 * The CopilotProvider is already configured with:
 *   runtimeUrl="/api/copilotkit"
 *   agent="modelcraft_agent"
 *   properties={{ projectId, projectSlug }} (orgName from layout params — not yet in properties)
 *
 * NOTE: orgName is currently NOT in CopilotProvider properties (only projectId and projectSlug are).
 * org_name in state will be empty string until CopilotProvider is updated to include orgName.
 * See: src/web/components/features/copilot/CopilotProvider.tsx
 */
import { NextRequest, NextResponse } from 'next/server'

export const maxDuration = 60

const AGENT_SERVICE_URL = process.env.AGENT_SERVICE_URL ?? 'http://localhost:8000'

async function handler(req: NextRequest): Promise<NextResponse> {
  const upstreamUrl = `${AGENT_SERVICE_URL}/copilotkit/`

  const headers = new Headers()
  headers.set('Content-Type', req.headers.get('Content-Type') ?? 'application/json')

  // Forward Authorization so Python agent can carry JWT to Gateway
  const authHeader = req.headers.get('Authorization') ?? ''
  if (authHeader) headers.set('Authorization', authHeader)

  // Forward Cookie for refresh token support
  const cookieHeader = req.headers.get('cookie')
  if (cookieHeader) headers.set('Cookie', cookieHeader)

  let body: string | undefined
  if (req.method !== 'GET' && req.method !== 'HEAD') {
    const rawBody = await req.text()
    // Inject auth context into body.state so LangGraph AgentState receives them
    try {
      const parsed = JSON.parse(rawBody) as Record<string, unknown>
      const properties = (parsed.properties ?? {}) as Record<string, string>
      parsed.state = {
        ...((parsed.state ?? {}) as Record<string, unknown>),
        authorization: authHeader,
        // orgName is not yet forwarded by CopilotProvider — will be empty until
        // CopilotProvider is updated to include orgName in properties
        org_name: properties.orgName ?? '',
        project_slug: properties.projectSlug ?? '',
      }
      body = JSON.stringify(parsed)
    } catch {
      // Not JSON (e.g. SSE ping) — forward as-is
      body = rawBody
    }
  }

  let upstreamRes: Response
  try {
    upstreamRes = await fetch(upstreamUrl, {
      method: req.method,
      headers,
      body,
      signal: AbortSignal.timeout(55000),
    })
  } catch {
    return NextResponse.json(
      { error: 'Agent service unreachable' },
      { status: 502 }
    )
  }

  // Pass through the response stream directly — do NOT buffer with arrayBuffer().
  // The agent returns SSE; buffering would block the client until the full response
  // is ready and break streaming UX entirely.
  const response = new NextResponse(upstreamRes.body, {
    status: upstreamRes.status,
    statusText: upstreamRes.statusText,
  })

  upstreamRes.headers.forEach((value, key) => {
    if (
      ['content-encoding', 'content-length', 'transfer-encoding', 'connection'].includes(
        key.toLowerCase()
      )
    )
      return
    response.headers.append(key, value)
  })

  return response
}

export const GET = handler
export const POST = handler
