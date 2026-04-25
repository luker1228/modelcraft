# CLI 错误处理

---

## 1. 设计原则

- 所有正常输出（含错误响应）到 **stdout**，格式为 JSON
- 诊断/调试信息到 **stderr**（`--verbose` 开启）
- 每个错误包含 `code`、`message`、`retryable`、`suggestion`，使 Agent 可自修正
- 语义化退出码，Agent 可通过 `$?` 快速分类错误

---

## 2. 统一错误输出格式

```json
{
  "ok": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error description",
    "retryable": true,
    "suggestion": "Actionable fix suggestion for the agent",
    "details": {}
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `ok` | bool | 永远为 `false` |
| `error.code` | string | 机器可读错误码（UPPER_SNAKE_CASE） |
| `error.message` | string | 人类可读描述 |
| `error.retryable` | bool | Agent 是否可以自动重试/修正 |
| `error.suggestion` | string | 具体的修正建议（可直接执行的命令） |
| `error.details` | object | 错误上下文（可选，如请求的值 vs 服务端限制） |

---

## 3. 退出码表

| 退出码 | 含义 | 错误码示例 |
|--------|------|-----------|
| `0` | 成功 | — |
| `1` | 通用/未知错误 | `UNKNOWN_ERROR` |
| `2` | 参数错误（Agent 可自修正） | `INVALID_ARGUMENT`, `INVALID_WHERE_SYNTAX`, `MISSING_REQUIRED_FLAG` |
| `3` | 认证失败 | `UNAUTHENTICATED`, `TOKEN_EXPIRED`, `INVALID_CREDENTIALS` |
| `4` | 权限不足 | `PERMISSION_DENIED`, `PROJECT_ACCESS_DENIED` |
| `5` | 资源不存在 | `NOT_FOUND`, `MODEL_NOT_FOUND`, `DATABASE_NOT_FOUND` |
| `6` | 服务端限制 | `TAKE_EXCEEDS_LIMIT`, `RATE_LIMITED` |
| `7` | 服务端错误（Agent 无法修正） | `INTERNAL_ERROR`, `SERVICE_UNAVAILABLE` |

Agent 可以用退出码快速决策：
- `0` → 成功
- `2` → 检查 `suggestion` 修正参数后重试
- `3` → 执行 `mc auth refresh` 或 `mc auth login`
- `4` → 无法自修正，报告给用户
- `5` → 检查 `suggestion`（可能有拼写建议），或运行 `mc catalog`
- `6` → 读取 `details.limit` 调整参数后重试
- `7` → 等待后重试或报告

---

## 4. 错误示例

### 4.1 参数错误（退出码 2）

```json
{
  "ok": false,
  "error": {
    "code": "INVALID_WHERE_SYNTAX",
    "message": "Failed to parse --where flag: invalid JSON at position 15",
    "retryable": true,
    "suggestion": "Check JSON syntax. Expected format: '{\"field\":{\"operator\":\"value\"}}'"
  }
}
```

```json
{
  "ok": false,
  "error": {
    "code": "MISSING_REQUIRED_FLAG",
    "message": "--where is required for 'mc get' command",
    "retryable": true,
    "suggestion": "Provide --where flag: mc get <path> --where '{\"id\":{\"equals\":\"...\"}}'"
  }
}
```

### 4.2 认证失败（退出码 3）

```json
{
  "ok": false,
  "error": {
    "code": "TOKEN_EXPIRED",
    "message": "Access token has expired",
    "retryable": true,
    "suggestion": "Run 'mc auth refresh' to obtain a new token"
  }
}
```

```json
{
  "ok": false,
  "error": {
    "code": "UNAUTHENTICATED",
    "message": "No credentials found",
    "retryable": true,
    "suggestion": "Run 'mc auth login --server <url> --org <org> --username <user> --password <pass>'"
  }
}
```

### 4.3 资源不存在（退出码 5）

```json
{
  "ok": false,
  "error": {
    "code": "MODEL_NOT_FOUND",
    "message": "Model 'user' not found in database 'maindb'",
    "retryable": false,
    "suggestion": "Run 'mc catalog models --database maindb' to see available models. Did you mean 'users'?"
  }
}
```

### 4.4 服务端限制（退出码 6）

```json
{
  "ok": false,
  "error": {
    "code": "TAKE_EXCEEDS_LIMIT",
    "message": "Requested take (500) exceeds server maximum (100)",
    "retryable": true,
    "suggestion": "Use --take 100 and paginate with --skip",
    "details": {
      "requested": 500,
      "limit": 100
    }
  }
}
```

### 4.5 服务端错误（退出码 7）

```json
{
  "ok": false,
  "error": {
    "code": "SERVICE_UNAVAILABLE",
    "message": "Backend server is unreachable",
    "retryable": true,
    "suggestion": "Check network connectivity and server status. Retry after a few seconds."
  }
}
```

---

## 5. Limit 执行机制

### 5.1 原则

- CLI 是薄客户端，**不做本地 limit 校验**
- 服务端是唯一权威
- 错误响应包含实际 limit 值，Agent 据此自修正

### 5.2 典型流程

```
Agent: mc query users --take 500
  │
  ├─ CLI → Backend: GraphQL findMany(take: 500)
  │
  ├─ Backend: take(500) > maxTake(100) → reject
  │
  ├─ CLI ← Backend: error
  │
  └─ Agent 收到:
     {
       "ok": false,
       "error": {
         "code": "TAKE_EXCEEDS_LIMIT",
         "details": { "requested": 500, "limit": 100 }
       }
     }
     │
     ├─ Agent 读取 details.limit = 100
     │
     └─ Agent 自修正:
        mc query users --take 100 --skip 0
        mc query users --take 100 --skip 100
        mc query users --take 100 --skip 200
        ...（直到 meta.hasMore = false）
```

### 5.3 Limit 发现

Agent 可通过 `mc describe` 提前获取 limit 信息，避免试错：

```bash
mc describe sales.maindb.users | jq '.model.limits'
```

```json
{
  "maxTake": 100,
  "defaultTake": 20,
  "note": "Limits are enforced server-side. Use --take and --skip for pagination."
}
```

---

## 6. TTY 感知行为

| 场景 | stdout 是 TTY | stdout 不是 TTY |
|------|--------------|----------------|
| 默认输出格式 | JSON（pretty-printed） | JSON（compact，一行） |
| 进度指示 | stderr 输出 spinner | 无 |
| 颜色 | stderr 带颜色 | 无颜色 |
| 交互提示 | v1 不支持 | N/A |

> v1 Agent-only，不支持交互式提示。所有参数必须通过标志提供。
