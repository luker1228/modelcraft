# End-User Authentication Architecture Analysis

## Overview
The ModelCraft frontend has a complete, symmetrical end-user authentication system mirroring the developer auth pattern. Here's how it works:

---

## 1. END-USER TOKEN STORAGE

### Store: `useEndUserAuthStore` (src/shared/stores/end-user-auth-store.ts)

**In-Memory Storage (Zustand):**
```typescript
interface EndUserAuthState {
  accessToken: string | null          // JWT access token
  expiresAt: number | null            // Unix timestamp (milliseconds)
  userInfo: EndUserInfo | null        // User metadata + context
  
  setAccessToken(token, expiresIn)    // Store token with expiry
  setUserInfo(info)                   // Store user details
  clearSession()                      // Clear everything on logout
  isTokenExpired()                    // Check if token needs refresh
}
```

**Key Features:**
- ✅ Refresh check: 5 minutes before actual expiry (`expiresAt - 5 * 60 * 1000`)
- ✅ Automatic user info caching
- ✅ Clean logout that clears token + user data

### Refresh Token Storage (Server-Side)

**HttpOnly Cookie:** `mc_enduser_refresh_token`
- ✅ Immune to XSS (HttpOnly flag)
- ✅ Automatically included in requests (credentials: 'include')
- ✅ Set by Go backend in login response
- ✅ Rotated on each refresh operation
- ✅ Cleared on logout

---

## 2. EXISTING REFRESH MECHANISM

### BFF Route: POST `/api/bff/org/{orgName}/end-user/auth/refresh`
**File:** `src/app/api/bff/org/[orgName]/end-user/auth/refresh/route.ts`

**How it works:**
1. Browser sends: POST `/api/bff/org/{orgName}/end-user/auth/refresh`
2. BFF extracts `mc_enduser_refresh_token` from HttpOnly cookie
3. BFF injects refreshToken into request body
4. BFF forwards to: `{BACKEND_URL}/api/end-user/auth/refresh`
5. Go backend validates refresh token and returns new access token
6. BFF transparently passes back Set-Cookie header (token rotation)
7. Frontend store gets updated with new token

**Request Flow:**
```
Browser (with HttpOnly cookie)
    ↓
BFF Route Handler (/api/bff/org/[orgName]/end-user/auth/refresh)
    ├─ Read: mc_enduser_refresh_token from cookie
    ├─ Inject: refreshToken into body
    ├─ Merge: orgName into body
    ↓
Go Backend (/api/end-user/auth/refresh)
    ├─ Validate refresh token
    ├─ Issue new access token (JWT)
    ├─ Rotate refresh token
    ├─ Return: EndUserAuthResponse (accessToken, refreshToken, expiresAt)
    ↓
BFF Response Handler
    ├─ Transparently pass Set-Cookie header back to browser
    ├─ Return: EndUserAuthResponse to frontend
    ↓
Frontend Store Update
    ├─ setAccessToken(newToken, expiresIn)
    ├─ Decode JWT → extract org_name, project_slug, sub
    └─ Auto-populate userInfo for display
```

### Proxy Helper: `proxyEndUserAuth()` (src/app/api/bff/org/[orgName]/end-user/auth/_proxy.ts)

**Cookie Handling:**
- ✅ Transparently forwards HttpOnly cookies to backend
- ✅ Transparently returns Set-Cookie headers to browser
- ✅ On logout: appends Set-Cookie with Max-Age=0 to clear cookie
- ✅ Supports multiple Set-Cookie headers (getSetCookie() fallback)

---

## 3. CLIENT-SIDE REFRESH IMPLEMENTATION

### Function: `refreshEndUserAccessToken()` (src/api-client/end-user/end-user-auth-client.ts)

