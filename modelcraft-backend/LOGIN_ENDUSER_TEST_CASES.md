# Pre-Merge Test Cases for Login/LoginEndUser Consolidation

## Overview
This document specifies all test cases that MUST be added before merging `Login` and `LoginEndUser` methods.

**Current Status:**
- Login: 3 tests (phone-based only)
- LoginEndUser: 0 tests
- Missing: 1 + 3 + backward-compat = 5+ new tests

---

## Test 1: Login by USERNAME (UNIT TEST)

**File:** `token_service_test.go`

**Test Name:** `TestTokenService_Login_ByUsername`

**Purpose:** Verify Login works with USERNAME identifier type (not just PHONE)

**Setup:**
1. Create test service with all mocks
2. Register a user with username "john_doe"
3. Call Login with Identifier="john_doe", IdentifierType=USERNAME, Password="securePassword1"

**Assertions:**
- ✓ Login succeeds
- ✓ result.UserID matches registered user
- ✓ result.AccessToken is not empty
- ✓ result.RefreshToken is not empty

**Code Template:**
```go
func TestTokenService_Login_ByUsername(t *testing.T) {
    svc, _, _, _, _ := createTestService(t)
    ctx := context.Background()
    
    // Register first
    registerTestUser(t, svc, "13800138000", "securePassword1")
    
    // Login by username
    result, err := svc.Login(ctx, LoginCommand{
        Identifier:     "testuser_8000",  // from registerTestUser helper
        IdentifierType: IdentifierTypeUsername,
        Password:       "securePassword1",
    })
    
    require.NoError(t, err)
    assert.NotEmpty(t, result.UserID)
    assert.NotEmpty(t, result.AccessToken)
    assert.NotEmpty(t, result.RefreshToken)
}
```

---

## Test 2: Login with Deprecated Phone Field (UNIT TEST)

**File:** `token_service_test.go`

**Test Name:** `TestTokenService_Login_DeprecatedPhoneField`

**Purpose:** Verify backward compatibility when Phone field is used instead of Identifier

**Setup:**
1. Create test service
2. Register user with phone
3. Call Login with Phone field populated, Identifier empty
4. (This mimics old clients using deprecated API)

**Assertions:**
- ✓ Login succeeds (via backward-compat path)
- ✓ User is correctly identified by phone
- ✓ Tokens are issued

**Code Template:**
```go
func TestTokenService_Login_DeprecatedPhoneField(t *testing.T) {
    svc, _, _, _, _ := createTestService(t)
    ctx := context.Background()
    
    registerTestUser(t, svc, "13800138000", "securePassword1")
    
    // Login using deprecated Phone field (not Identifier)
    result, err := svc.Login(ctx, LoginCommand{
        Phone:    "13800138000",  // Deprecated: use Phone instead of Identifier
        Password: "securePassword1",
        // Identifier left empty to trigger backward-compat path
    })
    
    require.NoError(t, err)
    assert.NotEmpty(t, result.UserID)
}
```

---

## Test 3: LoginEndUser by Phone with OrgName Scope (UNIT TEST)

**File:** `token_service_test.go`

**Test Name:** `TestTokenService_LoginEndUser_ByPhoneInOrgScope`

**Purpose:** Verify LoginEndUser works with phone identifier in org-scoped lookup

**Setup:**
1. Create test service
2. Need EndUser mock repository that supports GetByPhone with orgName
3. Register/create an EndUser
4. Call LoginEndUser with Identifier=phone, IdentifierType=PHONE, OrgName="org-slug"

**Assertions:**
- ✓ LoginEndUser succeeds
- ✓ User found in org scope
- ✓ Correct tokens issued
- ✓ Result includes OrgName and ExpiresAt

**NOTE:** Requires mock EndUser repository

**Code Template:**
```go
func TestTokenService_LoginEndUser_ByPhoneInOrgScope(t *testing.T) {
    svc, _, _, _, _ := createTestService(t)
    ctx := context.Background()
    
    // TODO: Setup EndUser mock repository
    // svc.endUserRepoFactory = mockFactory
    
    // Create test enduser somehow (TBD based on EndUser model)
    
    result, err := svc.LoginEndUser(ctx, LoginEndUserCommand{
        OrgName:        "test-org",
        Identifier:     "13800138000",
        IdentifierType: IdentifierTypePhone,
        Password:       "enduser_password",
    })
    
    require.NoError(t, err)
    assert.Equal(t, "test-org", result.OrgName)
    assert.NotEmpty(t, result.AccessToken)
    assert.NotEmpty(t, result.RefreshToken)
    assert.NotZero(t, result.ExpiresAt)
}
```

