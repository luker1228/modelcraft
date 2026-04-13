# 00 - Model Enum 方案总览

## 1. 背景与目标

字段建模需要同时满足两类诉求：

1. `ENUM` 字段必须绑定明确枚举定义，避免脏数据。
2. `ENUM_LABEL` 字段必须通过稳定 relation 定位 source ENUM 字段与枚举，避免运行时歧义。

本方案以**数据一致性 + 关系可追踪 + 交互可落地**为目标。

---

## 2. 范围与边界

### In Scope

1. `ENUM` 字段创建规则（`relateEnumName`）。
2. `ENUM_LABEL` 字段创建规则（`enumRelationId`）。
3. `FieldEnumRelation` 关系模型与唯一约束。
4. 字段 `format` 不可变规则（创建可设、编辑不可改）。
5. 前后端子页/方案设计与验收标准。

### Out of Scope

1. `FieldEnumRelation` 历史数据迁移、回填、双写兼容。
2. 复制/导入流程专项设计。
3. 历史脏数据自动修复工具。

---

## 3. 术语与字段单一真相

- `format`: 字段类型（含 `ENUM` / `ENUM_LABEL`）
- `relateEnumName`: `ENUM` 创建时的枚举绑定入参
- `enumRelationId`: `ENUM_LABEL` 创建时必须提供的 relation 引用
- `FieldEnumRelation`: relation 实体，记录 `modelId + sourceFieldName + enumName (+ project scope)`

强规则：

1. `format=ENUM` → 必须 `relateEnumName`，且不使用 `enumRelationId`。
2. `format=ENUM_LABEL` → 必须 `enumRelationId`，且不使用 `relateEnumName`。
3. `ENUM` 字段不写 relation（A 不入 relation）。
4. 唯一约束：`UNIQUE(modelId, sourceFieldName)`（同一 source 只能产生一个 label）。
5. `format` 创建后不可修改。

---

## 4. 规则矩阵（交互与后端一致）

| 字段类型 | 创建 | 编辑 | 删除 |
|---|---|---|---|
| ENUM | 必须传 `relateEnumName`；不得传 `enumRelationId` | `format` 与关联语义只读，不可改绑 | 若被 relation 作为 source 引用，删除阻断 |
| ENUM_LABEL | 必须传 `enumRelationId`；不得传 `relateEnumName` | `format` 与 `enumRelationId` 只读 | 删除字段时联动清理/失效 relation（以后端策略为准） |
| 非 ENUM/ENUM_LABEL | 不得传两类关联字段 | 不可改 `format` | 常规删除 |

---

## 5. 接口契约摘要（设计口径）

## 5.1 AddFieldInput（摘要）

- `format=ENUM`
  - 必填：`relateEnumName`
  - 禁止：`enumRelationId`
- `format=ENUM_LABEL`
  - 必填：`enumRelationId`
  - 禁止：`relateEnumName`

## 5.2 UpdateFieldInput（摘要）

- 允许：`title` / `description` / `validationConfig`
- 禁止：`format`、`relateEnumName`、`enumRelationId` 变更

## 5.3 FieldEnumRelation（摘要）

- 关键字段：`id, modelId, sourceFieldName, enumName, orgName, projectSlug`
- 唯一约束：`UNIQUE(modelId, sourceFieldName)`

---

## 6. 错误码基线（简化版）

- `InvalidInput`（模型上下文参数缺失/非法）
- `InvalidInput`（字段参数缺失/非法，含 ENUM/ENUM_LABEL 必填缺失）
- `FIELD_ENUM_SOURCE_CONFLICT`（同一 source 重复创建 label relation）
- `FIELD_FORMAT_IMMUTABLE`（format 不可修改）
- `FIELD_REFERENCE_IN_USE`（删除被引用 source 阻断）

---

## 7. 验收总则（Given / When / Then）

### AC-01 ENUM 创建成功
Given 项目存在枚举 `CustomerLevel`  
When 创建字段，`format=ENUM` 且传 `relateEnumName=CustomerLevel`  
Then 创建成功，字段不携带 `enumRelationId`

### AC-02 ENUM 创建失败（缺失绑定）
Given 创建字段  
When `format=ENUM` 且未传 `relateEnumName`  
Then 返回 `InvalidInput`

### AC-03 ENUM_LABEL 创建失败（缺失 relation）
Given 模型存在 source ENUM 字段  
When 创建字段，`format=ENUM_LABEL` 且未传 `enumRelationId`  
Then 返回 `InvalidInput`

### AC-04 source 唯一冲突
Given 已存在 relation 绑定 `sourceFieldName=level`  
When 再次为该 source 创建 relation  
Then 返回 `FIELD_ENUM_SOURCE_CONFLICT`

### AC-05 format 不可修改
Given 字段已创建  
When 更新请求尝试修改 `format` 或关联语义字段  
Then 返回 `FIELD_FORMAT_IMMUTABLE`

### AC-06 删除被引用 source 阻断
Given source ENUM 字段被 relation 引用  
When 删除该 source 字段  
Then 返回 `FIELD_REFERENCE_IN_USE`

---

## 8. 数据关系说明（示例）

- A（`ENUM`）
  - 只保存 `relateEnumName` / `enumName`
  - 不写 relation
- A_label（`ENUM_LABEL`）
  - 通过 `enumRelationId -> FieldEnumRelation`
  - relation 再关联 `sourceFieldName=A` 与 `enumName=EnumA`

---

## 9. 子页索引

| 文档 | 路径 | 说明 |
|---|---|---|
| 01 | `./01-field-create-enum-binding.md` | 创建 ENUM 字段交互与校验 |
| 02 | `./02-field-edit-format-immutable.md` | 字段编辑页（format 只读） |
| 03 | `./03-backend-design.md` | 后端详细方案（无迁移 + BDD + 简化错误码） |
| 04 | `./04-frontend-subpage-design.md` | 前端交互合并设计（基于 01/02） |
| Domain | `./model-enum-domain.puml` | Field / Enum / FieldEnumRelation 领域模型 |