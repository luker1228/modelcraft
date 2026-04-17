Feature: 组织级 API Key 角色绑定

  @apikey
  Scenario: 创建 API Key 时可绑定角色
    Given 我已登录并持有 access token
    And 存在一个可绑定的组织角色
    When 我创建一个绑定该角色的 API Key
    Then API Key 应创建成功且角色包含该角色

  @apikey
  Scenario: 更新 API Key 时可修改角色绑定
    Given 我已登录并持有 access token
    And 存在一个可绑定的组织角色
    And 我创建一个绑定该角色的 API Key
    When 我将该 API Key 的角色更新为空列表
    Then API Key 角色应更新为空
