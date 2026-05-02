# ModelCraft Frontend Authentication System
## Recommendations & Implementation Roadmap (Updated)

---

## Executive Summary

Previous analysis identified 8 design issues with a 7/10 architecture score. However, deeper investigation discovered that the "duplicate routes" problem (#5) is actually an **intentional dual-layer architecture pattern** supporting multiple client types (web, mobile, desktop, server-to-server). This finding elevates the overall score to **8/10** and shifts recommendations from code refactoring toward documentation and optimization.

---

## Priority-Ranked Recommendations

### 🔴 HIGH PRIORITY (Implement Next Sprint)

#### 1. Create Architecture Decision Record (ADR) for Dual API Layer Pattern

**Problem**: The sophisticated dual API layer pattern (`/api/bff/org/[orgName]/end-user/auth/*` + `/end-user/auth/*`) is completely undocumented, leading to potential confusion and incorrect consolidation attempts.

**Solution**: Create detailed ADR explaining:
- Why two layers exist (multi-client support)
- Which client type uses which layer
- Request/response flows for each layer
- Cookie vs. explicit token handling
- Migration paths for different integrations

**File**: Create `/ai-metadata/front/development/adr-dual-api-layer.md`

**Example Structure**:
```markdown
# ADR: Dual API Layer Pattern for End-User Authentication

## Context
ModelCraft needs to support multiple client types with different token management capabilities.

## Decision
Implement two API layers:
1. Org-scoped routes (/api/bff/org/[orgName]/end-user/auth/*) → Browser clients
2. Convenience routes (/end-user/auth/*) → Mobile/desktop/external clients

## Consequences
✓ Web browsers use secure HttpOnly cookies
✓ Mobile apps can use explicit token management
✓ Backend integration simplified (JWT extraction)
✗ Two parallel implementations to maintain
✗ Must document routing logic clearly
```

**Estimated Effort**: 4 hours (document creation + team review)

---

#### 2. Add Inline Documentation to Key Route Handlers

**Problem**: Developers working on `/end-user/auth/*` routes won't understand why they decode JWT and forward internally.

**Solution**: Add comprehensive comments to:
- `/end-user/auth/me/route.ts` - Explain JWT org_name extraction
- `/end-user/auth/refresh/route.ts` - Explain body-based orgName routing
- `/end-user/auth/logout/route.ts` - Explain fallback pattern

**Example Addition**:
```typescript
/**
 * Convenience Route: GET /end-user/auth/me
 * 
 * This route acts as a gateway adapter for non-web clients (mobile, desktop)
 * that don't have orgName in the request path.
 * 
 * Flow:
 * 1. Extract orgName from Bearer JWT claim (org_name)
 * 2. Validate orgName is present
 * 3. Forward to org-scoped route: /api/bff/org/{orgName}/end-user/auth/me
 * 
 * Why this exists:
 * - Web browsers: Use /api/bff/org/{orgName}/end-user/auth/me (known orgName)
 * - Mobile/Desktop: Use /end-user/auth/me (orgName only in JWT)
 * 
 * Token Handling:
 * - Input: Bearer token (JWT with org_name claim)
 * - Output: User info (id, username, createdAt)
 * - Cookie: Not used at convenience layer (mobile clients can't handle)
 * 
 * Security: Bearer token is validated by backend API
 */
export async function GET(req: NextRequest) {
  // implementation...
}
```

**Files to Update**:
- `/src/app/end-user/auth/me/route.ts`
- `/src/app/end-user/auth/refresh/route.ts`
- `/src/app/end-user/auth/logout/route.ts`
- `/src/app/api/bff/org/[orgName]/end-user/auth/_proxy.ts` (explain cookie strategy)

**Estimated Effort**: 3 hours (documentation + code comments)

---

#### 3. Refactor Concurrent Refresh Deduplication into Shared Utility

**Problem**: Developer auth and end-user auth have ~90% identical refresh concurrency handling code (separate `_isRefreshing` flags and `_refreshPromise` in each module).

**Solution**: Extract common logic into `src/api-client/shared/concurrent-refresh-helper.ts`

**Before** (Current - Duplicated in 2 places):
```typescript
// In auth-client.ts
let _isRefreshing = false
let _refreshPromise: Promise<string | null> | null = null

export async function refreshAccessToken(): Promise<string | null> {
  if (_isRefreshing && _refreshPromise) return _refreshPromise
  _isRefreshing = true
  _refreshPromise = (async () => {
    // actual refresh logic
    return token
  })()
  return _refreshPromise
}

// In end-user-auth-client.ts (IDENTICAL PATTERN)
let _isRefreshing = false
let _refreshPromise: Promise<string | null> | null = null

export async function refreshEndUserAccessToken(): Promise<string | null> {
  if (_isRefreshing && _refreshPromise) return _refreshPromise
  _isRefreshing = true
  _refreshPromise = (async () => {
    // actual refresh logic
    return token
  })()
  return _refreshPromise
}
```

**After** (Refactored):
```typescript
// src/api-client/shared/concurrent-refresh-helper.ts
export class ConcurrentRefreshHelper {
  private isRefreshing = false
  private refreshPromise: Promise<string | null> | null = null

  async execute(fn: () => Promise<string | null>): Promise<string | null> {
    if (this.isRefreshing && this.refreshPromise) {
      return this.refreshPromise
    }
    
    this.isRefreshing = true
    this.refreshPromise = (async () => {
      try {
        return await fn()
      } finally {
        this.isRefreshing = false
        this.refreshPromise = null
      }
    })()
    
    return this.refreshPromise
  }
}

// In auth-client.ts
const refreshHelper = new ConcurrentRefreshHelper()

export async function refreshAccessToken(): Promise<string | null> {
  return refreshHelper.execute(async () => {
    // actual refresh logic only
    return token
  })
}

// In end-user-auth-client.ts
const endUserRefreshHelper = new ConcurrentRefreshHelper()

export async function refreshEndUserAccessToken(): Promise<string | null> {
  return endUserRefreshHelper.execute(async () => {
    // actual refresh logic only
    return token
  })
}
```

**Benefits**:
- ✓ Reduces code duplication by ~50 lines
- ✓ Single point of maintenance for concurrent logic
- ✓ Consistent behavior across both auth systems
- ✓ Easier to add new auth clients in future

**Files to Create/Modify**:
- Create: `/src/api-client/shared/concurrent-refresh-helper.ts`
- Modify: `/src/api-client/auth/auth-client.ts`
- Modify: `/src/api-client/end-user/end-user-auth-client.ts`

**Estimated Effort**: 6 hours (refactoring + comprehensive testing)

---

### 🟡 MEDIUM PRIORITY (Next 2 Sprints)

#### 4. Centralize JWT Decoding Logic

**Problem**: JWT decoding logic appears in 4 separate places with slight variations:
- `auth-client.ts` - `decodeJWT()` with multiple field variants
- `end-user-auth-client.ts` - `decodeEndUserJWT()` with specific fields
- `end-user/auth/me/route.ts` - Inline decoding for orgName
- `end-user/auth/refresh/route.ts` - Inline decoding for orgName

**Solution**: Create `/src/api-client/shared/jwt-decoder.ts` with typed helpers

```typescript
// src/api-client/shared/jwt-decoder.ts
export interface JWTDecoderOptions {
  verify?: boolean  // Future: add signature verification
  throwOnError?: boolean
}

export function decodeJWT<T = Record<string, unknown>>(
  token: string,
  options?: JWTDecoderOptions
): T | null {
  try {
    const base64Url = token.split('.')[1]
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    )
    return JSON.parse(jsonPayload) as T
  } catch (err) {
    if (options?.throwOnError) throw err
    return null
  }
}

// Typed helpers for specific token types
export function decodeAuthJWT(token: string) {
  return decodeJWT<{
    exp: number
    sub: string
    username?: string
    userName?: string
    phone?: string
  }>(token)
}

export function decodeEndUserJWT(token: string) {
  return decodeJWT<{
    sub: string
    org_name: string
    project_slug: string
    role: string
    exp: number
    iat: number
  }>(token)
}

// Extract specific claims with defaults
export function extractOrgNameFromToken(token: string): string | null {
  const decoded = decodeEndUserJWT(token)
  return decoded?.org_name ?? null
}
```

**Usage in route handlers**:
```typescript
// Before
const normalized = payload.replace(/-/g, '+').replace(/_/g, '/')
const json = Buffer.from(normalized, 'base64').toString('utf-8')
const decoded = JSON.parse(json) as { org_name?: string }

// After
const decoded = decodeEndUserJWT(token)
const orgName = decoded?.org_name
```

**Files to Create/Modify**:
- Create: `/src/api-client/shared/jwt-decoder.ts`
- Modify: `/src/api-client/auth/auth-client.ts`
- Modify: `/src/api-client/end-user/end-user-auth-client.ts`
- Modify: `/src/app/end-user/auth/me/route.ts`
- Modify: `/src/app/end-user/auth/refresh/route.ts`

**Estimated Effort**: 5 hours (extraction + testing)

---

#### 5. Add Client Type Documentation to TypeScript Types

**Problem**: Type definitions don't indicate which client types use which flows.

**Solution**: Add JSDoc comments to key types and interfaces

```typescript
// src/types/end-user-auth.ts

/**
 * End-user authentication response after successful login/refresh
 * 
 * Client Types:
 * - Web Browser: Receives in response body, processes accessToken,
 *   backend sets mc_enduser_refresh_token HttpOnly cookie
 * - Mobile App: Receives in response body, must manually store accessToken,
 *   Set-Cookie header ignored by native app
 * 
 * @property accessToken JWT token for API authorization (Bearer token)
 * @property expiresAt ISO 8601 timestamp when token expires (from Go backend)
 * @property expiresIn Token lifetime in seconds (fallback if expiresAt missing)
 * @property refreshToken New refresh token (in response, moved to cookie by BFF)
 */
export interface EndUserAuthResponse {
  accessToken: string
  expiresAt?: string // ISO 8601 from Go backend
  expiresIn?: number // seconds
  refreshToken?: string
}

/**
 * Request to refresh end-user access token
 * 
 * Client Types:
 * - Web Browser: orgName from URL route, refreshToken from HttpOnly cookie
 *   (browser auto-includes in request)
 * - Mobile App: orgName from app state/JWT, refreshToken must be 
 *   manually included in request body
 */
export interface EndUserRefreshRequest {
  orgName: string
  projectSlug?: string // client context
  refreshToken?: string // only needed if not in cookie
}
```

**Files to Modify**:
- `/src/types/end-user-auth.ts`
- `/src/types/auth.ts`

**Estimated Effort**: 2 hours (documentation only)

---

#### 6. Document Cookie Strategy Differences in Architecture Guide

**Problem**: `ai-metadata/front/development/architecture.md` doesn't explain why cookie handling differs between auth systems.

**Solution**: Add new section explaining browser-specific considerations

```markdown
## Cookie Management Strategy

### Developer Authentication
- **Cookie Name**: `mc_refresh_token`
- **HttpOnly**: true
- **Secure**: true (production only)
- **SameSite**: lax
- **Scope**: Global (entire app)
- **Lifecycle**: 30 days

### End-User Authentication  
- **Cookie Name**: `mc_enduser_refresh_token`
- **HttpOnly**: true
- **Secure**: true (production only)
- **SameSite**: lax
- **Scope**: Org-specific (but set globally, filtered by middleware)
- **Lifecycle**: 30 days

### Why HttpOnly?
HttpOnly cookies cannot be accessed via JavaScript, protecting against 
XSS attacks that might steal tokens. Browser automatically includes them 
in requests, so client code never needs to handle them explicitly.

### Why SameSite=lax?
Prevents CSRF attacks by only including cookies in same-site requests 
and top-level navigations.

### Design Pattern: Web vs. Non-Web Clients
```

**File to Modify**: `/ai-metadata/front/development/architecture.md`

**Estimated Effort**: 3 hours (research + documentation)

---

### 🟢 LOW PRIORITY (Future Improvements)

#### 7. Create Mobile Client Integration Guide

**Problem**: Mobile team doesn't have clear documentation on how to integrate with the convenience API layer.

**Solution**: Create `/ai-metadata/integration/mobile-auth-integration-guide.md`

```markdown
# Mobile Client Integration Guide

## Quick Start

### Option 1: Using Convenience Routes (Recommended for Simple Cases)

```typescript
// 1. Login
const loginRes = await fetch('/end-user/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    email: 'user@example.com',
    password: 'password',
    orgName: 'acme'
  })
})

