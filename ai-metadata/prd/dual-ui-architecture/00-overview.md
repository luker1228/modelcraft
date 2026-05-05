---
版本: v1.0
状态: Spec（待实现）
日期: 2026-05-05
---

# ModelCraft 前端双 UI 架构 PRD

## 1. 背景与问题陈述

ModelCraft 历史上试图用一套 UI 同时服务两类受众：管理模型和配置权限的开发者，以及通过工作区操作数据的终端用户。

这一做法带来的核心问题：

| 维度 | 表现 |
|------|------|
| **路由冲突** | 租户端路由 `/org/[orgName]/...` 与用户端路由 `/u/[orgName]/...` 概念上清晰，但实现时边界不断模糊 |
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

- 两套 UI 路由前缀完全隔离，代码层面无跨界耦合
- 认证体系各自独立：租户端走 Developer Token 体系，用户端走 EndUser 两阶段认证（Org Token → Project Token）
- 导航结构各自独立设计，不共享侧边栏或顶部导航
- **唯一共享层**：Design System（shadcn/ui + Tailwind CSS 变量）

---

## 3. 受众与角色定义

### 租户端用户（Tenant UI）

| 角色 | 描述 | 登录方式 |
|------|------|---------|
| Org Owner | 创建 Org，管理成员，拥有全部权限 | `/login` → OAuth / Developer Auth |
| Org Admin | 管理成员、EndUser 账号、数据库集群 | `/login` → OAuth / Developer Auth |
| Developer | 设计模型、配置字段、管理枚举、设置 RBAC | `/login` → OAuth / Developer Auth |

### 用户端用户（App UI）

| 角色 | 描述 | 登录方式 |
|------|------|---------|
| EndUser | 使用 Org 搭建的内部工具，对数据执行 CRUD 操作 | `/u/[orgName]/login` → EndUser 两阶段认证 |

> **职责分离**：Org 管人（EndUser 账号生命周期由租户端管理），Project 管权（EndUser 能访问哪些 Project、拥有哪些 Role 由租户端配置）。

---

## 4. 核心架构决策

### 路由边界

| UI | 路由前缀 | 登录入口 |
|----|---------|---------|
| **租户端（Tenant UI）** | `/org/[orgName]/...` | `/login` |
| **用户端（App UI）** | `/u/[orgName]/...` | `/u/[orgName]/login` |

两个前缀在 Next.js App Router 中是完全独立的路由树，互不嵌套。

### Token 体系差异

| UI | Token 类型 | 存储方式 | 生命周期 |
|----|-----------|---------|---------|
| 租户端 | Developer Token（`scope=org`） | AuthProvider SDK 管理 | 由 OAuth Provider 决定 |
| 用户端 | EndUser Org Token（`scope=org`）→ exchange → Project Token（`scope=project`） | httpOnly cookie（BFF 管理） | 短 TTL，BFF 静默刷新 |

### 共享层

仅共享 Design System：`shadcn/ui` 组件、Tailwind CSS 语义化变量、图标（Lucide React）。

**不共享**：路由、认证逻辑、BFF 路由、Apollo Client 实例、全局状态（Zustand Store）、导航组件。

---

## 5. 租户端（Tenant UI）

### 入口与登录

- **登录入口**：`/login` → 触发 AuthProvider OAuth 跳转
- **登录后**：`/org-selector` → 选择 Org → 进入 `/org/[orgName]/projects`

### 路由结构

```
/login                                          # OAuth 登录入口
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

- **登录入口**：`/u/[orgName]/login`（Org 级统一入口，不绑定特定 Project）
- **登录流程**：见下节

### 登录流程

```
EndUser 访问 /u/[orgName]/login
        ↓
输入 username + password
        ↓
BFF → Backend 验证（Org 级账号池）
        ↓
返回 EndUser 有权访问的 Project 列表（通过 end_user_role_users 查询）
        ↓
若列表为空 → 页面显示错误："您暂无项目访问权限，请联系管理员授权"
若只有 1 个 Project → 直接 exchange 签发 Project Token → 跳转至数据页
若有 N > 1 个 Project → 跳转至 /u/[orgName]/select-project（选择页）
        ↓
/u/[orgName]/select-project
        ↓
选择 Project → BFF exchange → 签发 Project Token（httpOnly cookie）
        ↓
跳转至 /u/[orgName]/[projectSlug]/data
```

### 路由结构

```
/u/[orgName]/
├── login/                                      # EndUser 登录页
├── select-project/                             # 多 Project 时的选择页
└── [projectSlug]/
    └── data/                                   # Workspace 数据操作页（运行时 CRUD）
```

### 功能范围

- 登录 / 登出
- 选择 Project（多 Project 时）
- Workspace 数据操作：
  - 按模型分类浏览数据
  - 新增记录（表单，字段类型渲染由 JSON Schema 决定）
  - 编辑记录
  - 删除记录
  - 列表筛选、排序、分页
- **不提供**：模型设计、字段管理、枚举管理、RBAC 配置、成员管理等任何设计时功能

### 导航结构

顶部：当前 Org + Project 名称 + 用户信息 + 登出  
侧边栏：模型列表（仅运行时可见的模型）  
主内容区：数据表格 + 操作面板

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
| 租户端与用户端共享导航组件 | 导航结构差异过大，共享只会产生大量条件分支 |
| 用户端提供模型预览（只读） | v1 聚焦数据操作，模型结构查看不是核心需求 |
| EndUser 自助注册账号 | 账号由 Org 管理员创建，v1 不开放自注册 |
| 跨 Org 账号共享 | Org 是最大隔离单元 |
| 用户端 SSO / OAuth 登录 | EndUser 使用 username/password，OAuth 是租户端体系 |
| 渐进式路由迁移（保留旧路径）| 明确切换，旧路径仅保留 `/u/[orgName]/[projectSlug]/login` → redirect，其余废弃 |

---

## 9. 成功指标

| 指标 | 目标值 |
|------|--------|
| 租户端 `/org/` 路由不引用任何 EndUser Token 处理逻辑 | 100% |
| 用户端 `/u/` 路由不引用任何 Developer Token 处理逻辑 | 100% |
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
