# Register Flow Analysis: Complete Documentation

This directory contains a comprehensive analysis of the Register flow in ModelCraft and the user/enduser merge planning.

## Document Structure

### 1. **REGISTER_FLOW_SUMMARY.md** (START HERE)
**Quick reference for decision-makers** (7.5 KB, 191 lines)

Read this first for:
- High-level comparison of Register (global user) vs EndUser (tenant user) architectures
- 7 critical merge decisions with decision points
- Proposed merged User struct
- Implementation priority phases
- Field mapping overview

**Use when:** You need to quickly understand the scope and make architectural decisions.

---

### 2. **REGISTER_FLOW_ANALYSIS.md** (DETAILED REFERENCE)
**Complete technical analysis** (21.9 KB, 590 lines)

Answers all 7 key questions in detail:

- **A. Database INSERT Equivalence** — Are SqlUserRepository.Create and SqlEndUserOrgRepository.Save equivalent?
  - Answer: **NO** — critical differences in phone field, user_orgs table creation, soft-delete handling
  
- **B. Validation & Entity Construction** — What does NewUser validate that NewEndUser doesn't?
  - Answer: Phone required + format, stricter username rules, reserved words check
  
- **C. Username Validation** — ValidateUserName vs ValidateUsername comparison
  - Answer: Major differences (3-32 vs 3-64, first char rules, reserved words)
  
- **D. Phone Number Validation** — Is there an equivalent in domain/enduser?
  - Answer: **NO** — EndUser has no phone field or validation
  
- **E. Password Hashing** — PasswordHasher interface vs HashedPassword VO
  - Answer: Different patterns (interface vs VO), same crypto (bcrypt 12)
  
- **F. Profile Creation** — Register creates profile, EndUser doesn't?
  - Answer: **YES** — Register creates default profile with avatar
  
- **G. Field Comparison** — What fields differ?
  - Answer: User has 7 fields, EndUser has 8, significant non-overlapping fields

**Use when:** You need to understand the detailed reasoning behind each decision.

---

### 3. **REGISTER_FLOW_CHECKLIST.md** (IMPLEMENTATION GUIDE)
**Field-by-field merge checklist** (12.9 KB, 282 lines)

Structured checklists for:

- **Section A-G:** Field-by-field action items (Keep? Remove? Replace? Add?)
- **Section H:** Methods/constructors to merge (what to keep, deprecate, delete, import)
- **Database Schema Migration:** SQL changes required for users table
- **Implementation Checklist:** 5 phases of work with detailed tasks
- **Risk Assessment:** Severity and mitigation for each risk
- **Validation Checklist:** Pre-merge testing requirements

**Use when:** You're implementing the merge and need step-by-step guidance.

---

## Quick Navigation by Question

| Question | Answer | Location |
|----------|--------|----------|
| **What's the big difference?** | Register = global, EndUser = tenant-scoped | SUMMARY, section intro |
| **Are the database INSERTs the same?** | NO - phone, user_orgs, timestamps differ | ANALYSIS.A |
| **What validation conflicts exist?** | Phone required vs optional, username 3-32 vs 3-64 | ANALYSIS.B & C |
| **Should I keep/remove/add fields?** | See field mapping tables | CHECKLIST.G |
| **What methods should I merge?** | See method tables | CHECKLIST.H |
| **What's the implementation plan?** | 5 phases from schema → domain → validation → repo → cleanup | SUMMARY & CHECKLIST |
| **What are the risks?** | Phone lookups, validation strictness, password VO migration | CHECKLIST "Risk Assessment" |
| **What must I test?** | Username rules, phone lookups, register flow, login, org creation, profile | CHECKLIST "Validation Checklist" |

---

## Key Findings Summary

### Two Completely Different Architectures

**Register (domain/user)**
- Global user identity based on phone + username
- 7 fields: ID, ExternalID, Name, Phone, PasswordHash, CreatedAt, UpdatedAt
- Phone required and validated (11-digit China format)
- Username: 3-32 chars, starts with [a-zA-Z_-], reserved words blocked
- Creates profile automatically
- Single INSERT to users table (no user_orgs)
- PasswordHash stored as string

**EndUser (domain/enduser)**
- Tenant-scoped identity (org-specific)
- 8 fields: ID, OrgName, Username, Password, IsForbidden, IsAdmin, CreatedAt, UpdatedAt
- Phone hardcoded to empty string (NOT preserved)
- Username: 3-64 chars, any first char, no reserved check
- No profile creation
- Double INSERT (users + user_orgs)
- Password stored as HashedPassword VO

