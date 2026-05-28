# Feature: 数据库接管管理
# 覆盖 model_database 的接管、更新、取消接管等核心场景。
# 前置条件：项目已配置数据库集群，测试环境 MySQL 上存在 test_db 库。

Feature: 数据库接管管理

  Background:
    Given 我以管理员身份登录

  @smoke
  Scenario: 成功接管一个自建数据库
    When 我接管数据库 "test_db"，模式为 "SELF_HOSTED"，友好名称为 "测试库"
    Then 接管应该成功
    And 已接管数据库列表中应包含名为 "test_db" 的数据库
    And 该数据库的模式应该是 "SELF_HOSTED"

  Scenario: 成功接管一个托管数据库
    When 我接管数据库 "test_db"，模式为 "MANAGED"，友好名称为 "托管测试库"
    Then 接管应该成功
    And 已接管数据库列表中应包含名为 "test_db" 的数据库
    And 该数据库的模式应该是 "MANAGED"

  Scenario: 重复接管同一数据库应报错
    Given 已接管数据库 "test_db"，模式为 "SELF_HOSTED"
    When 我再次接管数据库 "test_db"，模式为 "SELF_HOSTED"
    Then 应该返回接管错误 "InvalidInput"

  Scenario: 列出已接管数据库
    Given 已接管数据库 "test_db"，模式为 "SELF_HOSTED"
    When 我查询已接管数据库列表
    Then 列表中应包含名为 "test_db" 的数据库

  Scenario: 更新数据库友好名称和描述
    Given 已接管数据库 "test_db"，模式为 "SELF_HOSTED"
    When 我将该数据库的友好名称更新为 "新名称"，描述更新为 "新描述"
    Then 更新应该成功
    And 查询该数据库，友好名称应该是 "新名称"
    And 查询该数据库，描述应该是 "新描述"

  Scenario: 将自建数据库切换为托管模式
    Given 已接管数据库 "test_db"，模式为 "SELF_HOSTED"
    When 我将该数据库的模式更新为 "MANAGED"
    Then 更新应该成功
    And 查询该数据库，模式应该是 "MANAGED"

  Scenario: 取消接管数据库
    Given 已接管数据库 "test_db"，模式为 "SELF_HOSTED"
    When 我取消接管该数据库
    Then 取消接管应该成功
    And 已接管数据库列表中不应包含名为 "test_db" 的数据库

  Scenario: 取消接管后可重新接管
    Given 已接管数据库 "test_db"，模式为 "SELF_HOSTED"
    And 我取消接管该数据库
    When 我接管数据库 "test_db"，模式为 "MANAGED"，友好名称为 "重新接管"
    Then 接管应该成功
    And 该数据库的模式应该是 "MANAGED"

  Scenario: 查询集群原始数据库列表
    When 我查询集群原始数据库列表
    Then 原始列表中应包含 "test_db"
    And 原始列表中不应包含系统库 "information_schema"
    And 原始列表中不应包含系统库 "mysql"
