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
│   │   ├── gateway-architecture.md
│   │   ├── developer-enduser-system.md
│   │   ├── comments.md
│   │   ├── context-handling.md
│   │   ├── contract-sync.md
│   │   ├── domain-development.md
│   │   ├── error-handling.md
│   │   ├── logging.md
│   │   ├── repo-develop.md
│   │   ├── sqlc-custom-types.md
│   │   ├── soft-delete-sqlc.md
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
├── cli/                              # ⭐ ModelCraft CLI 使用指南
│   └── README.md                     # 命令参考、认证、输出格式、退出码
└── front/                            # 前端知识库
    ├── development/
    │   ├── README.md                 # 前端开发规范总览
    │   ├── architecture.md
    │   ├── api-client-design.md
    │   ├── code-conventions.md
    │   ├── eslint-rules.md
    │   ├── known-issues.md
    │   ├── react-best-practices.md
    │   ├── workspace-mode-boundary.md
    │   └── typescript-guide.md
    └── style/
        ├── STYLE.md
        ├── color-system.md
        ├── design-system-demo-v2.html
        ├── quick-start.md
        └── tailwind-usage-policy.md
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
| [backend/design/domain-model/1-auth.md](./backend/design/domain-model/1-auth.md) | 登录与认证域领域模型（统一用户体系、双视角、Token 设计） |
| [backend/design/domain-model/2-tenant.md](./backend/design/domain-model/2-tenant.md) | 租户域领域模型（Org 生命周期、用户与 Org 关系）|
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
| [backend/development/gateway-architecture.md](./backend/development/gateway-architecture.md) | ⭐ Gateway 架构、认证代理链路与运行配置 |
| [backend/development/developer-enduser-system.md](./backend/development/developer-enduser-system.md) | ⭐ Developer / EndUser 双体系（认证、路由、边界） |
| [backend/development/user-vs-end-user.md](./backend/development/user-vs-end-user.md) | ⭐ `User` / `EndUser` 身份边界、runtime 语义与 context 命名约定 |
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

## CLI 文档索引

| 路径 | 说明 |
|------|------|
| [cli/README.md](./cli/README.md) | ⭐ CLI 设计指南（PAT 登录、凭证文件结构、`describe` 获取 schema、`run` 执行查询、AI Agent 工作流） |

---

## Frontend 文档索引

### 开发规范

| 路径 | 说明 |
|------|------|
| [front/development/README.md](./front/development/README.md) | 前端开发规范总览、快速开始、常见错误 |
| [front/development/architecture.md](./front/development/architecture.md) | 目录分层、组件约定、GraphQL Codegen、BFF 双体系路由 |
| [front/development/api-client-design.md](./front/development/api-client-design.md) | API Client 层设计规范（GraphQL 文档分模块 + 页面级 MSW mock + 双体系 BFF 映射） |
| [front/development/code-conventions.md](./front/development/code-conventions.md) | 命名约定、导入顺序、代码风格 |
| [front/development/eslint-rules.md](./front/development/eslint-rules.md) | ESLint 配置与规则说明 |
| [front/development/known-issues.md](./front/development/known-issues.md) | 已知问题与临时解决方案 |
| [front/development/react-best-practices.md](./front/development/react-best-practices.md) | React 组件设计、State 管理、性能优化 |
| [front/development/workspace-mode-boundary.md](./front/development/workspace-mode-boundary.md) | ⭐ Workspace 复用组件的 Design/End User 能力边界 |
| [front/development/copilot-architecture.md](./front/development/copilot-architecture.md) | ⭐ Copilot AI 助手架构（CopilotKit Provider 体系、知识注入、前端工具、AiTarget 高亮导航） |
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
| [front/style/icons.md](./front/style/icons.md) | ⭐ 图标体系文档 — 62 个图标的语义分类、使用位置、渲染上下文及 AI Prompt |
| [front/style/icon-prompts.md](./front/style/icon-prompts.md) | 图标 AI 生成 Prompt 清单（P0–P2，15 个）— 含批量 sprite sheet 生成指令和切图命令 |
| [front/style/design-system-demo-v2.html](./front/style/design-system-demo-v2.html) | 设计系统可视化 Demo |
