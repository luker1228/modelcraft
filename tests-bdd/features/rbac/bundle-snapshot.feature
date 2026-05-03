# tests-bdd/features/rbac/bundle-snapshot.feature
Feature: 权限包版本快照

  Background:
    Given 我以管理员身份登录

  @smoke
  Scenario: 添加权限点后自动生成快照
    Given 已创建名为 "SnapModel" 的模型
    And 已为该模型创建权限点，名称="权限A"，动作=SELECT，行策略=ALL
    And 已为该模型创建权限点，名称="权限B"，动作=SELECT，行策略=ALL
    And 已创建名为 "快照测试包" 的权限包
    And 已将名为"权限A"的权限点添加到该权限包
    When 我将名为"权限B"的权限点添加到该权限包
    Then 添加操作成功
    And 权限包的 currentVersion 为 2
    And 权限包快照列表长度为 2

  Scenario: 移除权限点后自动生成快照
    Given 已创建名为 "RemoveSnapModel" 的模型
    And 已为该模型创建权限点，名称="权限X"，动作=SELECT，行策略=ALL
    And 已创建名为 "移除快照包" 的权限包
    And 已将名为"权限X"的权限点添加到该权限包
    When 我从该权限包移除名为"权限X"的权限点
    Then 移除操作成功
    And 权限包的 currentVersion 为 2

  Scenario: 超出 5 个版本时滚动删除
    Given 已创建名为 "RollModel" 的模型
    And 已为该模型创建 6 个权限点
    And 已创建名为 "滚动删除包" 的权限包
    And 已依次将 6 个权限点添加到该权限包
    Then 权限包的 currentVersion 为 6
    And 权限包快照列表长度为 5

  Scenario: 回滚到历史版本
    Given 已创建名为 "RestoreModel" 的模型
    And 已为该模型创建权限点，名称="权限P"，动作=SELECT，行策略=ALL
    And 已为该模型创建权限点，名称="权限Q"，动作=SELECT，行策略=ALL
    And 已创建名为 "回滚测试包" 的权限包
    And 已将名为"权限P"的权限点添加到该权限包
    And 已将名为"权限Q"的权限点添加到该权限包
    When 我将权限包回滚到版本 1
    Then 回滚操作成功
    And 回滚后权限包内只有 1 个权限点
    And 回滚后权限包的 currentVersion 大于 2
    And 回滚生成的快照 restoredFrom 字段等于 1

  Scenario: 修改名称不触发快照
    Given 已创建名为 "NoSnapBundle" 的权限包
    When 我将该权限包名称修改为 "NoSnapBundle已改名"
    Then 更新操作成功
    And 权限包的 currentVersion 为 0
    And 权限包快照列表为空

  Scenario: 回滚到不存在的版本返回错误
    Given 已创建名为 "ErrBundle" 的权限包
    When 我将权限包回滚到版本 99
    Then 应该返回错误类型 "ResourceNotFound"
