---
name: backend-knowledge
description: >
  ModelCraft Go 后端分层开发知识库与最佳实践指引。覆盖业务逻辑开发的五层架构
  (pkg/interface/app/domain/repo) + 公共能力 (日志、错误处理)。
  使用场景：(1) 开发新的 pkg 包组件, (2) 设计业务接口 (Interface), (3) 实现应用层
  (App) 业务流程, (4) 定义领域模型与聚合 (Domain), (5) 实现数据访问层 (Repo),
  (6) 集成日志与错误处理。遇到不确定该在哪层实现、如何组织代码、如何正确使用
  日志/错误时，触发此技能。
---

# ModelCraft Go 后端分层开发指引

## 概览

ModelCraft 采用**分层架构**组织业务逻辑，从底向上分为 5 层：

| 层级 | 名称 | 职责 | 依赖方向 |
|------|------|------|--------|
| **1** | **Repo** | 数据持久化、SQL 执行 | ↓ 仅依赖 DB、pkg |
| **2** | **Domain** | 领域模型、聚合、业务规则 | ↓ 仅依赖 pkg |
| **3** | **App** | 应用流程、业务编排 | ↓ 依赖 domain、repo、pkg |
| **4** | **Interface** | API 契约、DTO、验证 | ↓ 依赖 app、pkg |
| **5** | **Pkg** | 公共工具、配置、中间件 | 无向上依赖 |

**公共能力**（横贯所有层）：日志、错误处理、追踪。

## 快速导航

选择你正在处理的场景：

- **我要开发 Pkg** → [详见 references/01-pkg-development.md](references/01-pkg-development.md)
- **我要设计 Interface** → [详见 references/02-interface-design.md](references/02-interface-design.md)
- **我要实现 App** → [详见 references/03-app-implementation.md](references/03-app-implementation.md)
- **我要定义 Domain** → [详见 references/04-domain-modeling.md](references/04-domain-modeling.md)
- **我要实现 Repo** → [详见 references/05-repository-pattern.md](references/05-repository-pattern.md)
- **日志与错误处理** → [详见 references/06-logging-error-handling.md](references/06-logging-error-handling.md)

## 核心原则

### 分层的好处

1. **清晰的职责边界** - 每层知道自己做什么，不越界
2. **易于测试** - 各层独立单元测试，通过 mock 进行集成测试
3. **易于维护** - 改动影响范围明确，重构成本低
4. **易于扩展** - 新增功能按分层添加，无需改动已有层

### 分层的约束

```
Interface (GraphQL/HTTP)
     ↓
   App (业务流程)
     ↓
  Domain (领域模型)  ←  Repo (数据访问)
     ↓
   Pkg (公共能力)
```

**记住：上层依赖下层，下层不能依赖上层。**

## 何时使用本技能

✅ **应该触发此技能的场景：**

1. **不确定某段代码该写在哪一层**
   - "这个验证逻辑应该在 Domain 还是 Interface？"
   - "API 返回值处理应该在 App 还是 Interface？"

2. **新建一个 pkg，不知道如何组织**
   - "我需要实现缓存能力，应该怎么设计 pkg 结构？"
   - "日志系统应该暴露什么接口？"

3. **不确定如何正确使用日志和错误处理**
   - "在 Repo 层应该记什么日志？"
   - "业务错误应该如何定义和传递？"
   - "何时返回自定义错误，何时返回 nil？"

4. **设计数据库查询或 SQL**
   - "这个查询应该放在 Repo 的哪个方法？"
   - "如何组织复杂的多表 JOIN？"

5. **设计 API 接口或 GraphQL Schema**
   - "Input/Output 结构应该怎么定义？"
   - "哪些字段需要验证？"

❌ **不需要此技能的场景：**

- 调试语法错误或编译错误
- 学习 Go 基础语法
- 使用某个第三方库的具体方法

## 扩展指南

新增功能指南时：

1. **在 SKILL.md 中添加指向新文件的链接**
   ```markdown
   - **功能名称** → [详见 references/07-new-feature.md](references/07-new-feature.md)
   ```

2. **在 references/ 下创建 `NN-feature-name.md`**
   - 编号格式保持一致（如 `07-new-feature.md`）
   - 包含具体例子和代码片段

3. **更新本文件的导航表**（保持统一编号）

## 相关技能

- [backend-coding-standards](../backend-coding-standards/SKILL.md) - 代码风格与规范
- [backend-patterns](../backend-patterns/SKILL.md) - 架构模式与设计原则
- [backend-debug](../../backend-debug/SKILL.md) - 问题排查与日志分析
