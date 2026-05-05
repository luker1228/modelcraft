# 阶段 1：Token 核心统一 — 研究报告

**研究日期**：2026-05-05  
**技术域**：Go JWT 认证体系迁移  
**整体置信度**：HIGH（全部基于直接代码阅读，无假设）

---

<phase_requirements>
## 阶段需求

| ID | 描述 | 研究支撑 |
|----|------|----------|
| TOKEN-01 | 统一 JWT issuer 为 `mc-platform`；新增 `scope` 字段（`"org"`/`"project"`），废弃旧 issuer | 见第 2-4 节：issuer 引用清单 + 变更范围 |
| TOKEN-02 | 平台管理员 `POST /api/auth/login` 返回 `scope=org` JWT，格式与端用户一致 | 见第 3.1 节：JWTSigner 调用路径 |
| TOKEN-03 | 端用户 `POST /api/end-user/{orgSlug}/auth/login` 返回 `scope=org` JWT，不再用 `mc-enduser` issuer 和 HMAC | 见第 3.2 节：端用户 token 签发路径 |
</phase_requirements>

---

## 摘要

本阶段目标是**将两套平行 Token 体系合并为一**。调查发现：

1. **平台管理员 Token**（`mc-developer`）：由 `JWTSigner`（ES256）签发，调用链为 `POST /api/auth/login` → `TokenService.Login` → `JWTSigner.IssueAccessToken`。  
2. **端用户 Token**（`mc-enduser`）：由独立的 `endUserJWTIssuer`（**HMAC-SHA256**）签发，与平台管理员 Token 不是同一算法也不是同一结构。

**两套体系在算法上是异构的**：平台管理员用 ES256（ECDSA P-256），端用户用 HS256（HMAC-SHA256）。这是本阶段最重要的技术结论——端用户 token 迁移不只是改 `issuer` 字符串，还需要**切换签名算法**（HS256 → ES256）和**替换 claims 结构**。

**一次核心结论**：Gateway 侧（`modelcraft-gateway`）对 issuer 字符串**没有硬编码检查**，只做签名算法校验（ES256 或 HMAC），所以 gateway 只需同步升级端用户 token 的验证逻辑，不需要修改 issuer 字符串相关判断。

---

## 架构责任分配

| 能力 | 主要层 | 次要层 | 说明 |
|------|--------|--------|------|
| 定义 PlatformClaims 结构 & issuer 常量 | Domain（`auth`） | — | 领域层是 JWT 结构的唯一真相源 |
| 签发 Org Token（平台管理员） | App（`token_service`） | Domain（JWTSigner） | `IssueAccessToken` 调用链 |
| 签发 Org Token（端用户） | Infra（`endUserJWTIssuer`） | Domain（JWTSigner） | 需从 HMAC → ES256 替换 |
| `scope` 注入平台管理员登录 | App（`token_service`） | Interfaces（auth handler） | `OrgName` 已存在，加 `Scope` |
| `scope` 注入端用户登录 | Infra（`endUserJWTIssuer`）| App（`enduser auth service`） | 整个签发链路改造 |
| Validate 逻辑（issuer 检查） | Domain（`user_claims.go`、`modelcraft_claims.go`） | — | 改为检查 `IssuerPlatform` |
| RLS/Runtime 中间件 issuer 检查 | Interfaces/middleware | Domain（rls） | 改 scope 判断取代 issuer 判断 |
| Gateway issuer 验证 | Gateway（`auth/service.go`） | — | 端用户 token 算法从 HMAC → ES256 |

---

## 现有代码引用清单

### 生产代码中的旧 issuer 引用（共 10 处）

