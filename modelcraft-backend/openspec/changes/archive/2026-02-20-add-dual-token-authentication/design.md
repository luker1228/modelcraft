# Design: Dual-Token Authentication Architecture

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────────────┐
│                         Client Application                           │
│  - Stores ModelCraft JWT                                            │
│  - Uses JWT for all API requests                                    │
│  - Has user info + roles + permissions cached                       │
└────────────────────────┬─────────────────────────────────────────────┘
                         │
                         │ Bearer {ModelCraft-JWT}
                         ▼
┌──────────────────────────────────────────────────────────────────────┐
│                    ModelCraft API Gateway                            │
│  ┌────────────────────────────────────────────────────────────┐    │
│  │ JWT Middleware (validates ModelCraft JWT)                  │    │
│  │ - Verifies signature with ModelCraft secret                │    │
│  │ - Extracts claims (user, org, roles, permissions)          │    │
│  │ - Injects into request context                             │    │
│  └────────────────────────────────────────────────────────────┘    │
│                         │                                            │
│                         ▼                                            │
│  ┌────────────────────────────────────────────────────────────┐    │
│  │ Authorization Logic                                        │    │
│  │ - Reads roles/permissions from context (no DB query!)      │    │
│  │ - Checks against required permissions                      │    │
│  └────────────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────────────┘
```

## Token Exchange Flow Design

### Component Responsibilities

**AuthHandler** (`internal/handlers/auth_handler.go`)
- Orchestrates the dual-token exchange
- Communicates with Casdoor
- Calls TokenService to generate ModelCraft JWT

**TokenService** (NEW: `internal/app/auth/token_service.go`)
- Generates ModelCraft JWT with custom claims
- Queries user roles and permissions
- Validates token expiration and structure

**JWT Middleware** (`internal/middleware/jwt_auth.go`)
- Currently validates Casdoor JWT
- Will be updated to validate ModelCraft JWT
- Maintains backward compatibility during migration

### Detailed Flow

```go
// Stage 1: Exchange authorization code for ModelCraft JWT
POST /api/auth/token
{
  "code": "oauth-authorization-code"
}

// Internal flow:
1. AuthHandler.ExchangeToken()
   ├─> Exchange code with Casdoor for Casdoor JWT
   ├─> Parse Casdoor JWT to extract user identity
   ├─> Call TokenService.IssueToken()
   │   ├─> Verify user exists in DB
   │   ├─> Query UserRoleService.ListUserRoles()
   │   ├─> Query PermissionService.ListRolePermissions() for each role
   │   ├─> Build ModelCraft JWT claims
   │   └─> Sign JWT with ModelCraft secret
   └─> Return TokenResponse with JWT + user info + roles + permissions

// Response:
{
  "accessToken": "eyJhbGc...",  // ModelCraft JWT
  "tokenType": "Bearer",
  "expiresIn": 3600,
  "user": {
    "id": "uuid",
    "externalId": "casdoor-user-id",
    "name": "John Doe",
    "email": "john@example.com"
  },
  "organization": {
    "name": "modelcraft"
  },
  "roles": [
    {"id": 1, "name": "owner", "displayName": "Owner"}
  ],
  "permissions": [
    "model:read", "model:write", "model:delete",
    "cluster:read", "cluster:write", "cluster:delete"
  ]
}
```

## Data Models

### ModelCraft JWT Claims

```go
// internal/domain/auth/modelcraft_claims.go
type ModelCraftClaims struct {
    jwt.RegisteredClaims

    // User identity
    UserID       string `json:"user_id"`       // ModelCraft user UUID
    ExternalID   string `json:"external_id"`   // Casdoor user ID
    Name         string `json:"name"`
    Email        string `json:"email"`

    // Authorization
    Organization string   `json:"organization"`
    Roles        []string `json:"roles"`        // Role names: ["owner", "editor"]
    Permissions  []string `json:"permissions"`  // Permission strings: ["model:read", "model:write"]

    // Metadata
    Issuer       string `json:"iss"`   // Always "modelcraft"
}
```

### Enhanced Token Response

```go
// internal/models/auth_response.go (or update existing)
type EnhancedTokenResponse struct {
    AccessToken  string              `json:"accessToken"`
    TokenType    string              `json:"tokenType"`
    ExpiresIn    int                 `json:"expiresIn"`
    RefreshToken string              `json:"refreshToken,omitempty"` // Future use

    // User information
    User         UserInfo            `json:"user"`
    Organization OrganizationInfo    `json:"organization"`

    // Authorization information
    Roles        []RoleInfo          `json:"roles"`
    Permissions  []string            `json:"permissions"`
}

