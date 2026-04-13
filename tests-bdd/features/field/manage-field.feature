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
    Then 应该返回错误类型 "InvalidInput"

    Examples:
      | name         |
      | 123invalid   |
      | has space    |
      | has-hyphen   |

  Scenario: AddFields 逐字段返回 results 并允许部分成功
    Given 已创建名为 "CustomerLevel" 的枚举，选项为 "VIP,NORMAL"
    When 我批量添加字段
      | name       | title      | format     | relateEnumName |
      | level      | 客户等级   | ENUM       | @lastEnum      |
      | levelLabel | 客户等级标签 | ENUM_LABEL  |                |
    Then addFields 结果中字段 "level" 应该成功
    And addFields 结果中字段 "levelLabel" 应该失败并返回 "InvalidInput"
    And 模型应该包含名为 "level" 的字段

  Scenario: 同一 source 重复创建关系返回冲突错误
    Given 已创建名为 "CustomerLevel" 的枚举，选项为 "VIP,NORMAL"
    And 模型已有名为 "level" 格式为 "ENUM" 且关联最近创建枚举的字段
    And 已创建字段枚举关联 source "level" label "levelLabelA"
    When 我创建字段枚举关联 source "level" label "levelLabelB"
    Then 应该返回错误类型 "FieldEnumSourceConflict"
    And 错误码应该是 "FIELD_ENUM_SOURCE_CONFLICT"

  Scenario: 删除被关系引用的 ENUM source 字段被阻断
    Given 已创建名为 "CustomerLevel" 的枚举，选项为 "VIP,NORMAL"
    And 模型已有名为 "level" 格式为 "ENUM" 且关联最近创建枚举的字段
    And 已创建字段枚举关联 source "level" label "levelLabelA"
    When 我删除名为 "level" 的字段
    Then 应该返回错误类型 "FieldReferenceInUse"
    And 错误码应该是 "FIELD_REFERENCE_IN_USE"
