# Register Flow Analysis: User vs EndUser Creation Paths

**Date:** 2026-06-02  
**Goal:** Compare Register (domain/user path) vs CreateEndUser (domain/enduser path) to understand what can be cleaned up post user/enduser merge.

---

## A. Database INSERT Equivalence: SqlUserRepository.Create vs SqlEndUserOrgRepository.Save

### SqlUserRepository.Create (Register path)
**File:** `modelcraft-backend/internal/infrastructure/repository/sql_org_repository.go:173`

**SQL Operation:**
```go
params := dbgen.CreateUserParams{
    ID:           u.ID,                    // UUID
    ExternalID:   StrToNullStr(u.ExternalID),  // NULL for phone+password register
    Name:         u.Name,                  // userName from user input
    Phone:        u.Phone.String(),        // 11-digit China phone
    PasswordHash: u.PasswordHash,          // bcrypt hash from passwordHasher.Hash()
    DisplayName:  sql.NullString{},        // Always NULL
}
return r.q.CreateUser(ctx, params)  // sqlc generated, inserts to users table
```

**Table: `users`**
- `id` (PK): UUID
- `external_id`: NULL for phone+password
- `name`: user-provided userName
- `phone`: 11-digit validated phone number
- `password_hash`: bcrypt hash
- `display_name`: NULL

---

### SqlEndUserOrgRepository.Save (EndUser path)
**File:** `modelcraft-backend/internal/infrastructure/repository/sql_enduser_repository.go:44`

**SQL Operations:**

#### Insert 1: users table
```sql
INSERT INTO users (id, name, phone, password_hash, deleted_at, delete_token, created_at, updated_at)
VALUES (?, ?, '', ?, 0, 0, NOW(3), NOW(3))
```
With params: `user.ID, user.Username, user.Password.Hash`

