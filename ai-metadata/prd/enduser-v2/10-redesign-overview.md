# EndUser 身份系统重设计 — 总览

> **版本**: v2.0  
> **状态**: Spec（待实现）  
> **取代文档**: `00-end-user-auth.md`（v1 设计，账号 Project 级隔离）

---

## 问题陈述

v1 设计中 EndUser 账号绑定到 `(OrgName, ProjectSlug)`，每个 Project 独立账号池。

这带来一个核心限制：**同一家企业用 ModelCraft 搭了多个内部工具（Project A = 销售系统，Project B = HR 系统），内部员工必须为每个 Project 分别注册账号，且无法统一管理。**

---

## 设计目标

| 目标 | 说明 |
|------|------|
| **统一账号** | 同一人在 Org 内只有一个 EndUser 账号，跨 Project 复用 |
| **Project 级权限** | 同一 EndUser 在不同 Project 中拥有不同的 PermissionBundle |
| **职责分离** | Org 管人（账号生命周期），Project 管权（访问控制） |
| **统一登录入口** | EndUser 通过 Org 级统一入口登录，登录后选择 Project |

---

## 核心设计原则

```
Org 层：管"人"
  ├── EndUser 账号池（注册/禁用/删除/改密码）
  └── 统一登录入口 /org/{orgName}/login

Project 层：管"权"
  ├── 授权哪些 EndUser 可以进入此 Project
  └── 授权进入的 EndUser 拥有哪些 PermissionBundle
```

---

## UI 架构决策

**核心原则：管理端和用户端是两套独立 UI，不共享主导航结构。**

历史设计的困境根源在于试图用一套 UI 同时服务两类完全不同的受众。两者登录入口不同、权限体系不同、功能集合不重叠，强行共用只会让两端都难用。

### 管理端（Admin / Developer UI）

- **入口：** `/login` → 平台统一登录
- **受众：** 平台管理员、开发者
- **功能：** 模型设计、字段管理、RBAC 配置、EndUser 账号管理、数据库集群管理
- **路由前缀：** `/org/[orgName]/...`（现有结构保持不变）
- **特征：** 功能密集、配置导向、设计时操作

### 用户端（EndUser / Workspace UI）

- **入口：** `/u/[orgName]/login`（Org 级统一登录）
- **受众：** 终端用户（EndUser）
- **功能：** 登录 → 选择项目 → 进入 Workspace 操作数据（CRUD）
- **路由前缀：** `/u/[orgName]/...` → `/u/[orgName]/[projectSlug]/data`
- **特征：** 精简、数据操作导向、运行时视图

### 为什么不用一套 UI

| 维度 | 管理端 | 用户端 |
|------|--------|--------|
| 登录入口 | `/login` | `/u/[orgName]/login` |
| Token 体系 | `scope=org` JWT | `scope=org` → exchange → `scope=project` |
| 功能复杂度 | 高（设计时：模型/字段/RBAC） | 低（运行时：数据 CRUD） |
| 导航结构 | 多层级侧边栏（Project 内嵌套 Org） | 扁平 Workspace 面板 |
| 权限来源 | RBAC（Developer 角色） | RBAC PermissionBundle（EndUser 角色） |

两者功能集合不重叠，也不存在"共用组件因 role 差异需要大量条件渲染"的场景。两套 UI 的唯一共享层是 Design System（shadcn/ui + Tailwind 变量）。

---

## 变更范围

### 1. EndUser 账号归属

| 维度 | v1（现状） | v2（目标） |
|------|-----------|-----------|
| 账号作用域 | `(OrgName, ProjectSlug)` | `OrgName` |
| Username 唯一性 | Project 内唯一 | Org 内唯一 |
| 同一人跨项目 | 多个独立账号 | 同一账号，不同 Project 权限 |

### 2. 新增 EndUser ↔ Project 授权关系

新增 `EndUserProjectAccess` 实体：
- 表达"某个 EndUser 可以访问某个 Project，且在该 Project 内拥有某个 PermissionBundle"
- 一个 EndUser 可以有多条记录（一对多个 Project）
- 一个 Project 的一个 EndUser 可有多条 PermissionBundle 授权

### 3. API 边界迁移

| 接口 | v1 位置 | v2 位置 |
|------|--------|--------|
| 账号 CRUD（创建/查询/禁用/删除） | Project GraphQL | **Org GraphQL** |
| 登录 / 刷新 Token | Project BFF路由 | **Org BFF 路由** |
| Project 访问控制（授权/撤销/Bundle） | 无 | **Project GraphQL（新增）** |

### 4. 登录流程

```
EndUser 访问 /org/{orgName}/login
        ↓
输入 username + password
        ↓
后端验证（Org 级账号池）
        ↓
返回 EndUser 有权访问的 Project 列表
        ↓
若列表为空 → 报错："您暂无项目访问权限，请联系管理员授权"
若有项目   → 展示列表，EndUser 选择要进入的 Project
        ↓
签发 JWT: { endUserId, orgName, projectSlug, iss: "mc-enduser" }
        ↓
跳转至 /{orgName}/{projectSlug}/data
```

### 5. JWT 结构（不变）

```json
{
  "endUserId": "uuid",
  "orgName": "acme",
  "projectSlug": "sales-system",
  "iss": "mc-enduser",
  "role": "end_user"
}
```

> RLS 体系（`owner` 字段 + Policy）**完全不变**，继续按 `endUserId` 做行级隔离。

---

## 不做（v2 范围外）

| 项目 | 原因 |
|------|------|
| EndUser 自助申请 Project 访问 | 复杂度高，v1 由 Developer 主动授权 |
| 跨 Org 账号共享 | Org 是最大隔离单元，不打破 |
| Project 级角色体系另起炉灶 | 复用现有 PermissionBundle，避免两套系统 |
| 忘记密码 / 邮件重置 | 不在此次范围 |
| 短信 / OAuth 第三方登录 | 不在此次范围 |

---

## 子页文档

| 文件 | 说明 |
|------|------|
| [11-domain-model-changes.md](./11-domain-model-changes.md) | 领域模型变更与 PlantUML |
| [12-graphql-api-design.md](./12-graphql-api-design.md) | GraphQL 接口变更（新增/迁移/废弃） |
| [13-database-schema.md](./13-database-schema.md) | 数据库 Schema 变更（Atlas 迁移） |
| [14-frontend-design.md](./14-frontend-design.md) | 前端页面与路由变更 |
| [15-bdd-scenarios.md](./15-bdd-scenarios.md) | BDD 验收场景 |
