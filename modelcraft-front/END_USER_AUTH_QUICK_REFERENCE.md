# End-User Authentication - Quick Reference

## Key Files & Locations

### Core Files
| File | Purpose |
|------|---------|
| `src/shared/stores/end-user-auth-store.ts` | Zustand store (token + userInfo storage) |
| `src/api-client/end-user/end-user-auth-client.ts` | Refresh logic + token helpers |
| `src/app/api/bff/org/[orgName]/end-user/auth/refresh/route.ts` | BFF refresh endpoint |
| `src/app/api/bff/org/[orgName]/end-user/auth/_proxy.ts` | Proxy to Go backend |
| `src/api-client/apollo/clients.ts` | Apollo client factories (needs fix) |
| `src/types/end-user-auth.ts` | TypeScript types + error codes |
| `src/middleware.ts` | Route protection + cookie checks |

---

## Storage Layer

### 1. In-Memory Store
```typescript
// File: src/shared/stores/end-user-auth-store.ts
useEndUserAuthStore.getState() → {
  accessToken: string | null           // JWT access token
  expiresAt: number | null             // Unix timestamp (ms)
  userInfo: EndUserInfo | null         // {id, username, orgName, projectSlug}
}
```

**Methods:**
- `setAccessToken(token, expiresIn)` - Store new token
- `setUserInfo(info)` - Update user metadata
- `clearSession()` - Clear all (on logout)
- `isTokenExpired()` - Check if refresh needed (5 min buffer)

### 2. HttpOnly Cookie
```
Cookie: mc_enduser_refresh_token
├─ Stored: Browser (HttpOnly, immune to XSS)
├─ Lifetime: Long-lived (backend-configured)
├─ Rotation: On each refresh
├─ Used by: BFF refresh endpoint (auto-included with credentials: include)
└─ Cleared: On logout (Set-Cookie with Max-Age=0)
```

### 3. Apollo Cache
- Separate cache per Apollo client instance
- Cleared on logout: `cache.clearStore()`

---

## Token Refresh Flow

### When to Refresh
```typescript
// Check expiry (5 min BEFORE actual expiry)
isTokenExpired() → Date.now() > (expiresAt - 5*60*1000)
```

### How to Refresh
```typescript
import { refreshEndUserAccessToken } from '@api-client/end-user/end-user-auth-client'

// Call manually or from Apollo link
const newToken = await refreshEndUserAccessToken({
  orgName: 'acme',
  projectSlug: 'crm'
})

// Automatically updates:
// ├─ useEndUserAuthStore.accessToken
// ├─ useEndUserAuthStore.expiresAt
// ├─ useEndUserAuthStore.userInfo (org/project from JWT)
// └─ Browser HttpOnly cookie (via Set-Cookie response)
```

### Concurrency Protection
```
Multiple refresh calls → All wait for first request
→ Share same Promise → Only 1 backend call ✓
```

---

## Apollo Client Pattern

### Current State (BROKEN)
```typescript
createEndUserScopedClient(orgName, projectSlug, endUserToken)
// ❌ Takes static token
// ❌ No refresh on expiry
// ❌ Token becomes invalid after 1 hour
```

### Tenant Pattern (TO REPLICATE)
```typescript
function createAuthLink() {
  return setContext(async (operation, { headers }) => {
    // Get current token from store
    let token = useAuthStore.getState().accessToken
    
    // Refresh if expired
    if (!token && !isEndUserPath()) {
      token = await refreshAccessToken()
    }
    
    // Use token in headers
    return { headers: { authorization: `Bearer ${token}` } }
  })
}
```

### End-User Pattern (FIX NEEDED)
```typescript
function createEndUserAuthLink(orgName: string, projectSlug: string) {
  return setContext(async (operation, { headers }) => {
    // Get current token from store
    let token = getEndUserToken()
    
    // Refresh if expired (5 min buffer)
    if (!token || isEndUserTokenExpired()) {
      const refreshParams = { orgName, projectSlug }
      token = await refreshEndUserAccessToken(refreshParams)
    }
    
    // Use token in headers
    if (!token) {
      // Clear session & redirect to login
      removeEndUserSession()
      window.location.href = `${getEndUserLoginPath(orgName)}?redirect=${window.location.pathname}`
      return { headers }
    }
    
    return { headers: { authorization: `Bearer ${token}` } }
  })
}
```

---

## Endpoints