| 文件 | 行 | 内容 | 操作 |
|------|----|------|------|
| `internal/domain/auth/issuer.go` | 7-8 | `IssuerDeveloper = "mc-developer"`、`IssuerEndUser = "mc-enduser"` | **替换** 为 `IssuerPlatform = "mc-platform"` |
| `internal/domain/auth/issuer.go` | 14 | `IsValid()` 判断两个旧 issuer | **改为** 检查 `IssuerPlatform` |
| `internal/domain/auth/jwt_signer.go` | 69, 84 | 默认 issuer 硬编码为 `IssuerDeveloper` | **改为** `IssuerPlatform` |
| `internal/domain/auth/user_claims.go` | 33-34 | `Validate()` 硬检查 `IssuerDeveloper` | **改为** `IssuerPlatform` |
| `internal/domain/auth/modelcraft_claims.go` | 60-61 | `Validate()` 硬检查 `IssuerDeveloper` | **改为** `IssuerPlatform`（或废弃此结构，见下文） |
| `internal/domain/rls/end_user_identity.go` | 11, 16 | `IsEndUser()` 检查 `"mc-enduser"`，`IsDeveloper()` 检查 `"mc-developer"` | **改为** 检查 `scope` 字段 |
| `internal/interfaces/http/middleware/runtime_auth_middleware.go` | 23, 28, 89, 92, 95 | `IsEndUser()` 检查 `"mc-enduser"`，`IsDeveloper()` 检查 `"mc-developer"`，issuer 校验逻辑 | **整体改造**：issuer 改为 `mc-platform`，加 scope 校验 |
| `internal/interfaces/http/routes.go` | 168 | `endUserJWTClaims` 签发时 `Issuer: string(domainAuth.IssuerEndUser)` | **改为** `IssuerPlatform`（并换 ES256） |

### 测试代码中的旧 issuer 引用

- **零处**：经搜索 `internal/domain/auth/`、`tests-bdd/` 内所有 `*_test.go` 和 `*.feature` 文件，均未发现对 `mc-developer`、`mc-enduser`、`IssuerDeveloper`、`IssuerEndUser` 的直接字符串引用。
- BDD 步骤文件（`auth.steps.ts`、`end-user-auth.steps.ts`）无 issuer 相关断言。
- 无需专门修改测试代码中的 issuer 字符串（**但** 需要新增 `scope` 断言的 BDD 场景，属于阶段 5 范畴）。

---

## 变更范围评估

### 阶段 1 核心改动（5 个文件）

```
modelcraft-backend/
└── internal/domain/auth/
    ├── issuer.go               ← 新增 IssuerPlatform，废弃旧常量
    ├── user_claims.go          ← 新增 Scope/OrgName 字段，Validate() 改 issuer 检查
    └── jwt_signer.go           ← IssueAccessToken 新增 orgName/scope 参数
└── internal/interfaces/http/
    ├── routes.go               ← endUserJWTIssuer 改用 ES256 + PlatformClaims
    └── handlers/enduser/auth_handler.go  ← /me 端点改用 ES256 验证
```

### 连带影响（需同步修改但不属于 Token 域核心）

```
internal/interfaces/http/middleware/runtime_auth_middleware.go  ← issuer 校验改 scope 校验
internal/domain/rls/end_user_identity.go                        ← IsEndUser/IsDeveloper 改 scope 判断
```

### 不需要修改的文件

- `internal/middleware/chi_jwt_auth.go`：已只依赖 `X-User-ID` 头，与 issuer 无关
- `internal/domain/auth/modelcraft_claims.go`：此结构体实际**未被任何生产代码使用**（`JWTSigner.IssueAccessToken` 签发的是 `UserClaims` 而非 `ModelCraftClaims`）。本阶段可暂不修改或直接废弃。
- `modelcraft-gateway/`：**不需要**改 issuer 相关字符串。网关对 developer token 只做 ES256 签名校验（`VerifyAccessToken`，基于公钥），不检查 issuer 字符串。但 `VerifyEndUserAccessToken` 使用 HMAC，**需要同步切换为 ES256 验证**（见风险点 2）。
- BDD 测试文件：无旧 issuer 字符串引用，本阶段无需修改。

---

## 实现策略建议

