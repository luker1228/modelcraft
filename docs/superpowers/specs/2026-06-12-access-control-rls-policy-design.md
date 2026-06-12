# Access Control 页面简化 — 仅保留 RLS 策略

**日期:** 2026-06-12
**状态:** 已确认

## 目标

将访问控制页面 (`/org/:orgName/project/:projectSlug/access-control`) 从 3 个 tab（角色/权限包/权限点）简化为仅保留 RLS 策略管理。

## 核心设计

- **Role** = 自由文本字段，用户在 Policy 编辑弹窗中直接输入，不是独立实体
- **Policy** = 按 `(modelId, action, role)` 定位，同一 model 下多条 policy 通过 role 隔开
- **页面**: 模型选择器 + 策略表格 + 添加/编辑/删除

## 实现阶段

### 阶段 1: 后端 — 实现 V2 RLS Policy Resolver

- `queryResolver.RlsPolicies(modelId)` — 查询模型全部策略
- `mutationResolver.UpsertRlsPolicy(modelId, input)` — 创建/更新
- `mutationResolver.DeleteRlsPolicy(id)` — 删除单条
- `mutationResolver.DeleteRlsPoliciesByModel(modelId)` — 删除模型全部策略

### 阶段 2: 前端 — Contract 同步

运行 `front-contract-pull` 拉取 `rls_policy_v2.graphql`。

### 阶段 3: 前端 — UI 构建 + 旧代码清理

**新建:**
- `_components/rls-policy/RlsPolicyContent.tsx` — 模型选择器 + 策略表格
- `_components/rls-policy/PolicyEditorDialog.tsx` — 编辑弹窗（Role 自由输入）
- `_hooks/rls-policy/useRlsPolicyList.ts`
- `_hooks/rls-policy/useRlsPolicyManage.ts`

**修改:**
- `page.tsx` — 移除 tab 逻辑
- `_components/index.ts` — 更新导出
- `AppLayout.tsx` — 简化侧边栏
- `route-catalog.ts` — 合并路由

**删除:**
- `_components/roles/`, `_components/bundles/`, `_components/permissions/`
- `_hooks/roles/`, `_hooks/bundles/`, `_hooks/permissions/`
- `[roleId]/`, `bundles/` 子页面
