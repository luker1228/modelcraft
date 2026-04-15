🗂️  Checklist Review — SqlModelRuntimeRepository.getFields (inline snippet)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

共 1 条规则

❌ 命中 1 条：
  [BM-20260415-0001] 凡是查询带 `project_slug` 的 SQL，必须同时带 `org_name`。
  → 位置：inline snippet（getFields 函数体）
  → 问题：调用 `GetEnumsByNames` 时，`dbgen.GetEnumsByNamesParams` 只传了 `ProjectSlug`，
           缺少 `OrgName` 字段。函数签名中已有 `orgName` 参数但未使用。
           `project_slug` 在不同 org 之间可能重名，缺少 `org_name` 条件会导致
           跨租户数据污染，拿到其他 org 的同名枚举定义。
  → 参考案例：BM-20260415-0001

✅ 通过 0 条（无其他规则）

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## 修改建议

```go
// ❌ 当前代码（缺少 OrgName）
enumRows, _ := r.q.GetEnumsByNames(ctx, dbgen.GetEnumsByNamesParams{
    ProjectSlug: projectSlug,
    Names:       enumNames,
})

// ✅ 修复后（必须同时传 OrgName）
enumRows, _ := r.q.GetEnumsByNames(ctx, dbgen.GetEnumsByNamesParams{
    OrgName:     orgName,
    ProjectSlug: projectSlug,
    Names:       enumNames,
})
```

如果 `dbgen.GetEnumsByNamesParams` 尚未包含 `OrgName` 字段，需同步修改：
1. SQL 查询加 `org_name = ?` 过滤条件
2. `dbgen.GetEnumsByNamesParams` struct 加 `OrgName` 字段（重新运行 `just generate-sqlc`）
3. 所有调用处补传 `OrgName`
