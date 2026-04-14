# 🧪 测试策略

> **优先级: 中** - 定义如何验证实现是否符合设计，确保代码质量。

## 概述

测试策略主要关注集成测试，验证各组件协作是否正确，确保系统行为符合业务预期。

## 📚 文档列表

| 文档 | 说明 |
|------|------|
| [debugging-workflow.md](./debugging-workflow.md) | **开发调试流程** - 日常开发必读 ⭐ |
| [bdd-testing-guidelines.md](./bdd-testing-guidelines.md) | BDD 验收测试注意要点（默认不耦合注册） |
| [integration-testing.md](./integration-testing.md) | 集成测试指南 |
| [test-strategy.md](./test-strategy.md) | 测试策略总览 |
| [test-data.md](./test-data.md) | 测试数据管理 |
| [coverage-requirements.md](./coverage-requirements.md) | 覆盖率要求 |

## 🧬 测试金字塔

```
        /\
       /  \      E2E 测试 (少量)
      /    \     - 端到端场景验证
     /──────\
    /        \   集成测试 (重点) ⭐
   /          \  - API 测试
  /            \ - 数据库集成
 /──────────────\
/                \ 单元测试 (大量)
                   - Domain 层 95%+ 覆盖率
                   - 纯函数、业务逻辑
```

## 🎯 测试重点

### 1. 单元测试 (Unit Tests)

- **覆盖范围**: Domain 层
- **覆盖率要求**: 95%+
- **特点**: 快速、隔离、无外部依赖

```bash
# 运行单元测试
task test-unit

# 检查覆盖率
task test-coverage
```

### 2. 集成测试 (Integration Tests)

- **覆盖范围**: API 端点、数据库操作
- **特点**: 验证组件协作、使用真实依赖

```bash
# 运行集成测试
task test-integration
```

### 3. E2E 测试 (End-to-End Tests)

- **覆盖范围**: 关键业务流程
- **特点**: 模拟真实用户场景

## 📋 测试命名规范

```go
// 单元测试
func TestEntityName_MethodName_Scenario(t *testing.T)

// 示例
func TestUser_Create_WithValidInput(t *testing.T)
func TestUser_Create_WithEmptyName_ReturnsError(t *testing.T)
```

## ✅ 测试检查清单

- [ ] Domain 层覆盖率 ≥ 95%
- [ ] 所有公开 API 有集成测试
- [ ] 关键业务流程有 E2E 测试
- [ ] 测试可独立运行，无顺序依赖
- [ ] 测试数据自动清理

## 📖 阅读顺序

### 对于新手开发者

1. **先阅读** `debugging-workflow.md` 了解日常调试流程 (最实用)
2. 再阅读 `test-strategy.md` 了解整体测试策略
3. 再阅读 `integration-testing.md` 了解集成测试详情
4. 按需阅读其他专题文档

### 对于测试编写

1. 先阅读 `test-strategy.md` 了解整体策略
2. 再阅读 `integration-testing.md` 了解集成测试详情
3. 按需阅读其他专题文档

## ⚠️ 前置要求

阅读本目录前，请确保已理解：
- [1-design](../1-design/) - 设计理念
- [2-development](../2-development/) - 开发规范

## 🔗 相关文档

- 测试工具请参考 [5-tools](../5-tools/)
- 部署验证请参考 [4-deployment](../4-deployment/)
