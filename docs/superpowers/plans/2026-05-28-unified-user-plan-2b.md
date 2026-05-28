# Unified User System — Plan 2b: App Service + Token + Handler 合并

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 `TokenService.Login` 支持普通用户（username-only，无手机号），在 JWT 中注入 `is_admin`，统一 cookie 为 `mc_refresh_token`，在 `enduser` handler 中将所有端点路由到已有的 `auth` handler/service，并更新 Gateway 注入 `X-Is-Admin` header。

**Architecture:** 不重写 `TokenService`，而是扩展它：加 `is_admin` 到 JWT claims，处理 username-only 登录（无 phone 校验）。`EndUserAuthAppService` 的 login/refresh/logout 全部重定向到 `TokenService`，session 存储复用 `refresh_tokens` 表（Plan 2a 已完成 stub，只需连接到正确的 service）。

**Tech Stack:** Go, JWT ES256, httpOnly cookie

**Plan 2a 结束状态：**
- `go build ./...` — ✅ 0 errors
- `IssueAccessToken(userID, orgName, aud)` — 3 参数，无 isAdmin
- `TokenService.Login` — 通过 `membershipRepo.ListByUserWithDetails` 获取 orgName，但不读 IsAdmin
- `EndUserAuthAppService` — stub，所有方法返回 `errEndUserRepoDeprecated`
- `mc_enduser_refresh_token` cookie — 还存在于 enduser handler

---

## 文件变更地图

| 操作 | 文件 |
|------|------|
| **修改** | `internal/domain/auth/jwt_signer.go` — `IssueAccessToken` 加 `isAdmin bool` 参数 |
| **修改** | `internal/domain/auth/jwt_signer_test.go` — 更新测试调用 |
| **修改** | `internal/app/auth/token_service.go` — Login/Refresh 读取 `IsAdmin`，传给 `IssueAccessToken` |
| **修改** | `internal/app/auth/token_service.go` — `LoginCommand` 允许 username-only（无 phone 强校验） |
| **修改** | `internal/app/auth/commands.go`（如果存在）— 或在 token_service.go 中找到 LoginCommand |
| **修改** | `internal/interfaces/http/handlers/enduser/auth_handler.go` — `EndUserLogin/Refresh/Logout/Me` 全部路由到 `auth.Handler` |
| **修改** | `internal/interfaces/http/routes.go` — enduser auth 路由复用 auth handler，删除对 `NewSqlEndUserRepository` 的调用 |
| **修改** | `modelcraft-gateway/` — 找到 JWT 验证和 header 注入处，加 `X-Is-Admin` |

---

## Task 1: IssueAccessToken 加 isAdmin 参数

**Files:**
- Modify: `modelcraft-backend/internal/domain/auth/jwt_signer.go`
- Modify: `modelcraft-backend/internal/domain/auth/jwt_signer_test.go`

- [ ] **Step 1: 读取 jwt_signer.go 中的 IssueAccessToken 方法**

```bash
grep -n "IssueAccessToken\|PlatformClaims{" \
  modelcraft-backend/internal/domain/auth/jwt_signer.go
```

- [ ] **Step 2: 修改 IssueAccessToken 签名，加 isAdmin 参数**

将 `IssueAccessToken` 改为：

```go
// IssueAccessToken 为指定用户签发短效 ES256 JWT（PlatformClaims 格式）。
// orgName 不能为空。aud 用于标识受众类型（tenant / end_user）。
// isAdmin 表示用户在该 Org 中是否为管理员，注入 is_admin claim 供 Gateway 读取。
func (s *JWTSigner) IssueAccessToken(
    userID, orgName string,
    aud jwt.ClaimStrings,
    isAdmin bool,
) (string, error) {
    if orgName == "" {
        return "", errors.New("jwt_signer: orgName is required")
    }
    now := time.Now()
    claims := &PlatformClaims{
        UserID:  userID,
        OrgName: orgName,
        IsAdmin: isAdmin,
        Key:     ApisixConsumerKey,
        RegisteredClaims: jwt.RegisteredClaims{
            Issuer:    string(IssuerPlatform),
            Subject:   userID,
            Audience:  aud,
            IssuedAt:  jwt.NewNumericDate(now),
            ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
    return token.SignedString(s.privateKey)
}
```

- [ ] **Step 3: 更新 jwt_signer_test.go 中的调用**

