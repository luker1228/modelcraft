# Token 业务设计规范

**日期**: 2026-03-30
**状态**: 已批准
**背景**: 将原有无状态双 Token 设计升级为现代化 SaaS 标准方案，支持 Web + CLI/CI/CD 两种客户端。

---

## 1. 整体架构

两套独立凭证体系，共享 Go 后端统一验证中间件：

```
┌─────────────────────────────────────────────────────────────┐
│  Web 浏览器                                                  │
│                                                             │
│  Access Token (内存, JWT 1h)                                │
│  Refresh Token (httpOnly Cookie, 7天, 滑动过期)              │
│                    ↕                                        │
│            Next.js BFF (src/bff/)                           │
│            · token 交换 (Casdoor 授权码 → token 对)          │
│            · token 刷新 (读写 httpOnly Cookie)               │
│            · logout (清除 Cookie + 通知 Go 后端吊销)          │
│                    ↕                                        │
│              Go Backend                                     │
│            · 验证 Access Token JWT 签名                      │
│            · 存储 Refresh Token hash (MySQL)                 │
│            · 提供 /api/auth/* 内部接口给 BFF 调用             │
├─────────────────────────────────────────────────────────────┤
│  CLI / CI/CD                                                │
│  API Key (mc_ 前缀, 直连 Go Backend, DB hash, 永久或自定义过期)│
└─────────────────────────────────────────────────────────────┘
```

### 职责划分

| 职责 | 负责方 |
|------|--------|
| Casdoor 授权码交换 | Next.js BFF |
| Refresh Token 存储（客户端） | httpOnly Cookie（BFF 设置） |
| Refresh Token 存储（服务端） | Go Backend MySQL |
| Refresh Token 轮换 + 盗用检测 | Go Backend |
| Access Token 签发 | Go Backend |
| Access Token 验证 | Go Backend 中间件 |
| API Key 管理和验证 | Go Backend |

### Web 登录流程

```
浏览器 → Casdoor 授权页
  ← 授权码
浏览器 → POST /bff/auth/token (Next.js BFF)
  BFF → POST /api/auth/token (Go Backend)
      Go: 验证授权码 → 查/创 user → 签发 Access Token + Refresh Token
      Go: Refresh Token 写入 DB
  ← { accessToken, refreshToken, expiresIn }
  BFF: 将 refreshToken 写入 httpOnly Cookie
  BFF: 将 accessToken 返回浏览器（存内存）
```

### Access Token 自动续期（用户无感知）

```
Access Token 过期 (1h后)
  → 浏览器检测到 401
  → 请求 POST /bff/auth/refresh (Next.js BFF)
      BFF: 从 httpOnly Cookie 读取 Refresh Token（JS 不可见）
      BFF → POST /api/auth/refresh (Go Backend)
          Go: 验证 → 轮换（旧 token revoked，签发新 token 对）
      ← { accessToken, refreshToken }
      BFF: 更新 httpOnly Cookie（新 Refresh Token）
      BFF: 返回新 Access Token 给浏览器
  → 浏览器重试原始请求
  → 用户毫无感知
```

每次轮换时新 Refresh Token 的 `expires_at = 创建时间 + 7天`，实际效果为滑动过期——持续使用的用户永远不会被踢出。只有 7 天内完全未使用才需要重新登录。

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

> **注意**：`refresh_tokens` 和 `security_audit_logs` 表存储在 Go Backend 的 MySQL 中，由 Go 后端管理。BFF 不直接访问数据库。

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

### BFF 端点（Next.js，供浏览器调用）

这些端点仅供 Web 浏览器使用，CLI 不走这里。

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/bff/auth/token` | 授权码换 token，设置 httpOnly Cookie |
| `POST` | `/bff/auth/refresh` | 从 Cookie 读 Refresh Token，换新 Access Token |
| `POST` | `/bff/auth/logout` | 清除 Cookie，通知 Go 后端吊销 Refresh Token |

#### POST /bff/auth/token

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
  "expiresIn": 3600
}
```
同时设置 httpOnly Cookie：
```
Set-Cookie: refresh_token=f7a2bc9d...; HttpOnly; Secure; SameSite=Strict; Max-Age=604800; Path=/bff/auth
```

**响应（失败）：**
- HTTP 400：授权码无效、已过期、`redirectUri` 不匹配
- HTTP 502：Casdoor 或 Go 后端服务不可用

#### POST /bff/auth/refresh

**请求：** 无 Body，Refresh Token 自动从 Cookie 读取

**响应（成功）：**
```json
{
  "accessToken": "eyJhbGci...",
  "expiresIn": 3600
}
```
同时更新 httpOnly Cookie（新 Refresh Token）。

**响应（失败）：** HTTP 401，同时清除 Cookie

#### POST /bff/auth/logout

**请求：** 无 Body

**响应：** HTTP 204，清除 httpOnly Cookie

---

### Go Backend 端点（供 BFF 内部调用 + CLI 直连）

| 方法 | 路径 | 调用方 | 说明 |
|------|------|--------|------|
| `POST` | `/api/auth/token` | BFF | 授权码换 token 对 |
| `POST` | `/api/auth/refresh` | BFF | Refresh Token 轮换 |
| `POST` | `/api/auth/logout` | BFF | 吊销 Refresh Token |

BFF → Go Backend 之间的通信使用服务内网，Go Backend 的这些端点**不对外暴露**（通过网关/防火墙限制）。