**Differences from Register:**
- `phone`: **EMPTY STRING** (hardcoded `''`) — NOT the actual phone
- `name` field stores: `user.Username` (same as Register's `Name`)
- Explicit soft-delete fields: `deleted_at=0, delete_token=0`
- Explicit timestamps: `created_at=NOW(3), updated_at=NOW(3)`

#### Insert 2: user_orgs table (NOT created by Register)
```sql
INSERT INTO user_orgs 
  (id, user_id, org_name, is_admin, status, deleted_at, delete_token, created_at, updated_at)
VALUES (?, ?, ?, 0, 'active', 0, 0, NOW(3), NOW(3))
```
With generated `userOrgID` and `orgName`.

**New table involved:** `user_orgs` — **NOT touched by Register**.

---

### **ANSWER TO A: NO, the two INSERTs are NOT equivalent**

| Aspect | Register (SqlUserRepository.Create) | EndUser (SqlEndUserOrgRepository.Save) |
|--------|-------------------------------------|---------------------------------------|
| **Phone field** | Populated with 11-digit validated phone | **EMPTY STRING** (`''`) |
| **ExternalID** | NULL (for phone+password) | Not set (NULL by default) |
| **DisplayName** | NULL (hardcoded) | Not set (NULL by default) |
| **Soft-delete fields** | Not set (likely NULL/0 by default) | Explicitly set (`deleted_at=0, delete_token=0`) |
| **Timestamps** | Not set (likely DB default NOW()) | Explicitly set (`NOW(3)`) |
| **user_orgs table** | **NOT created** | **Created** with `is_admin=0, status='active'` |
| **OrgName context** | Not stored (personal org created separately) | Stored in `user_orgs.org_name` |

**Critical Issue:** EndUser.Save hardcodes `phone=''`, losing the phone number that was used during registration. This suggests EndUser is designed for **multi-tenant end-users within an org**, not global phone-based registration.

---

## B. Validation & Entity Construction: domain/user.NewUser vs domain/enduser.NewEndUser

### domain/user.NewUser
**File:** `modelcraft-backend/internal/domain/user/user.go:79`

```go
func NewUser(id, userName string, phone PhoneNumber, passwordHash string) (*User, error) {
    // 1. Phone must not be zero (validated PhoneNumber required)
    if phone.IsZero() {
        return nil, fmt.Errorf("phone number is required")
    }
    // 2. Password hash must not be empty
    if passwordHash == "" {
        return nil, fmt.Errorf("password hash is required")
    }
    // 3. userName must pass ValidateUserName (3-32 chars, format, reserved words)
    if err := ValidateUserName(userName); err != nil {
        return nil, err
    }
    // 4. Set timestamps
    now := time.Now()
    user := &User{
        ID:           id,
        Name:         userName,              // ✓ user-provided
        Phone:        phone,                 // ✓ validated PhoneNumber
        PasswordHash: passwordHash,          // ✓ provided
        CreatedAt:    now,
        UpdatedAt:    now,
    }
    // 5. Entity validation (only checks ID not empty)
    if err := user.Validate(); err != nil {
        return nil, err
    }
    return user, nil
}
```

**Validations:**
1. ✓ Phone required (must be non-zero PhoneNumber VO)
2. ✓ PasswordHash required
3. ✓ userName format & reserved words check (3-32 chars, `^[a-zA-Z_-][a-zA-Z0-9_-]{2,31}$`, NO admin/root/etc)

**Field defaults:**
- `ExternalID`: NOT set (remains `""`)
- `CreatedAt/UpdatedAt`: Set to `time.Now()`

---

### domain/enduser.NewEndUser
**File:** `modelcraft-backend/internal/domain/enduser/end_user.go:21`

```go
func NewEndUser(id, orgName, username string, hashedPwd HashedPassword) (*EndUser, error) {
    // 1. ID required
    if id == "" {
        return nil, fmt.Errorf("user ID is required")
    }
    // 2. OrgName required (tenant scope)
    if orgName == "" {
        return nil, fmt.Errorf("org name is required")
    }
    // 3. Username must pass ValidateUsername (3-64 chars, alphanumeric + _-)
    if err := ValidateUsername(username); err != nil {
        return nil, err
    }
    // 4. HashedPassword is a value object (already validated at construction)
    
    now := time.Now()
    return &EndUser{
        ID:          id,
        OrgName:     orgName,               // ✓ required tenant key
        Username:    username,              // ✓ validated
        Password:    hashedPwd,             // ✓ provided VO
        IsForbidden: false,                 // ✓ default active
        IsAdmin:     false,                 // ✓ default non-admin (can be set after)
        CreatedAt:   now,
        UpdatedAt:   now,
    }, nil
}
```

**Validations:**
1. ✓ ID required (same as user)
2. ✓ OrgName required (NEW — tenant scope)
3. ✓ Username format check (3-64 chars, `^[a-zA-Z0-9_-]{3,64}$`, NO format checks for start char)

**Field defaults:**
- `IsForbidden`: Set to `false`
- `IsAdmin`: Set to `false`

---

### **ANSWER TO B: Significant validation differences**

| Aspect | domain/user.NewUser | domain/enduser.NewEndUser |
|--------|---------------------|--------------------------|
| **Phone validation** | ✓ Required, PhoneNumber VO with 11-digit China format check | ✗ NO phone field at all |
| **UserName format** | ✓ 3-32 chars, start with `[a-zA-Z_-]`, reserved words check | ✓ 3-64 chars, alphanumeric+_-, NO reserved check |
| **OrgName** | ✗ NOT present (implicit global user) | ✓ Required (explicit tenant scope) |
| **Password validation** | ✗ None (only checks not empty) | ✓ Via HashedPassword.NewHashedPasswordFromPlain() which calls ValidatePasswordStrength (8+ chars, letter, digit) |
| **ExternalID** | ✗ Not set (OAuth path uses NewOAuthUser) | ✗ Not present |
| **Active/Forbidden state** | ✗ No disable/enable methods | ✓ IsForbidden field + Enable/Disable methods |

**Key insight:** user.NewUser is for **global registration with phone+password**. enduser.NewEndUser is for **tenant-scoped users within an org**, with no phone requirement but with tenant context.

---

## C. Username Validation: ValidateUserName vs ValidateUsername

### domain/user.ValidateUserName
**File:** `modelcraft-backend/internal/domain/user/user.go:38`

```go
func ValidateUserName(name string) error {
    // 1. Length: 3-32 characters
    if len(name) < 3 || len(name) > 32 {
        return fmt.Errorf("userName must be 3-32 characters, got %d", len(name))
    }
    
    // 2. Pattern: ^[a-zA-Z_-][a-zA-Z0-9_-]{2,31}$
    //    → Starts with letter/underscore/hyphen, then alphanumeric/underscore/hyphen
    pattern := regexp.MustCompile(`^[a-zA-Z_-][a-zA-Z0-9_-]{2,31}$`)
    if !pattern.MatchString(name) {
        return fmt.Errorf("userName must start with letter/underscore/hyphen and contain only alphanumeric/underscore/hyphen")
    }
    
    // 3. Reserved words (case-insensitive): admin, administrator, root, system, sys, 
    //    modelcraft, support, help, api, www, null, undefined, anonymous
    lowerName := strings.ToLower(name)
    reservedWords := []string{
        "admin", "administrator", "root", "system", "sys",
        "modelcraft", "support", "help", "api", "www",
        "null", "undefined", "anonymous",
    }
    for _, reserved := range reservedWords {
        if lowerName == reserved {
            return fmt.Errorf("userName '%s' is reserved", name)
        }
    }
    
    return nil
}
```

**Validations:**
1. ✓ Length: 3-32 (inclusive)
2. ✓ First char: MUST be `[a-zA-Z_-]`
3. ✓ Remaining 2-31 chars: `[a-zA-Z0-9_-]`
4. ✓ Reserved words check (13 words)

---

### domain/enduser.ValidateUsername
**File:** `modelcraft-backend/internal/domain/enduser/hashed_password.go:61`

```go
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,64}$`)

