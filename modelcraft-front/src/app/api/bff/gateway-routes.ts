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
// Developer/tenant GraphQL endpoints proxied through the gateway.
// These routes are valid external entry points and rely on gateway-injected identity headers.

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

// ── End-User GraphQL ───────────────────────────────────────────────────────
// End-user GraphQL endpoints for PAT-based flows and other end-user clients.

export const endUserOrgGraphQL = (orgName: string) =>
  `${gatewayUrl()}/end-user/graphql/org/${orgName}`

export const endUserProjectGraphQL = (orgName: string, projectSlug: string) =>
  `${gatewayUrl()}/end-user/graphql/org/${orgName}/project/${projectSlug}`

export const endUserRuntimeGraphQL = (
  orgName: string,
  projectSlug: string,
  db: string,
  model: string,
) => `${gatewayUrl()}/end-user/graphql/org/${orgName}/project/${projectSlug}/db/${db}/model/${model}`
