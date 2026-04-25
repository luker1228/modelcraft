# Model Enum API Contract（Project GraphQL）

## 1. 协议归属

本需求归属 **Project GraphQL**：
- Schema 目录：`modelcraft-backend/api/graph/project/schema/`
- Endpoint：`/graphql/org/{orgName}/project/{projectSlug}/`

---

## 2. Contract 变更总览

| 变更点 | 目标 |
|---|---|
| Field | 增补 `enumName` / `enumRelationId` 语义（按 format 互斥） |
| AddFieldInput | 扁平化：`relateEnumName`、`enumRelationId`，移除 `enumConfig/enumLabelConfig` |
| FieldEnumRelation | 新增类型与 Query/Mutation |
| UpdateFieldInput | 继续不支持 format 变更 |

---

## 3. 错误模型（统一 InvalidInput）

### 3.1 规则

1. **参数缺失/参数非法**统一使用通用错误：`InvalidInput`
2. 仅对业务冲突/状态约束使用专用错误：
   - `FieldEnumSourceConflict`
   - `FieldFormatImmutable`
   - `FieldReferenceInUse`

> 也即：不再为“缺少 relateEnumName / enumRelationId”重复定义模块专属错误类型。

### 3.2 建议错误类型（Project GraphQL）

```graphql
# 通用输入错误
# 建议在 project/base.graphql 中统一定义并全模块复用

type InvalidInput implements Error {
  message: String!
  suggestion: String
}

# 业务冲突/约束错误

type FieldEnumSourceConflict implements Error {
  message: String!
  code: String!         # FIELD_ENUM_SOURCE_CONFLICT
  suggestion: String
}

type FieldFormatImmutable implements Error {
  message: String!
  code: String!         # FIELD_FORMAT_IMMUTABLE
}

type FieldReferenceInUse implements Error {
  message: String!
  code: String!         # FIELD_REFERENCE_IN_USE
  suggestion: String
}
```

---

## 4. GraphQL Contract（建议稿）

## 4.1 Field / AddFieldInput / UpdateFieldInput

```graphql
type Field {
  name: String!
  title: String!
  format: FormatType!
  isArray: Boolean!
  description: String
  validationConfig: ValidationConfig

  enum: EnumDefinition
  enumName: String        # format=ENUM 时有值
  enumRelationId: ID      # format=ENUM_LABEL 时有值

  createdAt: String!
  updatedAt: String!
}

input AddFieldInput {
  name: String!
  title: String!
  format: FormatType!
  storageHint: String
  nonNull: Boolean = false
  required: Boolean = false
  isUnique: Boolean = false
  isArray: Boolean = false
  description: String
  validationConfig: ValidationConfigInput
  relateFkId: String

  # 新增扁平字段
  relateEnumName: String  # format=ENUM 时必填
  enumRelationId: ID      # format=ENUM_LABEL 时必填

  # 移除：enumConfig / enumLabelConfig
}

input UpdateFieldInput {
  title: String
  description: String
  validationConfig: ValidationConfigInput
  # 不允许 format / relateEnumName / enumRelationId 更新语义
}
```

### 4.1.1 AddFieldInput 规则矩阵

| format | 必填 | 容错处理 |
|---|---|---|
| ENUM | `relateEnumName` | 若传 `enumRelationId`，忽略该参数 |
| ENUM_LABEL | `enumRelationId` | 若传 `relateEnumName`，忽略该参数 |
| 其他类型 | 无 | 忽略 `relateEnumName`、`enumRelationId` |

参数缺失/参数非法统一返回 `InvalidInput`。互斥参数同时传按容错策略处理，不返回错误。

---

## 4.2 FieldEnumRelation（新增）

```graphql
type FieldEnumRelation implements Node {
  id: ID!
  modelId: ID!
  labelFieldName: String!
  sourceFieldName: String!
  enumName: String!
  orgName: String!
  projectSlug: String!
  createdAt: String!
  updatedAt: String!
}

input CreateFieldEnumRelationInput {
  modelId: ID!
  labelFieldName: String!
  sourceFieldName: String!
  enumName: String!
}
```

---

## 4.3 Union / Payload 建议

