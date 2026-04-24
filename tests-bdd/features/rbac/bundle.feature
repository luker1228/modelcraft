# tests-bdd/features/rbac/bundle.feature
Feature: RBAC 权限包管理

  Background:
    Given 我以管理员身份登录

  @smoke
  Scenario: 成功创建权限包
    When 我创建名为 "基础查看包" 的权限包
    Then 权限包创建成功
    And 响应包含权限包 ID

  Scenario: 成功向权限包添加权限点
    Given 已创建名为 "BundlePerm" 的模型
    And 已为该模型创建权限点，名称="打包权限"，动作=SELECT，行策略=ALL
    And 已创建名为 "测试权限包" 的权限包
    When 我将该权限点添加到该权限包，排序=1
    Then 添加操作成功
    And 权限包内权限点列表不为空

  Scenario: 成功从权限包移除权限点
    Given 已创建名为 "RemovePerm" 的模型
    And 已为该模型创建权限点，名称="可移除权限"，动作=SELECT，行策略=ALL
    And 已创建名为 "移除测试包" 的权限包
    And 已将该权限点添加到该权限包
    When 我从该权限包移除该权限点
    Then 移除操作成功

  Scenario: 成功删除权限包
    Given 已创建名为 "待删除包" 的权限包
    When 我删除该权限包
    Then 删除操作成功
