# End-User API Token Runtime Auth — Design

**Date:** 2026-06-04  
**Status:** Approved  
**Scope:** Backend only (`modelcraft-backend`)

---

## 背景

End-User 已可在 Dashboard 创建 `mc_pat_*` 格式的 API Token。但当前 Runtime GraphQL 端点只接受 JWT Bearer Token，End-User 无法直接用 API Token 调用 Runtime 接口，必须先走登录换取 JWT。

本设计目标：让 End-User 可以用 `mc_pat_*` API Token 作为 Bearer 直接调用 Runtime GraphQL 端点，无需先换取 JWT。

---

## 范围

### 受影响路由（仅此一条）

```
/end-user/graphql/org/{orgName}/project/{projectSlug}/db/{db}/model/{model}
```

### 不受影响

- 设计时路由 `/graphql/org/...` — 无任何改动
- End-User Org/Project GraphQL 路由 `/end-user/graphql/org/...` — 无改动
- 现有 JWT 认证路径 — 完全保留

---

## 架构设计

### 中间件链（修改后，仅 `/end-user/` runtime 路由）

```
Request
  ↓
requestIDInjectorMiddleware
  ↓
ApiTokenToIdentityMiddleware          ← 新增，仅此路由
  · Bearer 以 "mc_pat_" 开头 → DB 验证 → 注入 EndUserIdentity → 继续
  · Bearer 不是 mc_pat_*      → pass-through，不做任何事
  · 验证失败                   → 401 立即返回
  ↓
ChiJWTAuthMiddleware                  ← 已有（微调：context 有 EndUserIdentity 则跳过）
  ↓
ChiGraphQLOrgMiddleware / cacheMW
  ↓
ModelRuntimeHandler.HandleQuery / HandlePlayground
```

### 设计原则

- `ApiTokenToIdentityMiddleware` 注入的 `EndUserIdentity` 与 JWT 认证后结构完全相同（`Issuer = "mc-platform"`），下游 handler 零感知
- 现有 runtime handler 和所有下游逻辑零改动
- `APITokenService.ValidateToken` 已存在，无需新增应用层逻辑

---

## 实现细节

### 新文件：`internal/interfaces/http/middleware/api_token_middleware.go`