#### POST /api/auth/token（Go Backend）

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
  "refreshToken": "f7a2bc9d...（64位随机 hex，不透明随机串）",
  "expiresIn": 3600
}
```

Refresh Token 格式：32 字节 CSPRNG → hex 编码 → 64 位字符串，存储时取 SHA256 hash。

#### POST /api/auth/refresh（Go Backend）

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

**响应（失败）：** HTTP 401

#### POST /api/auth/logout（Go Backend）

**请求：**
```json
{
  "refreshToken": "f7a2bc9d..."
}
```

**响应：** HTTP 204

---

### GraphQL（API Key 管理，Org 域，Go Backend）

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
| Access Token 存储 | 浏览器内存（不写 localStorage，防 XSS） |
| Refresh Token 有效期 | 7 天（滑动过期） |
| Refresh Token 存储（客户端） | httpOnly + Secure + SameSite=Strict Cookie，Path=/bff/auth |
| Refresh Token 存储（服务端） | MySQL hash，不存明文 |
| Refresh Token 策略 | 轮换：每次 refresh 生成新 token 对，旧 token 立即 revoked |
| API Key 格式 | `mc_` 前缀 + 40位 Base62（CSPRNG 生成，约 238 bit 熵） |
| API Key 前缀 | 完整 key 的前 10 位（含 `mc_`），用于 UI 识别 |
| API Key 有效期 | 默认永久，可选设置过期时间 |

**API Key 示例：**
```
mc_a1B2c3D4e5F6g7H8i9J0k1L2m3N4o5P6q7R8
└──┘└──────────────────────────────────┘
前缀  40位 Base62（CSPRNG）
```

### Refresh Token 轮换逻辑（Go Backend）

```
BFF → POST /api/auth/refresh { refreshToken }
  → Go: SHA256(token) → 查 DB
  → 找到且未 revoked → 正常流程
      旧 token: revoked_at = now()
      新 token 对: 写入 DB（expires_at = now + 7天）
      → 返回新 token 对给 BFF
  → 找到但已 revoked → 安全事件处理（见下方）
  → 未找到或已过期 → HTTP 401
```

### Refresh Token 疑似盗用处理

已轮换的 Refresh Token 被再次使用，是 token 盗用的强信号：

```
BFF → POST /api/auth/refresh（携带已 revoked 的 token）
  → Go: 检测到 revoked_at IS NOT NULL
  → 吊销该 user_id 下所有 Refresh Token（全设备强制下线）
  → 写入审计日志（event=REUSE_DETECTED）
  → 返回 HTTP 401
  → BFF: 清除 httpOnly Cookie
  → 浏览器: 跳转登录页
```

### 审计日志

安全事件写入 `security_audit_logs` 表：

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

多个标签页同时检测到 401，可能并发请求 BFF refresh 端点：

- BFF 负责防止并发：使用单例 Promise，同一时刻只发起一个 refresh 请求（现有实现已有此逻辑，保持不变）
- Go 后端不设宽限窗口，token 轮换后旧 token 立即无效

### Go Backend 统一验证中间件

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

**清理策略：** Go Backend 后台 goroutine（使用 `bizutils.GoWithCtx`），每天执行一次。

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

### Go Backend 变更

| 项目 | 变更内容 |
|------|---------|
| Refresh Token | 从无状态 JWT 改为写入 `refresh_tokens` DB 表，格式改为不透明 hex 串 |
| `/api/auth/refresh` | 验证后执行轮换逻辑 + 盗用检测 |
| `/api/auth/logout` | 新增：将指定 Refresh Token 标记 revoked |
| Casdoor JWT 兼容层 | 移除：Casdoor JWT 只在 `/api/auth/token` 内部验证，不再对外暴露 |

### Next.js BFF 变更

| 项目 | 变更内容 |
|------|---------|
| `/bff/auth/token` | 新增：BFF 层包装，设置 httpOnly Cookie |
| `/bff/auth/refresh` | 新增：从 Cookie 读取 Refresh Token，转发 Go 后端，更新 Cookie |
| `/bff/auth/logout` | 新增：清除 Cookie + 通知 Go 后端吊销 |
| Access Token 存储 | 从 localStorage 改为内存（页面变量/Zustand store） |
| Refresh Token 存储 | 从 localStorage 改为 httpOnly Cookie（BFF 管理，JS 不可见） |

### 保持不变的部分

| 项目 | 说明 |
|------|------|
| Access Token 结构 | JWT + HMAC-SHA256，payload 只含 userId |
| Casdoor 集成流程 | 授权码 → 验证 → 查/创 user，逻辑不变 |
| Go 中间件注入 userId 到 Context | 接口不变，业务层无感知 |
| BFF 并发 refresh 防护 | 单例 Promise 机制保持不变 |
| API Key 直连 Go Backend | CLI 不经过 BFF |

### 新增的部分

| 项目 | 说明 |
|------|------|
| `refresh_tokens` 表 | Go Backend MySQL 新增表 |
| `api_keys` 表 | Go Backend MySQL 新增表 |
| `security_audit_logs` 表 | Go Backend MySQL 新增表 |
| API Key CRUD | Org GraphQL 新增 1 个 Query + 3 个 Mutation |
| API Key 验证路径 | Go 中间件新增 `mc_` 前缀识别逻辑 |
| 数据清理 goroutine | Go Backend 后台定时清理过期记录 |

---

## 7. 不在本次范围内

- `logout-all`（吊销所有设备的 Refresh Token）：可在后续迭代添加
- API Key 权限范围（scope）：当前所有 key 等同于用户完整权限
- OAuth2 第三方授权：不在本次设计范围
- 将 BFF 迁移为独立服务：后续架构演进方向
