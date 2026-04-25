# AI Metadata 路径索引

> 本文件是 `ai-metadata/` 目录的完整路径索引，所有知识文档的统一入口。
> 文档按模块和优先级组织，冲突时以 **设计理念 > 开发规范 > 测试策略 > 部署指南 > 工具手册** 为准。

---

## 目录结构

```
ai-metadata/
├── index.md                          # 本文件 - 路径总索引
├── backend/                          # Go 后端知识库
│   ├── design/                       # ⭐ 设计理念（最高优先级）
│   │   ├── README.md
│   │   ├── core-principles.md
│   │   ├── roadmap.md
│   │   └── domain-model/
│   │       ├── README.md
│   │       ├── 1-auth.md
│   │       ├── 2-tenant.md
│   │       ├── 3-project.md
│   │       ├── 4-rbac.md
│   │       ├── 5-model/
│   │       │   ├── README.md
│   │       │   ├── artifact.md
│   │       │   └── design.md
│   │       ├── 6-database-cluster.md
│   │       └── 7-sql-editor.md
│   ├── development/                  # 开发规范
│   │   ├── README.md
│   │   ├── architecture.md
│   │   ├── comments.md
│   │   ├── context-handling.md
│   │   ├── contract-sync.md
│   │   ├── domain-development.md
│   │   ├── error-handling.md
│   │   ├── logging.md
│   │   ├── repo-develop.md
│   │   ├── sqlc-custom-types.md
│   │   ├── tenant-scope-and-propagation.md
│   │   └── type-conversion.md
│   ├── testing/                      # 测试策略
│   │   ├── README.md
│   │   ├── adapter-testing.md
│   │   ├── bdd-testing-guidelines.md
│   │   └── debugging-workflow.md
│   ├── deployment/                   # 部署指南
│   │   └── README.md
│   ├── tools/                        # 工具手册
│   │   ├── README.md
│   │   ├── justfile-guide.md
│   │   └── tools-installation.md
│   └── common-mistakes.md            # ⚠️ 错题本（真实 Bug Checklist）
├── front/                            # 前端知识库
│   ├── development/
│   │   ├── README.md                 # 前端开发规范总览
│   │   ├── architecture.md
│   │   ├── bff-design.md
│   │   ├── code-conventions.md
│   │   ├── eslint-rules.md
│   │   ├── known-issues.md
│   │   ├── react-best-practices.md
│   │   ├── workspace-mode-boundary.md
│   │   └── typescript-guide.md
│   └── style/
│       ├── STYLE.md
│       ├── color-system.md
│       ├── design-system-demo-v2.html
│       ├── quick-start.md
│       └── tailwind-usage-policy.md
└── prd/                              # 产品需求文档
    ├── auth/
    │   ├── 00-auth.md
    │   ├── auth-api-design.md
    │   ├── auth-domain.puml
    │   ├── 01-auth-login.md
    │   └── 02-auth-register.md
    ├── model-enum/
    │   ├── 00-model-enum.md
    │   ├── 01-field-create-enum-binding.md
    │   ├── 02-field-edit-format-immutable.md
    │   ├── 03-backend-design.md
    │   ├── 04-frontend-subpage-design.md
    │   └── model-enum-domain.puml
    ├── rbac/                             # ⭐ 权限模型（RBAC）
    │   ├── 00-rbac-overview.md           # 总览：三层定位、核心原则
    │   ├── 01-permission-model.md        # 权限点、权限包、授权对象
    │   ├── 02-implicit-roles.md          # 内置隐式角色
    │   ├── 03-auth-flow.md               # 鉴权流程与判定规则
    │   └── 04-department-scope.md        # 部门与数据范围
    ├── enduser-v2/                       # ⭐ EndUser 身份系统 v2（Org 级账号）
    │   ├── 10-redesign-overview.md       # 总览：问题、目标、核心设计决策
    │   ├── 11-domain-model-changes.md    # 领域模型变更（实体/关系图）
    │   ├── 12-graphql-api-design.md      # GraphQL API 设计（Org/Project Schema）
    │   ├── 13-database-schema.md         # 数据库 Schema 变更（Atlas 迁移）
    │   ├── 14-frontend-design.md         # 前端页面/路由/BFF 变更
    │   └── 15-bdd-scenarios.md           # BDD 验收场景
    └── cli/                              # ⭐ ModelCraft CLI（Agent-First）
        ├── 00-cli-overview.md            # 总览：设计目标、命令树、v1 范围
        ├── 01-auth-flow.md               # 认证流程与 Token 管理
        ├── 02-data-commands.md           # 数据命令（query/get/create/update/delete）
        ├── 03-discovery-and-introspection.md  # 资源发现与 Agent 自省
        ├── 04-error-handling.md          # 错误处理与 Limit 机制
        └── 05-architecture.md            # CLI 架构与后端变更需求
```

