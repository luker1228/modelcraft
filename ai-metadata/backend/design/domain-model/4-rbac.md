# 4. 用户鉴权（RBAC）

> 代码位置：
> - `internal/domain/membership/`
> - `internal/domain/role/`
> - `internal/domain/permission/`

## 概述

ModelCraft 使用基于角色的访问控制（RBAC）。用户通过 Membership 加入 Org，在 Org 内分配角色，角色绑定权限集合。

## 核心实体

### Membership — 用户与 Org 的关联

```
internal/domain/membership/membership.go

Membership
├── ID         string
├── UserID     string
├── OrgID      string
├── OrgName    string           // 冗余字段，避免 JOIN
├── Status     MembershipStatus // active | suspended | invited
├── InvitedBy  string           // 邀请人 UserID（可空）
├── InvitedAt  *time.Time
└── JoinedAt   *time.Time
```

### Role — 角色定义

```
internal/domain/role/role.go

Role
├── ID          string
├── OrgName     string
├── Name        string       // 角色标识
├── DisplayName string
├── IsSystem    bool         // 系统角色不可删除
└── Permissions []string     // 权限列表
```

### Permission — 权限值对象

```
internal/domain/permission/permission.go

格式："{resource}:{action}"
示例：
  model:read
  model:write
  project:*      // 通配符：project 下所有操作
  *              // 超级权限
```

## 关系图

```
User ──── Membership ────▶ Organization
                │
                ▼
             UserRole ────▶ Role ────▶ Permission[]
```

## 成员生命周期

```
NewInvitation()          NewMembership()
      │                        │
      ▼                        ▼
  invited ──AcceptInvitation()──▶ active
      │                              │
      └──────── Suspend() ──────▶ suspended
```

## 相关文件

- `internal/domain/membership/membership.go`
- `internal/domain/role/role.go`
- `internal/domain/permission/permission.go`
- `internal/domain/permission/role.go` — Role 与 Permission 关联
- `internal/domain/permission/user_role.go` — UserRole 实体
