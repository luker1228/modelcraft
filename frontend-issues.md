# ModelCraft 前端问题报告

> 初次测试时间：2026-05-12  
> 第二轮测试时间：2026-05-12（修复验证）  
> 测试环境：http://localhost:3002  
> 测试方式：Playwright MCP 自动化测试 + 手工探索  
> 测试账号：testuser01 / 13800138000

---

## 修复状态总览

| # | 问题 | 状态 |
|---|------|------|
| BUG-001 | 登录密码错误页面崩溃 | ✅ 已修复 |
| BUG-002 | MSW mock 在生产环境激活 | ✅ 已修复 |
| BUG-003 | 终端用户登录后无跳转 | ⚠️ 后端 BFF 未返回 accessToken（后端问题） |
| BUG-004 | 数据库错误显示英文 | ✅ 已修复 |
| BUG-005 | 注册失败提示模糊 | ✅ 已修复 |
| BUG-006 | 用户菜单"设置"语义歧义 | ✅ 已修复（改为"组织设置"） |
| UI-001 | 登录字段标签与验证提示不一致 | ✅ 已修复 |
| UI-002 | 个人资料展示 mock:// URL | ✅ 已修复 |
| UI-003 | 昵称显示随机字符串 | ✅ 已修复（显示"未设置昵称"） |
| UI-004 | 账号状态英文未翻译 | ✅ 已修复（REGISTERED→已注册） |
| UI-005 | 开发者状态英文未翻译 | ✅ 已修复（active→正常）|
| UI-006 | 开发者角色英文未翻译 | ✅ 已修复（owner→所有者） |
| UI-007 | 终端用户创建人显示 UUID | ✅ 已修复（截断+tooltip） |
| UI-008 | 中文项目名转拼音 slug 无说明 | ✅ 已修复（提示文案更新） |
| UI-009 | Token Issuer 字段未翻译 | ✅ 已修复 |
| UI-010 | 日期时区问题 | ⏭️ 跳过（服务器时区问题，需后端配合） |
| UI-011 | favicon.ico 404 | ⏭️ 跳过（需设计资源） |
| UI-012 | 表单缺少 autocomplete 属性 | ✅ 已修复 |
| CODE-001 | Apollo onCompleted/onError 废弃用法 | ✅ 已修复 |
| NEW-001 | 登录错误信息仍含英文 "incorrect password" | 🆕 新发现 |
| NEW-002 | 终端用户 BFF 登录接口不返回 accessToken | 🆕 新发现（后端 Bug） |

---

## 🔴 严重问题（Crash / 阻断流程）

### BUG-001：登录密码错误时页面崩溃
**状态**：✅ 已修复

**修复方案**：`use-auth-form.ts` 新增 `extractErrorMessage()` 函数，确保从 API 响应中提取的错误值始终是字符串

**修复后效果**：输入错误密码点击登录，现在正确显示红色错误横幅，页面不再崩溃

---

### BUG-002：MSWProvider 在生产/部署环境中激活
**状态**：✅ 已修复

**修复方案**：`MSWProvider.tsx` 加 `process.env.NODE_ENV === 'production'` 双重保护

---

### BUG-003：终端用户登录成功后无跳转
**状态**：⚠️ 后端 Bug（前端已优化，但根本原因在后端）

**根本原因（新发现）**：后端 BFF `/api/bff/org/{orgName}/end-user/auth/login` 接口返回的 JSON 中**没有 `accessToken` 字段**：
```json
{"expiresAt":"2026-05-19T02:53:19Z","requestId":"...","userId":"..."}
```
缺少 `accessToken`，导致前端无法构建认证状态，跳转后 workspace 页 token 为空，触发 refresh 并返回 401，最终踢回登录页。

**需后端修复**：BFF 登录接口返回体补充 `accessToken` 字段

---

## 🆕 新发现问题