---

## Backend 文档索引

### 设计理念（优先级最高）

| 路径 | 说明 |
|------|------|
| [backend/design/README.md](./backend/design/README.md) | 设计理念总览，核心原则摘要 |
| [backend/design/core-principles.md](./backend/design/core-principles.md) | 项目定位与重大技术抉择 |
| [backend/design/roadmap.md](./backend/design/roadmap.md) | 里程碑与功能规划 |
| [backend/design/domain-model/README.md](./backend/design/domain-model/README.md) | 领域模型总览 |
| [backend/design/domain-model/1-auth.md](./backend/design/domain-model/1-auth.md) | 认证域领域模型 |
| [backend/design/domain-model/2-tenant.md](./backend/design/domain-model/2-tenant.md) | 租户域领域模型 |
| [backend/design/domain-model/3-project.md](./backend/design/domain-model/3-project.md) | 项目域领域模型 |
| [backend/design/domain-model/4-rbac.md](./backend/design/domain-model/4-rbac.md) | RBAC 权限域领域模型 |
| [backend/design/domain-model/5-model/README.md](./backend/design/domain-model/5-model/README.md) | 模型域领域模型总览 |
| [backend/design/domain-model/5-model/artifact.md](./backend/design/domain-model/5-model/artifact.md) | 模型产物设计 |
| [backend/design/domain-model/5-model/design.md](./backend/design/domain-model/5-model/design.md) | 模型设计细节 |
| [backend/design/domain-model/6-database-cluster.md](./backend/design/domain-model/6-database-cluster.md) | 数据库集群域领域模型 |
| [backend/design/domain-model/7-sql-editor.md](./backend/design/domain-model/7-sql-editor.md) | SQL 编辑器域领域模型 |
| [backend/design/domain-model/8-runtime/jsonschema-contract.md](./backend/design/domain-model/8-runtime/jsonschema-contract.md) | ⭐ Runtime JSON Schema 契约（`x-mc` 命名空间、`widget` 规范、字段全集） |

### 开发规范

| 路径 | 说明 |
|------|------|
| [backend/development/README.md](./backend/development/README.md) | 开发规范总览 & AI 使用指南 |
| [backend/development/architecture.md](./backend/development/architecture.md) | DDD 分层架构、依赖规则、目录映射 |
| [backend/development/comments.md](./backend/development/comments.md) | 代码注释规范 |
| [backend/development/context-handling.md](./backend/development/context-handling.md) | Context 传递与使用规范 |
| [backend/development/contract-sync.md](./backend/development/contract-sync.md) | GraphQL Schema 规范与代码生成工作流 |
| [backend/development/domain-development.md](./backend/development/domain-development.md) | Domain 层 Repository 接口设计规范 |
| [backend/development/error-handling.md](./backend/development/error-handling.md) | 错误包体系、各层错误职责、RecordNotFound 处理 |
| [backend/development/logging.md](./backend/development/logging.md) | logfacade 使用规范、Stack() 使用约束 |
| [backend/development/repo-develop.md](./backend/development/repo-develop.md) | Repository 层开发规范，Go Wrapper 架构 |
| [backend/development/sqlc-custom-types.md](./backend/development/sqlc-custom-types.md) | sqlc 自定义类型实现标准 |
| [backend/development/tenant-scope-and-propagation.md](./backend/development/tenant-scope-and-propagation.md) | 租户隔离（org / org+project）与参数全链路传递规范 |
| [backend/development/type-conversion.md](./backend/development/type-conversion.md) | 类型转换规范 |

### 测试策略

| 路径 | 说明 |
|------|------|
| [backend/testing/README.md](./backend/testing/README.md) | 测试策略总览、测试金字塔、覆盖率要求 |
| [backend/testing/adapter-testing.md](./backend/testing/adapter-testing.md) | Adapter 契约测试规范（table-driven、golden、fuzz、invariants） |
| [backend/testing/bdd-testing-guidelines.md](./backend/testing/bdd-testing-guidelines.md) | BDD 验收测试注意要点（默认不耦合注册） |
| [backend/testing/debugging-workflow.md](./backend/testing/debugging-workflow.md) | ⭐ 日常开发调试流程（必读） |

### 部署指南

