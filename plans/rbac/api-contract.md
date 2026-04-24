# RBAC 行列级权限管理 — GraphQL API 协议

> **文件路径**：`plans/rbac/api-contract.md`
> **状态**：草稿
> **日期**：2026-04-24
> **作用域**：Project GraphQL 域（`/graphql/org/{orgName}/project/{projectSlug}/`）

---

## 一、概述

### 1.1 新增内容范围

本协议定义 ModelCraft **数据表行列级权限（Data-Level RBAC）** 管理的 GraphQL API，涵盖：

| 实体 | 说明 |
|------|------|
| `EndUserPermission`（权限点） | 描述对某张 Model 的 动作 × 列策略 × 行策略 的最小授权单元 |
| `EndUserPermissionBundle`（权限包） | 多个权限点的有序组合，系统唯一正式授权单位 |
| `EndUserRole`（RBAC 角色） | Project 维度角色；支持 `isImplicit` 内置隐式标记 |
| `EndUserBundleAssignment` | 终端用户直接绑定权限包（通道 1） |
| `EndUserRoleAssignment` | 终端用户绑定显式 RBAC 角色（通道 2；隐式角色不落库） |
| `RolePermissionBundle` | 角色绑定权限包（通道 2 / 3 共用） |
| 有效权限查询 | 三通道并集合并后的结果，供 Runtime 鉴权引擎和审计使用 |

**新增 Schema 文件**：`api/graph/project/schema/rbac.graphql`

### 1.2 与现有 Casbin 权限的边界说明

| 维度 | 现有 Casbin 功能权限 | 新增 Data-Level RBAC |
|------|---------------------|---------------------|
| **定位** | 系统功能权限（开发者侧） | 数据表行列级权限（End User 侧） |
| **模型** | `obj + act`（如 `model:create`） | EndUserPermission × EndUserPermissionBundle × EndUserRole |
| **作用阶段** | Design 阶段 API 鉴权 | Runtime 阶段数据查询过滤 |
| **授权对象** | Org User（开发者 / 管理员） | EndUser（终端用户） |
| **GraphQL 域** | Org 域（`permission.graphql`） | Project 域（`rbac.graphql`，本协议） |
| **执行机制** | `@hasPermission` 指令拦截 | `effectivePermissions` 查询 + RLS 引擎注入 WHERE |
| **类型命名** | `PermissionRole`、`PermissionDef` | `EndUserRole`、`EndUserPermission`、`EndUserPermissionBundle` |

> ⚠️ **命名隔离**：RBAC 角色统一命名为 `EndUserRole`（而非 `PermissionRole`），避免与 Casbin 类型冲突。
> 后端 Go struct 同步使用 `EndUserRole`、`EndUserPermission`、`EndUserPermissionBundle` 命名。

---

## 二、GraphQL Schema

