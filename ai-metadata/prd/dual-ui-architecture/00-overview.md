---
版本: v1.3
状态: Spec（待实现）
日期: 2026-05-05
---

# ModelCraft 前端双 UI 架构 PRD

## 1. 背景与问题陈述

ModelCraft 历史上试图用一套 UI 同时服务两类受众：管理模型和配置权限的开发者，以及通过工作区操作数据的终端用户。

这一做法带来的核心问题：

| 维度 | 表现 |
|------|------|
| **路由冲突** | 租户端路由 `/org/[orgName]/...` 与用户端路由 `/end-user/[orgName]/...` 概念上清晰，但实现时边界不断模糊 |
| **认证混用** | 租户端使用 OAuth/Developer Token，用户端使用 EndUser Token；同一 UI 框架需要同时处理两种 Token 生命周期 |
| **导航结构撕裂** | 租户端需要多层级配置导航，用户端需要精简的数据操作面板；一套导航无法同时满足 |
| **条件渲染失控** | 大量 `if (isEndUser)` / `if (isAdmin)` 分支散落各处，组件复杂度持续上升 |
| **功能边界不清** | 设计时功能（模型编辑、字段管理）和运行时功能（数据 CRUD）共存于同一组件体系 |

---

## 2. 设计目标

### 用户目标

| 受众 | 目标 |
|------|------|
| **租户端用户**（开发者、Org 管理员）| 高效完成模型设计、权限配置、EndUser 账号管理等设计时工作；导航清晰，功能密集 |
| **用户端用户**（EndUser / 终端用户）| 用最少步骤完成数据操作（增删改查）；界面精简，无无关功能干扰 |

### 产品目标

- 两套 UI 路由前缀完全隔离，入口独立
- 共用一套 JWT Token 体系，通过 `aud` claim 区分身份：`aud=tenant`（租户端）/ `aud=end-user`（用户端）
- 导航结构各自独立设计，不共享侧边栏或顶部导航
- **共享层**：Design System（shadcn/ui + Tailwind CSS 变量）+ 通用业务组件（如数据表格、弹窗、表单组件等可复用部分）

---

## 3. 受众与角色定义

### 租户端用户（Tenant UI）

| 角色 | 描述 | 登录方式 |
|------|------|---------|
| Org Owner | 创建 Org，管理成员，拥有全部权限 | `/tenant/login` → OAuth / Developer Auth |
| Org Admin | 管理成员、EndUser 账号、数据库集群 | `/tenant/login` → OAuth / Developer Auth |
| Developer | 设计模型、配置字段、管理枚举、设置 RBAC | `/tenant/login` → OAuth / Developer Auth |

### 用户端用户（App UI）

| 角色 | 描述 | 登录方式 |
|------|------|---------|
| EndUser | 使用 Org 搭建的内部工具，对数据执行 CRUD 操作 | `/end-user/[orgName]/login` → EndUser 两阶段认证 |

> **职责分离**：Org 管人（EndUser 账号生命周期由租户端管理），Project 管权（EndUser 能访问哪些 Project、拥有哪些 Role 由租户端配置）。

---

## 4. 核心架构决策

### 路由边界

| UI | 路由前缀 | 登录入口 |
|----|---------|---------|
| **租户端（Tenant UI）** | `/org/[orgName]/...` | `/tenant/login` |
| **用户端（App UI）** | `/end-user/[orgName]/...` | `/end-user/[orgName]/login` |

两个前缀在 Next.js App Router 中是完全独立的路由树，互不嵌套。

### Token 体系

两套 UI **共用一套 JWT 体系**，Token 格式统一，通过 `aud`（audience）claim 区分身份类型：

| `aud` 值 | 颁发场景 | 可访问路由 |
|----------|---------|-----------|
| `tenant` | 租户端 OAuth 登录成功后颁发 | `/org/[orgName]/...`、`/tenant/login` |
| `end-user` | EndUser username/password 登录后颁发 | `/end-user/[orgName]/...` |

**不存在 Project Token**：EndUser 登录后只颁发一个 `aud=end-user` Token，不做 scope exchange。Project 的访问控制由后端 RBAC（`end_user_role_users`）在查询时动态校验，不体现在 Token 中。

**中间件强制校验 `aud`：**
- 访问 `/org/[orgName]/...` 时，Token `aud` 必须为 `tenant`，否则 403
- 访问 `/end-user/[orgName]/...` 时，Token `aud` 必须为 `end-user`，否则 403

### 共享层

