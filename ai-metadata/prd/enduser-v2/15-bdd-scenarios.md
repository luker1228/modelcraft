# EndUser v2 验收标准（BDD 场景）

> 本文档定义 EndUser v2 改造的验收标准，以 BDD（Gherkin）格式描述核心场景，可直接用于自动化验收测试框架。

---

## 特性 1：EndUser 账号 Org 级管理

### 场景 1.1：Developer 在 Org 下创建 EndUser

```gherkin
Feature: Org 级 EndUser 账号管理

  Background:
    Given 存在 Org "demo-org"
    And 当前用户是 "demo-org" 的 Developer，具有 "end-user:create" 权限

  Scenario: 成功创建 EndUser
    When Developer 通过 Org GraphQL 调用 createOrgEndUser
      """
      mutation {
        createOrgEndUser(input: { username: "alice", password: "Secure@123" }) {
          user { id username isForbidden }
          error { __typename }
        }
      }
      """
    Then 返回创建成功的 EndUser：
      | 字段       | 值      |
      | username   | alice   |
      | isForbidden | false  |
    And 该用户存储在 Org "demo-org" 下（不绑定任何 Project）
    And 该用户尚无任何 Project 访问权

  Scenario: 用户名在同一 Org 内不可重复
    Given Org "demo-org" 已有 EndUser "alice"
    When Developer 再次创建 username 为 "alice" 的 EndUser
    Then 返回错误 EndUserAlreadyExists

  Scenario: 同名用户名在不同 Org 下可以并存
    Given Org "org-a" 已有 EndUser "alice"
    When Developer 在 Org "org-b" 下创建 username 为 "alice" 的 EndUser
    Then 创建成功，返回新 EndUser
    And 两个 "alice" 账号相互独立，不共享密码和权限
```

### 场景 1.2：禁用/启用 EndUser

```gherkin
  Scenario: Developer 禁用 EndUser
    Given Org "demo-org" 有 EndUser "bob"（状态：启用）
    When Developer 调用 updateOrgEndUserStatus(userId: "bob", isForbidden: true)
    Then 返回成功
    And EndUser "bob" 登录时返回 ACCOUNT_DISABLED 错误

  Scenario: Developer 重新启用 EndUser
    Given Org "demo-org" 有 EndUser "bob"（状态：禁用）
    When Developer 调用 updateOrgEndUserStatus(userId: "bob", isForbidden: false)
    Then 返回成功
    And EndUser "bob" 可以正常登录
```

### 场景 1.3：删除 EndUser

```gherkin
  Scenario: 删除 EndUser 会级联清理所有 Project 访问权
    Given Org "demo-org" 有 EndUser "charlie"
    And "charlie" 已被授权访问 Project "project-a" 和 "project-b"
    When Developer 调用 deleteOrgEndUser(userId: "charlie")
    Then 返回成功
    And end_user_project_access 中 charlie 的所有记录被删除（CASCADE）
    And charlie 的所有 refresh token 会话被清除（CASCADE）
```

---

## 特性 2：EndUser Project 访问控制

### 场景 2.1：授权 EndUser 访问 Project

```gherkin
Feature: Project 级 EndUser 访问控制

  Background:
    Given 存在 Org "demo-org"，Project "project-a"
    And Org 中已有 EndUser "alice"（无任何 Project 访问权）
    And 当前用户是 "project-a" 的 Developer，具有 "end-user-access:manage" 权限

  Scenario: 成功授权 EndUser 访问 Project
    When Developer 调用 grantEndUserProjectAccess
      """
      mutation {
        grantEndUserProjectAccess(input: {
          endUserId: "alice",
          projectSlug: "project-a",
          permissionBundleId: "bundle-readonly"
        }) {
          access { endUserId projectSlug permissionBundleId }
          error { __typename }
        }
      }
      """
    Then 返回成功，access 记录创建
    And alice 可以通过登录 → 选择 project-a 的流程进入该 Project

  Scenario: 重复授权同一 Project 应报错
    Given "alice" 已被授权访问 "project-a"
    When Developer 再次授权 "alice" 访问 "project-a"
    Then 返回错误 EndUserProjectAccessAlreadyExists

  Scenario: 授权不存在的 EndUser
    When Developer 授权不存在的 userId 访问 "project-a"
    Then 返回错误 EndUserNotFound

  Scenario: 同一 EndUser 可授权多个 Project，权限可不同
    When Developer 授权 "alice" 访问 "project-a"（bundle: readonly）
    And Developer 授权 "alice" 访问 "project-b"（bundle: editor）
    Then alice 的 Project 访问列表包含 project-a 和 project-b
    And 两个 Project 的 permissionBundleId 不同
```

