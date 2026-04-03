# Proposal: Add Dual-Token Authentication Flow

## Overview

Implement a dual-token authentication flow to separate concerns between external identity provider (Casdoor) and internal authorization (ModelCraft). This approach improves security, reduces database queries, and provides better user experience by including role/permission information in the authentication flow.

## Problem Statement

Currently, the authentication system has several limitations:

1. **Single Token Flow**: The system directly uses Casdoor JWT tokens for authentication, mixing identity verification with authorization
2. **Repeated Database Queries**: Every permission check requires querying the database to fetch user roles and permissions (see `user_role_service.go:153-237`)
3. **No Role/Permission in Login Response**: Clients must make additional requests to fetch user's roles and permissions after login
4. **Tight Coupling to Casdoor**: The application logic is directly coupled to Casdoor's JWT structure

## Proposed Solution

Implement a **dual-token authentication flow**:

### Token Flow

```
┌─────────┐                 ┌─────────┐                 ┌───────────┐
│ Client  │                 │ Casdoor │                 │ ModelCraft│
└────┬────┘                 └────┬────┘                 └─────┬─────┘
     │                           │                            │
     │ 1. Redirect to login      │                            │
     │──────────────────────────>│                            │
     │                           │                            │
     │ 2. Auth code              │                            │
     │<──────────────────────────│                            │
     │                           │                            │
     │ 3. POST /api/auth/token   │                            │
     │   (code)                  │                            │
     │────────────────────────────────────────────────────────>│
     │                           │                            │
     │                           │ 4. Exchange code for token │
     │                           │<───────────────────────────│
     │                           │                            │
     │                           │ 5. Casdoor JWT             │
     │                           │────────────────────────────>│
     │                           │                            │
     │                           │                     6. Validate user
     │                           │                     7. Query roles/permissions
     │                           │                     8. Generate ModelCraft JWT
     │                           │                            │
     │ 9. ModelCraft JWT + User Info + Roles/Permissions      │
     │<────────────────────────────────────────────────────────│
     │                           │                            │
     │ 10. Use ModelCraft JWT for subsequent requests         │
     │────────────────────────────────────────────────────────>│
```

### Two-Stage Token Exchange

**Stage 1: Casdoor Token Exchange** (External Identity)
- Endpoint: `POST /api/auth/token` (existing, will be enhanced)
- Input: Authorization code from OAuth flow
- Process:
  1. Exchange code with Casdoor for Casdoor JWT
  2. Validate Casdoor JWT and extract user identity (`sub`, `name`, `email`, `owner`)
  3. Verify user exists in ModelCraft database
  4. Query user's roles and permissions from ModelCraft database
  5. Generate ModelCraft JWT containing user identity + roles + permissions
- Output:
  ```json
  {
    "accessToken": "ModelCraft-signed-JWT",
    "tokenType": "Bearer",
    "expiresIn": 3600,
    "user": {
      "id": "uuid",
      "externalId": "casdoor-user-id",
      "name": "User Name",
      "email": "user@example.com"
    },
    "organization": {
      "name": "org-name"
    },
    "roles": [
      {"id": 1, "name": "owner", "displayName": "Owner"}
    ],
    "permissions": [
      "model:read", "model:write", "cluster:manage"
    ]
  }
  ```

**Stage 2: Subsequent Requests** (Authorization)
- Clients use ModelCraft JWT for all subsequent API requests
- JWT contains claims:
  ```json
  {
    "sub": "user-uuid",
    "name": "User Name",
    "email": "user@example.com",
    "organization": "org-name",
    "roles": ["owner"],
    "permissions": ["model:read", "model:write", "cluster:manage"],
    "exp": 1234567890,
    "iat": 1234567890,
    "iss": "modelcraft"
  }
  ```
- Middleware validates ModelCraft JWT (not Casdoor JWT)
- Permission checks use claims directly (no database query needed)

## Key Benefits

1. **Single Database Query**: Roles/permissions fetched once during login, cached in JWT
2. **Better UX**: Frontend receives all user info (identity + authorization) in one response
3. **Separation of Concerns**:
   - Casdoor handles authentication (who you are)
   - ModelCraft handles authorization (what you can do)
4. **Performance**: No repeated database queries for permission checks
5. **Flexibility**: Can easily switch identity providers without changing authorization logic
6. **Security**: ModelCraft controls token lifetime and content independently

## Implementation Scope

This change affects the following capabilities:

1. **jwt-middleware**: Update to validate ModelCraft-signed JWT instead of Casdoor JWT
2. **casdoor-provider**: Enhance to support dual-token flow
3. **permission-management**: Update permission check logic to use JWT claims
4. **model-integration-testing**: Update test fixtures to use dual-token flow

New capability:
1. **dual-token-auth**: New authentication flow orchestration

## Migration Strategy

The implementation will maintain backward compatibility during rollout:

1. Support both Casdoor JWT and ModelCraft JWT initially (based on `iss` claim)
2. Gradually migrate clients to use new dual-token flow
3. Eventually deprecate direct Casdoor JWT usage

### Integration Test Migration

**Critical**: Integration tests currently use Casdoor JWT directly (via password flow in `tests/conftest.py:auth_token` fixture). Tests must be updated to:

1. Exchange Casdoor token for ModelCraft token via `/api/auth/token`
2. Use ModelCraft JWT for all API calls
3. Validate enhanced response includes roles and permissions
4. Verify permission checks use JWT claims (no database queries)

Test migration is part of Phase 4 (Testing) and must be completed before full deployment.

## Security Considerations

1. **Token Signing**: ModelCraft JWT signed with separate secret key (from `jwt.secret` config)
2. **Token Lifetime**: Configurable expiration (default: 1 hour, aligned with Casdoor)
3. **Refresh Flow**: Clients must re-authenticate with Casdoor when ModelCraft JWT expires
4. **Permission Staleness**: Max 1 hour delay for permission changes to take effect (acceptable trade-off for performance)

## Non-Goals

- This proposal does NOT include refresh token implementation (separate future enhancement)
- This proposal does NOT change the Casdoor integration itself
- This proposal does NOT modify existing GraphQL permission directives behavior

## Questions for Approval

1. Is 1 hour token lifetime acceptable? (matches Casdoor default)
2. Should we include refresh token support in this change, or defer to future work?
3. Should we support token revocation (adds complexity, requires Redis/database)?

## Related Work

- Current implementation: `internal/handlers/auth_handler.go:89-185` (ExchangeToken)
- Current JWT middleware: `internal/middleware/jwt_auth.go`
- Permission checking: `internal/app/permission/user_role_service.go:153-237`