```graphql
# ============================================================
# RBAC — Data-Level Row & Column Permission Management
# File: api/graph/project/schema/rbac.graphql
#
# 依赖（定义于其他文件，此处直接引用）：
#   - interface Error           → cluster.graphql
#   - type InvalidInput         → field.graphql
#   - type ProjectNotFound      → cluster.graphql
#   - type ModelNotFound        → model.graphql
#   - type Model                → model.graphql
#   - type EndUser              → end_user.graphql
#   - scalar Time               → base.graphql
#   - interface Node            → base.graphql
#   - type PageInfo             → base.graphql
# ============================================================


# ============================================================
# Enums
# ============================================================

"""
数据操作动作：终端用户对数据表可执行的操作类型
"""
enum RbacAction {
  SELECT
  INSERT
  UPDATE
  DELETE
  EXPORT
}

"""
行策略范围：控制终端用户可见哪些数据行

ALL                — 全部行，不过滤
SELF               — 仅当前用户自己的行
                     （要求 Model 存在 owner 字段，类型为 END_USER_REF）
DEPT               — 仅当前用户所在部门的行
                     （要求 Model 存在 dept_id 字段）
DEPT_AND_CHILDREN  — 当前部门及所有下级部门的行
                     （要求 Model 存在 dept_id 字段）
"""
enum RowScopeType {
  ALL
  SELF
  DEPT
  DEPT_AND_CHILDREN
}

"""
列访问模式：描述某列在特定权限点下的可访问程度
"""
enum ColumnAccessMode {
  """可见且完整展示（SELECT 可见，INSERT/UPDATE 可写）"""
  VISIBLE

  """可见但只读（SELECT 可见，INSERT/UPDATE 不可写入该字段）"""
  READONLY

  """脱敏：返回遮挡后的值（如 138****8888）"""
  MASKED

  """隐藏：查询结果中不返回该列，INSERT/UPDATE 也不可写入"""
  HIDDEN
}


# ============================================================
# Column Policy — 列策略建模（结构化，非 JSON Scalar）
# ============================================================

"""
单列访问规则
"""
type ColumnRule {
  """字段名（对应 Field.name）"""
  fieldName: String!

  """该字段的访问模式"""
  mode: ColumnAccessMode!

  """脱敏掩码规则（仅 mode=MASKED 时有效，如 "138****{last4}"）"""
  maskPattern: String
}

"""
权限点的列访问策略

defaultMode 作用于所有未在 rules 中显式列出的字段；
rules 中的字段覆盖默认模式，只需列出与默认不同的字段。

示例（全部可见，phone 脱敏，salary 隐藏）：
  defaultMode: VISIBLE
  rules: [
    { fieldName: "phone",  mode: MASKED, maskPattern: "138****{last4}" },
    { fieldName: "salary", mode: HIDDEN }
  ]
"""
type ColumnPolicy {
  """未显式配置字段的默认访问模式"""
  defaultMode: ColumnAccessMode!

  """逐字段覆盖规则（空列表表示全部字段使用 defaultMode）"""
  rules: [ColumnRule!]!
}

"""
创建 / 更新权限点时的列策略输入
"""
input ColumnPolicyInput {
  defaultMode: ColumnAccessMode!
  rules: [ColumnRuleInput!]!
}

input ColumnRuleInput {
  fieldName: String!
  mode: ColumnAccessMode!
  """仅 mode=MASKED 时提供"""
  maskPattern: String
}


# ============================================================
# EndUserPermission（权限点）
# 底层表：end_user_permissions
# ============================================================

"""
权限点：系统中最小权限定义单元。

描述"对 某张 Model 执行 某个动作 时，能看哪些列（columnPolicy）、
能看哪些行（rowScope）"。

- 权限点只承担"能力定义"职责，不直接授权
- 必须通过 EndUserPermissionBundle（权限包）才能授权给用户或角色
- modelId + action 创建后不可修改（需删除重建）
"""
type EndUserPermission implements Node {
  id: ID!

  """所属 Model ID（创建后不可修改）"""
  modelId: ID!

  """所属 Model（关联字段）"""
  model: Model!

  """操作动作（创建后不可修改）"""
  action: RbacAction!

  """列访问策略"""
  columnPolicy: ColumnPolicy!

  """行策略范围"""
  rowScope: RowScopeType!

  """显示名称（如"订单查看-本人"），用于管理界面标识"""
  displayName: String

  """描述"""
  description: String

  createdAt: Time!
  updatedAt: Time!
}

type EndUserPermissionEdge {
  node: EndUserPermission!
  cursor: String!
}

type EndUserPermissionConnection {
  edges: [EndUserPermissionEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}


# ============================================================
# EndUserPermissionBundle（权限包）
# 底层表：end_user_permission_bundles
# ============================================================

"""
权限包：系统中唯一的正式授权单位。

由多个权限点有序组合而成；用户和角色只能关联权限包，
不能直接关联权限点。权限包属于 Project 维度。
"""
type EndUserPermissionBundle implements Node {
  id: ID!

  """权限包名称（Project 内唯一）"""
  name: String!

  """描述"""
  description: String

  """包含的权限点（按 sortOrder 升序排列）"""
  permissions: [EndUserBundlePermissionEntry!]!

  createdAt: Time!
  updatedAt: Time!
}

"""
权限包内的权限点条目（携带排序信息）
"""
type EndUserBundlePermissionEntry {
  """在本权限包中的显示排序（升序）"""
  sortOrder: Int!
  permission: EndUserPermission!
}

type EndUserPermissionBundleEdge {
  node: EndUserPermissionBundle!
  cursor: String!
}

type EndUserPermissionBundleConnection {
  edges: [EndUserPermissionBundleEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}


# ============================================================
# EndUserRole（RBAC 角色）
# 底层表：end_user_roles
# ============================================================

"""
RBAC 角色（Project 维度）。

与 Org 域的 Casbin PermissionRole 完全独立：
- Casbin PermissionRole → 控制开发者对系统资源的 API 访问
- EndUserRole           → 控制终端用户对数据行列的运行时访问

isImplicit = true 的角色为内置隐式角色：
  - 角色定义落库（可查看、可配置权限包），但用户关联不落库
  - 所有已认证终端用户在鉴权时由系统自动注入（无需手工分配）
  - 不可通过 assignEndUserRole 手工分配
  - 不可通过 updateEndUserRole / deleteEndUserRole 修改或删除
  - 可以通过 assignBundleToEndUserRole / revokeBundleFromEndUserRole 配置其权限包
"""
type EndUserRole implements Node {
  id: ID!

  """角色名称（Project 内唯一）"""
  name: String!

  """描述"""
  description: String

  """是否为内置隐式角色"""
  isImplicit: Boolean!

  """关联的权限包列表"""
  permissionBundles: [EndUserRoleBundleEntry!]!

  createdAt: Time!
  updatedAt: Time!
}

"""
角色关联的权限包条目
"""
type EndUserRoleBundleEntry {
  bundle: EndUserPermissionBundle!
  assignedAt: Time!
}

type EndUserRoleEdge {
  node: EndUserRole!
  cursor: String!
}

type EndUserRoleConnection {
  edges: [EndUserRoleEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}


# ============================================================
# 授权关联类型
# ============================================================

"""
终端用户直接绑定的权限包关联记录（通道 1）
底层表：end_user_bundle_assignments
"""
type EndUserBundleAssignment {
  endUserId: ID!
  endUser: EndUser!
  bundle: EndUserPermissionBundle!
  assignedAt: Time!
}

"""
终端用户绑定的显式 RBAC 角色关联记录（通道 2；隐式角色不落库）
底层表：end_user_role_assignments
"""
type EndUserRoleAssignment {
  endUserId: ID!
  endUser: EndUser!
  role: EndUserRole!
  assignedAt: Time!
}


# ============================================================
# 有效权限查询结果（三通道合并）
# ============================================================

"""
某终端用户对指定 Model 的有效权限计算结果。

计算规则（三通道并集）：
  有效权限集 =
      通道 1（EndUserBundleAssignment）展开的权限点
    ∪ 通道 2（EndUserRoleAssignment → RolePermissionBundle）展开的权限点
    ∪ 通道 3（Role.isImplicit=true → RolePermissionBundle）展开的权限点

行策略合并：同一 modelId × action 出现多个 rowScope 时，取最宽范围
  （ALL > DEPT_AND_CHILDREN > DEPT > SELF）

列策略合并：同一 modelId × action 出现多个 columnPolicy 时，
  逐字段取可见度最高的 mode（VISIBLE > READONLY > MASKED > HIDDEN）
"""
type EffectivePermissions {
  """目标终端用户 ID"""
  endUserId: ID!

  """目标 Model ID"""
  modelId: ID!

  """
  合并后的有效权限条目（按 action 分组，每个 action 最多一条合并结果）
  action 未出现时表示无权限（默认拒绝）
  """
  grants: [EffectiveGrant!]!

  """
  来源明细（供审计界面展示"此权限来自哪个角色/权限包"）
  """
  sources: EffectivePermissionSources!
}

"""
单个 action 的有效权限合并结果
"""
type EffectiveGrant {
  action: RbacAction!

  """合并后的列访问策略"""
  columnPolicy: ColumnPolicy!

  """合并后的行策略（最宽范围）"""
  rowScope: RowScopeType!
}

"""
有效权限的三通道来源明细（审计用）
"""
type EffectivePermissionSources {
  """通道 1：用户直接绑定的权限包"""
  directBundles: [EndUserPermissionBundle!]!

  """通道 2：用户显式角色绑定的权限包（含来源角色信息）"""
  explicitRoleBundles: [EndUserRoleBundleSource!]!

  """通道 3：隐式角色自动注入的权限包（含来源角色信息）"""
  implicitRoleBundles: [EndUserRoleBundleSource!]!
}

type EndUserRoleBundleSource {
  role: EndUserRole!
  bundles: [EndUserPermissionBundle!]!
}


# ============================================================
# Error Types
# ============================================================

type EndUserPermissionNotFound implements Error {
  message: String!
}

type EndUserPermissionBundleNotFound implements Error {
  message: String!
}

type EndUserPermissionBundleAlreadyExists implements Error {
  message: String!
  suggestion: String
}

type EndUserRoleNotFound implements Error {
  message: String!
}

type EndUserRoleAlreadyExists implements Error {
  message: String!
  suggestion: String
}

"""
rowScope 要求 Model 上存在特定字段，但该字段不存在时返回此错误。
SELF         → 要求 owner 字段（END_USER_REF 类型）
DEPT / DEPT_AND_CHILDREN → 要求 dept_id 字段
"""
type RowScopeFieldMissing implements Error {
  message: String!
  """缺失的字段名（"owner" 或 "dept_id"）"""
  missingField: String!
  """触发校验的 rowScope 值"""
  requiredByRowScope: RowScopeType!
  suggestion: String
}

"""
对内置隐式角色执行了不允许的操作（修改名称 / 删除）
"""
type EndUserImplicitRoleCannotBeModified implements Error {
  message: String!
  suggestion: String
}

"""
尝试手工将隐式角色分配给终端用户（隐式角色由系统自动注入）
"""
type EndUserCannotAssignImplicitRole implements Error {
  message: String!
  suggestion: String
}

"""
终端用户已绑定该权限包（幂等）
"""
type UserBundleAlreadyAssigned implements Error {
  message: String!
}

"""
终端用户已绑定该角色
"""
type UserRoleAlreadyAssigned implements Error {
  message: String!
}

"""
未找到对应的终端用户（RBAC 上下文）
"""
type EndUserNotFoundInProject implements Error {
  message: String!
}

"""
权限包已被角色或用户绑定，无法删除（需先解除所有绑定）
"""
type EndUserPermissionBundleInUse implements Error {
  message: String!
  suggestion: String
}

"""
权限点已被某权限包引用，无法直接删除（需先从权限包移除）
"""
type EndUserPermissionInUse implements Error {
  message: String!
  suggestion: String
}


# ============================================================
# Error Unions
# ============================================================

# EndUserPermission
union CreateEndUserPermissionError = ModelNotFound | InvalidInput | RowScopeFieldMissing | ProjectNotFound
union UpdateEndUserPermissionError = EndUserPermissionNotFound | InvalidInput | RowScopeFieldMissing | ProjectNotFound
union DeleteEndUserPermissionError = EndUserPermissionNotFound | EndUserPermissionInUse | ProjectNotFound

# EndUserPermissionBundle
union CreateEndUserPermissionBundleError = EndUserPermissionBundleAlreadyExists | InvalidInput | ProjectNotFound
union UpdateEndUserPermissionBundleError = EndUserPermissionBundleNotFound | EndUserPermissionBundleAlreadyExists | InvalidInput | ProjectNotFound
union DeleteEndUserPermissionBundleError = EndUserPermissionBundleNotFound | EndUserPermissionBundleInUse | ProjectNotFound

# Bundle ↔ Permission 关联
union AddEndUserPermissionToBundleError      = EndUserPermissionBundleNotFound | EndUserPermissionNotFound | InvalidInput | ProjectNotFound
union RemoveEndUserPermissionFromBundleError = EndUserPermissionBundleNotFound | EndUserPermissionNotFound | ProjectNotFound
union ReorderEndUserBundlePermissionsError   = EndUserPermissionBundleNotFound | InvalidInput | ProjectNotFound

# EndUserRole
union CreateEndUserRoleError = EndUserRoleAlreadyExists | InvalidInput | ProjectNotFound
union UpdateEndUserRoleError = EndUserRoleNotFound | EndUserImplicitRoleCannotBeModified | EndUserRoleAlreadyExists | InvalidInput | ProjectNotFound
union DeleteEndUserRoleError = EndUserRoleNotFound | EndUserImplicitRoleCannotBeModified | ProjectNotFound

# Role ↔ Bundle 关联
union AssignBundleToEndUserRoleError   = EndUserRoleNotFound | EndUserPermissionBundleNotFound | ProjectNotFound
union RevokeBundleFromEndUserRoleError = EndUserRoleNotFound | EndUserPermissionBundleNotFound | ProjectNotFound

# User ↔ Bundle 直接绑定
union AssignBundleToEndUserError   = EndUserNotFoundInProject | EndUserPermissionBundleNotFound | UserBundleAlreadyAssigned | ProjectNotFound
union RevokeBundleFromEndUserError = EndUserNotFoundInProject | EndUserPermissionBundleNotFound | ProjectNotFound

# User ↔ Role 绑定
union AssignEndUserRoleError = EndUserNotFoundInProject | EndUserRoleNotFound | EndUserCannotAssignImplicitRole | UserRoleAlreadyAssigned | ProjectNotFound
union RevokeEndUserRoleError = EndUserNotFoundInProject | EndUserRoleNotFound | ProjectNotFound

# 有效权限查询
union GetEffectivePermissionsError = EndUserNotFoundInProject | ModelNotFound | ProjectNotFound


# ============================================================
# Payload Types
# ============================================================

# --- EndUserPermission ---
type CreateEndUserPermissionPayload {
  permission: EndUserPermission
  error: CreateEndUserPermissionError
}

type UpdateEndUserPermissionPayload {
  permission: EndUserPermission
  error: UpdateEndUserPermissionError
}

type DeleteEndUserPermissionPayload {
  success: Boolean!
  error: DeleteEndUserPermissionError
}

# --- EndUserPermissionBundle ---
type CreateEndUserPermissionBundlePayload {
  bundle: EndUserPermissionBundle
  error: CreateEndUserPermissionBundleError
}

type UpdateEndUserPermissionBundlePayload {
  bundle: EndUserPermissionBundle
  error: UpdateEndUserPermissionBundleError
}

type DeleteEndUserPermissionBundlePayload {
  success: Boolean!
  error: DeleteEndUserPermissionBundleError
}

type AddEndUserPermissionToBundlePayload {
  bundle: EndUserPermissionBundle
  error: AddEndUserPermissionToBundleError
}

type RemoveEndUserPermissionFromBundlePayload {
  bundle: EndUserPermissionBundle
  error: RemoveEndUserPermissionFromBundleError
}

type ReorderEndUserBundlePermissionsPayload {
  bundle: EndUserPermissionBundle
  error: ReorderEndUserBundlePermissionsError
}

# --- EndUserRole ---
type CreateEndUserRolePayload {
  role: EndUserRole
  error: CreateEndUserRoleError
}

type UpdateEndUserRolePayload {
  role: EndUserRole
  error: UpdateEndUserRoleError
}

type DeleteEndUserRolePayload {
  success: Boolean!
  error: DeleteEndUserRoleError
}

type AssignBundleToEndUserRolePayload {
  role: EndUserRole
  error: AssignBundleToEndUserRoleError
}

type RevokeBundleFromEndUserRolePayload {
  role: EndUserRole
  error: RevokeBundleFromEndUserRoleError
}

# --- User ↔ Bundle ---
type AssignBundleToEndUserPayload {
  endUserId: ID!
  bundle: EndUserPermissionBundle
  error: AssignBundleToEndUserError
}

type RevokeBundleFromEndUserPayload {
  success: Boolean!
  error: RevokeBundleFromEndUserError
}

# --- User ↔ Role ---
type AssignEndUserRolePayload {
  endUserId: ID!
  role: EndUserRole
  error: AssignEndUserRoleError
}

type RevokeEndUserRolePayload {
  success: Boolean!
  error: RevokeEndUserRoleError
}

# --- 有效权限查询 ---
type GetEffectivePermissionsPayload {
  effectivePermissions: EffectivePermissions
  error: GetEffectivePermissionsError
}


# ============================================================
# Input Types
# ============================================================

# --- EndUserPermission ---
input CreateEndUserPermissionInput {
  """所属 Model ID"""
  modelId: ID!
  """操作动作（创建后不可变）"""
  action: RbacAction!
  """列访问策略"""
  columnPolicy: ColumnPolicyInput!
  """行策略范围"""
  rowScope: RowScopeType!
  """显示名称（可选）"""
  displayName: String
  """描述（可选）"""
  description: String
}

input UpdateEndUserPermissionInput {
  """列访问策略（不传则不修改）"""
  columnPolicy: ColumnPolicyInput
  """行策略范围（不传则不修改；修改时重新校验 Model 字段前提）"""
  rowScope: RowScopeType
  displayName: String
  description: String
  # 注意：modelId 和 action 不可修改，不在此 input 中
}

input ListEndUserPermissionsInput {
  """按 Model ID 过滤（不传则返回 Project 下所有权限点）"""
  modelId: ID
  """按动作过滤"""
  action: RbacAction
  first: Int
  after: String
}

# --- EndUserPermissionBundle ---
input CreateEndUserPermissionBundleInput {
  name: String!
  description: String
}

input UpdateEndUserPermissionBundleInput {
  name: String
  description: String
}

input ListEndUserPermissionBundlesInput {
  """名称模糊搜索"""
  search: String
  first: Int
  after: String
}

input AddEndUserPermissionToBundleInput {
  bundleId: ID!
  permissionId: ID!
  """排序值（不传时追加到末尾）"""
  sortOrder: Int
}

input RemoveEndUserPermissionFromBundleInput {
  bundleId: ID!
  permissionId: ID!
}

input ReorderEndUserBundlePermissionsInput {
  bundleId: ID!
  """权限点 ID 的完整有序列表；列表顺序即为新的 sortOrder（从 0 开始）"""
  permissionIds: [ID!]!
}

# --- EndUserRole ---
input CreateEndUserRoleInput {
  name: String!
  description: String
}

input UpdateEndUserRoleInput {
  name: String
  description: String
}

input ListEndUserRolesInput {
  """
  是否包含隐式角色（默认 true）。
  设为 false 时只返回可手工分配的普通角色（适用于"分配角色"选择器场景）。
  """
  includeImplicit: Boolean = true
  """名称模糊搜索"""
  search: String
  first: Int
  after: String
}

# --- Role ↔ Bundle ---
input AssignBundleToEndUserRoleInput {
  roleId: ID!
  bundleId: ID!
}

input RevokeBundleFromEndUserRoleInput {
  roleId: ID!
  bundleId: ID!
}

# --- User ↔ Bundle ---
input AssignBundleToEndUserInput {
  endUserId: ID!
  bundleId: ID!
}

input RevokeBundleFromEndUserInput {
  endUserId: ID!
  bundleId: ID!
}

# --- User ↔ Role ---
input AssignEndUserRoleInput {
  endUserId: ID!
  roleId: ID!
}

input RevokeEndUserRoleInput {
  endUserId: ID!
  roleId: ID!
}

# --- 有效权限查询 ---
input GetEffectivePermissionsInput {
  endUserId: ID!
  modelId: ID!
}


# ============================================================
# Query Extensions
# ============================================================

extend type Query {

  # ── EndUserPermission ────────────────────────────────────

  """
  查询单个权限点（含列策略、行策略、所属 Model）
  """
  endUserPermission(id: ID!): EndUserPermission @hasPermission(action: "rbac:read")

  """
  列出 Project 下的权限点，支持按 modelId / action 过滤，游标分页
  """
  endUserPermissions(input: ListEndUserPermissionsInput): EndUserPermissionConnection! @hasPermission(action: "rbac:read")

  # ── EndUserPermissionBundle ──────────────────────────────

  """
  查询单个权限包（含有序权限点列表）
  """
  endUserPermissionBundle(id: ID!): EndUserPermissionBundle @hasPermission(action: "rbac:read")

  """
  列出 Project 下所有权限包，支持名称模糊搜索，游标分页
  """
  endUserPermissionBundles(input: ListEndUserPermissionBundlesInput): EndUserPermissionBundleConnection! @hasPermission(action: "rbac:read")

  # ── EndUserRole ──────────────────────────────────────────

  """
  查询单个 RBAC 角色（含关联权限包）
  """
  endUserRole(id: ID!): EndUserRole @hasPermission(action: "rbac:read")

  """
  列出 Project 下所有 RBAC 角色；includeImplicit=false 只返回可手工分配的普通角色
  """
  endUserRoles(input: ListEndUserRolesInput): EndUserRoleConnection! @hasPermission(action: "rbac:read")

  # ── 用户授权绑定查询 ──────────────────────────────────────

  """
  查询指定终端用户直接绑定的权限包列表（通道 1 数据）
  """
  endUserBundleAssignments(endUserId: ID!): [EndUserBundleAssignment!]! @hasPermission(action: "rbac:read")

  """
  查询指定终端用户绑定的显式 RBAC 角色列表（通道 2 数据；不含隐式角色）
  """
  endUserRoleAssignments(endUserId: ID!): [EndUserRoleAssignment!]! @hasPermission(action: "rbac:read")

  # ── 有效权限（三通道合并） ────────────────────────────────

  """
  查询某终端用户对指定 Model 的有效权限（三通道并集合并结果）。

  供 Runtime 鉴权引擎调用；也可用于管理界面的权限调试和审计。
  返回 grants（合并后每个 action 的最终策略）和 sources（来源明细）。
  """
  effectivePermissions(input: GetEffectivePermissionsInput!): GetEffectivePermissionsPayload! @hasPermission(action: "rbac:read")
}


# ============================================================
# Mutation Extensions
# ============================================================

extend type Mutation {

  # ── EndUserPermission CRUD ───────────────────────────────

  """
  创建权限点。
  - rowScope = SELF              → 校验 Model 存在 owner 字段（END_USER_REF 类型）
  - rowScope = DEPT              → 校验 Model 存在 dept_id 字段
  - rowScope = DEPT_AND_CHILDREN → 校验 Model 存在 dept_id 字段
  字段不存在时返回 RowScopeFieldMissing 错误（前端应提前禁用对应 rowScope 选项）。
  """
  createEndUserPermission(input: CreateEndUserPermissionInput!): CreateEndUserPermissionPayload! @hasPermission(action: "rbac:manage")

  """
  更新权限点的列策略、行策略、显示名称或描述。
  modelId 和 action 不可修改（不在 input 中）。
  修改 rowScope 时重新执行 Model 字段前提校验。
  """
  updateEndUserPermission(id: ID!, input: UpdateEndUserPermissionInput!): UpdateEndUserPermissionPayload! @hasPermission(action: "rbac:manage")

  """
  删除权限点。
  若权限点已被某权限包引用，返回 EndUserPermissionInUse 错误；
  需先调用 removeEndUserPermissionFromBundle 解除引用后再删除。
  """
  deleteEndUserPermission(id: ID!): DeleteEndUserPermissionPayload! @hasPermission(action: "rbac:manage")

  # ── EndUserPermissionBundle CRUD ─────────────────────────

  """
  创建权限包（名称在 Project 内唯一）
  """
  createEndUserPermissionBundle(input: CreateEndUserPermissionBundleInput!): CreateEndUserPermissionBundlePayload! @hasPermission(action: "rbac:manage")

  """
  更新权限包的 name 或 description
  """
  updateEndUserPermissionBundle(id: ID!, input: UpdateEndUserPermissionBundleInput!): UpdateEndUserPermissionBundlePayload! @hasPermission(action: "rbac:manage")

  """
  删除权限包。
  若权限包已被角色或用户绑定，返回 EndUserPermissionBundleInUse 错误；
  需先解除所有绑定后再删除。
  """
  deleteEndUserPermissionBundle(id: ID!): DeleteEndUserPermissionBundlePayload! @hasPermission(action: "rbac:manage")

  # ── Bundle ↔ Permission 关联 ──────────────────────────────

  """
  向权限包中添加权限点（携带 sortOrder；不传时追加到末尾）
  """
  addEndUserPermissionToBundle(input: AddEndUserPermissionToBundleInput!): AddEndUserPermissionToBundlePayload! @hasPermission(action: "rbac:manage")

  """
  从权限包中移除权限点
  """
  removeEndUserPermissionFromBundle(input: RemoveEndUserPermissionFromBundleInput!): RemoveEndUserPermissionFromBundlePayload! @hasPermission(action: "rbac:manage")

  """
  对权限包内的权限点进行重新排序（传入完整的有序 permissionId 列表，全量替换 sortOrder）
  """
  reorderEndUserBundlePermissions(input: ReorderEndUserBundlePermissionsInput!): ReorderEndUserBundlePermissionsPayload! @hasPermission(action: "rbac:manage")

  # ── EndUserRole CRUD ─────────────────────────────────────

  """
  创建 RBAC 角色（isImplicit 固定为 false；内置隐式角色由系统初始化，不通过此接口）
  """
  createEndUserRole(input: CreateEndUserRoleInput!): CreateEndUserRolePayload! @hasPermission(action: "rbac:manage")

  """
  更新 RBAC 角色的 name / description。
  isImplicit = true 的角色不可修改，返回 EndUserImplicitRoleCannotBeModified。
  """
  updateEndUserRole(id: ID!, input: UpdateEndUserRoleInput!): UpdateEndUserRolePayload! @hasPermission(action: "rbac:manage")

  """
  删除 RBAC 角色。
  isImplicit = true 的角色不可删除，返回 EndUserImplicitRoleCannotBeModified。
  """
  deleteEndUserRole(id: ID!): DeleteEndUserRolePayload! @hasPermission(action: "rbac:manage")

  # ── Role ↔ Bundle 关联 ────────────────────────────────────

  """
  将权限包关联到角色。
  隐式角色也允许配置权限包（符合"角色落库，定义可见"原则）。
  """
  assignBundleToEndUserRole(input: AssignBundleToEndUserRoleInput!): AssignBundleToEndUserRolePayload! @hasPermission(action: "rbac:manage")

  """
  从角色解除权限包关联
  """
  revokeBundleFromEndUserRole(input: RevokeBundleFromEndUserRoleInput!): RevokeBundleFromEndUserRolePayload! @hasPermission(action: "rbac:manage")

  # ── User ↔ Bundle 直接绑定（通道 1）─────────────────────────

  """
  将权限包直接授予终端用户（EndUserBundleAssignment 关联，通道 1）
  """
  assignBundleToEndUser(input: AssignBundleToEndUserInput!): AssignBundleToEndUserPayload! @hasPermission(action: "rbac:manage")

  """
  解除终端用户的权限包直接绑定
  """
  revokeBundleFromEndUser(input: RevokeBundleFromEndUserInput!): RevokeBundleFromEndUserPayload! @hasPermission(action: "rbac:manage")

  # ── User ↔ EndUserRole 绑定（通道 2）─────────────────────────

  """
  将显式 RBAC 角色分配给终端用户（EndUserRoleAssignment 关联，通道 2）。
  isImplicit = true 的角色不可分配，返回 EndUserCannotAssignImplicitRole。
  """
  assignEndUserRole(input: AssignEndUserRoleInput!): AssignEndUserRolePayload! @hasPermission(action: "rbac:manage")

  """
  撤销终端用户的 RBAC 角色分配
  """
  revokeEndUserRole(input: RevokeEndUserRoleInput!): RevokeEndUserRolePayload! @hasPermission(action: "rbac:manage")
}
```

