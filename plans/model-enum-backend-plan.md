# Model Enum 后端落地技术方案

## 0. 文档目标

基于以下输入输出后端落地方案（仅方案，不含代码实现）：

- `ai-metadata/prd/model-enum/00-model-enum.md`
- `ai-metadata/prd/model-enum/model-enum-domain.puml`
- `ai-metadata/prd/model-enum/01-field-create-enum-binding.md`
- `ai-metadata/prd/model-enum/02-field-edit-format-immutable.md`
- `ai-metadata/prd/model-enum/03-backend-design.md`
- `ai-metadata/prd/model-enum/04-frontend-subpage-design.md`
- `ai-metadata/prd/model-enum/api-contract.md`

本方案严格对齐约束：

1. `FieldEnumRelation` 不考虑迁移（不做回填/双写/兼容期）
2. `ENUM` 使用 `relateEnumName`；`ENUM_LABEL` 使用 `enumRelationId`
3. 参数缺失/参数非法统一走 `InvalidInput`
4. 唯一约束：`UNIQUE(model_id, source_field_name)`
5. `format` 不可修改
6. 补充 BDD / unit 验收口径（integration 不纳入本期）

---

## 1. 现状基线（用于落地对齐）

### 1.1 Contract 现状

- 当前 `AddFieldInput` 仍使用 `enumConfig` / `enumLabelConfig`，尚未扁平化为 `relateEnumName` / `enumRelationId`：`modelcraft-backend/api/graph/project/schema/field.graphql:119`
- 当前 `AddFieldsError` 仅有 `InvalidModelInput`：`modelcraft-backend/api/graph/project/schema/field.graphql:160`

### 1.2 领域与应用现状

- 字段实体当前仍有 `EnumLabelConfig`，并以 `ENUM_LABEL` + `sourceField` 做校验：`modelcraft-backend/internal/domain/modeldesign/field_definition.go:11`
- `AddFieldSync` 当前按 `IsEnumLabelField()` 分支处理“虚拟字段”，未使用 `enumRelationId`：`modelcraft-backend/internal/app/modeldesign/model_app.go:310`

### 1.3 数据库现状

- `field_definitions` 当前有 `enum_name` 与 `enum_label_config`，无 `enum_relation_id`：`modelcraft-backend/db/schema/mysql/03_model_domain.sql:93`
- 现有枚举关联表为 `model_field_enum_associations`，非本次目标 `field_enum_relations`：`modelcraft-backend/db/schema/mysql/03_model_domain.sql:231`

---

## 2. 目标态设计概览

目标态采用“两条绑定链路”并彻底切换：

- `ENUM` 字段：
  - 输入：`relateEnumName`
  - 落库：`field_definitions.enum_name`
  - 不使用 `enumRelationId`
- `ENUM_LABEL` 字段：
  - 输入：`enumRelationId`
  - 落库：`field_definitions.enum_relation_id`
  - 不使用 `relateEnumName`

`FieldEnumRelation` 作为 `ENUM_LABEL` 的唯一中间层，核心约束：`UNIQUE(model_id, source_field_name)`。

---

## 3. 数据库设计

## 3.1 表结构调整

### A. `field_definitions`（增量调整）

新增字段：

- `enum_relation_id VARCHAR(36) NULL`：仅 `format=ENUM_LABEL` 使用

删除字段：

- `enum_label_config`：直接删除，不再保留历史兼容

说明：

- `relateEnumName` 是 API 输入参数，不新增数据库列；仍映射到 `field_definitions.enum_name`

建议索引：

- `idx_field_enum_relation_id (enum_relation_id)`
- 保持现有 `PRIMARY KEY(model_id, name)` 与项目维度索引

### B. 新增 `field_enum_relations`

核心字段：

- `id VARCHAR(36) PK`
- `model_id VARCHAR(36) NOT NULL`
- `label_field_name VARCHAR(64) NOT NULL`
- `source_field_name VARCHAR(64) NOT NULL`
- `org_name VARCHAR(36) NOT NULL`
- `project_slug VARCHAR(64) NOT NULL`
- `enum_name VARCHAR(64) NOT NULL`
- `created_at/updated_at`

关键约束：

