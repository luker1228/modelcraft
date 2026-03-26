# 5.2 模型产物（运行态）

> 代码位置：`internal/domain/modelruntime/`

## 概述

运行态的输入是已同步到目标数据库的模型结构，输出是动态生成的 GraphQL Schema 和对应的 SQL 执行能力。

## 核心实体

### RuntimeModel — 运行时模型快照

```
internal/domain/modelruntime/runtimemodel.go

RuntimeModel
├── ProjectSlug  string
├── DatabaseName string
├── ModelName    string
└── Fields       []RuntimeField   // 字段快照，用于生成 GraphQL Schema
```

运行态不直接使用 `modeldesign` 的实体，而是使用自己的 `RuntimeModel` 快照，确保两态解耦。

### 动态 GraphQL Schema

每个模型对应一套自动生成的 GraphQL 操作：

```
查询操作：
  findUnique     根据主键查询单条
  findFirst      条件查询第一条
  findMany       条件查询多条（含分页）
  count          统计数量
  aggregate      聚合查询（sum/avg/min/max）

变更操作：
  createOne      创建单条
  createMany     批量创建
  updateOne      更新单条
  updateMany     批量更新
  deleteOne      删除单条
  deleteMany     批量删除
```

### 查询能力

过滤条件（Prisma 风格）：

```
WhereInput 支持：
  equals / not
  in / notIn
  lt / lte / gt / gte
  contains / startsWith / endsWith（字符串）
  AND / OR / NOT（逻辑组合）
```

## 运行态入口

```
URL 格式：/:orgName/:projectSlug/:databaseName/:modelName

示例：
  POST /myorg/myproject/mydb/user
  → 操作 myorg.myproject 项目下，mydb 数据库的 user 表
```

每次请求：
1. 解析 URL 参数，定位 Cluster 连接信息
2. 查询 RuntimeModel 快照（或从缓存获取）
3. 动态构建 GraphQL Schema
4. 执行查询，翻译为 SQL，返回结果

## 相关文件

- `internal/domain/modelruntime/runtimemodel.go` — 运行时模型
- `internal/domain/modelruntime/graphqlschema_manager.go` — 动态 Schema 管理
- `internal/domain/modelruntime/graphql_input.go` — 输入类型定义
- `internal/domain/modelruntime/graphql_where_input_builder.go` — 过滤条件构建
- `internal/domain/modelruntime/model_resolver.go` — 查询解析与 SQL 翻译
- `internal/domain/modelruntime/graphql_field_conditions.go` — 字段过滤条件
