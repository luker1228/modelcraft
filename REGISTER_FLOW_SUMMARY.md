# Register Flow: Key Findings & Merge Decisions

## Quick Reference: Two Completely Different Architectures

### Register (domain/user path) — **Global Phone-based Identity**
- **Scope:** Global ModelCraft user
- **Identity key:** Phone number (11-digit China) + userName
- **Validation:** 
  - Phone: Required, validated format
  - UserName: 3-32 chars, must start with `[a-zA-Z_-]`, reserved words blocked (admin, root, system, etc.)
- **Password:** Stored as string hash (from passwordHasher.Hash interface)
- **Organization:** Created separately via createOrgService → personal org
- **Admin status:** Not tracked in User entity (stored in user_orgs.is_admin)
- **Account state:** No disable/enable (always active)
- **Profile:** Created automatically with default avatar
- **Database:** Single INSERT to `users` table only
- **Fields:** ID, ExternalID (NULL), Name, Phone (VO), PasswordHash (string), CreatedAt, UpdatedAt

### EndUser (domain/enduser path) — **Tenant-scoped Identity**
- **Scope:** Within a specific organization (multi-tenant)
- **Identity key:** Username + OrgName
- **Validation:**
  - Username: 3-64 chars, alphanumeric+_-, NO reserved word check, any first char allowed
  - OrgName: Required (tenant scope key)
- **Password:** Stored as HashedPassword VO (wraps Hash + Algorithm fields)
- **Organization:** Passed as constructor parameter, stored in user_orgs
- **Admin status:** Tracked as IsAdmin boolean field
- **Account state:** Tracked as IsForbidden boolean field (Enable/Disable methods)
- **Profile:** Not created by EndUser entity
- **Database:** Double INSERT (users + user_orgs tables)
- **Phone field:** Hardcoded to empty string — NOT preserved
- **Fields:** ID, OrgName, Username, Password (VO), IsForbidden, IsAdmin, CreatedAt, UpdatedAt

---

## Critical Merge Decisions Required

### 1. **Phone Field** ⚠️ MAJOR CONFLICT
- **User:** Phone is required, 11-digit validated, stored per-user
- **EndUser:** Phone is EMPTY STRING (not stored)
- **Decision:**
  - [ ] Keep phone required for global register flow
  - [ ] Make phone optional for tenant users
  - [ ] Create two different user types (global vs tenant)