```bash
grep -n "IssueAccessToken" modelcraft-backend/internal/domain/auth/jwt_signer_test.go
```

将所有 3-arg 调用改为 4-arg（加 `false` 或 `true`）。

- [ ] **Step 4: 找到所有其他 IssueAccessToken 调用方并修复**

```bash
grep -rn "IssueAccessToken" modelcraft-backend/internal/ --include="*.go"
```

对每个调用，在末尾加 `false`（默认非 admin）——实际 isAdmin 值将在 Task 2 的 token_service.go 中正确设置。

- [ ] **Step 5: 编译检查**

```bash
cd modelcraft-backend && go build ./internal/domain/auth/... 2>&1
```

预期：0 errors。

- [ ] **Step 6: 运行测试**

```bash
cd modelcraft-backend && go test ./internal/domain/auth/... -v 2>&1 | tail -10
```

- [ ] **Step 7: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-backend/internal/domain/auth/
git commit -m "domain: add isAdmin param to IssueAccessToken, inject into PlatformClaims"
```

---

## Task 2: TokenService.Login 读取 IsAdmin，传给 token

**Files:**
- Modify: `modelcraft-backend/internal/app/auth/token_service.go`

`MembershipWithDetails` 已有 `IsAdmin` 字段（Plan 2a 中 `Membership` struct 已更新）。只需在 login 时读取并传给 `IssueAccessToken`。

- [ ] **Step 1: 读取 Login 方法中的 membership 查询段（约 390-415 行）**

```bash
sed -n '388,415p' modelcraft-backend/internal/app/auth/token_service.go
```

- [ ] **Step 2: 修改 Login 中的 membership 读取和 token 签发**

将当前：
```go
// Fetch user's primary organization (first one)
var orgName string
if s.membershipRepo != nil {
    memberships, listErr := s.membershipRepo.ListByUserWithDetails(ctx, u.ID, 1)
    if listErr == nil && len(memberships) > 0 {
        orgName = memberships[0].OrgName
    }
}

accessToken, err := s.jwtSigner.IssueAccessToken(u.ID, orgName, jwt.ClaimStrings{domainauth.AudienceTenant})
```

改为：
```go
// Fetch user's primary organization and admin status (single Org per user)
var orgName string
var isAdmin bool
if s.membershipRepo != nil {
    memberships, listErr := s.membershipRepo.ListByUserWithDetails(ctx, u.ID, 1)
    if listErr == nil && len(memberships) > 0 {
        orgName = memberships[0].OrgName
        isAdmin = memberships[0].IsAdmin
    }
}

accessToken, err := s.jwtSigner.IssueAccessToken(u.ID, orgName, jwt.ClaimStrings{domainauth.AudienceTenant}, isAdmin)
```

- [ ] **Step 3: 同样修改 Refresh 方法中的 IssueAccessToken 调用（约 537-545 行）**

```bash
sed -n '530,550p' modelcraft-backend/internal/app/auth/token_service.go
```

Refresh 方法也需要读取 isAdmin 并传入。找到对应段，加入同样的 membership 查询（读 IsAdmin）。

- [ ] **Step 4: 确认 LoginResult 不需要返回 isAdmin（前端从 token 解析）**

```bash
grep -n "LoginResult\|IsAdmin" modelcraft-backend/internal/app/auth/result.go 2>/dev/null || \
  grep -n "type LoginResult\|IsAdmin" modelcraft-backend/internal/app/auth/token_service.go
```

`LoginResult` 不需要加 `IsAdmin` 字段——前端从 JWT 本身解析即可。

- [ ] **Step 5: 编译检查**

```bash
cd modelcraft-backend && go build ./internal/app/auth/... 2>&1
```

- [ ] **Step 6: 运行测试**

```bash
cd modelcraft-backend && go test ./internal/app/auth/... -v 2>&1 | tail -15
```

- [ ] **Step 7: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-backend/internal/app/auth/token_service.go
git commit -m "app: inject is_admin into JWT token from membership on login/refresh"
```

---

## Task 3: LoginCommand 支持 username-only（无 phone 格式校验）

**Files:**
- Modify: `modelcraft-backend/internal/app/auth/token_service.go`（或 commands.go）

当前 `IdentifierTypeUsername` 路径通过 `userRepo.GetByName` 查找用户，已经支持用户名登录。但需要确认 `enduser` 场景（普通用户，无 phone）可以正常走此路径。

- [ ] **Step 1: 读取 LoginCommand 定义**

