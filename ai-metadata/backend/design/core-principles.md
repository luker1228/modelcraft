# Core Principles

> 本文档定义 ModelCraft 的项目定位和重大技术抉择。这些决定是不可随意推翻的，任何与此冲突的实现方案都应以此为准。

## 项目定位

ModelCraft 是一个**面向开发团队的低代码数据模型管理平台**。

核心价值链：**可视化设计模型 → 自动同步数据库 Schema → 自动生成 GraphQL API**

用户无需编写代码，即可完成从数据建模到 API 消费的完整链路。

### 两个核心阶段

```
设计态 (Design-time)                运行态 (Runtime)
─────────────────────               ────────────────────
用户在 ModelCraft 中                 客户应用通过 GraphQL
定义模型、字段、关联关系               消费数据，CRUD 操作
        │                                    ▲
        ▼                                    │
  同步到目标 MySQL 数据库 ────────────────────┘
```

两态完全解耦：设计态变更不影响运行态的可用性；运行态可独立部署。

---

## 重大技术抉择

### 1. 运行态 API 只提供 GraphQL，不提供 REST

**决定**：客户通过 GraphQL 查询和操作数据，不暴露 REST endpoint。

**理由**：
- GraphQL 天然支持按需取字段，避免 over-fetch / under-fetch
- Schema 可根据模型定义动态生成，无需手写接口
- 客户端可通过单一 endpoint 完成所有数据操作

**影响**：运行态的所有 CRUD、过滤、聚合能力，均通过 GraphQL 表达。

---

### 2. 双 GraphQL 入口：设计态静态 Schema + 运行态动态 Schema

**决定**：系统存在两个独立的 GraphQL 入口，Schema 生成方式完全不同。

```
设计态 GraphQL                        运行态 GraphQL
─────────────────────────────         ──────────────────────────────────
/org/modelcraft/design/graphql        /:orgName/:projectSlug/:db/:model
静态 Schema（.graphql 文件定义）        动态 Schema（根据模型定义实时生成）
管理模型、字段、集群、项目               查询和操作客户数据
```

**理由**：
- 设计态操作的是 ModelCraft 自身的元数据，Schema 固定，适合静态定义
- 运行态操作的是客户的业务数据，每个模型结构不同，必须动态生成
- 两者职责边界清晰，互不干扰

---

### 3. 认证委托给 AuthProvider，ModelCraft 不自建用户体系

**决定**：用户身份由 AuthProvider 管理，ModelCraft 只持有 `ExternalID`（来自 JWT.sub），不存储密码。

**理由**：
- 认证是通用基础设施，不是 ModelCraft 的核心竞争力
- AuthProvider 提供完整的 SSO、OAuth2、OIDC 支持
- ModelCraft 只关心"这个用户在哪个 Org 里有什么权限"

**影响**：
- `User` 实体通过 `ExternalID` 与 AuthProvider 用户关联
- JWT token 由 AuthProvider 签发，ModelCraft 只做验证
- 未来可扩展支持 Keycloak / 通用 OIDC（代码中已预留 `ProviderType`）

---

### 4. 长期只支持 SQL 系数据库，不支持 NoSQL

**决定**：ModelCraft 的目标数据库永远是 SQL 系（当前 MySQL，未来可扩展 PostgreSQL 等 SQL 方言），**不支持 MongoDB 等 NoSQL 数据库**。

**理由**：
- SQL 的强 Schema 约束与 ModelCraft 的"设计态定义 Schema"理念天然契合
- NoSQL 的 Schema-less 特性与模型设计的核心价值冲突
- 专注 SQL 系可以做深，而不是泛泛支持所有数据库

**影响**：
- 所有运行态查询翻译目标为 SQL
- 字段类型体系以 SQL 类型为基础设计
- 架构上不需要抽象 NoSQL 适配层

---

### 5. 设计态与运行态解耦，运行态可独立部署

**决定**：运行态（Runtime GraphQL）不依赖设计态的实时状态，通过快照/同步机制独立工作。

**影响**：
- 设计态修改模型后，需要显式"同步"到目标数据库
- 运行态读取的是已同步的 Schema，而非设计态的实时状态
- 部署上两者可以分开，互不影响可用性
