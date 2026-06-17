# User 与 EndUser 区别

> 适用范围：后端 context 命名、runtime 鉴权、RLS、PAT、`X-MC-Auth-*` 相关开发。

## 核心定义

ModelCraft 中有两套不同语义的用户标识，必须分开理解：

- `User`：ModelCraft **系统本身**的用户
- `EndUser`：只在 **runtime** 场景存在的终端用户概念，是 ModelCraft 的业务特色概念

两者不是一回事，禁止混用。

## 1. User 是什么

`User` 表示平台内部身份，用于：

- 登录管理端
- 查看个人 Profile
- 查看自己在 org 中的 membership
- 创建 Project、设计模型、管理平台资源
- 执行 tenant / developer 侧 GraphQL 和 HTTP 接口

可以简单理解为：**操作 ModelCraft 平台的人**。

## 2. EndUser 是什么

`EndUser` 不是平台管理身份，而是 runtime 里的业务终端身份，用于：

- 访问 runtime GraphQL
- 参与 RLS 过滤
- 作为 `END_USER_REF` 的 owner / subject
- 使用 end-user API token / PAT
- 从 `X-MC-Auth-*`、end-user access token、runtime PAT 中解析身份

可以简单理解为：**通过 ModelCraft 构建出来的业务系统的终端使用者**。

## 3. 最重要的边界

`EndUser` **只存在于 runtime 语义里**。

不应把 `EndUser` 扩散到以下链路：

- 管理端 Profile / Membership
- tenant / developer 管理接口
- 普通后台用户中心逻辑
- 平台内部 `User` 相关 Domain / App 命名

也不应把平台 `User` 误当成 runtime 的 `EndUser`。

## 4. context 命名约定

必须拆开两套 context 字段：

- `UserID`：系统内部用户 ID
- `EndUserID`：runtime 终端用户 ID

推荐约定：

```go
ctxutils.SetUserID(ctx, userID)
ctxutils.GetUserIDFromContext(ctx)

ctxutils.SetEndUserID(ctx, endUserID)
ctxutils.GetEndUserIDFromContext(ctx)
```

禁止做法：

```go
// 禁止：把 end-user 身份塞进 UserID
ctxutils.SetUserID(ctx, token.EndUserID)

// 禁止：runtime 代码从 UserID 里读 end-user 身份
endUserID, _ := ctxutils.GetUserIDFromContext(ctx)
```

## 5. Header / Token 来源约定

### 平台用户

- 来源：Gateway 注入的 `X-User-ID`
- 路径：`/graphql/*`（内部链路）
- 语义：系统内部 `UserID`

### Runtime EndUser

- 来源：
  - `X-MC-Auth-*`
  - end-user access token
  - end-user PAT
- 语义：`EndUserID`

### 设计时 EndUser 链路

- 来源：Gateway 注入的 `X-User-ID`（注意：此链路的 `X-User-ID` 携带的是 **end-user ID**，不是平台 user ID）
- 路径：`/end-user/graphql/*`
- 语义：`EndUserID`
- **禁止**同时设置 `UserID` — userId ≠ endUserId，两个身份体系不应混淆

这里的重点不是”有没有 userId”，而是**这个 ID 属于哪套身份体系**。

## 5a. ChiJWTAuthMiddleware: link-based identity routing

```go
// /end-user/*  → X-User-ID 是 end-user ID → SetEndUserID（不设 UserID）
// /graphql/*   → X-User-ID 是 tenant user ID → SetUserID（不设 EndUserID）
if strings.HasPrefix(r.URL.Path, “/end-user/”) {
    ctx = ctxutils.SetEndUserID(ctx, userID)
} else {
    ctx = ctxutils.SetUserID(ctx, userID)
}
```

## 5b. @hasPermission validateContext: dual identity fallback

`validateContext` 尝试 `UserID`（内部链路），失败时回退到 `EndUserID`（EndUser 链路）：

```go
userID, err = ctxutils.GetUserIDFromContext(ctx)
if err != nil || userID == “” {
    userID, err = ctxutils.GetEndUserIDFromContext(ctx)
}
```

## 5c. Link-based allowEndUser gating (no context leak)

`allowEndUser` 门禁通过 **独立的 handler 实例** 区分链路，不通过 context 传递链路信息：

- 内部链路 handler：`NewHasPermissionDirective` → `enforceEndUserGate: false`，忽略 `allowEndUser`，统一走 RBAC
- EndUser 链路 handler：`NewEndUserHasPermissionDirective` → `enforceEndUserGate: true`，执行 `allowEndUser` 门禁

路由层为两种链路选择不同 handler，下游无需感知链路类型。

## 6. 代码判断原则

看到下面这些关键词时，优先想到 `EndUserID`，而不是 `UserID`：

- runtime
- RLS
- owner
- `END_USER_REF`
- `X-MC-Auth-*`
- end-user PAT
- 终端用户数据权限

看到下面这些关键词时，优先想到 `UserID`：

- profile
- membership
- org 管理
- project 管理
- 模型设计
- 平台登录态

## 7. 一句话准则

如果这段逻辑是在“操作 ModelCraft 平台”，用 `UserID`。  
如果这段逻辑是在“代表业务终端用户访问 runtime 数据”，用 `EndUserID`。
