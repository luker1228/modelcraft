## 1. Auth Model Foundation

- [ ] 1.1 Define `tenant token` and `project access token` claims, issuers, and server-side principal mapping rules.
- [ ] 1.2 Implement a unified project auth middleware that accepts both token families and injects `ProjectPrincipal`.
- [ ] 1.3 Enforce scope boundaries so tenant/org APIs reject `project access token` while project APIs accept both token families.

## 2. Backend Project API Convergence

- [ ] 2.1 Inventory existing project-capable GraphQL/runtime endpoints and decide which end-user project capabilities move into the unified project schema first.
- [ ] 2.2 Adapt project-level GraphQL/runtime handlers to authorize by `ProjectPrincipal` instead of route-specific token semantics.
- [ ] 2.3 Define the migration strategy for existing `/graphql/end-user/...` project capabilities so they become compatibility facades instead of long-term primary contracts.

## 3. Project Authorization Model

- [ ] 3.1 Add or formalize the protected built-in `project_admin` role for every project.
- [ ] 3.2 Introduce project function permissions and map them to page/feature visibility such as `model.design`, `project.role.manage`, and `data.access`.
- [ ] 3.3 Keep data permissions as a separate layer and wire project role assignments to expand both function permissions and data grants for project principals.

## 4. Frontend Workspace and Login Flow

- [ ] 4.1 Update tenant login and project-access login clients so their responses align with the new token semantics and workspace entry contract.
- [ ] 4.2 Implement unified `/org/{orgName}/workspace` routing with `tenant mode` and `project-access mode`.
- [ ] 4.3 In `project-access mode`, show only the authorized project list and project pages, hide the global sidebar, and gate page visibility by function permissions.

## 5. Verification and Rollout

- [ ] 5.1 Add backend tests for token family acceptance, principal expansion, and tenant/project API scope rejection.
- [ ] 5.2 Add frontend/integration tests for workspace mode switching, project selection, and page visibility gating.
- [ ] 5.3 Document the migration path from standalone end-user shell/schema usage to unified workspace + unified project API usage.
