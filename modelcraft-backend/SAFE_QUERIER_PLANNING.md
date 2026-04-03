# ModelCraft SafeQuerier Wrapper 实现规划文档

## 📊 项目现状分析

### 1. Querier 接口概览

**📍 位置**: `internal/infrastructure/dbgen/querier.go`

- **总方法数**: 125 个方法
- **代码生成工具**: sqlc v1.30.0（自动生成）
- **状态**: `DO NOT EDIT` - 由 sqlc 自动管理

#### 方法分类统计

| 类别 | 方法模式 | 示例数 |
|-----|--------|------|
| **Create** | 创建记录 | CreateModel, CreateUser, CreateOrganization, CreateProject... (~15个) |
| **Get** | 按ID/条件获取单条 | GetModelByID, GetUserByID, GetRoleByID... (~25个) |
| **List** | 列表查询 | ListModels, ListRoles, ListProjects... (~15个) |
| **Update** | 更新记录 | UpdateModel, UpdateUser, UpdateRole... (~13个) |
| **Delete** | 删除记录 | DeleteModel, DeleteUser, DeleteRole... (~15个) |
| **Find** | 条件查询(返回[]行) | FindLogicalForeignKeysByModelID, FindFieldsByBelongsToFKID... (~5个) |
| **Count** | 计数 | CountModels, CountFieldsByModelID, CountMembershipsByUser... (~4个) |
| **Exists** | 存在性检查 | ExistsProjectBySlug, ExistsEnumByName... (~5个) |
| **Other** | 其他操作 | Revoke*, Archive, Insert* | (~22个) |

#### 返回值类型分布

```
error                     (~85个方法)
(T, error)               (~30个方法)  // 单个对象查询
([]T, error)             (~20个方法)  // 列表查询
(int64, error)           (~10个方法)  // 计数
(sql.Result, error)      (~3个方法)   // 批量操作
nil                      (少见)
```

---

### 2. SQL 错误处理基础设施

**📍 位置**: `internal/infrastructure/repository/sql_error_analyzer.go`

#### WrapSQLError 函数签名
```go
func WrapSQLError(err error) error {
    if err == nil {
        return nil
    }
    return AnalyzeSQLError(err)
}
```

#### 相关辅助函数
```go
func AnalyzeSQLError(err error) error          // 分析并包装错误
func IsNotFoundError(err error) bool           // 检查 NotFound 错误
func classifyError(err error) shared.RepositoryErrorType  // 错误分类
```

#### 错误分类支持
- **MySQL 错误码**: 1062 (重复键), 1451/1452 (FK约束), 1064 (SQL语法), 1146 (表不存在)
- **通用模式**: Duplicate, Unique, Foreign Key, Connection, Timeout, Permission, Access Denied, Deadlock
- **返回类型**: `shared.RepositoryErrorType` (领域层定义的错误类型)

---

### 3. Repository 实现文件列表

**📍 位置**: `internal/infrastructure/repository/`

#### 核心 Repository 实现 (11个)
| 文件名 | 用途 |
|--------|------|
| `sql_modeldesign_repository.go` | 模型设计CRUD (23.7KB) |
| `sql_org_repository.go` | 组织管理 (15.5KB) |
| `sql_casbin_repository.go` | RBAC权限管理 (13.0KB) |
| `sql_database_cluster_repository.go` | 数据库集群配置 (9.7KB) |
| `sql_enum_repository.go` | 枚举定义管理 (12.7KB) |
| `sql_logical_foreign_key_repository.go` | 逻辑外键管理 (7.6KB) |
| `sql_model_group_repository.go` | 模型分组 (5.5KB) |
| `sql_api_key_repository.go` | API密钥管理 (4.2KB) |
| `sql_modelruntime_repository.go` | 模型运行时 (3.8KB) |
| `sql_refresh_token_repository.go` | 刷新令牌 (2.5KB) |
| `sql_security_audit_log_repository.go` | 安全审计日志 (1.2KB) |
| `project_repository.go` | 项目管理 (6.3KB) |

