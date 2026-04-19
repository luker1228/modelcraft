# Policy 配置 BDD 测试
# 对应 PRD: ai-metadata/prd/rls/05-policy-configuration.md
# 验收标准: AC-9, AC-10, AC-11, AC-12, AC-13

Feature: Policy 配置
  作为开发者
  我希望能配置 Model 的 RLS 策略
  以便灵活控制数据访问权限

  Background:
    Given 我以管理员身份登录

  @smoke
  Scenario: 查询 Model 的 RLS 策略
    Given 已创建名为 "PolicyTestModel" 的模型
    When 我查询该模型的 RLS 策略
    Then 应该返回默认的 READ_WRITE_OWNER 策略
    And Policy 的 selectPredicate 应该为 '{"owner":{"_eq":{"_auth":"uid"}}}'
    And Policy 的 insertCheck 应该为 '{"owner":{"_eq":{"_auth":"uid"}}}'
    And Policy 的 preset 应该为 "READ_WRITE_OWNER"

  Scenario: 设置 Policy 为 READ_ALL_WRITE_OWNER
    Given 已创建名为 "ReadAllWriteOwnerModel" 的模型
    When 我将该模型的 Policy 设置为 READ_ALL_WRITE_OWNER
    Then Policy 设置成功
    And Policy 的 selectPredicate 应该为 "true"
    And Policy 的 insertCheck 应该为 '{"owner":{"_eq":{"_auth":"uid"}}}'

  Scenario: 设置 Policy 为 READ_ALL
    Given 已创建名为 "ReadAllModel" 的模型
    When 我将该模型的 Policy 设置为 READ_ALL
    Then Policy 设置成功
    And Policy 的 selectPredicate 应该为 "true"
    And Policy 的 insertCheck 应该为 "false"
    And Policy 的 updatePredicate 应该为 "false"

  Scenario: 设置 Policy 为 READ_WRITE_ALL（高危策略）
    Given 已创建名为 "ReadWriteAllModel" 的模型
    When 开发者确认风险并将该模型的 Policy 设置为 READ_WRITE_ALL
    Then Policy 设置成功
    And 所有谓词都应该为 "true"

  Scenario: 设置 Policy 为 NO_ACCESS
    Given 已创建名为 "NoAccessModel" 的模型
    When 我将该模型的 Policy 设置为 NO_ACCESS
    Then Policy 设置成功
    And 所有谓词都应该为 "false"

  Scenario: 验证 RLS 表达式合法性
    Given 已创建名为 "ValidateExprModel" 的模型
    When 我验证表达式 '{"owner":{"_eq":{"_auth":"uid"}}}' 用于 selectPredicate
    Then 验证应该通过

  Scenario: 验证包含未声明 auth 变量的表达式
    Given 已创建名为 "ValidateExprModel" 的模型
    When 我验证表达式 '{"owner":{"_eq":{"_auth":"undeclared_var"}}}' 用于 selectPredicate
    Then 验证应该失败并返回错误类型 "InvalidAuthVariable"

  Scenario: insertCheck 中不允许使用 _exists 操作符
    Given 已创建名为 "ValidateExprModel" 的模型
    When 我验证表达式 '{"_exists":{"model":"test","where":{}}}' 用于 insertCheck
    Then 验证应该失败并返回错误类型 "InvalidRLSExpression"

  Scenario: 无 owner 字段的模型无法设置 Policy
    Given 已创建名为 "NoOwnerFieldModel" 的模型
    And 我删除名为 "owner" 的字段
    When 我尝试为该模型设置 Policy
    Then 操作失败并返回错误类型 "ModelHasNoOwnerField"

  Scenario: 删除 owner 字段后 Policy 同步删除
    Given 已创建名为 "CascadeDeleteModel" 的模型
    And 我将该模型的 Policy 设置为 READ_ALL_WRITE_OWNER
    When 我删除名为 "owner" 的字段
    Then 字段应该删除成功
    And 该模型应该不存在 RLS 策略

  # AC-13: auth_schema 配置
  Scenario: 配置 Project 的 auth_schema
    When 我为项目设置 auth_schema，添加变量 "tenant_id"，source 为 "jwt.tenant_id"，type 为 "UUID"
    Then auth_schema 设置成功
    And 项目应该包含 auth 变量 "tenant_id"

  Scenario: 使用已声明的 auth 变量验证表达式
    Given 已为项目设置 auth_schema，包含变量 "tenant_id"
    And 已创建名为 "AuthVarModel" 的模型
    When 我验证表达式 '{"owner":{"_eq":{"_auth":"tenant_id"}}}' 用于 selectPredicate
    Then 验证应该通过

  Scenario: 使用未声明的 auth 变量验证表达式失败
    Given 已创建名为 "AuthVarModel" 的模型
    When 我验证表达式 '{"owner":{"_eq":{"_auth":"undeclared_var"}}}' 用于 selectPredicate
    Then 验证应该失败

  Scenario: 设置自定义的 RLS 策略（非 Preset）
    Given 已创建名为 "CustomPolicyModel" 的模型
    When 我设置该模型的 RLS 策略为以下五件套:
      | predicateType    | expression                             |
      | selectPredicate  | {"owner":{"_eq":{"_auth":"uid"}}}    |
      | insertCheck      | {"owner":{"_eq":{"_auth":"uid"}}}    |
      | updatePredicate  | {"owner":{"_eq":{"_auth":"uid"}}}    |
      | updateCheck      | {"owner":{"_eq":{"_auth":"uid"}}}    |
      | deletePredicate  | {"_and":[{"owner":{"_eq":{"_auth":"uid"}}},{"status":{"_eq":"active"}}]} |
    Then Policy 设置成功
    And preset 应该为 null（自定义策略）
