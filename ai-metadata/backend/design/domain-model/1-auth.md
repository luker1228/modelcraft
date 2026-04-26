# 1. 登录与认证

> 代码位置：`internal/domain/auth/`

## 概述

ModelCraft 不自建用户体系，认证委托给外部 IdP（当前为 AuthProvider）。系统只负责验证 JWT token 的合法性，并从中提取用户身份。

## 核心实体

### User
```
internal/domain/user/user.go

User
├── ID          string    // ModelCraft 内部 UUID
├── ExternalID  string    // 来自 JWT.sub（AuthProvider 用户 ID）
├── Name        string    // 来自 AuthProvider
└── Phone       string    // 来自 AuthProvider
```

User 的 `ExternalID` 是与外部 IdP 的绑定点，首次登录时自动创建本地 User 记录。

### ProjectAuthConfig
```
internal/domain/auth/project_auth_config.go

ProjectAuthConfig
├── OrgName      string
├── ProjectSlug  string
├── Provider     ProviderType   // auth_provider | keycloak | oidc
├── Enabled      bool
└── Config       map[string]interface{}  // provider 专属配置
```

每个 Project 可以独立配置认证提供者，用于运行态 GraphQL 的访问鉴权。

## 认证流程

```
客户端携带 JWT token
        │
        ▼
ModelCraft 验证 token 签名（使用 AuthProvider 公钥）
        │
        ▼
从 JWT Claims 提取 ExternalID / OrgName / 权限信息
        │
        ▼
查找或创建本地 User 记录（通过 ExternalID 关联）
        │
        ▼
注入 context，传递给下游
```

## 当前状态

- AuthProvider 集成：**完整实现**
- Keycloak / OIDC：**预留结构，未实现**（`ProviderType` 已定义）

## 相关文件

- `internal/domain/auth/modelcraft_claims.go` — JWT Claims 结构
- `internal/domain/auth/user_claims.go` — 用户身份提取
- `internal/domain/auth/project_auth_config.go` — 项目级认证配置
- `internal/domain/auth/config_validator.go` — 各 provider 配置校验
