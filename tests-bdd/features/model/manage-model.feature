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

  Scenario: 通过 model query 请求 jsonSchema 字段时返回合法 JSON Schema
    Given 已创建名为 "JsonSchemaModel" 的模型
    And 模型已有名为 "email" 格式为 "STRING" 的字段
    And 模型已有名为 "age" 格式为 "NUMBER" 的字段
    When 我通过 model query 请求该模型的 jsonSchema 字段
    Then 返回的 jsonSchema 应该是合法的 JSON Schema 字符串
    And jsonSchema 中应该包含字段名 "email"
    And jsonSchema 中应该包含字段名 "age"

  Scenario: 不请求 jsonSchema 字段时 model query 正常返回（向后兼容）
    Given 已创建名为 "BackCompatModel" 的模型
    When 我通过 model query 请求该模型的基础字段（不含 jsonSchema）
    Then 操作应该成功
    And 返回的模型应该包含 id 和 name 字段
