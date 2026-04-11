Feature: Profile 分表能力

  Scenario: phone+userName+password 注册成功后自动创建 profile 并应用默认昵称
    When 我用唯一手机号 "13810138000" 和唯一用户名 "profile_user_a" 和密码 "password123" 注册
    Then 注册应该成功
    And 响应中应包含 userId
    And 响应中应包含 orgName
    And 响应中应包含 profile 快照
    And 响应中的 profile.userId 应与 userId 一致
    And 响应中的 profile.nickname 应匹配默认规则

  Scenario: updateMyProfile 按 PATCH 语义更新
    Given 已注册唯一手机号 "13810138001" 唯一用户名 "profile_user_b" 密码 "password123"
    And 我使用当前注册用户的访问令牌
    And 我调用 updateMyProfile 设置完整资料，昵称 "old_nick" 头像 "mock://avatar/default-1.png" 简介 "old bio"
    When 我调用 updateMyProfile 仅更新昵称为 "new_nick"
    Then updateMyProfile 应该成功
    And 返回的 profile.nickname 应为 "new_nick"
    And 返回的 profile.avatarUrl 应为 "mock://avatar/default-1.png"
    And 返回的 profile.bio 应为 "old bio"
    When 我查询 myUserProfile
    Then myUserProfile 应返回用户资料
    And myUserProfile.user.profile.nickname 应为 "new_nick"
    And myUserProfile.user.profile.avatarUrl 应为 "mock://avatar/default-1.png"
    And myUserProfile.user.profile.bio 应为 "old bio"

  Scenario: myUserProfile 在 profile 缺失时返回 ProfileNotFound
    Given 存在一个仅有 user 无 profile 的用户
    And 我使用该用户的访问令牌并初始化组织
    When 我查询 myUserProfile
    Then myUserProfile 应返回错误类型 "ProfileNotFound"

  Scenario: me 查询兼容性（行为不变）
    Given 已注册唯一手机号 "13810138002" 唯一用户名 "profile_user_c" 密码 "password123"
    And 我使用当前注册用户的访问令牌
    When 我查询 me
    Then me 查询应成功
    And me.id 应等于当前用户ID
    And me 应包含 externalID 字段
    And me 应包含 email 字段
    And me 应包含 name 字段
    And me.permissions 应为数组
