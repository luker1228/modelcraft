# 用户登录页详细设计

> **所属模块**：终端用户认证（End-User Auth）  
> **父文档**：[00-end-user-auth.md](./00-end-user-auth.md)  
> **优先级**：P0

---

## 用户故事

> 作为终端用户，我希望通过独立的登录入口使用用户名和密码登录，以便进入数据管理页操作数据，并且不看到任何项目构建或配置相关的界面。

---

## 页面路由

| 项目 | 值 |
|------|----|
| 前端路由路径 | `/org/[orgName]/project/[projectSlug]/user/login` |
| 完整示例 URL | `http://host:3000/org/acme/project/crm/user/login` |
| 登录成功跳转 | `/org/[orgName]/project/[projectSlug]/data` |
| 路由隔离要求 | 与开发者登录页 `/login` 完全独立，不共享任何布局组件或路由守卫逻辑 |

### 路由守卫行为

- **未登录用户**访问 `/org/.../data/*` 任意子页：重定向到对应项目的 `/user/login`，并携带 `?redirect=<原始路径>` 参数，登录成功后回跳。
- **已登录用户**（持有有效用户 Token）访问 `/user/login`：自动跳转到 `/data` 落地页，不显示登录表单。
- **开发者 Token** 不能用于访问用户侧路由，访问 `/data` 页时如果只携带开发者 Token 应跳转到用户登录页（而不是开发者工作区）。

---

## 页面结构

```
┌─────────────────────────────────────────────────────┐
│                    [Project Logo / Name]              │
│                                                       │
│              欢迎回来，请登录                         │
│                                                       │
│   ┌─────────────────────────────────────────────┐   │
│   │  用户名                                      │   │
│   │  ┌───────────────────────────────────────┐  │   │
│   │  │  (输入框)                              │  │   │
│   │  └───────────────────────────────────────┘  │   │
│   │                                              │   │
│   │  密码                                        │   │
│   │  ┌───────────────────────────────────────┐  │   │
│   │  │  (输入框，密码掩码，可切换明文)         │  │   │
│   │  └───────────────────────────────────────┘  │   │
│   │                                              │   │
│   │  [错误提示区域（可选显示）]                  │   │
│   │                                              │   │
│   │  ┌───────────────────────────────────────┐  │   │
│   │  │              登 录                    │  │   │
│   │  └───────────────────────────────────────┘  │   │
│   └─────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────┘
```

---

## 表单字段定义

### 用户名字段

| 属性 | 规则 |
|------|------|
| 标签 | 用户名 |
| 类型 | `text` |
| 占位符 | 请输入用户名 |
| 最大长度 | 64 字符 |
| 必填 | 是 |
| 前端校验 | 不可为空；不校验格式（格式由后端管理员控制） |
| 自动完成 | `autocomplete="username"` |

### 密码字段

| 属性 | 规则 |
|------|------|
| 标签 | 密码 |
| 类型 | `password`（默认）/ `text`（切换明文时） |
| 占位符 | 请输入密码 |
| 最大长度 | 128 字符 |
| 必填 | 是 |
| 前端校验 | 不可为空；不做强度检验（登录场景） |
| 明文切换 | 字段右侧眼睛图标，点击切换显示/隐藏 |
| 自动完成 | `autocomplete="current-password"` |

### 登录按钮

- 默认文案：**登录**
- 提交中状态：按钮禁用 + 显示加载动画，文案变为 **登录中…**
- 两个必填字段均非空时，按钮可点击

---

## 前端校验规则（提交前）

1. 用户名为空 → 显示「请输入用户名」，阻止提交
2. 密码为空 → 显示「请输入密码」，阻止提交
3. 以上校验在 `onBlur`（失焦）时触发，以及提交按钮点击时全量触发

---

## 登录流程（时序）

```
用户点击"登录"
    │
    ▼
前端表单校验（非空）
    │ 通过
    ▼
调用 BFF 接口
POST /org/{orgName}/project/{projectSlug}/user/auth/signin
{ username, password }
    │
    ├─ 200 OK → 存储用户 Token（Cookie / SessionStorage）→ 跳转 /data（或 redirect 参数目标）
    │
    ├─ 401 { code: "INVALID_CREDENTIALS" }
    │     → 显示「用户名或密码错误」（不区分用户名不存在 / 密码错误，防止用户枚举）
    │
    ├─ 403 { code: "ACCOUNT_DISABLED" }
    │     → 显示「该账号已被禁用，请联系管理员」
    │
    └─ 5xx / 网络错误
          → 显示「登录服务暂时不可用，请稍后重试」
```