### 策略一：新建 `PlatformClaims`，保留 `UserClaims` 作过渡（推荐）

```go
// internal/domain/auth/platform_claims.go（新文件）
type PlatformClaims struct {
    UserID  string `json:"user_id"`
    OrgName string `json:"org_name"`
    Scope   string `json:"scope"` // "org" | "project" | "service_key"（预留）
    jwt.RegisteredClaims
}

const (
    IssuerPlatform      Issuer = "mc-platform"
    TokenScopeOrg              = "org"
    TokenScopeProject          = "project"
    TokenScopeServiceKey       = "service_key" // 预留，本期不签发
)
```

**好处**：`UserClaims` 不删除，若后续有兼容需求可保留；新 claims 结构清晰独立。  
**代价**：多一个文件。

### 策略二：`JWTSigner.IssueAccessToken` 方法签名扩展

当前签名：`IssueAccessToken(userID, userName string) (string, error)`

目标签名：`IssueAccessToken(userID, orgName, scope string) (string, error)`

```go
func (s *JWTSigner) IssueAccessToken(userID, orgName, scope string) (string, error) {
    now := time.Now()
    claims := &PlatformClaims{
        UserID:  userID,
        OrgName: orgName,
        Scope:   scope,
        RegisteredClaims: jwt.RegisteredClaims{
            Issuer:    string(IssuerPlatform),
            Subject:   userID,
            IssuedAt:  jwt.NewNumericDate(now),
            ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
    return token.SignedString(s.privateKey)
}
```

**调用方需同步修改**：
- `internal/app/auth/token_service.go` 第 313 行：`IssueAccessToken(u.ID, u.Name)` → `IssueAccessToken(u.ID, orgName, TokenScopeOrg)`  
- `internal/app/auth/token_service.go` 第 446 行：`IssueAccessToken(token.UserID, "")` → `IssueAccessToken(token.UserID, orgName, TokenScopeOrg)`  
  - ⚠️ **注意**：refresh 场景需要从 refresh token 关联中找到 `orgName`，当前 `RefreshToken` 结构体没有 `OrgName` 字段，需要查询用户的 membership 取第一个 org（与 `Login` 现有逻辑一致）

### 策略三：端用户 token 迁移 — 用 JWTSigner 替换 endUserJWTIssuer

当前：`endUserJWTIssuer`（HMAC-SHA256）→ 独立签发，走 `jwt.SigningMethodHS256`

目标：复用同一个 `JWTSigner`（ES256）→ `IssueAccessToken(userID, orgName, TokenScopeOrg)`

**改动路径**：
1. `EndUserAuthAppService.tokenIssuer` 接口（`EndUserTokenIssuer`）改为接受 `JWTSigner` 注入
2. `endUserJWTIssuer` 的实现替换为对 `JWTSigner.IssueAccessToken` 的包装
3. `routes.go` 中的注入从 `&endUserJWTIssuer{secret: ...}` 改为传入已有 `jwtSigner`
4. `auth_handler.go` 中的 `parseEndUserJWT`（HMAC 验证，仅用于 `/me` 端点）改为 ES256 验证

### 策略四：runtime_auth_middleware.go 改造

当前逻辑：验证 issuer == `"mc-enduser"`  
目标逻辑：验证 issuer == `"mc-platform"` 且 scope 字段存在（阶段 1 只需验证 issuer，scope 强制校验属于阶段 2）

本阶段最小改动：
```go
// 改：issuer 检查
if issuer != string(auth.IssuerPlatform) {
    // 返回 401
}

// 从 claims 中取 user_id（不变）
// scope 字段本阶段读取但不做强制路由检查（属于阶段 2）
```

### 策略五：Gateway 端用户 token 验证迁移

**当前**：`VerifyEndUserAccessToken` 使用 HMAC-SHA256  
**目标**：端用户 token 改为 ES256 后，gateway 只需用同一个 `VerifyAccessToken`（已有，ES256 公钥验证）