**Flow:**
```typescript
export async function refreshEndUserAccessToken(
  params?: EndUserRefreshParams
): Promise<string | null> {
  // Prevent concurrent refresh requests (share same promise)
  if (_isRefreshing && _refreshPromise) {
    return _refreshPromise
  }

  _isRefreshing = true
  _refreshPromise = (async () => {
    try {
      const orgName = params?.orgName ?? ''
      
      // POST to BFF refresh endpoint
      const res = await fetch(`/api/bff/org/${orgName}/end-user/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ orgName }),
        credentials: 'include',  // ← Browser sends HttpOnly cookie
      })

      if (!res.ok) {
        useEndUserAuthStore.getState().clearSession()
        return null
      }

      const data = (await res.json()) as EndUserAuthResponse

      if (data.accessToken) {
        // Priority: expiresAt (ISO 8601) > expiresIn > default 1h
        const expiresIn = data.expiresAt
          ? Math.max(1, Math.floor((new Date(data.expiresAt).getTime() - Date.now()) / 1000))
          : (data.expiresIn ?? 3600)

        const store = useEndUserAuthStore.getState()
        store.setAccessToken(data.accessToken, expiresIn)

        // Parse JWT to extract org/project context
        const decoded = decodeEndUserJWT(data.accessToken)
        if (decoded) {
          store.setUserInfo({
            id: decoded.sub,
            username: '', // Filled later by fetchAndCacheEndUserInfo
            orgName: decoded.org_name,
            projectSlug: decoded.project_slug,
          })
        }

        return data.accessToken
      }

      return null
    } catch {
      return null
    } finally {
      _isRefreshing = false
      _refreshPromise = null
    }
  })()

  return _refreshPromise
}
```

**Key Features:**
- ✅ Concurrency protection (shared promise for simultaneous refresh calls)
- ✅ Parses JWT response to auto-populate org/project in store
- ✅ Handles both Go backend format (expiresAt ISO 8601) and legacy format (expiresIn seconds)
- ✅ Clears session on failure (triggers re-login)
- ✅ Returns new token or null

### Helper Functions:
```typescript
// Get token from store
export function getEndUserToken(): string | null {
  return useEndUserAuthStore.getState().accessToken
}

// Check if token is valid
export function isEndUserAuthenticated(): boolean {
  const { accessToken, isTokenExpired } = useEndUserAuthStore.getState()
  return !!accessToken && !isTokenExpired()
}