---

## 三、接口清单表

### 3.1 Query 接口

| 接口名 | 参数 | 返回类型 | 功能说明 |
|--------|------|----------|---------|
| `endUserPermission` | `id: ID!` | `EndUserPermission` | 查询单个权限点 |
| `endUserPermissions` | `input: ListEndUserPermissionsInput` | `EndUserPermissionConnection!` | 列出 Project 下权限点（支持按 modelId / action 过滤，游标分页） |
| `endUserPermissionBundle` | `id: ID!` | `EndUserPermissionBundle` | 查询单个权限包（含有序权限点列表） |
| `endUserPermissionBundles` | `input: ListEndUserPermissionBundlesInput` | `EndUserPermissionBundleConnection!` | 列出 Project 下所有权限包（支持名称搜索，游标分页） |
| `endUserRole` | `id: ID!` | `EndUserRole` | 查询单个 RBAC 角色（含关联权限包） |
| `endUserRoles` | `input: ListEndUserRolesInput` | `EndUserRoleConnection!` | 列出 Project 下所有 RBAC 角色；`includeImplicit=false` 过滤隐式角色（适用于角色分配选择器）|
| `endUserBundleAssignments` | `endUserId: ID!` | `[EndUserBundleAssignment!]!` | 查询终端用户直接绑定的权限包（通道 1 数据）|
| `endUserRoleAssignments` | `endUserId: ID!` | `[EndUserRoleAssignment!]!` | 查询终端用户绑定的显式角色（通道 2 数据，不含隐式角色）|
| `effectivePermissions` | `input: GetEffectivePermissionsInput!` | `GetEffectivePermissionsPayload!` | 三通道并集合并指定用户对指定 Model 的有效权限，附带来源明细（审计用）|

