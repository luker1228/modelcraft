'use client'
/* eslint-disable no-restricted-syntax */

import { generateUUID } from '@/shared/utils/uuid'
import { buildXAction } from './x-action'

// Gateway base URL — empty string means same-origin (when behind a reverse proxy)
export const GATEWAY_URL = ''

// ── Developer endpoint builders ──────────────────────────────────────────────

export function buildRuntimeEndpoint(
  orgName: string,
  projectSlug: string,
  databaseName: string,
  modelName: string
): string {
  return `${GATEWAY_URL}/api/bff/graphql/org/${orgName}/project/${projectSlug}/db/${databaseName}/model/${modelName}`
}

/**
 * Build project-scoped BFF path (no trailing slash — must match Next.js route handler).
 * Shared between ApolloClient factory and operation context so both use the same URI.
 * @see useProjectScopedContext in context.ts
 */
export function buildProjectScopedEndpoint(orgName: string, projectSlug: string): string {
  return `/api/bff/graphql/org/${orgName}/project/${projectSlug}`
}

// ── Shared utilities exposed for the link chains below ──────────────────────

export { generateUUID, buildXAction }
