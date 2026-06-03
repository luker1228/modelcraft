# Register Flow: Field-by-Field Merge Checklist

## A. Database INSERT Equivalence ✗ NOT EQUIVALENT

| Item | SqlUserRepository.Create | SqlEndUserOrgRepository.Save | Action Required |
|------|--------------------------|------------------------------|-----------------|
| **Phone field** | Populated (11-digit) | Empty string `''` | ❌ CONFLICT - decide phone handling |
| **Name/Username** | `u.Name` | `u.Username` | ✓ SAME semantic, rename to Name |
| **PasswordHash** | `u.PasswordHash` (string) | `u.Password.Hash` (from VO) | ✓ Adopt VO pattern |
| **ExternalID** | NULL for phone+pwd | Not set (NULL) | ✓ SAME (keep for OAuth) |
| **DisplayName** | NULL | Not set | ✓ SAME (always NULL) |
| **Soft-delete** | DB default | Explicit `deleted_at=0` | 🔄 Standardize explicitly |
| **Timestamps** | DB default | Explicit NOW(3) | 🔄 Standardize explicitly |
| **user_orgs** | NOT created | Created (is_admin=0, active) | ❌ CONFLICT - decide ownership |

**Decision:** Merge creates NEW unified INSERT logic that handles:
- Phone (required for Register, optional for tenant users?)
- org_name + is_admin directly in users table (not user_orgs)
- Explicit soft-delete and timestamps

---

## B. Entity Validation: NewUser vs NewEndUser