### 3.2 Mutation 接口

#### 权限点（EndUserPermission）管理

| 接口名 | 参数 | 返回类型 | 功能说明 |
|--------|------|----------|---------|
| `createEndUserPermission` | `input: CreateEndUserPermissionInput!` | `CreateEndUserPermissionPayload!` | 创建权限点；后端校验 rowScope 对应 Model 字段前提 |
| `updateEndUserPermission` | `id: ID!`，`input: UpdateEndUserPermissionInput!` | `UpdateEndUserPermissionPayload!` | 更新权限点的列策略/行策略/名称；`modelId`+`action` 不可变 |
| `deleteEndUserPermission` | `id: ID!` | `DeleteEndUserPermissionPayload!` | 删除权限点；已被权限包引用时返回 `EndUserPermissionInUse` |

#### 权限包（EndUserPermissionBundle）管理

| 接口名 | 参数 | 返回类型 | 功能说明 |
|--------|------|----------|---------|
| `createEndUserPermissionBundle` | `input: CreateEndUserPermissionBundleInput!` | `CreateEndUserPermissionBundlePayload!` | 创建权限包（名称 Project 内唯一）|
| `updateEndUserPermissionBundle` | `id: ID!`，`input: UpdateEndUserPermissionBundleInput!` | `UpdateEndUserPermissionBundlePayload!` | 更新权限包 name / description |
| `deleteEndUserPermissionBundle` | `id: ID!` | `DeleteEndUserPermissionBundlePayload!` | 删除权限包；已被绑定时返回 `EndUserPermissionBundleInUse` |
| `addEndUserPermissionToBundle` | `input: AddEndUserPermissionToBundleInput!` | `AddEndUserPermissionToBundlePayload!` | 向权限包添加权限点（携带 sortOrder）|
| `removeEndUserPermissionFromBundle` | `input: RemoveEndUserPermissionFromBundleInput!` | `RemoveEndUserPermissionFromBundlePayload!` | 从权限包移除权限点 |
| `reorderEndUserBundlePermissions` | `input: ReorderEndUserBundlePermissionsInput!` | `ReorderEndUserBundlePermissionsPayload!` | 对权限包内权限点重新排序（全量替换式）|