| 层级 | 是否共享 | 说明 |
|------|---------|------|
| Design System | ✓ 共享 | shadcn/ui 组件、Tailwind CSS 语义变量、图标 |
| 通用业务组件 | ✓ 允许共享 | 数据表格、弹窗、表单、Pagination 等可复用组件 |
| 路由 / 导航结构 | ✗ 不共享 | 两套 UI 导航结构完全不同 |
| 认证逻辑 / BFF 路由 | ✗ 不共享 | `aud` 不同，BFF 路由和 Token 处理各自独立 |
| Apollo Client 实例 | ✗ 不共享 | 端点不同，缓存隔离 |
| 全局状态（Zustand Store）| ✗ 不共享 | 两端状态互不干扰 |

---

## 5. 租户端（Tenant UI）

### 入口与登录

- **登录入口**：`/tenant/login` → 触发 AuthProvider OAuth 跳转
- **登录后**：`/org-selector` → 选择 Org → 进入 `/org/[orgName]/projects`

### 路由结构

```
/tenant/login                                   # OAuth 登录入口（租户专属）
/auth/callback                                  # OAuth 回调
/org-selector                                   # 登录后选择 Org
/org/[orgName]/
├── projects/                                   # 项目列表
│   └── [projectSlug]/
│       ├── model-editor/                       # 模型设计器（字段、外键、枚举）
│       ├── enums/                              # 枚举管理
│       ├── database/                           # 数据库集群配置
│       ├── rbac/                               # RBAC 权限配置（Role、PermissionBundle）
│       └── end-user-access/                    # EndUser 访问控制（Project 级 Role 分配）
├── end-users/                                  # EndUser 账号管理（Org 级）
├── team/                                       # 成员管理
└── settings/                                   # Org 设置
```

### 功能范围

**Org 级：**
- 项目 CRUD（创建、编辑、删除项目）
- 成员管理（邀请成员、设置成员 Role）
- EndUser 账号管理：创建账号（username + password）、禁用账号、删除账号、查看账号的 Project 访问列表
- Org 设置（名称、配置）

**Project 级：**
- 模型设计：创建模型、添加字段（Text/Number/Boolean/Date/Enum/Relation 等类型）、设置外键（Logical Foreign Key）
- 枚举管理：创建枚举、管理枚举值及排序
- 数据库集群：绑定数据库集群、配置连接信息
- RBAC：创建 Role、配置 PermissionBundle、版本快照管理
- EndUser 访问控制：为 Org 内 EndUser 分配 Project Role、修改角色、撤销访问

### 导航结构

顶部：Org 选择器 + 全局搜索  
侧边栏（Project 内）：模型设计器 / 枚举 / 数据库 / RBAC / 访问控制 / 项目设置  
侧边栏（Org 级）：项目列表 / 终端用户 / 成员 / 设置

---

## 6. 用户端（App UI）

### 入口与登录

- **登录入口**：`/end-user/[orgName]/login`（Org 级统一入口，不绑定特定 Project）
- **登录后**：统一跳转至 `/end-user/[orgName]/workspace`（Workspace 主页）

### 登录流程

```
EndUser 访问 /end-user/[orgName]/login
        ↓
输入 username + password
        ↓
BFF → Backend 验证（Org 级账号池）
        ↓
颁发 aud=end-user Token（httpOnly cookie）
        ↓
若 EndUser 在该 Org 无任何 Project 可访问（无 end_user_role_users 记录）
  → 登录成功，进入 Workspace，Projects tab 展示空状态 + 引导文案
若有可访问 Project
  → 登录成功，进入 Workspace，Projects tab 展示 Project 列表
```

> **不做 exchange，不存在 Project Token**。Project 访问权限由后端 RBAC 在每次请求时校验，Token 本身只携带 `endUserId` 和 `orgName`。

### 路由结构

```
/end-user/[orgName]/
├── login/                                      # EndUser 登录页
└── workspace/                                  # EndUser Workspace 主页（登录后落地页）
    ├── (tab: projects)                         # Tab 1：可访问的 Project 列表，点击进入数据操作
    ├── (tab: TBD)                              # Tab 2：待定（未来规划 API Key 管理等）
    └── [projectSlug]/
        └── data/                               # 某个 Project 的数据操作页（运行时 CRUD）
```

### Workspace 页面结构

**`/end-user/[orgName]/workspace`**

```
┌─────────────────────────────────────────────────┐
│  [Org名称]  [用户名]  [登出]                      │  ← 顶部栏
├─────────────────────────────────────────────────┤
│  Projects  |  （待定）                            │  ← Tab 导航
├─────────────────────────────────────────────────┤
│                                                 │
│  Projects Tab:                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐      │
│  │ Project A│  │ Project B│  │ Project C│      │  ← Project 卡片列表
│  └──────────┘  └──────────┘  └──────────┘      │
│                                                 │
│  点击任意 Project → 进入 /end-user/[org]/[proj]/data │
│                                                 │
└─────────────────────────────────────────────────┘
```

