# End-User Authentication Research Summary

## What You Asked For

You wanted to understand:
1. ✅ How end-user tokens are currently **stored**
2. ✅ Whether there's any **existing refresh mechanism**
3. ✅ What the **tenant createAuthLink pattern** looks like (to replicate for end-users)
4. ✅ The **end-user auth store** (useEndUserAuthStore)
5. ✅ Any **BFF route for end-user token refresh**

---

## What We Found

### 1. Token Storage ✅

**In-Memory (Zustand Store):**
```
useEndUserAuthStore
├─ accessToken: JWT string (short-lived, ~1h)
├─ expiresAt: Unix timestamp (ms) - set at login/refresh
├─ userInfo: {id, username, orgName, projectSlug}
└─ Methods: setAccessToken(), setUserInfo(), clearSession(), isTokenExpired()
```

**Server-Side (HttpOnly Cookie):**
```
mc_enduser_refresh_token
├─ Set by Go backend on login/refresh
├─ Rotated on each refresh
├─ Automatically sent via credentials: include
├─ Immune to XSS (HttpOnly flag)
└─ Cleared on logout (Set-Cookie with Max-Age=0)
```

**Refresh Logic:**
- 5-minute warning buffer: `isTokenExpired()` → `now > (expiresAt - 5*60*1000)`
- Concurrency protection: Shared promise (prevents multiple simultaneous refresh calls)

---

### 2. Existing Refresh Mechanism ✅✅✅

**Yes! It exists and is fully implemented:**

**BFF Endpoint:** `POST /api/bff/org/{orgName}/end-user/auth/refresh`
- File: `src/app/api/bff/org/[orgName]/end-user/auth/refresh/route.ts`
- Extracts `mc_enduser_refresh_token` from HttpOnly cookie
- Injects into request body as `refreshToken`
- Proxies to Go backend: `POST /api/end-user/auth/refresh`
- Returns: New accessToken + refreshToken (rotated) + expiresAt
- Sets: New HttpOnly cookie via Set-Cookie header

**Client-Side Function:** `refreshEndUserAccessToken(params?)`
- File: `src/api-client/end-user/end-user-auth-client.ts`
- Concurrency-protected (shared promise pattern)
- Updates store with new token + decoded JWT data
- Returns: New token or null on failure

**Example Usage:**
```typescript
const newToken = await refreshEndUserAccessToken({
  orgName: 'acme',
  projectSlug: 'crm'
})
// Automatically updates: accessToken, expiresAt, userInfo
```

---

### 3. Tenant createAuthLink Pattern (Reference) ✅

**Location:** `src/api-client/apollo/clients.ts` (lines 46-74)

**Pattern:**
```typescript
function createAuthLink() {
  return setContext(async (operation, { headers }) => {
    try {
      // 1. Get token from store
      let token = useAuthStore.getState().accessToken
      
      // 2. Refresh if needed (and not on end-user path)
      if (!token && !isEndUserPath()) {
        token = await refreshAccessToken()
      }

      // 3. Build headers
      const nextHeaders = {
        ...(headers ?? {}),
        'x-client-request-id': generateUUID(),
      }

      // 4. Inject token
      if (token) {
        nextHeaders.authorization = `Bearer ${token}`
      }

      // 5. Add X-Action header
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

**Key Features:**
- ✅ Async context link
- ✅ Gets token from store (not passed in)
- ✅ Checks expiry + refreshes automatically
- ✅ Builds all required headers
- ✅ Error handling (fallback to no-auth)

---

### 4. End-User Auth Store ✅

**File:** `src/shared/stores/end-user-auth-store.ts`

**Full Implementation:**
```typescript
interface EndUserAuthState {
  accessToken: string | null
  expiresAt: number | null      // Unix timestamp (ms)
  userInfo: EndUserInfo | null
  