| Validation | domain/user.NewUser | domain/enduser.NewEndUser | Action Required |
|------------|---------------------|--------------------------|-----------------|
| **Phone required** | ✓ YES (IsZero check) | ✗ NO | ❌ CONFLICT - phone optional? |
| **Phone format** | ✓ 11-digit China | N/A | ✓ Keep for phone users |
| **ID required** | ✓ Simple check | ✓ Simple check | ✓ SAME |
| **OrgName required** | ✗ NO | ✓ YES | ✓ ADD to merged User |
| **Username format** | ✓ 3-32, start `[a-zA-Z_-]`, reserved check | ✓ 3-64, any start, NO reserved | ❌ CONFLICT - adopt stricter (user's) |
| **PasswordHash required** | ✓ YES (empty check) | N/A (VO handles) | ✓ SAME |
| **Password strength** | ✓ Done in Register line 127 | ✓ In HashedPassword constructor | ✓ SAME logic, move to VO |

**Decision:**
- [ ] Merged NewUser requires: id, orgName, name (3-32, format, reserved), phone (optional), hashedPassword
- [ ] Username validation: 3-32 chars, must start with `[a-zA-Z_-]`, reserved words blocked

---

## C. Username Validation: ValidateUserName vs ValidateUsername

| Rule | ValidateUserName | ValidateUsername | Merged Rule |
|------|------------------|------------------|-------------|
| **Min length** | 3 | 3 | 3 |
| **Max length** | 32 | 64 | **32** (stricter) |
| **First char** | MUST `[a-zA-Z_-]` | Any `[a-zA-Z0-9_-]` | **MUST `[a-zA-Z_-]`** (stricter) |
| **Other chars** | `[a-zA-Z0-9_-]` | `[a-zA-Z0-9_-]` | ✓ SAME |
| **Reserved words** | YES (13 words) | NO | **YES, (13 words)** (stricter) |

**Reserved words to block:** admin, administrator, root, system, sys, modelcraft, support, help, api, www, null, undefined, anonymous

**Decision:**
- [ ] Create merged ValidateUsername that enforces ALL of the stricter user rules
- [ ] Deprecate both old validators
- [ ] This validation will be stricter than EndUser but better security

---

## D. Phone Number Validation: NewPhoneNumber

| Item | Requirement | Action |
|------|-------------|--------|
| **Regex pattern** | `^1[3-9]\d{9}$` | ✓ Keep (11-digit mainland China) |
| **Value object** | PhoneNumber struct | ✓ Keep |
| **String() method** | Raw phone | ✓ Keep |
| **Masked() method** | "138****1234" format | ✓ Keep |
| **IsZero() method** | Empty string check | ✓ Keep |
| **Usage in merged User** | Optional for phone users, empty for tenant-only users? | ❌ DECIDE phone semantics |

**Decision:**
- [ ] Keep PhoneNumber VO as-is
- [ ] Phone field in merged User should be: `Phone PhoneNumber` (may be zero)
- [ ] Register flow requires phone, tenant-only users can have empty phone

---

## E. Password Hashing: PasswordHasher vs HashedPassword

| Item | PasswordHasher (user) | HashedPassword (enduser) | Merged Approach |
|------|----------------------|------------------------|--------------------|
| **Type pattern** | Interface (injectable) | Value object | **Adopt VO pattern** |
| **Hash generation** | bcrypt(cost=12) | bcrypt(cost=12) | ✓ SAME algorithm |
| **Algorithm field** | Not stored | Stored explicitly | **Add to User Password field** |
| **Validation** | Before Hash() call | In NewHashedPasswordFromPlain | ✓ SAME (password strength check) |
| **Verify method** | Interface method | VO method | **Use VO.Verify()** |

**Decision:**
- [ ] Remove PasswordHasher interface from domain/user
- [ ] Import and use enduser.HashedPassword VO in merged User
- [ ] NewUser constructor takes `HashedPassword` (not string)
- [ ] Remove domain/auth.PasswordHasher injection from Register flow (use HashedPassword directly)
- [ ] Register calls `HashedPassword.NewHashedPasswordFromPlain(password)` before NewUser

---

## F. Profile Creation: Handled ✓

| Item | Register | EndUser | Merged |
|------|----------|---------|--------|
| **Creates profile** | ✓ YES | ✗ NO | ✓ Keep YES (same as Register) |
| **Default avatar** | "mock://avatar/default-1.png" | N/A | ✓ Keep |
| **Profile persisted** | ✓ YES | N/A | ✓ Keep |
| **Returned to client** | ✓ YES | N/A | ✓ Keep |

**Decision:** No changes needed — keep Register's profile creation logic.

---

## G. Field Mapping: User vs EndUser

### Current domain/user.User
```go
type User struct {
    ID           string           // ✓ Keep
    ExternalID   string           // ✓ Keep (OAuth)
    Name         string           // ✓ Keep (rename from Name to align?)
    Phone        PhoneNumber      // ✓ Keep (make optional)
    PasswordHash string           // ❌ REMOVE (replace with Password VO)
    CreatedAt    time.Time        // ✓ Keep
    UpdatedAt    time.Time        // ✓ Keep
}
```

### Current domain/enduser.EndUser
```go
type EndUser struct {
    ID          string         // ✓ Merge (same as User.ID)
    OrgName     string         // ✓ ADD to merged User
    Username    string         // = User.Name (unify)
    Password    HashedPassword // ✓ ADD to merged User (replaces PasswordHash)
    IsForbidden bool           // ✓ ADD to merged User
    IsAdmin     bool           // ✓ ADD to merged User
    CreatedAt   time.Time      // ✓ Keep (same)
    UpdatedAt   time.Time      // ✓ Keep (same)
}
```

### Proposed Merged User
```go
type User struct {
    ID            string          // UUID (required)
    ExternalID    string          // OAuth provider ID (optional)
    OrgName       string          // Tenant scope key (required for EndUser, optional for legacy?)
    Name          string          // Username (3-32, start [a-zA-Z_-], reserved words blocked)
    Phone         PhoneNumber     // 11-digit China phone (optional, may be empty for tenant users)
    Password      HashedPassword  // Bcrypt hash VO (Hash + Algorithm fields)
    IsForbidden   bool            // Account disabled flag (default false)
    IsAdmin       bool            // Org admin flag (default false)
    CreatedAt     time.Time       // Creation timestamp
    UpdatedAt     time.Time       // Last update timestamp
}
```

**Checklist:**
- [ ] Keep: ID, ExternalID, Name, Phone, CreatedAt, UpdatedAt
- [ ] Replace: PasswordHash (string) → Password (HashedPassword VO)
- [ ] Add: OrgName, IsForbidden, IsAdmin
- [ ] Update: Make Phone optional (can be empty PhoneNumber)
- [ ] Update: Move from PasswordHash string to Password VO structure

---

## H. Methods/Constructors to Merge

### From User domain
| Method | Keep? | Move to merged? | Replace? |
|--------|-------|-----------------|----------|
| `NewUser()` | Keep signature but UPDATE params | YES | Add orgName, use HashedPassword |
| `NewOAuthUser()` | Keep for OAuth path | YES | No changes |
| `ValidateUserName()` | NO - deprecate | YES | Merged into ValidateUsername (stricter) |
| `generateRegisterDisplayName()` | NO - unused | DELETE | Remove completely |
| `registerNameAdjectives` | NO - unused | DELETE | Remove completely |
| `registerNameNouns` | NO - unused | DELETE | Remove completely |

### From EndUser domain
| Method | Keep? | Move to merged? | Import? |
|--------|-------|-----------------|---------|
| `NewEndUser()` | NO - merge with NewUser | REPLACE | Merged into NewUser |
| `ValidateUsername()` | NO - deprecate | REPLACE | Merged (stricter version) |
| `NewHashedPasswordFromPlain()` | YES - keep | IMPORT | Import HashedPassword VO |
| `NewHashedPasswordFromHash()` | YES - keep | IMPORT | Import HashedPassword VO |
| `Enable()` | YES - add to merged | YES (new) | Add to merged User |
| `Disable()` | YES - add to merged | YES (new) | Add to merged User |
| `IsActive()` | YES - add to merged | YES (new) | Add to merged User |
| `VerifyPassword()` | YES - add to merged | YES (new) | Add to merged User |

---

## Database Schema Migration Checklist

**Current users table (sqlc generated from schema):**
```sql
CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    external_id VARCHAR(255) NULL,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NULL,
    created_at TIMESTAMP(3),
    updated_at TIMESTAMP(3),
    deleted_at INT DEFAULT 0,
    delete_token INT DEFAULT 0
);
```

**After merge (add 3 new columns):**
```sql
ALTER TABLE users ADD COLUMN org_name VARCHAR(255) NULL;  -- Tenant scope key
ALTER TABLE users ADD COLUMN is_forbidden BOOLEAN DEFAULT FALSE;  -- Account disable flag
ALTER TABLE users ADD COLUMN is_admin BOOLEAN DEFAULT FALSE;  -- Org admin flag
```

**Then update sqlc schema definition and regenerate:**
- [ ] Update users table schema in migration file
- [ ] Run `sqlc generate` to create new CreateUserParams
- [ ] Update SqlUserRepository.Create() to map new fields

---

## Implementation Checklist

- [ ] **Phase 1: Domain Model Merge**
  - [ ] Create new User struct with merged fields (see Section G)
  - [ ] Import HashedPassword VO from domain/enduser
  - [ ] Create new NewUser constructor signature: `NewUser(id, orgName, name, phone, password, orgName)`
  - [ ] Add Enable(), Disable(), IsActive(), VerifyPassword() methods
  - [ ] Create merged ValidateUsername() (adopt stricter user rules)
  - [ ] Keep NewOAuthUser() for OAuth path
  - [ ] Mark as deprecated: ValidateUserName(), registerNameAdjectives, registerNameNouns, generateRegisterDisplayName
  - [ ] Delete unused: generateRegisterDisplayName, registerNameAdjectives, registerNameNouns

- [ ] **Phase 2: Validation Unification**
  - [ ] Create merged ValidateUsername: 3-32 chars, start `[a-zA-Z_-]`, reserved words blocked
  - [ ] Update Register validation (line 116-128 in token_service.go) to use merged validator
  - [ ] Verify backward compat: no existing usernames become invalid

- [ ] **Phase 3: Repository Update**
  - [ ] Update SqlUserRepository.Create() to handle new fields (org_name, is_forbidden, is_admin)
  - [ ] Verify sqlc params match merged User struct
  - [ ] Test: Insert User with all fields populated
  - [ ] Consider: Can SqlEndUserOrgRepository.Save now call SqlUserRepository.Create()?

- [ ] **Phase 4: Register Flow Update**
  - [ ] Update token_service.Register() to:
    - [ ] Use HashedPassword.NewHashedPasswordFromPlain() instead of passwordHasher.Hash()
    - [ ] Pass orgName to NewUser()
    - [ ] Pass HashedPassword VO instead of string hash
  - [ ] Remove passwordHasher dependency from TokenService? (or keep for backward compat)

- [ ] **Phase 5: Cleanup**
  - [ ] Delete unused code from domain/user
  - [ ] Deprecate domain/enduser.NewEndUser in favor of merged User
  - [ ] Update all callers of ValidateUserName to use merged ValidateUsername
  - [ ] Verify no breaking changes for OAuth flow (NewOAuthUser still works)

---

## Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|-----------|
| Phone becomes optional, breaking phone-based lookups | HIGH | Thorough testing of phone lookup queries, add NOT NULL constraint option if needed |
| Username validation stricter, existing EndUsers can't login | MEDIUM | Pre-migration validation to check if any existing EndUser violates stricter rules |
| Password VO requires migration of existing hashes | MEDIUM | HashedPassword.NewHashedPasswordFromHash() handles this, no re-hashing needed |
| org_name required but Register doesn't pass it consistently | HIGH | Add validation in NewUser to ensure org_name provided |
| user_orgs table becomes redundant | LOW | Keep for transition period, deprecate gradually |

---

## Validation Checklist (Pre-Merge Testing)

- [ ] No existing users violate merged username rules (3-32, starts with [a-zA-Z_-], no reserved)
- [ ] All existing phone numbers are either valid 11-digit format OR empty (for migration)
- [ ] Register flow still creates profile correctly
- [ ] Login still works for both phone and username lookups
- [ ] EndUser operations still work after merge
- [ ] OAuth flow still works with NewOAuthUser
- [ ] Account enable/disable works with new IsForbidden field
- [ ] Admin status correctly reflected in IsAdmin field

