Feature: 组织初始化

  Scenario: 创建组织时自动生成 builtin admin 用户
    Given 我已登录并持有 access token
    When 我初始化组织 displayName 为 "BDD Builtin Test Org" 并设置 endUserAdminPassword 为 "Admin@12345"
    Then 组织初始化应该成功
    And 该组织下应存在一个 isBuiltin 为 true 的终端用户
    And 该 builtin 用户的用户名应为 "admin"
    And 该 builtin 用户的 isForbidden 应为 false
