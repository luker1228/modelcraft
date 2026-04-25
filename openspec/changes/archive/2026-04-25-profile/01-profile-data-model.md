# 数据模型拆分（User / Profile）

## 页面清单定位

| # | 页面名称 | 页面类型 | 路由路径 | 主体实体 | 展示字段 | 操作 |
|---|----------|----------|----------|----------|----------|------|
| 1 | 数据模型拆分 | 设计页 | N/A（数据层） | User, Profile | user.id, user.phone, user.user_name, profile.nickname, profile.avatar_url, profile.bio | 建表、建约束 |

---

## ASCII 布局

```text
┌──────────────────────────────────────────────────────────────┐
│ // ── data schema: user + profile (1:1) ──                  │
│                                                              │
│ ┌──────────────────────┐   1 : 1   ┌──────────────────────┐ │
│ │ user                 │───────────│ profile              │ │
│ │ · id (PK)            │           │ · id (PK)            │ │
│ │ · phone (UNIQUE)     │           │ · user_id (FK,UNIQUE)│ │
│ │ · user_name (UNIQUE) │           │ · nickname           │ │
│ │ · password_hash      │           │ · avatar_url         │ │
│ │ · status             │           │ · bio                │ │
│ │ · created_at         │           │ · created_at         │ │
│ └──────────────────────┘           └──────────────────────┘ │
└──────────────────────────────────────────────────────────────┘
```

---

## 页面信息

- **页面名称**: 数据模型拆分（User / Profile）
- **页面类型**: 设计页
- **路由路径**: N/A（后端数据层能力）
- **所属领域**: 认证 / 用户资料

## 数据依赖

| 领域实体 | 使用方式 | 关键字段 |
|----------|----------|----------|
| User | 账号主体（认证、授权主键） | id, phone, user_name, password_hash, status |
| Profile | 资料扩展（一对一） | user_id, nickname, avatar_url, bio |

## 区块说明

### 1. User 主体字段
- **用途**: 承载登录与身份识别必要字段。
- **包含字段**: id、phone、user_name、password_hash、status、created_at。
- **交互**: 被注册、登录、鉴权流程直接依赖。

### 2. Profile 扩展字段
- **用途**: 承载可演进的用户资料信息。
- **包含字段**: user_id、nickname、avatar_url、bio、created_at。
- **交互**: 被资料编辑、资料展示、联合查询流程读写。

### 3. 约束与索引
- **用途**: 保证一对一关系与查询性能。
- **包含字段**: profile.user_id UNIQUE + FK(user.id)。
- **交互**: 防止重复 profile，保证级联一致性策略可执行。

## 用户操作

| 操作 | 触发方式 | 影响 |
|------|----------|------|
| 创建 user | 注册成功 | 产生账号主体记录 |
| 创建 profile | 注册联动创建 | 建立资料记录 |
| 读取 profile | 详情查询 | 返回扩展信息 |

## 错误状态

| 错误场景 | 提示内容 | 来源 |
|----------|----------|------|
| profile.user_id 重复 | `profile already exists` | 唯一约束 |
| profile.user_id 无对应 user | `invalid user reference` | 外键约束 |
| 非法字段长度 | `nickname too long` | 数据校验 |

## 设计决策

- User 与 Profile 强制一对一，避免“一号多档案”带来的读写歧义。
- 账号字段与资料字段解耦，后续资料扩展不影响认证主链路。

## 不做什么

- 不在本次引入多 profile 模式。
- 不在本次引入社交关系字段（关注/粉丝）。

## 待确认

- [ ] nickname 是否允许为空；为空时前端展示策略。
- [ ] avatar_url 默认值（空字符串 / 默认头像 URL）。
