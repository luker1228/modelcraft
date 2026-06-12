/**
 * Gateway upstream URL builders — single source of truth for all BFF proxy paths.
 *
 * RULES (enforced by ESLint):
 * - ALL BFF route files MUST import upstream paths from this file.
 * - ALL routes MUST use tenant* builders (prefix: /graphql/).
 * - Never write /graphql/ path strings directly in BFF route files.
 */

function gatewayUrl(): string {
  const url = process.env.BACKEND_URL
  if (!url) throw new Error('BACKEND_URL is not set')
  return url
}

// ── Tenant GraphQL ──────────────────────────────────────────────────────────
// Only tenant (developer/admin) tokens may call these endpoints.

export const tenantOrgGraphQL = (orgName: string) =>
  `${gatewayUrl()}/graphql/org/${orgName}/`

export const tenantProjectGraphQL = (orgName: string, projectSlug: string) =>
  `${gatewayUrl()}/graphql/org/${orgName}/project/${projectSlug}/`

export const tenantRuntimeGraphQL = (
  orgName: string,
  projectSlug: string,
  db: string,
  model: string,
) => `${gatewayUrl()}/graphql/org/${orgName}/project/${projectSlug}/db/${db}/model/${model}`

// ── Tenant Auth ─────────────────────────────────────────────────────────────

export const tenantAuthPath = (path: string) =>
  `${gatewayUrl()}/api/tenant/auth/${path}`