### Frontend → BFF (via credentials: include)
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/bff/org/{orgName}/end-user/auth/login` | Login |
| POST | `/api/bff/org/{orgName}/end-user/auth/register` | Register |
| **POST** | **`/api/bff/org/{orgName}/end-user/auth/refresh`** | **← Key for refresh** |
| GET | `/api/bff/org/{orgName}/end-user/auth/me` | Get user info |
| POST | `/api/bff/org/{orgName}/end-user/auth/logout` | Logout |

### BFF → Go Backend
| Endpoint | Used by |
|----------|---------|
| `/api/end-user/auth/login` | BFF login route |
| `/api/end-user/auth/refresh` | BFF refresh route (extracts cookie → injects to body) |
| `/api/end-user/auth/me` | BFF me route |
| `/api/end-user/auth/logout` | BFF logout route |

### GraphQL Endpoints
| Endpoint | Usage |
|----------|-------|
| `/api/bff/graphql/end-user/org/{orgName}/project/{projectSlug}` | Org/project queries |
| `/api/bff/graphql/end-user/org/{orgName}/project/{projectSlug}/db/{db}/model/{model}` | Model runtime queries |

---

## Error Handling

### Refresh Failure
1. `refreshEndUserAccessToken()` catches errors
2. Clears session: `useEndUserAuthStore.getState().clearSession()`
3. Returns `null`
4. Apollo link should: Redirect to login

### Error Codes
| Code | Meaning | User Message |
|------|---------|--------------|
| `INVALID_REFRESH_TOKEN` | Refresh token expired/revoked | "登录已过期，请重新登录" |
| `INVALID_CREDENTIALS` | Wrong username/password | "用户名或密码错误，请重试" |
| `CLUSTER_NOT_CONFIGURED` | Project has no cluster | "服务暂时不可用，请联系管理员" |
| `PRIVATE_DB_NOT_INITIALIZED` | User DB not ready | "私有库尚未初始化" |

See `src/types/end-user-auth.ts` for complete error mapping.

---

## JWT Payload Structure

```typescript
interface EndUserJWTPayload {
  sub: string              // userId
  org_name: string         // Organization name
  project_slug: string     // Project slug
  role: 'end_user'         // Always 'end_user'
  exp?: number             // Expiry (unix timestamp)
  iat?: number             // Issued at
}
```

**Decoded at:** `src/api-client/end-user/end-user-auth-client.ts`
**Used for:** Auto-populating userInfo with org/project context

---

## Middleware Protection

### Public Routes
- ✅ `/end-user/{orgName}/login`
- ✅ `/end-user/{orgName}/no-project-access`

### Protected Routes
- 🔒 `/end-user/{orgName}/workspace` (requires `mc_enduser_refresh_token`)
- 🔒 `/end-user/{orgName}/projects/{projectSlug}/*` (requires `mc_enduser_refresh_token`)

**Redirect on Missing Cookie:**
```
/end-user/acme/workspace (no cookie)
  → /end-user/acme/login?redirect=/end-user/acme/workspace
```

---

## Type Definitions

### Key Types (in `src/types/end-user-auth.ts`)
```typescript
// Request
interface EndUserLoginRequest {
  orgName: string
  projectSlug: string
  username: string
  password: string
}

// Response
interface EndUserAuthResponse {
  accessToken?: string
  expiresAt?: string  // ISO 8601 (Go backend)
  expiresIn?: number  // Seconds (legacy)
  projects?: EndUserAccessibleProject[]
}

// User Info
interface EndUserInfo {
  id: string
  username: string
  orgName: string
  projectSlug: string
}
```

---

## Testing Refresh Scenario

### Manual Test
```typescript
// In browser console (on end-user page)

// 1. Check current token
useEndUserAuthStore.getState()
// → { accessToken: "JWT...", expiresAt: 1234567890, userInfo: {...} }

// 2. Manually trigger refresh
import { refreshEndUserAccessToken } from '@/api-client/end-user/end-user-auth-client'
await refreshEndUserAccessToken({ orgName: 'acme', projectSlug: 'crm' })
// → Returns new token

// 3. Verify store updated
useEndUserAuthStore.getState()
// → { accessToken: "JWT2...", expiresAt: 1234567891, userInfo: {...} }

// 4. Verify cookie rotated
document.cookie  // Can't see HttpOnly cookie but it's there
// Check DevTools → Application → Cookies → mc_enduser_refresh_token
```

---

## TODO: Fix Apollo Client

The current `createEndUserScopedClient()` needs to be refactored to:

1. ✅ Get token from store (not passed in)
2. ✅ Check expiry on each operation
3. ✅ Call `refreshEndUserAccessToken()` if needed
4. ✅ Use refreshed token in headers
5. ✅ Handle refresh failure (redirect to login)
6. ✅ Support concurrency (shared promise already exists)

See `src/api-client/apollo/clients.ts` for implementation details.