```go
// ApiTokenToIdentityMiddleware 检查 Bearer token 是否为 mc_pat_* 格式。
// 是则通过 APITokenService 验证，成功后将 EndUserIdentity 注入 context。
// 否则 pass-through，交由下游 JWT 中间件处理。
type ApiTokenToIdentityMiddleware struct {
    tokenService ApiTokenValidator
    logger       logfacade.Logger
}

// ApiTokenValidator 是 appEnduser.APITokenService 的最小接口（便于测试 mock）
type ApiTokenValidator interface {
    ValidateToken(ctx context.Context, plaintext string) (*domainenduser.APIToken, error)
}

func (m *ApiTokenToIdentityMiddleware) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        bearer := extractBearer(r)
        if !strings.HasPrefix(bearer, "mc_pat_") {
            next.ServeHTTP(w, r) // pass-through
            return
        }
        token, err := m.tokenService.ValidateToken(r.Context(), bearer)
        if err != nil || token == nil {
            http.Error(w, `{"error":"Unauthorized: invalid API token"}`, http.StatusUnauthorized)
            return
        }
        identity := &EndUserIdentity{
            EndUserID: token.EndUserID,
            Issuer:    issuerPlatform,
        }
        ctx := context.WithValue(r.Context(), endUserContextKey, identity)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 微调：`ChiJWTAuthMiddleware`（已有文件）

在 JWT 验证逻辑最前面新增短路：

```go
// 上游已注入 EndUserIdentity（来自 API Token 验证），跳过 JWT
if GetEndUserIdentity(r.Context()) != nil {
    next.ServeHTTP(w, r)
    return
}
// ... 原有 JWT 验证逻辑不变
```

### 改动：`routes.go` — `SetupRuntimeGraphQLRoutesOnChi`

函数签名新增 `apiTokenSvc` 参数，仅为 `/end-user/` 路径组合新的中间件链：

```go
func SetupRuntimeGraphQLRoutesOnChi(
    router chi.Router,
    handlers *RuntimeHandlers,
    cfg *config.Config,
    apiTokenSvc appEnduser.APITokenService, // ← 新增
) {
    ...
    apiTokenMW := NewApiTokenToIdentityMiddleware(apiTokenSvc, logger)

    // /end-user/ 路径：API Token 优先，JWT fallback
    endUserRuntimeMW := func(next http.Handler) http.Handler {
        return requestIDInjectorMiddleware(
            apiTokenMW.Middleware(jwtMW(orgMW(cacheMW(next)))))
    }

    // /graphql/ 路径：原有 runtimeMW 完全不变
    runtimePath := "/graphql/org/{orgName}/project/{projectSlug}/db/{db}/model/{model}"
    router.With(runtimeMW).Get(runtimePath, handlers.ModelRuntimeHandler.HandlePlayground)
    router.With(runtimeMW).Post(runtimePath, handlers.ModelRuntimeHandler.HandleQuery)

    endUserRuntimePath := "/end-user/graphql/org/{orgName}/project/{projectSlug}/db/{db}/model/{model}"
    router.With(endUserRuntimeMW).Get(endUserRuntimePath, handlers.ModelRuntimeHandler.HandlePlayground)
    router.With(endUserRuntimeMW).Post(endUserRuntimePath, handlers.ModelRuntimeHandler.HandleQuery)
}
```

### 调用方：`chi_setup.go`（或 `server.go`）

`SetupRuntimeGraphQLRoutesOnChi` 调用处需传入 `handlers.EndUserAPITokenService`。

---

## 错误处理

| 场景 | HTTP 状态 | 响应体 |
|------|-----------|--------|
| `mc_pat_*` token 不存在 | 401 | `{"error":"Unauthorized: invalid API token"}` |
| `mc_pat_*` token 已过期 / 已删除 | 401 | `{"error":"Unauthorized: invalid API token"}` |
| DB 查询失败（系统错误） | 401 | 同上（内部记录 error log，不暴露） |
| 非 `mc_pat_*`，JWT 也无效 | 401 | 原有 JWT 中间件响应（不变） |

> 统一返回 401 而非 403，避免暴露 token 存在性信息。

---

## 测试策略

### 新增单元测试：`api_token_middleware_test.go`

Table-driven，3 个核心 case：

| Case | 输入 | 预期 |
|------|------|------|
| Bearer 非 `mc_pat_*` | `Bearer eyJ...` | pass-through，`ValidateToken` 不调用 |
| `mc_pat_*` 验证失败 | `Bearer mc_pat_invalid` | 401，`next` 不调用 |
| `mc_pat_*` 验证成功 | `Bearer mc_pat_abc` | `EndUserIdentity` 注入 context，`next` 调用 |

`ApiTokenValidator` 接口设计使得测试可直接 mock，无需真实 DB。

### 现有测试

`ChiJWTAuthMiddleware` 的现有测试无需修改（短路条件在原测试场景中不触发）。

---

## 改动文件清单

| 文件 | 类型 |
|------|------|
| `internal/interfaces/http/middleware/api_token_middleware.go` | 新增 |
| `internal/interfaces/http/middleware/api_token_middleware_test.go` | 新增 |
| `internal/interfaces/http/middleware/chi_middleware.go`（或 JWT 中间件所在文件） | 微改（+3 行短路） |
| `internal/interfaces/http/routes.go` | 微改（函数签名 +1 参数，end-user runtime MW 组合） |
| `internal/interfaces/http/chi_setup.go`（或 server.go） | 微改（传参 +1） |

---

## 不在本次范围内

- API Token 的 scope 限制（如只允许调用特定 model）— 留待后续迭代
- 前端 Dashboard UI 变更 — API Token 管理页已存在，无需改动
- OpenAPI YAML 文档更新 — runtime 端点目前未在 openapi.yaml 中描述