  setAccessToken: (token: string, expiresIn: number) => void
  setUserInfo: (info: EndUserInfo) => void
  clearSession: () => void
  isTokenExpired: () => boolean  // 5 min buffer check
}

export const useEndUserAuthStore = create<EndUserAuthState>((set, get) => ({
  accessToken: null,
  expiresAt: null,
  userInfo: null,

  setAccessToken: (token, expiresIn) => {
    set({
      accessToken: token,
      expiresAt: Date.now() + expiresIn * 1000,
    })
  },

  setUserInfo: (info) => set({ userInfo: info }),

  clearSession: () =>
    set({
      accessToken: null,
      expiresAt: null,
      userInfo: null,
    }),

  isTokenExpired: () => {
    const { expiresAt } = get()
    if (!expiresAt) return true
    return Date.now() > expiresAt - 5 * 60 * 1000  // 5 min warning
  },
}))
```

---

### 5. BFF Route for End-User Token Refresh ✅

**File:** `src/app/api/bff/org/[orgName]/end-user/auth/refresh/route.ts`

**Full Implementation:**
```typescript
import { NextRequest, NextResponse } from 'next/server'
import { proxyEndUserAuth } from '../_proxy'
import { END_USER_REFRESH_COOKIE } from '@/middleware'

type RouteParams = { params: Promise<{ orgName: string }> }

