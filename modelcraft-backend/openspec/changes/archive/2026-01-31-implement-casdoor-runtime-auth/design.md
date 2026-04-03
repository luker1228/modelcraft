# Design: Casdoor Runtime Authentication Architecture

## Purpose

This document captures the architectural decisions and design rationale for implementing JWT-based authentication for ModelCraft's runtime GraphQL API using Casdoor as the default identity provider.

## Problem Context

ModelCraft's runtime GraphQL API (`/graphql/*`) directly accesses customer databases without any authentication mechanism. This creates several security and operational risks:

1. **Unauthorized Access**: Anyone with network access can query/mutate customer data
2. **No Audit Trail**: Cannot track which user performed which operations
3. **No Multi-Tenancy**: Cannot isolate different organizations' data access
4. **Compliance Violations**: Fails security audits for enterprise deployments

The system needs authentication that:
- Works across distributed runtime instances (stateless)
- Supports multiple identity providers (different organizations use different SSO)
- Integrates with enterprise identity sources (LDAP, corporate SSO)
- Minimizes performance overhead (<1ms per request)

## Architectural Decisions

### Decision 1: JWT-Based Stateless Authentication

**Options Considered**:
1. **Session-Based Authentication**: Server-side session storage with cookies
2. **JWT Tokens**: Stateless bearer tokens
3. **API Keys**: Long-lived static credentials

**Decision**: JWT tokens (OAuth 2.0 / OIDC pattern)

**Rationale**:
- **Scalability**: No server-side state, works across multiple instances without session replication
- **Standard**: Industry-standard OAuth 2.0 flow, compatible with all SSO providers
- **Performance**: Verification is fast (RS256 ~0.1-0.5ms), no database lookup per request
- **Distributed**: Works seamlessly in Kubernetes multi-pod deployments

**Trade-offs**:
- ❌ Cannot revoke tokens before expiry (mitigation: short token lifetime, refresh tokens)
- ❌ Tokens can be large (mitigation: use compact JWT format, no unnecessary claims)
- ✅ Simpler than session management (no Redis/database for sessions)
- ✅ Native support in all modern identity providers

### Decision 2: Multi-Provider Architecture

**Options Considered**:
1. **Single Provider (Casdoor Only)**: Hard-code Casdoor integration
2. **Multi-Provider with Interface**: Abstract AuthProvider interface
3. **Generic OIDC Only**: Only support standard OIDC, no provider-specific logic

**Decision**: Multi-provider architecture with Casdoor as default implementation

**Rationale**:
- **Flexibility**: Different projects may need different identity providers (enterprise LDAP, social login, etc.)
- **Future-Proof**: Easy to add Keycloak, Okta, Auth0 without changing middleware
- **Cost**: Minimal additional complexity (interface + factory pattern), huge long-term flexibility
- **Real-World**: Large enterprises often have multiple identity systems during mergers/transitions

**Implementation Pattern**:
```go
// Single interface, multiple implementations
type AuthProvider interface {
    GetPublicKey(ctx) -> interface{}
    ParseClaims(jwt.MapClaims) -> *UnifiedClaims
    GetSigningMethod() -> string
    Type() -> string
}

// Factory creates provider from config
func createProvider(config *ProjectAuthConfig) AuthProvider {
    switch config.Provider {
    case "casdoor":   return NewCasdoorProvider(config.Config)
    case "keycloak":  return NewKeycloakProvider(config.Config)
    case "oidc":      return NewOIDCProvider(config.Config)
    }
}
```

**Trade-offs**:
- ✅ Zero cost for simple use cases (just use default Casdoor)
- ✅ Pays for complexity only when needed (multi-provider scenarios)
- ❌ Slightly more code than hard-coded Casdoor
- ✅ Easier to test (mock providers)

### Decision 3: Project-Level Configuration

**Options Considered**:
1. **Global Authentication**: Single SSO for entire platform
2. **Project-Level Configuration**: Each project can configure its own provider
3. **User-Level Configuration**: Users choose their preferred login method

**Decision**: Project-level authentication configuration

