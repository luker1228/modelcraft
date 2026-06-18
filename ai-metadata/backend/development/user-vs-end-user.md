# 用户身份体系

> 适用范围：认证、网关代理、context 命名、runtime 鉴权、RLS、PAT、`X-MC-Auth-*` 相关开发。

ModelCraft 是低代码平台，**所有请求的发起方永远是 tenant user**（设计者 / 管理员）。

平台只有一种用户：**tenant user**。`user_orgs.is_admin` 字段存在，但目前只是一个标记——表示该用户是这个 Org 的管理员，暂无基于此字段的权限控制衍生逻辑。

"EndUser" 只在 RLS 上下文里有意义：tenant user 调用 Open Data API 时，通过 `X-MC-Auth-*` headers 声明"我当前代表哪个业务终端用户在操作"，RLS 引擎用这个值过滤数据行。

---

## 请求结构

Open Data API 目前**仅支持 PAT 认证**。

```
tenant PAT 调用方
  │
  ├── Authorization: Bearer <PAT>         ← 认证：PAT 对应的 tenant user
  ├── X-User-ID: <tenant-user-id>         ← Gateway 注入（来自 PAT 持有者身份）
  ├── X-MC-Auth-Userid-Str: <end-user-id> ← 声明：代表哪个终端用户（RLS 过滤用）
  ├── X-MC-Auth-Roles: viewer             ← 声明：该终端用户的角色
  └── X-MC-Auth-Useadmin: true            ← 可选：设计者以 admin 身份操作数据，跳过 RLS 用户过滤
```

`X-MC-Auth-Useadmin: true` 的语义：**设计者本人想用 admin 权限操作 runtime 数据**，不扮演任何终端用户，RLS 策略中 role=admin 的 policy 生效。

## `X-MC-Auth-*` header 完整列表

| Header | Go 常量 | 类型 | 说明 |
|--------|---------|------|------|
| `X-MC-Auth-Userid-Str` | `XMCAuthUserIDStr` | string | 字符串型 end-user ID，优先使用，与 `-Int` 互斥 |
| `X-MC-Auth-Userid-Int` | `XMCAuthUserIDInt` | int64 | 数值型 end-user ID，`-Str` 不存在时使用，与 `-Str` 互斥 |
| `X-MC-Auth-Username` | `XMCAuthUserName` | string | 终端用户名 |
| `X-MC-Auth-Roles` | `XMCAuthRoles` | string（逗号分隔） | 角色列表，用于 RLS policy `role` 字段匹配 |
| `X-MC-Auth-Useadmin` | `XMCAuthUseAdmin` | `"true"` | 设计者以 admin 身份操作 runtime 数据，跳过 RLS 用户过滤 |

> **`X-MC-Auth-Userid`（无后缀）已废弃，不存在**。传该 header 会被 middleware 忽略，导致 `auth.userid` 为空、RLS CHECK 必然失败。

## RLS 表达式中的 `auth.*` 变量映射

| 表达式变量 | 来源 header |
|-----------|------------|
| `auth.userid` | `X-MC-Auth-Userid-Str` 或 `X-MC-Auth-Userid-Int` |
| `auth.username` | `X-MC-Auth-Username` |
| `auth.roles` | `X-MC-Auth-Roles` |

---

## context 命名约定

```go
// tenant user（所有链路）
ctxutils.SetUserID(ctx, userID)
ctxutils.GetUserIDFromContext(ctx)

// end-user 角色（只在 Open Data API / runtime SQL）
ctxutils.SetEndUserID(ctx, endUserID)
ctxutils.GetEndUserIDFromContext(ctx)
```

禁止做法：

```go
// 禁止：把 end-user ID 塞进 UserID context
ctxutils.SetUserID(ctx, token.EndUserID)

// 禁止：runtime 代码从 UserID 里读 end-user 身份
endUserID, _ := ctxutils.GetUserIDFromContext(ctx)

// 禁止：UserID 和 EndUserID 互相 fallback——两者正交，不可替代
if userID == "" {
    userID, _ = ctxutils.GetEndUserIDFromContext(ctx) // 错误
}
```

## ChiJWTAuthMiddleware: link-based identity routing

`X-User-ID` 永远是 tenant user ID。在 `/end-user/*` 路径上，middleware 将其写入 `EndUserID` context（这条路径的调用方以 end-user 角色操作，需要映射到 end-user context）：

```go
// /end-user/*  → 以 end-user 角色操作的链路 → 写入 EndUserID
// /graphql/*   → tenant 管理链路 → 写入 UserID
if strings.HasPrefix(r.URL.Path, "/end-user/") {
    ctx = ctxutils.SetEndUserID(ctx, userID)
} else {
    ctx = ctxutils.SetUserID(ctx, userID)
}
```

## Link-based allowEndUser gating

`allowEndUser` 门禁通过**独立的 handler 实例**区分链路，不通过 context 传递链路信息：

- 管理链路 handler：`NewHasPermissionDirective` → `enforceEndUserGate: false`，统一走 RBAC
- EndUser 链路 handler：`NewEndUserHasPermissionDirective` → `enforceEndUserGate: true`，执行 `allowEndUser` 门禁

---

## 代码锚点

- RLS context middleware：`modelcraft-backend/internal/interfaces/http/middleware/rls_context_middleware.go`
- Header 常量：`modelcraft-backend/pkg/httpheader/headers.go`
- Gateway 认证：`modelcraft-gateway/internal/auth/service.go`

---

## 一句话准则

如果这段逻辑是"操作 ModelCraft 平台本身"，用 `UserID`（tenant user）。  
如果这段逻辑是"tenant 代表业务终端用户过滤 SQL 数据行"，用 `EndUserID`（RLS 上下文）。
