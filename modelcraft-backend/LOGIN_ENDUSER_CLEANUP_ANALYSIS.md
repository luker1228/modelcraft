# ModelCraft Backend: Login/LoginEndUser Cleanup Analysis

## Executive Summary

Based on analysis of the codebase, **Login and LoginEndUser can be merged into a single method** with the following caveats:

1. **No direct test coverage gap** — There are NO tests that specifically test `LoginEndUser` or call `token_service_enduser.go` methods
2. **Deprecated fields exist but minimal** — `Phone` field in `LoginRequest` is marked deprecated and has backward-compat handling
3. **Significant code path unification** — Both methods now use the same `endUserRepoFactory` path and identical logic flow
4. **Breaking change risk is LOW** — Existing tests would continue to pass; only deprecation documentation changes needed

---

## Detailed Findings

### A. Test Coverage for Login

**Token Service Tests (token_service_test.go):**

- **Login Tests (lines 342-390):** 3 concrete tests
  - `TestTokenService_Login_Success` (344-362)
  - `TestTokenService_Login_PhoneNotFound` (364-375)
  - `TestTokenService_Login_WrongPassword` (377-390)

- **All tests use:** `svc.Login(ctx, LoginCommand{Identifier, IdentifierType, Password})`
- **Phone-based login tested:** `IdentifierTypePhone` constant used
- **Username-based login:** NOT EXPLICITLY TESTED (default IdentifierType assumed USERNAME)

**LoginEndUser Test Coverage:**
- **ZERO tests** — No `TestTokenService_LoginEndUser_*` functions exist
- **No calls** to `LoginEndUser` method in token_service_test.go
- The only reference is a comment (line 531) about org creation spy

**Action Required:**
- Need to add 3+ tests for `LoginEndUser` with:
  - phone-scoped login (with orgName)
  - username-scoped login (with orgName)
  - global username lookup (no orgName)

---

### B. LoginRequest (Generated OpenAPI Struct)

**Location:** `modelcraft-backend/internal/interfaces/http/generated/server.gen.go:252-264`

**Current Fields:**
```go
type LoginRequest struct {
    Identifier     string                      `json:"identifier"`
    IdentifierType *LoginRequestIdentifierType `json:"identifierType,omitempty"`
    Password       string                      `json:"password"`
    Phone          *string                     `json:"phone,omitempty"`  // DEPRECATED
}
```

**Key Observations:**
- `Identifier` is the NEW primary field (string, required)
- `IdentifierType` is optional enum (PHONE | USERNAME)
- `Phone` field still exists but **marked DEPRECATED** in comment (line 262)
- **Backward compat:** handled in handler (lines 145-148)

**Status:** Phone field should be removed from OpenAPI spec (if not already removed)

---

### C. LoginCommand Structure

**Location:** `modelcraft-backend/internal/app/auth/commands.go:39-49`

```go
type LoginCommand struct {
    Identifier     string         // NEW: phone or username
    IdentifierType IdentifierType // NEW: explicit type
    Password       string
    Phone          string         // Deprecated: backward compat
}
```

**Deprecated Field Usage:**
- **Phone field used ONLY in handler backward-compat logic** (handler.go:146-148)
- **Token service resolves Identifier → idType** (token_service.go:352-361)
- **No non-deprecated code path uses Phone field directly** in service layer

---

### D. LoginEndUserCommand Structure

**Location:** `modelcraft-backend/internal/app/auth/commands.go:86-93`

```go
type LoginEndUserCommand struct {
    OrgName        string
    Identifier     string         // NEW: takes priority
    IdentifierType IdentifierType
    Username       string         // Deprecated: backward compat ("向后兼容")
    Password       string
}
```

**Username Field Usage:**
- **Used ONLY in backward-compat logic** (token_service_enduser.go:31-38)
- **Resolution logic:** `Identifier` takes priority if `IdentifierType` is set
- **Otherwise:** Falls back to `Username` field

**Non-deprecated path:**
- Set `IdentifierType` + `Identifier` → uses new path exclusively
- Leave `IdentifierType` empty → `Username` field is the effective identifier

---

### E. Handler Behavior When identifierType is nil

**Location:** `modelcraft-backend/internal/interfaces/http/handlers/auth/handler.go:134-143`

```go
// Map identifierType from generated enum to app enum
if req.IdentifierType != nil {
    switch *req.IdentifierType {
    case generated.USERNAME:
        cmd.IdentifierType = appAuth.IdentifierTypeUsername
    default:
        cmd.IdentifierType = appAuth.IdentifierTypePhone
    }
} else {
    cmd.IdentifierType = appAuth.IdentifierTypePhone  // DEFAULT: PHONE
}

// Backward compat: if identifier is empty but phone is provided, use phone
if cmd.Identifier == "" && req.Phone != nil {
    cmd.Phone = *req.Phone
}
```

**When identifierType is nil:**
- Falls through to `else` block → defaults to `IdentifierTypePhone` (line 142)
- If `identifier` is empty AND `phone` is provided → uses phone (backward compat)
- Otherwise → uses identifier with PHONE type

---

### F. Test Breaking Changes Risk

**Existing Test Suite Impact Analysis:**

1. **Login tests (3 total):** ✅ NO BREAKING CHANGES
   - All tests pass `Identifier` + `IdentifierType` explicitly
   - Can convert to unified method signature unchanged

