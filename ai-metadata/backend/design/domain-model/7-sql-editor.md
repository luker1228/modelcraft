# 7. SQL 编辑器

> 状态：**未实现，Future Milestone**

## 规划定位

SQL 编辑器允许用户在 ModelCraft 界面中直接对目标数据库执行 SQL 查询，用于：

- 数据探查和调试
- 临时查询，不经过 GraphQL 层
- 辅助模型设计阶段的数据验证

## 与现有架构的关系

```
SQL 编辑器
    │
    ├── 复用 DatabaseCluster 的连接信息
    ├── 需要独立的权限控制（非所有角色可执行裸 SQL）
    └── 执行结果不经过 modelruntime，直接返回原始行数据
```

## 边界约束

- 只支持 SQL 系数据库（MySQL，未来 PostgreSQL）
- 不支持 MongoDB 等 NoSQL（与 core-principles.md 一致）
- 写操作需要额外权限确认

## 相关文档

- [core-principles.md](../core-principles.md) — 仅支持 SQL 系原则
- [roadmap.md](../roadmap.md) — 里程碑规划