#### RBAC 角色（EndUserRole）管理

| 接口名 | 参数 | 返回类型 | 功能说明 |
|--------|------|----------|---------|
| `createEndUserRole` | `input: CreateEndUserRoleInput!` | `CreateEndUserRolePayload!` | 创建普通 RBAC 角色（isImplicit 固定 false）|
| `updateEndUserRole` | `id: ID!`，`input: UpdateEndUserRoleInput!` | `UpdateEndUserRolePayload!` | 更新角色名称/描述；隐式角色不可操作 |
| `deleteEndUserRole` | `id: ID!` | `DeleteEndUserRolePayload!` | 删除角色；隐式角色不可删除 |
| `assignBundleToEndUserRole` | `input: AssignBundleToEndUserRoleInput!` | `AssignBundleToEndUserRolePayload!` | 将权限包关联到角色（隐式角色也允许配置权限包）|
| `revokeBundleFromEndUserRole` | `input: RevokeBundleFromEndUserRoleInput!` | `RevokeBundleFromEndUserRolePayload!` | 从角色解除权限包关联 |

#### 用户授权

| 接口名 | 参数 | 返回类型 | 功能说明 |
|--------|------|----------|---------|
| `assignBundleToEndUser` | `input: AssignBundleToEndUserInput!` | `AssignBundleToEndUserPayload!` | 直接将权限包授予终端用户（通道 1）|
| `revokeBundleFromEndUser` | `input: RevokeBundleFromEndUserInput!` | `RevokeBundleFromEndUserPayload!` | 解除终端用户的直接权限包绑定 |
| `assignEndUserRole` | `input: AssignEndUserRoleInput!` | `AssignEndUserRolePayload!` | 将显式角色分配给终端用户（通道 2）；隐式角色返回 `EndUserCannotAssignImplicitRole` |
| `revokeEndUserRole` | `input: RevokeEndUserRoleInput!` | `RevokeEndUserRolePayload!` | 撤销终端用户的角色分配 |

