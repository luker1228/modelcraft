# Proposal: Implement Casdoor Runtime Authentication

## Status
- **Change ID**: `implement-casdoor-runtime-auth`
- **Status**: Draft
- **Created**: 2026-01-31
- **Author**: System

## Problem Statement

ModelCraft currently has no authentication/authorization mechanism for the runtime GraphQL API (`/graphql/*`). The runtime API directly exposes customer database operations without user identity verification, creating security risks:

1. **No Access Control**: Anyone can query/mutate data in customer databases
2. **No Multi-tenancy**: Cannot isolate different organizations' data access
3. **No Audit Trail**: Cannot track which user performed which operations
4. **Compliance Risk**: Violates security best practices for enterprise applications

The design-time API (`/api/design/*`) currently doesn't require authentication either, but this proposal focuses on securing the runtime GraphQL API first, as it directly accesses customer data.

## Proposed Solution

Implement a **multi-provider authentication system** with **Casdoor as the default provider**. The architecture follows these principles:

### Core Architecture

1. **AuthProvider Interface**: Abstract interface supporting multiple SSO providers (Casdoor, Keycloak, generic OIDC)
2. **Provider Registry**: Dynamic provider selection based on project-level configuration
3. **Unified Claims**: Standardized user identity structure across all providers
4. **JWT Middleware**: Generic middleware that delegates to the appropriate provider

### Key Design Decisions

**Decision 1: Multi-Provider vs Single Provider**
- **Choice**: Multi-provider architecture with Casdoor as default
- **Rationale**:
  - Different projects may have different SSO requirements (enterprise LDAP, social login, etc.)
  - Future-proof: Easy to add new providers without changing middleware
  - Casdoor fits our use case (Go-native, lightweight, supports enterprise identity sources)

**Decision 2: Project-Level vs Global Authentication**
- **Choice**: Project-level authentication configuration
- **Rationale**:
  - Each project may belong to different organizations with different SSO systems
  - Aligns with ModelCraft's multi-tenant project container model
  - More flexible for SaaS scenarios

**Decision 3: JWT-Only vs Session-Based**
- **Choice**: JWT-only (stateless)
- **Rationale**:
  - Scalable: No server-side session storage needed
  - Standard: Industry-standard OAuth 2.0 / OIDC flow
  - Distributed: Works across multiple runtime instances

**Decision 4: When to Apply Authentication**
- **Choice**: Only runtime GraphQL API (`/graphql/*`), not design-time API
- **Rationale**:
  - Runtime API directly accesses customer databases (higher risk)
  - Design-time API manages metadata (lower risk, can be secured later)
  - Incremental rollout reduces implementation complexity

## Capabilities Introduced

This change introduces four new capabilities:

1. **auth-provider-interface**: Core abstraction for authentication providers
2. **casdoor-provider**: Casdoor-specific implementation
3. **jwt-middleware**: Generic JWT authentication middleware for GraphQL endpoints
4. **project-auth-config**: Project-level authentication configuration management

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     Runtime GraphQL Request                      │
│                   /graphql/:projectId/:cluster/...               │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     JWTAuthMiddleware                            │
│                      (generic, reusable)                         │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     ProviderRegistry                             │
│              GetProvider(projectID) -> AuthProvider              │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                ┌───────────────┼───────────────┐
                ▼               ▼               ▼