### NEW-001：登录错误信息含英文
**页面**：`/login`  
**现象**：输入错误密码，错误提示显示"认证失败: **incorrect password**"，后半段是英文  
**根因**：后端 BFF 返回的错误 `message` 字段本身是英文字符串  
**建议**：后端将常见认证错误映射为中文，或前端在 `extractErrorMessage` 后再做关键词翻译

### NEW-002：终端用户 BFF 登录接口不返回 accessToken（后端 Bug）
**接口**：`POST /api/bff/org/{orgName}/end-user/auth/login`  
**问题**：响应体缺少 `accessToken` 字段，只返回 `expiresAt`、`requestId`、`userId`  
**影响**：终端用户完全无法登录进入 workspace，整个终端用户体验链路断路

---

## 🟠 功能缺陷（影响核心使用）

### BUG-004：数据库连接失败时显示原始英文错误信息
**状态**：✅ 已修复  
**修复后效果**：
- `connection refused` → "连接被拒绝，请检查主机地址和端口"
- `authentication failed` → "认证失败，请检查用户名和密码"
- 其他 → 保持原始消息

---

### BUG-005：注册失败时错误提示过于模糊
**状态**：✅ 已修复（同 BUG-001 的 extractErrorMessage 改进）

---

### BUG-006：用户菜单"设置"跳转到组织设置而非个人设置
**状态**：✅ 已修复  
**修复方案**：菜单项文本改为"组织设置"，明确语义

---

## 🟡 UI/UX 问题

### UI-002：个人资料页展示内部 Mock 数据
**状态**：✅ 已修复  
- 头像地址 `mock://...` 显示为"未设置"
- 编辑页 placeholder 改为 `https://example.com/avatar.png`

### UI-003：个人资料显示名称为随机字符串
**状态**：✅ 已修复  
- 检测 `user_XXXXXX` 格式的自动生成昵称
- 显示为"未设置昵称"并提示用户"点击「编辑资料」完善昵称信息"

### UI-004/005/006：英文状态/角色未本地化
**状态**：✅ 已修复  
- 账号状态：`REGISTERED`→已注册，`ACTIVE`→正常，`SUSPENDED`→已停用
- 开发者状态：`active`→正常，`ACTIVE`→正常（大小写兼容）
- 角色：`owner`→所有者，`admin`→管理员，`editor`→编辑者，`viewer`→查看者

### UI-007：终端用户创建人显示原始 UUID
**状态**：✅ 已修复  
- 列表页：截断显示前8位 + `…`，鼠标悬停显示完整 UUID

### UI-008/009/012：其他 UI 细节
**状态**：✅ 已修复（slug 提示文案、Token Issuer 翻译、autocomplete 属性）

---

## 附录：截图
screenshots 保存在项目根目录，按问题编号命名。


## 🔴 严重问题（Crash / 阻断流程）

### BUG-001：登录密码错误时页面崩溃

**页面**：`/login`  
**复现步骤**：输入正确手机号 + 错误密码，点击登录  
**现象**：
- 前端弹出 Next.js `Unhandled Runtime Error` 弹窗
- 错误信息：`Objects are not valid as a React child (found: object with keys {code, message})`
- 关闭弹窗后页面变为空白，只剩 "1 error" 提示
- 用户无法继续操作，必须刷新页面

**根因**：登录失败后，错误响应对象 `{code, message}` 被直接传入 React 渲染树，未提取 `.message` 字符串

**影响**：所有输入错误密码的用户都会遇到页面崩溃，体验极差  
**截图**：`login-error-unhandled.png`

---

### BUG-002：MSWProvider 在生产/部署环境中激活

**页面**：所有页面  
**现象**：Console 日志中出现 `MSWProvider`，说明 Mock Service Worker 在当前环境中被加载  
**根因**：MSW 未正确通过环境变量限制为仅在 `development` 模式下启用  
**影响**：可能拦截真实 API 请求，导致数据异常；mock 数据可能混入真实数据展示

---

### BUG-003：终端用户登录成功后无跳转

