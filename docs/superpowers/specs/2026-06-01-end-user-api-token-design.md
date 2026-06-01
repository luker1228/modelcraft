# End-User API Token (PAT) Design

**Date:** 2026-06-01  
**Status:** Approved  
**Scope:** End-user Personal Access Token — Org 级，前后端完整实现

---

## 1. 概述

为 End-User 提供长期有效的 API Token（Personal Access Token，PAT），使 CLI / 外部程序无需每次用户名密码登录，直接携带 token 调用 ModelCraft API。

**方案选型：** 方案 A — PAT 在后端 middleware 层验证，不改 APISIX Gateway 配置。

---

## 2. Token 格式与存储

### 2.1 Token 格式

```
mc_pat_<32位随机hex>
示例：mc_pat_a3f9b2c1d4e5f6789012345678901234
```

- 前缀 `mc_pat_` 用于 middleware 快速识别，与 JWT Bearer token 区分
- 明文仅在创建时展示一次，不落库
- 落库存 SHA-256 hash（hex 编码，64字符）

### 2.2 数据库 Schema

表名：`end_user_api_tokens`（在系统 DB，不在 tenant DB）

```sql
CREATE TABLE end_user_api_tokens (
  id            VARCHAR(36)   NOT NULL PRIMARY KEY,  -- UUID v7
  org_name      VARCHAR(255)  NOT NULL,
  end_user_id   VARCHAR(36)   NOT NULL,
  name          VARCHAR(255)  NOT NULL,              -- 用户自定义名称
  token_hash    VARCHAR(64)   NOT NULL UNIQUE,       -- SHA-256(plaintext) hex
  expires_at    DATETIME      NULL,                  -- NULL = 永不过期
  last_used_at  DATETIME      NULL,                  -- 每次验证时更新
  created_at    DATETIME      NOT NULL,
  deleted_at    DATETIME      NULL,                  -- 软删除时间
  delete_token  VARCHAR(36)   NOT NULL DEFAULT ''    -- 软删除唯一标记
);

-- 唯一索引：同一用户不能有同名 token（软删除后可复用）
CREATE UNIQUE INDEX uq_end_user_api_tokens_name
  ON end_user_api_tokens (org_name, end_user_id, name, delete_token);

-- 查询索引
CREATE INDEX idx_end_user_api_tokens_user
  ON end_user_api_tokens (org_name, end_user_id, deleted_at);
```

---

## 3. 后端实现

### 3.1 领域层

**文件：** `internal/domain/enduser/api_token.go`

```go
type APIToken struct {
    ID          string
    OrgName     string
    EndUserID   string
    Name        string
    TokenHash   string
    ExpiresAt   *time.Time
    LastUsedAt  *time.Time
    CreatedAt   time.Time
    DeletedAt   *time.Time
    DeleteToken string
}

func (t *APIToken) IsValid() bool {
    if t.DeletedAt != nil { return false }
    if t.ExpiresAt != nil && time.Now().After(*t.ExpiresAt) { return false }
    return true
}
```

**Repository 接口：** `internal/domain/enduser/end_user_repository.go` 新增

```go
type APITokenRepository interface {
    Save(ctx context.Context, token *APIToken) error
    FindByHash(ctx context.Context, hash string) (*APIToken, error)
    ListByUser(ctx context.Context, orgName, endUserID string) ([]*APIToken, error)
    SoftDelete(ctx context.Context, id, orgName, endUserID string) error
    UpdateLastUsed(ctx context.Context, id string, at time.Time) error
}
```

### 3.2 应用层

**文件：** `internal/app/enduser/api_token_service.go`

命令与结果类型：

```go
type CreateAPITokenCommand struct {
    OrgName     string
    EndUserID   string
    Name        string
    ExpiresAt   *time.Time
}

type CreateAPITokenResult struct {
    Token     *domain.APIToken
    Plaintext string  // 仅此一次，不落库
}
```

操作：
- `CreateAPIToken(ctx, cmd)` — 生成随机 token，存 hash，返回明文
- `ListAPITokens(ctx, orgName, endUserID)` — 列出未删除 token
- `RevokeAPIToken(ctx, id, orgName, endUserID)` — 软删除

### 3.3 验证 Middleware

**文件：** `internal/middleware/chi_pat_auth.go`

```go
// ChiPATAuthMiddleware 识别 "mc_pat_" 前缀的 Bearer token，
// 验证通过后将 EndUser 身份注入 context（与 JWT 路径等价）。
// 若 token 不是 mc_pat_ 前缀，直接调用 next（不拦截 JWT 流程）。
func ChiPATAuthMiddleware(repo domain.APITokenRepository) func(http.Handler) http.Handler
```

验证流程：
1. 读取 `Authorization: Bearer mc_pat_xxx`
2. 识别 `mc_pat_` 前缀，否则 `next()`
3. SHA-256 hash → 查 DB
4. 检查 `IsValid()`（未删除、未过期）
5. 异步更新 `last_used_at`
6. 注入 `ctxutils.WithEndUserID`、`ctxutils.WithOrgName` 到 context