---

## Test 4: LoginEndUser by Username with OrgName Scope (UNIT TEST)

**File:** `token_service_test.go`

**Test Name:** `TestTokenService_LoginEndUser_ByUsernameInOrgScope`

**Purpose:** Verify LoginEndUser works with username identifier in org scope

**Setup:**
1. Create test service with EndUser mock
2. Create EndUser with username in specific org
3. Call LoginEndUser with Identifier=username, OrgName="org-slug"

**Assertions:**
- ✓ LoginEndUser succeeds
- ✓ User found within org scope only
- ✓ Correct tokens issued

**Code Template:**
```go
func TestTokenService_LoginEndUser_ByUsernameInOrgScope(t *testing.T) {
    svc, _, _, _, _ := createTestService(t)
    ctx := context.Background()
    
    // TODO: Setup EndUser mock repository
    // svc.endUserRepoFactory = mockFactory
    
    result, err := svc.LoginEndUser(ctx, LoginEndUserCommand{
        OrgName:        "test-org",
        Identifier:     "end_user_1",
        IdentifierType: IdentifierTypeUsername,
        Password:       "enduser_password",
    })
    
    require.NoError(t, err)
    assert.Equal(t, "test-org", result.OrgName)
}
```

---

## Test 5: LoginEndUser by Username Global Scope (No OrgName) (UNIT TEST)

**File:** `token_service_test.go`

**Test Name:** `TestTokenService_LoginEndUser_ByUsernameGlobal`

**Purpose:** Verify LoginEndUser works for global username lookup (no orgName provided)

**Setup:**
1. Create test service
2. Create EndUser
3. Call LoginEndUser with OrgName empty
4. Should still work and resolve user globally

**Assertions:**
- ✓ LoginEndUser succeeds
- ✓ result.OrgName is populated from user record
- ✓ Tokens issued

**Code Template:**
```go
func TestTokenService_LoginEndUser_ByUsernameGlobal(t *testing.T) {
    svc, _, _, _, _ := createTestService(t)
    ctx := context.Background()
    
    // TODO: Setup EndUser mock
    
    result, err := svc.LoginEndUser(ctx, LoginEndUserCommand{
        OrgName:        "",  // No org scope
        Identifier:     "global_user",
        IdentifierType: IdentifierTypeUsername,
        Password:       "enduser_password",
    })
    
    require.NoError(t, err)
    assert.NotEmpty(t, result.OrgName)  // Should be resolved from user
}
```

---

## Test 6: LoginEndUser with Disabled Account (UNIT TEST)

**File:** `token_service_test.go`

**Test Name:** `TestTokenService_LoginEndUser_DisabledAccount`

**Purpose:** Verify LoginEndUser rejects disabled/forbidden end-users

**Setup:**
1. Create test service
2. Create EndUser with IsForbidden=true
3. Try to login with this user

**Assertions:**
- ✗ LoginEndUser fails
- ✓ Error code is EndUserAccountDisabled or similar
- ✓ No tokens issued

**Code Template:**
```go
func TestTokenService_LoginEndUser_DisabledAccount(t *testing.T) {
    svc, _, _, _, _ := createTestService(t)
    ctx := context.Background()
    
    // TODO: Create disabled EndUser in mock
    
    _, err := svc.LoginEndUser(ctx, LoginEndUserCommand{
        OrgName:    "test-org",
        Identifier: "disabled_user",
        Password:   "password",
    })
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "ACCOUNT_DISABLED")
}
```

---

## Test 7: Handler Integration - POST /api/auth/login (INTEGRATION TEST)

**File:** (new file or separate integration test file)

**Test Name:** `TestHandleLogin_IdentifierTypeNil_DefaultsToPhone`

**Purpose:** Verify handler correctly defaults to PHONE when identifierType is nil

**Setup:**
1. Create HTTP test server with auth handler
2. Send POST /api/auth/login with JSON:
   ```json
   {
     "identifier": "13800138000",
     "password": "password",
     "identifierType": null
   }
   ```

**Assertions:**
- ✓ Response status 200
- ✓ Response contains accessToken
- ✓ Cookie mc_refresh_token is set (httpOnly)

**Code Template:**
```go
func TestHandleLogin_IdentifierTypeNil_DefaultsToPhone(t *testing.T) {
    // Setup handler with test service
    handler := NewHandler(testService, testCookieConfig)
    
    req := httptest.NewRequest("POST", "/api/auth/login", 
        strings.NewReader(`{"identifier":"13800138000","password":"password"}`))
    w := httptest.NewRecorder()
    
    handler.HandleLogin(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
    // Verify cookie
    cookies := w.Result().Cookies()
    assert.True(t, hasCookie(cookies, "mc_refresh_token"))
}
```

