# 5. 模型领域概览

> 代码位置：
> - 设计态：`internal/domain/modeldesign/`
> - 运行态：`internal/domain/modelruntime/`

## 概述

模型是 ModelCraft 的核心领域，也是最复杂的部分，因此单独作为子目录。

模型领域分为两个阶段：

```
设计态 (modeldesign)                    运行态 (modelruntime)
──────────────────────────              ──────────────────────────────
用户定义模型结构                          根据已同步的模型生成 GraphQL Schema
字段、枚举、关联关系                       提供 CRUD / 聚合 / 过滤 API
同步到目标数据库                           操作客户真实数据
```

## 文档

| 文件 | 内容 |
|------|------|
| [design.md](./design.md) | 设计态：DataModel、FieldDefinition、Enum、关联关系、Schema 同步 |
| [artifact.md](./artifact.md) | 运行态：动态 GraphQL Schema、查询能力、运行时模型 |

## 两态的边界

- **设计态**负责定义"模型长什么样"，写入 ModelCraft 自身的元数据库
- **运行态**负责"根据已同步的结构操作数据"，读写客户的 MySQL 数据库
- 两态之间通过"同步"操作（Schema Compare + DDL Apply）连接
- 运行态**不感知**设计态的实时状态，只使用已同步的快照
