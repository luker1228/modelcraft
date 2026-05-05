# ModelCraft — Unified Token System & Workspace Entrypoint

## What This Is

ModelCraft currently has two parallel auth systems — one for platform admins (`mc-developer`) and one for end users (`mc-enduser`). This milestone eliminates that split: all users share a single JWT format with a `scope` claim (`"org"` | `"project"`), the redundant `end_user` GraphQL schema is deleted, and end users get a proper workspace entrypoint (login → project list → workspace CRUD).

## Core Value

Every user — whether org admin or end user — authenticates through one unified token pipeline, and end users can reach their data via the runtime GraphQL workspace without any separate auth path.

## Requirements

### Validated

- ✓ Multi-tenant org isolation (OrgName as tenant key) — existing
- ✓ Project CRUD with `(OrgName, Slug)` composite key — existing
- ✓ RBAC: role assignment, Casbin enforcer, permission cache — existing
- ✓ Runtime GraphQL (dynamic schema, full CRUD + filters + pagination) — existing
- ✓ End-user login at `/api/end-user/{orgSlug}/auth/login` — existing (to be modified)
- ✓ Platform admin login at `/api/auth/login` — existing (to be modified)

### Active

- [ ] **TOKEN-01**: Unify JWT issuer to `mc-platform`; add `scope` claim (`"org"` | `"project"`)
- [ ] **TOKEN-02**: Platform admin login returns `scope=org` JWT (same format as end-user)
- [ ] **TOKEN-03**: End-user login returns `scope=org` JWT (no longer uses `mc-enduser` issuer)
- [ ] **TOKEN-04**: `POST /api/auth/exchange` — swap Org Token for Project Token (`scope=project`, no `projectSlug` in token)
- [ ] **TOKEN-05**: Middleware enforces scope: `scope=org` blocked from `/graphql/org/{orgName}/project/*`; `scope=project` blocked from org-management routes
- [ ] **SCHEMA-01**: Delete `api/graph/end_user/` directory and all associated handlers/resolvers/routes
- [ ] **SCHEMA-02**: Migrate surviving end_user queries to org/project schemas (user/project queries already covered, remove dead ones)
- [ ] **WORKSPACE-01**: Frontend `/u/{orgSlug}/login` page — **用户端 UI**（非管理端）
- [ ] **WORKSPACE-02**: After login, show project list filtered by RBAC membership — **用户端 UI**
- [ ] **WORKSPACE-03**: Project selection triggers exchange to get Project Token (stored as httpOnly cookie, BFF-mediated) — **用户端 UI**
- [ ] **WORKSPACE-04**: Redirect to `/u/{orgSlug}/{projectSlug}/data`; workspace shows only runtime GraphQL CRUD tab — **用户端 UI**（`/u/` 前缀，独立于管理端 `/org/` 前缀）
- [ ] **TEST-01**: BDD scenarios: login→exchange→workspace, scope boundary enforcement (403 on wrong scope), invalid token 401

### Out of Scope

- Service Key (`scope=service_key`) — architecture placeholder, next milestone
- Functional RBAC (admin vs data-operator tab differences) — depends on feature permission system
- Row-level security (RLS) — independent subsystem
- End-user self-registration (`/api/end-user/auth/register`) — keep endpoint as-is, not modified
- Gradual issuer migration — hard cut: `mc-developer` and `mc-enduser` issuers are immediately invalid after deploy

## Context

- The `api/graph/end_user/` schema is a read-only shadow of the project schema with 6 queries. Most map directly to existing org/project schema queries or can be deleted outright.
- Existing `X-Internal-Token` middleware (BFF internal calls) is **not touched** — it's a separate auth path.
- Refresh tokens use short TTL (1h); no DB migration needed — old tokens will simply expire. No `issuer` field migration required.
- Token spec: ES256 (ECDSA P-256), using existing `jwt_signer.go`.

## Constraints

- **Auth**: ES256 JWT only — no symmetric HMAC secrets for access tokens
- **Hard cut**: Old issuers (`mc-developer`, `mc-enduser`) become invalid immediately on deploy — no grace period
- **Token design**: Project Token does NOT contain `projectSlug` — RBAC decides accessible projects server-side
- **Frontend token storage**: Project Token stored in httpOnly cookie; BFF handles the exchange call server-side
- **Never**: Edit `internal/interfaces/graphql/generated/` manually or run `just clean-gql`

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Single issuer `mc-platform` | Eliminates middleware fork, one validation path | — Pending |
| `scope` claim instead of separate token types | Backward-compatible JWT struct extension, no new signing key | — Pending |
| Project Token has no `projectSlug` | RBAC is the access control plane, not the token payload | — Pending |
| Hard cut on old issuers | Simplifies implementation; 1h TTL means users re-auth within an hour | — Pending |
| httpOnly cookie for Project Token | Security over convenience; BFF mediates exchange | — Pending |
| Delete `end_user` schema entirely | Shadow schema was maintenance debt; all data is accessible via project schema | — Pending |
| 双 UI 架构（管理端 `/org/` vs 用户端 `/u/`）| 两类受众需求完全不同；一套 UI 无法同时胜任配置密集的设计时界面和精简的数据操作界面；路由前缀隔离，两套 UI 共享 Design System，业务逻辑和 BFF 完全独立 | — Pending（WORKSPACE 阶段实现用户端） |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-05-05 after initialization*
