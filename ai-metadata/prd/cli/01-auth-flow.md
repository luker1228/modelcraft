# CLI 认证流程

---

## 1. 问题：BFF-Only 认证阻断 CLI 访问

现有 EndUser 认证路由全部在 `/internal/end-user/auth/*` 下，需要 BFF 的内部 Token 中间件（`ChiInternalTokenMiddleware`），CLI 无法直接调用。

参考代码：`modelcraft-backend/internal/interfaces/http/routes.go:672-686`

```go
router.Route("/internal/end-user/auth", func(r chi.Router) {
    r.Use(requestIDInjectorMiddleware)
    r.Use(internalTokenMW)  // ← BFF 专用，CLI 无法通过
    r.Post("/login", handlers.EndUserAuthHandler.Login)
    ...
})
```

---

## 2. 方案：新增公共 EndUser Auth REST 端点

后端新增公共端点（不经过 `internalTokenMW`），CLI 直接获取 JWT：

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/end-user/auth/login` | POST | 登录，返回 refreshToken + 可访问项目列表 |
| `/api/end-user/auth/select-project` | POST | 选择项目，签发 accessToken |
| `/api/end-user/auth/refresh` | POST | 刷新 accessToken |
| `/api/end-user/auth/logout` | POST | 登出，吊销 refreshToken |
| `/api/end-user/auth/me` | GET | 当前身份信息 |

这些端点复用现有 `AuthHandler` 的业务逻辑，区别仅在中间件栈：
- 不使用 `internalTokenMW`
- 使用 `X-Org-Name` header 进行 Org 定位（与现有 handler 一致）
- 可选增加 rate limiting 中间件防止暴力破解

---

## 3. 登录时序

```
Agent                     CLI                    Backend
  │                        │                        │
  │  mc auth login         │                        │
  │  --server <url>        │                        │
  │  --org acme            │                        │
  │  --username alice      │                        │
  │  --password ****       │                        │
  │───────────────────────>│                        │
  │                        │                        │
  │                        │  POST /api/end-user/   │
  │                        │  auth/login             │
  │                        │  X-Org-Name: acme       │
  │                        │  { username, password }  │
  │                        │───────────────────────>│
  │                        │                        │
  │                        │  200 {                  │
  │                        │    userId,              │
  │                        │    refreshToken,        │
  │                        │    projects: [{slug,    │
  │                        │      title}...]         │
  │                        │  }                      │
  │                        │<───────────────────────│
  │                        │                        │
  │                        │  POST /api/end-user/   │
  │                        │  auth/select-project    │
  │                        │  { refreshToken,        │
  │                        │    projectSlug }         │
  │                        │───────────────────────>│
  │                        │                        │
  │                        │  200 {                  │
  │                        │    accessToken,         │
  │                        │    expiresAt            │
  │                        │  }                      │
  │                        │<───────────────────────│
  │                        │                        │
  │  { ok: true,           │                        │
  │    projects: [...],    │                        │
  │    currentProject: ... │                        │
  │  }                     │                        │
  │<───────────────────────│                        │
```

登录成功后 CLI 自动选择第一个项目。可通过 `--project` 指定：

```bash
mc auth login --server ... --org acme --username alice --password "..." --project hr
```

---

## 4. Token 管理

### 4.1 存储位置

```
~/.config/modelcraft/credentials.json
```

### 4.2 存储格式

```json
{
  "server": "https://mc.example.com",
  "orgName": "acme",
  "projectSlug": "sales",
  "userId": "01944...",
  "accessToken": "eyJ...",
  "refreshToken": "...",
  "expiresAt": "2026-04-25T12:00:00Z"
}
```

### 4.3 自动刷新

CLI 在每次请求前检查 `expiresAt`：
1. 若距过期 > 60s → 正常使用 accessToken
2. 若距过期 ≤ 60s → 自动调用 `/api/end-user/auth/refresh` 续期
3. 续期失败 → 返回 `TOKEN_EXPIRED` 错误，suggestion 中提示 `mc auth login`

### 4.4 环境变量覆盖

支持通过环境变量覆盖 credentials，方便 CI/CD 场景：

```bash
export MC_SERVER=https://mc.example.com
export MC_ORG=acme
export MC_ACCESS_TOKEN=eyJ...
```

优先级：环境变量 > credentials.json

---

## 5. 命令详细设计

### 5.1 `mc auth login`

```bash
mc auth login \
  --server https://mc.example.com \
  --org acme \
  --username alice \
  --password "s3cret"