| 路径 | 说明 |
|------|------|
| [backend/deployment/README.md](./backend/deployment/README.md) | Docker 环境要求、部署流程、常用命令 |

### 工具手册

| 路径 | 说明 |
|------|------|
| [backend/tools/README.md](./backend/tools/README.md) | 工具手册总览 |
| [backend/tools/justfile-guide.md](./backend/tools/justfile-guide.md) | 所有 `just` 命令参考 |
| [backend/tools/tools-installation.md](./backend/tools/tools-installation.md) | goenv / just / Atlas / jq 安装指南 |

### ⚠️ 错题本（最高优先级参考）

| 路径 | 说明 |
|------|------|
| [backend/common-mistakes.md](./backend/common-mistakes.md) | 真实 Bug 案例 + Checklist 规则，代码审查前必读 |

---

## Frontend 文档索引

### 开发规范

| 路径 | 说明 |
|------|------|
| [front/development/README.md](./front/development/README.md) | 前端开发规范总览、快速开始、常见错误 |
| [front/development/architecture.md](./front/development/architecture.md) | 目录分层、组件约定、GraphQL Codegen |
| [front/development/bff-design.md](./front/development/bff-design.md) | BFF 层设计规范 |
| [front/development/code-conventions.md](./front/development/code-conventions.md) | 命名约定、导入顺序、代码风格 |
| [front/development/eslint-rules.md](./front/development/eslint-rules.md) | ESLint 配置与规则说明 |
| [front/development/known-issues.md](./front/development/known-issues.md) | 已知问题与临时解决方案 |
| [front/development/react-best-practices.md](./front/development/react-best-practices.md) | React 组件设计、State 管理、性能优化 |
| [front/development/workspace-mode-boundary.md](./front/development/workspace-mode-boundary.md) | ⭐ Workspace 复用组件的 Design/End User 能力边界 |
| [front/development/typescript-guide.md](./front/development/typescript-guide.md) | TypeScript 严格模式、组件类型、Hooks 类型 |

### 设计系统

| 路径 | 说明 |
|------|------|
| [front/style/ui-spec.md](./front/style/ui-spec.md) | ⭐ **UI 规范 v2（单一真相源）** — Stripe 方向：token、组件规范、禁止事项全集 |
| [front/style/STYLE.md](./front/style/STYLE.md) | 设计系统整体规范（旧版，ui-spec.md 优先）|
| [front/style/color-system.md](./front/style/color-system.md) | 颜色系统与语义化变量 |
| [front/style/quick-start.md](./front/style/quick-start.md) | 设计系统快速上手 |
| [front/style/tailwind-usage-policy.md](./front/style/tailwind-usage-policy.md) | Tailwind CSS 使用策略 |
| [front/style/ui-checklist.md](./front/style/ui-checklist.md) | ⭐ UI Checklist — 选中态、颜色、布局、Lint 提交前必查 |
| [front/style/design-system-demo-v2.html](./front/style/design-system-demo-v2.html) | 设计系统可视化 Demo |

---

## PRD 文档索引

### 认证模块

| 路径 | 说明 |
|------|------|
| [prd/auth/00-auth.md](./prd/auth/00-auth.md) | 认证模块需求总览 |
| [prd/auth/auth-api-design.md](./prd/auth/auth-api-design.md) | 认证 API 设计 |
| [prd/auth/auth-domain.puml](./prd/auth/auth-domain.puml) | 认证域 PlantUML 类图 |
| [prd/auth/01-auth-login.md](./prd/auth/01-auth-login.md) | 登录流程需求 |
| [prd/auth/02-auth-register.md](./prd/auth/02-auth-register.md) | 注册流程需求 |

### 字段展示协议

| 路径 | 说明 |
|------|------|
| [prd/field/00-field-label-field.md](./prd/field/00-field-label-field.md) | 关系字段统一展示协议（`__label` + 模型级 `displayField`） |
| [prd/field/plan.md](./prd/field/plan.md) | 后端实现逻辑计划（直接切换、无 fallback） |

### 枚举字段关联（Model Enum）