- **唯一约束（强制）**：`UNIQUE(model_id, source_field_name)`
- `FOREIGN KEY (model_id) -> models(id) ON DELETE CASCADE`

建议索引：

- `idx_field_enum_relations_project (org_name, project_slug)`
- `idx_field_enum_relations_model (model_id)`
- `idx_field_enum_relations_enum (org_name, project_slug, enum_name)`

## 3.2 本期明确不做

- 不做历史 `enum_label_config` → `enum_relation_id` 回填
- 不做 `model_field_enum_associations` 与 `field_enum_relations` 双写
- 不做兼容分支读取旧协议

---

## 4. 分层落点

## 4.1 Domain 层

新增/调整：

1. 新增 `FieldEnumRelation` 领域实体与仓储接口（`Create/Get/List/Delete`）
2. `FieldDefinition` 不变量收敛：
   - `format=ENUM`：`enumName` 必填，`enumRelationId` 为空
   - `format=ENUM_LABEL`：`enumRelationId` 必填，`enumName` 为空
3. 更新语义：`format` 及其绑定语义不可变

## 4.2 Repository 层

1. 新增 `db/queries/field_enum_relation.sql` 与 sqlc 生成
2. 新增 `SqlFieldEnumRelationRepository`
3. `SqlModelDesignRepository` 字段映射补充 `enum_relation_id`（读写）
4. 唯一冲突（`UNIQUE(model_id, source_field_name)`）映射为业务冲突错误

## 4.3 App 层

### A. `AddFieldSync`

- `ENUM`：校验 `relateEnumName`，校验枚举存在，落库 `enum_name`
- `ENUM_LABEL`：校验 `enumRelationId`，查 relation 并验证 model/source/enum 一致性，落库 `enum_relation_id`

### B. `UpdateFieldSync`

- 禁止修改 `format` 与枚举绑定语义（`FIELD_FORMAT_IMMUTABLE`）

### C. `RemoveFieldSync`

- 删除 `ENUM` 源字段前检查 `field_enum_relations` 是否引用，存在则拒绝（`FIELD_REFERENCE_IN_USE`）

### D. 新增 relation 用例

- `CreateFieldEnumRelation`
- `DeleteFieldEnumRelation`
- `ListFieldEnumRelations`

## 4.4 Resolver / Adapter 层

1. Project GraphQL schema 变更：
   - `field.graphql`：输入扁平化、Error Union 扩展
   - 新增 `field_enum_relation.graphql`
2. `FieldMapper` 从 `AddFieldInput` 解析：
   - `relateEnumName`
   - `enumRelationId`
3. 错误适配器统一映射 `InvalidInput` 与 relation 专项错误

---

## 5. 核心流程与事务边界

## 5.1 AddFields（批量）

事务边界：**按字段独立事务**（每个字段单独提交）

流程顺序（每个字段）：

1. 输入合法性（format 与参数矩阵）
2. 关联对象存在性（enum / relation）
3. 跨对象一致性（relation.source 必须是 ENUM；relation.enumName 对齐）
4. 持久化字段

失败策略：任一字段失败，仅回滚当前字段；已成功添加字段不回滚。其余字段继续执行。

## 5.2 CreateFieldEnumRelation

事务边界：**单事务**

1. 校验 source 字段存在且 `format=ENUM`
2. 校验 `input.enumName == source.enumName`
3. 插入 relation（依赖 DB 唯一约束兜底）

## 5.3 UpdateField

事务边界：单事务。

- 若触发 format 或绑定语义变更尝试，直接失败，不产生部分更新。

## 5.4 RemoveField

事务边界：单事务。

- 删除 source ENUM 字段前先查 `field_enum_relations` 引用；有引用即拒绝。

---

## 6. 错误映射（统一 InvalidInput）

| 场景 | Biz 错误语义 | GraphQL 错误类型 |
|---|---|---|
| modelID/projectSlug/orgName 缺失或非法 | ParamInvalid(模型上下文) | `InvalidInput` |
| `format=ENUM` 缺少/非法 `relateEnumName` | ParamInvalid(字段参数) | `InvalidInput` |
| `format=ENUM_LABEL` 缺少/非法 `enumRelationId` | ParamInvalid(字段参数) | `InvalidInput` |
| 互斥参数同时传（如 ENUM 同时传 relationId） | 容错过滤（忽略多余参数） | 无错误 |
| source 重复创建 relation | Conflict | `FieldEnumSourceConflict` (`FIELD_ENUM_SOURCE_CONFLICT`) |
| 更新时尝试改 format/绑定语义 | OperationFailed | `FieldFormatImmutable` (`FIELD_FORMAT_IMMUTABLE`) |
| 删除被引用 source 字段 | OperationFailed | `FieldReferenceInUse` (`FIELD_REFERENCE_IN_USE`) |