func ValidateUsername(username string) error {
    if !usernameRegex.MatchString(username) {
        return fmt.Errorf("username must be 3-64 characters and contain only letters, numbers, underscores, or hyphens")
    }
    return nil
}
```

**Validations:**
1. ✓ Length: 3-64 (inclusive)
2. ✓ Pattern: `[a-zA-Z0-9_-]` (any first char is allowed)
3. ✗ NO reserved words check
4. ✗ NO requirement that first char be letter/underscore/hyphen

---

### **ANSWER TO C: Major differences**

| Aspect | ValidateUserName | ValidateUsername |
|--------|------------------|------------------|
| **Length** | 3-32 | 3-64 |
| **First char** | MUST be `[a-zA-Z_-]` | Can be any `[a-zA-Z0-9_-]` |
| **Pattern** | `^[a-zA-Z_-][a-zA-Z0-9_-]{2,31}$` | `^[a-zA-Z0-9_-]{3,64}$` |
| **Reserved words** | 13 words blocked (admin, root, system, etc.) | **NO check** |

**Impact:** A username like "0admin" would fail in Register but pass for EndUser. Reserved words like "admin" would fail in Register but pass for EndUser. Merge must adopt stricter validation.

---

## D. Phone Number Validation: domain/user.NewPhoneNumber

### domain/user.NewPhoneNumber
**File:** `modelcraft-backend/internal/domain/user/phone_number.go:16`

```go
var phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)  // 11 digits, mainland China

type PhoneNumber struct {
    value string
}

func NewPhoneNumber(value string) (PhoneNumber, error) {
    // Validates 11-digit mainland China phone numbers
    if !phoneRegex.MatchString(value) {
        return PhoneNumber{}, fmt.Errorf("invalid phone number format: must be 11-digit mainland China number")
    }
    return PhoneNumber{value: value}, nil
}

