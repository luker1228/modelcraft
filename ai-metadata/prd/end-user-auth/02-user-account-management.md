# 开发者管理终端用户账号

> **所属模块**：终端用户认证（End-User Auth）
> **父文档**：[00-end-user-auth.md](./00-end-user-auth.md)
> **优先级**：P0

---

## 用户故事

> 作为开发者，我希望在项目的用户管理页创建、查看、禁用和删除终端用户账号，以便控制哪些用户可以访问数据管理页。

---

## 入口路径

```
/org/{orgName}/project/{projectSlug}/end-users
```

独立页面，与项目主导航并列。所有操作经 **BFF 代理到 Go Backend**，Go Backend 直接操作 `mc_private`。

---

## BFF → Go Backend 接口映射

| 操作 | BFF Route | Go 内部接口 |
|------|-----------|------------|
| 列表 | `GET /api/bff/end-users` | `GET /internal/end-users` |
| 创建 | `POST /api/bff/end-users` | `POST /internal/end-users` |
| 禁用/启用 | `PATCH /api/bff/end-users/{userId}/status` | `PATCH /internal/end-users/{userId}/status` |
| 删除 | `DELETE /api/bff/end-users/{userId}` | `DELETE /internal/end-users/{userId}` |

---

## 页面设计（最小 MVP）

```
┌────────────────────────────────────────────────────────────┐
│  用户管理                              [+ 创建用户]         │
├────────────────────────────────────────────────────────────┤
│  搜索用户名...                                              │
├────────────────┬──────────┬────────────────┬───────────────┤
│  用户名        │  状态    │  创建时间      │  操作          │
├────────────────┼──────────┼────────────────┼───────────────┤
│  alice         │ ● 启用   │ 2026-04-10     │ 禁用  删除    │
│  bob           │ ○ 禁用   │ 2026-04-12     │ 启用  删除    │
└────────────────┴──────────┴────────────────┴───────────────┘
```

---

## 创建用户

点击「+ 创建用户」弹 Modal：

| 字段 | 规则 |
|------|------|
| 用户名 | 3–64 字符，`^[a-zA-Z0-9_-]+$`，同一 Project 内唯一 |
| 初始密码 | 至少 8 位，包含字母 + 数字 |
| 确认密码 | 与初始密码一致 |

**后端错误映射**：

| Go 返回 | 前端行为 |
|---------|---------|
| 201 | 关闭 Modal，列表刷新，Toast「创建成功」 |
| 409 | Modal 内提示「该用户名已被使用」 |
| 400 | 显示字段级错误 |

---

## 禁用 / 启用

- **禁用**：二次确认 → `PATCH { isForbidden: true }` → Go 更新 `mc_private.users.is_forbidden=1`
- **启用**：无需确认 → `PATCH { isForbidden: false }`
- 禁用后 access token 在 1h 内自然过期（MVP 可接受）；下次登录返回 403

---

## 删除

确认对话框 → `DELETE` → Go 删除 `mc_private.users` 记录（同时 revoke 所有 accounts）

> 数据管理页的历史数据记录不受影响（数据归属 Project，非用户）。

---

## 验收标准

| # | 场景 | 预期结果 |
|---|------|----------|
| AC-1 | 访问 `/end-users` 页面 | 展示用户列表 |
| AC-2 | 创建合法用户 | 201，列表刷新，mc_private.users 新增记录 |
| AC-3 | 创建重名用户 | 409，Modal 提示「用户名已被使用」 |
| AC-4 | 禁用用户 | is_forbidden=1，该用户登录返回 403 |
| AC-5 | 启用用户 | is_forbidden=0，该用户可正常登录 |
| AC-6 | 删除用户 | 记录删除，所有 accounts revoked |
| AC-7 | 非开发者访问此页 | middleware 拦截，redirect 到 /login |