---

## 错误提示规则

| 场景 | 展示位置 | 文案 |
|------|----------|------|
| 用户名为空（提交时） | 用户名字段下方 | 请输入用户名 |
| 密码为空（提交时） | 密码字段下方 | 请输入密码 |
| 账号不存在 / 密码错误 | 表单顶部红色 Banner | 用户名或密码错误，请重试 |
| 账号被禁用 | 表单顶部红色 Banner | 该账号已被禁用，请联系管理员 |
| 服务异常 | 表单顶部红色 Banner | 登录服务暂时不可用，请稍后重试 |

- 错误 Banner 出现时，密码字段自动清空，聚焦密码字段
- 错误文案不得泄露「账号是否存在」的信息（账号不存在与密码错误统一返回同样文案）

---

## 与开发者登录页的隔离要求

| 隔离维度 | 要求 |
|----------|------|
| 路由 | 独立路由文件，不复用 `/login` 的 Next.js Page 组件 |
| 布局 | 独立 Layout，不引入开发者侧的导航栏、侧边栏 |
| 组件 | 可复用基础 UI 组件（Button、Input 等来自 `@/components/ui`），但不复用开发者登录的业务组件 |
| Token 存储 | 使用独立的 Cookie / Storage key，不与开发者 Token 混用 |
| 守卫逻辑 | 路由守卫使用独立的 hook（如 `useUserAuth`），不复用开发者侧的 `useAuth` |
| 跳转逻辑 | 登录成功跳转 `/data`，而非开发者工作区；开发者登录成功不跳转到用户侧页面 |

---

## 前端 BFF 接口集成

前端调用 BFF Route Handler，BFF 内部代理到 Go Backend（不感知 Go 实现细节）：

```typescript
// src/web/hooks/auth/use-end-user-auth.ts（示意）
export async function endUserSignIn(
  orgName: string,
  projectSlug: string,
  username: string,
  password: string,
): Promise<{ accessToken: string }> {
  const res = await fetch('/api/bff/end-user/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ orgName, projectSlug, username, password }),
    credentials: 'same-origin',
  })
  if (!res.ok) {
    const err = await res.json()
    throw new EndUserAuthError(err.code)
  }
  return res.json()
}
```

Token 存储策略（与现有开发者认证体系保持一致）：
- **access token**：BFF 响应 body 返回，前端存入内存 store（参考现有 `useAuthStore`，新增 `useEndUserStore`）
- **refresh token**：BFF 写入 `end_user_refresh_token` HttpOnly Cookie（7d），与开发者 `refresh_token` cookie 使用不同 key，完全隔离
- access token 过期后，前端静默调用 `/api/bff/end-user/auth/refresh` 用 cookie 换新 token

---

## 页面元数据

```typescript
// Next.js 页面 metadata
export const metadata = {
  title: `登录 · ${projectName}`,
  robots: 'noindex, nofollow', // 用户登录页不应被搜索引擎收录
}
```

---

## 验收标准

| # | 场景 | 预期结果 |
|---|------|----------|
| AC-1 | 访问 `/org/acme/project/crm/user/login` | 显示独立用户登录页，无开发者导航栏 |
| AC-2 | 提交空表单 | 显示「请输入用户名」「请输入密码」，不发起网络请求 |
| AC-3 | 输入正确用户名密码登录 | 跳转到 `/org/acme/project/crm/data` |
| AC-4 | 携带 `?redirect=/org/acme/project/crm/data/records` 登录 | 登录成功后跳转到 `redirect` 指定路径 |
| AC-5 | 输入错误密码 | 显示「用户名或密码错误，请重试」，密码字段清空并聚焦 |
| AC-6 | 使用被禁用账号登录 | 显示「该账号已被禁用，请联系管理员」 |
| AC-7 | 已持有有效用户 Token 时再次访问登录页 | 自动跳转到 `/data`，不展示登录表单 |
| AC-8 | 开发者登录 `/login` 不受影响 | 开发者登录页功能正常，两个入口完全独立 |
| AC-9 | 用户 Token 访问 `/org/.../project/.../settings` | 返回 403，不可访问项目配置 |
| AC-10 | 登录按钮点击中（请求进行中） | 按钮禁用，显示加载状态，防止重复提交 |
