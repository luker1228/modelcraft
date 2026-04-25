# Domain 层开发规范

> **触发条件**: 当开发或修改 `internal/domain/**/*.go` 文件时适用本规范，尤其是定义 Repository 接口时。

## 核心原则

Domain 层定义 Repository **接口**（不实现），Infrastructure 层负责实现。接口的方法签名直接决定了安全边界是否正确——参数设计错误会导致租户隔离漏洞或调试困难。

---

## 方法签名规范

### 规则 1：所有方法第一个参数必须是 `ctx context.Context`

`ctx` 承载：请求追踪 ID、超时控制、取消信号、日志字段注入。**没有 ctx 的方法无法被正确追踪和取消。**

```go
// ✅ 正确
type EnumRepository interface {
    Create(ctx context.Context, enum *EnumDefinition) error
    FindByID(ctx context.Context, orgName, id string) (*EnumDefinition, error)
}

// ❌ 错误：缺少 ctx
type EnumRepository interface {
    Create(enum *EnumDefinition) error
    FindByID(id string) (*EnumDefinition, error)
}
```

### 规则 2：`orgName` 是所有方法的必传参数

`orgName` 是**多租户系统的核心安全边界**。每次数据库操作都必须按 `org_name` 过滤，防止跨租户数据泄露。

- 即使是按 ID 查询（`FindByID`、`GetByID`），也必须传入 `orgName`
- 不能依赖 ID 全局唯一来代替 `orgName` 过滤——ID 可能被枚举、猜测或泄露

```go
// ✅ 正确：FindByID 也必须有 orgName
FindByID(ctx context.Context, orgName, id string) (*EnumDefinition, error)

// ❌ 错误：只有 id，无法保证租户隔离
FindByID(id string) (*EnumDefinition, error)
```

**为什么 FindByID 也要 orgName？**

假设攻击者知道另一个租户资源的 ID，如果 `FindByID` 不过滤 `orgName`，他可以直接读取该资源。加入 `orgName` 后，SQL 会变为：

```sql
-- ✅ 安全：即使 ID 泄露也无法跨租户读取
SELECT * FROM enum_definitions WHERE id = ? AND org_name = ?

-- ❌ 不安全：只靠 ID 查询
SELECT * FROM enum_definitions WHERE id = ?
```

### 规则 3：`projectSlug` 按 API Contract 归属决定是否传入

判断 `org scoped` 还是 `org + project scoped`，**优先看 GraphQL Contract 所在目录**（`modelcraft-backend/api/graph/`）：

| Contract 目录 | 领域范围 | 方法参数要求 |
|--------------|---------|-------------|
| `api/graph/org/schema/` | `org scoped` | `ctx + orgName` |
| `api/graph/project/schema/` | `org + project scoped` | `ctx + orgName + projectSlug` |

也就是说：
- 当功能定义在 `api/graph/org/schema/` 下，Repository 查询/删除/检查方法通常不需要 `projectSlug`
- 当功能定义在 `api/graph/project/schema/` 下，Repository 查询/删除/检查方法必须带 `projectSlug`

> 该判定规则比“记资源名”更可靠，新增模块时也不会漏。

---

## 完整示例

### ✅ 正确的 Repository 接口设计

以 `EnumRepository` 为例（所有方法有 `ctx` + `orgName`，项目域方法有 `projectSlug`）：

```go
package modeldesign

import "context"

// EnumRepository 枚举定义仓储接口
type EnumRepository interface {
    // Create 创建枚举定义
    Create(ctx context.Context, enum *EnumDefinition) error

    // Update 更新枚举定义
    Update(ctx context.Context, enum *EnumDefinition) error

    // Delete 删除枚举定义（org + project scoped）
    Delete(ctx context.Context, orgName, projectSlug, name string) error

    // FindByName 根据名称查找枚举定义（org + project scoped）
    FindByName(ctx context.Context, orgName, projectSlug, name string) (*EnumDefinition, error)

    // FindByID 根据ID查找枚举定义
    // orgName 必传，确保跨租户隔离，防止 ID 被猜测或枚举后跨租户读取
    FindByID(ctx context.Context, orgName, id string) (*EnumDefinition, error)

    // List 列出项目下的所有枚举定义（org + project scoped）
    List(ctx context.Context, orgName, projectSlug string) ([]*EnumDefinition, error)

    // IsReferencedByFields 检查枚举是否被字段引用（org + project scoped）
    IsReferencedByFields(ctx context.Context, orgName, projectSlug, name string) (bool, []string, error)

    // ExistsByName 检查项目下指定 name 的枚举是否存在（org + project scoped）
    ExistsByName(ctx context.Context, orgName, projectSlug, name string) (bool, error)
}
```

### ❌ 错误的接口设计（来自 enum_repository.go 的反例）

```go
type EnumRepository interface {
    Create(enum *EnumDefinition) error              // ❌ 缺少 ctx
    Update(enum *EnumDefinition) error              // ❌ 缺少 ctx
    FindByID(id string) (*EnumDefinition, error)   // ❌ 缺少 ctx + orgName，租户隔离漏洞
    Delete(orgName, projectSlug, name string) error // ❌ 缺少 ctx
    // ...
}
```

---

## Create / Update 方法中的 orgName

`Create` 和 `Update` 方法通过实体对象（如 `*EnumDefinition`）传入 `orgName`，因为实体本身就嵌入了 `ProjectScope`（包含 `OrgName` 和 `ProjectSlug`）。

这种情况下不需要在参数列表中额外传入 `orgName`，但 Infrastructure 层实现时**必须从实体中提取 `orgName` 并写入 SQL 条件**：

```go
// ✅ 正确：orgName 通过实体携带
Create(ctx context.Context, enum *EnumDefinition) error

// Infrastructure 实现中：
func (r *SqlEnumRepo) Create(ctx context.Context, enum *EnumDefinition) error {
    params := dbgen.CreateEnumParams{
        OrgName:     enum.OrgName,     // ← 从实体取出，写入 DB
        ProjectSlug: enum.ProjectSlug,
        Name:        enum.Name,
        // ...
    }
    return sqlerr.ExecWithErrorHandling(func() error {
        return r.q.CreateEnum(ctx, params)
    })
}
```

---

## 检查清单

定义 Domain Repository 接口时，确认：

- [ ] **所有方法第一个参数是 `ctx context.Context`**
- [ ] **所有查询/删除/检查方法都有显式 `orgName` 参数**（包括 `FindByID`、`GetByID`）
- [ ] **若 Contract 在 `api/graph/project/schema/` 下，则方法必须包含 `projectSlug` 参数**
- [ ] `Create` / `Update` 方法通过实体对象携带 `orgName`（而非重复传参），但 Infrastructure 实现必须用到它
- [ ] 接口方法签名有注释说明 scope（`org scoped` 或 `org + project scoped`）

---

## 参考索引

| 主题 | 文件 |
|------|------|
| 值对象 ProjectScope（含 OrgName + ProjectSlug） | `internal/domain/project/project_scope.go` |
| EnumDefinition 实体定义 | `internal/domain/modeldesign/enum.go` |
| Infrastructure 层实现规范 | `ai-metadata/backend/development/repo-develop.md` |
| 错误处理规范 | `ai-metadata/backend/development/error-handling.md` |
