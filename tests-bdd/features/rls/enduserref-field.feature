# EndUserRef 字段类型 BDD 测试
# 对应 PRD: ai-metadata/prd/rls/01-enduserref-field.md
# 验收标准: AC-1, AC-7

Feature: EndUserRef 字段类型
  作为开发者
  我希望在 Model 中添加 EndUserRef 类型的字段
  以便实现数据行级隔离

  Background:
    Given 我以管理员身份登录
    And 已创建名为 "RLSTestModel" 的模型

  @smoke
  Scenario: 为模型添加 EndUserRef 字段
    When 我为该模型添加名为 "owner" 格式为 "END_USER_REF" 的字段
    Then 字段应该添加成功
    And 模型应该包含名为 "owner" 的字段

  Scenario: 同一 Model 不能添加第二个 EndUserRef 字段
    Given 模型已有名为 "owner" 格式为 "END_USER_REF" 的字段
    When 我为该模型添加名为 "owner2" 格式为 "END_USER_REF" 的字段
    Then addFields 结果中字段 "owner2" 应该失败并返回 "EndUserRefAlreadyExists"

  Scenario: EndUserRef 字段必须使用固定名称 "owner"
    When 我为该模型添加名为 "customOwner" 格式为 "END_USER_REF" 的字段
    Then addFields 结果中字段 "customOwner" 应该失败并返回 "InvalidInput"

  Scenario: 删除 EndUserRef 字段后策略同步删除
    Given 模型已有名为 "owner" 格式为 "END_USER_REF" 的字段
    When 我删除名为 "owner" 的字段
    Then 字段应该删除成功
    When 我查询该模型的 RLS 策略
    Then 应该返回 null（无策略）