---

## Test 8: Handler Integration - POST /api/end-user/auth/login (INTEGRATION TEST)

**File:** (new integration test file for enduser handlers)

**Test Name:** `TestEndUserLogin_ByUsernameWithOrgName`

**Purpose:** Verify EndUserLogin handler correctly calls LoginEndUser

**Setup:**
1. Create HTTP test server with enduser auth handler
2. Send POST /api/end-user/auth/login with JSON:
   ```json
   {
     "orgName": "test-org",
     "identifier": "end_user",
     "identifierType": "USERNAME",
     "password": "password"
   }
   ```

**Assertions:**
- ✓ Response status 200
- ✓ Response contains userId, orgName, accessToken
- ✓ Cookie mc_refresh_token is set

**Code Template:**
```go
func TestEndUserLogin_ByUsernameWithOrgName(t *testing.T) {
    handler := NewAuthHandler(testService, testEndUserSvc, testJWTSigner, sharedHandler, logger)
    
    req := httptest.NewRequest("POST", "/api/end-user/auth/login",
        strings.NewReader(`{
            "orgName":"test-org",
            "identifier":"end_user",
            "identifierType":"USERNAME",
            "password":"password"
        }`))
    w := httptest.NewRecorder()
    
    handler.EndUserLogin(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
    // Verify response contains orgName
}
```

---

## Mock EndUser Repository Needed

To test LoginEndUser, we need a mock that implements `EndUserRepository`:

```go
type mockEndUserRepository struct {
    users map[string]*enduser.EndUser
}

func (m *mockEndUserRepository) GetByPhone(ctx context.Context, orgName, phone string) (*enduser.EndUser, error) {
    for _, u := range m.users {
        if u.Phone.String() == phone && u.OrgName == orgName {
            return u, nil
        }
    }
    return nil, nil
}

func (m *mockEndUserRepository) GetByUsername(ctx context.Context, orgName, username string) (*enduser.EndUser, error) {
    for _, u := range m.users {
        if u.Username == username && u.OrgName == orgName {
            return u, nil
        }
    }
    return nil, nil
}

func (m *mockEndUserRepository) GetByUsernameGlobal(ctx context.Context, username string) (*enduser.EndUser, error) {
    for _, u := range m.users {
        if u.Username == username {
            return u, nil
        }
    }
    return nil, nil
}

func (m *mockEndUserRepository) GetByID(ctx context.Context, orgName, userID string) (*enduser.EndUser, error) {
    u, ok := m.users[userID]
    if !ok || (orgName != "" && u.OrgName != orgName) {
        return nil, nil
    }
    return u, nil
}
```

---

## Summary

| Test # | Name | Type | Current | Status | Effort |
|--------|------|------|---------|--------|--------|
| 1 | Login by USERNAME | Unit | Missing | HIGH | Low |
| 2 | Login deprecated Phone | Unit | Missing | MEDIUM | Low |
| 3 | LoginEndUser by Phone+Org | Unit | Missing | HIGH | Medium |
| 4 | LoginEndUser by Username+Org | Unit | Missing | HIGH | Medium |
| 5 | LoginEndUser by Username global | Unit | Missing | MEDIUM | Medium |
| 6 | LoginEndUser disabled account | Unit | Missing | MEDIUM | Low |
| 7 | Handler login nil type defaults | Integration | Missing | MEDIUM | Medium |
| 8 | Handler enduser login | Integration | Missing | HIGH | Medium |

**Total New Tests Required:** 8
**Estimated Implementation Time:** 3-4 hours

---

## Testing Before Merge Checklist

- [ ] Test 1: Login by USERNAME — PASSING
- [ ] Test 2: Login deprecated Phone — PASSING
- [ ] Test 3: LoginEndUser by Phone+Org — PASSING
- [ ] Test 4: LoginEndUser by Username+Org — PASSING
- [ ] Test 5: LoginEndUser by Username global — PASSING
- [ ] Test 6: LoginEndUser disabled account — PASSING
- [ ] Test 7: Handler login nil type defaults — PASSING
- [ ] Test 8: Handler enduser login — PASSING
- [ ] All existing tests still pass — ✓ VERIFIED
- [ ] Code coverage remains >= baseline — ✓ VERIFIED

Only after ALL checks are complete should merge proceed.

