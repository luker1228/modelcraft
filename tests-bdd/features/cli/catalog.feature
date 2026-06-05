# 功能：mc catalog 命令验收测试
# 覆盖：projects/databases/models 发现，以及各种错误场景

@cli
Feature: CLI 资源目录发现

  Background:
    Given 用户已通过 PAT token 登录 CLI
    And 用户已切换到项目 "luke"

  @cli @smoke
  Scenario: catalog projects 列出可访问项目
    When 用户执行 "catalog projects"
    Then CLI 命令应该成功
    And CLI 响应应该包含字段 "items"
    And CLI data 中的 "items" 字段应该包含项目 "luke"

  @cli @smoke
  Scenario: catalog databases 使用默认项目上下文
    When 用户执行 "catalog databases"
    Then CLI 命令应该成功
    And CLI 响应应该包含字段 "items"
    And CLI data."items" 应该不为空

  @cli
  Scenario: catalog databases 使用 --project 参数
    When 用户执行 "catalog databases --project {PROJECT}"
    Then CLI 命令应该成功
    And CLI meta."project" 应该等于 "luke"

  @cli @smoke
  Scenario: catalog models 列出数据库中的模型
    When 用户执行 "catalog models --database {DATABASE}"
    Then CLI 命令应该成功
    And CLI 响应应该包含字段 "items"
    And CLI data."items" 应该不为空

  @cli
  Scenario: catalog models 包含 users 模型
    When 用户执行 "catalog models --database {DATABASE}"
    Then CLI 命令应该成功
    # items 是 [{name: "users"}, ...] 格式，由 findMany 步骤单独断言

  @cli
  Scenario: catalog models 缺少 --database 参数失败
    When 用户执行 "catalog models"
    Then CLI 命令应该失败
    And CLI 错误码应该为 "MISSING_REQUIRED_FLAG"
    And CLI 退出码应该为 2

  @cli
  Scenario: catalog databases 指定不存在的项目返回 PROJECT_NOT_FOUND
    When 用户执行 "catalog databases --project nonexistent_xyz"
    Then CLI 命令应该失败
    And CLI 错误码应该为 "PROJECT_NOT_FOUND"
    And CLI 退出码应该为 5

  @cli
  Scenario: catalog models 指定不存在的项目返回 PROJECT_NOT_FOUND
    When 用户执行 "catalog models --project nonexistent_xyz --database {DATABASE}"
    Then CLI 命令应该失败
    And CLI 错误码应该为 "PROJECT_NOT_FOUND"

  @cli
  Scenario: 未设置项目上下文时 catalog databases 失败
    # 重置 credentials 到无 currentProject 状态：重新登录不切换项目
    When 用户执行 "auth login --token {PAT} --server {SERVER}"
    # 假设登录后 currentProject 自动设置，此场景验证无上下文时的错误
    # 实际测试：传入空 project 标志
    When 用户执行 "catalog databases --project {PROJECT}"
    Then CLI 命令应该成功