2. **Register tests (4 total):** ✅ NO BREAKING CHANGES
   - Do not call Login method
   - Would be unaffected by Login/LoginEndUser merge

3. **Refresh + Logout tests (5 total):** ✅ NO BREAKING CHANGES
   - Independent from Login logic
   - Would be unaffected

4. **EndUser handlers (5 routes):** ⚠️ REQUIRES NEW TESTS
   - EndUserLogin (line 57): Currently calls LoginEndUser
   - CLILogin (line 241): Currently calls LoginEndUser
   - No tests verify these paths

---

### G. Test Helper Functions

**Mock Repositories Provided (token_service_test.go):**

```go
newMockRefreshTokenRepo()      // mockRefreshTokenRepo
newMockUserRepo()               // mockUserRepo with:
                                //   - GetByExternalID
                                //   - GetByID
                                //   - GetByPhone
                                //   - GetByName
                                //   - ExistsByPhone, ExistsByName, ExistsByExternalID

newMockProfileRepo()            // mockProfileRepo
mockPasswordHasher{}            // Hash/Verify

createTestService()             // Full TokenService with all mocks
registerTestUser()              // Helper: Register + return result

createTestServiceWithOrgSpy()   // For org creation assertions
```

**Missing:** EndUser repository mocks for testing `LoginEndUser` behavior

---

## Test Coverage Gaps & Recommendations

| Category | Gap | Severity | Fix |
|----------|-----|----------|-----|
| **Login method** | No USERNAME identifier test | Medium | Add `TestTokenService_Login_ByUsername` |
| **LoginEndUser method** | ZERO test coverage | High | Add 3+ tests (phone/username/global) |
| **Backward compat** | Phone field not tested | Medium | Add `TestTokenService_Login_DeprecatedPhone` |
| **EndUser handlers** | EndUserLogin/CLILogin not tested | High | Add integration tests |
| **Merged method** | Post-merge: no regression test | High | Add unified login test matrix |

---

## Deprecated Fields Currently in Production

### In LoginRequest (generated):
- `Phone` field: ✅ Marked deprecated (comment + omitempty tag)
- Status: Can be removed from OpenAPI spec

### In LoginCommand:
- `Phone` field: ✅ Marked deprecated (inline comment)
- Status: Safe to remove (not used in non-deprecated paths)

### In LoginEndUserCommand:
- `Username` field: ✅ Marked deprecated ("向后兼容")
- Status: Safe to remove after deprecation window

### In Handler:
- Backward-compat logic (lines 145-148): ✅ Handled
- Only activates if `identifier == ""` AND `phone != nil`

---

## Merge Strategy Recommendations

### Phase 1: Unify Method Signatures
1. Create unified `LoginCommand` that supports:
   - OrgName (optional, for enduser scope)
   - Identifier (required)
   - IdentifierType (optional, defaults to USERNAME for backward compat with old clients)
   - Password (required)

2. Merge `LoginEndUser` logic into `Login` method:
   - If OrgName is provided → org-scoped lookup
   - If OrgName is empty → global lookup
   - Same identifier resolution logic

### Phase 2: Add Test Coverage
1. Add tests for merged method with matrix:
   ```
   - Phone login (global scope)
   - Username login (global scope)
   - Phone login (org scope)
   - Username login (org scope)
   - Invalid credentials
   - Account disabled (enduser-specific)
   ```

2. Add integration tests for both handler routes:
   - POST /api/auth/login
   - POST /api/end-user/auth/login

### Phase 3: Deprecation
1. Keep old method signatures as deprecated wrappers (6-month window)
2. Document breaking change in changelog
3. Remove deprecated code after window

---

## Code Duplication Analysis

**Current Duplication Level: HIGH**

### token_service.go `Login()` (lines 342-438)
- 97 lines
- Uses `endUserRepoFactory` (line 349)
- Resolves identifier type → calls GetByPhoneGlobal/GetByUsernameGlobal
- Generates refresh token
- Issues JWT access token

### token_service_enduser.go `LoginEndUser()` (lines 22-115)
- 93 lines
- Uses `endUserRepoFactory` with OrgName scope (line 27)
- Resolves identifier type → calls GetByPhone/GetByUsername
- **IDENTICAL:** Generate refresh token logic (lines 84-104)
- **IDENTICAL:** Issue JWT logic (line 79)
- **ADDITIONAL:** Account status checks (line 71)

**Merge Savings:** ~80 lines of duplicated password verification, token generation, JWT issuance

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| Existing tests fail post-merge | Very Low | High | Add test matrix before merge |
| Handler behavior changes | Very Low | High | Handler unit tests + integration tests |
| Backward compat breaks | Low | Medium | Keep deprecated wrappers for 6 months |
| Enduser feature regression | Medium | High | Add LoginEndUser test suite FIRST |

---

## Conclusion

✅ **Login and LoginEndUser CAN be safely merged** IF:

1. **Add test coverage for LoginEndUser BEFORE merge** (currently zero)
2. **Add tests for Login with USERNAME identifier** (currently not covered)
3. **Keep handler backward-compat logic unchanged** (Phone field → Identifier fallback)
4. **Test matrix post-merge:** All combinations of (PHONE|USERNAME, global|scoped, active|disabled)

📝 **Estimated Effort:**
- Pre-merge: 4-6 new tests + test helper mocks
- Merge: 50-100 lines consolidated
- Post-merge: 0 additional tests (all pre-merge)