#### 辅助模型/工具文件 (20个)
- **转换函数**:
  - `modeldesign_convert_test.go` (11.4KB - 包含转换逻辑)
  - `org_convert_test.go` (4.4KB)
  - `enum_convert_test.go` (5.9KB)
  - `cluster_convert_test.go` (5.6KB)
  - `model_group_convert_test.go` (1.9KB)
  - `modelruntime_convert_test.go` (1.3KB)

- **模型定义**:
  - `membership_model.go` (8.8KB)
  - `role_model.go` (5.0KB)
  - `organization_model.go` (4.0KB)
  - `model_group_model.go` (1.6KB)
  - `user_model.go` (3.5KB)
  - `field_enum_association_model.go` (1.8KB)

- **其他工具**:
  - `error_helper.go` - 错误辅助函数
  - `sql_connection.go` - 连接工厂
  - `connection_factory.go` - 连接工厂
  - `gorm_connection.go` - GORM连接
  - `gorm_logger.go` - GORM日志记录器
  - `sqlc_logger.go` - sqlc日志记录器 (5.0KB)
  - `cluster_connection_manager.go` - 集群连接管理 (13.1KB)
  - `tx_manager.go` - 事务管理器

---

### 4. Repository 实现模式分析

**📍 示例**: `sql_modeldesign_repository.go` (前100行)

#### 类型定义
```go
type SqlModelDesignRepository struct {
    q dbgen.Querier  // ← Querier 接口注入
}
```

#### 构造函数模式
```go
func NewSqlModelDesignRepository(q dbgen.Querier) modeldesign.ModelRepository {
    return &SqlModelDesignRepository{q: q}
}
```

#### 方法调用模式
```go
// 方式1: 直接调用 + ExecWithErrorHandling 包装
func (r *SqlModelDesignRepository) Save(...) error {
    return r.q.CreateModel(ctx, params)  // 有时直接返回
}

// 方式2: 使用 ExecWithErrorHandling
func (r *SqlModelDesignRepository) Save(...) error {
    if err := ExecWithErrorHandling(func() error {
        return r.q.CreateModel(ctx, ModelToCreateParams(model, orgName))
    }); err != nil {
        return err
    }
    return r.AddFields(ctx, orgName, model.Fields)
}

// 方式3: 查询 + 转换 + 错误处理
row, err := r.q.GetModelByID(ctx, id)
if err != nil {
    return nil, WrapSQLError(err)
}
```

#### 错误处理函数
```go
func ExecWithErrorHandling(op func() error) error {
    return WrapSQLError(op())
}

func QueryWithSQLErrorHandling(op func() error) error {
    return WrapSQLError(op())
}
```

#### 转换函数示例（第16行开始）
```go
func ModelToDomain(row dbgen.Model) *modeldesign.DataModel { ... }
func ModelToCreateParams(m *modeldesign.DataModel, orgName string) dbgen.CreateModelParams { ... }
func ModelToUpdateParams(m *modeldesign.DataModel) dbgen.UpdateModelParams { ... }
```

---

### 5. Justfile 中的 sqlc 相关命令

**📍 位置**: `justfile` (第259-330行)

#### install-sqlc (私有recipe)
```bash
install-sqlc:
    #!/usr/bin/env bash
    if ! command -v sqlc > /dev/null 2>&1; then
        echo "sqlc 未安装，正在安装..."
        go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
    fi
    echo "sqlc 工具已就绪"
```

#### generate-sqlc (公开recipe)
```bash
generate-sqlc: install-sqlc
    echo "📊 生成 sqlc 代码..."
    sqlc generate
    echo "✅ sqlc 代码生成完成"
```

