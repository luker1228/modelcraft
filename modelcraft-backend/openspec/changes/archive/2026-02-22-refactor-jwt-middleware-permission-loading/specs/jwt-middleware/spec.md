# jwt-middleware Specification Delta

## MODIFIED Requirements

### Requirement: JWT Authentication Without Authorization

The JWT authentication middleware MUST validate JWT tokens and extract ONLY user identity information, deferring permission loading to the organization context middleware.

#### Scenario: ModelCraft JWT extracts identity only

**Given** a request with a valid ModelCraft JWT containing:
- `user_id`: "user-uuid-123"
- `email`: "user@example.com"
- `name`: "John Doe"
- `organization`: "modelcraft"
- `roles`: ["owner", "editor"] (embedded in JWT)
- `permissions`: ["model:read", "model:write"] (embedded in JWT)

**When** the JWT middleware validates the token

**Then** the middleware injects into context:
- `ContextKeyUserID` = "user-uuid-123"
- `ContextKeyEmail` = "user@example.com"
- `ContextKeyName` = "John Doe"
- `ContextKeyOrganization` = "modelcraft"

**And** the middleware does NOT inject:
- `ContextKeyRoles` (ignored, will be loaded by org context middleware)
- `ContextKeyPermissions` (ignored, will be loaded by org context middleware)

**And** logs: "JWT authentication successful: userID=user-uuid-123, email=user@example.com, org=modelcraft (authorization deferred to org context middleware)"

#### Scenario: Casdoor JWT extracts identity only

**Given** a request with a valid Casdoor JWT containing:
- `sub`: "casdoor-user-456"
- `email`: "user@example.com"
- `name`: "Jane Smith"
- `owner`: "acme-corp"

**When** the JWT middleware validates the token

**Then** the middleware injects into context:
- `ContextKeyUserID` = "casdoor-user-456"
- `ContextKeyEmail` = "user@example.com"
- `ContextKeyName` = "Jane Smith"
- `ContextKeyOrganization` = "acme-corp"

**And** the middleware does NOT inject roles or permissions

**And** logs: "JWT authentication successful: userID=casdoor-user-456, email=user@example.com, org=acme-corp (authorization deferred to org context middleware)"

#### Scenario: Chi middleware behaves identically

**Given** a request using Chi router with valid JWT

**When** the Chi JWT middleware (`ChiJWTAuthMiddleware`) validates the token

**Then** the middleware injects ONLY identity fields into `r.Context()`:
- `ContextKeyUserID`, `ContextKeyEmail`, `ContextKeyName`, `ContextKeyOrganization`

**And** does NOT inject `ContextKeyRoles` or `ContextKeyPermissions`

**And** log messages match Chi middleware format

---

## REMOVED Requirements

### Requirement: Permission Injection from JWT Claims

~~The JWT middleware SHOULD extract roles and permissions from JWT claims and inject them into context if available.~~

**Rationale**: Permission loading is now the responsibility of `OrgContextMiddleware`, which queries fresh permissions from the database with Redis caching. This ensures permissions are always up-to-date and organization-scoped.

---

## Dependencies

- **Requires**: `org-context-middleware` capability MUST be used after JWT middleware to load roles and permissions
- **Breaking Change**: Code that reads `ContextKeyRoles` or `ContextKeyPermissions` immediately after JWT middleware will fail (must use org context middleware first)
