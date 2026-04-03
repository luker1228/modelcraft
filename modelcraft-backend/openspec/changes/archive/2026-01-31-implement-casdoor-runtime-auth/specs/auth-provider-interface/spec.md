# Spec: Authentication Provider Interface

## Overview

This specification defines the core abstraction for authentication providers in ModelCraft. The `AuthProvider` interface enables support for multiple SSO/identity providers (Casdoor, Keycloak, generic OIDC) through a common contract.

## Context

ModelCraft's runtime GraphQL API requires authentication to secure customer database access. Different organizations may use different identity providers (enterprise LDAP, cloud SSO, etc.). The AuthProvider interface decouples the authentication middleware from specific provider implementations, allowing dynamic provider selection based on project configuration.

## ADDED Requirements

### Requirement: AuthProvider Interface Definition

The system MUST define an `AuthProvider` interface that abstracts JWT verification logic.

#### Scenario: Interface declares public key retrieval method

**Given** an authentication provider implementation
**When** the system needs to verify a JWT signature
**Then** the provider must expose a `GetPublicKey(ctx context.Context) (interface{}, error)` method
**And** the method returns a key compatible with `jwt.Parse` (e.g., *rsa.PublicKey, *ecdsa.PublicKey)
**And** the method supports caching (repeated calls may return cached key)

#### Scenario: Interface declares signing method

**Given** an authentication provider implementation
**When** the system needs to validate the JWT signature algorithm
**Then** the provider must expose a `GetSigningMethod() string` method
**And** the method returns a valid algorithm identifier (e.g., "RS256", "HS256", "ES256")

#### Scenario: Interface declares claims parsing method

**Given** an authentication provider implementation
**When** the system receives a validated JWT with provider-specific claims structure
**Then** the provider must expose a `ParseClaims(claims jwt.MapClaims) (*UnifiedClaims, error)` method
**And** the method transforms provider-specific claim fields to the unified structure
**And** the method handles missing optional fields gracefully

#### Scenario: Interface declares provider type method

**Given** an authentication provider implementation
**When** the system needs to identify the provider for logging or debugging
**Then** the provider must expose a `Type() string` method
**And** the method returns a unique identifier (e.g., "casdoor", "keycloak", "oidc")

---

### Requirement: UnifiedClaims Structure

The system MUST define a `UnifiedClaims` struct that normalizes user identity across providers.

#### Scenario: UnifiedClaims contains required user identity fields

**Given** a JWT token validated by any provider
**When** the provider parses claims into UnifiedClaims
**Then** the struct must contain:
- `UserID string` - unique user identifier (from `sub` claim)
- `Username string` - human-readable username
- `Email string` - user email address (may be empty if not provided)
- `Organization string` - tenant/organization identifier for multi-tenancy
- `ExpiresAt time.Time` - token expiration timestamp (from `exp` claim)
- `Raw map[string]interface{}` - original claims for provider-specific extensions

#### Scenario: UnifiedClaims validates required fields

**Given** a UnifiedClaims instance
**When** the system calls `Validate() error` method
**Then** the method must return error if `UserID` is empty
**And** the method must return error if `Organization` is empty
**And** the method must return error if `ExpiresAt` is zero/past

#### Scenario: UnifiedClaims checks expiration

**Given** a UnifiedClaims instance with an expiration time
**When** the system calls `IsExpired() bool` method
**Then** the method returns `true` if current time >= ExpiresAt
**And** the method returns `false` otherwise

---

### Requirement: Provider Registry

The system MUST provide a `ProviderRegistry` to manage and cache provider instances per project.

#### Scenario: Registry returns provider for project

**Given** a ProviderRegistry initialized with auth config repository
**When** the system calls `GetProvider(projectID string) AuthProvider`
**Then** the registry queries the repository for project auth config
**And** if config exists, the registry creates/caches the appropriate provider
**And** if config does not exist, the registry returns the default provider
**And** subsequent calls with same projectID return cached provider (no repository query)

#### Scenario: Registry creates provider based on config

**Given** a project auth config with provider type "casdoor"
**When** the registry calls internal `createProvider(config)` method
**Then** the registry instantiates a CasdoorProvider with config values
**And** stores the provider in cache with key = projectID

#### Scenario: Registry handles missing default provider

**Given** a ProviderRegistry without a default provider configured
**When** the system calls `GetProvider(projectID)` for a project without auth config
**Then** the registry returns `nil`
**And** the caller must handle the nil case (e.g., reject authentication)

#### Scenario: Registry cache is thread-safe

