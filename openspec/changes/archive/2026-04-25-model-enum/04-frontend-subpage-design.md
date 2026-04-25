# 04 - 前端交互合并设计（基于 01/02）

> 本文为前端统一交互设计稿，合并以下子页内容：
> - `01-field-create-enum-binding.md`
> - `02-field-edit-format-immutable.md`
> - ENUM_LABEL 新增子页（relation 模式）

---

## 1. 页面清单定位（合并版）

| # | 页面名称 | 页面类型 | 路由路径 | 主体实体 | 核心字段 | 关键操作 |
|---|---|---|---|---|---|---|
| 1 | 字段创建页（ENUM 绑定） | 创建页 | 沿用现有字段创建路由 | Field | name, title, format, relateEnumName | 取消、保存 |
| 2 | 字段创建页（ENUM_LABEL 绑定 relation） | 创建页 | 沿用现有字段创建路由 | Field + FieldEnumRelation | name, title, format, sourceField, enumRelationId | 取消、保存 |
| 3 | 字段编辑页（Format 不可变） | 编辑页 | 沿用现有字段编辑路由 | Field | name(只读), format(只读), relation(只读), title/description | 取消、保存 |

---

## 2. 全局交互规则

1. `format` 在创建阶段可设，进入编辑后不可修改。
2. `ENUM` 字段仅使用 `relateEnumName`，不显示/不提交 `enumRelationId`。
3. `ENUM_LABEL` 字段必须使用 `enumRelationId`，且需先确定 `sourceField`（必须是 ENUM）。
4. 同一 `sourceField` 只能生成一个 label（前端预校验 + 后端唯一约束兜底）。
5. 保存动作统一：前端校验通过后再调用接口，失败时保留输入态并展示明确错误。

---

## 3. 子页一：字段创建页（ENUM 绑定）

## 3.1 ASCII 布局

```text
┌──────────────────────────────────────────────────────────────┐
│ ▸ 项目 / 模型 / 字段 / 新建                                  │
│ // ── 基本信息 ──────────────────────────────────────────── │
│ 字段名称: [____________________]   ← Field.name             │
│ 标题:     [____________________]   ← Field.title            │
│ 描述:     [____________________]   ← Field.description      │
│ 字段类型: «ENUM»（创建时可选）                               │
│                                                              │
│ // ── 枚举绑定 ──────────────────────────────────────────── │
│ 关联枚举: (请选择枚举 ▾)               ← relateEnumName      │
│ · 无可用枚举: ░░░ 请先创建枚举 ░░░                           │
│                                                              │
│ // ── 操作区 ────────────────────────────────────────────── │
│ [取消]                                         [保存]        │
└──────────────────────────────────────────────────────────────┘
```

## 3.2 表单字段与校验

- 必填：`name`、`title`、`format=ENUM`、`relateEnumName`
- 禁止：提交 `enumRelationId`
- 保存前校验：
  - 未选枚举 -> 阻断提交并提示“枚举字段必须关联枚举”

## 3.3 交互细节

- 枚举下拉仅展示当前项目可用枚举
- 枚举为空时允许填写基础字段，但保存按钮禁用或保存时报错阻断

## 3.4 API 编排

1. 查询枚举列表
2. `addFields` 提交 `format=ENUM + relateEnumName`
3. 成功后更新字段列表缓存

## 3.5 错误态

- `InvalidInput`（参数缺失/非法，含 relateEnumName 缺失、枚举不存在、归属不匹配）

---

## 4. 子页二：字段创建页（ENUM_LABEL + relation）

## 4.1 ASCII 布局

```text
┌────────────────────────────────────────────────────────────────────┐
│ ▸ 项目 / 模型 / 字段 / 新建                                        │
│ // ── 基本信息 ────────────────────────────────────────────────── │
│ 字段名称: [____________________]    ← Field.name                  │
│ 标题:     [____________________]    ← Field.title                 │
│ 描述:     [____________________]    ← Field.description           │
│ 字段类型: «ENUM_LABEL»（创建时可选）                              │
│                                                                    │
│ // ── Label 关联配置 ─────────────────────────────────────────── │
│ 源字段:   (请选择 ENUM 字段 ▾)         ← sourceField              │
│ Relation: (选择已有 relation ▾)         ← enumRelationId          │
│          [新建 relation]                                           │
│                                                                    │
│ 新建 relation 抽屉（可选）                                         │
│   sourceFieldName: 自动带入 sourceField（只读）                    │
│   enumName:        自动带入 sourceField.enumName（只读/校验）      │
│                                                                    │
│ 提示: 同一 sourceField 只能生成一个 label                           │
│                                                                    │
│ // ── 操作区 ──────────────────────────────────────────────────── │
│ [取消]                                               [保存]        │
└────────────────────────────────────────────────────────────────────┘
```

