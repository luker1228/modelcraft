# RLS Context + 开放数据 API 设计

**日期**: 2026-06-12
**状态**: Draft
**范围**: Backend + SDK

---

## 1. 背景

ModelCraft 是一个低代码平台。客户（开发者）有自己的用户体系，不需要 ModelCraft 管理用户、角色、权限。他们只需要：

1. 用 API Token 认证
2. 传当前用户上下文（user_id / user_name / user_role）
3. ModelCraft 按客户配置的 RLS 规则过滤数据行
4. 拿到过滤后的数据，返回给自己的用户

### 与现有 RBAC 的关系

现有 [End-User Runtime 数据权限体系](./2026-05-06-enduser-runtime-data-permission-design.md) 假设 ModelCraft 管理用户和角色（EndUser + Permission Bundle + Role）。本设计面向的是**另一种场景**：客户自有用户体系，ModelCraft 不感知用户。两者互补，不互相替代。

---

## 2. 架构

```
客户后端（自有用户体系）
    │
    │  Authorization: Bearer mc_pat_xxx
    │  X-MC-User-ID: customer_user_123
    │  X-MC-User-Name: zhangsan
    │  X-MC-User-Role: manager
    │
    ▼
ModelCraft 数据 API
    │
    │  1. PAT 验证（复用现有 ChiPATAuthMiddleware）
    │  2. 提取 RLS context header → 注入 context
    │  3. RLS 引擎：将 {{user_id}} 替换为 "customer_user_123"
    │     WHERE tenant_id = {{user_id}}
    │     → WHERE tenant_id = 'customer_user_123'
    │
    ▼
过滤后的数据 → 返回客户后端
```

### 中间件链

```
Request
  ↓
ApiTokenToIdentityMiddleware    ← 已有（2026-06-04 设计），验证 mc_pat_xxx
  ↓
RLSContextMiddleware             ← 新增，提取 X-MC-User-* → context
  ↓
ModelRuntimeHandler             ← 已有，执行 GraphQL + RLS 过滤
```

---

## 3. RLS Context Header

客户后端通过 HTTP Header 传递当前用户信息。三个关键字全部可选。

| Header | RLS 变量 | 示例 |
|--------|---------|------|
| `X-MC-User-ID` | `{{user_id}}` | `customer_user_123` |
| `X-MC-User-Name` | `{{user_name}}` | `zhangsan` |
| `X-MC-User-Role` | `{{user_role}}` | `manager` |

### 安全边界

Header 值由客户后端控制。ModelCraft 不做校验、不定义语义——只是透传变量。信任边界在 PAT 层：客户持有 PAT 即拥有对 project 数据的全部访问权。Header 的作用是**缩小**数据范围（通过 RLS），而非扩大。

不存在"客户篡改 header 骗自己"的场景——客户全权控制自己的后端。

---

## 4. 认证 — 复用现有 PAT

无需新增认证机制。客户后端使用现有 `mc_pat_xxx` 格式的 API Token。

已有组件：
- `end_user_api_tokens` 表
- `ChiPATAuthMiddleware`（PAT 验证 + EndUser 身份注入）
- PAT 管理接口（GraphQL createEndUserAPIToken / revokeEndUserAPIToken）

### 两层身份

| 身份层 | 来源 | 含义 |
|--------|------|------|
| PAT 的 EndUserID | `mc_pat_xxx` token → DB 查询 | 客户在 ModelCraft 的"服务账号"，决定能访问哪些 project |
| RLS Context | `X-MC-User-*` headers | 客户的终端用户，PAT 代表客户后端，RLS context 代表客户后端正在服务的那个人 |

两个身份互不干扰：
- PAT 决定"能不能访问这个 project"
- RLS context 决定"能访问哪些数据行"

---

## 5. RLS 策略 DSL

语义上对齐 PostgreSQL RLS（USING / WITH CHECK / PERMISSIVE / RESTRICTIVE），但使用 ModelCraft 自己的 DSL，不与 PG SQL 语法绑定。底层数据库为 MySQL，ModelCraft 在应用层实现 RLS。

### 5.1 DSL 结构

每条策略包含以下字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `policyName` | string | 策略名称，model 内唯一 |
| `action` | enum | `read` / `create` / `update` / `delete` |
| `using` | string? | 行过滤表达式（SQL WHERE 片段） |
| `withCheck` | string? | 新行校验表达式（SQL WHERE 片段） |
| `mode` | enum | `PERMISSIVE`（默认）/ `RESTRICTIVE` |

### 5.2 action 与 USING / WITH CHECK 的关系

| action | using | withCheck |
|--------|-------|-----------|
| `read` | ✅ 必须 | ❌ 不可配置 |
| `create` | ❌ 不可配置 | ✅ 必须 |
| `update` | ✅ 必须 | ✅ 必须 |
| `delete` | ✅ 必须 | ❌ 不可配置 |

- **using**：过滤已有行，不满足的行**静默忽略**
- **withCheck**：校验新行，不满足则**报错拒绝**

### 5.3 DSL 示例

```json
[
  {
    "policyName": "user_own_orders",
    "action": "read",
    "mode": "PERMISSIVE",
    "using": "tenant_id = '{{user_id}}'"
  },
  {
    "policyName": "user_own_orders",
    "action": "create",
    "mode": "PERMISSIVE",
    "withCheck": "tenant_id = '{{user_id}}'"
  },
  {
    "policyName": "user_own_orders",
    "action": "update",
    "mode": "PERMISSIVE",
    "using": "tenant_id = '{{user_id}}'",
    "withCheck": "tenant_id = '{{user_id}}'"
  },
  {
    "policyName": "user_own_orders",
    "action": "delete",
    "mode": "PERMISSIVE",
    "using": "tenant_id = '{{user_id}}'"
  },
  {
    "policyName": "admin_all_orders",
    "action": "read",
    "mode": "PERMISSIVE",
    "using": "'{{user_role}}' = 'admin'"
  },
  {
    "policyName": "hide_deleted",
    "action": "read",
    "mode": "RESTRICTIVE",
    "using": "deleted_at IS NULL"
  }
]
```

