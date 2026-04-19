# Runtime JWT 认证 BDD 测试
# 对应 PRD: ai-metadata/prd/rls/03-runtime-jwt-auth.md
# 验收标准: AC-6

Feature: Runtime JWT 认证
  作为终端用户
  我只能使用 EndUser JWT 访问 Runtime API
  以便确保数据隔离的安全性

  Background:
    Given 存在已配置数据库集群的组织 "testorg" 和项目 "testproject"
    And 我以管理员身份登录
    And 已创建名为 "JWTTestModel" 的模型

  @smoke
  Scenario: EndUser JWT 可以访问 Runtime
    Given 已存在终端用户 "runtimeuser"，密码 "Pass1234"
    And 终端用户 "runtimeuser" 已登录
    When 终端用户调用 Runtime 查询 "JWTTestModel"
    Then Runtime 应该返回成功响应

  Scenario: Developer JWT 无法访问 Runtime
    Given 我已获取开发者 JWT
    When 开发者使用 JWT 调用 Runtime 查询 "JWTTestModel"
    Then Runtime 应该返回 401 未授权错误

  Scenario: 无效的 JWT 无法访问 Runtime
    When 使用无效的 JWT 调用 Runtime 查询 "JWTTestModel"
    Then Runtime 应该返回 401 未授权错误

  Scenario: 无 JWT 无法访问 Runtime
    When 不带 JWT 调用 Runtime 查询 "JWTTestModel"
    Then Runtime 应该返回 401 未授权错误
