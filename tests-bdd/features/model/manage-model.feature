# tests-bdd/features/model/manage-model.feature
Feature: 模型管理

  Background:
    Given 我以管理员身份登录

  @smoke
  Scenario: 成功创建模型
    When 我创建名为 "SmokeModel" 的模型
    Then 模型应该创建成功
    And 模型名称应该是 "SmokeModel"

  Scenario: 成功删除模型
    Given 已创建名为 "DeleteMe" 的模型
    When 我删除该模型
    Then 操作应该成功

  Scenario: 创建重名模型时报错
    Given 已创建名为 "DupModel" 的模型
    When 我再次创建名为 "DupModel" 的模型
    Then 应该返回错误类型 "ModelAlreadyExists"

  Scenario Outline: 创建非法名称模型时报错
    When 我创建名为 "<name>" 的模型
    Then 应该返回错误类型 "InvalidInput"

    Examples:
      | name             |
      | 123startsWithNum |
      | has space        |
      | has-hyphen       |
      | _startsUnderscore |