### 2. **Username Validation** ⚠️ MAJOR CONFLICT
- **User:** 3-32 chars, first char MUST be `[a-zA-Z_-]`, reserved words blocked (13 words)
- **EndUser:** 3-64 chars, any char allowed, no reserved check
- **Decision:** Adopt **stricter rules** (User's validation) to prevent conflicts
  - Username must be 3-32 chars
  - Must start with `[a-zA-Z_-]`
  - Must not be a reserved word

### 3. **Password Storage** ⚠️ ARCHITECTURAL CHOICE
- **User:** String hash only (`PasswordHash: string`)
- **EndUser:** Value object (`Password: HashedPassword` with Hash + Algorithm)
- **Decision:** Adopt **EndUser's HashedPassword VO pattern**
  - More type-safe
  - Explicit algorithm tracking (future-proof for multiple algorithms)
  - Both already use bcrypt(cost=12)

### 4. **Organization Context** ⚠️ SCHEMA CHANGE
- **User:** Global user, personal org created separately
- **EndUser:** org_name stored in user_orgs table, implicit in context
- **Decision:** Add `org_name` field to User entity
  - Makes organization relationship explicit
  - Aligns with EndUser model
  - Requires users table schema change

### 5. **Account State Management** ⚠️ NEW CAPABILITY REQUIRED
- **User:** No disable/enable (always active)
- **EndUser:** IsForbidden field + Enable/Disable methods
- **Decision:** Add `is_forbidden` field to merged User
  - Replaces user_orgs.deleted_at model
  - Enables account suspension without soft-delete

### 6. **Admin Status** ⚠️ SCHEMA CHANGE
- **User:** Not tracked (stored in user_orgs.is_admin)
- **EndUser:** IsAdmin field in EndUser struct
- **Decision:** Add `is_admin` field to User entity
  - Makes it explicit in User aggregate
  - Simplifies queries and business logic

### 7. **Profile Creation** ✓ NO CONFLICT
- **User:** Creates profile with default avatar
- **EndUser:** No profile creation
- **Decision:** Keep profile creation in merged Register flow

---

## Database Schema Changes Required

**Current users table (via Register):**
```
id, external_id, name, phone, password_hash, display_name, created_at, updated_at
```

**After merge (merged User entity to adopt EndUser context):**
```
id, external_id, name, phone, password_hash, display_name,
org_name,         (NEW - tenant scope key)
is_forbidden,     (NEW - account disable flag)
is_admin,         (NEW - org admin flag)
created_at, updated_at
```

**Then user_orgs can be simplified or deprecated** (org_name + is_admin now in users table).

---

## Field Mapping for Merged User

**Proposed merged `domain/user.User` struct:**

```go
type User struct {
    ID            string          // UUID (global or per-org)
    ExternalID    string          // OAuth provider ID (optional, NULL for phone+password)
    OrgName       string          // Tenant scope key (NEW from EndUser)
    Name          string          // Username (unified field name)
    Phone         PhoneNumber     // 11-digit China phone (optional: may be empty for tenant users)
    Password      HashedPassword  // bcrypt hash VO (NEW from EndUser, replaces PasswordHash)
    IsForbidden   bool            // Account disabled flag (NEW from EndUser)
    IsAdmin       bool            // Org admin flag (NEW from EndUser)
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

**Removed:**
- ExternalID → nullable for OAuth path (keep for backward compat)
- Separate PasswordHash string → replaced by Password VO

**Added:**
- OrgName
- IsForbidden
- IsAdmin
- Password (VO)

---

## Cleanup Opportunities

### ✓ CAN REMOVE from domain/user:
1. `registerNameAdjectives` & `registerNameNouns` (marked unused)
2. `generateRegisterDisplayName()` (unused)
3. Consider deprecating `NewOAuthUser()` if OAuth path uses EndUser

### ✓ MUST ADOPT from domain/enduser:
1. `HashedPassword` value object → use in merged User
2. `Enable()` / `Disable()` methods for account state
3. `IsActive()` method
4. `VerifyPassword()` method
5. `ValidatePasswordStrength()` validation

### ✓ VALIDATIONS TO UNIFY:
1. Merge `ValidateUserName` (user) + `ValidateUsername` (enduser) → adopt stricter (user's)
2. Adopt user's reserved words list in merged validation
3. Keep user's username length 3-32, not enduser's 3-64

---

## Register Flow Post-Merge

**Same as today, but with unified user structure:**

1. HTTP: HandleRegister(`{phone, password, userName, organizationName}`)
2. Validate: userName (3-32, format, reserved), phone (11-digit China), password (strength)
3. Hash: password → bcrypt via HashedPassword.NewHashedPasswordFromPlain()
4. Create User: `domain/user.NewUser(id, userName, phone, hashedPassword, orgName)`
5. Create Profile: `domain/profile.NewInitialProfile(...)`
6. Persist (tx):
   - userRepo.Create(user) → INSERT users table (with new org_name, is_forbidden, is_admin fields)
   - profileRepo.CreateInitialProfile(profile)
7. Create Org (or link existing): createOrgService.Execute()
8. Return: RegisterResult with userId, orgName, profile

**Result:** Single User domain entity serving both:
- Global phone-based registration (Register)
- Tenant-scoped user management (EndUser use cases)

---

## Implementation Priority

1. **Phase 1 (Schema):** Add org_name, is_forbidden, is_admin to users table
2. **Phase 2 (Domain):** Merge User + EndUser entities, adopt EndUser's fields + methods
3. **Phase 3 (Validation):** Unify username validation rules (adopt user's stricter rules)
4. **Phase 4 (Repository):** Consolidate SqlUserRepository + SqlEndUserOrgRepository
5. **Phase 5 (Cleanup):** Remove unused code, deprecate old APIs

