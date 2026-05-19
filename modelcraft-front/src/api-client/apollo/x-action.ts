import { getOperationDefinition } from '@apollo/client/utilities'
import type { DocumentNode } from '@apollo/client'

/**
 * Builds the X-Action header value required by the backend ChiGraphQLActionMiddleware.
 * Format: "{type}:{operationName}", e.g. "query:GetProjects", "mutation:CreateModel".
 *
 * Falls back to the operation name embedded in the document when the call site does not
 * pass operationName explicitly (e.g. useQuery / client.query without explicit operationName).
 */
export function buildXAction(operation: { operationName?: string; query: DocumentNode }): string | null {
  try {
    const def = getOperationDefinition(operation.query) as {
      operation?: string
      name?: { value?: string }
    } | null | undefined
    if (!def) return null
    const opType = def.operation ?? 'query'
    const opName = operation.operationName ?? def.name?.value
    if (!opName) return null
    return `${opType}:${opName}`
  } catch {
    return null
  }
}