type UserInfo struct {
    ID         string `json:"id"`
    ExternalID string `json:"externalId"`
    Name       string `json:"name"`
    Email      string `json:"email"`
}

type OrganizationInfo struct {
    Name string `json:"name"`
}

type RoleInfo struct {
    ID          int    `json:"id"`
    Name        string `json:"name"`
    DisplayName string `json:"displayName"`
}
```

## Implementation Components

### 1. Token Service (NEW)

**File**: `internal/app/auth/token_service.go`

```go
type TokenService struct {
    userRepo         user.UserRepository
    userRoleService  *permission.UserRoleService
    permService      *permission.PermissionService
    jwtSecret        []byte
    tokenExpiration  time.Duration
}

func (s *TokenService) IssueToken(ctx context.Context, casdoorClaims *CasdoorClaims) (*EnhancedTokenResponse, error)
func (s *TokenService) ValidateToken(ctx context.Context, tokenString string) (*ModelCraftClaims, error)
func (s *TokenService) generateJWT(claims *ModelCraftClaims) (string, error)
```

**Responsibilities**:
- Generate ModelCraft JWT with user identity and authorization data
- Query roles and permissions from database
- Sign JWT with ModelCraft secret
- Validate and parse ModelCraft JWT

### 2. Updated Auth Handler

**File**: `internal/handlers/auth_handler.go`

```go
type AuthHandler struct {
    // Existing fields
    casdoorURL       string
    clientID         string
    clientSecret     string

    // New dependency
    tokenService     *auth.TokenService  // NEW
}

// Enhanced ExchangeToken method
func (h *AuthHandler) ExchangeToken(c *gin.Context) {
    // 1. Exchange code with Casdoor (existing logic)
    casdoorToken := exchangeWithCasdoor(req.Code)

    // 2. Parse Casdoor JWT to get user identity
    casdoorClaims := parseCasdoorJWT(casdoorToken)

    // 3. Issue ModelCraft token (NEW)
    tokenResponse, err := h.tokenService.IssueToken(ctx, casdoorClaims)

    // 4. Return enhanced response with user info + roles + permissions
    c.JSON(http.StatusOK, tokenResponse)
}
```

### 3. Updated JWT Middleware

**File**: `internal/middleware/jwt_auth.go`

```go
// Add backward compatibility: detect token issuer
func JWTAuthMiddleware(config *JWTAuthConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := extractToken(c)

        // Parse token to check issuer (without validation)
        unverifiedToken, _ := jwt.ParseUnverified(tokenString, &jwt.MapClaims{})
        claims := unverifiedToken.Claims.(jwt.MapClaims)
        issuer := claims["iss"].(string)

        if issuer == "modelcraft" {
            // NEW: Validate ModelCraft JWT
            validateModelCraftJWT(c, tokenString, config.ModelCraftSecret)
        } else {
            // OLD: Validate Casdoor JWT (backward compatibility)
            validateCasdoorJWT(c, tokenString, config.CasdoorPublicKey)
        }
    }
}
```

### 4. Updated Permission Middleware

**File**: `internal/middleware/permission.go`

```go
// Simplify permission check - read from context instead of DB query
func PermissionMiddleware(requiredPermission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get permissions from context (injected by JWT middleware)
        permissions, exists := c.Get(ContextKeyPermissions)

        if !exists {
            // Fallback: query database (for backward compatibility)
            permissions = queryFromDatabase(ctx, userID, orgName)
        }

        if !hasPermission(permissions, requiredPermission) {
            c.AbortWithStatusJSON(http.StatusForbidden, ...)
            return
        }

        c.Next()
    }
}
```

## Configuration Changes

**File**: `configs/config.yaml`

```yaml
jwt:
  secret: ""  # MUST override via JWT_SECRET (used for ModelCraft JWT signing)
  expiration: 3600  # Token lifetime in seconds (default: 1 hour)
  issuer: "modelcraft"  # JWT issuer claim

auth:
  design:
    # Support dual-token during migration
    accept_casdoor_jwt: true   # Accept Casdoor JWT (backward compatibility)
    accept_modelcraft_jwt: true  # Accept ModelCraft JWT (new flow)
```

## Database Queries

The token issuance process requires these queries:

```sql
-- 1. Verify user exists
SELECT id, external_id, name, email
FROM users
WHERE external_id = ?

-- 2. Get user's roles in organization
SELECT r.id, r.name, r.display_name, ur.org_name
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id
WHERE ur.user_id = ? AND ur.org_name = ?

