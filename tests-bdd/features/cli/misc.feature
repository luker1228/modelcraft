# 功能：mc schema/version 命令以及 exit code 契约测试
# 覆盖：schema commands 结构、version 格式、exit code 约定

@cli
Feature: CLI 工具命令与 Exit Code 契约

  Background:
    Given 用户已通过 PAT token 登录 CLI
    And 用户已切换到项目 "luke"

  # ─── version ──────────────────────────────────────────────────────

  @cli @smoke
  Scenario: version 命令返回版本信息
    When 用户执行 "version"
    Then CLI 命令应该成功
    And CLI 响应应该包含字段 "version"
    And CLI 响应应该包含字段 "commit"
    And CLI 响应应该包含字段 "buildTime"

  # ─── schema commands ──────────────────────────────────────────────

  @cli @smoke
  Scenario: schema commands 包含所有 auth 子命令
    When 用户执行 "schema commands"
    Then CLI 命令应该成功
    And CLI 响应应该包含字段 "commands"

  # ─── exit code 契约 ───────────────────────────────────────────────

  @cli
  Scenario: 缺少必要参数退出码为 2
    When 用户执行 "catalog models"
    Then CLI 退出码应该为 2

  @cli
  Scenario: 未认证退出码为 3
    # 此场景重新创建一个新的临时 credentials 文件（不登录），验证 exit 3
    # Note: 由于 Background 已登录，使用额外标志使 step 逻辑创建新临时文件
    Given 用户未登录 CLI
    When 用户执行 "auth status"
    Then CLI 退出码应该为 3

  @cli
  Scenario: 资源路径格式错误退出码为 6
    When 用户执行 "describe {MODEL}"
    Then CLI 退出码应该为 6

  @cli
  Scenario: 资源不存在退出码为 5
    When 用户执行 "auth switch-project nonexistent_xyz"
    Then CLI 退出码应该为 5

  @cli
  Scenario: 未知命令退出码为 2
    When 用户执行 "unknown_command_xyz"
    Then CLI 退出码应该为 2

  # ─── 路径解析 ──────────────────────────────────────────────────────

  @cli
  Scenario: 路径含空格会被 trim 后正常解析
    # "luke .demo_ecommerce.users" 的空格应被 trim 为 "luke"
    When 用户执行 "describe 'luke .{DATABASE}.{MODEL}'"
    Then CLI 命令应该成功
    And CLI meta."project" 应该等于 "luke"
