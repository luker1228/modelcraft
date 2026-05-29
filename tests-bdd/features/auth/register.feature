# tests-bdd/features/auth/register.feature
Feature: 用户注册

  Scenario: 成功注册新用户
    When 我用手机号 "13800138000" 和密码 "password123" 注册
    Then 注册应该成功
    And 响应中应包含 userId
    And 响应中应包含 orgName

  Scenario: 注册后可以登录
    Given 已注册手机号 "13900139000" 密码 "password123"
    When 我用手机号 "13900139000" 和密码 "password123" 登录
    Then 登录应该成功

  # 回归：注册用户应拥有管理员身份（user_orgs.is_admin = 1）
  # 修复 bug：CreateMembership SQL 漏写 is_admin 字段，导致注册用户 is_admin = 0
  Scenario: 注册后登录的 JWT 包含管理员身份
    Given 已注册手机号 "13800138001" 密码 "password123"
    When 我用手机号 "13800138001" 和密码 "password123" 登录
    Then 登录应该成功
    And 登录 JWT 中 is_admin 应为 true

  Scenario: 手机号已注册时报错
    Given 已注册手机号 "13700137000" 密码 "password123"
    When 我用手机号 "13700137000" 和密码 "newpassword1" 注册
    Then 应该返回 HTTP 状态码 409
    And 应该返回错误码 "CONFLICT.USER"

  Scenario: 用户名已被占用时报错
    Given 已注册手机号 "13700137001" 用户名 "dupuser" 密码 "password123"
    When 我用手机号 "13700137002" 和用户名 "dupuser" 和密码 "password123" 注册
    Then 应该返回 HTTP 状态码 409
    And 应该返回错误码 "CONFLICT.USER"

  Scenario Outline: 手机号格式不合法时报错
    When 我用手机号 "<phone>" 和密码 "password123" 注册
    Then 应该返回 HTTP 状态码 400
    And 应该返回错误码 "PARAM_INVALID.AUTH"

    Examples:
      | phone        |
      | 1380013800   |
      | 138001380000 |
      | 23800138000  |
      | abcdefghijk  |
      |              |

  Scenario Outline: 密码不满足要求时报错
    When 我用手机号 "13600136000" 和密码 "<password>" 注册
    Then 应该返回 HTTP 状态码 400
    And 应该返回错误码 "PARAM_INVALID.AUTH"

    Examples:
      | password |
      | 1234567  |
      | short   |
      | abc     |
      |         |

  Scenario Outline: 用户名不满足要求时报错
    When 我用手机号 "13600136001" 和用户名 "<userName>" 和密码 "password123" 注册
    Then 应该返回 HTTP 状态码 400
    And 应该返回错误码 "PARAM_INVALID.AUTH"

    Examples:
      | userName                          |
      | ab                                |
      | 1abc                              |
      | abc@def                           |
      | admin                             |
      | aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa |
