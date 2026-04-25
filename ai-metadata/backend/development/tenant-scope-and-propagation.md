# 租户隔离与参数传递规范

> 适用范围：后端所有设计态与运行态请求链路。

## 目标

统一多租户隔离模型，明确两类 scope：
- `org scoped`
- `org + project scoped`

并保证租户参数从入口到 SQL 全链路显式传递，避免中途丢失导致越权访问。

## 先看 API Contract 目录再定 scope

判定标准以 `modelcraft-backend/api/graph/` 为准：

| Contract 目录 | 领域范围 | 典型参数 |
|--------------|---------|----------|
| `api/graph/org/schema/` | `org scoped` | `ctx + orgName` |
| `api/graph/project/schema/` | `org + project scoped` | `ctx + orgName + projectSlug` |

这条规则用于快速判断一个新功能应不应该携带 `projectSlug`。

## 两类隔离边界

| scope | 典型资源 | 必须具备的过滤条件 |
|------|----------|-------------------|
| `org scoped` | Organization / Project / Cluster / Role / UserOrg | `org_name` |
| `org + project scoped` | Model / Field / Enum / ModelGroup / LogicalForeignKey | `org_name` + `project_slug` |

规则：
- 任意 `FindByID/GetByID` 都不能只靠 `id`，至少要加 `org_name`。
- 项目域资源查询必须同时带 `org_name` 和 `project_slug`。

## 参数传递链路（必须显式）

```
Interfaces -> Application -> Domain 接口 -> Infrastructure -> SQL
```

### 1) Interfaces

- 通过 `ctxutils` 提取 `orgName`、`userID`。
- 从路由/输入提取 `projectSlug`（若为项目域）。
- 调用 App 方法时显式传参。

### 2) Application

- 方法签名显式声明 `orgName` / `projectSlug` / `userID`。
- 禁止在 App 层重新从 context 提取租户参数。
- 调用 Domain/Repository 时继续显式下传。

### 3) Domain

- Repository 接口签名直接表达隔离边界。
- 查询/删除/检查方法显式参数必须包含 scope 所需字段。
- 写方法（Create/Update）可由实体携带 scope。

### 4) Infrastructure

- SQL 或 sqlc 参数必须落地完整 scope 过滤。
- 不允许项目域资源出现 `WHERE id = ?` 这类单键查询。

## ProjectScope 嵌入模式

项目域实体和值对象推荐复用 `project.ProjectScope`：

```go
type ModelLocator struct {
    project.ProjectScope        // 嵌入: OrgName + ProjectSlug
    DatabaseName         string
    ModelName            string
}
```

作用：
- 避免重复字段定义。
- 保证 `OrgName` 与 `ProjectSlug` 成对出现。
- 让 Create/Update 这类写接口通过实体携带 scope。

注意：
- 即使实体嵌入了 `ProjectScope`，读/删/查重等方法仍应显式传 `orgName/projectSlug`，不要依赖隐式状态。

## 反例

- `FindByID(ctx, id)`：缺少 `orgName`。
- `List(ctx, orgName)`（项目域资源）：缺少 `projectSlug`。
- Resolver 提取了租户参数，但 App 方法签名未体现。
- Repository SQL 缺 `org_name` 或 `project_slug` 条件。

## 开发检查清单

- [ ] 已按 `api/graph/org|project` 目录判定当前 scope
- [ ] 方法签名包含 scope 所需参数
- [ ] 参数从 Resolver 一路显式传递到 SQL
- [ ] SQL WHERE 条件与 scope 完全一致
- [ ] `FindByID/GetByID` 未遗漏 `orgName`