---

## 四、设计说明

### 4.1 columnPolicy 建模方式

**决策**：使用结构化 GraphQL 类型（`ColumnPolicy` / `ColumnRule` / `ColumnAccessMode`），而非 JSON Scalar。

**理由**：
- **类型安全**：前端 codegen 可生成强类型，直接绑定枚举值，避免手写 JSON 字符串产生的序列化错误
- **后端校验**：`ColumnRuleInput.fieldName` 可对照 `Model.fields[]` 做字段名白名单校验，阻止引用不存在的字段
- **可扩展性**：`maskPattern` 等脱敏规则有明确位置，后续增加如 `encryptedMode` 等属性不破坏结构

**"默认模式 + 例外覆盖"设计**（`defaultMode` + `rules`）：
- 只需列出与默认不同的字段，减少配置冗余
- `defaultMode: VISIBLE, rules: [{fieldName:"salary", mode:HIDDEN}]` → 全部可见，仅薪资隐藏
- `defaultMode: HIDDEN,  rules: [{fieldName:"id", mode:VISIBLE}, {fieldName:"status", mode:VISIBLE}]` → 仅 id/status 可见

**Go 后端 ColumnPolicy 映射**：

```go
// ColumnAccessMode 对应 GraphQL enum ColumnAccessMode
type ColumnAccessMode string // "VISIBLE" | "READONLY" | "MASKED" | "HIDDEN"

// ColumnRule 对应 GraphQL type ColumnRule
type ColumnRule struct {
    FieldName   string           `json:"field_name"`
    Mode        ColumnAccessMode `json:"mode"`
    MaskPattern string           `json:"mask_pattern,omitempty"`
}

// ColumnPolicy 对应 GraphQL type ColumnPolicy
type ColumnPolicy struct {
    DefaultMode ColumnAccessMode `json:"default_mode"`
    Rules       []ColumnRule     `json:"rules"`
}

// JSON 序列化后存储于 end_user_permissions.column_policy 列（TEXT / JSON 类型）
```

