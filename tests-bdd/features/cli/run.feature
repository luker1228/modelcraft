# 功能：mc run 命令验收测试
# 覆盖：findMany/findFirst/findUnique/count/aggregate、where 过滤、orderBy、分页、错误处理

@cli
Feature: CLI 运行时 GraphQL 查询

  Background:
    Given 用户已通过 PAT token 登录 CLI
    And 用户已切换到项目 "luke"

  # ─── findMany ────────────────────────────────────────────────────

  @cli @smoke
  Scenario: findMany 基础查询返回数据和 totalCount
    When 用户执行 mc run "{PROJECT}.{DATABASE}.{MODEL}" 查询 "{ findMany(take: 5) { items { id username email } totalCount } }"
    Then CLI 命令应该成功
    And CLI data."findMany" 应该不为空
    And CLI data.totalCount 应该大于 0
    And CLI findMany 第一条记录应该包含字段 "id"
    And CLI findMany 第一条记录应该包含字段 "username"

  @cli
  Scenario: findMany take 参数限制返回数量
    When 用户执行 mc run "{PROJECT}.{DATABASE}.{MODEL}" 查询 "{ findMany(take: 3) { items { id } totalCount } }"
    Then CLI 命令应该成功
    And CLI data.items 数量应该等于 3
    And CLI data.totalCount 应该大于 3

  @cli
  Scenario: findMany take=0 返回空列表但 totalCount 正确
    When 用户执行 mc run "{PROJECT}.{DATABASE}.{MODEL}" 查询 "{ findMany(take: 0) { items { id } totalCount } }"
    Then CLI 命令应该成功
    And CLI data.items 数量应该等于 0
    And CLI data.totalCount 应该大于 0

  @cli
  Scenario: findMany orderBy 降序排列
    When 用户执行 mc run "{PROJECT}.{DATABASE}.{MODEL}" 查询 "{ findMany(take: 3, orderBy: [{username: desc}]) { items { id username } totalCount } }"
    Then CLI 命令应该成功
    And CLI data.items 数量应该等于 3

  @cli
  Scenario: findMany skip/take 分页
    When 用户执行 mc run "{PROJECT}.{DATABASE}.{MODEL}" 查询 "{ findMany(skip: 5, take: 3) { items { id } totalCount } }"
    Then CLI 命令应该成功
    And CLI data.items 数量应该等于 3
    And CLI data.totalCount 应该大于 5

  # ─── where 过滤 ──────────────────────────────────────────────────

  @cli @smoke
  Scenario: findMany where equals 过滤
    When 用户执行 mc run "{PROJECT}.{DATABASE}.{MODEL}" 查询 "{ findMany(where: { username: { equals: \"alice\" } }) { items { username } totalCount } }"
    Then CLI 命令应该成功
    And CLI data.totalCount 应该等于 1
    And CLI data.items 数量应该等于 1

  @cli
  Scenario: findMany where contains 过滤
    When 用户执行 mc run "{PROJECT}.{DATABASE}.{MODEL}" 查询 "{ findMany(where: { email: { contains: \"example\" } }) { items { email } totalCount } }"
    Then CLI 命令应该成功
    And CLI data.totalCount 应该大于 0

  @cli
  Scenario: findMany where startsWith 过滤
    When 用户执行 mc run "{PROJECT}.{DATABASE}.{MODEL}" 查询 "{ findMany(where: { username: { startsWith: \"a\" } }) { items { username } totalCount } }"
    Then CLI 命令应该成功
    And CLI data.totalCount 应该大于 0

  @cli
  Scenario: findMany where in 列表过滤
    When 用户执行 mc run "{PROJECT}.{DATABASE}.{MODEL}" 查询 "{ findMany(where: { username: { in: [\"alice\", \"bob\"] } }) { items { username } totalCount } }"
    Then CLI 命令应该成功
    And CLI data.totalCount 应该等于 2

  @cli
  Scenario: findMany where not 排除
    When 用户执行 mc run "{PROJECT}.{DATABASE}.{MODEL}" 查询 "{ findMany(where: { username: { not: \"alice\" } }) { items { username } totalCount } }"
    Then CLI 命令应该成功
    And CLI data.totalCount 应该大于 0

  @cli
  Scenario: findMany where AND 逻辑
    When 用户执行 mc run "{PROJECT}.{DATABASE}.{MODEL}" 查询 "{ findMany(where: { AND: [{ username: { startsWith: \"a\" } }, { email: { contains: \"example\" } }] }) { items { username } totalCount } }"
    Then CLI 命令应该成功
    And CLI data.totalCount 应该大于 0

  @cli
  Scenario: findMany where OR 逻辑
    When 用户执行 mc run "{PROJECT}.{DATABASE}.{MODEL}" 查询 "{ findMany(where: { OR: [{ username: { equals: \"alice\" } }, { username: { equals: \"bob\" } }] }) { items { username } totalCount } }"
    Then CLI 命令应该成功
    And CLI data.totalCount 应该等于 2

  @cli
  Scenario: findMany where NOT 逻辑排除单条件
    When 用户执行 mc run "{PROJECT}.{DATABASE}.{MODEL}" 查询 "{ findMany(where: { NOT: { username: { equals: \"alice\" } } }) { items { username } totalCount } }"
    Then CLI 命令应该成功
    And CLI data.totalCount 应该大于 0

  # ─── 其他查询操作 ─────────────────────────────────────────────────

  @cli @smoke
  Scenario: count 返回总行数
    When 用户执行 mc run "{PROJECT}.{DATABASE}.{MODEL}" 查询 "{ count { count } }"
    Then CLI 命令应该成功
    And CLI data."count" 应该不为空

  @cli
  Scenario: count 带 where 条件
    When 用户执行 mc run "{PROJECT}.{DATABASE}.{MODEL}" 查询 "{ count(where: { username: { equals: \"alice\" } }) { count } }"
    Then CLI 命令应该成功

  @cli
  Scenario: count 带 select 字段级计数
    When 用户执行 mc run "{PROJECT}.{DATABASE}.{MODEL}" 查询 "{ count(select: { _all: true, email: true }) { fieldsCount { _all email } } }"
    Then CLI 命令应该成功
    And CLI 响应应该包含字段 "count"

  @cli
  Scenario: count 混用 count 和 select 报错
    When 用户执行 mc run "{PROJECT}.{DATABASE}.{MODEL}" 查询 "{ count(select: { _all: true }) { count fieldsCount { _all } } }"
    Then CLI 命令应该失败
    And CLI 错误码应该为 "GRAPHQL_ERROR"
    And CLI 错误信息应该包含 "mutually exclusive"

  @cli
  Scenario: aggregate 聚合查询
    When 用户执行 mc run "{PROJECT}.{DATABASE}.{MODEL}" 查询 "{ aggregate(_count: { _all: true }) { _count { _all } } }"
    Then CLI 命令应该成功

  @cli
  Scenario: findFirst 返回第一条匹配记录
    When 用户执行 mc run "{PROJECT}.{DATABASE}.{MODEL}" 查询 "{ findFirst(where: { username: { startsWith: \"a\" } }) { item { id username } } }"
    Then CLI 命令应该成功
    And CLI 响应应该包含字段 "findFirst"

  @cli
  Scenario: findUnique 按唯一键查找
    When 用户执行 mc run "{PROJECT}.{DATABASE}.{MODEL}" 查询 "{ findUnique(where: { id: \"u001\" }) { item { id username } } }"
    Then CLI 命令应该成功
    And CLI 响应应该包含字段 "findUnique"

  # ─── 错误场景 ─────────────────────────────────────────────────────

  @cli
  Scenario: 查询不存在的模型返回 NOT_FOUND
    When 用户执行 "run luke.{DATABASE}.nonexistent_model_xyz '{ count { count } }'"
    Then CLI 命令应该失败
    And CLI 错误码应该为 "NOT_FOUND"
    And CLI 退出码应该为 5

  @cli
  Scenario: 未提供 query 时报错
    When 用户执行 "run {PROJECT}.{DATABASE}.{MODEL}"
    # mc 会从 stdin 读取，但 stdin 是 pipe 且无内容，应报 MISSING_REQUIRED_FLAG
    Then CLI 命令应该失败
    And CLI 错误码应该为 "MISSING_REQUIRED_FLAG"

  @cli
  Scenario: 路径格式错误返回 INVALID_RESOURCE_PATH
    When 用户执行 "run {MODEL} '{ count { count } }'"
    Then CLI 命令应该失败
    And CLI 退出码应该为 6

  @cli
  Scenario: 语法错误的 GraphQL 返回 INVALID_ARGUMENT 而非 500
    When 用户执行 "run {PROJECT}.{DATABASE}.{MODEL} '{ findMany( '"
    Then CLI 命令应该失败
    And CLI 错误码应该为 "INVALID_ARGUMENT"
    And CLI 退出码应该为 2

  @cli
  Scenario: 只读模型拒绝 mutation
    When 用户执行 mc run "{PROJECT}.{DATABASE}.{MODEL}" 查询 "mutation { create(data: {}) { id } }"
    Then CLI 命令应该失败
    And CLI 错误码应该为 "PERMISSION_DENIED"

  # ─── 2-part path 快捷写法 ─────────────────────────────────────────

  @cli
  Scenario: 使用 2-part path（数据库.模型）省略项目
    When 用户执行 "run {DATABASE}.{MODEL} '{ count { count } }'"
    Then CLI 命令应该成功
