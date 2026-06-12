# RLS Context + 开放数据 API 设计

**日期**: 2026-06-12
**状态**: Draft
**范围**: Backend + SDK

---

## 1. 背景

ModelCraft 是一个低代码平台。客户（开发者）有自己的用户体系，不需要 ModelCraft 管理用户、角色、权限。他们只需要：

1. 用 API Token 认证
2. 传当前用户上下文（user_id / user_name / user_roles）
3. ModelCraft 按角色匹配 RLS 策略，过滤数据行
4. 拿到过滤后的数据，返回给自己的用户

### 与现有 RBAC 的关系

现有 [End-User Runtime 数据权限体系](./2026-05-06-enduser-runtime-data-permission-design.md) 假设 ModelCraft 管理用户和角色。本设计面向**客户自有用户体系**的场景，两者互补。

---

## 2. 核心设计

### 2.1 三层模型

```
Header                               策略存储                  请求执行
──────                               ────────                  ────────
X-MC-User-ID: 123                    ┌──────────────────┐     
X-MC-User-Name: zhangsan             │ policyName        │     action = "read"
X-MC-User-Roles: admin,manager       │ action: read      │     roles = [admin, manager]
                                     │ role: admin       │───→ 匹配 policy (action=read AND role IN roles)
                                     │ using: {...}      │          ↓
                                     └──────────────────┘     OR 合并所有匹配策略的 expression
                                     ┌──────────────────┐          ↓
                                     │ policyName        │     注入 SQL WHERE
                                     │ action: read      │
                                     │ role: manager     │
                                     │ using: {...}      │
                                     └──────────────────┘
```

### 2.2 role 职责分离

| role 做的事 | role 不做的事 |
|------------|-------------|
| 决定哪些策略生效（匹配键） | 不参与表达式内的行过滤 |
| 多条匹配的策略 OR 合并 | 不在 `{{...}}` 变量中出现 |

### 2.3 空 roles

客户端不传 `X-MC-User-Roles` → 只匹配 `role = ""` 的策略。无匹配时默认拒绝。

---

## 3. 架构

```
客户后端（自有用户体系）
    │
    │  Authorization: Bearer mc_pat_xxx
    │  X-MC-User-ID: customer_user_123
    │  X-MC-User-Name: zhangsan
    │  X-MC-User-Roles: admin, manager          ← 逗号分隔
    │
    ▼
ModelCraft 数据 API
    │
    │  1. PAT 验证（复用现有 ChiPATAuthMiddleware）
    │  2. RLSContextMiddleware：提取 Header → context
    │  3. 匹配策略：action + role IN (roles)
    │  4. 表达式编译：{{user_id}} → "customer_user_123"
    │  5. 注入 SQL WHERE / CHECK
    │
    ▼
过滤后的数据 → 返回客户后端
```

### 中间件链

```
Request
  ↓
ApiTokenToIdentityMiddleware    ← 已有，验证 mc_pat_xxx
  ↓
RLSContextMiddleware             ← 新增，提取 X-MC-User-* → context
  ↓
ModelRuntimeHandler             ← 已有，执行 GraphQL + RLS 过滤
```

---

## 4. RLS Context Header

客户后端通过 HTTP Header 传递当前用户信息。

| Header | 类型 | 示例 |
|--------|------|------|
| `X-MC-User-ID` | string | `customer_user_123` |
| `X-MC-User-Name` | string | `zhangsan` |
| `X-MC-User-Roles` | comma-separated list | `admin, manager` |

### 安全边界

Header 值由客户后端控制。ModelCraft 不做校验——只是透传。信任边界在 PAT 层：客户持有 PAT 即拥有 project 数据访问权，Header 的作用是**缩小**范围。

---

## 5. 认证 — 复用现有 PAT

无需新增认证机制。客户后端使用现有 `mc_pat_xxx` 格式的 API Token。

已有组件：
- `end_user_api_tokens` 表
- `ApiTokenToIdentityMiddleware`（PAT 验证 + EndUser 身份注入）
- PAT 管理接口

### 两层身份