## 4.2 表单字段与校验

- 必填：`name`、`title`、`format=ENUM_LABEL`、`sourceField`、`enumRelationId`
- 约束：
  - `sourceField` 必须是 ENUM 字段
  - 同一 source 不可重复创建 label
  - `enumRelationId` 必须存在且与当前模型匹配

## 4.3 交互细节

- 若模型内无 ENUM 字段：展示空态“请先创建 ENUM 字段”
- relation 选择器支持两种路径：
  1) 选已有
  2) 先创建 relation，再回填 `enumRelationId`
- 若 source 已有关联 label：在 source 下拉项上标记“已占用”并禁用

## 4.4 API 编排

1. 查询模型字段（过滤出 ENUM）
2. 查询 relation 列表（按 model/source）
3. 可选：创建 relation
4. `addFields` 提交 `format=ENUM_LABEL + enumRelationId`

## 4.5 错误态

- `InvalidInput`（参数缺失/非法，含 enumRelationId 缺失、relation/source 不合法）
- `FIELD_ENUM_SOURCE_CONFLICT`

---

## 5. 子页三：字段编辑页（Format 不可变）

## 5.1 ASCII 布局

```text
┌──────────────────────────────────────────────────────────────┐
│ ▸ 项目 / 模型 / 字段 / 编辑                                  │
│ // ── 基本信息（可编辑）──────────────────────────────────── │
│ 字段名称: «user_status»（只读）                               │
│ 标题:     [____________________]                              │
│ 描述:     [____________________]                              │
│                                                              │
│ // ── 格式信息（只读）────────────────────────────────────── │
│ 字段类型: «ENUM / ENUM_LABEL / ...»                          │
│ ENUM:       关联枚举: «relateEnumName»                        │
│ ENUM_LABEL: relationId: «enumRelationId»                      │
│ 提示: 字段格式创建后不可修改                                 │
│                                                              │
│ // ── 操作区 ────────────────────────────────────────────── │
│ [取消]                                         [保存]        │
└──────────────────────────────────────────────────────────────┘
```

## 5.2 编辑权限

- 只读：`name`、`format`、`relateEnumName`、`enumRelationId`
- 可编辑：`title`、`description`、`validationConfig`

## 5.3 API 编排

- 调用 `updateField`
- 不提交 `format` 及关联字段变更

## 5.4 错误态

- `FIELD_FORMAT_IMMUTABLE`
- 通用保存失败兜底提示

---

## 6. 前端状态与组件拆分（统一）

```text
app/.../model-editor/
├── _components/field-pages/
│   ├── CreateEnumFieldPage.tsx
│   ├── CreateEnumLabelFieldPage.tsx
│   ├── EnumRelationSelector.tsx
│   ├── FieldDetailEditPage.tsx
│   └── index.ts
└── _hooks/
    ├── use-create-enum-field.ts
    ├── use-create-enum-label-field.ts
    ├── use-field-detail-edit.ts
    ├── use-enum-relation-options.ts
    └── types.ts
```

状态建议：
- 表单：RHF + schema 校验
- 数据：GraphQL query/mutation + cache update
- UI：`idle/loading/saving/error` 统一状态机

---

## 7. Contract 同步点（执行清单）

1. 后端更新 `project schema`（field + field_enum_relation）
2. 前端通过 subtree 同步 `contract/`
3. 前端执行 codegen
4. 更新前端 gql 操作与 hook 签名
5. 回归三类页面（创建 ENUM、创建 ENUM_LABEL、编辑）

---

## 8. 合并验收标准（总）

1. 创建 ENUM：必须 `relateEnumName`，且无 `enumRelationId`
2. 创建 ENUM_LABEL：必须 `enumRelationId`
3. source 唯一：同一 source 不能重复 label
4. 编辑页中 format 永远只读
5. 所有失败分支均有可识别提示
