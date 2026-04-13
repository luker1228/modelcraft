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
│   │   └── type-conversion.md
│   ├── testing/                      # 测试策略
│   │   ├── README.md
│   │   └── debugging-workflow.md
│   ├── deployment/                   # 部署指南
│   │   └── README.md
│   └── tools/                        # 工具手册
│       ├── README.md
│       ├── justfile-guide.md
│       └── tools-installation.md
├── front/                            # 前端知识库
│   ├── development/
│   │   ├── README.md                 # 前端开发规范总览
│   │   ├── architecture.md
│   │   ├── bff-design.md
│   │   ├── code-conventions.md
│   │   ├── eslint-rules.md
│   │   ├── known-issues.md
│   │   ├── react-best-practices.md
│   │   └── typescript-guide.md
│   └── style/
│       ├── STYLE.md
│       ├── color-system.md
│       ├── design-system-demo-v2.html
│       ├── quick-start.md
│       └── tailwind-usage-policy.md
└── prd/                              # 产品需求文档
    ├── auth/
    │   ├── auth.md
    │   ├── auth-api-design.md
    │   ├── auth-domain.puml
    │   ├── auth-login.md
    │   └── auth-register.md
    └── model-enum/
        ├── 00-model-enum.md
        ├── 01-field-create-enum-binding.md
        ├── 02-field-edit-format-immutable.md
        ├── 03-backend-design.md
        ├── 04-frontend-subpage-design.md
        └── model-enum-domain.puml
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
| [backend/development/type-conversion.md](./backend/development/type-conversion.md) | 类型转换规范 |

### 测试策略

| 路径 | 说明 |
|------|------|
| [backend/testing/README.md](./backend/testing/README.md) | 测试策略总览、测试金字塔、覆盖率要求 |
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
| [front/development/typescript-guide.md](./front/development/typescript-guide.md) | TypeScript 严格模式、组件类型、Hooks 类型 |

### 设计系统

| 路径 | 说明 |
|------|------|
| [front/style/STYLE.md](./front/style/STYLE.md) | 设计系统整体规范 |
| [front/style/color-system.md](./front/style/color-system.md) | 颜色系统与语义化变量 |
| [front/style/quick-start.md](./front/style/quick-start.md) | 设计系统快速上手 |
| [front/style/tailwind-usage-policy.md](./front/style/tailwind-usage-policy.md) | Tailwind CSS 使用策略 |
| [front/style/design-system-demo-v2.html](./front/style/design-system-demo-v2.html) | 设计系统可视化 Demo |

---

## PRD 文档索引

### 认证模块

| 路径 | 说明 |
|------|------|
| [prd/auth/auth.md](./prd/auth/auth.md) | 认证模块需求总览 |
| [prd/auth/auth-api-design.md](./prd/auth/auth-api-design.md) | 认证 API 设计 |
| [prd/auth/auth-domain.puml](./prd/auth/auth-domain.puml) | 认证域 PlantUML 类图 |
| [prd/auth/auth-login.md](./prd/auth/auth-login.md) | 登录流程需求 |
| [prd/auth/auth-register.md](./prd/auth/auth-register.md) | 注册流程需求 |

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

