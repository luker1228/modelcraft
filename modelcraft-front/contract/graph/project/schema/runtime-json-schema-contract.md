# Runtime JSON Schema 协议（Project 域）

本文档定义 `model-editor` 使用的 Runtime JSON Schema 协议。

目标：前端仅依赖 schema 渲染，不再从设计态 `Field[]` 推导行为。

## 1. 协议原则

1. Schema 是唯一渲染数据源。
2. ModelCraft 私有元数据统一放在 `x-mc` 下。
3. 后端通过 `x-mc.widget` 显式声明渲染意图。
4. 枚举与关系元数据必须完整到可被前端直接渲染。
5. 枚举值集合必须生成到标准 JSON Schema 字段 `enum`，并作为唯一枚举值来源。

## 2. 模型级字段

| 字段 | 类型 | 必填 | 用途 |
| --- | --- | --- | --- |
| `$schema` | string | 是 | JSON Schema 草案标识，当前为 draft-07 |
| `type` | `"object"` | 是 | 根类型 |
| `title` | string | 是 | 模型展示名 |
| `description` | string | 否 | 模型描述 |
| `required` | string[] | 是 | 必填字段名列表（无必填时使用 `[]`） |
| `properties` | object | 是 | 字段 schema 映射 |
| `x-modelName` | string | 是 | Runtime 模型名 |
| `x-databaseName` | string | 是 | Runtime 数据库名 |

## 3. 属性标准字段

| 字段 | 类型 | 必填 | 用途 |
| --- | --- | --- | --- |
| `title` | string | 是 | 字段展示名 |
| `description` | string | 否 | 字段描述 |
| `type` | string | 是 | JSON Schema 基础类型 |
| `format` | string | 否 | 仅允许标准格式（`uuid`/`date`/`date-time`/`time`） |
| `enum` | string[] | 条件必填 | 枚举 code 列表（用于校验） |
| `readOnly` | boolean | 条件必填 | 不可编辑字段标记（`id`、关系只读字段） |
| `nullable` | boolean | 否 | 可空标记 |

## 4. `x-mc` 扩展字段

| 字段 | 类型 | 必填 | 用途 |
| --- | --- | --- | --- |
| `widget` | string | 条件必填 | 前端控件选择器 |
| `isPrimary` | boolean | 是 | 主键标记 |
| `isUnique` | boolean | 是 | 唯一约束标记 |
| `displayOrder` | string | 是 | 字段排序键 |
| `storageHint` | string | 否 | 渲染提示（`TEXT` => textarea） |
| `validateRule` | string | 否 | 领域校验提示（`email`/`url`/`phone`） |
| `precision` | number | 否 | Decimal 精度 |
| `scale` | number | 否 | Decimal 小数位 |
| `relation` | object | 条件必填 | 关系元数据对象（见 4.2） |
| `enum` | object | 条件必填 | 枚举元数据对象 |

### 4.1 `x-mc.widget` 取值

| 取值 | 含义 |
| --- | --- |
| `enum-select` | 枚举选择控件 |
| `date` | 日期输入 |
| `datetime-local` | 日期时间输入 |
| `time` | 时间输入 |
| `textarea` | 多行文本输入 |
| `relation-selector` | 多对一关系选择器 |
| `relation-multi-readonly` | 一对多只读列表 |

### 4.2 `x-mc.relation` 对象

```json
{
  "databaseName": "users_db",
  "modelName": "User",
  "belongsToFkId": "fk-123",
  "relateFkId": "",
  "relationType": "MANY_TO_ONE",
  "relationDirection": "normal"
}
```

字段说明：

1. `databaseName`：目标 Runtime 数据库名。
2. `modelName`：目标 Runtime 模型名。
3. `belongsToFkId`：多对一外键字段 ID（多对一字段时必填）。
4. `relateFkId`：关系虚拟字段外键 ID（RELATION 字段时必填）。
5. `relationType`：`ONE_TO_MANY` 或 `MANY_TO_ONE`。
6. `relationDirection`：`reverse` 或 `normal`。

### 4.3 `x-mc.enum` 对象

```json
{
  "labelFieldName": "status_label"
}
```

规则：

1. `labelFieldName` 表示枚举标签投影字段名，属于 `x-mc.enum` 的唯一字段。
2. 后端必须将枚举定义生成到属性顶层标准字段 `enum`（JSON Schema Draft 7）。
3. 枚举取值列表仅通过属性顶层 `enum` 提供，前端也仅以该字段作为枚举值来源。
4. `x-mc.enum` 不承载枚举值集合，仅承载 `labelFieldName`。
5. 协议不提供旧字段兼容兜底；缺失任一必需字段视为协议违规。

## 5. 关系元数据必备项

关系控件字段必须包含：

- `x-mc.relation.databaseName`
- `x-mc.relation.modelName`
- `x-mc.relation.belongsToFkId`（多对一字段）或 `x-mc.relation.relateFkId`（关系虚拟字段）
- `x-mc.relation.relationType`
- `x-mc.relation.relationDirection`

缺少上述字段时，前端无法可靠查询关系目标。

## 6. 兼容性策略

以下属于破坏性变更：

1. 重命名或删除已有 `x-mc` 字段。
2. 修改已有 `x-mc.widget` 枚举值的语义。
3. 删除前端依赖的关系/枚举元数据。

发生破坏性变更时：

1. 必须先更新本协议文档。
2. 必须在同一 PR 提供迁移说明。
3. 必须在同一发布窗口协调前后端改造。

## 7. 协作流程

1. 在 `api/` 中先更新协议/schema。
2. 后端按更新后的协议实现。
3. 推送 `api/` subtree：
   - `git subtree push --prefix=api contracts main`
4. 前端通过 `front-contract-pull` skill 同步协议。
5. 前端实现并做回归验证。

## 8. 属性示例

```json
"status": {
  "type": "string",
  "title": "状态",
  "enum": ["pending", "paid", "cancelled"],
  "x-mc": {
    "widget": "enum-select",
    "isPrimary": false,
    "isUnique": false,
    "displayOrder": "a2",
    "enum": {
      "labelFieldName": "status_label"
    }
  }
}
```