**页面**：`/end-user/testuser01/login`  
**复现步骤**：以 enduser001 账号登录终端用户  
**现象**：点击登录后页面停留在登录页，无任何跳转或成功提示  
**控制台错误**：`401 Unauthorized @ /api/bff/org/testuser01/end-user/auth/refresh`  
**影响**：终端用户无法正常进入系统，登录功能不可用  
**截图**：`enduser-login-result.png`

---

## 🟠 功能缺陷（影响核心使用）

### BUG-004：数据库连接失败时显示原始英文错误信息

**页面**：创建项目弹窗第 2 步  
**现象**：测试数据库连接失败时，直接展示后端原始英文报错：  
`数据库连接失败: dial tcp [::1]:3306: connect: connection refused: Please verify host, port, username, and password are correct`  
**问题**：
1. 错误信息是英文，与中文界面不一致
2. 暴露了内部网络诊断信息（`dial tcp [::1]:3306`）给用户
3. 错误消息过长，阅读困难

**应改为**：简洁中文提示，如"无法连接到数据库，请检查主机地址、端口和账号信息"  
**截图**：`create-project-step2-error.png`

---

### BUG-005：注册失败时错误提示过于模糊

**页面**：`/register`  
**复现步骤**：用已注册的用户名/手机号重新注册  
**现象**：显示"注册失败，请稍后重试"，没有告知具体原因  
**应改为**："该用户名已被占用"或"该手机号已注册"

---

### BUG-006：用户菜单"设置"跳转到组织设置而非个人设置

**位置**：右上角用户头像菜单  
**现象**：点击用户下拉菜单中的"设置"选项，跳转到 `/org/testuser01/settings/general`（组织设置），而不是用户个人账号设置  
**期望**：应跳转到用户个人设置（修改密码、通知偏好等）  
**截图**：`user-menu.png`

---

## 🟡 UI/UX 问题

### UI-001：登录页手机号字段验证提示与字段标签不一致

**页面**：`/login`  
**现象**：字段 Label 是"手机号"，但空提交后错误提示是"请输入手机号**或用户名**"  
**问题**：用户看字段标签只知道输手机号，验证提示却说还可以输用户名，歧义明显  
**建议**：字段 Label 改为"手机号 / 用户名"，或在 placeholder 中注明两者均可

---

### UI-002：个人资料页展示内部 Mock 数据

**页面**：`/org/testuser01/profile`  
**现象**：
- 头像地址字段显示 `mock://avatar/default-1.png`（mock 协议地址，不是真实 URL）
- 编辑页头像地址 placeholder 是 `例如 /mocks/avatar-user.png`（开发时 mock 路径）

**影响**：向用户泄露系统内部实现细节，体验很差  
**截图**：`profile-page.png`、`profile-edit.png`

---

### UI-003：个人资料显示名称为随机字符串

**页面**：`/org/testuser01/profile`  
**现象**：用户注册时输入了用户名 `testuser01`，但个人资料页显示的昵称是系统随机生成的 `user_KMKQSU`  
**期望**：注册时应使用用户名作为初始昵称，或在注册流程中增加昵称字段

---

### UI-004：账号状态显示英文未翻译

**页面**：`/org/testuser01/profile`  
**现象**：账号状态显示 `REGISTERED`（全大写英文），与其他中文界面不一致  
**建议**：翻译为"已注册"或"正常"

---

### UI-005：开发者页面角色和状态未本地化

**页面**：`/org/testuser01/developers/members`  
**现象**：
- 角色列显示英文：`owner`
- 状态列显示英文：`active`

**建议**：翻译为中文，如"所有者"、"启用"

---

### UI-006：角色列表名称未翻译

**页面**：`/org/testuser01/developers/roles`  
**现象**：系统内置角色名称 `admin`、`editor`、`owner`、`viewer` 均为英文  
**建议**：添加中文显示名称，或在名称旁加括号注释中文  
**截图**：`roles-page.png`

---

