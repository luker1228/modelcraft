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

## 5. RLS 策略配置

客户在低代码后台为每个 model 配置 RLS 规则。

### 5.1 规则定义

```yaml
# 示例：orders model 的 RLS 策略
model: orders
rules:
  select:
    filter: "tenant_id = '{{user_id}}'"          # 只能看自己的数据
  insert:
    force:                                        # 写入时强制注入
      tenant_id: "{{user_id}}"
  update:
    filter: "tenant_id = '{{user_id}}'"
  delete:
    filter: "tenant_id = '{{user_id}}' AND {{user_role}} = 'admin'"
```

### 5.2 变量替换

RLS 引擎在执行查询前，将 `{{user_id}}` / `{{user_name}}` / `{{user_role}}` 替换为 Header 中的实际值。未传的 Header 对应的变量保持为空字符串。

### 5.3 策略执行

```
请求进入
  ↓
提取 RLS context（从 Header）
  ↓
查询 model 的 RLS 策略
  ↓
替换策略中的变量 → 生成实际 WHERE 条件
  ↓
将 WHERE 条件注入 SQL 查询
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
| 权限模型 | Role → Bundle → Permission | RLS 表达式 + Header 变量 |
| 行过滤 | SELF / ALL（固定 scope） | 自定义 WHERE 表达式 |
| 适用场景 | 客户无自有用户体系 | 客户有自有用户体系 |
| 用户标识 | EndUserID（ModelCraft 内部 ID） | Header 透传（客户定义） |

---

## 10. 实现顺序

1. RLS Context Middleware（提取 Header → context）
2. RLS 策略存储（DB schema + domain + CRUD API）
3. RLS 引擎（策略匹配 + 变量替换 + WHERE 注入）
4. 接入 modelruntime（现有 GraphQL 执行引擎）
5. Python SDK
