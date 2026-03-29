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
│  API Key (mc_ 前缀, DB hash, 永久或自定义过期)            │
├─────────────────────────────────────────────────────────┤
│  统一中间件                                              │
│  识别类型 → 验证 → 注入 userId → 业务层无感知             │
└─────────────────────────────────────────────────────────┘
```

### Web 登录流程

```
Casdoor 授权码
  → POST /api/auth/token
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
  → 自动发送 Refresh Token 到 POST /api/auth/refresh
  → 服务端验证通过 → 旧 token revoked，签发新 token 对
  → 前端重试原始请求
  → 用户毫无感知
```

只有 Refresh Token 过期才需要重新登录。每次轮换时重新签发新 token，新 token 的 `expires_at = 创建时间 + 7天`，因此实际效果为滑动过期——持续使用的用户永远不会被踢出。

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
    key_prefix   VARCHAR(10)  NOT NULL,     -- 完整 key 的前 10 位，如 "mc_a1b2c3d4"，用于 UI 识别
    last_used_at DATETIME     NULL,         -- 上次使用时间（防抖：仅在距上次 > 1 分钟时更新）
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
| `key_prefix` 存完整 key 的前 10 位（含 `mc_` 前缀） | UI 显示 `mc_a1b2c3d4****` 让用户辨认 |
| `revoked_at` 而非删除 | 保留吊销历史，支持审计 |
| `last_used_at` 防抖更新 | 避免每次请求都写库，1 分钟内只更新一次 |

---

## 3. API 端点

### REST（Web Auth）

#### POST /api/auth/token

**请求：**
```json
{
  "code": "casdoor_authorization_code",
  "redirectUri": "https://app.example.com/callback"
}
```

**响应（成功）：**
```json
{
  "accessToken": "eyJhbGci...",
  "refreshToken": "f7a2bc9d...（64位随机 hex，不透明随机串，非 JWT）",
  "expiresIn": 3600
}
```

Refresh Token 格式：32 字节 CSPRNG → hex 编码 → 64 位字符串，存储时取 SHA256 hash。

**响应（失败）：**
- HTTP 400：授权码无效、已过期、`redirectUri` 不匹配
- HTTP 502：Casdoor 服务不可用

#### POST /api/auth/refresh

**请求：**
```json
{
  "refreshToken": "f7a2bc9d...（64位 hex）"
}
```

**响应（成功）：**
```json
{
  "accessToken": "eyJhbGci...",
  "refreshToken": "e8b3cd0e...（新的 64位 hex）",
  "expiresIn": 3600
}
```

**响应（失败，token 已过期、已吊销或不存在）：** HTTP 401

#### POST /api/auth/logout

**请求 Header：**
```
Authorization: Bearer <access_token>
```

**请求 Body：**
```json
{
  "refreshToken": "eyJhbGci..."
}
```

**响应：** HTTP 204 No Content

---

### GraphQL（API Key 管理，Org 域）

所有 API Key 操作均基于当前认证用户的 `user_id`（从 Context 提取），用户只能查询和操作自己的 Key。

每个用户最多创建 **20 个**有效（未吊销）API Key。已吊销的 key 不计入限额。

```graphql
type Query {
  # 返回当前认证用户的所有 API Key
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
  keyPrefix:   String!       # 完整 key 的前 10 位，如 "mc_a1b2c3d4"
  lastUsedAt:  Time
  expiresAt:   Time          # null = 永不过期
  revokedAt:   Time          # null = 有效；非 null = 已吊销
  createdAt:   Time!
}

type CreateApiKeyResult {
  id:        ID!
  name:      String!
  key:       String!         # 完整明文，只出现一次，之后无法找回
  keyPrefix: String!
  createdAt: Time!
}

input CreateApiKeyInput {
  name:      String!
  expiresAt: Time            # 可选，null = 永不过期
}

input UpdateApiKeyInput {
  name:      String
  expiresAt: Time            # 可选；显式传 null = 改为永久有效（移除过期时间）
}

# Payload（遵循项目联合错误模式）
type CreateApiKeyPayload {
  result: CreateApiKeyResult
  error:  CreateApiKeyError
}

type RevokeApiKeyPayload {
  apiKey: ApiKey             # 返回已吊销的 key 记录（含 revokedAt）
  error:  RevokeApiKeyError
}

type UpdateApiKeyPayload {
  apiKey: ApiKey
  error:  UpdateApiKeyError
}

# 错误类型（在 errors.graphql 中定义）
union CreateApiKeyError =
  | ApiKeyLimitExceeded      # 已达到每用户 20 个上限
  | InvalidInput             # name 为空，或 expiresAt 已过期

union RevokeApiKeyError =
  | ApiKeyNotFound           # key 不存在或不属于当前用户
  # 注：对已吊销的 key 再次吊销为幂等操作，返回成功（含已吊销的 key 记录）

union UpdateApiKeyError =
  | ApiKeyNotFound           # key 不存在或不属于当前用户
  | InvalidInput             # 无效的 expiresAt（已过期）