-- 3. Get permissions for each role
SELECT p.obj, p.act
FROM role_permissions rp
JOIN permissions p ON rp.permission_id = p.id
WHERE rp.role_id = ?
```

**Performance**: These queries run ONCE per login, not on every request.

## Error Handling

```go
// Token Service errors
var (
    ErrUserNotFound       = bizerrors.New(bizerrors.NotFound, "user not found in ModelCraft database")
    ErrInvalidCasdoorJWT  = bizerrors.New(bizerrors.ParamInvalid, "invalid Casdoor JWT")
    ErrTokenGeneration    = bizerrors.New(bizerrors.SystemError, "failed to generate ModelCraft JWT")
    ErrTokenExpired       = bizerrors.New(bizerrors.ParamInvalid, "token has expired")
    ErrInvalidToken       = bizerrors.New(bizerrors.ParamInvalid, "invalid or malformed token")
)
```

## Migration Path

### Phase 1: Implementation (Week 1-2)
- Implement TokenService
- Update AuthHandler.ExchangeToken
- Deploy with feature flag disabled

### Phase 2: Rollout (Week 3)
- Enable dual-token support (both Casdoor JWT and ModelCraft JWT accepted)
- Update frontend to use new token response format
- Monitor for issues

### Phase 3: Migration (Week 4-6)
- Migrate all clients to use ModelCraft JWT
- Monitor usage of Casdoor JWT (should decline to zero)

### Phase 4: Cleanup (Week 7+)
- Deprecate Casdoor JWT support
- Remove backward compatibility code
- Simplify middleware

## Testing Strategy

### Unit Tests
- `token_service_test.go`: Test JWT generation, validation, expiration
- `auth_handler_test.go`: Test enhanced ExchangeToken flow
- `jwt_auth_test.go`: Test middleware with both token types

### Integration Tests

**Current State**: Integration tests use `auth_token` fixture which obtains Casdoor JWT directly:
```python
# tests/conftest.py (lines 130-165)
@pytest.fixture(scope="session")
def auth_token(test_config, test_user_with_owner_role):
    """Returns Casdoor JWT from password flow"""
    from common.auth import get_test_access_token
    token = get_test_access_token(test_config)  # Returns Casdoor JWT
    return token
```

**Migration Required**: Tests must be updated to use ModelCraft JWT:

```python
# Updated: tests/common/auth.py
def exchange_for_modelcraft_token(test_config, casdoor_token):
    """
    Exchange Casdoor JWT for ModelCraft JWT via /api/auth/token.

    Note: Since tests use password flow (not OAuth code), we may need
    a special test endpoint that accepts Casdoor JWT directly, or
    mock the OAuth code flow.

    Returns:
        str: ModelCraft JWT
    """
    # Option A: Mock OAuth code exchange
    # This requires test environment to support code generation from JWT

    # Option B: Add test-only endpoint POST /api/auth/token-exchange
    # that accepts Casdoor JWT directly and returns ModelCraft JWT
    # (only enabled when AUTH_DESIGN_ENABLED=false or in test mode)

    response = requests.post(
        f"{test_config.get_base_url()}/api/auth/token",
        json={"casdoor_token": casdoor_token},  # Test-only parameter
        timeout=10
    )
    response.raise_for_status()

    data = response.json()
    return data['accessToken']  # ModelCraft JWT

def get_modelcraft_token(test_config):
    """Get ModelCraft JWT for tests (convenience function)"""
    casdoor_token = get_test_access_token(test_config)
    return exchange_for_modelcraft_token(test_config, casdoor_token)

# Updated: tests/conftest.py
@pytest.fixture(scope="session")
def auth_token(test_config, test_user_with_owner_role):
    """Returns ModelCraft JWT for authenticated API calls"""
    from common.auth import get_modelcraft_token
    token = get_modelcraft_token(test_config)  # Now returns ModelCraft JWT
    print(f"✅ Obtained ModelCraft JWT token")
    return token
```

**New Tests Required**:
```python
# tests/design/auth/test_dual_token_exchange.py
def test_token_exchange_response_structure(base_url, test_config):
    """Verify enhanced token response includes roles and permissions"""
    casdoor_token = get_test_access_token(test_config)
    response = requests.post(f"{base_url}/api/auth/token",
                            json={"casdoor_token": casdoor_token})

    assert response.status_code == 200
    data = response.json()

    # Verify structure
    assert "accessToken" in data
    assert "user" in data
    assert "organization" in data
    assert "roles" in data
    assert "permissions" in data

    # Verify user info
    assert data["user"]["id"]
    assert data["user"]["externalId"]
    assert data["user"]["name"]
    assert data["user"]["email"]

    # Verify roles
    assert isinstance(data["roles"], list)
    assert len(data["roles"]) > 0
    assert "id" in data["roles"][0]
    assert "name" in data["roles"][0]

    # Verify permissions
    assert isinstance(data["permissions"], list)
    assert len(data["permissions"]) > 0
    assert all(isinstance(p, str) for p in data["permissions"])

    # Verify JWT claims
    import jwt
    claims = jwt.decode(data["accessToken"], options={"verify_signature": False})
    assert claims["iss"] == "modelcraft"
    assert claims["organization"] == data["organization"]["name"]
    assert set(claims["permissions"]) == set(data["permissions"])

