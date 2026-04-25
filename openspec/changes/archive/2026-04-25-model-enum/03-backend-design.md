# 03 - 后端详细方案设计（ENUM / ENUM_LABEL / FieldEnumRelation）

## 1. 目标与范围

### 1.1 目标

围绕以下业务规则落地后端方案：

1. `ENUM` 字段只保存 `relateEnumName`（持久化为 `enumName`），不写 `enumRelationId`。  
2. `ENUM_LABEL` 字段必须保存 `enumRelationId`，不允许仅靠 `enumName` 创建。  
3. `FieldEnumRelation` 用于描述 `sourceFieldName + enumName` 的稳定关系，供 `ENUM_LABEL` 解析链路使用。  
4. 同一模型下，`sourceFieldName` 唯一（一个 source 只能绑定一条 relation）。  
5. 字段 `format` 创建可设、更新不可改（immutable）。

### 1.2 范围（In Scope）

- Project GraphQL Contract（`field.graphql` + `field_enum_relation.graphql`）
- Domain 模型与不变量
- App 用例编排（Add / Update / Remove / Relation Create）
- 错误模型与校验顺序
- BDD 测试设计（feature 级可执行场景）

### 1.3 明确不做（Out of Scope，关键）

> **本方案明确：FieldEnumRelation 不考虑迁移。**

以下内容全部不在本方案内：

- 不做“阶段化迁移”
- 不做“历史数据回填”
- 不做“双写兼容”
- 不保留旧入参长期兼容策略（一次性切换到新协议）

---

## 2. 核心设计决策

### 决策 A：ENUM 与 ENUM_LABEL 走不同绑定路径

- `ENUM`：`relateEnumName`（落库为 `enumName`）  
- `ENUM_LABEL`：`enumRelationId`（引用 `FieldEnumRelation.id`）

### 决策 B：FieldEnumRelation 作为 ENUM_LABEL 的唯一中间层

`FieldEnumRelation` 不承载普通字段定义，仅服务 label 解析与一致性校验。

### 决策 C：source 唯一约束前置到领域规则 + 存储约束双保险

- 领域规则：同一 `modelId + sourceFieldName` 只能创建一条 relation  
- 存储层：唯一约束兜底冲突

### 决策 D：format 强不可变

更新字段时禁止修改 `format`（包含类型与枚举绑定语义）。

### 决策 E：一次性切换（无迁移）

上线即只认新模型，不设计过渡期双轨逻辑。

---

## 3. GraphQL 契约设计（Project Schema）

## 3.1 Field 契约调整（`field.graphql`）

### `type Field` 字段

- `enumName: String`（仅 `format=ENUM` 有值）
- `enumRelationId: ID`（仅 `format=ENUM_LABEL` 有值）

### `AddFieldInput`（扁平输入）

- `relateEnumName: String`（`format=ENUM` 时必填）
- `enumRelationId: ID`（`format=ENUM_LABEL` 时必填）
- 移除旧嵌套输入（若存在）：`enumConfig` / `enumLabelConfig`

### `UpdateFieldInput`

- 不允许包含 `format` 变更语义字段  
- 仅允许更新展示/描述/校验等非 format 字段

---

## 3.2 FieldEnumRelation 契约（新增 `field_enum_relation.graphql`）

