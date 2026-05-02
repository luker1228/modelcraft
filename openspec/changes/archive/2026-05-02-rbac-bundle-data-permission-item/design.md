## Context

当前实现里有两个混淆点：

1. `PRESET` 和 `CUSTOM` 都复用了 `end_user_data_permissions` 这一层实体。
2. bundle 直接关联 permission，导致“绑定模板”在写入时必须先把模板展开并落成 permission。

这套模型在 MVP 阶段虽然能跑，但表达能力和约束都不对：
- `preset` 本质是内置模板定义，不应该伪装成一个可独立管理的持久化权限实体。
- `custom` 本质是管理员定义的模型级策略实体，适合复用和编辑。
- bundle 关心的不是 permission 本身，而是“某个模型在这个 bundle 下采用哪一种数据权限配置”。

本次变更允许破坏性调整，因此直接把数据结构收窄到以 bundle item 为中心：
- custom permission 保留为模型级实体
- preset 保留为虚拟模板定义
- bundle 内对数据权限的唯一正式绑定单元变为 item

## Goals / Non-Goals

**Goals:**
- 建立清晰的三层语义：preset template、custom permission、bundle data permission item。
- 保证“同一 bundle 下同一模型最多一个数据权限 item”。
- 支持两条显式写路径：绑定 preset、绑定 custom。
- 让 bundle 查看链路直接返回 item，避免前端再猜测绑定来源。
- 让鉴权链路以 item 为输入，统一展开为最终的模型 row/column 策略。
- 移除“preset 必须先落库成 permission 才能授权”的旧前提。

**Non-Goals:**
- 不在本次变更支持一个 item 内混合多个 custom permission。
- 不支持同一 bundle 下同一模型叠加 preset + custom。
- 不引入管理员自定义 preset 模板系统。
- 不处理功能权限；本次仅收敛数据权限模型。

## Decisions

### D1: bundle 的正式授权单元改为 `data permission item`

**决策**
- 新增 bundle-item 层，表示“bundle 在某模型上的数据权限配置”。
- item 至少包含：`bundle_id`、`model_id`、`grant_type`、`preset`、`custom_permission_id`、`sort_order`。
- `grant_type` 仅允许两种值：`PRESET`、`CUSTOM`。
- 当 `grant_type=PRESET` 时，`preset` 必填、`custom_permission_id` 为空。
- 当 `grant_type=CUSTOM` 时，`custom_permission_id` 必填、`preset` 为空。

**备选方案**
- 方案 A：继续让 bundle 直接关联 permission。
- 方案 B：增加 item 层表达 bundle-model 绑定关系。

**选择理由**
- 方案 A 无法自然表达“preset 是模板而不是实体”。
- 方案 B 能把“模板选择”和“实体引用”统一进同一授权模型，并承载 bundle-model 唯一约束。

### D2: preset 只作为虚拟模板定义存在，授权时必须带 model

**决策**
- 保留 `READ_ALL`、`READ_WRITE_ALL`、`READ_WRITE_OWNER`、`READ_ALL_WRITE_OWNER` 这些 preset 枚举。
- preset 自身不落库到 `end_user_data_permissions`。
- 管理员给 bundle 绑定 preset 时，接口必须显式传 `modelId + preset`，后端在运行时基于 model 上下文将 preset 展开成实际 row/column policy。

**备选方案**
- 方案 A：preset 先落库为 permission，再让 bundle 关联 permission。
- 方案 B：preset 只保留定义，授权时由 item 记录 `model + preset`。

**选择理由**
- 方案 A 让模板伪装成实体，造成多余生命周期管理。
- 方案 B 与“定义不耦合模型、授权必须选模型”的业务语义完全一致。

### D3: custom permission 保留为模型级实体，并允许与 preset 语义重合

**决策**
- `end_user_data_permissions` 收窄为只存 `CUSTOM` 类型数据权限实体。
- 不禁止 custom 策略与某个 preset 在语义上完全一致。
- preset 的价值定义为“降低配置心智负担的快捷模板”，而不是“语义上独占的一类能力”。

**备选方案**
- 方案 A：强制 custom 不能与 preset 重合。
- 方案 B：允许重合，但在模型层区分“模板”与“管理员实体”。

**选择理由**
- 方案 A 需要昂贵且脆弱的策略等价判断，也不符合管理员认知。
- 方案 B 更贴近真实业务动机：preset 是为了省事，不是为了表达一种无法被 custom 覆盖的新能力。