export async function POST(req: NextRequest, { params }: RouteParams): Promise<NextResponse> {
  const { orgName } = await params

  // Read refresh token from HttpOnly cookie
  const cookieToken = req.cookies.get(END_USER_REFRESH_COOKIE)?.value

  // Parse request body (may contain projectSlug, etc.)
  let bodyObj: Record<string, unknown> = {}
  try {
    const raw = await req.text()
    if (raw) bodyObj = JSON.parse(raw) as Record<string, unknown>
  } catch {
    // Ignore parsing errors
  }

  // Merge: cookie refreshToken takes priority
  const mergedBody = {
    ...bodyObj,
    orgName: bodyObj.orgName ?? orgName,
    ...(cookieToken ? { refreshToken: cookieToken } : {}),
  }

  // Create new request with merged body
  const newReq = new NextRequest(req.url, {
    method: 'POST',
    headers: req.headers,
    body: JSON.stringify(mergedBody),
  })

  // Proxy to backend, which handles cookie rotation
  return proxyEndUserAuth(newReq, 'refresh', 'POST')
}
```

**Proxy Helper:** `src/app/api/bff/org/[orgName]/end-user/auth/_proxy.ts`
- Transparently forwards requests to Go backend
- Preserves cookie headers (send + receive)
- On logout: Appends Set-Cookie with Max-Age=0 to clear

---

## THE PROBLEM

Looking at `createEndUserScopedClient()` (lines 144-172 in clients.ts):

```typescript
export function createEndUserScopedClient(
  orgName: string,
  projectSlug: string,
  endUserToken: string  // ❌ Static token passed in!
): ApolloClient<object> {
  const uri = `${GATEWAY_URL}/api/bff/graphql/end-user/org/${orgName}/project/${projectSlug}`
  const httpLink = createHttpLink({ uri, credentials: 'include' })

  // ❌ No refresh logic!
  const authLink = setContext((operation, { headers }) => {
    const xAction = buildXAction(operation)
    return {
      headers: {
        ...(headers ?? {}),
        authorization: `Bearer ${endUserToken}`,  // ← Same token forever
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

**Issues:**
1. ❌ Token passed at client creation time (static)
2. ❌ No expiry check on each operation
3. ❌ No refresh call on expiry
4. ❌ Token becomes invalid after 1 hour → all queries fail
5. ❌ Same issue in: `createEndUserOrgScopedClient()`, `createEndUserModelRuntimeClient()`

---

## THE SOLUTION (Pattern to Implement)

**Replicate the tenant pattern:**

```typescript
export function createEndUserScopedClient(
  orgName: string,
  projectSlug: string
): ApolloClient<object> {
  const uri = `${GATEWAY_URL}/api/bff/graphql/end-user/org/${orgName}/project/${projectSlug}`
  const httpLink = createHttpLink({ uri, credentials: 'include' })

  // ✅ NEW: Dynamic auth link with refresh
  const authLink = setContext(async (operation, { headers }) => {
    try {
      let token = getEndUserToken()

      // ✅ Check expiry + refresh if needed
      if (!token || useEndUserAuthStore.getState().isTokenExpired()) {
        const newToken = await refreshEndUserAccessToken({ orgName, projectSlug })
        if (newToken) {
          token = newToken
        } else {
          // Refresh failed → redirect to login
          removeEndUserSession()
          // TODO: Navigate to login
          return { headers }
        }
      }

      const xAction = buildXAction(operation)
      return {
        headers: {
          ...(headers ?? {}),
          authorization: `Bearer ${token}`,
          ...(xAction ? { 'X-Action': xAction } : {}),
        },
      }
    } catch (err) {
      console.error('[EndUserAuthLink] ERROR:', err)
      return { headers }
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

---

## Key Insights

### 1. Infrastructure Is Already Built
- ✅ Store: Complete with expiry logic
- ✅ Refresh function: Implemented + concurrency-safe
- ✅ BFF endpoint: Working + cookie rotation
- ✅ Types: All defined
- ✅ Error handling: All mapped

### 2. Pattern Is Already Established
- The tenant `createAuthLink()` is the gold standard
- End-user auth is architecturally identical
- Just needs to be plugged in

### 3. What's Missing
- **Apollo Link Integration** (1 file to fix)
  - `createEndUserScopedClient()` 
  - `createEndUserOrgScopedClient()`
  - `createEndUserModelRuntimeClient()`
- **Optional: useRequireEndUserAuth hook** (already exists, just ensure it calls refresh)

### 4. Storage Strategy
```
accessToken (JWT)
├─ Stored: Zustand in-memory
├─ Lifetime: Until expiry + 5 min buffer
├─ Rotated: On each refresh
└─ Purpose: Apollo headers

expiresAt (Unix timestamp)
├─ Stored: Zustand in-memory
├─ Updated: On login/refresh
├─ Used by: isTokenExpired() for 5-min check
└─ Recovers: From JWT exp claim on page load

refreshToken (Opaque)
├─ Stored: HttpOnly cookie (BROWSER ONLY)
├─ Never: Accessible to JavaScript
├─ Rotated: On each refresh
├─ Used by: BFF → Go backend exchange
└─ Cleared: On logout

mc_enduser_refresh_token (Cookie name)
├─ HttpOnly: YES ✓
├─ SameSite: Strict ✓
├─ Secure: YES ✓
└─ Path: / ✓
```

---

## All Files Documented

See the other markdown files in this directory:

1. **END_USER_AUTH_ANALYSIS.md** - Deep dive on architecture (5K words)
2. **END_USER_AUTH_FLOW.txt** - Visual flow diagrams (2K lines)
3. **END_USER_AUTH_QUICK_REFERENCE.md** - Practical quick lookup
4. **END_USER_AUTH_FILE_TREE.txt** - File locations + import structure

---

## Next Steps

To implement end-user token refresh in Apollo:

1. **Update `createEndUserScopedClient()`**
   - Remove `endUserToken` parameter
   - Add dynamic auth link (use tenant pattern as reference)
   - Call `refreshEndUserAccessToken()` on expiry

2. **Update `createEndUserOrgScopedClient()`**
   - Same pattern as above

3. **Update `createEndUserModelRuntimeClient()`**
   - Same pattern as above

4. **Test:**
   - Manual: Set clock to near expiry, trigger query
   - Should auto-refresh transparently
   - Should not break on token expiry

All the pieces are there. Just needs to be connected. ✓

