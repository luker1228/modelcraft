# Token 业务设计规范

**日期**: 2026-03-30
**状态**: 已批准
**背景**: 将原有无状态双 Token 设计升级为现代化 SaaS 标准方案，支持 Web + CLI/CI/CD 两种客户端。

---

## 1. 整体架构

两套独立凭证体系，BFF 负责 Web auth 全流程，Go Backend 负责验证与业务：

```
┌──────────────────────────────────────────────────────────────────┐
│  Web 浏览器                                                       │
│  Access Token（内存，JWT 1h，只含 userId）                        │
│  Refresh Token（httpOnly Cookie，7天滑动过期）                    │
│                         ↕                                        │
│                 Next.js BFF (src/bff/)                           │
│                 · Casdoor OAuth2 交互                             │
│                 · 签发 Access Token（用共享 JWT_SECRET）           │
│                 · Refresh Token Cookie 生命周期管理               │
│                         ↕  (内网)                                │
│                    Go Backend                                    │
│                 · 验证 Access Token 签名                          │
│                 · 实时查 Casbin 权限 → 注入 Context               │
│                 · Refresh Token DB 存储                           │
│                 · user sync（BFF 登录时调用）                     │
├──────────────────────────────────────────────────────────────────┤
│  CLI / CI/CD                                                     │
│  API Key（mc_ 前缀，直连 Go Backend，DB hash，永久或自定义过期）    │
└──────────────────────────────────────────────────────────────────┘
```

### 职责边界

| 职责 | 负责方 |
|------|--------|
| Casdoor OAuth2 交互 | BFF（Go 完全不知道 Casdoor） |
| Access Token 签发 | BFF |
| Access Token 验证 | Go Backend（用共享 JWT_SECRET） |
| Refresh Token 生成 | BFF |
| Refresh Token Cookie 管理 | BFF |
| Refresh Token DB 存储 | Go Backend |
| user 查/创（sync） | Go Backend（BFF 提供 externalId） |
| 权限查询（Casbin） | Go Backend 实时查，BFF 不感知 |
| API Key 全套逻辑 | Go Backend |

BFF 和 Go Backend 之间**只共享一个秘密**：`JWT_SECRET`（用于签发和验签 Access Token）。

---

### Web 登录流程

```
1. 浏览器 → Casdoor 登录页
2. 用户登录 → Casdoor redirect 到 /auth/callback?code=xxx
3. 回调页 JS → POST /bff/auth/token { code }

4. BFF → Casdoor: 用 code 换 Casdoor token
5. BFF: 从 Casdoor token 提取 { externalId, email, name }
6. BFF → Go: POST /api/users/sync { externalId, email, name }
              Go: 查/创 user → 返回 { userId }

7. BFF: 用 JWT_SECRET 签发 Access Token { userId, exp: now+1h }
8. BFF: 生成 Refresh Token（32字节 CSPRNG → 64位 hex）
9. BFF → Go: POST /api/auth/refresh-tokens { userId, tokenHash: SHA256(token) }
              Go: 写入 refresh_tokens 表

10. BFF: Set-Cookie refresh_token=<token>; HttpOnly; Secure; SameSite=Strict; Max-Age=604800
11. BFF → 浏览器: { accessToken, expiresIn: 3600 }
```

### Access Token 自动续期（用户无感知）

```
Access Token 过期（1h后）
  → 浏览器检测到 401
  → POST /bff/auth/refresh（无 Body，Cookie 自动携带）
      BFF: 从 httpOnly Cookie 读取 Refresh Token
      BFF → Go: POST /api/auth/refresh { tokenHash: SHA256(token) }
              Go: 验证 → 轮换（旧 revoked，签发新 tokenHash 写 DB）
              Go → BFF: { newTokenHash, userId }
      BFF: 用 JWT_SECRET 签发新 Access Token
      BFF: 生成新 Refresh Token hex
      BFF → Go: POST /api/auth/refresh-tokens { userId, tokenHash }
      BFF: 更新 httpOnly Cookie
      BFF → 浏览器: { accessToken, expiresIn: 3600 }
  → 浏览器重试原始请求
  → 用户毫无感知
```