```

---

## 4. 安全规范

### Token 参数

| 项目 | 规范 |
|------|------|
| Access Token 有效期 | 1 小时 |
| Refresh Token 有效期 | 7 天 |
| Refresh Token 策略 | 轮换：每次 refresh 生成新 token 对，旧 token 立即 revoked |
| 存储方式 | 只存 SHA256 hash，明文不落库 |
| API Key 格式 | `mc_` 前缀 + 40位 Base62（CSPRNG 生成，约 238 bit 熵） |
| API Key 前缀 | 完整 key 的前 10 位（含 `mc_`），用于 UI 识别 |
| API Key 有效期 | 默认永久，可选设置过期时间 |

**API Key 示例：**
```
mc_a1B2c3D4e5F6g7H8i9J0k1L2m3N4o5P6q7R8
└──┘└──────────────────────────────────┘
前缀  40位 Base62（CSPRNG）
```

### Refresh Token 轮换逻辑

```
客户端发送 Refresh Token
  → 服务端查询 DB（按 hash）
  → 找到且未 revoked → 正常流程
      旧 token: revoked_at = now()   (立即失效)
      新 token 对: 写入 DB
      → 返回新 token 对给客户端
  → 找到但已 revoked → 安全事件处理（见下方）
  → 未找到或已过期 → HTTP 401
```

### Refresh Token 疑似盗用处理

已轮换的 Refresh Token 被再次使用，是 token 盗用的强信号：

```
客户端发送已 revoked 的 Refresh Token
  → 服务端检测到 revoked_at IS NOT NULL
  → 吊销该 user_id 下所有 Refresh Token（全设备强制下线）
  → 写入审计日志（event=REUSE_DETECTED，见下方审计日志）
  → 返回 HTTP 401
```

### 审计日志

安全事件写入 `security_audit_logs` 表（新增）：

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

当前阶段只写入 `REUSE_DETECTED` 事件，后续可扩展。

### 并发 Refresh 竞态处理

多个标签页同时检测到 401，可能并发发送相同 Refresh Token：

- 前端负责防止并发：使用单例 Promise，同一时刻只发起一个 refresh 请求（现有实现已有此逻辑，保持不变）
- 服务端不设宽限窗口，token 轮换后旧 token 立即无效

### 统一验证中间件

```
Authorization: Bearer <value>
  ↓
以 "mc_" 开头？
  ├─ YES → API Key 路径
  │         SHA256(value) → 查 api_keys 表
  │         检查 revoked_at IS NULL
  │         检查 expires_at（如有）
  │         异步更新 last_used_at（防抖：距上次 > 1 分钟才更新，失败静默忽略）
  │         → 注入 userId 到 Context
  │
  └─ NO  → JWT 路径
            验证签名（HMAC-SHA256，密钥来自环境变量 JWT_SECRET，最低 32 字节）
            验证 exp、iss
            → 注入 userId 到 Context
```

---

## 5. 数据清理

过期和已吊销的记录需定期清理，避免表无限增长。

**清理策略：** 后台 goroutine（使用 `bizutils.GoWithCtx`），每天执行一次。

```sql
-- 清理 refresh_tokens：已过期超过 30 天，或已吊销超过 30 天
DELETE FROM refresh_tokens
WHERE (expires_at < NOW() - INTERVAL 30 DAY)
   OR (revoked_at IS NOT NULL AND revoked_at < NOW() - INTERVAL 30 DAY);

-- 清理 api_keys：已吊销超过 90 天（保留更长用于审计）
DELETE FROM api_keys
WHERE revoked_at IS NOT NULL AND revoked_at < NOW() - INTERVAL 90 DAY;
```

---

## 6. 与现有系统的衔接

### 需要变更的部分

| 项目 | 变更内容 |
|------|---------|
| Refresh Token | 从无状态 JWT 改为写入 `refresh_tokens` DB 表 |
| `/api/auth/refresh` | 验证后执行轮换逻辑 + 盗用检测 |
| `/api/auth/logout` | 新增：将当前 Refresh Token 标记 revoked |
| Casdoor JWT 兼容层 | 移除：Casdoor JWT 只在 `/api/auth/token` 内部验证，不再对外暴露 |

### 保持不变的部分

| 项目 | 说明 |
|------|------|
| Access Token 结构 | JWT + HMAC-SHA256，payload 只含 userId |
| Casdoor 集成流程 | 授权码 → 验证 → 查/创 user，逻辑不变 |
| 中间件注入 userId 到 Context | 接口不变，业务层无感知 |
| 前端并发 refresh 防护 | 单例 Promise 机制保持不变 |

### 新增的部分

| 项目 | 说明 |
|------|------|
| `refresh_tokens` 表 | 新增数据库表 |
| `api_keys` 表 | 新增数据库表 |
| API Key CRUD | Org GraphQL 新增 1 个 Query + 3 个 Mutation |
| API Key 验证路径 | 中间件新增 `mc_` 前缀识别逻辑 |
| 数据清理 goroutine | 后台定时清理过期记录 |

---

## 7. 不在本次范围内

- `logout-all`（吊销所有设备的 Refresh Token）：可在后续迭代添加
- API Key 权限范围（scope）：当前所有 key 等同于用户完整权限
- OAuth2 第三方授权：不在本次设计范围
