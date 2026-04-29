## ADDED Requirements

### Requirement: 权限包详情页展示真实权限点列表

权限包详情页的已配置权限点区域 SHALL 直接渲染 `bundle.permissions` 数组，不得使用任何硬编码 fallback 数据。当 `bundle.permissions` 为空时，SHALL 显示空状态提示（「暂无权限点，点击添加」）。

#### Scenario: 权限包有权限点时正常展示
- **WHEN** `bundle.permissions` 包含至少一个权限点
- **THEN** 页面按 model 分组展示所有权限点，每条显示操作类型、行范围、列策略摘要

#### Scenario: 权限包无权限点时显示空状态
- **WHEN** `bundle.permissions` 为空数组
- **THEN** 权限点区域显示空状态 UI（图标 + 文案 + 「添加权限点」按钮），不显示任何 mock 数据

---

### Requirement: 添加策略 Dialog 使用真实权限点数据

「添加策略」Dialog 的可选权限点列表 SHALL 使用 `GET_END_USER_PERMISSIONS` query 返回的真实数据（通过 `useBundleManage` 的 `allPermissions` 字段），不得使用硬编码 fallback 权限点。

#### Scenario: API 返回权限点时展示可选列表
- **WHEN** `allPermissions` 包含至少一个权限点，且权限点未被当前权限包包含
- **THEN** Dialog 右侧区域展示可添加的权限点列表

#### Scenario: API 返回空列表时显示空状态
- **WHEN** `allPermissions` 为空，或所有权限点均已在当前权限包中
- **THEN** Dialog 显示「暂无可添加权限点」的空状态，不展示任何 mock 数据

---

### Requirement: 添加策略 Dialog 移除数据库维度筛选

「添加策略」Dialog SHALL 不显示数据库选择下拉控件，因为当前权限点模型不包含 clusterId 字段。Dialog 布局退化为「左侧 model 列表 + 右侧权限选择」两列。

#### Scenario: Dialog 打开时不显示数据库下拉
- **WHEN** 用户点击「添加权限点」打开 Dialog
- **THEN** Dialog 顶部不出现数据库选择控件，直接展示 model 列表和权限选择器