### 场景 2.2：修改 Project 访问权限

```gherkin
  Scenario: Developer 修改 EndUser 的 PermissionBundle
    Given "alice" 已被授权访问 "project-a"（bundle: readonly）
    When Developer 调用 updateEndUserProjectAccess(endUserId: "alice", projectSlug: "project-a", permissionBundleId: "bundle-editor")
    Then 返回成功
    And alice 在 project-a 的 permissionBundleId 变为 "bundle-editor"
```

### 场景 2.3：撤销 Project 访问权

```gherkin
  Scenario: Developer 撤销 EndUser 的 Project 访问权
    Given "alice" 已被授权访问 "project-a"
    When Developer 调用 revokeEndUserProjectAccess(endUserId: "alice", projectSlug: "project-a")
    Then 返回成功
    And alice 的 Project 访问列表中不再包含 "project-a"
    And alice 现有的 "project-a" JWT 仍然有效直到过期（不强制失效）
    And alice 下次登录时 project-a 不再出现在 Project 选择列表中
```

---

## 特性 3：EndUser 登录流程（Org 级）

### 场景 3.1：有且仅有一个 Project 访问权时的登录

```gherkin
Feature: EndUser Org 级统一登录

  Background:
    Given 存在 Org "demo-org"，Project "project-a"
    And EndUser "alice" 绑定在 Org "demo-org"，密码为 "Secure@123"
    And "alice" 仅有一个 Project 访问权：project-a

  Scenario: 单 Project 直接登录
    When alice 访问 /u/demo-org/login 并输入正确密码
    Then BFF 返回：{ singleProject: true, projectSlug: "project-a", accessToken: "<jwt>" }
    And 前端直接跳转到 /u/demo-org/project-a/data
    And 不经过 select-project 页面
    And JWT 的 payload 包含 { orgName: "demo-org", projectSlug: "project-a", role: "end_user" }
```

### 场景 3.2：有多个 Project 访问权时的登录

```gherkin
  Background:
    Given EndUser "bob" 有两个 Project 访问权：project-a 和 project-b

  Scenario: 多 Project 登录后进入 Project 选择页
    When bob 访问 /u/demo-org/login 并输入正确密码
    Then BFF 返回：{ singleProject: false, projects: [{ slug: "project-a", name: "..." }, { slug: "project-b", name: "..." }] }
    And 前端跳转到 /u/demo-org/select-project
    And 页面显示 project-a 和 project-b 两个选项

  Scenario: bob 选择 project-a 后签发最终 JWT
    When bob 在 select-project 页面点击 project-a
    Then BFF 调用 /api/bff/org/demo-org/end-user/auth/select-project
    And 返回最终 accessToken（包含 projectSlug: "project-a"）
    And 设置 refresh token Cookie（path: /u/demo-org/project-a/）
    And 前端跳转到 /u/demo-org/project-a/data
```

### 场景 3.3：无 Project 访问权时的登录

```gherkin
  Background:
    Given EndUser "newuser" 刚在 Org "demo-org" 注册，尚未被授权任何 Project

  Scenario: 无 Project 访问权时登录报错
    When newuser 访问 /u/demo-org/login 并输入正确密码
    Then BFF 返回：{ error: { code: "NO_PROJECT_ACCESS", message: "您暂无项目访问权限，请联系管理员授权" } }
    And 页面不跳转，在登录页内显示该错误信息
    And 不签发任何 JWT 或 Cookie
```

### 场景 3.4：错误密码/账号不存在

```gherkin
  Scenario: 错误密码
    When alice 用错误密码尝试登录
    Then BFF 返回 INVALID_CREDENTIALS 错误
    And 不签发任何 JWT

  Scenario: 账号被禁用
    Given alice 的账号已被 Developer 禁用
    When alice 尝试登录
    Then BFF 返回 ACCOUNT_DISABLED 错误
    And 不签发任何 JWT
```

