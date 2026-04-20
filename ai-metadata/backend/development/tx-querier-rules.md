# Tx Querier 规范（Application 层）

适用范围：`modelcraft-backend/internal/app/**/*.go`

## 背景

`TxManager.WithTx` 会向回调传入与当前事务绑定的 `dbgen.Querier`。如果回调里继续使用服务对象上注入的常规 repo（通常绑定到非事务 DB 连接），就会出现“看起来在事务里，实际部分操作不在事务里”的问题。

## 强制规则

1. `WithTx` 回调参数禁止使用 `_ dbgen.Querier`。
2. `WithTx` 回调内禁止直接调用 `s.*Repo.*`（例如 `s.projectRepo.Create(...)`）。
3. 回调内必须基于回调参数 `q` 创建 tx-scoped repo，再执行写操作。

## 推荐写法

```go
err := s.txManager.WithTx(ctx, func(ctx context.Context, q dbgen.Querier) error {
    txProjectRepo := repository.NewSqlProjectRepository(q)
    txClusterRepo := repository.NewSqlDatabaseClusterRepository(q)

    if err := txProjectRepo.Create(ctx, proj); err != nil {
        return err
    }
    if err := txClusterRepo.Create(ctx, clusterEntity); err != nil {
        return err
    }
    return nil
})
```

## 自动检查

项目提供静态检查脚本：

- `modelcraft-backend/scripts/check-withtx-repo-usage.sh`

该脚本会检查：

- 是否存在 `_ dbgen.Querier`
- `WithTx` 回调内是否存在 `s.*Repo.*` 调用

并已集成到 `just lint` / `just lint-fix` 的预检查阶段。
