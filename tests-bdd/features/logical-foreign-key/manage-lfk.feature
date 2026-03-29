# tests-bdd/features/logical-foreign-key/manage-lfk.feature
Feature: 逻辑外键管理

  Background:
    Given 我以管理员身份登录
    And 已创建名为 "OrderModel" 的模型
    And 已创建名为 "UserModel" 的模型
    And "OrderModel" 已有名为 "userId" 格式为 "STRING" 的字段
    And "UserModel" 已有名为 "id" 格式为 "STRING" 的字段

  Scenario: 成功创建逻辑外键
    When 我创建从 "OrderModel.userId" 到 "UserModel.id" 的逻辑外键
    Then 逻辑外键应该创建成功

  Scenario: 字段不存在时报错
    When 我创建从 "OrderModel.nonExistent" 到 "UserModel.id" 的逻辑外键
    Then 应该返回 LFK 错误类型 "FKColumnsNotFoundError"