具体做法：
- `EndUserGraphQLHandler` 中从 `VerifyEndUserAccessToken` 改为 `VerifyAccessToken`
- `Service` 结构体的 `endUserJWTSecret` 字段本阶段可保留（给后续清理阶段），也可直接移除

---

## 风险点与注意事项

### 风险 1：`tokenService.Refresh` 拿不到 orgName

**问题**：`Refresh` 路径（`token_service.go` 第 446 行）当前传 `IssueAccessToken(token.UserID, "")` 没有 orgName。切换到 `PlatformClaims` 后必须填 orgName，否则 token 中 `org_name` 为空。  
**解法**：与 `Login` 路径保持一致——从 membership 查用户的第一个 org（`membershipRepo.ListByUserWithDetails`）。该 repo 已注入 `TokenService`，无需新增依赖。  
**影响范围**：`internal/app/auth/token_service.go` 的 `Refresh` 方法。

### 风险 2：端用户 /me 端点 HMAC → ES256 迁移

**问题**：`auth_handler.go` 中 `EndUserMe` 端点使用 `parseEndUserJWT`（HMAC 验证），切换后必须改为 ES256 验证。  
**解法**：`AuthHandler` 目前注入了 `jwtSecret []byte`。方案 A：改注入 `JWTSigner` 公钥并用标准 `jwt.ParseWithClaims` + ES256；方案 B：AuthHandler 完全不做 token 验证，让请求先经过 RuntimeAuthMiddleware（但需路由调整）。  
**推荐方案 A**：AuthHandler 注入 `auth.JWTSigner`（或其 `Verifier` 接口），复用现有公钥。  
**影响范围**：`handlers/enduser/auth_handler.go` + `routes.go` 的注入处。

### 风险 3：endUserClaims 和 endUserJWTClaims 两个地方重复定义了 claims 结构

- `internal/interfaces/http/routes.go`：`endUserJWTClaims`（签发用，HMAC）
- `internal/interfaces/http/handlers/enduser/auth_handler.go`：`endUserClaims`（验证用，HMAC）

迁移后两个结构都要废弃，统一用 `PlatformClaims`。注意 `endUserClaims` 包含 `OrgName` 字段但没有 `scope`，切换后 `EndUserMe` 的 orgName 从 `PlatformClaims.OrgName` 取（不变），`user_id` 从 `UserID` 取。

### 风险 4：`modelcraft_claims.go` 僵尸代码

`ModelCraftClaims` 结构体（复杂，含 Memberships、Permissions 等字段）的 `Validate()` 方法硬检查 `IssuerDeveloper`，但调查发现**该结构体在生产代码中未被任何处使用**（`JWTSigner.IssueAccessToken` 内部用 `UserClaims`，没有用 `ModelCraftClaims`）。  
**建议**：本阶段可只更新 `Validate()` 中的 issuer 字符串（不删除此文件），或直接添加 `// Deprecated` 注释，避免回归风险。

### 风险 5：`rls.EndUserIdentity` 与 `middleware.EndUserIdentity` 重复定义

**现象**：`internal/domain/rls/end_user_identity.go` 和 `internal/interfaces/http/middleware/runtime_auth_middleware.go` 各自定义了同名结构 `EndUserIdentity`，且都有 `IsEndUser()`/`IsDeveloper()` 方法，判断依据是 issuer 字符串。  
**本阶段改动**：两处都需要把 issuer 检查改为 scope 判断（或新增 scope 字段）。但 `IsDeveloper()/IsEndUser()` 的语义在统一 token 体系后已不准确——所有用户都用同一 issuer，只有 scope 区分。  
**建议**：本阶段改为检查 `scope == "org"` / `scope == "project"`，并重命名方法（`HasOrgScope()`/`HasProjectScope()`）以匹配新语义。但如果 Runtime 下游有大量依赖 `IsEndUser()` 的代码，本阶段可只改 issuer 字符串，方法重命名放阶段 2/3。

