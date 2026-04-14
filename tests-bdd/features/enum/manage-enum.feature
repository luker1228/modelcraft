# tests-bdd/features/enum/manage-enum.feature
Feature: 枚举管理

  Background:
    Given 我以管理员身份登录

  @smoke
  Scenario: 成功创建枚举
    When 我创建名为 "Status" 的枚举，选项为 "ACTIVE,INACTIVE"
    Then 枚举应该创建成功
    And 枚举名称应该包含 "Status"

  Scenario: 创建重名枚举时报错
    Given 已创建名为 "Priority" 的枚举，选项为 "HIGH,LOW"
    When 我再次创建名为 "Priority" 的枚举，选项为 "HIGH,LOW"
    Then 应该返回错误类型 "EnumAlreadyExists"

  Scenario Outline: 创建非法名称枚举时报错
    When 我创建名为 "<name>" 的枚举，选项为 "A,B"
    Then 应该返回错误类型 "InvalidInput"

    Examples:
      | name         |
      | 123invalid   |
      | has space    |
      | has-hyphen   |