```

**成功输出**：

```json
{
  "ok": true,
  "userId": "01944...",
  "orgName": "acme",
  "projects": [
    { "slug": "sales", "title": "销售系统" },
    { "slug": "hr", "title": "HR 系统" }
  ],
  "currentProject": "sales"
}
```

**失败输出**：

```json
{
  "ok": false,
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "用户名或密码错误",
    "retryable": true,
    "suggestion": "请检查用户名和密码是否正确"
  }
}
```

```json
{
  "ok": false,
  "error": {
    "code": "ACCOUNT_FORBIDDEN",
    "message": "账号已被禁用",
    "retryable": false,
    "suggestion": "请联系管理员启用账号"
  }
}
```

```json
{
  "ok": false,
  "error": {
    "code": "NO_PROJECT_ACCESS",
    "message": "您暂无项目访问权限",
    "retryable": false,
    "suggestion": "请联系管理员授予项目访问权限"
  }
}
```

### 5.2 `mc auth status`

```bash
mc auth status
```

```json
{
  "ok": true,
  "authenticated": true,
  "server": "https://mc.example.com",
  "orgName": "acme",
  "userId": "01944...",
  "currentProject": "sales",
  "tokenExpiresAt": "2026-04-25T12:00:00Z",
  "projects": [
    { "slug": "sales", "title": "销售系统" },
    { "slug": "hr", "title": "HR 系统" }
  ]
}
```

### 5.3 `mc auth switch-project`

```bash
mc auth switch-project --project hr
```

```json
{
  "ok": true,
  "previousProject": "sales",
  "currentProject": "hr"
}
```

### 5.4 `mc auth logout`

```bash
mc auth logout
```

```json
{
  "ok": true,
  "message": "Successfully logged out"
}
```

### 5.5 `mc auth refresh`

```bash
mc auth refresh
```

```json
{
  "ok": true,
  "expiresAt": "2026-04-25T14:00:00Z"
}
```

---

## 6. 后端 API 规格

### 6.1 `POST /api/end-user/auth/login`

**Request**：
```
POST /api/end-user/auth/login
X-Org-Name: acme
Content-Type: application/json

{ "username": "alice", "password": "s3cret" }
```

**Response 200**：
```json
{
  "requestId": "req-xxx",
  "userId": "01944...",
  "refreshToken": "rt-xxx",
  "projects": [
    { "slug": "sales", "title": "销售系统" }
  ]
}
```

**Response 401**：`{ "error": "INVALID_CREDENTIALS" }`  
**Response 403**：`{ "error": "ACCOUNT_FORBIDDEN" }` 或 `{ "error": "NO_PROJECT_ACCESS" }`

### 6.2 `POST /api/end-user/auth/select-project`

**Request**：
```
POST /api/end-user/auth/select-project
X-Org-Name: acme
Content-Type: application/json

{ "refreshToken": "rt-xxx", "projectSlug": "sales" }
```

**Response 200**：
```json
{
  "requestId": "req-xxx",
  "accessToken": "eyJ...",
  "selectedProject": "sales",
  "expiresAt": "2026-04-25T12:00:00Z"
}
```

**Response 403**：`{ "error": "PROJECT_ACCESS_DENIED" }`

### 6.3 `POST /api/end-user/auth/refresh`

**Request**：
```
POST /api/end-user/auth/refresh
X-Org-Name: acme
Content-Type: application/json

{ "refreshToken": "rt-xxx" }
```

**Response 200**：
```json
{
  "requestId": "req-xxx",
  "accessToken": "eyJ...",
  "refreshToken": "rt-new-xxx",
  "expiresAt": "2026-04-25T14:00:00Z"
}
```

### 6.4 `POST /api/end-user/auth/logout`

**Request**：
```
POST /api/end-user/auth/logout
X-Org-Name: acme
Content-Type: application/json

{ "refreshToken": "rt-xxx" }
```

**Response 200**：`{ "requestId": "req-xxx" }`

### 6.5 `GET /api/end-user/auth/me`

**Request**：
```
GET /api/end-user/auth/me
Authorization: Bearer eyJ...
X-Org-Name: acme
```

**Response 200**：
```json
{
  "requestId": "req-xxx",
  "endUser": {
    "id": "01944...",
    "username": "alice",
    "isForbidden": false,
    "createdAt": "2026-01-01T00:00:00Z"
  }
}
```
