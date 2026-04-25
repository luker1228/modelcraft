# 终端用户认证（End-User Auth）

## 是什么问题

ModelCraft 目前只有**开发者**角色可以登录，使用自研 JWT 认证。

随着平台演进，开发者需要为其构建的应用提供**终端用户**访问能力：终端用户不参与项目构建，只访问开发者搭建好的数据管理页面，操作实际数据记录。

两种角色的职责完全不同，因此需要**完全独立的登录入口和会话体系**。

## 角色定义

| 角色 | 定义 | 登录入口 | 登录后落地页 |
|------|------|----------|-------------|
| **开发者**（Developer） | 构建和管理 ModelCraft 项目的人 | `/login`（现有） | 工作区 / 项目管理页 |
| **终端用户**（End User） | 使用开发者所构建应用的用户 | `/user/login`（新增） | 数据管理页（Data Manager） |

> 每个 Project 下的终端用户账号彼此隔离，即同一个人在不同 Project 中是不同账号。

## 技术栈与架构

**认证层：BFF（Next.js）作为轻量网关代理，Go Backend 使用 `goth` 实现终端用户认证，不依赖 Casdoor 等外部服务。**

```
终端用户浏览器
    │
    ▼
Next.js BFF（轻量代理网关）
    ├── middleware.ts — 路由保护，按 cookie key 区分角色
    ├── /api/bff/end-user/auth/* — 代理到 Go Backend
    └── /org/.../end-users — 用户账号管理，代理到 Go Backend

Go Backend（goth）
    ├── 验证用户名/密码（mc_private.users + accounts）
    ├── 签发 / 轮换 refresh token（存 mc_private）
    └── 返回 { userId, refreshToken } 给 BFF

BFF
    ├── 收到 Go 响应 → 自签 end-user access token（jose, 1h）
    └── 写 HttpOnly Cookie: end_user_refresh_token（7d）
```

### 数据库分库

| 数据库 | 用途 |
|--------|------|
| `mc_meta` | ModelCraft 平台元数据（原 `modelcraft`，单独任务改名） |
| `private_{projectSlug}` | 终端用户身份数据，每个 Project 一个独立库（`users` + `accounts` 表） |

**连接路由**：`private_{projectSlug}` 库与 Project 关联的 `DatabaseCluster` 在同一 MySQL 实例上，Go Backend 通过读取 `mc_meta` 的 Project → Cluster 链动态路由，**不依赖独立配置**。

> 终端用户本质是 Project 的产物，其数据库连接来源于 Project 配置，而非平台配置。

### Cookie 隔离

| Cookie Key | 角色 | 说明 |
|------------|------|------|
| `refresh_token` | 开发者 | 现有，不变 |
| `end_user_refresh_token` | 终端用户 | 新增 |

## 核心功能

### 终端用户登录入口

- 独立 URL：`/org/{orgName}/project/{projectSlug}/user/login`
- 前端路由与开发者 `/login` 完全隔离
- 支持凭证：用户名 + 密码
- 登录成功后：跳转至 `/org/{orgName}/project/{projectSlug}/data`

### 用户账号管理（开发者操作）

- 入口：`/org/{orgName}/project/{projectSlug}/end-users`
- 开发者可创建 / 禁用 / 删除终端用户账号
- 账号归属于特定 Project，不跨 Project 共享

### 会话隔离

- 终端用户 access token payload 含 `role: "end_user"`、`orgName`、`projectSlug`
- middleware 按 cookie key 区分路由保护规则
- 终端用户 token 无法访问项目配置接口

## 子页文档

| 文件 | 说明 |
|------|------|
| [01-user-login.md](./01-user-login.md) | 用户登录页详细设计 |
| [02-user-account-management.md](./02-user-account-management.md) | 开发者管理终端用户账号 |
| [03-backend-design.md](./03-backend-design.md) | Go Backend 认证方案（goth + mc_private） |

## 不做什么

- 不修改现有开发者登录流程
- 不做终端用户自助注册（v1 由开发者创建账号）
- 不做短信 / 邮箱验证码登录
- 不做第三方 OAuth 登录
- 不做忘记密码 / 重置密码
- 不做用户间权限分级
- 不跨 Project 共享终端用户账号
- 不在本需求中改名 `mc_meta`（单独任务）

## 验收标准

1. 访问 `/user/login` 与开发者登录页完全独立
2. 终端用户登录后落地数据管理页，无法访问项目配置页
3. 开发者 cookie 无法通过终端用户路由保护
4. 终端用户 cookie 无法通过开发者路由保护
5. 开发者可创建 / 禁用 / 删除终端用户账号