---

## 特性 4：EndUser 自注册

```gherkin
Feature: EndUser Org 级自注册

  Background:
    Given 存在 Org "demo-org"

  Scenario: EndUser 成功自注册
    When 未注册用户访问 /u/demo-org/login 并点击"注册"
    And 填写 username "newuser" 和符合强度要求的密码
    Then 账号创建成功，绑定在 Org "demo-org"
    And 不绑定任何 Project
    And 注册后自动登录
    And 因无 Project 访问权，显示错误："您暂无项目访问权限，请联系管理员授权"

  Scenario: 注册时用户名已被占用
    Given Org "demo-org" 已有 EndUser "alice"
    When 新用户尝试注册 username "alice"
    Then 返回 CONFLICT 错误，提示"用户名已存在"
```

---

## 特性 5：数据隔离（RLS 不变）

```gherkin
Feature: EndUser 数据隔离（RLS）

  Background:
    Given Project "project-a" 有两个 EndUser：alice 和 bob
    And 均被授权访问 project-a（RLS 基于 JWT 中的 userId 作为 owner）

  Scenario: EndUser 只能查询自己的数据
    Given alice 已登录 project-a，JWT 包含 { sub: "alice-id" }
    And 数据表中有 alice 的数据（owner = "alice-id"）和 bob 的数据（owner = "bob-id"）
    When alice 查询数据表
    Then 只返回 owner = "alice-id" 的行
    And bob 的数据不可见

  Scenario: 账号归属上移不影响 RLS
    Given v2 改造后，alice 的账号归属从 (org, project) 改为 (org)
    But alice 登录 project-a 后签发的 JWT 仍包含正确的 projectSlug 和 userId
    When alice 查询 project-a 的数据
    Then RLS 行为与 v1 完全相同
    And 数据隔离不受账号归属层级变更影响
```

---

## 特性 6：向后兼容性

```gherkin
Feature: 旧 URL 兼容性

  Scenario: 旧版 Project 级登录 URL 重定向
    When 用户访问旧版 URL /u/demo-org/project-a/login
    Then 服务器返回 301 重定向到 /u/demo-org/login
    And 不显示 404 页面

  Scenario: 现有 refresh token Cookie 在 v2 仍有效
    Given alice 在 v1 系统中已登录，持有有效的 refresh token Cookie
    When v2 上线后 alice 的 access token 过期
    And alice 发起 token 刷新请求
    Then BFF 能正常处理 v1 格式的 refresh token
    And 返回新的 access token
    And 不强制要求 alice 重新登录
```

---

## 验收检查清单

### 后端

- [ ] `end_user_users` 表不再含 `project_slug` 列
- [ ] 新表 `end_user_project_access` 创建成功
- [ ] Org GraphQL：`createOrgEndUser` / `listOrgEndUsers` / `updateOrgEndUserStatus` / `deleteOrgEndUser` 可用
- [ ] Project GraphQL：`grantEndUserProjectAccess` / `listEndUserProjectAccesses` / `updateEndUserProjectAccess` / `revokeEndUserProjectAccess` 可用
- [ ] 登录接口返回 Project 访问列表（不再直接返回单 Project JWT）
- [ ] 无 Project 访问权时返回 `NO_PROJECT_ACCESS` 错误
- [ ] JWT 结构不变：`{ endUserId, orgName, projectSlug, iss: "mc-enduser" }`
- [ ] RLS 行为与 v1 完全一致

### 前端

- [ ] `/u/[orgName]/login` 页面正常工作
- [ ] `/u/[orgName]/select-project` 页面正常工作（多 Project 场景）
- [ ] 单 Project 场景不经过 select-project 页，直接跳转
- [ ] 无权限时在登录页显示错误提示
- [ ] `/org/[orgName]/end-users` 展示 Org 级 EndUser 管理
- [ ] `/org/[orgName]/project/[projectSlug]/end-user-access` 展示 Project 访问控制
- [ ] 旧 `/u/[orgName]/[projectSlug]/login` 重定向到新 URL

### 数据迁移

- [ ] 现有 EndUser 数据正确迁移（用户名唯一性无冲突）
- [ ] 现有 Project 绑定关系迁移到 `end_user_project_access` 表
- [ ] Atlas 迁移成功执行，无数据丢失