```graphql
# ============================================
# FieldEnumRelation Error Types
# ============================================

type FieldEnumSourceAlreadyHasLabel implements Error {
  message: String!
  suggestion: String
}

type FieldEnumSourceNotFound implements Error {
  message: String!
}

type FieldEnumSourceNotEnum implements Error {
  message: String!
}

type EnumNotFound implements Error {
  message: String!
}

type FieldEnumRelationNotFound implements Error {
  message: String!
}

# ============================================
# FieldEnumRelation Union Types
# ============================================

union CreateFieldEnumRelationError =
    FieldEnumSourceAlreadyHasLabel
  | FieldEnumSourceNotFound
  | FieldEnumSourceNotEnum
  | EnumNotFound

union DeleteFieldEnumRelationError =
    FieldEnumRelationNotFound

# ============================================
# FieldEnumRelation Payload Types
# ============================================

type CreateFieldEnumRelationPayload {
  relation: FieldEnumRelation
  error: CreateFieldEnumRelationError
}

type DeleteFieldEnumRelationPayload {
  success: Boolean!
  error: DeleteFieldEnumRelationError
}

# ============================================
# FieldEnumRelation Types
# ============================================

type FieldEnumRelation implements Node {
  id: ID!
  modelId: ID!
  sourceFieldName: String!
  enumName: String!
  createdAt: String!
  updatedAt: String!
}

# ============================================
# FieldEnumRelation Input Types
# ============================================

input CreateFieldEnumRelationInput {
  modelId: ID!
  sourceFieldName: String!
  enumName: String!
}

# ============================================
# Queries & Mutations
# ============================================

extend type Query {
  fieldEnumRelations(modelId: ID!): [FieldEnumRelation!]!
    @hasPermission(action: "field:read")
}

extend type Mutation {
  createFieldEnumRelation(input: CreateFieldEnumRelationInput!): CreateFieldEnumRelationPayload!
    @hasPermission(action: "field:update")

  deleteFieldEnumRelation(id: ID!): DeleteFieldEnumRelationPayload!
    @hasPermission(action: "field:update")
}
```

---

## 4. Domain 设计

## 4.1 FieldDefinition 不变量

1. `format=ENUM`：`enumName != nil && enumRelationId == nil`
2. `format=ENUM_LABEL`：`enumRelationId != nil && enumName == nil`
3. 其他 format：`enumName == nil && enumRelationId == nil`
4. `format` 创建后不可修改（更新流程必须对比旧值）

## 4.2 FieldEnumRelation 不变量

1. `sourceFieldName` 必须存在
2. source 字段 `format` 必须为 `ENUM`
3. `relation.enumName` 必须与 source 字段 `enumName` 一致
4. 唯一性：`UNIQUE(modelId, sourceFieldName)`

---

## 5. App 用例编排

## 5.1 AddField（创建字段）

### A) 创建 ENUM 字段

1. 校验 `relateEnumName` 必填
2. 若同时传入 `enumRelationId`，按容错策略忽略该参数（不报错）
3. 校验枚举存在（org + project + enumName）
4. 持久化：`enumName=relateEnumName`，`enumRelationId=nil`
5. 不自动创建 relation

### B) 创建 ENUM_LABEL 字段

1. 校验 `enumRelationId` 必填
2. 若同时传入 `relateEnumName`，按容错策略忽略该参数（不报错）
3. 根据 `enumRelationId` 查询 relation，必须存在且归属当前模型
4. 校验 relation 对应 source 字段仍为 `ENUM`
5. 持久化：`enumRelationId=...`，`enumName=nil`

### AddFields（批量）返回语义与回滚策略

1. 按字段独立处理（每个字段独立事务）。
2. 任一字段失败，仅回滚当前字段；已成功字段不回滚。
3. API 返回建议包含 `results[]`（逐字段 success/error）+ 最新 `model` 快照。

## 5.2 CreateFieldEnumRelation（创建 relation）

1. 校验 source 字段存在
2. 校验 source 字段 `format=ENUM`
3. 校验 `input.enumName == source.enumName`
4. 执行创建，若命中唯一约束返回冲突错误

## 5.3 UpdateField（更新字段）

- 任何 format 变更尝试一律拒绝：`FIELD_FORMAT_IMMUTABLE`

## 5.4 RemoveField（删除字段）

- 删除 source（ENUM）字段前，检查是否被 relation/label 引用  
- 若存在引用，拒绝删除：`FIELD_REFERENCE_IN_USE`

---

## 6. 持久化约束要求（无迁移版）

> 仅定义约束要求，**不包含迁移计划**。

1. `field_enum_relations` 必须有唯一约束：`(model_id, source_field_name)`
2. `field_definitions.enum_relation_id` 与 relation 主键建立引用约束
3. 领域不变量需在 App 层校验；数据库约束作为兜底
4. 本期不提供旧数据回填逻辑

---

## 7. 错误码与校验顺序

## 7.1 错误码建议

- `InvalidInput`（参数缺失/参数非法，统一入口）
- `FIELD_ENUM_SOURCE_CONFLICT`（同一 source 重复创建 label relation）
- `FIELD_FORMAT_IMMUTABLE`（format 不可修改）
- `FIELD_REFERENCE_IN_USE`（删除被引用 source 阻断）

