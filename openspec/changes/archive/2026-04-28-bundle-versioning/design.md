## Context

ModelCraft 后端采用 DDD 分层架构（Domain → UseCase → Infrastructure → Interfaces）。权限包（`EndUserPermissionBundle`）已在 `internal/domain/enduserrbac/` 中定义，通过 GraphQL Project Schema 对外暴露。当前权限包没有任何历史记录机制，每次修改权限列表后状态即时覆盖。

相关文件路径：
- GraphQL Schema: `api/graph/project/schema/permission.graphql`
- 领域层: `internal/domain/enduserrbac/`
- Use Case 层: `internal/usecase/enduserrbac/`
- Infrastructure: `internal/infrastructure/postgres/enduserrbac/`
- Resolver: `internal/interfaces/graphql/`
- 数据库迁移: `db/schema/mysql/`（Atlas 管理）

## Goals / Non-Goals

**Goals:**
- 每次权限列表变更（添加/移除权限点）后自动保存快照
- 每个权限包最多保留最近 5 个历史版本，超出时滚动删除最旧版本
- 提供 `restoreEndUserPermissionBundle` mutation 支持一键回滚
- `EndUserPermissionBundle` GraphQL 类型新增 `currentVersion` 和 `snapshots` 字段

**Non-Goals:**
- 前端 UI 实现（后端就绪后单独跟进）
- 快照自动过期（基于时间的 TTL）
- 快照的权限点完整数据归档（存 ID 引用，不复制权限点内容）
- 修改名称/描述触发快照

## Decisions

### 决策 1：快照存储策略——ID 引用 vs 完整数据复制

**选择**：JSONB 数组存储权限点 ID + sortOrder，不复制权限点内容

**理由**：
- 权限点（`EndUserPermission`）本身是稳定的命名实体，不频繁变更
- 存完整数据会造成大量冗余，且权限点改名后快照内容不一致
- 查询快照详情时通过 ID JOIN 当前权限点，已删除的标注 `[已删除]`，信息完整

**替代方案**：存储完整权限点快照 → 放弃，数据膨胀且引入一致性问题

---

### 决策 2：版本号管理——`current_version` 冗余字段 vs 从 snapshot 表 MAX 派生

**选择**：从 `end_user_permission_bundle_snapshots` 表 MAX(version) 派生，不在 bundle 表新增冗余字段

**理由**：
- 避免两表之间的数据同步问题
- 快照表写入和版本号计算在同一事务内，MAX 查询一致性有保障
- 简化 Schema 变更（无需修改 bundle 表结构）

**替代方案**：在 bundle 表加 `current_version` 列 → 放弃，冗余字段需额外维护

---

### 决策 3：滚动保留清理时机——写入后同事务清理 vs 异步后台任务

**选择**：写入新快照后，在同一数据库事务内执行 DELETE，保留最新 5 个

**理由**：
- 数据量小（每个 bundle 最多 5+1 条），事务内清理开销可忽略
- 保证清理操作与写入原子一致，避免后台任务失败导致超出上限

**替代方案**：异步清理 → 放弃，实现复杂度高且一致性弱

---

### 决策 4：回滚操作的语义——原地覆盖 vs 生成新版本

**选择**：回滚生成新版本号，`restored_from` 字段指向来源版本

**理由**：
- 保持快照链完整可追溯，审计时可知"v4 是从 v2 回滚来的"
- 与添加/移除权限点操作一致，都是"状态变更 → 生成快照"的模式
- 支持从回滚后的状态再次回滚，无歧义

**替代方案**：原地覆盖 → 放弃，破坏审计链，违背"不可变快照"原则

## Risks / Trade-offs

- **[风险] 快照表并发写入竞争** → `UNIQUE (bundle_id, version)` 约束 + 数据库序列化隔离，冲突时报错重试
- **[风险] 已删除权限点 JOIN 返回 null** → GraphQL `EndUserPermissionSnapshotEntry.permission` 字段设计为可空，前端需处理 null 展示为 `[已删除]`
- **[Trade-off] 查询快照详情需 JOIN 权限点表** → 相比纯 JSONB 读取多一次查询，但数据量小，可接受

## Migration Plan

1. 编写 Atlas 迁移文件，创建 `end_user_permission_bundle_snapshots` 表
2. 运行 `just db up` 应用迁移
3. 新增 sqlc 查询，运行 `just generate-sqlc` 生成 Go 代码
4. 实现 Domain 层值对象和 Repository 接口
5. 修改 Use Case 层，在事务内写快照
6. 实现 restore Use Case
7. 运行 `just generate-gql` 生成 GraphQL resolver 接口
8. 实现 resolver

**回滚策略**：如部署失败，回滚迁移（表为新增，DROP TABLE 即可），代码回滚至上一版本，无数据丢失风险。

## Open Questions

- 快照 `created_by` 字段从 context 中获取操作人标识，需确认当前 Use Case 层 context 是否已携带该信息
