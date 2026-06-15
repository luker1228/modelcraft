# Model owner 字段生命周期 BDD 测试
# 对应 PRD: ai-metadata/prd/rls/02-model-owner-lifecycle.md
# 验收标准: AC-2, AC-8

Feature: Model owner 字段生命周期
  作为开发者
  我希望新建 Model 自动生成 owner 字段
  以便零配置实现数据隔离

  Background:
    Given 我以管理员身份登录

  @smoke
  Scenario: 新建 Model 自动生成 owner 字段和默认 Policy
    When 我创建名为 "AutoOwnerModel" 的模型
    Then 模型应该创建成功
    And 模型应该包含名为 "owner" 的 EndUserRef 字段

  Scenario: 新建 Model 的 owner 字段可被删除
    Given 已创建名为 "RemovableOwnerModel" 的模型
    When 我删除名为 "owner" 的字段
    Then 字段应该删除成功
    And 该模型应该不存在 RLS 策略

  Scenario: 导入的 Model 不自动生成 owner 字段
    # 导入功能需要单独测试，此处标记为待实现
    Given 从数据库表导入名为 "ImportedModel" 的模型
    Then 该模型不应该包含 "owner" 字段
    And 该模型应该不存在 RLS 策略

  Scenario: 删除 owner 字段后重新添加
    Given 已创建名为 "RecreateOwnerModel" 的模型
    And 我删除名为 "owner" 的字段
    When 我重新为模型添加 EndUserRef 字段
    Then 字段应该添加成功