**Given** multiple concurrent GraphQL requests for the same project
**When** multiple goroutines call `GetProvider(projectID)` simultaneously
**Then** the registry uses read-write locks (sync.RWMutex) to prevent race conditions
**And** the provider is initialized only once
**And** all goroutines receive the same cached provider instance

#### Scenario: Registry cache invalidation

**Given** a cached provider for a project
**When** the project's auth config is updated or deleted
**Then** the system must call `registry.InvalidateCache(projectID)` method
**And** the method removes the cached provider
**And** next `GetProvider(projectID)` call fetches fresh config from repository

---

### Requirement: Claims Validation

The system MUST validate UnifiedClaims before trusting the user identity.

#### Scenario: Validation rejects empty UserID

**Given** a UnifiedClaims with empty UserID
**When** the system calls `Validate()` method
**Then** the method returns error "UserID is required"

#### Scenario: Validation rejects empty Organization

**Given** a UnifiedClaims with empty Organization
**When** the system calls `Validate()` method
**Then** the method returns error "Organization is required"

#### Scenario: Validation rejects expired token

**Given** a UnifiedClaims with ExpiresAt = 2026-01-01 12:00:00
**And** current time is 2026-01-01 13:00:00
**When** the system calls `Validate()` method
**Then** the method returns error "token has expired"

#### Scenario: Validation accepts valid claims

**Given** a UnifiedClaims with:
- UserID = "user123"
- Organization = "tenant1"
- ExpiresAt = 1 hour from now
**When** the system calls `Validate()` method
**Then** the method returns `nil` (no error)

---

### Requirement: Provider Factory Pattern

The system MUST use a factory pattern to instantiate providers from configuration.

#### Scenario: Factory creates CasdoorProvider

**Given** a project auth config with:
```json
{
  "provider": "casdoor",
  "config": {
    "endpoint": "https://casdoor.example.com",
    "client_id": "abc123",
    "organization": "tenant1",
    "certificate": "-----BEGIN CERTIFICATE-----..."
  }
}
```
**When** the factory calls `createProvider(config)`
**Then** the factory instantiates `NewCasdoorProvider(config.Config)`
**And** returns an AuthProvider interface

#### Scenario: Factory creates KeycloakProvider (future)

**Given** a project auth config with provider = "keycloak"
**When** the factory calls `createProvider(config)`
**Then** the factory instantiates `NewKeycloakProvider(config.Config)`
**And** returns an AuthProvider interface

#### Scenario: Factory returns nil for unknown provider

**Given** a project auth config with provider = "unknown"
**When** the factory calls `createProvider(config)`
**Then** the factory returns `nil`
**And** the caller must handle the nil case (e.g., use default provider)

---

## Design Notes

**Thread Safety**: ProviderRegistry uses sync.RWMutex for cache access. Write lock only during cache insert/delete, read lock for lookups.

**Caching Strategy**: Providers are cached indefinitely until explicitly invalidated. This avoids repeated config queries and certificate parsing on every request.

**Extensibility**: New providers can be added by:
1. Implementing AuthProvider interface
2. Adding provider type to factory's switch statement
3. No changes required to middleware or registry

**Public Key Handling**: Interface returns `interface{}` to support different key types (RSA, ECDSA, HMAC secret). Caller (JWT middleware) uses `jwt.Parse` which accepts `interface{}` for key.

---

## Dependencies

**External Libraries**:
- `github.com/golang-jwt/jwt/v5` - JWT parsing and validation

**Internal Dependencies**:
- Project domain model (for multi-tenancy)
- Auth config repository (for loading project configurations)

---

## Testing Strategy

**Unit Tests**:
- Mock AuthProvider implementation
- ProviderRegistry with mock repository
- UnifiedClaims validation edge cases
- Concurrent cache access (race detector)

**Integration Tests**:
- Real provider instance (Casdoor) with test certificate
- Registry loading config from database
- Cache invalidation after config update

---

## Open Questions

1. **Key Rotation**: How should the system handle provider key rotation (e.g., JWKS endpoint updates)?
   - Current design: Providers can re-fetch on each GetPublicKey call or implement their own caching with TTL
   - Future: Add RefreshPublicKey() method to interface?

2. **Provider Health Checks**: Should the registry ping providers on initialization to validate config?
   - Concern: Startup delays if provider is unreachable
   - Alternative: Lazy validation on first request, log errors

3. **Multi-Provider per Project**: Should a single project support multiple providers simultaneously?
   - Current design: One provider per project
   - Use case: Support both LDAP and social login for same project
   - Complexity: Middleware would need to try multiple providers