// Fetch user info from /me endpoint and cache it
export async function fetchAndCacheEndUserInfo(orgName: string): Promise<EndUserInfo | null> {
  const token = getEndUserToken()
  if (!token) return null

  const res = await fetch(`/api/bff/org/${orgName}/end-user/auth/me`, {
    credentials: 'include',
    headers: { Authorization: `Bearer ${token}` },
  })

  if (!res.ok) return null

  const data = (await res.json()) as EndUserMeResponse
  const store = useEndUserAuthStore.getState()
  const currentInfo = store.userInfo

  const info: EndUserInfo = {
    id: data.id,
    username: data.username,
    orgName: currentInfo?.orgName ?? '',
    projectSlug: currentInfo?.projectSlug ?? '',
  }

  store.setUserInfo(info)
  return info
}
```

---

## 4. APOLLO CLIENT INTEGRATION

### File: `src/api-client/apollo/clients.ts`

#### Tenant Pattern (Developer)
```typescript
function createAuthLink() {
  return setContext(async (operation, { headers }: { headers?: Record<string, string> }) => {
    try {
      let token = typeof window !== 'undefined' 
        ? useAuthStore.getState().accessToken 
        : null
      
      // Only refresh if not on end-user path
      if (!token && typeof window !== 'undefined' && !isEndUserPath()) {
        token = await refreshAccessToken()
      }

      const nextHeaders: Record<string, string> = {
        ...(headers ?? {}),
        'x-client-request-id': generateUUID(),
      }

      if (token) {
        nextHeaders.authorization = `Bearer ${token}`
      }

      const xAction = buildXAction(operation)
      if (xAction) {
        nextHeaders['X-Action'] = xAction
      }

      return { headers: nextHeaders }
    } catch (err) {
      console.error('[AuthLink] ERROR:', err)
      return { headers }
    }
  })
}
```

**Pattern: Shared promise on refresh (prevents concurrent requests)**

#### End-User Pattern (NEW - Can be replicated)
```typescript
export function createEndUserScopedClient(
  orgName: string,
  projectSlug: string,
  endUserToken: string  // ← Static token passed in
): ApolloClient<object> {
  const uri = `${GATEWAY_URL}/api/bff/graphql/end-user/org/${orgName}/project/${projectSlug}`
  const httpLink = createHttpLink({ uri, credentials: 'include' })

  // Simple auth link: just inject token + X-Action header
  const authLink = setContext((operation, { headers }: { headers?: Record<string, string> }) => {
    const xAction = buildXAction(operation)
    return {
      headers: {
        ...(headers ?? {}),
        authorization: `Bearer ${endUserToken}`,
        ...(xAction ? { 'X-Action': xAction } : {}),
      },
    }
  })

  return new ApolloClient({
    link: authLink.concat(httpLink),
    cache: new InMemoryCache(),
    defaultOptions: {
      watchQuery: { errorPolicy: 'all' },
      query: { errorPolicy: 'all' },
      mutate: { errorPolicy: 'all' },
    },
  })
}
```

**Current Approach:** Token passed statically at client creation time. **Problem:** No refresh on expiry!

---

## 5. MIDDLEWARE PROTECTION

### File: `src/middleware.ts`

**End-User Routes:**
- ✅ Public paths: `/end-user/{orgName}/login`, `/end-user/{orgName}/no-project-access`
- ✅ Protected paths: `/end-user/{orgName}/workspace`, `/end-user/{orgName}/projects/{projectSlug}/*`

**Check:** HttpOnly cookie `mc_enduser_refresh_token` presence only
- No token validation (happens client-side)
- If missing: redirect to login with `?redirect=` parameter

---

## 6. TYPES & INTERFACES

### Exported from `src/types/end-user-auth.ts`:

**Request Types:**
- `EndUserLoginRequest` - {orgName, projectSlug, username, password}
- `EndUserRegisterRequest` - {orgName, projectSlug, username, password}

**Response Types:**
- `EndUserAuthResponse` - {accessToken, refreshToken, expiresAt (ISO 8601), projects}
- `EndUserMeResponse` - {id, username, createdAt}

**JWT Payload:**
```typescript
interface EndUserJWTPayload {
  sub: string              // userId
  org_name: string         // Organization
  project_slug: string     // Project
  role: 'end_user'         // Always 'end_user'
  exp?: number             // Expiry timestamp
  iat?: number             // Issued at
}
```

**User Info:**
```typescript
interface EndUserInfo {
  id: string               // userId (JWT sub)
  username: string         // From /me endpoint
  orgName: string          // From JWT org_name
  projectSlug: string      // From JWT project_slug
}
```

---

## SUMMARY: WHAT YOU NEED TO DO

### Current Problem:
- `createEndUserScopedClient()` takes a static token but doesn't refresh on expiry
- Apollo link has no refresh logic

### Solution Pattern (replicate tenant pattern):
1. ✅ Store has refresh mechanism: `refreshEndUserAccessToken(params)`
2. ✅ BFF route exists: `/api/bff/org/{orgName}/end-user/auth/refresh`
3. ✅ HttpOnly cookie rotation handled
4. ✅ Concurrency protection implemented

### What's Missing in Apollo Client:
The end-user Apollo link needs:
1. Get current token from store
2. Check expiry (5 min buffer)
3. If expired: call `refreshEndUserAccessToken(params)` → update store
4. Use refreshed token in Authorization header
5. On refresh failure: clear session → redirect to login

This mirrors `createAuthLink()` for tenants, but with end-user auth mechanisms.

---

## ENDPOINT SUMMARY

### Frontend → BFF Routes:
- `POST /api/bff/org/{orgName}/end-user/auth/login` - Login
- `POST /api/bff/org/{orgName}/end-user/auth/register` - Register
- `POST /api/bff/org/{orgName}/end-user/auth/refresh` - Refresh token ✅
- `GET /api/bff/org/{orgName}/end-user/auth/me` - Get user info
- `POST /api/bff/org/{orgName}/end-user/auth/logout` - Logout

### BFF → Go Backend:
- `/api/end-user/auth/login`
- `/api/end-user/auth/register`
- `/api/end-user/auth/refresh` (Go backend returns: accessToken, refreshToken, expiresAt)
- `/api/end-user/auth/me`
- `/api/end-user/auth/logout`

### GraphQL Endpoints:
- Design-time: `/api/bff/graphql/org/{orgName}/project/{projectSlug}/`
- End-user: `/api/bff/graphql/end-user/org/{orgName}/project/{projectSlug}`
- End-user runtime: `/api/bff/graphql/end-user/org/{orgName}/project/{projectSlug}/db/{db}/model/{model}`