### Critical Merge Decisions

1. **Phone Field** ❌ CONFLICT
   - Register requires it, EndUser discards it
   - Decision: Keep phone required for Register, optional for tenant users

2. **Username Validation** ❌ CONFLICT
   - User: 3-32 strict, EndUser: 3-64 permissive
   - Decision: Adopt stricter rules (user's)

3. **Password Storage** ⚠️ ARCHITECTURAL
   - User: string, EndUser: VO
   - Decision: Adopt EndUser's HashedPassword VO

4. **Org Context** ⚠️ SCHEMA CHANGE
   - User: global, EndUser: requires org_name
   - Decision: Add org_name field to merged User

5. **Account State** ⚠️ NEW CAPABILITY
   - User: no disable, EndUser: IsForbidden
   - Decision: Add IsForbidden to merged User

6. **Admin Status** ⚠️ SCHEMA CHANGE
   - User: in user_orgs.is_admin, EndUser: IsAdmin field
   - Decision: Add IsAdmin to merged User

7. **Profile** ✓ NO CONFLICT
   - Register creates it, EndUser doesn't
   - Decision: Keep profile creation

---

## Implementation Roadmap

### Phase 1: Schema (users table)
- Add 3 columns: org_name, is_forbidden, is_admin
- Regenerate sqlc types

### Phase 2: Domain Model
- Merge User struct with EndUser fields
- Import HashedPassword VO from enduser
- Add Enable/Disable/IsActive/VerifyPassword methods
- Create new NewUser constructor

### Phase 3: Validation
- Merge ValidateUserName + ValidateUsername → adopt stricter rules
- Update Register validation
- Verify backward compat

### Phase 4: Repository
- Update SqlUserRepository.Create for new fields
- Consider consolidating with SqlEndUserOrgRepository.Save

### Phase 5: Cleanup
- Delete unused code (registerNameAdjectives, etc.)
- Deprecate old validators
- Update all callers

---

## What Can Be Removed from User Domain

1. `registerNameAdjectives` array (marked unused)
2. `registerNameNouns` array (marked unused)
3. `generateRegisterDisplayName()` function (unused)
4. `ValidateUserName()` function (superseded by merged validator)
5. Consider: `NewOAuthUser()` if OAuth path adopts EndUser model

## What Must Be Adopted from EndUser

1. `HashedPassword` value object (type safety)
2. `Enable()` / `Disable()` methods
3. `IsActive()` method
4. `VerifyPassword()` method
5. Stricter `ValidatePasswordStrength()` validation

---

## Files Referenced

**Domain Models:**
- `modelcraft-backend/internal/domain/user/user.go` (7 fields)
- `modelcraft-backend/internal/domain/enduser/end_user.go` (8 fields)
- `modelcraft-backend/internal/domain/user/phone_number.go` (validation)
- `modelcraft-backend/internal/domain/enduser/hashed_password.go` (VO)

**Repositories:**
- `modelcraft-backend/internal/infrastructure/repository/sql_org_repository.go` (Register path)
- `modelcraft-backend/internal/infrastructure/repository/sql_enduser_repository.go` (EndUser path)

**Service:**
- `modelcraft-backend/internal/app/auth/token_service.go` (Register logic, lines 110-264)

**HTTP:**
- `modelcraft-backend/internal/interfaces/http/handlers/auth/handler.go` (HandleRegister)

---

## How to Use These Documents

1. **Planning meeting:** Read SUMMARY + show field mapping table (CHECKLIST.G)
2. **Implementation kickoff:** Reference CHECKLIST for step-by-step tasks
3. **Code review:** Use ANALYSIS for validation of decisions
4. **Testing:** Use CHECKLIST "Validation Checklist" for test cases
5. **Documentation:** Link to ANALYSIS for rationale on merged design

---

## Questions to Answer Before Starting

- [ ] Should phone be required for all users, or optional for tenant-only users?
- [ ] Are there existing EndUser accounts that violate the stricter (merged) username rules?
- [ ] Can we migrate existing password hashes to HashedPassword VO without re-hashing?
- [ ] Should user_orgs table be deprecated after merge, or kept for transition?
- [ ] Do OAuth flows need phone numbers, or is that Register-only?
- [ ] What's the backward compatibility requirement for old PasswordHasher interface?

