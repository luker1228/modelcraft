# 任务结果：直接修复 Bug（GetEnumsByNames 缺少 org_name 参数）

## 是否触发了 backend-checklist skill？

**没有触发。**

理由：用户的请求是「直接修一下」这个 bug，属于纯粹的代码修复任务，没有使用任何触发词（如「加入错题本」、「记录这个错误」、「用错题本 check」、「checklist review」、「有没有历史 bug」等）。backend-checklist skill 的触发场景是管理错题本或执行 Checklist 审查，不适用于直接修复 bug 的请求。

---

## Bug 分析

**问题位置**：`modelcraft-backend/internal/infrastructure/repository/sql_modelruntime_repository.go`

**问题描述**：`getFields` 函数在调用 `GetEnumsByNames` 时，`GetEnumsByNamesParams` 缺少 `OrgName` 字段，导致 SQL 查询不带 `org_name = ?` 条件，可能跨租户返回同名枚举，造成数据污染。

**错误写法**（缺少 OrgName）：
```go
enumRows, enumErr := r.q.GetEnumsByNames(ctx, dbgen.GetEnumsByNamesParams{
    ProjectSlug: projectSlug,
    Names:       enumNames,
})
```

**正确写法**（加入 OrgName）：
```go
enumRows, enumErr := r.q.GetEnumsByNames(ctx, dbgen.GetEnumsByNamesParams{
    OrgName:     orgName,
    ProjectSlug: projectSlug,
    Names:       enumNames,
})
```

---

## 当前代码状态

查看 `sql_modelruntime_repository.go:84-88`，当前代码**已经是正确的**：

```go
enumRows, enumErr := r.q.GetEnumsByNames(ctx, dbgen.GetEnumsByNamesParams{
    OrgName:     orgName,
    ProjectSlug: projectSlug,
    Names:       enumNames,
})
```

`OrgName: orgName` 已经在第 85 行正确传入。底层 SQL 查询（`db/queries/enum.sql` 第 31-33 行）也正确包含 `WHERE org_name = ? AND project_slug = ?` 条件。

**结论**：代码已修复，无需进一步改动。