说明：参数类错误统一收敛到 `InvalidInput`，不再区分 `InvalidModelInput/InvalidFieldInput`。

---

## 7. 实施顺序

1. **Contract 先行**
   - 更新 `api/graph/project/schema/field.graphql`
   - 新增 `api/graph/project/schema/field_enum_relation.graphql`
   - 执行 `just generate-gql`
2. **存储层**
   - 新增 `field_enum_relations` 表
   - `field_definitions` 增加 `enum_relation_id`
   - 更新 `db/queries/*.sql` 并执行 `just generate-sqlc`
3. **Domain + Repository**
   - 落地 relation 实体/仓储
   - 更新字段实体不变量
4. **App**
   - 改造 Add/Update/Remove 字段流程
   - 新增 relation 用例
5. **Resolver + Adapter**
   - 输入映射与错误映射切换
6. **测试与验收**
   - unit / BDD 通过后再切换前端联调

---

## 8. 风险点与回滚思路

## 8.1 风险点

1. **协议破坏性变更**：前端若未同步 contract，会出现字段不匹配
2. **无迁移策略风险**：历史依赖 `enum_label_config` 的存量数据不在本期保障范围
3. **并发冲突**：relation 并发创建触发唯一约束
4. **删除竞态**：source 删除与 relation 创建并发

## 8.2 缓解策略

1. 后端 contract 发布与前端 contract sync 同步执行
2. 所有关键冲突由 DB 约束兜底 + 业务错误映射
3. 删除与创建流程均在事务中执行，必要处加行级锁

## 8.3 回滚思路

1. **应用回滚**：回退到上一个后端版本（首选）
2. **数据库回滚策略**：采用“增量 DDL”，回滚时允许保留新增表/列（避免二次数据风险）
3. **接口回滚**：如需快速止血，先下线 relation mutation（读接口可保留）

---

## 9. 验收清单

## 9.1 BDD（业务场景验收）

新增/扩展场景（建议在 `tests-bdd/features/field/manage-field.feature` 拆分子场景）：

1. `ENUM` 创建成功：仅传 `relateEnumName`
2. `ENUM` 创建失败：缺少 `relateEnumName` -> `InvalidInput`
3. `ENUM_LABEL` 创建失败：缺少 `enumRelationId` -> `InvalidInput`
4. `ENUM_LABEL` 创建成功：`enumRelationId` 有效
5. relation source 唯一冲突 -> `FIELD_ENUM_SOURCE_CONFLICT`
6. 更新时改 format/绑定语义 -> `FIELD_FORMAT_IMMUTABLE`
7. 删除被引用 source -> `FIELD_REFERENCE_IN_USE`

## 9.2 Integration（本期不纳入）

- integration 测试已废弃，本期不作为验收门禁。

## 9.3 Unit（领域与应用验收）

1. `FieldDefinition` 不变量测试（ENUM/ENUM_LABEL/其他）
2. `AddFieldSync` 参数矩阵测试
3. `CreateFieldEnumRelation` 一致性与冲突测试
4. `UpdateFieldSync` immutable 测试
5. `RemoveFieldSync` 引用阻断测试
6. adapter 错误映射测试（InvalidInput + 业务错误）

---

## 10. 完成定义（DoD）

1. 新协议输入完全生效：`relateEnumName` / `enumRelationId`
2. 参数缺失/非法统一映射到 InvalidInput
3. 互斥参数同时传按容错策略过滤多余参数，不作为失败
4. 唯一约束 `UNIQUE(model_id, source_field_name)` 生效且错误可识别
5. `format` 不可修改规则在 App + Resolver 双层生效
6. 删除 source 引用阻断可稳定复现
7. BDD / unit 验收项通过（integration 不作为门禁）
8. 文档与实现中无迁移/回填/双写方案
