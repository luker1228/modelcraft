# tests-bdd/features/auth/token.feature
Feature: Token 管理（刷新与登出）

  Background:
    Given 已注册并登录手机号 "13800138002" 密码 "password123"

  Scenario: 成功刷新 Token
    When 我使用 refreshToken 刷新令牌
    Then 刷新应该成功
    And 响应中应包含 refreshToken
    And 响应中应包含 expiresAt
    And 新 refreshToken 应与旧 refreshToken 不同

  Scenario: 使用无效 refreshToken 刷新时报错
    When 我使用无效的 refreshToken 刷新令牌
    Then 应该返回 HTTP 状态码 401
    And 应该返回错误码 "UNAUTHORIZED"

  Scenario: 成功登出
    When 我使用 refreshToken 登出
    Then 登出应该成功

  Scenario: 登出后 refreshToken 失效
    When 我使用 refreshToken 登出
    And 我使用已登出的 refreshToken 刷新令牌
    Then 应该返回 HTTP 状态码 401
    And 应该返回错误码 "UNAUTHORIZED"