### UI-007：终端用户列表创建人显示为原始 UUID

**页面**：`/org/testuser01/end-users`  
**现象**：
- 列表中创建人列显示完整 UUID（如 `019e17d6-3754-7efd-9fab-b07884f580fb`）
- 详情页创建人显示截断 UUID（如 `019e17d6…`）

**建议**：应显示创建人的用户名，UUID 可作为 tooltip 或完全隐藏  
**截图**：`end-users-page.png`、`enduser-detail.png`

---

### UI-008：创建项目时中文名称自动转拼音作为 slug

**页面**：创建项目弹窗第 1 步  
**现象**：输入中文项目名称"测试项目"，项目标识自动填充为 `ceshixiangmu`  
**问题**：拼音 slug 对中文用户不友好，且对应关系不直观  
**建议**：对中文名称，建议提示用户手动输入英文 slug，或默认生成 `project-001` 之类的通用标识

---

### UI-009：登录配置页 "Token Issuer" 字段标签未翻译

**页面**：`/org/testuser01/settings/login-settings`  
**现象**：`Token Issuer` 字段是英文，与其他中文字段不一致  
**建议**：翻译为"令牌颁发者"或添加中文说明

---

### UI-010：终端用户创建时间显示为昨天（时区问题疑似）

**页面**：`/org/testuser01/end-users`  
**现象**：当前时间是 2026-05-12，但今天新建的终端用户创建时间显示为 `2026/5/11`  
**疑因**：服务器时区与前端显示时区差异，或时间戳精度问题

---

### UI-011：`favicon.ico` 404

**页面**：所有页面  
**现象**：Console 报 `Failed to load resource: 404 @ http://localhost:3002/favicon.ico`  
**建议**：添加 favicon 或配置正确路径

---

### UI-012：表单输入框缺少 autocomplete 属性

**页面**：登录页、注册页  
**现象**：Console 报 `[DOM] Input elements should have autocomplete attributes (suggested: "current-password" / "new-password")`  
**影响**：浏览器无法智能填充密码，降低用户便利性，也是无障碍合规问题

---

## 🔵 代码质量 / 潜在风险

### CODE-001：Apollo Client useQuery 错误回调使用方式被废弃

**现象**：Console 警告 `An error occurred! onCompleted/onError callback sets local state...`  
**详情**：使用 `useQuery` 的 `onCompleted` 和 `onError` 回调中直接 `setState`，违反 Apollo Client 3.x 最佳实践  
**风险**：可能导致 React 渲染无限循环或状态不一致

---

### CODE-002：Next.js 版本过旧

**现象**：Console 提示 `Next.js (14.2.35) is outdated`  
**建议**：升级到最新稳定版，修复已知的安全和性能问题

---

## 附录：测试截图列表

| 截图文件 | 描述 |
|---|---|
| `workspace-home.png` | 工作区首页（空项目列表） |
| `create-project-step1.png` | 创建项目步骤1 |
| `create-project-step2.png` | 创建项目步骤2（数据库配置） |
| `create-project-step2-error.png` | 数据库连接失败错误展示 |
| `developers-page.png` | 开发者列表页 |
| `roles-page.png` | 角色列表页 |
| `end-users-page.png` | 终端用户列表页 |
| `create-enduser-dialog.png` | 新增终端用户弹窗 |
| `create-enduser-validation.png` | 终端用户表单校验提示 |
| `enduser-detail.png` | 终端用户详情页 |
| `settings-general.png` | 组织通用设置页 |
| `settings-login.png` | 登录配置页 |
| `user-menu.png` | 用户下拉菜单 |
| `profile-page.png` | 个人资料页 |
| `profile-edit.png` | 编辑个人资料页 |
| `enduser-login.png` | 终端用户登录页 |
| `enduser-login-result.png` | 终端用户登录后（无跳转） |
| `login-error-unhandled.png` | 登录密码错误导致页面崩溃 |
| `register-duplicate.png` | 注册重复账号的错误提示 |
