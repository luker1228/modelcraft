# Runtime JSON Schema 契约

> 本文件定义 ModelCraft 运行时 JSON Schema 的完整字段规范。
> JSON Schema 是前端表单渲染的**唯一数据源**，前端不依赖 `Field[]` 等设计时数据结构。

---

## 设计原则

1. **顶层只放标准字段** — `type`、`format`、`title`、`required`、校验规则等均遵循 JSON Schema Draft 7 标准
2. **私有扩展收敛到 `x-mc`** — 所有 ModelCraft 专有元数据放在单一命名空间对象 `x-mc` 下，不污染顶层
3. **`x-mc.widget` 直接表达渲染意图** — 后端决定用哪个控件，前端零推断逻辑，不依赖 `Field[]`

### 关于 `format` 字段

JSON Schema 标准 `format` 是**注解性**的，校验器默认不强制执行。顶层 `format` 只用于标准值（`uuid`、`date`、`date-time`、`time`），供工具链识别。

ModelCraft 内部 format 枚举值（`ENUM`、`STRING`、`INTEGER` 等）**不放入顶层 `format`**——渲染意图通过 `x-mc.widget` 表达。自定义 `format` 值虽然规范允许，但 `x-mc.widget` 已经覆盖了所有需求，无需重复。

---

## 模型级字段

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "title": "User",
  "description": "用户模型",
  "required": ["name"],
  "properties": { ... },
  "x-modelName": "User",
  "x-databaseName": "users_db"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `$schema` | string | JSON Schema Draft 7 标识 |
| `type` | `"object"` | 固定 |
| `title` / `description` | string | 模型显示名 / 描述 |
| `required` | string[] | 必填字段名列表（`[]` 而非 `null`） |
| `x-modelName` | string | 模型名（用于构建 runtime endpoint） |
| `x-databaseName` | string | 所属数据库名（用于构建 runtime endpoint） |

---

## 字段级字段

### 标准字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `title` / `description` | string | 字段显示名 / 描述 |
| `type` | string | JSON Schema 基础类型：`string` / `number` / `integer` / `boolean` / `array` / `object` |
| `format` | string | 标准格式注解：`uuid` / `date` / `date-time` / `time`（仅这四种） |
| `enum` | string[] | ENUM 字段的合法 code 列表（供 JSON Schema 校验器使用） |
| `nullable` | bool | 字段是否可为 null（`NonNull=false` 时为 `true`） |
| `readOnly` | bool | 主键字段和 RELATION 虚拟字段标记为 `true` |
| `maxLength` / `minLength` / `pattern` | - | 字符串校验规则 |
| `maximum` / `minimum` | - | 数字校验规则 |
| `maxItems` / `minItems` | - | 数组校验规则 |

### `x-mc` 扩展对象

所有 ModelCraft 专有元数据统一放在 `x-mc` 对象下：

```json
"user_id": {
  "type": "string",
  "title": "用户 ID",
  "x-mc": {
    "widget": "relation-selector",
    "isPrimary": false,
    "isUnique": false,
    "displayOrder": "a0",
    "storageHint": "TEXT",
    "belongsToFkId": "fk-123",
    "relation": {
      "databaseName": "users_db",
      "modelName": "User"
    },
    "enum": {
      "name": "StatusEnum",
      "displayName": "状态",
      "isMultiSelect": false,
      "options": [
        { "code": "active", "label": "激活", "description": "" }
      ]
    }
  }
}
```

| `x-mc` 字段 | 类型 | 出现条件 | 说明 |
|---|---|---|---|
| `widget` | string | 需要非默认控件时 | 渲染控件指令（见下表） |
| `isPrimary` | bool | 始终 | 是否主键 |
| `isUnique` | bool | 始终 | 是否唯一约束 |
| `displayOrder` | string | 始终 | 字典序排序键（用于前端字段排序） |
| `storageHint` | string | 有值时 | 当前仅 `"TEXT"`（→ textarea） |
| `validateRule` | string | 有值时 | `"email"` / `"url"` / `"phone"` |
| `precision` / `scale` | int | DECIMAL 字段 | 精度和小数位数 |
| `minDate` / `maxDate` | string | DATE 字段有校验时 | ISO 8601 `YYYY-MM-DD` |
| `minTime` / `maxTime` | string | TIME 字段有校验时 | `HH:MM:SS` |
| `belongsToFkId` | string | 外键列字段 | 该字段引用的逻辑外键 ID |
| `relation` | object | 外键列字段（`belongsToFkId` 存在时） | `{ databaseName, modelName }`，前端据此构建 runtime endpoint |
| `relateFkId` | string | RELATION 虚拟字段 | 关联的逻辑外键 ID |
| `enum` | object | ENUM / ENUM_ARRAY 字段 | 枚举完整元数据，含 options |

---

## `x-mc.widget` 取值规范

后端直接告诉前端渲染意图，前端 `buildUiSchema` 做纯映射，零推断：

| `x-mc.widget` 值 | 触发条件（后端逻辑） | 前端控件 |
|---|---|---|
| `"enum-select"` | `format = ENUM` 或 `ENUM_ARRAY` | `EnumSchemaSelect` |
| `"date"` | `format = DATE` | 原生 `date` input |
| `"datetime-local"` | `format = DATETIME` | 原生 `datetime-local` input |
| `"time"` | `format = TIME` | 原生 `time` input |
| `"textarea"` | `storageHint = TEXT` | `textarea` |
| `"relation-selector"` | `BelongsToFKID != nil`（外键列） | `RelationSelector` |
| 不填 | 其余所有字段 | RJSF 按标准 `type` 默认渲染 |

> `widget` 缺省时前端使用 RJSF 默认控件（string → text input，number → number input，boolean → checkbox 等）。

---

## 完整示例

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "title": "订单",
  "description": "",
  "required": ["user_id", "status"],
  "x-modelName": "Order",
  "x-databaseName": "shop_db",
  "properties": {
    "id": {
      "type": "string",
      "format": "uuid",
      "title": "ID",
      "readOnly": true,
      "x-mc": {
        "isPrimary": true,
        "isUnique": true,
        "displayOrder": "a0"
      }
    },
    "user_id": {
      "type": "string",
      "title": "用户",
      "x-mc": {
        "widget": "relation-selector",
        "isPrimary": false,
        "isUnique": false,
        "displayOrder": "a1",
        "belongsToFkId": "fk-abc",
        "relation": {
          "databaseName": "users_db",
          "modelName": "User"
        }
      }
    },
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
          "name": "OrderStatus",
          "displayName": "订单状态",
          "isMultiSelect": false,
          "options": [
            { "code": "pending", "label": "待支付", "description": "" },
            { "code": "paid",    "label": "已支付", "description": "" },
            { "code": "cancelled", "label": "已取消", "description": "" }
          ]
        }
      }
    },
    "note": {
      "type": "string",
      "title": "备注",
      "x-mc": {
        "widget": "textarea",
        "isPrimary": false,
        "isUnique": false,
        "displayOrder": "a3",
        "storageHint": "TEXT"
      }
    },
    "created_at": {
      "type": "string",
      "format": "date-time",
      "title": "创建时间",
      "x-mc": {
        "widget": "datetime-local",
        "isPrimary": false,
        "isUnique": false,
        "displayOrder": "a4"
      }
    }
  }
}
```

---

## 迁移说明

当前后端仍使用平铺的 `x-storageHint`、`x-isPrimary` 等字段（旧格式），需迁移到 `x-mc` 结构。迁移完成后前端可移除对 `Field[]` prop 的依赖。