> 同一个 `policyName` 可以出现在多个 action 中（如 `user_own_orders` 覆盖了 read / create / update / delete）。每个 action 独立定义 using / withCheck。

### 5.4 策略合并逻辑

```
请求 action = "read"
  ↓
匹配 action 为 "read" 的所有策略
  ↓
PERMISSIVE 策略之间 → OR
    user_own_orders                    OR  admin_all_orders
    (tenant_id = '123')                OR  ('manager' = 'admin')
    → true                                       → false
    → 最终: true
  ↓
RESTRICTIVE 策略之间 → AND
    hide_deleted
    (deleted_at IS NULL)
    → true
  ↓
最终 → PERMISSIVE 结果 AND RESTRICTIVE 结果 → 放行
```

### 5.5 默认拒绝

当 model 开启 RLS 但没有配置任何策略时，所有操作被拒绝。

### 5.6 变量替换

RLS 引擎执行前，将模板变量替换为 Header 实际值：

| 模板变量 | Header 来源 | 未传时的行为 |
|---------|------------|-------------|
| `{{user_id}}` | `X-MC-User-ID` | 空字符串 `''` |
| `{{user_name}}` | `X-MC-User-Name` | 空字符串 `''` |
| `{{user_role}}` | `X-MC-User-Role` | 空字符串 `''` |

替换后的表达式拼入 SQL WHERE / CHECK 子句。

### 5.7 策略执行流程

```
请求进入（action = read / create / update / delete）
  ↓
RLS 是否开启？ → 否 → 跳过，正常执行
  ↓ 是
提取 RLS context（从 Header）
  ↓
匹配当前 action 的所有策略
  ├── PERMISSIVE → 替换变量 → using 表达式 OR 合并
  └── RESTRICTIVE → 替换变量 → using 表达式 AND 合并
  ↓
USING 表达式 → 注入 SQL WHERE（静默过滤）
  ↓
WITH CHECK 表达式（create / update）→ 注入 SQL CHECK（不满足则报错）
  ↓
执行查询 → 返回过滤后的数据
```

---

## 6. 数据 SDK

为客户后端提供轻量 SDK，封装 HTTP 调用和 Header 设置。

### 6.1 SDK 形态（Python 示例）

```python
from modelcraft import Client

client = Client(
    endpoint="https://your-modelcraft/api/data",
    api_token="mc_pat_xxx",
    org_name="my-org",
    project_slug="my-project",
)

# 查询数据 — 一行代码，RLS 自动生效
orders = client.query(
    db="main",
    model="orders",
    user_context={
        "user_id": "customer_123",
        "user_name": "zhangsan",
        "user_role": "manager",
    }
).list(limit=10)

# 创建数据 — RLS force 字段自动注入
client.query(
    db="main",
    model="orders",
    user_context={"user_id": "customer_123"}
).create({"amount": 100, "product": "widget"})
```

### 6.2 SDK 职责

- 封装 `Authorization: Bearer mc_pat_xxx`
- 自动设置 `X-MC-User-*` headers
- 封装数据 API 端点 URL 构造
- V1 支持 Python，后续按需扩展 Node.js / Go

---

## 7. 数据 API 端点

V1 使用现有 Runtime GraphQL 端点（复用现有 modelruntime 执行引擎）：

```
POST /end-user/graphql/org/{orgName}/project/{projectSlug}/db/{db}/model/{model}
```

SDK 封装此端点。客户也可直接调 HTTP。

---

## 8. 不在本次范围内

| 排除项 | 说明 |
|--------|------|
| OIDC / 社交登录 | 客户自有用户体系，ModelCraft 不需要 |
| 自建角色权限系统 | 客户管理自己的 RBAC |
| EndUser 管理 | 客户不通过 ModelCraft 创建用户 |
| Column-level RLS | V1 仅行级过滤 |
| DEPT / DEPT_AND_CHILDREN rowScope | 客户通过 RLS 表达式自行实现 |
| Webhook / 事件通知 | 后续迭代 |

---

## 9. 与现有设计的差异

| 维度 | 现有 RBAC 设计 (2026-05-06) | 本设计 |
|------|---------------------------|--------|
| 用户管理 | ModelCraft 管理 EndUser | 客户自行管理 |
| 权限模型 | Role → Bundle → Permission | RLS 策略（USING / WITH CHECK），对齐 PG RLS |
| 行过滤 | SELF / ALL（固定 scope） | 自定义 SQL 表达式 + PERMISSIVE / RESTRICTIVE |
| 策略合并 | 取最宽 rowScope（ALL > SELF） | PERMISSIVE OR / RESTRICTIVE AND |
| 写校验 | 无 | WITH CHECK 表达式 |
| 适用场景 | 客户无自有用户体系 | 客户有自有用户体系 |
| 用户标识 | EndUserID（ModelCraft 内部 ID） | Header 透传变量（客户定义） |

---

## 10. 实现顺序

1. RLS Context Middleware（提取 Header → context）
2. RLS 策略存储（DB schema + domain + CRUD API）
3. RLS 引擎（策略匹配 + 变量替换 + WHERE 注入）
4. 接入 modelruntime（现有 GraphQL 执行引擎）
5. Python SDK
