# Runtime JWT 认证（iss 区分 + 401 保护）

> 依赖：无（可与 `01-enduserref-field.md` 并行开发）
> 对应主 PRD 章节：M3

---

## 背景

Runtime GraphQL endpoint 是终端用户（EndUser）直接调用的入口。为了保证 RLS 可靠生效，**Runtime 只能接受 EndUser JWT**，必须拒绝开发者 JWT。

两类 JWT 通过 `iss`（issuer）字段区分：

| JWT 类型 | `iss` 值 | 可访问端点 |
|---------|---------|----------|
| Developer JWT | `mc-developer` | 设计态 GraphQL（管理后台） |
| EndUser JWT | `mc-enduser` | Runtime GraphQL |

> ⚠️ 现有 `iss: "modelcraft"` 的 Developer JWT 需同步迁移到 `mc-developer`（已确认，见主 PRD Q1）。

---

## 需求

### Runtime endpoint 认证规则

- Runtime endpoint 仅接受 `iss = mc-enduser` 的 JWT
- 收到 `iss = mc-developer` 的 JWT → 返回 **401 Unauthorized**，拒绝访问
- 收到无 JWT 或格式非法的请求 → 返回 **401 Unauthorized**
- 认证通过后，从 JWT 中提取 `endUserId`，注入 request context，供后续 RLS 注入使用

### 设计态 GraphQL 认证（不变）

- 设计态 endpoint 只接受 `iss = mc-developer`（现有逻辑，本期完成迁移）
- EndUser JWT 调用设计态 → 401（向后兼容，不在本期新增，确认现有行为即可）

### 开发者调试 Runtime 的方式

- 开发者**不存在**绕过 Runtime JWT 检查的模式
- 调试方式：在管理后台创建测试 EndUser 账号，用该账号登录拿到 EndUser JWT，再调用 Runtime
- Runtime 不接受 Developer JWT（无例外）

---

## 验收标准

### AC-6（部分）：Runtime 只接受 EndUser JWT

- [ ] Developer JWT（`iss = mc-developer`）调用 Runtime → 401，拒绝访问
- [ ] EndUser JWT（`iss = mc-enduser`）调用 Runtime → 认证通过，正常执行
- [ ] 无 JWT 或 JWT 格式非法 → 401

### AC-6（部分）：EndUser ID 注入

- [ ] 认证通过后，`endUserId` 从 JWT 中正确提取并注入 request context
- [ ] 后续 RLS 注入逻辑可从 context 中读取 `endUserId`

---

## 用户故事对应

**Story 3**（部分）：终端用户只能访问自己的数据
> EndUser 持有合法 EndUser JWT 才能调用 Runtime。

**Story 4**：开发者通过 EndUser 身份调试 Runtime
> Developer JWT 直接调用 Runtime → 401。
> 开发者用测试 EndUser 账号登录，拿到 EndUser JWT → 正常调用，RLS 生效。

---

## 领域模型关键元素

```
EndUserIdentity (Value Object)
  + endUserId: EndUserID
  + issuer: Issuer
  + isEndUser(): Boolean → issuer == MC_END_USER

Issuer (Enum)
  MC_DEVELOPER  ← iss = "mc-developer"
  MC_END_USER   ← iss = "mc-enduser"

RuntimeAuthPolicy (Domain Service)
  + validate(identity: EndUserIdentity): ValidationResult
    → issuer != MC_END_USER → 401 Unauthorized
    → issuer == MC_END_USER → 提取 endUserId 注入 context
  + extractFromJWT(token: String): EndUserIdentity
```

---

## 不做什么（本子页 Out of scope）

- WHERE 自动注入逻辑（见 `04-runtime-rls-injection.md`）
- EndUser 账号管理 UI（管理后台现有能力，不在本期新增）
- 开发者角色级访问控制（独立方向，Out of scope）