## 7.2 校验顺序（推荐）

1. 格式与必填参数校验
2. 关联对象存在性校验
3. 跨对象一致性校验（source/enumName）
4. 唯一约束与引用约束兜底

---

## 8. BDD 测试设计（Feature 级，可执行）

> 建议文件：`tests-bdd/features/model-enum/field-enum-relation.feature`

### Feature 1：ENUM 字段创建仅依赖 relateEnumName

#### Scenario 1.1 创建 ENUM 成功（仅 relateEnumName）
**Given** 当前项目存在枚举 `CustomerLevel`  
**And** 模型 `Order` 中不存在同名字段 `level`  
**When** 调用 `addField`，`format=ENUM`，仅传 `relateEnumName="CustomerLevel"`  
**Then** 返回成功并创建字段 `level`  
**And** 字段 `enumName="CustomerLevel"`  
**And** 字段 `enumRelationId` 为空

#### Scenario 1.2 ENUM 互斥参数同时传入（容错）
**Given** 当前项目存在枚举 `CustomerLevel`  
**When** 调用 `addField`，`format=ENUM`，同时传 `relateEnumName` 与 `enumRelationId`  
**Then** 返回成功  
**And** 服务端忽略 `enumRelationId`

---

### Feature 2：ENUM_LABEL 字段创建必须提供 enumRelationId

#### Scenario 2.1 缺少 enumRelationId 创建失败
**Given** 模型 `Order` 已存在 source 字段 `level`（format=ENUM）  
**When** 调用 `addField`，`format=ENUM_LABEL`，不传 `enumRelationId`  
**Then** 返回错误 `InvalidInput`  
**And** 字段不落库

#### Scenario 2.2 提供有效 enumRelationId 创建成功
**Given** 模型 `Order` 已存在 relation `R1`（sourceFieldName=`level`）  
**When** 调用 `addField`，`format=ENUM_LABEL`，传 `enumRelationId=R1`  
**Then** 返回成功并创建 label 字段  
**And** 新字段 `enumRelationId=R1`

---

### Feature 3：source 唯一约束冲突

#### Scenario 3.1 同一 source 重复创建 relation 失败
**Given** 模型 `Order` 已有 relation `R1`，`sourceFieldName="level"`  
**When** 再次调用 `createFieldEnumRelation`，`sourceFieldName="level"`  
**Then** 返回错误 `FIELD_ENUM_SOURCE_CONFLICT`  
**And** 关系表仍只有一条该 source 记录

---

### Feature 4：format 不可修改

#### Scenario 4.1 更新时尝试修改 format 被拒绝
**Given** 字段 `level` 当前 `format=ENUM`，`enumName="CustomerLevel"`  
**When** 调用 `updateField` 试图改为 `format=TEXT` 或改绑 `enumName`  
**Then** 返回错误 `FIELD_FORMAT_IMMUTABLE`  
**And** 字段 format 保持原值

---

### Feature 5：删除被引用 source 的阻断

#### Scenario 5.1 source 被 relation 引用时删除失败
**Given** source 字段 `level` 已被 relation `R1` 引用  
**When** 调用 `removeField(level)`  
**Then** 返回错误 `FIELD_REFERENCE_IN_USE`  
**And** source 字段不被删除  
**And** 关联 relation 仍存在

---

## 9. 实施步骤（无迁移）

1. 调整 GraphQL Schema（`field.graphql`、新增 `field_enum_relation.graphql`）
2. 执行 `just generate-gql`
3. 实现 Domain 不变量与 App 校验流程
4. 落地 Repository 约束与错误映射
5. 补齐 Unit + BDD 场景
6. 联调前端，仅使用新协议字段（`relateEnumName` / `enumRelationId`）

---

## 10. 完成定义（DoD）

1. `ENUM` 创建只接受 `relateEnumName` 路径并可成功落库
2. `ENUM_LABEL` 创建缺少 `enumRelationId` 必失败
3. `UNIQUE(modelId, sourceFieldName)` 冲突可被稳定复现并返回业务错误
4. `updateField` 对 format 变更全部拒绝
5. 删除被引用 source 字段必阻断
6. 文档与实现中不包含迁移/回填/双写方案