// Methods on PhoneNumber:
// - String() → raw phone
// - Masked() → "138****1234" format
// - IsZero() → true if empty
```

**Validation rule:**
- Pattern: `^1[3-9]\d{9}$`
  - Starts with `1`
  - Second digit: `[3-9]`
  - Followed by 9 more digits
  - Total: 11 digits
  - **Mainland China phone numbers only**

---

### **ANSWER TO D: NO equivalent in domain/enduser**

**domain/enduser has NO phone field, phone validation, or PhoneNumber value object.**

**Implication:** EndUser was designed for **tenant-scoped users without phone-based identity**. The merge must decide: should EndUser support phone-based login/registration?

---

## E. Password Hashing: passwordHasher interface in Register vs HashedPassword in EndUser

### Register path: domain/auth.PasswordHasher interface
**File:** `modelcraft-backend/internal/app/auth/token_service.go:65,150`

```go
// In TokenService:
passwordHasher domainauth.PasswordHasher  // injected dependency

// In Register():
hashedPassword, err := s.passwordHasher.Hash(ctx, cmd.Password)
// ...
u, err := domainUser.NewUser(id, cmd.UserName, phone, hashedPassword)
```

**interface:** `domain/auth/password.go:40`
```go
type PasswordHasher interface {
    Hash(ctx context.Context, password string) (string, error)
    Verify(ctx context.Context, password, hash string) error
}
```

**Implementation:** `internal/infrastructure/auth/bcrypt_hasher.go:21`
```go
func (h *BcryptPasswordHasher) Hash(_ context.Context, password string) (string, error) {
    // Calls bcrypt.GenerateFromPassword with cost=12
    hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    return string(hash), err
}
```

---

### EndUser path: domain/enduser.HashedPassword value object
**File:** `modelcraft-backend/internal/domain/enduser/hashed_password.go:16`

```go
const bcryptCost = 12

type HashedPassword struct {
    Hash      string  // bcrypt hash string
    Algorithm string  // always "bcrypt"
}

// Constructor 1: from plaintext
func NewHashedPasswordFromPlain(plain string) (HashedPassword, error) {
    if err := ValidatePasswordStrength(plain); err != nil {
        return HashedPassword{}, err
    }
    hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
    if err != nil {
        return HashedPassword{}, fmt.Errorf("failed to hash password: %w", err)
    }
    return HashedPassword{
        Hash:      string(hash),
        Algorithm: "bcrypt",
    }, nil
}

// Constructor 2: from existing hash
func NewHashedPasswordFromHash(hash string) HashedPassword {
    return HashedPassword{
        Hash:      hash,
        Algorithm: "bcrypt",
    }
}

// Methods
func (p HashedPassword) Verify(plain string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(p.Hash), []byte(plain))
    return err == nil
}
```

**Password strength validation:**
```go
func ValidatePasswordStrength(plain string) error {
    return domainauth.ValidatePasswordStrength(plain)
    // → 8+ chars, 1+ letter, 1+ digit
}
```

---

### **ANSWER TO E: Different patterns, same crypto**

| Aspect | PasswordHasher (Register) | HashedPassword (EndUser) |
|--------|--------------------------|-------------------------|
| **Type** | Interface (injectable) | Value object (self-contained) |
| **Hash call** | `passwordHasher.Hash(ctx, plain)` → string | `NewHashedPasswordFromPlain(plain)` → HashedPassword struct |
| **Crypto algorithm** | bcrypt (cost=12) | bcrypt (cost=12) |
| **Password validation** | No validation in interface (done in Register line 127 before hashing) | ValidatePasswordStrength() called in NewHashedPasswordFromPlain |
| **Verify** | `passwordHasher.Verify(ctx, plain, hash)` method | `hashedPwd.Verify(plain)` method |
| **Storage** | Register stores `user.PasswordHash` (string) | EndUser stores `user.Password` (HashedPassword struct with Hash + Algorithm) |

**Key difference:** 
- **Register**: Uses injected PasswordHasher interface → string hash → stored in user.PasswordHash
- **EndUser**: Constructs HashedPassword value object → encapsulates Hash + Algorithm fields

**After merge:** The enduser.HashedPassword pattern is more explicit (stores algorithm), but both use bcrypt(12). Could consolidate by:
1. Making PasswordHasher a method on HashedPassword, or
2. Wrapping HashedPassword in a PasswordHasher interface implementation

---

## F. Profile Record Creation

### Register path: Creates profile
**File:** `modelcraft-backend/internal/app/auth/token_service.go:167-176`

```go
// Create initial profile
profileID, err := bizutils.GenerateUUIDV7()
if err != nil {
    return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate profile id")
}
defaultAvatarURL := "mock://avatar/default-1.png"
initialProfile, err := domainProfile.NewInitialProfile(
    profileID, u.ID, "", &defaultAvatarURL, nil)