| 身份层 | 来源 | 含义 |
|--------|------|------|
| PAT 的 EndUserID | `mc_pat_xxx` → DB | 客户在 ModelCraft 的"服务账号"，决定能访问哪些 project |
| RLS Context | `X-MC-User-*` headers | 客户的终端用户，决定能看/写哪些数据行 |

---

## 6. RLS 策略模型

### 6.1 数据结构

每条策略 = `policyName` + `action` + `role` + 表达式。

| 字段 | 类型 | 说明 |
|------|------|------|
| `policyName` | string | 策略名称，model 内唯一 |
| `action` | enum | `read` / `create` / `update` / `delete` |
| `role` | string | 匹配键，和 Header `X-MC-User-Roles` 比对。空字符串 = 默认策略 |
| `using` | JSON object | 行过滤表达式（read / update / delete） |
| `withCheck` | JSON object | 新行校验表达式（create / update） |

### 6.2 action 与表达式的对应

| action | using | withCheck |
|--------|-------|-----------|
| `read` | ✅ 必须 | — |
| `create` | — | ✅ 必须 |
| `update` | ✅ 必须 | ✅ 必须 |
| `delete` | ✅ 必须 | — |

- **using**：已有行过滤，不满足的**静默忽略**
- **withCheck**：新行校验，不满足的**报错拒绝**

### 6.3 表达式格式 — 对齐现有 Condition DSL

使用和现有 `domain/query/field_conditions.go` 一致的 Prisma 风格 JSON，统一全项目查询语法。

**操作符**：`equals`, `not`, `in`, `gt`, `gte`, `lt`, `lte`, `contains`, `startsWith`, `endsWith`

**逻辑**：`AND`, `OR`, `NOT`

**变量**：`"{{user_id}}"`, `"{{user_name}}"` — 字符串模板占位

**NULL**：`{"equals": null}`

```json
// 简单等式
{"tenant_id": {"equals": "{{user_id}}"}}

// 逻辑组合
{
  "AND": [
    {"tenant_id": {"equals": "{{user_id}}"}},
    {"deleted_at": {"equals": null}},
    {"OR": [
      {"status": {"equals": "active"}},
      {"status": {"equals": "draft"}}
    ]}
  ]
}

// 范围 + 变量
{
  "amount": {"gte": 100, "lte": 1000},
  "owner": {"equals": "{{user_name}}"}
}
```

### 6.4 完整示例

```json
[
  {
    "policyName": "admin_all",
    "action": "read",
    "role": "admin",
    "using": {"equals": 1}
  },
  {
    "policyName": "admin_all",
    "action": "create",
    "role": "admin",
    "withCheck": {"equals": 1}
  },
  {
    "policyName": "manager_own",
    "action": "read",
    "role": "manager",
    "using": {
      "AND": [
        {"tenant_id": {"equals": "{{user_id}}"}},
        {"deleted_at": {"equals": null}}
      ]
    }
  },
  {
    "policyName": "manager_own",
    "action": "create",
    "role": "manager",
    "withCheck": {"tenant_id": {"equals": "{{user_id}}"}}
  },
  {
    "policyName": "manager_own",
    "action": "update",
    "role": "manager",
    "using": {"tenant_id": {"equals": "{{user_id}}"}},
    "withCheck": {"tenant_id": {"equals": "{{user_id}}"}}
  },
  {
    "policyName": "manager_own",
    "action": "delete",
    "role": "manager",
    "using": {"tenant_id": {"equals": "{{user_id}}"}}
  }
]
```

---

## 7. 策略匹配与合并

### 7.1 匹配逻辑

```
请求 action = "read", roles = ["admin", "manager"]
  ↓
查 model 的所有策略 WHERE action = "read"
  ↓
过滤：role IN ("admin", "manager", "")
       → role="" 总是被匹配（默认策略，代表所有角色）
  ↓
收集所有匹配策略的 using 表达式
  ↓
所有 using OR 合并 → 注入 SQL WHERE
```

### 7.2 默认拒绝

- model 开启 RLS + 无匹配策略 → 所有操作拒绝
- model 未开启 RLS → 不检查，正常执行

