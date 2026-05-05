---
phase: 01-token-core-unified
plan: 01
subsystem: domain/auth
tags: [jwt, claims, issuer, token-unification]
key-files:
  created:
    - modelcraft-backend/internal/domain/auth/platform_claims.go
    - modelcraft-backend/internal/domain/auth/platform_claims_test.go
    - modelcraft-backend/internal/domain/auth/jwt_signer_test.go
  modified:
    - modelcraft-backend/internal/domain/auth/issuer.go
    - modelcraft-backend/internal/domain/auth/jwt_signer.go
    - modelcraft-backend/internal/domain/auth/user_claims.go
    - modelcraft-backend/internal/domain/auth/modelcraft_claims.go
    - modelcraft-backend/internal/app/auth/token_service.go
metrics:
  tasks_completed: 2
  tests_added: 21
  files_changed: 8
---

## 完成内容

### 任务 1：新建 platform_claims.go

创建了 `PlatformClaims` 结构体（`user_id`、`org_name`、`scope` 字段）及 `TokenScopeOrg / TokenScopeProject / TokenScopeServiceKey` 常量。新增配套测试 `platform_claims_test.go`（8 个子测试全覆盖 Validate 逻辑）。

### 任务 2：统一 issuer 常量，迁移签发路径

- **issuer.go**：新增 `IssuerPlatform = "mc-platform"`，旧常量 `IssuerDeveloper / IssuerEndUser` 保留并标注 `Deprecated`，`IsValid()` 只接受 `IssuerPlatform`
- **jwt_signer.go**：`IssueAccessToken` 签名变更为 `(userID, orgName, scope string)`，内部签发 `PlatformClaims`（ES256），issuer 固定为 `IssuerPlatform`；`GenerateDevSigner` 和 `NewJWTSignerFromPEM` fallback 均改为 `IssuerPlatform`
- **user_claims.go**：`Validate()` 改为检查 `IssuerPlatform`
- **modelcraft_claims.go**：`Validate()` 改为检查 `IssuerPlatform`，文件头部添加 `Deprecated` 注释
- **token_service.go**（Login + Refresh）：更新两处 `IssueAccessToken` 调用至新签名（orgName + `TokenScopeOrg`）；Refresh 路径新增 membership 查询取 orgName

## 偏差记录

- token_service.go 的两处调用（Line 313, 446）原属 Wave 2 范围，为保持编译通过提前在 Wave 1 同步修复，Wave 2 计划中对应步骤可跳过

## Self-Check

- [x] `go build ./...` 通过
- [x] `go test ./internal/domain/auth/...` 21 个测试全部 PASS
- [x] `just lint` 通过
- [x] `IssuerPlatform = "mc-platform"` 已在 issuer.go
- [x] `PlatformClaims` 含 `UserID / OrgName / Scope` 三字段
- [x] `IssueAccessToken(userID, orgName, scope string)` 新签名生效
- [x] `user_claims.go` 和 `modelcraft_claims.go` 均检查 `IssuerPlatform`

**Self-Check: PASSED**

## Wave 2 接口契约（供 01-PLAN-02 引用）

```go
// IssueAccessToken 新签名
func (s *JWTSigner) IssueAccessToken(userID, orgName, scope string) (string, error)

// PlatformClaims 结构
type PlatformClaims struct {
    UserID  string `json:"user_id"`
    OrgName string `json:"org_name"`
    Scope   string `json:"scope"`
    jwt.RegisteredClaims
}

// Scope 常量
const (
    TokenScopeOrg        = "org"
    TokenScopeProject    = "project"
    TokenScopeServiceKey = "service_key"
)

// Issuer 常量
IssuerPlatform Issuer = "mc-platform"
```