### 风险 6：硬切策略下的 1h 窗口

**PROJECT.md 决策**：硬切，无兼容期。旧 token 自然在 1h 内过期。  
**实施注意**：部署时需保证前后端同步上线（或接受最多 1h 401 期）。这是架构决策，不是代码 bug，但规划任务时需标注"部署顺序依赖"。

---

## 代码结构示例

### PlatformClaims 签发示例（[VERIFIED: 代码库直接读取]）

```go
// 平台管理员登录
accessToken, err := s.jwtSigner.IssueAccessToken(u.ID, orgName, auth.TokenScopeOrg)

// 端用户登录
accessToken, err := jwtSigner.IssueAccessToken(user.ID, cmd.OrgName, auth.TokenScopeOrg)
```

### Gateway 端用户 token 验证切换示例（[VERIFIED: 代码库直接读取]）

```go
// 改前（EndUserGraphQLHandler）
claims, err := h.authService.VerifyEndUserAccessToken(tokenStr)  // HMAC

// 改后（复用同一 ES256 验证）
claims, err := h.authService.VerifyAccessToken(tokenStr)  // ES256
```

---

## 验证架构（Validation Architecture）

### 测试框架

| 属性 | 值 |
|------|----|
| Go 单元测试 | `go test ./...`（内建） |
| BDD | Cucumber.js（`tests-bdd/`），命令：`just bdd` 或 `npx cucumber-js` |
| 集成测试 | `tests/design/`（pytest），命令：`just test-integration` |
| 快速运行 | `just test-unit-pkg PKG=./internal/domain/auth/...` |

### 需求 → 测试映射

| 需求 ID | 行为 | 测试类型 | 建议命令 |
|---------|------|----------|----------|
| TOKEN-01 | `PlatformClaims` 签发带 issuer=`mc-platform` 和 scope 字段 | 单元测试 | `go test ./internal/domain/auth/...` |
| TOKEN-01 | 旧 issuer token 被 `Validate()` 拒绝 | 单元测试 | `go test ./internal/domain/auth/...` |
| TOKEN-02 | `POST /api/auth/login` 返回 JWT，`iss=mc-platform`，`scope=org` | BDD / 集成 | `tests-bdd/features/auth/login.feature` |
| TOKEN-03 | `POST /api/end-user/{orgSlug}/auth/login` 返回同格式 JWT | BDD / 集成 | `tests-bdd/features/end-user-auth/end-user-auth.feature` |

### Wave 0 缺口

- [ ] `internal/domain/auth/platform_claims_test.go` — 覆盖 TOKEN-01（PlatformClaims 签发和 Validate）
- [ ] `internal/domain/auth/jwt_signer_test.go` — 覆盖 IssueAccessToken 新签名

---

## 安全域

| ASVS 分类 | 适用 | 标准控制 |
|-----------|------|----------|
| V2 认证 | 是 | ES256 JWT，`IssuerPlatform` 签名校验 |
| V3 会话管理 | 是 | refresh token（opaque）+ httpOnly cookie |
| V5 输入验证 | 是 | claims 字段非空校验（`Validate()`） |
| V6 密码学 | 是 | 已用 ES256（ECDSA P-256），禁止手写 HMAC 替代 |

**本阶段安全注意**：端用户 token 从 HMAC（对称）切换到 ES256（非对称）是**安全加强**，不是降级。原 HMAC secret 在 gateway 和 backend 两侧都持有，存在泄露扩散风险；ES256 私钥只在 backend 持有，gateway 只需公钥验证。

---

## 来源清单

### 主要来源（HIGH 置信度，代码直接读取）

