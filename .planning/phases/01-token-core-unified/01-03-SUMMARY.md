---
phase: 01-token-core-unified
plan: 03
subsystem: gateway/auth, gateway/proxy
tags: [gateway, jwt, es256, token-migration]
key-files:
  modified:
    - modelcraft-gateway/internal/proxy/handler.go
    - modelcraft-gateway/internal/auth/service.go
    - modelcraft-gateway/internal/config/config.go
    - modelcraft-gateway/cmd/gateway/main.go
metrics:
  tasks_completed: 2
  files_changed: 4
---

## 完成内容

### 任务 1：handler.go — EndUserGraphQLHandler 改用 ES256

- `VerifyEndUserAccessToken(tokenStr)` → `VerifyAccessToken(tokenStr)`
- `claims.Subject` → `claims.UserID`（`auth.Claims` 结构体直接有 `UserID` 字段）
- 更新函数注释说明现在使用 ES256（mc-platform issuer）

### 任务 2：废弃标注

**service.go**：
- `endUserJWTSecret` 字段：添加 Deprecated 注释
- `EndUserClaims` 结构体：添加 Deprecated 注释
- `VerifyEndUserAccessToken` 方法：添加 Deprecated 注释，说明迁移到 `VerifyAccessToken`

**config.go**：
- `EndUserJWTSecret` 字段：注释替换为 Deprecated 说明（保留字段保持向后兼容）

**main.go**：
- `auth.NewService` 调用中 `cfg.EndUserJWTSecret` 参数行添加 Deprecated 注释

## 偏差记录

无偏差。

## Self-Check

- [x] `cd modelcraft-gateway && go build ./...` 通过
- [x] `grep "VerifyEndUserAccessToken" modelcraft-gateway/internal/proxy/handler.go` 输出为空
- [x] `grep "VerifyAccessToken" modelcraft-gateway/internal/proxy/handler.go` 输出 3 行（含注释）
- [x] `grep -c "Deprecated" modelcraft-gateway/internal/auth/service.go` 输出 3
- [x] `grep -c "Deprecated" modelcraft-gateway/internal/config/config.go` 输出 1

**Self-Check: PASSED**

## 阶段 1 整体完成状态

| 需求 ID | 描述 | 实现状态 |
|---------|------|---------|
| TOKEN-01 | 统一 issuer `mc-platform`，新增 `scope` claim，废弃旧 issuer | ✓ Wave 1 (domain/auth) |
| TOKEN-02 | 平台管理员登录返回 `scope=org` JWT（`mc-platform`，ES256） | ✓ Wave 1+2 (token_service) |
| TOKEN-03 | 端用户登录返回 `scope=org` JWT（`mc-platform`，ES256，不再 HMAC） | ✓ Wave 2+3 (enduser auth + gateway) |

**所有 Token 核心统一需求（TOKEN-01/02/03）已完成实现。**

## 阶段 2 可以开始的条件

- 所有登录路径均产出 `mc-platform` issuer + `scope=org` ES256 JWT ✓
- 端用户 token 已从 HMAC 迁移至 ES256 ✓
- Gateway 使用 `VerifyAccessToken`（ES256）验证端用户 token ✓
- 下一步：实现 scope 强制路由校验中间件 + `/api/auth/exchange` 端点（阶段 2）