| 路径 | 说明 |
|------|------|
| [prd/model-enum/00-model-enum.md](./prd/model-enum/00-model-enum.md) | 总览：规则边界、单一真相与子页索引 |
| [prd/model-enum/01-field-create-enum-binding.md](./prd/model-enum/01-field-create-enum-binding.md) | 前端子页：创建 ENUM 字段交互设计 |
| [prd/model-enum/02-field-edit-format-immutable.md](./prd/model-enum/02-field-edit-format-immutable.md) | 前端子页：字段编辑页（format 不可变） |
| [prd/model-enum/03-backend-design.md](./prd/model-enum/03-backend-design.md) | 后端方案：无迁移、含 BDD 场景设计 |
| [prd/model-enum/04-frontend-subpage-design.md](./prd/model-enum/04-frontend-subpage-design.md) | 前端合并交互方案（基于 01/02） |
| [prd/model-enum/model-enum-domain.puml](./prd/model-enum/model-enum-domain.puml) | 领域模型：Field / Enum / FieldEnumRelation 关系图 |

### 运行时（Runtime）

| 路径 | 说明 |
|------|------|
| [prd/runtime/field-relation-selector.md](./prd/runtime/field-relation-selector.md) | 多对一外键字段升级为下拉选择器，依赖 → [jsonschema-contract](./backend/design/domain-model/8-runtime/jsonschema-contract.md) |

### 权限模型（RBAC）

> 与 `end-user-auth/`（认证层）和 `rls/`（数据隔离层）协同，负责业务鉴权。

| 路径 | 说明 |
|------|------|
| [prd/rbac/00-rbac-overview.md](./prd/rbac/00-rbac-overview.md) | ⭐ 总览：三层定位关系、核心设计原则、子页索引 |
| [prd/rbac/01-permission-model.md](./prd/rbac/01-permission-model.md) | 权限点、权限包、授权对象完整数据模型 |
| [prd/rbac/02-implicit-roles.md](./prd/rbac/02-implicit-roles.md) | 内置隐式角色：落库定义、关系隐式、运行时自动注入 |
| [prd/rbac/03-auth-flow.md](./prd/rbac/03-auth-flow.md) | 鉴权流程：三通道权限来源合并 → 展开 → 判定 |
| [prd/rbac/04-department-scope.md](./prd/rbac/04-department-scope.md) | 部门职责定位：数据范围计算上下文，非授权载体 |

### EndUser 身份系统 v2（Org 级账号）

> EndUser 账号从 Project 级上移到 Org 级，建立 EndUser ↔ Project 多对多授权关系。

| 路径 | 说明 |
|------|------|
| [prd/enduser-v2/10-redesign-overview.md](./prd/enduser-v2/10-redesign-overview.md) | ⭐ 总览：问题陈述、目标、核心原则（Org 管人 Project 管权）、登录流程 |
| [prd/enduser-v2/11-domain-model-changes.md](./prd/enduser-v2/11-domain-model-changes.md) | 领域模型变更：EndUser 实体、新增 EndUserProjectAccess、Repository 接口更新 |
| [prd/enduser-v2/12-graphql-api-design.md](./prd/enduser-v2/12-graphql-api-design.md) | GraphQL API 设计：Org Schema 新接口、Project Schema access 接口、BFF 路由变更 |
| [prd/enduser-v2/13-database-schema.md](./prd/enduser-v2/13-database-schema.md) | 数据库 Schema 变更：DDL 对比、迁移策略、Atlas 操作指南 |
| [prd/enduser-v2/14-frontend-design.md](./prd/enduser-v2/14-frontend-design.md) | 前端设计：路由变更、新增页面、BFF 接口、登录流程时序 |
| [prd/enduser-v2/15-bdd-scenarios.md](./prd/enduser-v2/15-bdd-scenarios.md) | BDD 验收场景：账号管理、访问控制、登录流程、数据隔离、兼容性 |

### ModelCraft CLI（Agent-First 命令行工具）

> EndUser 通过 CLI 登录、查询和操作数据。第一版本面向 AI Agent，输出结构化 JSON。

| 路径 | 说明 |
|------|------|
| [prd/cli/00-cli-overview.md](./prd/cli/00-cli-overview.md) | ⭐ 总览：设计目标、资源路径约定、命令树、v1 范围 |
| [prd/cli/01-auth-flow.md](./prd/cli/01-auth-flow.md) | 认证流程：公共端点设计、登录时序、Token 管理 |
| [prd/cli/02-data-commands.md](./prd/cli/02-data-commands.md) | 数据命令：query/get/create/update/delete/count/aggregate |
| [prd/cli/03-discovery-and-introspection.md](./prd/cli/03-discovery-and-introspection.md) | 资源发现与 Agent 自省：catalog/describe/schema |
| [prd/cli/04-error-handling.md](./prd/cli/04-error-handling.md) | 错误处理：统一格式、退出码、Agent 自修正、Limit 机制 |
| [prd/cli/05-architecture.md](./prd/cli/05-architecture.md) | CLI 内部架构与后端变更需求 |
