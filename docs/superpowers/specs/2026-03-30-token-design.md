# Token 业务设计规范

**日期**: 2026-03-30
**状态**: 已批准
**背景**: 将原有无状态双 Token 设计升级为现代化 SaaS 标准方案，支持 Web + CLI/CI/CD 两种客户端。

---

## 1. 整体架构

两套独立凭证体系，共享统一验证中间件：

```
┌─────────────────────────────────────────────────────────┐
│  Web 浏览器                                              │
│  Access Token (JWT 1h) + Refresh Token (DB 7天, 轮换)   │
├─────────────────────────────────────────────────────────┤
│  CLI / CI/CD                                            │
│  API Key (mc_xxxxx, DB hash, 永久或自定义过期)            │
├─────────────────────────────────────────────────────────┤
│  统一中间件                                              │
│  识别类型 → 验证 → 注入 userId → 业务层无感知             │
└─────────────────────────────────────────────────────────┘
```

### Web 登录流程

```
Casdoor 授权码
  → /api/auth/token
  → 验证授权码，提取 external_id
  → 查/创建本地 user 记录
  → 签发 Access Token (JWT, 1h)
  → 签发 Refresh Token (写入 DB, 7天)
  → 返回两个 token 给前端
```

### Access Token 自动续期（用户无感知）

```
Access Token 过期 (1h后)
  → 前端检测到 401 响应
  → 自动发送 Refresh Token 到 /api/auth/refresh
  → 服务端返回新 Access Token + 新 Refresh Token（轮换）
  → 前端重试原始请求
  → 用户毫无感知
```

只有 Refresh Token 过期（7天未使用）才需要重新登录。

### CLI 使用流程

```
用户在 Web 控制台创建 API Key
  → 系统生成 key，明文只展示一次
  → 用户复制到 CI/CD 环境变量
  → CLI 每次请求携带: Authorization: Bearer mc_<key>
```

---

## 2. 数据模型

### refresh_tokens 表

```sql
CREATE TABLE refresh_tokens (
    id          VARCHAR(36) PRIMARY KEY,  -- UUID
    user_id     VARCHAR(36) NOT NULL,
    token_hash  VARCHAR(64) NOT NULL,     -- SHA256 hash，不存明文
    expires_at  DATETIME   NOT NULL,
    created_at  DATETIME   NOT NULL,
    revoked_at  DATETIME   NULL,          -- NULL = 有效

    INDEX idx_token_hash (token_hash),
    INDEX idx_user_id (user_id)
);
```

### api_keys 表

```sql
CREATE TABLE api_keys (
    id           VARCHAR(36)  PRIMARY KEY,  -- UUID
    user_id      VARCHAR(36)  NOT NULL,
    name         VARCHAR(100) NOT NULL,     -- 用户命名，如 "GitHub Actions"
    key_hash     VARCHAR(64)  NOT NULL,     -- SHA256 hash，不存明文
    key_prefix   VARCHAR(8)   NOT NULL,     -- 明文前8位，如 "mc_a1b2"，用于 UI 识别
    last_used_at DATETIME     NULL,         -- 上次使用时间
    expires_at   DATETIME     NULL,         -- NULL = 永不过期
    created_at   DATETIME     NOT NULL,
    revoked_at   DATETIME     NULL,         -- NULL = 有效

    INDEX idx_key_hash (key_hash),
    INDEX idx_user_id (user_id)
);
```

**设计决策：**

| 决策 | 原因 |
|------|------|
| token/key 只存 hash | 数据库泄露不会暴露真实凭证 |
| `key_prefix` 存明文 | UI 显示 `mc_a1b2****` 让用户辨认 |
| `revoked_at` 而非删除 | 保留吊销历史，支持审计 |
| `last_used_at` | 发现长期未使用 key，提醒用户清理 |

---

## 3. API 端点

### REST（Web Auth）

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/auth/token` | Casdoor 授权码换取 Access Token + Refresh Token |
| `POST` | `/api/auth/refresh` | Refresh Token 换取新的 Access Token（含轮换） |
| `POST` | `/api/auth/logout` | 吊销当前 Refresh Token |

### GraphQL（API Key 管理，Org 域）

```graphql
type Query {
  apiKeys: [ApiKey!]!
}