### 7.3 执行流程

```
请求进入
  ↓
PAT 验证 → EndUser 身份
  ↓
提取 X-MC-User-* headers → RLS context
  ↓
RLS 是否开启？ → 否 → 正常执行
  ↓ 是
查询策略：WHERE model=xxx AND action=xxx AND role IN (roles, "")
  ↓
匹配数 = 0 → 拒绝
  ↓
匹配的 using 表达式 替换变量 → OR 合并 → 注入 SQL WHERE
  ↓
（create/update）withCheck 替换变量 → SQL CHECK
  ↓
执行查询 → 返回过滤数据
```

---

## 8. 表达式编译

### 8.1 变量替换

| 模板 | Header 来源 | 未传 |
|------|-----------|------|
| `{{user_id}}` | `X-MC-User-ID` | `""` |
| `{{user_name}}` | `X-MC-User-Name` | `""` |

替换后通过参数化查询拼入 SQL，防止注入。

### 8.2 与现有 PolicyCompiler 的关系

现有 `internal/app/rls/policy_compiler.go` 已实现 JSON → SQL 编译，支持 `_and/_or/_not/_eq/_neq/_gt/_gte/_lt/_lte/_is_null/_in/_nin`。需要扩展：

1. **对齐操作符命名**：`_eq` → `equals`, `_and` → `AND` 等（统一 Condition DSL）
2. **新增操作符**：`contains` → `LIKE '%x%'`, `startsWith` → `LIKE 'x%'`, `endsWith` → `LIKE '%x'`
3. **变量支持**：`{{user_id}}` / `{{user_name}}` → 从 context 查值 → `?` 参数化
4. **AND/OR 合并**：多条匹配策略的表达式 OR 合并为一条 WHERE

---

## 9. SDK

### 9.1 Python 示例

```python
from modelcraft import Client

client = Client(
    endpoint="https://modelcraft/api/data",
    api_token="mc_pat_xxx",
    org="my-org",
    project="my-project",
)

# 查询 + RLS
orders = client.model("orders").list(
    db="main",
    limit=10,
    user_id="customer_123",
    user_name="zhangsan",
    user_roles="admin,manager",
)

# 创建
client.model("orders").create(
    db="main",
    data={"amount": 100},
    user_id="customer_123",
    user_roles="manager",
)
```

### 9.2 SDK 职责

- 封装 `Authorization: Bearer mc_pat_xxx`
- 自动设置 `X-MC-User-*` headers
- V1 支持 Python

---

## 10. 数据 API 端点

复用现有 Runtime GraphQL 端点：

```
POST /end-user/graphql/org/{orgName}/project/{projectSlug}/db/{db}/model/{model}
```

---

## 11. 不在本次范围

| 排除项 | 说明 |
|--------|------|
| OIDC / 社交登录 | 客户自有用户体系 |
| 自建用户/角色系统 | ModelCraft 不管理用户 |
| EndUser 管理界面 | 客户自行管理 |
| Column-level RLS | V1 仅行级 |
| PERMISSIVE / RESTRICTIVE | role 匹配替代，不需要双层合并 |

---

## 12. 与现有设计的差异

| 维度 | 现有 RBAC (2026-05-06) | 本设计 |
|------|----------------------|--------|
| 用户管理 | ModelCraft 管理 EndUser | 客户自行管理 |
| 策略匹配 | Role → Bundle → Permission | action + role 直接匹配 |
| 策略合并 | 取最宽 rowScope | 匹配上的全部 OR 合并 |
| 表达式 | SELF / ALL 固定 scope | 自由表达式 + 变量 |
| 角色定义 | ModelCraft 定义 role | 客户自定义 role 字符串 |
| 适用场景 | 客户无自有用户体系 | 客户有自有用户体系 |

---

## 13. 实现顺序

1. RLS Context Middleware（Header → context）
2. 表达式编译器（对齐 Condition DSL + 变量替换 + SQL 编译）
3. 策略存储（DB migration + CRUD API）
4. 策略匹配 + 合并引擎
5. 接入 modelruntime 执行链路
6. Python SDK
