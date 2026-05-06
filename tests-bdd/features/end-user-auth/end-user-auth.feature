# End-User Auth 终端用户认证 BDD 测试
# 测试后端内部接口 /internal/end-user/*

Feature: 终端用户认证管理
  作为开发者或终端用户
  我希望能管理终端用户账号（注册、登录、禁用、删除等）
  以便为项目提供终端用户认证能力

  Background:
    Given 存在已配置数据库集群的组织 "testorg" 和项目 "testproject"
    And 我以开发者身份登录

  # ==================== 用户管理（GraphQL/开发者侧）====================

  Scenario: 开发者成功创建终端用户
    When 我创建终端用户，用户名为 "alice"，密码为 "Pass1234"
    Then 终端用户应该创建成功
    And 返回的用户名应该是 "alice"
    And 用户状态应该是启用

  Scenario: 创建终端用户时用户名已存在
    Given 已存在终端用户 "bob"，密码 "Pass1234"
    When 我创建终端用户，用户名为 "bob"，密码为 "Pass5678"
    Then 返回终端用户错误码 "CONFLICT"

  Scenario: 创建终端用户时密码强度不足
    When 我创建终端用户，用户名为 "weakuser"，密码为 "123"
    Then 返回终端用户错误码 "PARAM_INVALID"

  Scenario: 创建终端用户时用户名格式不合法
    When 我创建终端用户，用户名为 "has space"，密码为 "Pass1234"
    Then 返回终端用户错误码 "PARAM_INVALID"

  Scenario: 开发者分页查询终端用户列表
    Given 已存在终端用户 "user1"，密码 "Pass1234"
    And 已存在终端用户 "user2"，密码 "Pass1234"
    When 我查询终端用户列表，每页 10 条
    Then 应该返回用户列表
    And 列表中应该包含用户 "user1"
    And 列表中应该包含用户 "user2"
    And 总用户数应该大于等于 2

  Scenario: 开发者搜索终端用户
    Given 已存在终端用户 "searchable_user"，密码 "Pass1234"
    When 我搜索终端用户，关键词为 "searchable"
    Then 应该返回用户列表
    And 列表中应该包含用户 "searchable_user"

  Scenario: 开发者禁用终端用户
    Given 已存在终端用户 "tobeforbidden"，密码 "Pass1234"
    When 我禁用终端用户 "tobeforbidden"
    Then 用户状态更新成功
    And 用户状态应该是禁用

  Scenario: 开发者启用被禁用的终端用户
    Given 已存在被禁用的终端用户 "disableduser"
    When 我启用终端用户 "disableduser"
    Then 用户状态更新成功
    And 用户状态应该是启用

  Scenario: 开发者删除终端用户
    Given 已存在终端用户 "tobedeleted"，密码 "Pass1234"
    When 我删除终端用户 "tobedeleted"
    Then 终端用户应该删除成功
    And 该用户应该不存在

  # ==================== 终端用户自助认证（OpenAPI/终端用户侧）====================

  Scenario: 终端用户自助注册
    When 终端用户自助注册，用户名为 "selfreg"，密码为 "Pass1234"
    Then 注册成功并返回 refresh token
    And 返回的 token 有效期为 3600 秒

  Scenario: 终端用户使用正确凭证登录
    Given 已存在终端用户 "logintest"，密码 "Pass1234"
    When 终端用户登录，用户名为 "logintest"，密码为 "Pass1234"
    Then 登录成功并返回 access token
    And 返回 refresh token
    And 返回可访问项目列表
    And 可访问项目列表应包含 "testproject"

  Scenario: 终端用户使用错误密码登录
    Given 已存在终端用户 "wrongpwd"，密码 "Pass1234"
    When 终端用户登录，用户名为 "wrongpwd"，密码为 "WrongPass123"
    Then 返回终端用户错误码 "INVALID_CREDENTIALS"
    And HTTP 状态码应该是 401

  Scenario: 终端用户使用错误用户名登录
    When 终端用户登录，用户名为 "nonexistent"，密码为 "Pass1234"
    Then 返回终端用户错误码 "INVALID_CREDENTIALS"

  Scenario: 被禁用的终端用户无法登录
    Given 已存在被禁用的终端用户 "forbiddenuser"
    When 终端用户登录，用户名为 "forbiddenuser"，密码为 "Pass1234"
    Then 返回终端用户错误码 "ACCOUNT_DISABLED"
    And HTTP 状态码应该是 403

  Scenario: 终端用户获取自己的信息
    Given 终端用户 "meuser" 已登录
    When 终端用户查询自己的信息
    Then 应该返回用户信息
    And 返回的用户名应该是 "meuser"
    And 返回的信息应该包含创建时间

  Scenario: 终端用户登出
    Given 终端用户 "logoutuser" 已登录
    When 终端用户登出
    Then 登出成功

  # ==================== Token 刷新 ====================

  Scenario: 终端用户刷新 access token
    Given 终端用户 "refreshtest" 已登录并持有 refresh token
    When 终端用户刷新 token
    Then 刷新成功并返回新的 access token
    And 返回新的 refresh token

  Scenario: 使用已撤销的 refresh token 刷新
    Given 终端用户 "revokedtest" 已登录并持有 refresh token
    And 该用户的 refresh token 已被撤销
    When 终端用户刷新 token
    Then 返回终端用户错误码 "INVALID_REFRESH_TOKEN"
    And HTTP 状态码应该是 401

  Scenario: 被禁用的终端用户无法刷新 token
    Given 已存在并登录后被禁用的终端用户 "disabledrefresh"
    When 使用该用户的 refresh token 刷新
    Then 返回终端用户错误码 "ACCOUNT_DISABLED"
    And HTTP 状态码应该是 403

  Scenario: 终端用户选择已授权项目上下文
    Given 终端用户 "projectselector" 已登录并持有 refresh token
    When 终端用户选择项目上下文 "testproject"
    Then 选择项目成功并返回用户信息
    And 返回已选择项目 "testproject"
    And 返回中不应包含 access token

  Scenario: 被禁用的终端用户无法选择项目上下文
    Given 已存在并登录后被禁用的终端用户 "disabledselect"
    When 使用该用户的 refresh token 选择项目 "testproject"
    Then 返回终端用户错误码 "ACCOUNT_DISABLED"
    And HTTP 状态码应该是 403

  # ==================== 删除用户后 Session 清理 ====================

  Scenario: 删除终端用户后其 session 失效
    Given 终端用户 "deleteduser" 已登录并持有 refresh token
    When 开发者删除终端用户 "deleteduser"
    Then 终端用户应该删除成功
    When 使用该用户的 refresh token 刷新
    Then 返回终端用户错误码 "INVALID_REFRESH_TOKEN"

  # ==================== findUsers / me GraphQL 查询 ====================

  Scenario: 开发者通过 findUsers 查询终端用户列表
    Given 已存在终端用户 "findusers_dev1"，密码 "Pass1234"
    And 已存在终端用户 "findusers_dev2"，密码 "Pass1234"
    When 开发者调用 findUsers 查询，take 为 20
    Then findUsers 应该返回用户列表
    And findUsers 结果中应包含用户名 "findusers_dev1"
    And findUsers 结果中应包含用户名 "findusers_dev2"

  Scenario: 终端用户通过 findUsers 查询用户列表
    Given 已存在终端用户 "findusers_eu1"，密码 "Pass1234"
    And 终端用户 "findusers_eu1" 已登录
    When 终端用户调用 findUsers 查询，take 为 20
    Then findUsers 应该返回用户列表
    And findUsers 结果中应包含用户名 "findusers_eu1"

  Scenario: 终端用户调用 me 查询获取自己信息
    Given 已存在终端用户 "mequery_eu"，密码 "Pass1234"
    And 终端用户 "mequery_eu" 已登录
    When 终端用户调用 me 查询
    Then me 查询应该返回用户信息
    And me 查询返回的用户名应该是 "mequery_eu"

  Scenario: 被禁用的终端用户调用 me 查询返回错误
    Given 已存在并登录后被禁用的终端用户 "mequery_disabled2"
    When 终端用户调用 me 查询
    Then me 查询应该返回错误
    And me 查询错误码应该包含 "ACCOUNT_DISABLED"

  Scenario: 开发者调用 me 查询返回 INVALID_CALLER 错误
    When 开发者调用 me 查询
    Then me 查询应该返回错误
    And me 查询错误码应该包含 "INVALID_CALLER"

  Scenario: 终端用户无法调用 listProjectEndUsers（默认拒绝）
    Given 已存在终端用户 "directive_eu"，密码 "Pass1234"
    And 终端用户 "directive_eu" 已登录
    When 终端用户调用 listProjectEndUsers 查询
    Then 应该返回权限拒绝错误
    And 错误码应该是 "PERMISSION_DENIED"
