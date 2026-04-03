# 🔒 ModelCraft SafeQuerier 规划文档

## 📚 文档导航

本项目包含以下规划文档，用于指导 **SafeQuerier Wrapper** 的实现：

### 1. 📋 **SAFE_QUERIER_SUMMARY.md** (6.5 KB) - ⭐ 从这里开始
   - **内容**: 执行摘要，核心发现，快速总览
   - **适用**: 想快速了解项目整体情况
   - **关键部分**: 
     - 核心发现
     - 方法分类统计
     - 实现方案对比表
     - 预期收益

### 2. 📊 **SAFE_QUERIER_PLANNING.md** (9.4 KB) - 详细规划
   - **内容**: 完整的项目分析、现状评估、下一步行动
   - **适用**: 需要深入理解项目结构和实现细节
   - **关键部分**:
     - 1️⃣ Querier 接口概览 (125 个方法分类统计)
     - 2️⃣ SQL 错误处理基础设施
     - 3️⃣ Repository 实现文件列表 (41 个文件)
     - 4️⃣ Repository 实现模式分析
     - 5️⃣ Justfile 中的 sqlc 命令
     - 6️⃣ 已有工具/配置检查
     - 7️⃣ Scripts 目录内容
     - 8️⃣ SafeQuerier 实现规划
     - 9️⃣ 下一步行动清单

### 3. 📖 **QUERIER_METHODS_REFERENCE.txt** (11 KB) - 方法参考
   - **内容**: 所有 125 个 Querier 方法的完整签名列表
   - **适用**: 代码生成时参考，查询特定方法签名
   - **分类**:
     - CREATE 操作 (15 个)
     - GET 操作 (25 个)
     - LIST 操作 (15 个)
     - UPDATE 操作 (13 个)
     - DELETE 操作 (15 个)
     - COUNT 操作 (4 个)
     - EXISTS 操作 (5 个)
     - FIND 操作 (7 个)
     - 其他操作 (6 个)

### 4. 🏗️ **ARCHITECTURE_DIAGRAM.txt** (14 KB) - 架构设计图
   - **内容**: ASCII 架构图、方法流转、代码改进示例
   - **适用**: 理解系统架构和数据流
   - **包含**:
     - 完整的分层架构图
     - 三个详细的方法流转场景
     - 文件结构对比 (现状 vs 目标)
     - 代码改进前后对比

---

## 🎯 快速开始路线

### 👤 角色1: 项目经理/决策者
阅读顺序:
1. 📋 `SAFE_QUERIER_SUMMARY.md` (5-10 分钟)
2. 🏗️ `ARCHITECTURE_DIAGRAM.txt` - "代码改进示例" 部分 (5 分钟)
3. 📊 `SAFE_QUERIER_PLANNING.md` - "实现方案对比" 部分 (3 分钟)

**关键结论**: 推荐使用 Python 脚本或 Go 代码生成方式

### 👨‍💻 角色2: 实现工程师
阅读顺序:
1. 📋 `SAFE_QUERIER_SUMMARY.md` (10 分钟)
2. 📊 `SAFE_QUERIER_PLANNING.md` - 完整阅读 (20-30 分钟)
3. 📖 `QUERIER_METHODS_REFERENCE.txt` - 参考查阅 (随需)
4. 🏗️ `ARCHITECTURE_DIAGRAM.txt` - 完整阅读 (15 分钟)

**关键产出**: 
- 理解 SafeQuerier 需要包装的 125 个方法
- 了解现有的错误处理基础设施
- 决定使用哪种生成方式

### 🧪 角色3: 测试/QA
阅读顺序:
1. 📋 `SAFE_QUERIER_SUMMARY.md` - "快速开始" 部分 (5 分钟)
2. 🏗️ `ARCHITECTURE_DIAGRAM.txt` - "方法流转示例" 部分 (10 分钟)
3. 📊 `SAFE_QUERIER_PLANNING.md` - "实现清单" 部分 (5 分钟)

**关键产出**: 
- 理解测试覆盖点
- 了解异常场景处理
- 准备测试用例

---

## 📊 核心统计数据

```
总方法数:        125 个
├── 无参数:      ~10 个 (如 ListProjects)
├── 单参数:      ~30 个 (如 GetModelByID(id))
├── 多参数:      ~85 个 (使用 Params 结构)

返回值分布:
├── error:       ~85 个 (68%)
├── (T, error):  ~30 个 (24%)
├── ([]T, error):~20 个 (16%)  # 重叠计数
├── (int64, error): ~10 个 (8%)
├── (sql.Result, error): ~3 个 (2%)

Repository 文件:
├── 核心实现:    11 个
├── 转换函数:    6 个
├── 模型定义:    6 个
├── 工具/辅助:   20 个
总计:           41 个文件
```