┌───────────────────┐ ┌─────────────────┐ ┌──────────────────┐
│CasdoorProvider    │ │KeycloakProvider │ │ OIDCProvider     │
│                   │ │  (future)       │ │  (future)        │
│- GetPublicKey     │ │                 │ │                  │
│- ParseClaims      │ │                 │ │                  │
│- GetSigningMethod │ │                 │ │                  │
└───────────────────┘ └─────────────────┘ └──────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     UnifiedClaims                                │
│  UserID, Username, Email, Organization, ExpiresAt, Raw           │
└─────────────────────────────────────────────────────────────────┘
```

## Impact Analysis

### What Changes

**New Components**:
- `internal/domain/auth/` - Domain layer for authentication concepts
- `internal/app/auth/` - Application services for auth configuration management
- `internal/infrastructure/repository/project_auth_config_repository.go` - Auth config persistence
- `internal/middleware/jwt_auth.go` - JWT authentication middleware
- `db/schema/mysql/04_auth.sql` - Database schema for auth configuration

**Modified Components**:
- `cmd/server/main.go` - Wire up JWT middleware to GraphQL routes
- `configs/config.yaml` - Add Casdoor default configuration section
- `.env.example` - Add Casdoor configuration variables
- `go.mod` - Add JWT library (already present: `github.com/golang-jwt/jwt/v5`)

**What Doesn't Change**:
- Design-time API remains unauthenticated (future enhancement)
- Existing GraphQL schema and resolvers (no breaking changes)
- Runtime query execution logic
- Project, Cluster, Model domain logic

### Backward Compatibility

**Breaking Changes**: None for API contracts, but behavior changes:
- Runtime GraphQL endpoints will require `Authorization: Bearer <token>` header
- Existing clients without tokens will receive 401 Unauthorized
- **Migration Path**: Provide grace period configuration flag `auth.runtime.required: false` (default `true`) to allow gradual client migration

### Security Considerations

**Enhancements**:
- ✅ Prevents unauthorized data access
- ✅ Enables multi-tenant isolation via organization claims
- ✅ Provides audit trail (user identity in logs)
- ✅ Supports enterprise SSO (LDAP, corporate identity providers)

**New Risks**:
- ⚠️ JWT secret compromise: Mitigated by using RS256 (public key verification)
- ⚠️ Token replay attacks: Mitigated by short token expiry (2 hours)
- ⚠️ Misconfigured providers: Mitigated by validation layer and default deny

**Security Requirements**:
- HTTPS must be enforced in production (config validation)
- JWT secrets/certificates must be in environment variables (not config files)
- Token expiry must be validated
- Public key verification must use certificate chains

## Dependencies

### External Dependencies

**Required**:
- Casdoor deployment (Docker Compose provided in docs)
- MySQL (already required for platform database)

**Optional**:
- Enterprise identity sources (LDAP, WeCom, DingTalk) - configured in Casdoor

### Internal Dependencies

**Must Complete First**: None (new feature, isolated implementation)

**Blocks**: Future enhancements (design-time API auth, API key support)

## Risk Assessment

**Technical Risks**:

1. **JWT Library Compatibility** (Low Risk)
   - Mitigation: Use standard `golang-jwt/jwt/v5` (already in go.mod)

2. **Performance Impact** (Low Risk)
   - Concern: JWT verification on every request
   - Mitigation: Public key caching, RS256 verification is fast (~0.1ms)

3. **Provider Misconfiguration** (Medium Risk)
   - Concern: Wrong certificate/endpoint breaks authentication
   - Mitigation: Configuration validation, health check endpoint, detailed error messages

4. **Multi-Tenant Isolation** (Medium Risk)
   - Concern: User from Org A accessing Org B's project
   - Mitigation: Explicit validation logic in middleware (claims.Organization == project.TenantID)

**Organizational Risks**:

1. **Learning Curve** (Low Risk)
   - Team needs to understand Casdoor setup
   - Mitigation: Comprehensive documentation, docker-compose setup

2. **Breaking Existing Clients** (High Risk - Mitigated)
   - All runtime clients must update to include JWT tokens
   - Mitigation: Feature flag for gradual rollout, clear migration guide

## Success Criteria

**Functional**:
- [ ] Runtime GraphQL requests without token receive 401 Unauthorized
- [ ] Valid Casdoor JWT allows GraphQL query execution
- [ ] Invalid/expired JWT is rejected with clear error message
- [ ] User identity is logged in all runtime operations
- [ ] Multi-tenant isolation: User from Org A cannot access Org B's project

**Non-Functional**:
- [ ] JWT verification overhead < 1ms per request
- [ ] System remains available if Casdoor is down (cached public keys)
- [ ] Configuration errors are caught at startup (fail-fast)
- [ ] Zero breaking changes to design-time API

**Documentation**:
- [ ] Casdoor setup guide (already exists in docs/runtime-auth-design.md)
- [ ] API authentication guide for clients
- [ ] Environment variable reference
- [ ] Troubleshooting guide (common JWT errors)

## Alternatives Considered

### Alternative 1: Single Provider (Casdoor Only)
**Pros**: Simpler implementation, fewer abstractions
**Cons**: Inflexible, hard to support different enterprise SSO later
**Decision**: Rejected - Multi-provider costs little extra but provides huge flexibility

### Alternative 2: Session-Based Authentication
**Pros**: Simpler token revocation, familiar pattern
**Cons**: Stateful, doesn't scale horizontally, not standard for APIs
**Decision**: Rejected - JWT is industry standard for stateless APIs

### Alternative 3: API Key Only (No SSO)
**Pros**: Simple, no external dependency
**Cons**: No user identity, no enterprise SSO, harder to manage/rotate keys
**Decision**: Rejected - Not suitable for multi-tenant SaaS, no audit trail

### Alternative 4: Secure Design-Time API First
**Pros**: Defense in depth
**Cons**: Design-time manages metadata (lower risk), larger scope
**Decision**: Deferred - Focus on runtime API first (higher risk), design-time can follow same pattern

## Open Questions

### Clarification Needed

1. **Default Project Authentication**
   - Q: Should the `default` project require authentication?
   - Context: Default project exists for backward compatibility
   - Proposed: Make auth optional for `default` project (config: `auth.default_project.required: false`)

2. **Token Refresh Strategy**
   - Q: Should ModelCraft backend handle token refresh or expect clients to do it?
   - Context: Access tokens expire in 2 hours
   - Proposed: Clients handle refresh (return 401 when expired, client exchanges refresh token)

3. **Health Check Endpoint**
   - Q: Should health check endpoint validate Casdoor connectivity?
   - Proposed: Add `/health/auth` endpoint that checks provider availability

### Investigation Required

1. **JWKS Dynamic Fetching**
   - Need to test Casdoor's JWKS endpoint reliability
   - Fallback strategy if JWKS fetch fails (cached certificate?)

2. **Concurrent Provider Initialization**
   - Registry caches providers per project
   - Need to design thread-safe cache with TTL

## Next Steps

After approval, proceed to **tasks.md** which breaks this into implementable work items.

## Related Changes

- **Future**: Design-time API authentication (will reuse same provider infrastructure)
- **Future**: API key authentication (for service-to-service calls)
- **Future**: Fine-grained permissions/RBAC (currently only authentication, not authorization)

## References

- Design Document: `docs/runtime-auth-design.md`
- Casdoor Documentation: https://casdoor.org/docs/overview
- JWT Best Practices: https://datatracker.ietf.org/doc/html/rfc8725
- OAuth 2.0 RFC: https://datatracker.ietf.org/doc/html/rfc6749