### 4.2 隐式角色的设计与防护

**"角色落库，关系隐式"**（对应 PRD `02-implicit-roles.md`）：

| 操作 | 隐式角色是否允许 | 说明 |
|------|-----------------|------|
| `endUserRoles(includeImplicit: true)` 查询 | ✅ | 管理界面可见，支持审计 |
| `assignBundleToEndUserRole` / `revokeBundleFromEndUserRole` | ✅ | 可修改隐式角色绑定的权限包 |
| `assignEndUserRole` | ❌ → `EndUserCannotAssignImplicitRole` | 隐式角色由系统自动注入，禁止手工分配 |
| `updateEndUserRole` / `deleteEndUserRole` | ❌ → `EndUserImplicitRoleCannotBeModified` | 内置角色定义由系统管理 |

`ListEndUserRolesInput.includeImplicit = false` 专为"角色选择器"场景设计，过滤掉隐式角色，避免误操作。

### 4.3 有效权限三通道合并规则

`effectivePermissions` Query 在服务端执行完整合并：

```
通道 1：EndUserBundleAssignment  → 直接绑定的权限包 → 展开权限点
通道 2：EndUserRoleAssignment    → 显式角色绑定的权限包 → 展开权限点
通道 3：Role.isImplicit=true     → 隐式角色绑定的权限包 → 展开权限点

有效权限集 = 通道 1 ∪ 通道 2 ∪ 通道 3
```

