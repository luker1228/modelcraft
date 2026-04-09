Feature: 组织初始化

  Scenario: memberships 为空时初始化组织
    Given 我已登录并持有 access token
    And 当前用户没有任何组织 memberships
    When 我初始化组织 displayName 为 "BDD 初始化组织"
    Then 组织初始化应该成功
    And 初始化结果 should have alreadyExists false
    And 当前用户 memberships 数量应为 1

  Scenario: 组织初始化可重入，只会创建一个组织
    Given 我已登录并持有 access token
    And 当前用户没有任何组织 memberships
    When 我首次初始化组织 displayName 为 "BDD 幂等组织"
    And 我再次使用相同 displayName 初始化组织
    Then 第二次初始化结果 should have alreadyExists true
    And 两次初始化返回的 orgName 应相同
    And 当前用户 memberships 数量应为 1
