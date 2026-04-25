## ADDED Requirements

### Requirement: 系统必须支持 EndUser 到 Project 的访问授权关系
系统 SHALL 提供独立的 `EndUserProjectAccess` 关系模型来表达 EndUser 对项目的访问授权，并 MUST 记录授权来源与权限包信息。

#### Scenario: 授予项目访问权限
- **WHEN** 管理员调用 `grantEndUserAccess` 为 EndUser 指定 `projectSlug` 与 `permissionBundleId`
- **THEN** 系统 MUST 创建访问关系并记录 `grantedBy` 与 `grantedAt`

### Requirement: 同一 EndUser 在同一项目不得重复授权
系统 MUST 保证 `(end_user_id, org_name, project_slug)` 的唯一性，重复授予时不得产生重复关系记录。

#### Scenario: 重复授予同一项目
- **WHEN** 对同一 EndUser 和同一项目再次执行授予
- **THEN** 系统 MUST 返回冲突错误或幂等成功，且最终仅保留一条关系记录

### Requirement: Project GraphQL 必须提供授权关系管理接口
系统 SHALL 在 Project GraphQL 提供授权关系的授予、撤销、列表、更新权限包能力。

#### Scenario: 按项目查询访问关系
- **WHEN** 调用 `listProjectEndUserAccess` 查询某 `projectSlug`
- **THEN** 系统 MUST 返回该项目的 EndUser 授权关系并支持分页与筛选

#### Scenario: 更新授权权限包
- **WHEN** 调用 `updateEndUserProjectAccess` 变更 `permissionBundleId`
- **THEN** 系统 MUST 更新目标关系记录并在后续查询中返回新权限包

### Requirement: 删除 EndUser 时必须清理授权关系
系统 MUST 在删除 EndUser 后级联删除其项目访问关系，避免孤儿授权数据。

#### Scenario: 删除账号后查询授权
- **WHEN** EndUser 被删除
- **THEN** 查询该 EndUser 的访问关系 MUST 返回空结果
