# 🎯 ModelCraft SafeQuerier 实现规划 - 执行摘要

## 📌 核心发现

### ✅ 项目现状
```
ModelCraft Backend 项目
├── sqlc v1.30.0 已集成
│   └── 125 个 Querier 方法自动生成
├── 完善的错误处理层
│   ├── WrapSQLError() 函数
│   ├── AnalyzeSQLError() 函数
│   └── ErrorPattern 分类系统
└── 11 个核心 Repository 实现
    └── 使用 r.q (Querier接口) 依赖注入
```

### ❌ 缺失部分
```
SafeQuerier Wrapper
├── ❌ 无 gowrap 配置
├── ❌ 无自动生成脚本
└── ❌ 需要手工或脚本生成 (125 个方法)
```

---

## 📊 Querier 方法分类统计

```
总计: 125 个方法

按操作类型:
  ├── Create ........... 15 个  ╔════════════════════════════╗
  ├── Get .............. 25 个  ║ Get 和 Update 占多数      ║
  ├── List ............. 15 个  ╚════════════════════════════╝
  ├── Update ........... 13 个
  ├── Delete ........... 15 个
  ├── Count ............  4 个
  ├── Exists ...........  5 个
  ├── Find ............. 7 个
  └── Other (Insert, Revoke) .. 6 个

按返回类型:
  ├── error ..................... ~85 个 (68%)  ← 主要
  ├── (T, error) ................ ~30 个 (24%)  ← 单个查询
  ├── ([]T, error) .............. ~20 个 (16%)  ← 列表查询
  ├── (int64, error) ............ ~10 个 ( 8%)  ← 计数/存在性
  └── (sql.Result, error) .......  ~3 个 ( 2%)  ← 批量操作
```

---

## 🏗️ 实现架构

### 当前 Repository 接入点
```go
type SqlModelDesignRepository struct {
    q dbgen.Querier  // ← Querier 接口
}

func (r *SqlModelDesignRepository) GetByID(ctx context.Context, id string) (..., error) {
    row, err := r.q.GetModelByID(ctx, id)  // ← 直接调用 Querier 方法
    if err != nil {
        return nil, WrapSQLError(err)      // ← 使用 WrapSQLError 包装错误
    }
    return ModelToDomain(row), nil
}
```

### 目标 SafeQuerier 架构
```go
type SafeQuerier interface {
    // 包装所有 125 个 Querier 方法，自动应用 WrapSQLError
    GetModelByID(ctx context.Context, id string) (Model, error)
    // ...
}

type safeQuerierImpl struct {
    q dbgen.Querier
}

func (sq *safeQuerierImpl) GetModelByID(ctx context.Context, id string) (Model, error) {
    result, err := sq.q.GetModelByID(ctx, id)
    return result, WrapSQLError(err)  // ← 统一错误处理
}
```

---

## 🔧 实现方案对比

| 方案 | 优点 | 缺点 | 工作量 |
|------|------|------|--------|
| **gowrap** | 自动化、易维护 | 需学习、配置复杂 | ⭐⭐ |
| **Python 脚本** | 灵活、快速 | 依赖Python环境 | ⭐⭐⭐ |
| **Go 代码生成** | 集成友好、标准 | 代码较多 | ⭐⭐⭐⭐ |
| **手工编写** | 无依赖 | 易出错、维护难 | ⭐⭐⭐⭐⭐ |

**推荐方案**: **Python 脚本** (快速、灵活)或 **Go 代码生成** (长期维护)

---

## 📁 项目结构现状

```
internal/infrastructure/
├── dbgen/
│   └── querier.go          # sqlc 生成 (125 个方法，DO NOT EDIT)
│   ├── safe_querier.go     # 🆕 SafeQuerier 接口
│   └── safe_querier_impl.go # 🆕 wrapper 实现
├── repository/
│   ├── sql_error_analyzer.go    # ✅ 错误处理 (WrapSQLError)
│   ├── sql_modeldesign_repository.go
│   ├── sql_org_repository.go
│   ├── sql_enum_repository.go
│   ├── ... (8 个其他 repo 文件)
│   └── error_helper.go         # ✅ 错误辅助
└── ...
```

---

## 🎯 错误处理流转

```
Querier 方法执行
    ↓
    ├─→ 成功: 返回 (result, nil)
    │   └─→ SafeQuerier: return (result, nil)
    │
    └─→ 失败: 返回 (zero, error)
        └─→ SafeQuerier: return (zero, WrapSQLError(error))
            └─→ AnalyzeSQLError(error)
                └─→ classifyError(error)
                    └─→ 根据错误消息匹配 errorPatterns
                        └─→ shared.RepositoryError
```

---

## 📋 实现清单

### 阶段 1: 准备 (1-2 天)
- [ ] 选择代码生成方案
- [ ] 编写生成脚本/工具
- [ ] 创建单元测试模板

### 阶段 2: 生成 (1 天)
- [ ] 生成 SafeQuerier 接口
- [ ] 生成 SafeQuerier 实现
- [ ] 运行初始测试

### 阶段 3: 集成 (2-3 天)
- [ ] 集成到 Justfile
- [ ] 更新现有 Repository 类
- [ ] 验证向后兼容性

### 阶段 4: 测试 (2-3 天)
- [ ] 单元测试覆盖
- [ ] 集成测试
- [ ] 性能基准测试

### 阶段 5: 发布 (1 天)
- [ ] 文档更新
- [ ] 迁移指南
- [ ] 代码审查

---

## 💡 关键决策点

### 1️⃣ 错误处理策略
- ✅ 使用现有的 `WrapSQLError()` 函数
- ✅ 支持链式错误（error.Cause）
- ✅ 保持 `shared.RepositoryError` 类型

### 2️⃣ 接口位置
- 建议在 `internal/infrastructure/dbgen/safe_querier.go` 中定义
- 配合 `querier.go` 并存于 dbgen 包

### 3️⃣ 实现方式
- 建议先使用脚本生成，后期可切换 gowrap
- 保持生成结果的可读性

### 4️⃣ 向后兼容性
- Repository 可逐步迁移到 SafeQuerier
- 无需一次性替换所有 Repository

---

## 📚 参考文件位置

| 文档 | 路径 |
|------|------|
| 详细规划 | `SAFE_QUERIER_PLANNING.md` |
| 方法参考 | `QUERIER_METHODS_REFERENCE.txt` |
| 错误处理 | `internal/infrastructure/repository/sql_error_analyzer.go` |
| 示例 Repo | `internal/infrastructure/repository/sql_modeldesign_repository.go` |
| Build 配置 | `justfile` (第259-330行) |

---

## 🚀 快速开始

### 第1步: 生成 SafeQuerier
```bash
# (假设选择 Python 脚本方案)
python3 scripts/generate_safe_querier.py
```

### 第2步: 集成到构建流程
```bash
# 在 justfile 中添加
generate-safe-querier:
    python3 scripts/generate_safe_querier.py
```

### 第3步: 使用 SafeQuerier
```go
// 创建 SafeQuerier 实例
sq := dbgen.NewSafeQuerier(queries)

// 使用（错误自动包装）
result, err := sq.GetModelByID(ctx, id)
// err 已通过 WrapSQLError 处理
```

---

## ✨ 预期收益

✅ **自动错误处理** - 无需重复 WrapSQLError 调用
✅ **代码简化** - Repository 实现更清晰
✅ **一致性保证** - 所有 DB 操作统一处理
✅ **易于维护** - 新方法自动包装
✅ **类型安全** - 编译期检查

---

**生成时间**: 2026-04-03
**项目**: ModelCraft Backend
**规划者**: Claude Code