if err != nil {
    return nil, bizerrors.WrapError(err, bizerrors.SystemError, "create profile entity")
}
```

**Persisted:** Line 312 in `persistUserAndProfile()`
```go
if err := profileRepo.CreateInitialProfile(ctx, p); err != nil {
    return bizerrors.ConvertRepositoryError(ctx, err)
}
```

**Result returned to client:** Line 180-186
```go
result := &RegisterResult{
    UserID: u.ID,
    Profile: RegisterProfileSnapshot{
        ID:        initialProfile.ID,
        UserID:    initialProfile.UserID,
        Nickname:  initialProfile.Nickname,
        AvatarURL: initialProfile.AvatarURL,
        Bio:       initialProfile.Bio,
    },
}
```

**Summary:** ✓ Register creates `profile` record with default avatar + empty nickname/bio.

---

### EndUser path: NO profile creation in CreateEndUser
**File:** `modelcraft-backend/internal/domain/enduser/end_user.go:21`

**No profile entity. EndUser is purely the user identity within org scope.**

---

### **ANSWER TO F: Register creates profile, EndUser does NOT**

| Aspect | Register | EndUser |
|--------|----------|---------|
| **Profile creation** | ✓ YES — domain/profile.NewInitialProfile called | ✗ NO — EndUser has no profile field |
| **Default avatar** | ✓ "mock://avatar/default-1.png" | N/A |
| **Profile persisted** | ✓ YES — via profileRepo.CreateInitialProfile | N/A |
| **Returned to client** | ✓ YES — profile object in RegisterResult | N/A |

**After merge:** Need to ensure merged Register still creates profile for new users.

---

## G. Field Comparison: domain/user.User vs domain/enduser.EndUser

### domain/user.User fields
**File:** `modelcraft-backend/internal/domain/user/user.go:12`

```go
type User struct {
    ID           string        // ModelCraft internal UUID
    ExternalID   string        // external auth provider user ID (from JWT.sub)
    Name         string        // user name (for phone+password register)
    Phone        PhoneNumber   // value object with validation
    PasswordHash string        // bcrypt hash (only for phone+password users)
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

**7 fields.**

---

### domain/enduser.EndUser fields
**File:** `modelcraft-backend/internal/domain/enduser/end_user.go:10`

```go
type EndUser struct {
    ID          string         // UUID, primary key
    OrgName     string         // org scope key (tenant)
    Username    string         // 3-64 chars, unique within org
    Password    HashedPassword // bcrypt hashed (value object)
    IsForbidden bool           // account disabled flag
    IsAdmin     bool           // org admin flag
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**8 fields.**

---

### **ANSWER TO G: Significant field differences**

| Field | domain/user.User | domain/enduser.EndUser | Notes |
|-------|------------------|----------------------|-------|
| **ID** | ✓ | ✓ | Both UUID, same purpose |
| **Name / Username** | ✓ (Name) | ✓ (Username) | Same semantic, different field names |
| **Phone** | ✓ PhoneNumber VO | ✗ | Register uses phone; EndUser does not |
| **ExternalID** | ✓ | ✗ | For OAuth users in Register; not in EndUser |
| **PasswordHash** | ✓ (string) | ✗ | Register stores string hash directly |
| **Password** | ✗ | ✓ (HashedPassword VO) | EndUser wraps hash + algorithm in VO |
| **OrgName** | ✗ | ✓ | EndUser is tenant-scoped; User is global |
| **IsForbidden** | ✗ | ✓ | EndUser supports account disable; User does not |
| **IsAdmin** | ✗ | ✓ | EndUser tracks org admin status; User does not |
| **CreatedAt / UpdatedAt** | ✓ | ✓ | Both track timestamps |

**Total fields:**
- **User:** 7 fields (ID, ExternalID, Name, Phone, PasswordHash, CreatedAt, UpdatedAt)
- **EndUser:** 8 fields (ID, OrgName, Username, Password, IsForbidden, IsAdmin, CreatedAt, UpdatedAt)

**Fields unique to User:** ExternalID, Phone, PasswordHash (string)  
**Fields unique to EndUser:** OrgName, IsForbidden, IsAdmin, Password (VO)

---

## Summary: Cleanup Opportunities Post User/EndUser Merge

### Can be REMOVED from User domain:
1. **ExternalID field** — appears to be OAuth-only, not used in Register
2. **NewOAuthUser constructor** — NewUser is used for phone+password; OAuth path is deprecated
3. **registerNameAdjectives / registerNameNouns** — unused (marked `//nolint:unused`)
4. **generateRegisterDisplayName** — unused

### Must be ADOPTED from EndUser:
1. **OrgName field** — to support multi-tenant users
2. **IsForbidden field** — to support account disable/suspension
3. **IsAdmin field** — to track org admin status (currently in user_orgs table)
4. **HashedPassword value object** — more explicit than string hash

### CONFLICTS to resolve:
1. **Phone field:** User has required 11-digit phone; EndUser has none. **Merge decision:** Should merged user support phone-based identity globally or only per-org?
2. **Username validation:** User is 3-32 with reserved words; EndUser is 3-64 with no reserved check. **Merge decision:** Adopt stricter (User's) rules to avoid conflicts.
3. **Password hash storage:** User stores string; EndUser stores VO. **Merge decision:** Adopt EndUser's VO pattern for type safety.
4. **Profile creation:** Register creates profile; EndUser doesn't. **Merge decision:** Keep profile creation in merged Register flow.

### Database schema implications:
- `users` table will need `org_name` column (currently implicit via `user_orgs`)
- `users` table will need `is_forbidden` column (currently implicit via `user_orgs.deleted_at`)
- `users` table phone field will be per-user, not empty string per EndUser approach
- Password hash storage remains bcrypt, cost=12 in both paths

---

## Register Flow Summary (Current State)

1. **HTTP:** HandleRegister receives `{phone, password, userName, organizationName}`
2. **Validate:** userName format, phone format, password strength, username/phone uniqueness
3. **Hash:** password → bcrypt hash via passwordHasher.Hash()
4. **Create User entity:** domainUser.NewUser(id, userName, phone, hash)
5. **Create Profile entity:** domainProfile.NewInitialProfile(id, userId, empty nickname, default avatar, nil bio)
6. **Persist (transactional):**
   - userRepo.Create(user) → INSERT users table
   - profileRepo.CreateInitialProfile(profile) → INSERT profile table
7. **Create Org:** createOrgService.Execute() → INSERT organizations + user_orgs
8. **Return:** RegisterResult with userId, orgName, profile snapshot

**Key Point:** Register creates both a global User and a personal Organization, then an EndUser within that org implicitly (via user_orgs). Post-merge, User must directly support org_name + is_admin + is_forbidden fields.

