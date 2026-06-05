# 功能：mc describe 命令验收测试
# 覆盖：GraphQL introspection 输出结构验证

@cli
Feature: CLI 模型 Introspection

  Background:
    Given 用户已通过 PAT token 登录 CLI
    And 用户已切换到项目 "luke"

  @cli @smoke
  Scenario: describe 返回模型类型结构
    When 用户执行 "describe {PROJECT}.{DATABASE}.{MODEL}"
    Then CLI 命令应该成功
    And CLI 响应应该包含字段 "types"
    And CLI 响应应该包含字段 "meta"

  @cli
  Scenario: describe 使用 2-part path
    When 用户执行 "describe {DATABASE}.{MODEL}"
    Then CLI 命令应该成功
    And CLI 响应应该包含字段 "types"

  @cli
  Scenario: describe meta 字段包含正确的 project/database/model
    When 用户执行 "describe {PROJECT}.{DATABASE}.{MODEL}"
    Then CLI 命令应该成功
    And CLI meta."project" 应该等于 "luke"
    And CLI meta."database" 应该等于 "demo_ecommerce"
    And CLI meta."model" 应该等于 "users"

  @cli
  Scenario: describe 不存在的模型返回错误
    When 用户执行 "describe {PROJECT}.{DATABASE}.nonexistent_model_xyz"
    Then CLI 命令应该失败
    And CLI 错误码应该为 "NOT_FOUND"
    And CLI 退出码应该为 5

  @cli
  Scenario: describe 1-part path 报路径格式错误
    When 用户执行 "describe {MODEL}"
    Then CLI 命令应该失败
    And CLI 退出码应该为 6

  @cli
  Scenario: describe 4-part path 报路径格式错误
    When 用户执行 "describe a.b.c.d"
    Then CLI 命令应该失败
    And CLI 退出码应该为 6
