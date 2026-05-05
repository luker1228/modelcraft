---
phase: 01-token-core-unified
plan: 02
subsystem: app/auth, interfaces/http
tags: [jwt, es256, enduser, token-migration, middleware, rls]
key-files:
  modified:
    - modelcraft-backend/internal/interfaces/http/routes.go
    - modelcraft-backend/internal/interfaces/http/handlers/enduser/auth_handler.go
    - modelcraft-backend/internal/interfaces/http/middleware/runtime_auth_middleware.go
    - modelcraft-backend/internal/domain/rls/end_user_identity.go
    - modelcraft-backend/internal/app/auth/token_service.go
    - modelcraft-backend/internal/app/auth/token_service_test.go
metrics:
  tasks_completed: 2
  files_changed: 6
  tests_fixed: 5
---

## 完成内容

### 任务 1：token_service.go 调用更新

- Login 路径（第 313 行）：`IssueAccessToken(u.ID, orgName, TokenScopeOrg)` ✓（Wave 1 中已完成）
- Refresh 路径（第 455 行）：新增 membership 查询获取 `refreshOrgName`，然后调用 `IssueAccessToken(token.UserID, refreshOrgName, TokenScopeOrg)` ✓
- `TokenService` 的 `membershipRepo` 字段类型改为精简接口 `MembershipOrgProvider`（只含 `ListByUserWithDetails`），减少外部依赖
- 测试中新增 `mockMembershipRepo` 并注入 `test-org`，修复了 5 个 Login/Refresh/Logout 相关测试失败

### 任务 2：接口层全面迁移

**routes.go**：
- 删除 `endUserJWTClaims` 结构体（HMAC 专用）
- `endUserJWTIssuer` 替换为 ES256 实现（持有 `*domainAuth.JWTSigner`，调用 `IssueAccessToken(userID, orgName, TokenScopeOrg)`）
- `endUserAuthHandler` 注入点：`[]byte(cfg.JWT.Secret)` → `jwtSigner`
- 移除未使用的 `sort` 和 `jwt` import

**auth_handler.go**：
- `AuthHandler.jwtSecret []byte` → `jwtSigner *domainAuth.JWTSigner`
- `NewAuthHandler` 参数：`jwtSecret []byte` → `jwtSigner *domainAuth.JWTSigner`
- 删除 `endUserClaims` 结构体（HMAC Claims）
- `parseEndUserJWT` 完全替换：HMAC 验证 → ES256 公钥验证，返回 `*PlatformClaims`
- 新增 `crypto/x509`、`encoding/pem`、`domainAuth` import

**runtime_auth_middleware.go**：
- `EndUserIdentity` 新增 `Scope string` 字段
- `IsEndUser()` / `IsDeveloper()` 改为检查 `issuerPlatform` 常量（`"mc-platform"`）
- issuer 校验：`"mc-enduser"` → `issuerPlatform`
- 注入 context 时同时传入 `scope` 字段（从 claims 取）

**rls/end_user_identity.go**：
- `EndUserIdentity` 新增 `Scope string` 字段
- `IsEndUser()` 改为 scope 判断：`issuer=="mc-platform" && (scope=="org"||scope=="project")`
- `IsDeveloper()` 改为 `issuer=="mc-platform" && scope=="org"`，标注 Deprecated

## 偏差记录

- `MembershipOrgProvider` 接口是新增的精简接口（非计划原本内容），原计划保留 `membership.MembershipRepository` 类型；改动原因：让测试更简单，同时也更符合接口隔离原则

## Self-Check

- [x] `go build ./...` 通过
- [x] `go test ./internal/app/auth/...` 19 个测试全部 PASS
- [x] `just lint` 通过
- [x] 端用户 token 签发路径：HMAC → ES256
- [x] `routes.go` 中无 `endUserJWTClaims`、无旧 `sort`/`jwt` import
- [x] `auth_handler.go` 中无 `jwtSecret`、无 `endUserClaims`、无 HMAC 验证
- [x] `runtime_auth_middleware.go` 中无 `mc-enduser`
- [x] `rls/end_user_identity.go` 中无 `mc-enduser`/`mc-developer`

**Self-Check: PASSED**

## Wave 3 接口契约（供 01-PLAN-03 Gateway 迁移引用）

端用户 token 现在是 **ES256 签名的 PlatformClaims**（`mc-platform` issuer），Gateway 的 `VerifyEndUserAccessToken`（HMAC）必须改为 `VerifyAccessToken`（ES256 公钥验证）才能接受新格式的端用户 token。