```bash
grep -n "type LoginCommand\|LoginCommand{" \
  modelcraft-backend/internal/app/auth/token_service.go \
  modelcraft-backend/internal/app/auth/*.go 2>/dev/null | head -20
```

- [ ] **Step 2: 确认 IdentifierTypeUsername 路径不要求 phone**

```bash
sed -n '344,360p' modelcraft-backend/internal/app/auth/token_service.go
```

预期：`case IdentifierTypeUsername:` 只调用 `s.userRepo.GetByName`，无 phone 相关逻辑。如果是，无需修改。

- [ ] **Step 3: 确认 HandleLogin 已支持 IdentifierType=username**

```bash
grep -n "IdentifierType\|USERNAME\|username" \
  modelcraft-backend/internal/interfaces/http/handlers/auth/handler.go
```

预期：已有 `case generated.USERNAME: cmd.IdentifierType = appAuth.IdentifierTypeUsername`。

如果确认已支持，此 Task 只需验证无需改动。

- [ ] **Step 4: 验证 go build 通过**

```bash
cd modelcraft-backend && go build ./... 2>&1
```

- [ ] **Step 5: 如无改动，不需要 commit**

---

## Task 4: 统一 enduser handler — 路由到 auth.Handler

**Files:**
- Modify: `modelcraft-backend/internal/interfaces/http/handlers/enduser/auth_handler.go`
- Modify: `modelcraft-backend/internal/interfaces/http/routes.go`

`EndUserAuthAppService` 所有方法已是 stub（返回 deprecated error）。需要让 enduser 的 login/refresh/logout/me 路由到 `auth.Handler` 的对应方法。

- [ ] **Step 1: 读取 enduser/auth_handler.go 的完整内容**

```bash
cat modelcraft-backend/internal/interfaces/http/handlers/enduser/auth_handler.go
```

- [ ] **Step 2: 读取 routes.go 中 enduser 路由的注册方式**

```bash
grep -n "enduser\|EndUser\|end-user" \
  modelcraft-backend/internal/interfaces/http/routes.go | head -30
```

- [ ] **Step 3: 修改 enduser/auth_handler.go**

将 `EndUserLogin`、`EndUserRefresh`、`EndUserLogout` 改为直接调用 `auth.Handler` 的对应方法：

```go
package enduser

import (
    "net/http"
    authhandler "modelcraft/internal/interfaces/http/handlers/auth"
)

// EndUserAuthHandler 将 end-user auth 路由到统一的 auth handler。
// 统一用户体系后，end-user 使用与 tenant 相同的登录接口。
type EndUserAuthHandler struct {
    authHandler *authhandler.Handler
}

// NewEndUserAuthHandler creates an EndUserAuthHandler backed by the unified auth handler.
func NewEndUserAuthHandler(h *authhandler.Handler) *EndUserAuthHandler {
    return &EndUserAuthHandler{authHandler: h}
}

// EndUserLogin delegates to the unified login handler.
// Cookie name is unified to mc_refresh_token.
func (h *EndUserAuthHandler) EndUserLogin(w http.ResponseWriter, r *http.Request) {
    h.authHandler.HandleLogin(w, r)
}

// EndUserRefresh delegates to the unified refresh handler.
func (h *EndUserAuthHandler) EndUserRefresh(w http.ResponseWriter, r *http.Request) {
    h.authHandler.HandleRefresh(w, r)
}

// EndUserLogout delegates to the unified logout handler.
func (h *EndUserAuthHandler) EndUserLogout(w http.ResponseWriter, r *http.Request) {
    h.authHandler.HandleLogout(w, r)
}

// EndUserMe delegates to the unified me handler (if exists) or returns current user info.
func (h *EndUserAuthHandler) EndUserMe(w http.ResponseWriter, r *http.Request) {
    h.authHandler.HandleMe(w, r)
}
```

注意：如果 `auth.Handler` 没有 `HandleMe` 方法，需要先添加。检查：

```bash
grep -n "func.*Handler.*HandleMe\|func.*Handle.*Me" \
  modelcraft-backend/internal/interfaces/http/handlers/auth/handler.go
```

如果没有，在 `auth/handler.go` 中添加简单的 `HandleMe`，从 JWT 中提取用户信息返回。

- [ ] **Step 4: 删除 mc_enduser_refresh_token cookie 相关代码**