const loginData = await loginRes.json() // { accessToken, refreshToken }

// 2. Store tokens securely
await saveToSecureStorage('accessToken', loginData.accessToken)
await saveToSecureStorage('refreshToken', loginData.refreshToken)

// 3. Use accessToken for API calls
const userRes = await fetch('/end-user/auth/me', {
  headers: { 'Authorization': `Bearer ${accessToken}` }
})

// 4. Refresh when needed
const refreshRes = await fetch('/end-user/auth/refresh', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    orgName: 'acme',
    refreshToken: storedRefreshToken
  })
})

const refreshData = await refreshRes.json() // { accessToken, refreshToken }
```

### Option 2: Using Org-Scoped Routes (If orgName Known Early)

Similar pattern but routes are `/api/bff/org/{orgName}/end-user/auth/*`

## Error Handling

Handle these specific error codes:
- `INVALID_CREDENTIALS` - Wrong email/password
- `ORG_NOT_FOUND` - Invalid orgName
- `TOKEN_EXPIRED` - Refresh token expired, re-login needed
- `ROLE_INSUFFICIENT` - User lacks project access
```

**Estimated Effort**: 4 hours (guide creation + examples + review)

---

#### 8. Add Type-Safe API Client for Frontend Integration

**Problem**: No type-safe API client helper; developers must manually construct fetch calls.

**Solution**: Create `/src/api-client/end-user/typed-client.ts` with full typing

```typescript
// src/api-client/end-user/typed-client.ts
export class TypedEndUserAuthClient {
  constructor(private baseUrl: string = '/api/bff/org') {}

  async login(
    orgName: string,
    email: string,
    password: string
  ): Promise<EndUserAuthResponse> {
    const res = await fetch(
      `${this.baseUrl}/${encodeURIComponent(orgName)}/end-user/auth/login`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      }
    )
    return res.json()
  }

  async getMe(orgName: string, token: string): Promise<EndUserMeResponse> {
    const res = await fetch(
      `${this.baseUrl}/${encodeURIComponent(orgName)}/end-user/auth/me`,
      {
        headers: { 'Authorization': `Bearer ${token}` },
      }
    )
    return res.json()
  }

  // ... other methods with full typing
}
```

**Estimated Effort**: 6 hours (implementation + documentation)

---

## Implementation Roadmap

### Phase 1: Documentation (Week 1-2)
1. ✓ Create ADR for dual API layer pattern
2. ✓ Add inline comments to route handlers
3. ✓ Update architecture guide with cookie strategy
4. **Time**: 10 hours

### Phase 2: Deduplication (Week 3-4)
5. ✓ Extract concurrent refresh helper
6. ✓ Centralize JWT decoding
7. **Time**: 11 hours

### Phase 3: Enhancement (Week 5-6)
8. ✓ Add type documentation
9. ✓ Mobile integration guide (optional if mobile team exists)
10. ✓ Type-safe client wrapper (optional, nice-to-have)
11. **Time**: 10 hours (core), +10 hours (optional)

### Total Effort: 31 hours (43 with optional tasks)

---

## Code Review Checklist

Use this checklist when reviewing future authentication changes:

- [ ] Does the change affect both developer and end-user auth systems?
  - If yes, must be implemented identically or with clear rationale for differences
- [ ] Does the change add a new API route?
  - If yes, must specify: is it web-only or multi-client?
  - If multi-client, must have both org-scoped and convenience variants
- [ ] Does the change modify cookie handling?
  - If yes, must update both `_proxy.ts` and relevant documentation
- [ ] Does the change add JWT decoding logic?
  - If yes, must use centralized `jwt-decoder.ts` utilities (after Phase 2)
- [ ] Does the change add concurrent request logic?
  - If yes, must use `ConcurrentRefreshHelper` (after Phase 2)
- [ ] Are related types documented with client type usage?
  - If no, add JSDoc comments explaining web vs. non-web implications

---

## Testing Checklist

### Unit Tests
- [ ] Concurrent refresh returns same promise for parallel calls
- [ ] Cookie is set only for successful responses
- [ ] Cookie is cleared on logout
- [ ] JWT decoding handles malformed tokens gracefully
- [ ] orgName extraction works from both JWT and body

### Integration Tests
- [ ] Web browser flow: orgName in route, refreshToken in cookie
- [ ] Mobile flow: orgName in body, refreshToken in body
- [ ] Token refresh updates both accessToken and expiry
- [ ] Middleware correctly allows/blocks based on cookie presence

### E2E Tests (Recommended)
- [ ] Full web login → refresh → access protected page → logout flow
- [ ] Mobile login → refresh → access API → logout flow
- [ ] Concurrent requests during token expiry
- [ ] Token expiry triggers refresh before access

---

## Success Metrics

After implementing these recommendations:

1. **Code Quality**
   - Duplicate code reduced from ~90% to <10%
   - No console warnings from TypeScript strict mode
   - 100% of JWT decoding uses centralized utility

2. **Developer Experience**
   - New developers can understand dual API layer in <2 hours (currently: undefined)
   - Mobile team has clear integration guide
   - Route handler comments explain rationale

3. **Architecture Score**
   - Current: 8/10
   - Target: 9/10 (after all recommendations)
   - Remaining -1 point: middleware regex patterns (complex but functional)

4. **Documentation Coverage**
   - All route handlers have JSDoc comments: ✓
   - Dual API layer has ADR: ✓
   - Cookie strategy documented: ✓
   - Client type requirements in types: ✓

---

## Appendix: Architecture Score Breakdown

| Category | Current | Target | Improvement |
|----------|---------|--------|-------------|
| Auth separation | 9/10 | 9/10 | - |
| Cookie security | 9/10 | 9/10 | - |
| Token refresh | 8/10 | 9/10 | Remove duplication |
| API routing | 8/10 | 9/10 | Add documentation |
| Type safety | 8/10 | 9/10 | Add client type hints |
| Middleware logic | 7/10 | 8/10 | Refactor regex (future) |
| Documentation | 5/10 | 9/10 | ⭐ Major improvement |
| Configuration | 8/10 | 9/10 | Add migration guide |
| **OVERALL** | **8/10** | **9/10** | +1 point |

---

## Conclusion

The ModelCraft frontend dual authentication system demonstrates sophisticated architecture supporting multiple client types. The primary opportunity for improvement is **documentation and consolidation of duplicate patterns**, not fundamental redesign. Implementing the high-priority recommendations (ADR + inline comments + concurrent helper) in the next sprint will significantly improve code maintainability and developer experience.