每次轮换新 Refresh Token 的 `expires_at = now + 7天`，实际效果为滑动过期——持续使用的用户永远不会被踢出。只有 7 天内完全未使用才需要重新登录。

### CLI 使用流程

```
用户在 Web 控制台创建 API Key
  → 系统生成 key，明文只展示一次
  → 用户复制到 CI/CD 环境变量
  → CLI 直接调用 Go Backend:
    Authorization: Bearer mc_<key>
```

---

## 2. 数据模型

> 所有表均在 Go Backend 的 MySQL 中，BFF 不直接访问数据库。

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
    id           VARCHAR(36)  PRIMARY KEY,
    user_id      VARCHAR(36)  NOT NULL,
    name         VARCHAR(100) NOT NULL,     -- 用户命名，如 "GitHub Actions"
    key_hash     VARCHAR(64)  NOT NULL,     -- SHA256 hash，不存明文
    key_prefix   VARCHAR(10)  NOT NULL,     -- 完整 key 前 10 位，如 "mc_a1b2c3d4"
    last_used_at DATETIME     NULL,         -- 上次使用时间（防抖：距上次 > 1 分钟才更新）
    expires_at   DATETIME     NULL,         -- NULL = 永不过期
    created_at   DATETIME     NOT NULL,
    revoked_at   DATETIME     NULL,         -- NULL = 有效

    INDEX idx_key_hash (key_hash),
    INDEX idx_user_id (user_id)
);
```

### security_audit_logs 表

```sql
CREATE TABLE security_audit_logs (
    id         VARCHAR(36) PRIMARY KEY,
    user_id    VARCHAR(36) NOT NULL,
    event      VARCHAR(50) NOT NULL,   -- 如 REUSE_DETECTED
    detail     JSON        NULL,        -- token_id, ip 等上下文
    created_at DATETIME    NOT NULL,
    INDEX idx_user_id_created (user_id, created_at)
);
```

---

## 3. API 端点

### BFF 端点（Next.js，供浏览器调用）

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/bff/auth/token` | 授权码 → 换 token，设置 httpOnly Cookie |
| `POST` | `/bff/auth/refresh` | 从 Cookie 读 Refresh Token，返回新 Access Token |
| `POST` | `/bff/auth/logout` | 清除 Cookie，通知 Go 吊销 Refresh Token |

#### POST /bff/auth/token

**请求：**
```json
{ "code": "casdoor_authorization_code", "redirectUri": "https://app.example.com/callback" }
```

**响应（成功）：**
```json
{ "accessToken": "eyJhbGci...", "expiresIn": 3600 }
```
同时设置：
```
Set-Cookie: refresh_token=f7a2bc9d...; HttpOnly; Secure; SameSite=Strict; Max-Age=604800; Path=/bff/auth
```

**响应（失败）：**
- HTTP 400：授权码无效、已过期、redirectUri 不匹配
- HTTP 502：Casdoor 或 Go Backend 不可用

#### POST /bff/auth/refresh

**请求：** 无 Body，Refresh Token 自动从 Cookie 读取

**响应（成功）：**
```json
{ "accessToken": "eyJhbGci...", "expiresIn": 3600 }
```
同时更新 httpOnly Cookie（新 Refresh Token）。

**响应（失败）：** HTTP 401，清除 Cookie

#### POST /bff/auth/logout

**请求：** 无 Body

**响应：** HTTP 204，清除 httpOnly Cookie

---