def test_modelcraft_jwt_accepted_by_api(graphql_client, modelcraft_token):
    """Verify ModelCraft JWT works for API calls"""
    # Use ModelCraft JWT for GraphQL query
    result = graphql_client.execute(
        gql("query { projects { nodes { id name } } }"),
        headers={"Authorization": f"Bearer {modelcraft_token}"}
    )
    assert result["projects"]

def test_permission_check_uses_jwt_claims(graphql_client, modelcraft_token, db_config):
    """Verify permission checks don't query database"""
    import pymysql

    # Monitor database queries
    connection = pymysql.connect(**db_config)
    cursor = connection.cursor()

    # Get query count before request
    cursor.execute("SHOW STATUS LIKE 'Questions'")
    before_count = int(cursor.fetchone()[1])

    # Make API call that requires permission check
    graphql_client.execute(
        gql("mutation { createModel(input: {...}) { id } }"),
        headers={"Authorization": f"Bearer {modelcraft_token}"}
    )

    # Get query count after request
    cursor.execute("SHOW STATUS LIKE 'Questions'")
    after_count = int(cursor.fetchone()[1])

    # Verify no user_roles or role_permissions queries
    # (JWT claims should be used instead)
    # Note: Some queries are expected (e.g., INSERT for createModel)
    # but user_roles/role_permissions queries should NOT occur
    query_log = cursor.execute("SELECT * FROM mysql.general_log WHERE ...")
    # Assert no "SELECT * FROM user_roles" or "SELECT * FROM role_permissions"
```

**Test Migration Checklist**:
- [ ] Add `exchange_for_modelcraft_token()` to `tests/common/auth.py`
- [ ] Update `auth_token` fixture to return ModelCraft JWT
- [ ] Add new test file `tests/design/auth/test_dual_token_exchange.py`
- [ ] Run all integration tests to verify no regressions
- [ ] Add test to verify permission checks don't query database

### Load Testing
- Measure performance improvement (expect >90% reduction in permission check latency)
- Verify no performance degradation during token issuance

## Security Considerations

1. **Secret Key Management**
   - ModelCraft JWT secret MUST be different from Casdoor credentials
   - Store in environment variable (`JWT_SECRET`)
   - Rotate periodically (requires re-authentication of all users)

2. **Token Lifetime**
   - Short-lived tokens (1 hour default) limit impact of permission changes
   - Balance between performance (longer = fewer re-auths) and security (shorter = fresher permissions)

3. **Permission Staleness**
   - Max 1 hour delay for permission changes to take effect
   - For critical permission revocations, implement token revocation list (future work)

4. **Token Size**
   - JWT size increases with number of permissions
   - For users with >100 permissions, consider storing permission hash or using compact format

## Performance Impact

**Before** (Current System):
- Login: 1 request (Casdoor token exchange)
- Per request: 1 JWT validation + 2-3 DB queries (roles + permissions)
- Permission check latency: ~10-50ms (DB query overhead)

**After** (Dual-Token System):
- Login: 1 request (Casdoor + ModelCraft token exchange + DB queries)
- Per request: 1 JWT validation (no DB queries)
- Permission check latency: <1ms (in-memory claim check)

**Expected Improvement**: 90-95% reduction in permission check latency.

## Open Questions

1. **Refresh Token Support**: Should we implement refresh tokens now or defer?
   - Pro: Better UX (no re-authentication every hour)
   - Con: Adds complexity (token storage, rotation, revocation)
   - **Recommendation**: Defer to future work

2. **Token Revocation**: How to handle immediate permission revocation?
   - Option A: Accept 1-hour staleness (simple, acceptable for most cases)
   - Option B: Implement revocation list (complex, adds Redis dependency)
   - **Recommendation**: Start with Option A, implement B if needed

3. **Multiple Organizations**: Should JWT support multiple organizations?
   - Current: User authenticated to single organization at a time
   - Future: User switches between organizations without re-authentication
   - **Recommendation**: Keep single-org for now, design allows future extension
