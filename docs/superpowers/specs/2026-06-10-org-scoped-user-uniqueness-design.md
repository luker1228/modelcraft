# Org-Scoped User Uniqueness & Admin Registration Redesign

**Date**: 2026-06-10  
**Status**: Approved  
**Scope**: Backend schema + Admin auth flow + Frontend login page

---

## Background

ModelCraft has two types of users sharing the same `users` table:

- **Admin（开发者）**: Self-registers, manages Org/Project/Model via the console
- **EndUser（开发者的用户）**: Created by Admin, consumes runtime APIs via SDK — not a console user

Currently `users.phone` and `users.name` are **globally unique**. This prevents multiple Orgs from having users with the same phone/username, which is incorrect for a multi-tenant system.

The fix: move uniqueness to org-scope for `users`, and use `organizations.phone` as the global anchor for Admin registration and login.

---

## Design Decisions

### Core Principle
- `organizations.phone` — globally unique, represents "one phone = one Org"
- `users.(org_name, phone)` — unique within an Org
- `users.(org_name, name)` — unique within an Org
- EndUser has no login flow on the frontend; SDK handles all token acquisition

### Login Strategy
| Role | Login Method | Lookup Logic |
|------|-------------|-------------|
| Admin | Phone + Password | `org.phone` (global unique) → locate org → `users WHERE org_name=? AND phone=?` |
| EndUser | (frontend entry removed) | Backend interface retained for SDK/internal use |

---

## Part 1: Schema Changes

### `organizations` table

Add phone field and global unique index:

```sql
`phone` VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'Org 注册手机号，全局唯一',

UNIQUE INDEX `uk_org_phone` (`phone`, `delete_token`)
```

### `users` table

Add `org_name` field, replace global unique indexes with org-scoped ones:

```sql
-- New field
`org_name` VARCHAR(36) NOT NULL DEFAULT '' COMMENT '所属 Org，创建时绑定'

-- Remove
-- UNIQUE INDEX `uk_phone` (`phone`, `delete_token`)
-- UNIQUE INDEX `uk_user_name` (`name`, `delete_token`)

-- Add
UNIQUE INDEX `uk_org_user_phone` (`org_name`, `phone`, `delete_token`)  -- org 内手机唯一
UNIQUE INDEX `uk_org_user_name`  (`org_name`, `name`,  `delete_token`)  -- org 内用户名唯一

-- Update existing index (name is now org-scoped, include org_name)
-- Remove: INDEX `idx_users_live_name` (`deleted_at`, `name`)
-- Add:
INDEX `idx_users_live_name` (`deleted_at`, `org_name`, `name`)

-- Add FK constraint
CONSTRAINT `fk_users_org` FOREIGN KEY (`org_name`) REFERENCES `organizations`(`name`) ON UPDATE CASCADE
```

`user_orgs` table: **no change**. Retains its role as Admin ↔ Org membership record.

---

## Part 2: Admin Registration Flow

### Request

```
POST /api/tenant/auth/register
Body: {
  org_display_name  string  // Org 展示名称（新增）
  phone             string  // 手机号（移到 org 层做全局唯一校验）
  password          string
  username          string
}
```

### Steps

```
1. Check organizations.phone global uniqueness
   → Duplicate: return error "该手机号已注册组织"

2. Generate org_name (random slug)

3. Create Organization
   → name = generated slug
   → display_name = org_display_name
   → phone = request.phone

4. Create User
   → org_name = created org's name
   → phone = request.phone  (org-scoped unique)
   → name = request.username (org-scoped unique)

5. Create user_orgs
   → is_admin = 1
```

### Key Change from Current
- Registration now **requires `org_display_name`** as input
- Phone uniqueness check moves from `users` table to `organizations` table
- User creation must bind `org_name` at creation time

---

## Part 3: Admin Login Flow

### Request

```
POST /api/tenant/auth/login
Body: {
  phone     string  // 唯一登录标识符（不再支持用户名登录）
  password  string
}
```

### Steps

```
1. SELECT org_name FROM organizations WHERE phone = ?
   → Not found: return error "手机号未注册"

2. SELECT * FROM users WHERE org_name = ? AND phone = ?
   → Not found: return error "用户不存在"

3. bcrypt.Verify(password, user.password_hash)
   → Mismatch: return error "密码错误"

4. Issue JWT
   → PlatformClaims { user_id, org_name, is_admin=true, ... }
   → RefreshToken → httpOnly Cookie
```

### Key Change from Current
- Login identifier: `identifier` (phone or username) → `phone` only
- Remove `GetByUsernameGlobal` lookup path
- `GetByPhoneGlobal` replaced by two-step: org lookup → user lookup

---

## Part 4: Frontend Changes

### Scope: Login page only

**Remove**: EndUser login entry point from the homepage/login page

**Keep**: Admin login form (phone + password)

Backend EndUser login endpoints are **retained** — not deleted. They remain available for SDK usage and internal testing. The frontend simply does not expose them.

---

## Part 5: PAT + API Docs → Admin Console

### PAT Management
- Already under `graphql/org/` (no path change)
- Confirm entry point is inside Admin console only, not on public pages

### API Documentation
- Move into Admin console (visible after login)
- Content: how to initialize SDK, how to obtain EndUser tokens, how to call runtime APIs

---

## Impact Analysis

### Backend

| Layer | Change |
|-------|--------|
| `db/schema/mysql/05_organizations.sql` | Add `phone` field + `uk_org_phone` index |
| `db/schema/mysql/06_users.sql` | Add `org_name` field, replace unique indexes |
| `internal/domain/organization/` | Add `Phone` field to Organization entity |
| `internal/domain/user/` | Add `OrgName` field to User entity |
| `internal/app/auth/token_service.go` | Login: two-step phone lookup (org → user) |
| `internal/app/user/` (registration) | Add org creation step, bind org_name to user |
| `internal/interfaces/http/handlers/auth/handler.go` | Update register/login request structs |
| Repository layer | Update `GetByPhone`, remove `GetByUsernameGlobal` |
| DB migration | Atlas migration for both tables |

### Frontend

| Area | Change |
|------|--------|
| Login / Home page | Remove EndUser login entry |
| PAT management | Confirm in Admin console (likely already there) |
| API docs page | Move behind auth gate |

---

## Out of Scope

- TypeScript SDK design (separate spec)
- EndUser token acquisition via SDK (separate spec)
- EndUser backend login endpoints (retained as-is, no changes)
- RBAC / permission changes
