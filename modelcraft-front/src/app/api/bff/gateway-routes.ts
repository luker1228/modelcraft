/**
 * Gateway upstream URL builders — single source of truth for all BFF proxy paths.
 *
 * RULES (enforced by ESLint):
 * - ALL BFF route files MUST import upstream paths from this file.
 * - End-user BFF routes MUST use endUser* builders (prefix: /end-user/graphql/).
 * - Tenant BFF routes MUST use tenant* builders (prefix: /graphql/).
 * - Never write /graphql/ path strings directly in BFF route files.
 */

const GATEWAY_URL = process.env.BACKEND_URL
if (!GATEWAY_URL) throw new Error('BACKEND_URL is not set')

// ── Tenant GraphQL ──────────────────────────────────────────────────────────
// Only tenant (developer/admin) tokens may call these endpoints.

export const tenantOrgGraphQL = (orgName: string) =>
  `${GATEWAY_URL}/graphql/org/${orgName}`

export const tenantProjectGraphQL = (orgName: string, projectSlug: string) =>
  `${GATEWAY_URL}/graphql/org/${orgName}/project/${projectSlug}`

export const tenantRuntimeGraphQL = (
  orgName: string,
  projectSlug: string,
  db: string,
  model: string,
) => `${GATEWAY_URL}/graphql/org/${orgName}/project/${projectSlug}/db/${db}/model/${model}`

// ── End-User GraphQL ────────────────────────────────────────────────────────
// Only end-user tokens may call these endpoints.
// APISIX route: enduser-graphql → enforces aud=end_user, injects X-User-ID + X-User-Type:end_user

export const endUserOrgGraphQL = (orgName: string) =>
  `${GATEWAY_URL}/end-user/graphql/org/${orgName}`

export const endUserProjectGraphQL = (orgName: string, projectSlug: string) =>
  `${GATEWAY_URL}/end-user/graphql/org/${orgName}/project/${projectSlug}`

// ── End-User Auth ───────────────────────────────────────────────────────────

export const endUserAuthPath = (path: string) =>
  `${GATEWAY_URL}/api/end-user/auth/${path}`

// ── Tenant Auth ─────────────────────────────────────────────────────────────

export const tenantAuthPath = (path: string) =>
  `${GATEWAY_URL}/api/tenant/auth/${path}`
