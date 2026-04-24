# tests-bdd/features/rbac/permission.feature
Feature: RBAC 权限点管理

  Background:
    Given 我以管理员身份登录

  @smoke
  Scenario: 成功创建 SELECT ALL 权限点
    Given 已创建名为 "OrdersModel" 的模型
    When 我为该模型创建权限点，名称="订单查看全部"，动作=SELECT，行策略=ALL
    Then 权限点创建成功
    And 响应包含权限点 ID

  Scenario: 成功列出项目下权限点
    Given 已创建名为 "ListPerm" 的模型
    And 已为该模型创建权限点，名称="查看权限点"，动作=SELECT，行策略=ALL
    When 我查询项目下所有权限点
    Then 权限点列表不为空

  Scenario: 成功删除权限点
    Given 已创建名为 "DelPerm" 的模型
    And 已为该模型创建权限点，名称="待删除权限"，动作=SELECT，行策略=ALL
    When 我删除该权限点
    Then 删除操作成功

  Scenario: 创建重名权限点时报错
    Given 已创建名为 "DupPerm" 的模型
    And 已为该模型创建权限点，名称="重名权限"，动作=SELECT，行策略=ALL
    When 我再次为该模型创建相同动作行策略的权限点，名称="重名权限"，动作=SELECT，行策略=ALL
    Then 应该返回权限点错误
