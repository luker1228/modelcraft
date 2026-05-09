## 1. API 与契约调整

- [x] 1.1 更新 GraphQL schema：将 apply 接口语义改为模型级 reconcile（必要时新增 reconcile mutation 并保留兼容入口）
- [x] 1.2 为“权限包绑定模型预设”定义后端契约（新增专用 mutation 或扩展现有输入）
- [x] 1.3 补充错误类型与返回结构：`PresetRequiresOwnerField`、`PresetDeleteBlockedByBundle` 等

## 2. Repository 与 SQL 能力补齐

- [x] 2.1 新增按模型查询 PRESET 集合与按 `(model_id,type,name)` 定位记录的查询
- [x] 2.2 新增 PRESET 原地更新查询（用于 `toUpdate`）
- [x] 2.3 新增 toDelete 引用检查查询（判断是否被 bundle 引用）
- [x] 2.4 保留并复用唯一键约束，补充并发冲突处理路径

## 3. App 层 Reconcile 主流程

- [x] 3.1 实现 `desired/existing` 计算与 `toCreate/toUpdate/toDelete` 差异算法
- [ ] 3.2 实现单事务执行：create/update/delete 与失败回滚
- [x] 3.3 实现 `toDelete` 安全删除阻断逻辑（被引用则报错并回滚）
- [x] 3.4 实现模型级虚拟预设计算服务（只读）

## 4. 绑定时 Ensure 能力

- [x] 4.1 实现 `EnsurePresetPermission(modelId,preset)` 幂等服务
- [x] 4.2 在权限包绑定链路接入 ensure，再执行绑定写入
- [x] 4.3 对 OWNER 预设绑定补充 `END_USER_REF` 校验与错误映射
- [x] 4.4 验证重复绑定幂等（无重复 PRESET、无重复绑定）

## 5. 测试与回归

- [x] 5.1 新增 Reconcile 用例：新增预设补齐、原地更新、阻断删除、CUSTOM 不受影响
- [x] 5.2 新增 Ensure 用例：存在复用、不存在创建、重复请求幂等
- [x] 5.3 新增 OWNER 校验与错误映射测试
- [x] 5.4 回归现有权限查询与鉴权链路，确保行为兼容

## 6. 迁移与发布

- [ ] 6.1 提供 dry-run reconcile 报告能力（仅输出差异，不落库）
- [ ] 6.2 发布后执行一次全量 reconcile 并记录失败清单
- [ ] 6.3 在兼容期结束后移除旧 replace 语义分支与相关文档说明