`EndUserAuthHandler` 不再设置 `mc_enduser_refresh_token` cookie（统一使用 `mc_refresh_token`）。确认旧的 `setEndUserRefreshCookie`、`clearEndUserRefreshCookie` 不再被调用。

- [ ] **Step 5: 更新 routes.go**

```bash
grep -n "enduser\|EndUserAuth\|NewEndUserAuth\|endUserAuthHandler" \
  modelcraft-backend/internal/interfaces/http/routes.go | head -20
```

将 enduser auth 路由的 handler 替换为 `NewEndUserAuthHandler(authHandler)`，移除对 `NewSqlEndUserRepository` 的依赖（已在 stub 中处理）。

- [ ] **Step 6: 编译检查**

```bash
cd modelcraft-backend && go build ./... 2>&1
```

预期：0 errors。

- [ ] **Step 7: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-backend/internal/interfaces/http/handlers/enduser/ \
        modelcraft-backend/internal/interfaces/http/routes.go
git commit -m "handler: route enduser auth to unified auth handler, unify cookie name"
```

---

## Task 5: Gateway 注入 X-Is-Admin header

**Files:**
- Modify: `modelcraft-gateway/` 中 JWT 验证和 header 注入处

- [ ] **Step 1: 找到 Gateway 中 JWT 验证和 header 注入的代码**

```bash
ls modelcraft-gateway/
grep -rn "X-User-ID\|X-Org-Name\|X-Is-Admin\|UserID\|OrgName\|platform_claims\|PlatformClaims" \
  modelcraft-gateway/ --include="*.go" | head -20
```

- [ ] **Step 2: 读取 JWT 验证中间件**

找到 Gateway 中设置 `X-User-ID`、`X-Org-Name` header 的代码，读取相关文件。

- [ ] **Step 3: 加入 X-Is-Admin header 注入**

在设置 `X-User-ID` 和 `X-Org-Name` 的代码旁边，加入：

```go
// Inject X-Is-Admin from token claims
if claims.IsAdmin {
    r.Header.Set("X-Is-Admin", "true")
} else {
    r.Header.Set("X-Is-Admin", "false")
}
```

- [ ] **Step 4: 编译 Gateway**

```bash
cd modelcraft-gateway && go build ./... 2>&1
```

预期：0 errors。

- [ ] **Step 5: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-gateway/
git commit -m "gateway: inject X-Is-Admin header from JWT is_admin claim"
```

---

## Task 6: 全局编译 + 验证

- [ ] **Step 1: 全量编译**

```bash
cd /data/home/lukemxjia/modelcraft/modelcraft-backend && go build ./... 2>&1
cd /data/home/lukemxjia/modelcraft/modelcraft-gateway && go build ./... 2>&1
```

预期：两个服务都 0 errors。

- [ ] **Step 2: 运行 domain + app 测试**

```bash
cd /data/home/lukemxjia/modelcraft/modelcraft-backend && \
  go test ./internal/domain/auth/... ./internal/app/auth/... -v 2>&1 | tail -20
```

预期：全部通过。

- [ ] **Step 3: 验证 is_admin 在 JWT 中存在**

```bash
cd /data/home/lukemxjia/modelcraft/modelcraft-backend && \
  go test ./internal/domain/auth/... -run TestIssueAccessToken -v 2>&1
```

如无此测试，添加一个简单验证：

```go
// In jwt_signer_test.go
func TestIssueAccessTokenAdmin(t *testing.T) {
    signer, _ := auth.GenerateDevSigner()
    token, err := signer.IssueAccessToken("user-1", "my-org", jwt.ClaimStrings{"tenant"}, true)
    require.NoError(t, err)
    claims, err := signer.ParsePlatformClaims(token)
    require.NoError(t, err)
    assert.True(t, claims.IsAdmin)
}

func TestIssueAccessTokenNonAdmin(t *testing.T) {
    signer, _ := auth.GenerateDevSigner()
    token, err := signer.IssueAccessToken("user-1", "my-org", jwt.ClaimStrings{"tenant"}, false)
    require.NoError(t, err)
    claims, err := signer.ParsePlatformClaims(token)
    require.NoError(t, err)
    assert.False(t, claims.IsAdmin)
}
```

- [ ] **Step 4: 最终 commit（如有未提交变更）**

```bash
cd /data/home/lukemxjia/modelcraft
git status
git add -A && git commit -m "chore: Plan 2b complete — unified auth service with is_admin token support"
```
