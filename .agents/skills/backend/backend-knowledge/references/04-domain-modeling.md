# 04 - Domain 开发指引

## 职责

`domain` 层沉淀核心业务概念：

- 实体（Entity）
- 值对象（Value Object）
- 聚合（Aggregate）
- 领域服务（Domain Service）
- 业务不变量（Invariant）

## 设计原则

1. **表达业务语言**：命名贴近业务语义
2. **保护不变量**：创建和变更必须满足约束
3. **持久化无关**：不关心 SQL、缓存、传输协议

## 开发步骤

1. 抽取核心实体和值对象
2. 定义状态变化方法（而非随意字段修改）
3. 在方法内强制校验业务规则
4. 返回明确的领域错误

## 常见反模式

- 将 domain 变成纯 DTO
- 在 domain 引入数据库/HTTP 依赖
- 把所有规则塞到 interface/app 而不是 domain