**Rationale**:
- **Multi-Tenant SaaS**: Different organizations (projects) use different identity systems
- **Isolation**: Aligns with ModelCraft's project-based resource isolation model
- **Enterprise Requirements**: Large orgs often have project-specific identity requirements (dev vs prod, different business units)

**Example Scenario**:
- Project "ecommerce-prod" uses enterprise LDAP (via Casdoor)
- Project "ecommerce-dev" uses social login (via generic OIDC)
- Project "internal-tools" uses corporate Azure AD (via Keycloak)

**Data Model**:
```sql
CREATE TABLE project_auth_configs (
    id           BIGINT PRIMARY KEY,
    project_id   VARCHAR(64) UNIQUE,  -- One config per project
    provider     VARCHAR(32),         -- casdoor | keycloak | oidc
    enabled      BOOLEAN,
    config       JSON,                -- Provider-specific settings
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);
```

**Trade-offs**:
- ✅ Maximum flexibility (each project independent)
- ❌ More complex than global config (need config management UI)
- ✅ Better security (breach in one project doesn't affect others)

### Decision 4: Casdoor as Default Provider

**Options Considered**:
1. **Casdoor**: Open-source Go-based IAM platform
2. **Keycloak**: Java-based, industry-standard IAM
3. **Ory Kratos**: Modern Go-based, headless identity system
4. **Auth0 / Okta**: Commercial SaaS solutions

**Decision**: Casdoor as default, with architecture supporting others

**Comparison Table**:

| Criterion | Casdoor | Keycloak | Ory Kratos | Auth0/Okta |
|-----------|---------|----------|------------|------------|
| Language | Go | Java | Go | N/A (SaaS) |
| Resource Usage | ~100MB | ~1GB | ~300MB | N/A |
| Chinese Identity Sources | Built-in (WeCom, DingTalk) | Plugin required | Not supported | Limited |
| OAuth 2.0 / OIDC | ✅ | ✅ | ✅ | ✅ |
| LDAP Support | ✅ | ✅ | ❌ (external) | ✅ |
| Admin UI | ✅ Excellent | ✅ Excellent | ❌ Headless | ✅ Excellent |
| Go SDK | ✅ Official | ⚠️ Community | ✅ Official | ✅ Official |
| Self-Hosted | ✅ | ✅ | ✅ | ❌ |
| Cost | Free | Free | Free | Paid per MAU |

**Rationale**:
- **Go Stack Alignment**: Same language as ModelCraft, easier debugging and integration
- **Resource Efficiency**: Critical for containerized deployments (100MB vs 1GB matters)
- **Chinese Enterprise**: Built-in WeCom/DingTalk/Feishu support (major ModelCraft target market)
- **Learning Curve**: Simpler than Keycloak's extensive configuration options
- **Deployment**: Single binary + MySQL (vs Keycloak's JVM + DB + complex realm config)

**Trade-offs**:
- ❌ Smaller community than Keycloak (fewer plugins, less StackOverflow content)
- ✅ Faster startup and lower memory footprint
- ✅ Native Chinese UI and documentation
- ❌ Less mature than Keycloak (but actively developed, 10k+ GitHub stars)

### Decision 5: RS256 Public Key Verification

**Options Considered**:
1. **RS256 (RSA Signature)**: Asymmetric, public key verification
2. **HS256 (HMAC)**: Symmetric, shared secret
3. **ES256 (ECDSA)**: Asymmetric, elliptic curve

**Decision**: RS256 for Casdoor, interface supports all methods

**Rationale**:
- **Security**: Public key verification means ModelCraft never needs Casdoor's private key
- **Key Rotation**: Casdoor can rotate signing keys without updating ModelCraft (JWKS endpoint)
- **Standard**: Most common algorithm for OAuth 2.0 / OIDC (GitHub, Google, Azure AD use RS256)
- **Performance**: Fast enough (~0.1-0.5ms per verification), RSA-2048 is sufficient

**Architecture**:
```
┌─────────────┐                    ┌──────────────┐
│  Casdoor    │                    │ ModelCraft   │
│             │                    │              │
│ Private Key │ ──signs JWT──────► │ Public Key   │
│ (secret)    │                    │ (verifies)   │
└─────────────┘                    └──────────────┘
```

**Trade-offs**:
- ✅ ModelCraft can never forge tokens (doesn't have private key)
- ✅ Key compromise in ModelCraft doesn't allow token generation
- ❌ Slightly slower than HS256 (but negligible: 0.5ms vs 0.1ms)
- ✅ Supports key rotation via JWKS

### Decision 6: UnifiedClaims Abstraction

**Problem**: Different SSO providers use different JWT claim structures:
- Casdoor: `owner` (organization), `name` (username)
- Keycloak: `azp` (client/tenant), `preferred_username`
- Generic OIDC: `org`, `name`, highly variable

**Decision**: Define `UnifiedClaims` struct that normalizes across providers

**Structure**:
```go
type UnifiedClaims struct {
    UserID       string                 // sub (universally required)
    Username     string                 // name | preferred_username | email
    Email        string                 // email
    Organization string                 // owner | azp | org (for multi-tenancy)
    Roles        []string               // roles (future RBAC)
    ExpiresAt    time.Time              // exp
    Raw          map[string]interface{} // Original claims
}
```

**Rationale**:
- **Loose Coupling**: Middleware doesn't know about provider-specific claims
- **Consistency**: GraphQL resolvers always work with same structure
- **Extensibility**: Raw map allows provider-specific extensions
- **Type Safety**: Go structs with validation vs map[string]interface{}

**Provider Responsibility**: Each provider implements `ParseClaims(jwt.MapClaims) -> *UnifiedClaims`

**Trade-offs**:
- ✅ Middleware is 100% provider-agnostic (testable with mocks)
- ✅ Easy to add new providers (just map their claims)
- ❌ Slight mapping overhead (but trivial: ~0.01ms)
- ✅ Type-safe access to common claims

### Decision 7: Provider Registry with Caching

**Problem**: Each GraphQL request needs to:
1. Load auth config from database
2. Create AuthProvider instance
3. Parse public key certificate

Doing this on every request would add 10-50ms latency.

**Decision**: ProviderRegistry with in-memory cache

**Architecture**:
```go
type ProviderRegistry struct {
    providers      map[string]AuthProvider  // Cache: projectID -> provider
    mu             sync.RWMutex            // Thread-safe cache
    configRepo     ProjectAuthConfigRepo   // Database access
    defaultProvider AuthProvider           // Fallback
}

func (r *ProviderRegistry) GetProvider(projectID string) AuthProvider {
    // 1. Check cache (read lock)
    r.mu.RLock()
    if provider, ok := r.providers[projectID]; ok {
        r.mu.RUnlock()
        return provider  // Cache hit: ~0.001ms
    }
    r.mu.RUnlock()

    // 2. Load from database (write lock)
    r.mu.Lock()
    defer r.mu.Unlock()

    // Double-check (another goroutine may have loaded)
    if provider, ok := r.providers[projectID]; ok {
        return provider
    }

    // 3. Load config and create provider
    config := r.configRepo.GetByProjectID(projectID)
    provider := r.createProvider(config)
    r.providers[projectID] = provider
    return provider  // First hit: ~10-50ms
}
```

**Cache Invalidation**:
- On config update: `registry.InvalidateCache(projectID)`
- On config delete: `registry.InvalidateCache(projectID)` → falls back to default

**Trade-offs**:
- ✅ 99.9% of requests are cache hits (0.001ms overhead)
- ✅ Thread-safe concurrent access (RWMutex allows multiple readers)
- ❌ Stale cache if config updated externally (mitigation: explicit invalidation)
- ✅ Memory usage is negligible (few KB per provider, <100 projects = <10MB)

**Multi-Instance Consideration** (Future):
- Current: Each instance has its own cache (eventual consistency)
- Future: Redis pub/sub for cross-instance invalidation

### Decision 8: Middleware Placement

**Options Considered**:
1. **Global Middleware**: Apply to all routes
2. **Selective Middleware**: Only `/graphql/*` routes
3. **Resolver-Level**: Check auth in each GraphQL resolver

**Decision**: Selective middleware on `/graphql/*` routes only

**Rationale**:
- **Separation of Concerns**: Runtime API needs auth, design-time doesn't (yet)
- **Incremental Rollout**: Can secure runtime first, design-time later
- **Performance**: No auth overhead on health checks, metrics endpoints
- **Standard Pattern**: Auth middleware is industry-standard approach (vs scattered resolver checks)

**Middleware Stack**:
```go
router := gin.New()
router.Use(logging.Middleware())      // Log all requests
router.Use(gin.Recovery())           // Panic recovery
router.Use(cors.Middleware())        // CORS headers

// Public routes (no auth)
router.GET("/health", healthHandler)
router.Group("/api/design").Use(/* future: design-time auth */)

// Protected routes (JWT auth)
graphqlGroup := router.Group("/graphql")
graphqlGroup.Use(auth.JWTAuthMiddleware(providerRegistry))  // ← Auth here
graphqlGroup.POST("/:projectId/:cluster/:db/:model", graphqlHandler)
```

**Trade-offs**:
- ✅ Clean separation (auth only where needed)
- ✅ Easy to test (can mock auth middleware)
- ❌ Must remember to apply middleware to new runtime endpoints
- ✅ Performance (no auth for internal/admin endpoints)

### Decision 9: Error Handling Strategy

**Problem**: Authentication failures can happen for many reasons:
- Missing token
- Invalid signature
- Expired token
- Wrong signing method
- Provider misconfiguration

Users need clear, actionable error messages without exposing security details.

**Decision**: Structured error responses with context-appropriate detail

**Principles**:
1. **401 Unauthorized**: Authentication failures (client can fix)
2. **500 Internal Server Error**: Provider/config errors (admin must fix)
3. **No Stack Traces**: Never expose internal errors to clients
4. **Actionable Messages**: Tell users what to do (e.g., "token has expired, please refresh")

**Example Responses**:
```json
// Client error (401)
{"error": "invalid token signature"}
{"error": "token has expired"}
{"error": "missing authorization header"}

// Server error (500)
{"error": "authentication provider not configured for project"}
{"error": "failed to load authentication configuration"}
```

**Logging Strategy**:
```go
// Client errors (WARN level - expected, no action needed)
logger.Warn("JWT authentication failed",
    "project", projectID,
    "error", "invalid signature",
    "path", c.Request.URL.Path)

// Server errors (ERROR level - needs investigation)
logger.Error("Provider configuration error",
    "project", projectID,
    "provider", provider.Type(),
    "error", err.Error())
```

**Trade-offs**:
- ✅ Clear user-facing errors (better UX)
- ✅ Detailed server logs (easier debugging)
- ✅ No security leaks (no stack traces, internal paths)
- ❌ More error handling code (but worth it)

## Data Flow

### Authentication Flow (Successful)

```
1. Client                    2. Middleware              3. Registry           4. Provider           5. GraphQL Handler
   │                            │                         │                     │                     │
   ├─ POST /graphql/proj/... ──►│                         │                     │                     │
   │  Authorization: Bearer XX  │                         │                     │                     │
   │                            │                         │                     │                     │
   │                            ├─ Extract projectID ────►│                     │                     │
   │                            │  Extract token          │                     │                     │
   │                            │                         │                     │                     │
   │                            │                         ├─ GetProvider(proj)─►│                     │
   │                            │                         │  (from cache)       │                     │
   │                            │                         │◄────────────────────┤                     │
   │                            │                         │                     │                     │
   │                            │◄───── provider ─────────┤                     │                     │
   │                            │                         │                     │                     │
   │                            ├─────────────────────────┼─ GetPublicKey() ──►│                     │
   │                            │                         │  (cached)           │                     │
   │                            │◄────────────────────────┼─────────────────────┤                     │
   │                            │                         │                     │                     │
   │                            ├─ jwt.Parse(token, key) ─┤                     │                     │
   │                            │  ✓ Valid signature      │                     │                     │
   │                            │                         │                     │                     │
   │                            ├─────────────────────────┼─ ParseClaims() ────►│                     │
   │                            │◄────────────────────────┼── UnifiedClaims ────┤                     │
   │                            │                         │                     │                     │
   │                            ├─ c.Set("user", claims) ─┤                     │                     │
   │                            ├─ c.Next() ──────────────┼─────────────────────┼─────────────────────►│
   │                            │                         │                     │                     │
   │◄──────────── 200 OK ───────┼─────────────────────────┼─────────────────────┼─────────────────────┤
   │  {...graphql response...}  │                         │                     │                     │
```

### Configuration Update Flow

```
1. Admin Client        2. HTTP Handler       3. Service            4. Repository        5. Registry
   │                      │                     │                     │                     │
   ├─ POST /api/.../auth─►│                     │                     │                     │
   │  config              │                     │                     │                     │
   │                      │                     │                     │                     │
   │                      ├─ Parse & Validate ─►│                     │                     │
   │                      │                     ├─ Validate config ───►│                     │
   │                      │                     │  (PEM, URL, etc.)  │                     │
   │                      │                     │                     │                     │
   │                      │                     ├───── Save ──────────►│                     │
   │                      │                     │◄──── OK ─────────────┤                     │
   │                      │                     │                     │                     │
   │                      │                     ├─ Invalidate Cache ──────────────────────►│
   │                      │                     │                     │                     │
   │◄──────── 200 OK ─────┤                     │                     │                     │
   │                      │                     │                     │                     │
   │                      │                     │                     │                     │
   [Next GraphQL Request]                                                                   │
   │                      │                     │                     │                     │
   ├─ POST /graphql/... ──┼───── GetProvider ───┼─────────────────────┼─────────────────────►│
   │                      │                     │                     │                     ├─ Cache miss
   │                      │                     │                     │                     ├─ Load fresh config
   │                      │                     │                     │                     ├─ Create new provider
   │                      │                     │                     │                     │
   │                      │◄──────────────────────────────────────────┼─────────────────────┤
```

## Performance Analysis

### Latency Breakdown (Target)

| Operation | Cold (First Request) | Warm (Cached) | Budget |
|-----------|---------------------|---------------|--------|
| Extract project ID | 0.001ms | 0.001ms | - |
| Extract token | 0.01ms | 0.01ms | - |
| Get provider (registry) | **10-50ms** (DB query + init) | **0.001ms** (cache hit) | <1ms |
| Get public key | **1-5ms** (parse cert) | **0.001ms** (cached) | <1ms |
| JWT signature verify (RS256) | 0.1-0.5ms | 0.1-0.5ms | <1ms |
| Parse claims | 0.01ms | 0.01ms | - |
| **Total Auth Overhead** | **11-56ms** | **~0.5ms** | <1ms |

**Optimization Strategies**:
1. ✅ Provider caching → 99% requests are cache hits
2. ✅ Public key caching → No repeated cert parsing
3. ✅ RS256 is fast enough (0.5ms acceptable overhead)
4. 🔮 Future: Preload providers on startup (eliminate first-hit penalty)

### Memory Footprint

| Component | Per Project | Estimate (100 projects) |
|-----------|-------------|------------------------|
| Provider instance | 1-2 KB | 100-200 KB |
| Public key (RSA-2048) | 512 bytes | 50 KB |
| Config JSON | 1-5 KB | 100-500 KB |
| **Total** | **3-8 KB** | **300-800 KB** |

Negligible compared to typical Go application memory (50-200MB).

## Security Considerations

### Threat Model

**Threats**:
1. **Token Theft**: Attacker steals valid JWT token
2. **Token Replay**: Attacker reuses intercepted token
3. **Token Forgery**: Attacker generates fake token
4. **Provider Misconfiguration**: Wrong certificate → accept invalid tokens
5. **Key Compromise**: Casdoor private key leaked

**Mitigations**:

| Threat | Mitigation | Status |
|--------|-----------|--------|
| Token Theft | HTTPS required (no plain HTTP) | ✅ Config validation |
| Token Replay | Short token lifetime (2h), HTTPS | ✅ JWT exp validation |
| Token Forgery | RS256 signature verification | ✅ Implemented |
| Misconfiguration | Certificate validation on save | ✅ Config validation |
| Key Compromise | Casdoor controls private key, ModelCraft only has public key | ✅ By design |

### Security Best Practices

**Production Checklist**:
- [ ] HTTPS enforced (reject HTTP or redirect)
- [ ] JWT secret/certificate in environment variables (not config files)
- [ ] Token expiry = 2 hours (balance between UX and security)
- [ ] Refresh token support (clients can renew without re-login)
- [ ] Rate limiting on auth endpoints (prevent brute force)
- [ ] Audit logging (track who accessed what, when)

## Testing Strategy

### Unit Tests
- Mock AuthProvider interface
- Mock ProviderRegistry
- Test all error scenarios (missing header, invalid signature, expired token, etc.)
- Test concurrent cache access (race detector)
- Test claims validation

### Integration Tests
- Real CasdoorProvider with test certificate
- Generate real JWT tokens signed with test key
- Test full request-response cycle
- Test configuration CRUD + cache invalidation
- Test multi-tenant isolation

### Performance Tests
- Benchmark JWT verification (should be <1ms)
- Load test: 1000 req/s with JWT auth (should not degrade)
- Memory profiling (ensure no leaks in provider cache)

### Security Tests
- Test with expired token (should reject)
- Test with wrong signing method (should reject)
- Test with invalid signature (should reject)
- Test HTTPS enforcement (if enabled)

## Migration Plan

### Phase 1: Implementation (This Change)
- ✅ Build all components
- ✅ Unit and integration tests
- ✅ Documentation

### Phase 2: Gradual Rollout
- Deploy with `auth.runtime.optional_for_projects = ["default"]`
- Monitor metrics (auth success/failure rates)
- Migrate one project at a time
- Remove optional flag after migration complete

### Phase 3: Future Enhancements
- Design-time API authentication (reuse same infrastructure)
- Fine-grained permissions (RBAC based on roles in UnifiedClaims)
- API key authentication (for service-to-service)
- Token introspection (real-time revocation)

## Open Questions & Future Work

### Q1: Certificate Rotation
**Question**: How to handle Casdoor certificate rotation without downtime?

**Options**:
1. Manual: Admin updates config, app restarts
2. Automatic: Periodic fetch from JWKS endpoint
3. Dual-Certificate: Support multiple certificates during rotation

**Recommendation**: Start with manual (simple), add JWKS auto-refresh in Phase 2

### Q2: Multi-Instance Cache Invalidation
**Question**: How to invalidate cache across multiple ModelCraft instances?

**Current**: Each instance has local cache (eventual consistency - config changes take effect on next request per instance)

**Future Options**:
1. Redis pub/sub: Publish invalidation events
2. Database triggers: Notify via PostgreSQL LISTEN/NOTIFY
3. TTL-based: Cache expires after N minutes (no explicit invalidation)

**Recommendation**: Start with local cache (simpler), add Redis pub/sub when multi-instance deployment is confirmed

### Q3: Authorization (RBAC)
**Question**: Should authentication include authorization (roles/permissions)?

**Current**: Only authentication (identity verification), no permission checking

**Future**: Use `UnifiedClaims.Roles` and project-level role definitions to implement:
- Read-only access to GraphQL (queries only, no mutations)
- Admin access to config endpoints
- Audit access (read audit logs)

**Recommendation**: Phase 3 enhancement (auth first, authz later)

## Conclusion

This design balances:
- **Security**: Strong authentication with industry-standard JWT/RS256
- **Flexibility**: Multi-provider architecture supports diverse identity systems
- **Performance**: Sub-millisecond overhead via aggressive caching
- **Simplicity**: Casdoor default provides out-of-box experience
- **Extensibility**: Clean interfaces enable future enhancements (RBAC, API keys, etc.)

The architecture follows DDD principles with clear separation between domain (AuthProvider interface), application (ProviderRegistry), infrastructure (CasdoorProvider), and interfaces (middleware, HTTP handlers).