**挂载位置：** 在 `chi_setup.go` 的条件认证 middleware 中，与 JWT 验证并联（先检查 PAT 前缀，再走 JWT）。

### 3.4 GraphQL API

**Schema 文件：** `api/graph/org/schema/end_user_api_token.graphql`

```graphql
type EndUserAPIToken {
  id:         ID!
  name:       String!
  createdAt:  String!
  expiresAt:  String      # null = 永不过期
  lastUsedAt: String      # null = 从未使用
}

type CreateEndUserAPITokenPayload {
  token:     EndUserAPIToken!
  plaintext: String!       # 仅创建时返回一次
}

extend type Query {
  endUserAPITokens: [EndUserAPIToken!]!
}

extend type Mutation {
  createEndUserAPIToken(name: String!, expiresAt: String): CreateEndUserAPITokenPayload!
  revokeEndUserAPIToken(id: ID!): Boolean!
}
```

**Endpoint：** Org-scoped GraphQL（`/graphql/org/{orgName}/`）  
**认证：** EndUser Bearer token（现有 JWT）调用管理接口；PAT 本身不能管理 PAT

---

## 4. 前端实现

### 4.1 路由

```
/end-user/{orgName}/dashboard/tokens   ← 新增页面
```

**文件结构：**
```
src/app/end-user/[orgName]/dashboard/tokens/
├── page.tsx         # Token 管理页面主体
└── _components/
    ├── TokenTable.tsx          # token 列表表格
    ├── CreateTokenDialog.tsx   # 新建 token 弹窗
    └── TokenRevealDialog.tsx   # 创建后一次性展示弹窗
```

### 4.2 导航

`EndUserAppLayout.tsx` 侧边栏新增导航项：

```typescript
type ActivePage = 'projects' | 'cli' | 'tokens'  // 新增 'tokens'

// 新增导航按钮：
// 图标：KeyRound（lucide-react）
// 文字：API Token
// 路由：/end-user/{orgName}/dashboard/tokens
// 激活条件：activePage === 'tokens'
```

### 4.3 页面布局

```
┌─────────────────────────────────────────────────┐
│ TopBar: ModelCraft > orgName        [用户菜单]   │
├──────────┬──────────────────────────────────────┤
│ 项目     │  API Token                           │
│ CLI 下载 │  ─────────────────────────────────── │
│▶API Token│  [+ 新建 Token]              右对齐   │
│          │                                      │
│          │  ┌──────────────────────────────┐   │
│          │  │名称    过期时间  最后使用  操作│   │
│          │  │my-cli  永不过期  2小时前  [撤销]│  │
│          │  │ci-bot  2026-12  从未使用  [撤销]│  │
│          │  └──────────────────────────────┘   │
│          │                                      │
│          │  (空状态：暂无 API Token，点击新建)    │
└──────────┴──────────────────────────────────────┘
```

### 4.4 新建 Token 弹窗

字段：
- **名称**（必填，文本输入）：同一用户不可重名
- **过期时间**（单选）：永不过期 / 30天 / 90天 / 自定义日期

### 4.5 Token 一次性展示弹窗

创建成功后立即弹出，用户必须主动点击"我已复制"才能关闭：

```
┌──────────────────────────────────────┐
│ ⚠️ 请立即复制，此后无法再次查看       │
│                                      │
│ [mc_pat_a3f9b2c1...]  [复制到剪贴板] │
│                                      │
│  复制后点击确认关闭此弹窗             │
│                         [我已复制]   │
└──────────────────────────────────────┘
```

### 4.6 GraphQL 文档

**文件：** `src/api-client/end-user/graphql-docs/api-tokens.ts`

包含：
- `END_USER_API_TOKENS` query
- `CREATE_END_USER_API_TOKEN` mutation
- `REVOKE_END_USER_API_TOKEN` mutation

调用 `createEndUserOrgScopedClient(orgName)` 发请求。

---

## 5. 安全约束

1. **明文不落库**：生成后立即 hash，明文仅在内存中传递到 response
2. **Token 不可管理 Token**：创建/撤销接口只接受 JWT 认证，PAT 无法调用管理接口
3. **软删除**：撤销使用软删除 + `delete_token`，保留审计记录
4. **last_used_at 异步更新**：不阻塞请求路径
5. **Token 数量限制**：每个 EndUser 最多 20 个有效 token（防止滥用）

---

## 6. 实现顺序

1. DB migration（Atlas）
2. 领域层 + Repository
3. 应用服务层
4. PAT middleware
5. GraphQL schema + codegen + resolver
6. 前端 GraphQL 文档
7. 前端 Token 页面（TokenTable + CreateTokenDialog + TokenRevealDialog）
8. 侧边栏导航项

---

## 7. 不在本期范围

- Token 权限范围细分（本期 token 拥有用户的全部权限）
- Token 使用审计日志
- Token 使用量统计
- Webhook 事件通知
