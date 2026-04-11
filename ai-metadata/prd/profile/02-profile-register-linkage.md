# 注册联动创建（Register → User + Profile）

## 页面清单定位

| # | 页面名称 | 页面类型 | 路由路径 | 主体实体 | 展示字段 | 操作 |
|---|----------|----------|----------|----------|----------|------|
| 2 | 注册联动创建 | 流程页 | /register | User, Profile | phone, userName, password, nickname(default) | 提交注册、自动建档 |

---

## ASCII 布局

```text
┌──────────────────────────────────────────────────────────────┐
│ ▸ Auth / Register                                            │
│                                                              │
│ ┌──────────────────────┐                                     │
│ │ Register Form        │                                     │
│ │ · phone              │ ← User.phone                        │
│ │ · userName           │ ← User.user_name                    │
│ │ · password           │ ← User.password_hash                │
│ │ [创建账号]           │                                     │
│ └──────────┬───────────┘                                     │
│            │ submit                                           │
│            v                                                  │
│   ┌─────────────────────────────┐                            │
│   │ Tx: create user + profile   │                            │
│   │ · insert user               │                            │
│   │ · insert profile(default)   │                            │
│   └─────────────────────────────┘                            │
└──────────────────────────────────────────────────────────────┘
```

---

## 页面信息

- **页面名称**: 注册联动创建
- **页面类型**: 流程页
- **路由路径**: /register
- **所属领域**: 认证 / 用户资料

## 数据依赖

| 领域实体 | 使用方式 | 关键字段 |
|----------|----------|----------|
| User | 注册主记录创建 | phone, user_name, password_hash |
| Profile | 注册后联动创建 | user_id, nickname, avatar_url, bio |

## 区块说明

### 1. 注册输入区块
- **用途**: 收集最小账号信息。
- **包含字段**: phone、userName、password（注册不需要 email）。
- **交互**: 点 `[创建账号]` 后触发注册流程。

### 2. 联动创建事务区块
- **用途**: 保证 user/profile 成对创建。
- **包含字段**: user.id、profile.user_id、默认资料字段。
- **交互**: user 创建成功后立即创建 profile，任一步失败则回滚。

### 3. 默认值初始化区块
- **用途**: 给新用户提供可用的初始资料。
- **包含字段**: nickname（注册时直接给默认值）、avatar_url（默认 mock 头像 URL）、bio（空）。
- **交互**: 创建后可在资料编辑页覆盖。

## 用户操作

| 操作 | 触发方式 | 影响 |
|------|----------|------|
| 提交注册 | 点击 `[创建账号]` | 创建 user 与 profile |
| 注册重试 | 接口失败后再次提交 | 幂等校验避免重复账号 |

## 错误状态

| 错误场景 | 提示内容 | 来源 |
|----------|----------|------|
| phone 已存在 | `phone already registered` | User 唯一约束 |
| user 成功但 profile 失败 | `registration failed, rollback` | 事务执行 |
| 默认值初始化异常 | `profile init failed` | 业务规则/配置 |
| mock 头像资源不可用 | `default avatar mock unavailable` | mock 资源配置 |

## 设计决策

- 注册采用单事务编排（推荐），避免“有 user 无 profile”的脏状态。
- profile 默认值初始化放在后端，保证不同客户端一致。

## 不做什么

- 不在注册流程中采集完整资料（昵称/头像/简介可后补）。
- 不在本次支持第三方登录联动创建。

## 待确认

- [ ] 是否强制“注册事务内创建 profile”（强一致）。
- [x] nickname 默认生成规则：固定前缀+随机后缀（例如 `user_A1B2C3`）。