#### clean-sqlc
```bash
clean-sqlc:
    echo "清理生成的 sqlc 代码..."
    rm -f internal/infrastructure/dbgen/*.go
    echo "✅ sqlc 代码清理完成"
```

**配置文件**: 应位于 `sqlc.yaml`（执行 `sqlc generate` 时读取）

---

### 6. 已有工具/配置检查

#### ✅ 已有
- ✓ `sql_error_analyzer.go` - 完善的错误处理层
- ✓ `ExecWithErrorHandling()` - 执行包装器
- ✓ `QueryWithSQLErrorHandling()` - 查询包装器
- ✓ 统一的 `WrapSQLError()` 错误处理函数
- ✓ 转换函数集合（ModelToDomain, ModelToCreateParams等）
- ✓ 领域错误类型（shared.RepositoryErrorType）

#### ❌ 不存在
- ✗ `.gowrap` 配置文件 - **没有任何 gowrap 相关文件**
- ✗ `gowrap` 代码生成 - **未集成 gowrap 工具**
- ✗ 现有的 SafeQuerier wrapper - **需要从零开始实现**

---

### 7. Scripts 目录内容

**📍 位置**: `scripts/` (10个shell脚本)

| 文件 | 用途 |
|------|------|
| `db-env.sh` | 数据库环境配置 (781B) |
| `setup-casdoor.sh` | Casdoor认证配置 (2.7KB) |
| `verify-casdoor-integration.sh` | 验证认证集成 |
| `add-missing-tests.sh` | 补充测试 (2.2KB) |
| `ai-generate-test.sh` | AI生成测试 (6.2KB) |
| `analyze-uncovered.sh` | 覆盖率分析 (2.1KB) |
| `auto-fix-coverage.sh` | 自动修复覆盖率 (5.0KB) |
| `generate-meaningful-tests.sh` | 生成测试模板 (2.9KB) |
| `generate-tests-with-agent.sh` | Agent生成测试 (8.2KB) |
| `generate-test-template.sh` | 测试模板生成 (2.0KB) |

**用途**: 主要用于测试覆盖率和自动化测试生成（非编码生成工具）

---

## 🎯 SafeQuerier Wrapper 实现规划

### 核心需求

基于上述分析，SafeQuerier wrapper 需要：

1. **包装所有125个Querier方法**
2. **统一错误处理** - 使用现有的 `WrapSQLError()`
3. **支持多种返回类型**:
   - `error`
   - `(T, error)`
   - `([]T, error)`
   - `(int64, error)`
   - `(sql.Result, error)`

4. **生成方式选择**:
   - **选项A**: gowrap 自动生成（需要 `.gowrap` 配置）
   - **选项B**: 手工编写模板代码
   - **选项C**: 代码生成脚本（Python/Go）

### 预期输出物

```
internal/infrastructure/dbgen/
├── querier.go                    # ← 原始接口（sqlc生成）
├── safe_querier.go               # ← SafeQuerier wrapper（新增）
├── safe_querier_impl.go          # ← 包装实现（新增）
└── safe_querier_test.go          # ← 测试覆盖（新增）
```

### 关键集成点

1. **现有错误处理集成**:
   ```go
   return WrapSQLError(q.GetModelByID(ctx, id))
   ```

2. **Repository 层集成**:
   ```go
   // 更新现有代码
   return r.q.GetModelByID(ctx, id)  // 自动包装
   ```

3. **Justfile 集成**:
   ```bash
   generate-safe-querier: install-gowrap
       gowrap gen ...
   ```

---

## 📋 下一步行动清单

- [ ] 决策代码生成方式（gowrap vs 脚本 vs 手工）
- [ ] 确定 SafeQuerier 接口签名
- [ ] 实现错误处理层集成
- [ ] 生成 safe_querier.go
- [ ] 编写单元测试
- [ ] 集成到 Justfile
- [ ] 更新现有 Repository 类使用新的 SafeQuerier
- [ ] 验证向后兼容性

