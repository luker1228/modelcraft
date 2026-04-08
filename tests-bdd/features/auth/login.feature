# tests-bdd/features/auth/login.feature
Feature: 用户登录

  Background:
    Given 已注册手机号 "13800138001" 密码 "password123"

  @smoke
  Scenario: 成功登录
    When 我用手机号 "13800138001" 和密码 "password123" 登录
    Then 登录应该成功
    And 响应中应包含 userId
    And 响应中应包含 refreshToken
    And 响应中应包含 expiresAt

  Scenario: 手机号不存在时报错
    When 我用手机号 "19999999999" 和密码 "password123" 登录
    Then 应该返回 HTTP 状态码 401
    And 应该返回错误码 "AUTHENTICATION_FAILED"

  Scenario: 密码错误时报错
    When 我用手机号 "13800138001" 和密码 "wrongpassword" 登录
    Then 应该返回 HTTP 状态码 401
    And 应该返回错误码 "AUTHENTICATION_FAILED"

  Scenario Outline: 登录参数格式不合法时报错
    When 我用手机号 "<phone>" 和密码 "<password>" 登录
    Then 应该返回 HTTP 状态码 400
    And 应该返回错误码 "PARAM_INVALID.AUTH"

    Examples:
      | phone       | password    |
      | 1380013800  | password123 |
      | abcdefghijk | password123 |