```graphql
union AddFieldsError =
    InvalidInput

union UpdateFieldError =
    InvalidInput
  | FieldFormatImmutable

union RemoveFieldError =
    InvalidInput
  | FieldReferenceInUse

union CreateFieldEnumRelationError =
    InvalidInput
  | FieldEnumSourceConflict

union DeleteFieldEnumRelationError =
    InvalidInput

type AddFieldItemResult {
  name: String!
  success: Boolean!
  error: AddFieldsError
}

type AddFieldsPayload {
  model: Model
  results: [AddFieldItemResult!]!
  error: AddFieldsError
}

type UpdateFieldPayload {
  model: Model
  error: UpdateFieldError
}

type RemoveFieldPayload {
  model: Model
  error: RemoveFieldError
}

type CreateFieldEnumRelationPayload {
  relation: FieldEnumRelation
  error: CreateFieldEnumRelationError
}

type DeleteFieldEnumRelationPayload {
  success: Boolean!
  error: DeleteFieldEnumRelationError
}

extend type Query {
  fieldEnumRelations(modelID: ID!): [FieldEnumRelation!]!
    @hasPermission(action: "field:read")
}

extend type Mutation {
  addFields(modelID: ID!, input: [AddFieldInput!]!): AddFieldsPayload!
    @hasPermission(action: "field:create")

  updateField(modelID: ID!, fieldName: String!, input: UpdateFieldInput!): UpdateFieldPayload!
    @hasPermission(action: "field:update")

  removeField(modelID: ID!, fieldName: String!): RemoveFieldPayload!
    @hasPermission(action: "field:delete")

  createFieldEnumRelation(input: CreateFieldEnumRelationInput!): CreateFieldEnumRelationPayload!
    @hasPermission(action: "field:create")

  deleteFieldEnumRelation(id: ID!): DeleteFieldEnumRelationPayload!
    @hasPermission(action: "field:delete")
}
```

---

## 5. 错误映射表

| 场景 | 返回错误类型 | code |
|---|---|---|
| `format=ENUM` 缺少 `relateEnumName` | `InvalidInput` | - |
| `format=ENUM_LABEL` 缺少 `enumRelationId` | `InvalidInput` | - |
| 互斥字段同时传（如 ENUM 同时传 relationId） | 容错过滤，不返回错误 | - |
| 同一 `modelId+sourceFieldName` 重复建 relation | `FieldEnumSourceConflict` | `FIELD_ENUM_SOURCE_CONFLICT` |
| 更新尝试变更 format 语义 | `FieldFormatImmutable` | `FIELD_FORMAT_IMMUTABLE` |
| 删除被引用 source 字段 | `FieldReferenceInUse` | `FIELD_REFERENCE_IN_USE` |

---

## 6. 请求/响应示例

## 6.1 成功：创建 ENUM 字段

### Request
```graphql
mutation AddEnumField($modelID: ID!, $input: [AddFieldInput!]!) {
  addFields(modelID: $modelID, input: $input) {
    model {
      id
      fields { name format enumName enumRelationId }
    }
    results {
      name
      success
      error {
        __typename
        ... on InvalidInput { message suggestion }
      }
    }
    error {
      __typename
      ... on InvalidInput { message suggestion }
    }
  }
}
```

```json
{
  "modelID": "mdl_order",
  "input": [{
    "name": "level",
    "title": "客户等级",
    "format": "ENUM",
    "relateEnumName": "CustomerLevel"
  }]
}
```

### Response
```json
{
  "data": {
    "addFields": {
      "model": {
        "id": "mdl_order",
        "fields": [{
          "name": "level",
          "format": "ENUM",
          "enumName": "CustomerLevel",
          "enumRelationId": null
        }]
      },
      "results": [{
        "name": "level",
        "success": true,
        "error": null
      }],
      "error": null
    }
  }
}
```

## 6.2 失败：创建 ENUM_LABEL 缺少 enumRelationId（统一 InvalidInput）

### Request
```json
{
  "modelID": "mdl_order",
  "input": [{
    "name": "levelLabel",
    "title": "客户等级标签",
    "format": "ENUM_LABEL"
  }]
}
```

### Response
```json
{
  "data": {
    "addFields": {
      "model": null,
      "results": [{
        "name": "levelLabel",
        "success": false,
        "error": {
          "__typename": "InvalidInput",
          "message": "enumRelationId is required when format=ENUM_LABEL",
          "suggestion": "请先创建或选择有效的 FieldEnumRelation"
        }
      }],
      "error": {
        "__typename": "InvalidInput",
        "message": "enumRelationId is required when format=ENUM_LABEL",
        "suggestion": "请先创建或选择有效的 FieldEnumRelation"
      }
    }
  }
}
```

## 6.3 失败：source 冲突（重复 relation）

### Response
```json
{
  "data": {
    "createFieldEnumRelation": {
      "relation": null,
      "error": {
        "__typename": "FieldEnumSourceConflict",
        "code": "FIELD_ENUM_SOURCE_CONFLICT",
        "message": "source field already has an enum label relation",
        "suggestion": "同一 sourceFieldName 只能绑定一个 label relation"
      }
    }
  }
}
```

## 6.4 失败：更新字段尝试修改 format

### Response
```json
{
  "data": {
    "updateField": {
      "model": null,
      "error": {
        "__typename": "FieldFormatImmutable",
        "code": "FIELD_FORMAT_IMMUTABLE",
        "message": "field format is immutable after creation"
      }
    }
  }
}
```

---

## 7. 与当前设计文档对齐

- 总览：`ai-metadata/prd/model-enum/00-model-enum.md`
- 后端方案：`ai-metadata/prd/model-enum/03-backend-design.md`
- 前端子页：`ai-metadata/prd/model-enum/04-frontend-subpage-design.md`