type Mutation {
  createApiKey(input: CreateApiKeyInput!): CreateApiKeyPayload!
  revokeApiKey(id: ID!): RevokeApiKeyPayload!
  updateApiKey(id: ID!, input: UpdateApiKeyInput!): UpdateApiKeyPayload!
}

type ApiKey {
  id:          ID!
  name:        String!
  keyPrefix:   String!       # "mc_a1b2"
  lastUsedAt:  Time
  expiresAt:   Time          # null = 永不过期
  createdAt:   Time!
}

type CreateApiKeyResult {
  id:        ID!
  name:      String!
  key:       String!         # 完整明文，只出现一次
  keyPrefix: String!
  createdAt: Time!
}

input CreateApiKeyInput {
  name:      String!
  expiresAt: Time            # 可选
}

input UpdateApiKeyInput {
  name:      String
  expiresAt: Time
}

type CreateApiKeyPayload {
  result: CreateApiKeyResult
  error:  CreateApiKeyError
}

type RevokeApiKeyPayload {
  success: Boolean
  error:   RevokeApiKeyError
}

type UpdateApiKeyPayload {
  apiKey: ApiKey
  error:  UpdateApiKeyError
}
```

---

## 4. 安全规范

### Token 参数

| 项目 | 规范 |
|------|------|
| Access Token 有效期 | 1 小时 |
| Refresh Token 有效期 | 7 天 |
| Refresh Token 策略 | 轮换：每次 refresh 生成新 token，旧的立即 revoked |
| 存储方式 | 只存 SHA256 hash，明文不落库 |
| API Key 前缀 | `mc_` 固定前缀，中间件快速识别 |
| API Key 有效期 | 默认永久，可选设置过期时间 |

### Refresh Token 轮换逻辑

```
客户端发送旧 Refresh Token
  → 服务端验证通过
  → 旧 token: revoked_at = now()   (立即失效)
  → 新 token: 写入 DB
  → 返回新 token 给客户端
  → 攻击者若用旧 token → 直接拒绝
```

### 统一验证中间件

```
Authorization: Bearer <value>
  ↓
以 "mc_" 开头？
  ├─ YES → API Key 路径
  │         SHA256(value) → 查 api_keys 表
  │         检查 revoked_at IS NULL
  │         检查 expires_at（如有）
  │         更新 last_used_at
  │         → 注入 userId 到 Context
  │
  └─ NO  → JWT 路径
            验证签名（HMAC-SHA256）
            验证 exp、iss
            → 注入 userId 到 Context
```

---

## 5. 与现有系统的衔接

### 需要变更的部分

| 项目 | 变更内容 |
|------|---------|
| Refresh Token | 从无状态 JWT 改为写入 `refresh_tokens` DB 表 |
| `/api/auth/refresh` | 验证后执行轮换逻辑，旧 token 立即失效 |
| `/api/auth/logout` | 新增：将当前 Refresh Token 标记 revoked |
| Casdoor JWT 兼容层 | 移除：Casdoor JWT 只在 `/api/auth/token` 内部验证，不再对外暴露 |

### 保持不变的部分

| 项目 | 说明 |
|------|------|
| Access Token 结构 | JWT + HMAC-SHA256，payload 只含 userId |
| Casdoor 集成流程 | 授权码 → 验证 → 查/创 user，逻辑不变 |
| 中间件注入 userId 到 Context | 接口不变，业务层无感知 |

### 新增的部分

| 项目 | 说明 |
|------|------|
| `refresh_tokens` 表 | 新增数据库表 |
| `api_keys` 表 | 新增数据库表 |
| API Key CRUD | Org GraphQL 新增 Query + 3 个 Mutation |
| API Key 验证路径 | 中间件新增 `mc_` 前缀识别逻辑 |

---

## 6. 不在本次范围内

- `logout-all`（吊销所有设备）：可在后续迭代添加
- API Key 权限范围（scope）：当前所有 key 等同于用户权限
- OAuth2 第三方授权：不在本次设计范围