---

## 🔧 关键技术点

### 错误处理链条
```
database/sql error
    ↓
WrapSQLError(err)
    ↓
AnalyzeSQLError(err)
    ↓
classifyError(err) + errorPatterns
    ↓
shared.RepositoryError
```

### 支持的错误分类
- ✅ DuplicatedKey (MySQL 1062, 唯一约束冲突)
- ✅ Constraint (MySQL 1451/1452, 外键约束)
- ✅ Connection (连接失败/重置)
- ✅ Timeout (超时)
- ✅ Transaction (死锁/锁超时)
- ✅ Permission (权限不足)
- ✅ NotFound (sql.ErrNoRows)
- ✅ Unknown (未分类)

### 现有工具
- ✅ WrapSQLError() - 错误包装函数
- ✅ ExecWithErrorHandling() - 执行包装
- ✅ QueryWithSQLErrorHandling() - 查询包装
- ✅ IsNotFoundError() - NotFound 检查

---

## 📁 项目文件位置

| 类型 | 文件 | 说明 |
|------|------|------|
| 错误处理 | `internal/infrastructure/repository/sql_error_analyzer.go` | WrapSQLError 实现 |
| Querier | `internal/infrastructure/dbgen/querier.go` | 125 个方法 (sqlc 生成) |
| Repository 示例 | `internal/infrastructure/repository/sql_modeldesign_repository.go` | 使用 Querier 的实现 |
| Build 配置 | `justfile` (259-330 行) | sqlc 相关命令 |
| Scripts | `scripts/` | 10 个辅助脚本 |

---

## ✨ 预期收益

### 代码质量
- ✅ 消除 125 处重复的 `WrapSQLError()` 调用
- ✅ 统一的错误处理策略
- ✅ 减少人为错误

### 开发效率
- ✅ Repository 代码更简洁
- ✅ 新方法自动包装
- ✅ 维护成本降低

### 可维护性
- ✅ 错误处理逻辑集中化
- ✅ 易于扩展新的错误分类
- ✅ 便于追踪错误来源

---

## 🚀 下一步

### 立即行动
1. 阅读 `SAFE_QUERIER_SUMMARY.md` 了解全貌
2. 查看 `ARCHITECTURE_DIAGRAM.txt` 理解架构
3. 选择实现方案 (gowrap / Python 脚本 / Go 代码生成)

### 准备工作
1. 设计 SafeQuerier 接口
2. 编写代码生成工具或脚本
3. 创建单元测试框架

### 实现阶段
1. 生成 SafeQuerier 接口和实现
2. 集成到 Justfile
3. 更新现有 Repository 类
4. 运行完整的测试套件

---

## 📞 参考信息

**项目**: ModelCraft Backend
**生成时间**: 2026-04-03
**规划版本**: 1.0

**主要发现**:
- ✅ 现有错误处理基础设施完善
- ✅ 125 个 Querier 方法需要包装
- ❌ 暂无 gowrap 配置
- 💡 建议: Python 脚本或 Go 代码生成

---

## 📖 如何使用本文档

### 对于新加入的开发者
1. 从 `SAFE_QUERIER_SUMMARY.md` 开始
2. 浏览 `ARCHITECTURE_DIAGRAM.txt` 理解架构
3. 参考 `QUERIER_METHODS_REFERENCE.txt` 了解方法列表
4. 阅读 `SAFE_QUERIER_PLANNING.md` 获取完整细节

### 对于代码生成脚本编写者
1. 重点阅读 `QUERIER_METHODS_REFERENCE.txt`
2. 参考 `ARCHITECTURE_DIAGRAM.txt` 的代码示例
3. 查看 `sql_error_analyzer.go` 了解 WrapSQLError 实现
4. 查看 `sql_modeldesign_repository.go` 了解使用模式

### 对于测试工程师
1. 阅读 `ARCHITECTURE_DIAGRAM.txt` 的"方法流转示例"
2. 理解三个核心场景 (成功, NOT FOUND, CONSTRAINT)
3. 参考 `sql_error_analyzer.go` 中的错误分类
4. 设计覆盖各错误类型的测试用例

---

**💾 备注**: 所有文档都保存在项目根目录，可通过 `grep` 或编辑器快速查找关键信息。