### Go Backend 内部端点（供 BFF 调用，不对外暴露）

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/users/sync` | 登录时同步用户，返回 userId |
| `POST` | `/api/auth/refresh-tokens` | 存储新 Refresh Token hash |
| `POST` | `/api/auth/refresh` | 验证并轮换 Refresh Token |
| `POST` | `/api/auth/logout` | 吊销指定 Refresh Token |

#### POST /api/users/sync

**请求：**
```json
{ "externalId": "casdoor_user_id", "email": "user@example.com", "name": "Alice" }
```

**响应：**
```json
{ "userId": "uuid-xxx" }
```

#### POST /api/auth/refresh-tokens

**请求：**
```json
{ "userId": "uuid-xxx", "tokenHash": "sha256hex...", "expiresAt": "2026-04-06T10:00:00Z" }
```

**响应：** HTTP 204

#### POST /api/auth/refresh

**请求：**
```json
{ "tokenHash": "sha256hex..." }
```

**响应（成功）：**
```json
{ "userId": "uuid-xxx" }
```
（Go 负责轮换旧 token，BFF 随后生成新 token 并单独调用 `/api/auth/refresh-tokens` 存储）

**响应（失败）：** HTTP 401

#### POST /api/auth/logout

**请求：**
```json
{ "tokenHash": "sha256hex..." }
```

**响应：** HTTP 204

---

### GraphQL（API Key 管理，Org 域，Go Backend）

所有操作基于当前认证用户的 `user_id`（从 Context 提取）。每用户最多 **20 个**有效 Key（已吊销不计入）。

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
  keyPrefix:   String!       # 完整 key 前 10 位，如 "mc_a1b2c3d4"
  lastUsedAt:  Time
  expiresAt:   Time          # null = 永不过期
  revokedAt:   Time          # null = 有效；非 null = 已吊销
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
  expiresAt: Time
}

input UpdateApiKeyInput {
  name:      String
  expiresAt: Time            # 显式传 null = 改为永久有效
}

type CreateApiKeyPayload {
  result: CreateApiKeyResult
  error:  CreateApiKeyError
}

type RevokeApiKeyPayload {
  apiKey: ApiKey
  error:  RevokeApiKeyError
}

type UpdateApiKeyPayload {
  apiKey: ApiKey
  error:  UpdateApiKeyError
}

union CreateApiKeyError = ApiKeyLimitExceeded | InvalidInput
union RevokeApiKeyError = ApiKeyNotFound
# 对已吊销的 key 再次吊销为幂等操作，返回成功
union UpdateApiKeyError = ApiKeyNotFound | InvalidInput
```

---

## 4. 安全规范

### Token 参数

| 项目 | 规范 |
|------|------|
| Access Token 有效期 | 1 小时 |
| Access Token 存储（客户端） | 浏览器内存（不写 localStorage，防 XSS） |
| Access Token payload | `{ userId, iat, exp }`，不含权限信息 |
| Refresh Token 有效期 | 7 天（滑动过期） |
| Refresh Token 存储（客户端） | httpOnly + Secure + SameSite=Strict Cookie，Path=/bff/auth |
| Refresh Token 存储（服务端） | MySQL hash，不存明文 |
| Refresh Token 格式 | 32 字节 CSPRNG → hex 编码 → 64 位字符串 |
| Refresh Token 策略 | 轮换：每次 refresh 旧 token 立即 revoked，生成新 token |
| JWT 签名 | HMAC-SHA256，密钥 `JWT_SECRET`（环境变量，最低 32 字节，BFF 和 Go 共享） |
| API Key 格式 | `mc_` + 40位 Base62（CSPRNG，约 238 bit 熵） |
| API Key 前缀存储 | 完整 key 前 10 位（含 `mc_`），用于 UI 识别 |

### Refresh Token 疑似盗用处理

已轮换的 Refresh Token 被再次使用是盗用强信号：

```
BFF → Go: POST /api/auth/refresh（携带已 revoked 的 tokenHash）
  Go: 检测到 revoked_at IS NOT NULL
  Go: 吊销该 user_id 下所有 Refresh Token（全设备强制下线）
  Go: 写入审计日志（event=REUSE_DETECTED）
  Go → BFF: HTTP 401
  BFF: 清除 httpOnly Cookie
  浏览器: 跳转登录页
```

