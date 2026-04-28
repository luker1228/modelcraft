## Context

现有实现中，`applyEndUserPresetPolicy` 采用“按模型删除全部 PRESET，再插入新记录”的流程。该流程在语义上更接近重置，而非同步；在工程上会带来两个问题：

1. 权限点 `permission_id` 不稳定，导致引用关系震荡。
2. 删除 PRESET 会触发外键级联，`end_user_bundle_permissions` 可能静默缩减。

同时，预设策略本身可由“内置预设目录 + 模型格式（是否有 END_USER_REF）”计算得到，查看阶段并不需要立即落库；真正需要实体权限点的时机通常在权限包绑定或显式同步。

## Goals / Non-Goals

**Goals:**
- 将 `applyEndUserPresetPolicy` 定义为“模型级预设集合差异同步（reconcile）”。
- 保持已存在预设权限点的 ID 稳定（优先原地更新）。
- 支持新增内置预设后的平滑补齐（`toCreate`）。
- 支持“查看虚拟预设 + 绑定时自动 ensure 实体权限”。
- 对 `*_OWNER` 预设继续执行 owner 字段校验。

**Non-Goals:**
- 不引入管理员自定义预设 DSL（仅内置预设）。
- 不在本次变更引入复杂版本编排系统（如多版本预设共存切换）。
- 不改变 CUSTOM 权限点模型与编辑能力。

## Decisions

### D1：`applyEndUserPresetPolicy` 从 Replace 改为 Reconcile

**决策**
- 接口语义改为：对指定模型执行一次“目标预设集合”的全量同步。
- 处理流程：
  - 计算 `desired`（由内置预设目录 + 模型格式得到）
  - 读取 `existing`（DB 中该模型 PRESET）
  - 计算差集：
    - `toCreate = desired - existing`
    - `toUpdate = desired ∩ existing`（row_policy 有差异）
    - `toDelete = existing - desired`（预设废弃/不适配）

**为何不用继续 Replace（全删全建）**
- 会导致 ID 抖动和级联删除副作用。
- 不利于后续新增预设的渐进升级。

---

### D2：对 `toUpdate` 采用原地更新，保持 permission_id 稳定

**决策**
- 通过 `(model_id, type, name)` 定位预设记录，更新 `row_policy/preset/name/description` 等可变字段。
- 不通过“先删后建”实现更新。

**备选方案**
- 方案A：删旧建新（放弃）
- 方案B：原地更新（采用）

**选择理由**
- 方案B保留引用稳定性，避免不必要的关联重建。

---

### D3：`toDelete` 采用安全删除策略（默认阻断被引用删除）

**决策**
- 当待删除 PRESET 仍被权限包引用时，默认返回结构化错误并终止本次 apply（事务回滚）。
- 仅删除未被引用的废弃 PRESET。

**备选方案**
- 方案A：直接删（依赖 FK CASCADE）
- 方案B：若被引用则阻断（采用）

**选择理由**
- 方案A会造成权限包静默缩减，排障成本高。
- 方案B行为可预期，便于前端提示用户先解绑或迁移。

---

### D4：查看阶段使用“虚拟预设视图”，绑定阶段执行 Ensure

**决策**
- 查询模型策略列表时，返回“可用预设视图”（可计算，不要求落库）。
- 在“权限包绑定模型预设”入口，执行 `EnsurePresetPermission(modelId, preset)`：
  - 存在则复用 permission_id
  - 不存在则按当前模型上下文创建

**备选方案**
- 方案A：查看即 apply 落库
- 方案B：绑定时 ensure（采用）

**选择理由**
- 方案B更贴合用户操作路径，降低初始化负担。

---

### D5：事务边界与幂等

**决策**
- 一次 `apply` 的 `toCreate/toUpdate/toDelete` 在单事务内提交。
- `EnsurePresetPermission` 设计为幂等：重复调用不产生重复记录。
- 并发冲突通过唯一键 + 重试策略处理。

## Risks / Trade-offs

- **[风险] 接口语义变更导致调用方误用** → **缓解**：在 GraphQL schema 与错误码中显式标注 Reconcile 语义；补充迁移说明。
- **[风险] 安全删除阻断导致 apply 失败率上升** → **缓解**：返回“被哪些 bundle 引用”的可读错误，指导先解绑或迁移。
- **[风险] 绑定时 ensure 增加写路径复杂度** → **缓解**：将 ensure 封装为单一 app service，集中测试幂等与并发。
- **[风险] 新增预设后全量同步耗时增加** → **缓解**：按 model 粒度执行，必要时引入批处理与分页同步。

## Migration Plan

1. 新增 reconcile 语义实现（保留旧入口一段兼容期，内部转发到 reconcile）。
2. 引入 `EnsurePresetPermission` 应用服务，并接入权限包“模型预设绑定”入口。
3. 将前端“模型策略查看”切换为虚拟视图来源（不依赖已落库 PRESET 才能展示）。
4. 发布前对现有模型执行一次 dry-run reconcile，输出差异报告。
5. 正式发布后执行一次全量 reconcile 作业；对失败模型输出可操作清单。
6. 兼容期结束后移除旧的 replace 语义分支。

## Open Questions

- `applyEndUserPresetPolicy` 是否保留原 mutation 名称，仅调整输入/语义，还是新增 `reconcile...` mutation 并逐步下线旧接口？
- `toDelete` 被引用时，是否需要支持 `force=false/true` 双模式，还是全程仅允许安全阻断？
- 权限包“绑定模型预设”是复用现有 `addEndUserPermissionToBundle` 扩展输入，还是新增专用 mutation 更清晰？
