# tests-bdd/features/field/manage-field.feature
Feature: 字段管理

  Background:
    Given 我以管理员身份登录
    And 已创建名为 "FieldTestModel" 的模型

  Scenario: 成功添加字段
    When 我为该模型添加名为 "email" 格式为 "STRING" 的字段
    Then 字段应该添加成功
    And 模型应该包含名为 "email" 的字段

  Scenario: 成功删除字段
    Given 模型已有名为 "age" 格式为 "NUMBER" 的字段
    When 我删除名为 "age" 的字段
    Then 字段应该删除成功

  Scenario Outline: 添加非法名称字段时报错
    When 我为该模型添加名为 "<name>" 格式为 "STRING" 的字段
    Then 应该返回错误类型 "InvalidModelInput"

    Examples:
      | name         |
      | 123invalid   |
      | has space    |
      | has-hyphen   |