### 并发 Refresh 竞态处理

- BFF 使用单例 Promise 防并发（现有实现保持不变）
- Go 后端不设宽限窗口，旧 token 轮换后立即无效

### Go Backend 统一验证中间件

```
Authorization: Bearer <value>
  ↓
以 "mc_" 开头？
  ├─ YES → API Key 路径
  │         SHA256(value) → 查 api_keys 表
  │         检查 revoked_at IS NULL + expires_at
  │         异步更新 last_used_at（防抖 1 分钟，失败静默忽略）
  │         → 注入 userId → 查 Casbin 权限 → 注入 Context
  │
  └─ NO  → JWT 路径
            验证 HMAC-SHA256 签名（JWT_SECRET）
            验证 exp、iss
            → 注入 userId → 查 Casbin 权限 → 注入 Context
```

---

## 5. 数据清理

**Go Backend 后台 goroutine**（`bizutils.GoWithCtx`），每天执行一次：

```sql
-- refresh_tokens：过期或吊销超过 30 天
DELETE FROM refresh_tokens
WHERE (expires_at < NOW() - INTERVAL 30 DAY)
   OR (revoked_at IS NOT NULL AND revoked_at < NOW() - INTERVAL 30 DAY);

-- api_keys：吊销超过 90 天（保留审计）
DELETE FROM api_keys
WHERE revoked_at IS NOT NULL AND revoked_at < NOW() - INTERVAL 90 DAY;
```

---

## 6. 与现有系统的衔接

### Go Backend 变更

| 项目 | 变更内容 |
|------|---------|
| Casdoor 集成 | **完全移除**，Go 不再知道 Casdoor 存在 |
| Access Token 签发 | **移至 BFF**，Go 只验证 |
| Refresh Token | 从无状态 JWT 改为 DB hash 存储，格式改为不透明 hex |
| 新增内部端点 | `/api/users/sync`、`/api/auth/refresh-tokens`、`/api/auth/refresh`、`/api/auth/logout` |

### Next.js BFF 变更

| 项目 | 变更内容 |
|------|---------|
| Casdoor OAuth2 | 保留，集中在 BFF |
| `/bff/auth/token` | 新增：完整登录流程（sync user → 签发 token） |
| `/bff/auth/refresh` | 新增：从 Cookie 读 token，转发 Go，更新 Cookie |
| `/bff/auth/logout` | 新增：清除 Cookie + 通知 Go 吊销 |
| Access Token 存储 | localStorage → 内存 |
| Refresh Token 存储 | localStorage → httpOnly Cookie（JS 不可见） |

### 保持不变

| 项目 | 说明 |
|------|------|
| Access Token 结构 | JWT，payload 只含 userId |
| Casbin 权限逻辑 | Go 中间件实时查，完全不变 |
| Go 中间件注入 userId 到 Context | 接口不变，业务层无感知 |
| BFF 并发 refresh 防护 | 单例 Promise 保持不变 |
| API Key 直连 Go Backend | CLI 不经过 BFF |

### 新增

| 项目 | 说明 |
|------|------|
| `refresh_tokens` 表 | Go Backend MySQL |
| `api_keys` 表 | Go Backend MySQL |
| `security_audit_logs` 表 | Go Backend MySQL |
| API Key CRUD | Org GraphQL，1 Query + 3 Mutation |
| API Key 验证路径 | Go 中间件新增 `mc_` 识别逻辑 |
| 数据清理 goroutine | Go Backend 后台定时 |

---

## 7. 不在本次范围内

- `logout-all`（吊销所有设备）：后续迭代添加
- API Key 权限 scope：当前等同用户完整权限
- OAuth2 第三方授权：不在范围
- BFF 迁移为独立服务：后续架构演进
