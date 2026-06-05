# 功能：mc auth 命令验收测试
# 覆盖：PAT 登录、登出、刷新、状态查询、切换项目

@cli
Feature: CLI 认证管理

  Background: 使用隔离的临时 credentials 文件
    # @cli 标签会在 Before 钩子中创建临时 credentials 文件

  @cli @smoke
  Scenario: PAT token 登录成功
    When 用户执行 "auth login --token {PAT} --server {SERVER}"
    Then CLI 命令应该成功
    And CLI 退出码应该为 0
    And CLI 响应应该包含字段 "orgName"
    And CLI 响应应该包含字段 "projects"
    And CLI data."currentProject" 应该等于 "luke"

  @cli
  Scenario: PAT 登录后 currentProject 自动设为唯一项目
    # 用户只有一个项目时，登录后应自动设置 currentProject
    When 用户执行 "auth login --token {PAT} --server {SERVER}"
    Then CLI 命令应该成功
    And CLI data."currentProject" 应该不为空

  @cli
  Scenario: 使用无效 token 登录失败
    When 用户执行 "auth login --token mc_pat_invalid_token_000 --server {SERVER}"
    Then CLI 命令应该失败
    And CLI 错误码应该为 "UNAUTHENTICATED"
    And CLI 退出码应该为 3

  @cli
  Scenario: 未提供 username 时登录失败
    When 用户执行 "auth login --password somepassword"
    Then CLI 命令应该失败
    And CLI 错误码应该为 "MISSING_REQUIRED_FLAG"
    And CLI 退出码应该为 2

  @cli
  Scenario: 未提供 password 时登录失败
    When 用户执行 "auth login --username someuser"
    Then CLI 命令应该失败
    And CLI 错误码应该为 "MISSING_REQUIRED_FLAG"
    And CLI 退出码应该为 2

  @cli @smoke
  Scenario: 登录后 auth status 返回正确信息
    Given 用户已通过 PAT token 登录 CLI
    When 用户执行 "auth status"
    Then CLI 命令应该成功
    And CLI data."orgName" 应该不为空
    And CLI data."currentProject" 应该不为空
    And CLI 响应应该包含字段 "projects"

  @cli
  Scenario: auth refresh 响应包含 projects 字段
    Given 用户已通过 PAT token 登录 CLI
    When 用户执行 "auth refresh"
    Then CLI 命令应该成功
    And CLI 响应应该包含字段 "projects"
    And CLI 响应应该包含字段 "currentProject"

  @cli
  Scenario: 切换到可访问的项目
    Given 用户已通过 PAT token 登录 CLI
    When 用户执行 "auth switch-project {PROJECT}"
    Then CLI 命令应该成功
    And CLI data."currentProject" 应该等于 "luke"

  @cli
  Scenario: 切换到不存在的项目失败
    Given 用户已通过 PAT token 登录 CLI
    When 用户执行 "auth switch-project nonexistent_project_xyz"
    Then CLI 命令应该失败
    And CLI 错误码应该为 "PROJECT_NOT_FOUND"
    And CLI 退出码应该为 5

  @cli
  Scenario: 未登录时 auth status 报错
    Given 用户未登录 CLI
    When 用户执行 "auth status"
    Then CLI 命令应该失败
    And CLI 错误码应该为 "UNAUTHENTICATED"
    And CLI 退出码应该为 3

  @cli
  Scenario: 登出后命令不可用
    Given 用户已通过 PAT token 登录 CLI
    When 用户执行 "auth logout"
    Then CLI 命令应该成功
    When 用户执行 "catalog projects"
    Then CLI 命令应该失败
    And CLI 错误码应该为 "UNAUTHENTICATED"
