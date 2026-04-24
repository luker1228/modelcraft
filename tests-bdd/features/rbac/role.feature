# tests-bdd/features/rbac/role.feature
Feature: RBAC 角色管理

  Background:
    Given 我以管理员身份登录

  @smoke
  Scenario: 成功创建普通角色
    When 我创建名为 "销售专员" 的 RBAC 角色
    Then 角色创建成功
    And 响应包含角色 ID
    And 角色不是隐式角色

  Scenario: 成功将权限包关联到角色
    Given 已创建名为 "RoleBundle" 的权限包
    And 已创建名为 "授权角色" 的 RBAC 角色
    When 我将该权限包关联到该角色
    Then 关联操作成功

  Scenario: 成功解除角色的权限包关联
    Given 已创建名为 "RevBundle" 的权限包
    And 已创建名为 "解除角色" 的 RBAC 角色
    And 已将该权限包关联到该角色
    When 我从该角色解除该权限包的关联
    Then 解除操作成功

  Scenario: 成功删除普通角色
    Given 已创建名为 "待删除角色" 的 RBAC 角色
    When 我删除该角色
    Then 删除操作成功

  Scenario: 成功列出项目下所有角色
    Given 已创建名为 "列表角色A" 的 RBAC 角色
    When 我查询项目下所有角色
    Then 角色列表不为空