### D4: 同一 bundle 下同一模型最多一个 item，绑定新项时走 replace 语义

**决策**
- 在数据层添加 `UNIQUE(bundle_id, model_id)`。
- 当管理员为某个 bundle/model 再次绑定 preset 或 custom 时，后端执行 replace：旧 item 被替换，新 item 成为该模型的唯一配置。
- GraphQL mutation 语义明确为 upsert/replace，而不是 append。

**备选方案**
- 方案 A：允许一个模型在 bundle 下有多个 item，再在鉴权阶段合并。
- 方案 B：bundle-model 唯一，只保留一个 item。

**选择理由**
- 方案 A 会显著增加冲突解释、优先级规则和 UI 复杂度。
- 方案 B 更符合当前产品认知，也与用户已确认的业务规则一致。

### D5: bundle 查看链路以 item 为中心返回“来源详情”

**决策**
- `EndUserPermissionBundle.permissions` 改造成 `dataPermissionItems` 或等价字段，条目名称使用 `item` 而不是 `grant`。
- 每个 item 返回统一结构：模型信息、grantType、presetSummary、自定义 permission 摘要、最终 row/column policy 摘要。
- 前端查看 bundle 时无需再推断一个 permission 是 preset 还是 custom 绑定出来的。

**备选方案**
- 方案 A：沿用 `grant` 命名，强调鉴权结果。
- 方案 B：使用 `item` 命名，强调 bundle 中的配置项。

**选择理由**
- 这里讨论的是“bundle 中配置了什么”，不是“用户最终被授予什么”。
- `item` 比 `grant` 更直观，且不会和鉴权聚合结果混淆。

### D6: 鉴权阶段按 item 展开为最终 permission grants

**决策**
- 鉴权查询不再直接展开 `bundle -> permission`，而是展开 `bundle -> item`。
- 对 `CUSTOM` item，读取其引用的 custom permission 实体。
- 对 `PRESET` item，根据 `modelId + preset` 现算 row/column policy。
- 二者统一转换为鉴权内部使用的 effective grants。

**备选方案**
- 方案 A：预先把 preset item 物化成 permission 再复用旧鉴权逻辑。
- 方案 B：鉴权时直接解析 item。

**选择理由**
- 方案 B 更干净，避免为了兼容旧模型引入额外持久化状态。

## Risks / Trade-offs

- **[风险] 破坏性变更会使既有 SQL/GraphQL/前端调用整体失配** → **缓解**：直接新增 change spec，按 schema、resolver、前端查询统一切换，不保留双轨兼容。
- **[风险] replace 语义可能让管理员误以为是在追加** → **缓解**：mutation 命名和文案显式包含 bind/replace 语义，前端在同模型已有 item 时展示“将替换现有配置”。
- **[风险] preset 不落库后，排查时少了实体记录** → **缓解**：bundle item 保留 preset 枚举、模型、摘要字段；快照也记录 item 明细，可追溯历史。
- **[风险] custom permission 被删除后，bundle item 可能悬挂** → **缓解**：对 custom permission 删除增加引用检查，或采用 FK RESTRICT 阻断删除。

## Migration Plan

1. 重写 `13_rbac_permissions.sql`，将 permission 表收窄为 custom permission，并新增 bundle data permission item 表。
2. 重写 `14_rbac_bundle_snapshots.sql`，快照改存 item 列表而非 permission ID 列表。
3. 调整 repository / app / authz 查询，统一基于 item 解析数据权限。
4. 调整 GraphQL schema：移除/废弃 apply preset to model 的落库语义，新增 bind preset/custom item 的 mutation 与 item 查询类型。
5. 调整前端 bundle 详情页与添加弹窗，使其消费 item 结构、preset 模板列表、custom permission 列表。
6. 对现有 MVP 数据不做兼容迁移要求；开发环境可直接重建表结构。

## Open Questions

- bundle GraphQL 字段最终命名是 `dataPermissionItems` 还是直接简化为 `items`？
- preset item 的返回结构里，是否需要直接返回“展开后的 rowPolicy JSON”，还是只返回摘要 + 单独 detail 字段？
- 自定义 permission 的编辑/删除入口，是否要在 bundle 详情页直接跳转到模型权限管理页？