**行策略合并（取最宽范围）**：
```
ALL > DEPT_AND_CHILDREN > DEPT > SELF
```
同一 `modelId × action` 出现多个 `rowScope` 时保留最宽者（`ALL` 覆盖 `SELF`）。

**列策略合并（取最宽松可见度）**：
- 同一字段出现多个 `mode` 时，取可见度最高的：`VISIBLE > READONLY > MASKED > HIDDEN`
- 多个 `defaultMode` 取最宽松值；`rules` 逐字段取最宽松值

`sources` 字段保留三通道原始来源明细（角色名 + 权限包名），供审计界面展示"此权限来自哪里"。

### 4.4 权限点不可变字段（modelId + action）

权限点创建后 `modelId` 和 `action` **不可修改**，理由：
- 修改会使所有引用该权限点的权限包语义发生不可追踪的变化
- 如需变更，应删除旧权限点并创建新权限点（有审计记录）

`UpdateEndUserPermissionInput` 因此不包含 `modelId` / `action` 字段。

### 4.5 权限点删除的引用检查

删除权限点采用**保守策略**（不提供 `force` 参数）：
- 若权限点已被某权限包引用，直接返回 `EndUserPermissionInUse` 错误
- 调用方需先调用 `removeEndUserPermissionFromBundle` 逐一解除引用，再执行删除

**理由**：强制显式解除引用，确保管理员知道哪些权限包会受到影响，避免静默删除引发的权限漏洞。

### 4.6 `@hasPermission` Action 命名规则

新增两个 Casbin action（对应系统功能权限层，控制开发者对 RBAC 配置的访问）：

| Action | 适用接口 | 对应 Casbin obj+act |
|--------|----------|---------------------|
| `rbac:read` | 所有 Query | `{projectSlug}/rbac` + `read` |
| `rbac:manage` | 所有 Mutation | `{projectSlug}/rbac` + `manage` |

> 注意：`@hasPermission` 校验的是**开发者**对 RBAC 配置管理的权限，而非 End User 的数据权限。两层权限不可混淆。

### 4.7 与 RLS 引擎的接口边界

```
RBAC（本协议）                        RLS 引擎（rls.graphql 体系）
    │                                         │
EndUserPermission.rowScope  ──────►  编译为参数化 SQL WHERE 子句
（策略定义层）                               （策略执行层）
```

RBAC 只负责写入策略（`rowScope` 枚举值），RLS 引擎只负责在 Runtime 查询时读取并注入 WHERE。
两者通过 `rowScope` 枚举值解耦，互不侵入实现细节。

### 4.8 EndUser 隔离

所有授权关联对象均为 **EndUser**（终端用户），`endUserId` 对应 `end_user.graphql` 的 `EndUser.id`。  
开发者（Org User）的权限管理继续使用 Org 域 `permission.graphql` + Casbin 体系。  
两套数据**完全隔离**，没有跨表外键。

### 4.9 权限包删除的保护策略

`deleteEndUserPermissionBundle` 采用**拒绝删除**策略（返回 `EndUserPermissionBundleInUse`）而非级联删除：
- 若强制级联解除所有用户/角色绑定，可能在无感知的情况下大规模撤销权限
- 调用方需先通过 `revokeBundleFromEndUser` / `revokeBundleFromEndUserRole` 显式解除所有绑定后再删除
