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

## 5. RLS 策略配置 — 对齐 PostgreSQL RLS

ModelCraft 的 RLS 模型直接对应 PostgreSQL 的 Row-Level Security 语义。用过 PG RLS 的开发者无需重新学习。

### 5.1 概念映射

| PostgreSQL | ModelCraft |
|---|---|
| `ALTER TABLE ... ENABLE ROW LEVEL SECURITY` | Model 级别的 RLS 开关 |
| `CREATE POLICY name ON table` | 客户在低代码后台为 model 创建策略 |
| `USING (expression)` | 读过滤：决定哪些**已有行**可见（SELECT / UPDATE / DELETE），不符合的行**静默过滤** |
| `WITH CHECK (expression)` | 写校验：决定哪些**新行**允许写入（INSERT / UPDATE），不符合则**报错拒绝** |
| `AS PERMISSIVE`（默认） | 多条策略 OR 合并，任一通过即可 |
| `AS RESTRICTIVE` | 多条策略 AND 合并，全部必须通过 |
| 默认拒绝（无策略 = 不可访问） | 开启 RLS 但无策略 → 所有操作被拒绝 |
| `current_setting('app.user_id')` | `{{user_id}}` 变量 |

### 5.2 USING vs WITH CHECK（关键语义）

这是 PG RLS 的核心设计，ModelCraft 完全保留：

| | USING | WITH CHECK |
|---|---|---|
| **作用** | 过滤已有行 | 校验新行 |
| **SELECT** | ✅ 可见行过滤 | ❌ 不适用 |
| **INSERT** | ❌ 不适用 | ✅ 新行校验 |
| **UPDATE** | ✅ 哪些行可被更新 | ✅ 更新后的行是否合法 |
| **DELETE** | ✅ 哪些行可被删除 | ❌ 不适用 |
| **不满足时** | 静默忽略该行 | 抛错，拒绝操作 |
| **省略时** | — | 复用 USING 表达式 |

### 5.3 策略定义示例

```sql
-- 这是客户在 ModelCraft 低代码后台为 orders model 配置的 RLS 策略
-- 语义完全等同于 PostgreSQL CREATE POLICY

-- 策略 1：用户只能看/改/删自己的订单（PERMISSIVE）
CREATE POLICY "user_own_orders" ON orders
    AS PERMISSIVE
    FOR ALL
    USING (tenant_id = '{{user_id}}');

-- 策略 2：管理员可以看所有订单（PERMISSIVE，与策略 1 OR 合并）
CREATE POLICY "admin_all_orders" ON orders
    AS PERMISSIVE
    FOR SELECT
    USING ('{{user_role}}' = 'admin');

-- 策略 3：无论如何，已删除的订单不可见（RESTRICTIVE，与所有 PERMISSIVE AND 合并）
CREATE POLICY "hide_deleted" ON orders
    AS RESTRICTIVE
    FOR ALL
    USING (deleted_at IS NULL);
```

### 5.4 策略合并逻辑

```
PERMISSIVE 策略之间 → OR
    policy_user_own_orders OR policy_admin_all_orders

RESTRICTIVE 策略之间 → AND
    (所有 RESTRICTIVE 都通过)

最终结果 → (任一 PERMISSIVE 通过) AND (所有 RESTRICTIVE 通过)
```

### 5.5 默认拒绝

当 model 开启 RLS 但没有配置任何策略时，所有操作被拒绝。这是 PG 的安全默认。

### 5.6 变量替换

RLS 引擎执行前，将模板变量替换为 Header 实际值：

| 模板变量 | Header 来源 | 未传时的行为 |
|---------|------------|-------------|
| `{{user_id}}` | `X-MC-User-ID` | 空字符串 `''` |
| `{{user_name}}` | `X-MC-User-Name` | 空字符串 `''` |
| `{{user_role}}` | `X-MC-User-Role` | 空字符串 `''` |

替换后的表达式直接拼入 SQL WHERE 子句。

### 5.7 策略执行流程

```
请求进入
  ↓
RLS 是否开启？ → 否 → 跳过，正常执行
  ↓ 是
提取 RLS context（从 Header）
  ↓
收集 model 的所有 PERMISSIVE 策略 → 替换变量 → OR 合并
  ↓
收集 model 的所有 RESTRICTIVE 策略 → 替换变量 → AND 合并
  ↓
USING 表达式 → 注入 SQL WHERE（静默过滤）
  ↓
WITH CHECK 表达式 → 注入 SQL CHECK（不满足则报错）
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