- `modelcraft-backend/internal/domain/auth/issuer.go` — issuer 常量定义
- `modelcraft-backend/internal/domain/auth/user_claims.go` — 当前 claims 结构和 Validate 逻辑
- `modelcraft-backend/internal/domain/auth/jwt_signer.go` — IssueAccessToken 实现
- `modelcraft-backend/internal/domain/auth/modelcraft_claims.go` — ModelCraftClaims（僵尸代码分析）
- `modelcraft-backend/internal/app/enduser/end_user_auth_service.go` — 端用户登录路径
- `modelcraft-backend/internal/interfaces/http/routes.go` — endUserJWTIssuer（HMAC 签发）
- `modelcraft-backend/internal/interfaces/http/handlers/enduser/auth_handler.go` — endUserClaims（HMAC 验证）
- `modelcraft-backend/internal/interfaces/http/middleware/runtime_auth_middleware.go` — issuer 校验中间件
- `modelcraft-backend/internal/domain/rls/end_user_identity.go` — RLS 层 issuer 判断
- `modelcraft-backend/internal/middleware/chi_jwt_auth.go` — 设计时 JWT 中间件（已无 issuer 检查）
- `modelcraft-gateway/internal/auth/service.go` — gateway token 验证逻辑
- `modelcraft-gateway/internal/proxy/handler.go` — gateway proxy 路由逻辑
- `modelcraft-backend/pkg/config/config.go` — JWTConfig 结构（`Issuer` 配置项）
- `.planning/PROJECT.md`、`REQUIREMENTS.md`、`ROADMAP.md` — 决策上下文
- `plans/unified-token-system.md` — 技术规格文档

### 假设项（Assumptions Log）

| # | 假设 | 章节 | 错误影响 |
|---|------|------|----------|
| A1 | `ModelCraftClaims` 在生产代码中未被实际使用（搜索未发现调用，但有可能被反射使用） | 变更范围 | 低：即使被用到，修改 issuer 字符串同样能修复 |
| A2 | Refresh 路径的 orgName 从 membership 查询（与 Login 相同逻辑）不会有性能问题 | 风险 1 | 低：一次额外 DB 查询，与现有 Login 路径相同 |

---

## RESEARCH COMPLETE

**阶段**：1 — Token 核心统一  
**整体置信度**：HIGH

### 关键发现

1. **旧 issuer 引用共 10 处**，全部在 `modelcraft-backend/`，无测试文件包含旧 issuer 字符串
2. **端用户 token 是 HMAC（HS256）**，平台管理员 token 是 ES256，两者算法不同；迁移需同时换算法 + 换 issuer
3. **Gateway 不检查 issuer 字符串**，只验证签名算法；端用户 GraphQL handler 使用 HMAC 验证需切换为 ES256
4. **`ModelCraftClaims` 是僵尸代码**，实际签发使用 `UserClaims`，本阶段可安全忽略或仅改 issuer 字符串
5. **核心改动文件 5 个**（全在 domain/auth 和 interfaces/http），无数据库迁移需求，refresh_token 表无 issuer 字段

### 置信度矩阵

| 域 | 置信度 | 原因 |
|----|--------|------|
| 旧 issuer 引用清单 | HIGH | 代码库全量 grep，零假设 |
| 变更范围 | HIGH | 调用链追踪到最终签发点 |
| Gateway 改动范围 | HIGH | gateway 源码直接读取 |
| 实现策略 | MEDIUM | 策略选择有多种，方案推荐基于代码结构，未经用户确认 |

### 开放问题

1. **`EndUserTokenIssuer` 接口**：阶段 1 后 `EndUserTokenIssuer` 接口是否继续保留？若端用户登录改为直接调用 `JWTSigner.IssueAccessToken`，此接口可废弃（属于 SCHEMA 清理阶段工作）。本阶段建议保留接口，仅替换其实现。
2. **`endUserClaims.OrgName` 字段的 /me 端点**：`EndUserMe` 用 HMAC 解析 token 取 OrgName，切换 ES256 后改从 `PlatformClaims.OrgName` 取，语义一致。但 `/me` 接口是否属于阶段 1 范围？如果端用户 /me 不经过 RuntimeAuthMiddleware，需要独立改造。

### 规划就绪

研究完成，规划者可依此创建 PLAN.md。