### 功能范围

- 登录 / 登出
- Workspace 主页（`/workspace`）：
  - **Tab 1 — Projects**：展示 EndUser 有权访问的 Project 列表（空状态时显示引导文案）；点击进入对应 Project 的数据页
  - **Tab 2 — 待定**：后期可扩展为 API Key 管理等功能
- Project 数据操作页（`/workspace/[projectSlug]/data`）：
  - 按模型分类浏览数据
  - 新增记录（表单，字段类型渲染由 JSON Schema 决定）
  - 编辑记录
  - 删除记录
  - 列表筛选、排序、分页
- **不提供**：模型设计、字段管理、枚举管理、RBAC 配置、成员管理等任何设计时功能

### 导航结构

顶部：当前 Org 名称 + 用户名 + 登出  
Tab 导航：Projects / （待定）  
Projects 页：Project 卡片网格  
数据页侧边栏：当前 Project 内的模型列表

---

## 7. 功能边界矩阵

| 功能 | 租户端 | 用户端 |
|------|--------|--------|
| 模型创建 / 编辑 / 删除 | ✓ | ✗ |
| 字段管理（添加/编辑/删除字段） | ✓ | ✗ |
| 枚举管理 | ✓ | ✗ |
| 逻辑外键（Logical Foreign Key） | ✓ | ✗ |
| 数据库集群配置 | ✓ | ✗ |
| RBAC Role / PermissionBundle 管理 | ✓ | ✗ |
| EndUser 账号创建 / 禁用 / 删除 | ✓（Org 级） | ✗ |
| EndUser Project Role 分配 / 撤销 | ✓（Project 级） | ✗ |
| 成员（Developer）管理 | ✓ | ✗ |
| 运行时数据 CRUD（增删改查） | ✗（设计时无需） | ✓ |
| 数据筛选 / 排序 / 分页 | ✗ | ✓ |
| 登录（Developer OAuth） | ✓ | ✗ |
| 登录（EndUser username/password） | ✗ | ✓ |
| Design System（shadcn/ui + Tailwind） | ✓（共享） | ✓（共享） |

---

## 8. 不做（Out of Scope）

| 项目 | 原因 |
|------|------|
| 租户端与用户端共享**导航组件** | 导航结构差异过大，共享只会产生大量条件分支 |
| 用户端独立的 Project 选择中间页（select-project）| Project 选择在 Workspace Tab 1 内完成，无需独立页面 |
| Project Token / scope exchange | Token 统一为 `aud=end-user`，Project 访问权限由后端 RBAC 校验，不做 Token 降维 |
| EndUser 自助注册账号 | 账号由 Org 管理员创建，v1 不开放自注册 |
| 跨 Org 账号共享 | Org 是最大隔离单元 |
| 用户端 SSO / OAuth 登录 | EndUser 使用 username/password，OAuth 是租户端体系 |
| 渐进式路由迁移（保留旧路径）| 明确切换，旧路径（`/login`、`/u/[orgName]/...`）废弃，不保留重定向 |

---

## 9. 成功指标

| 指标 | 目标值 |
|------|--------|
| Token `aud` 校验覆盖所有受保护路由 | 100% |
| 租户端 `/org/` 路由拒绝 `aud=end-user` Token | 403 |
| 用户端 `/end-user/` 路由拒绝 `aud=tenant` Token | 403 |
| 用户端页面组件不包含设计时功能入口（模型编辑、字段管理等）| 0 处 |
| EndUser 登录到数据页操作路径步骤数 ≤ 3（单 Project 场景）| ≤ 3 步 |
| 两套 UI 共用 Design System，无样式一致性问题 | 通过 UI Review |

---

## 10. 子文档索引

| 文件 | 说明 |
|------|------|
| [01-tenant-ui-routes.md](./01-tenant-ui-routes.md) | 租户端完整路由表与页面详细设计（待补充） |
| [02-app-ui-routes.md](./02-app-ui-routes.md) | 用户端完整路由表与页面详细设计（待补充） |
| [03-auth-boundary.md](./03-auth-boundary.md) | 两端认证体系边界与 BFF 路由映射（待补充） |

### 关联文档

- [enduser-access-model spec](../../openspec/specs/enduser-access-model/spec.md) — EndUser 可访问 Project 由 Role Assignment 决定
- [enduser-frontend spec](../../openspec/specs/enduser-frontend/spec.md) — 租户端 EndUser 访问控制页详细 spec
- [enduser-two-phase-auth spec](../../openspec/specs/enduser-two-phase-auth/spec.md) — 用户端两阶段认证详细 spec
- [前端架构总览](../front/development/architecture.md) — BFF 双体系路由、组件约定
