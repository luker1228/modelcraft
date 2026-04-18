# EndUserRef 字段 Format

> 依赖：无
> 对应主 PRD 章节：M1

---

## 背景

RLS 行级数据隔离依赖一种新的字段 Format：**`EndUserRef`**。该字段是整个 RLS 体系的基础，所有后续功能（自动生成 owner、WHERE 注入）都建立在此字段类型之上。

字段即开关：**有 `EndUserRef` 字段 = RLS 开启；无 = RLS 关闭**，不存在单独的 RLS toggle。

---

## 需求

### 字段定义

- 新增字段格式类型 `END_USER_REF`（扩展现有 `FieldFormat` 枚举）
- 字段名固定为 `owner`，不可自定义命名
- 存储语义：UUID 字符串，指向当前项目 `private_{projectSlug}.users.id`
- 数据库层生成外键约束：`REFERENCES private_{projectSlug}.users(id)`

### 唯一性约束

- 同一个 Model 最多只允许存在一个 `EndUserRef` 字段（第一期）
- 尝试添加第二个时，返回错误 `END_USER_REF_ALREADY_EXISTS`，提示"每个 Model 只允许一个归属字段"

### owner 字段的可见性规则

| 调用者 | Query 返回结果 | Mutation input 类型 |
|--------|--------------|-------------------|
| **EndUser** | 包含 `owner` 字段（值为自己的 ID，只读） | **不暴露** `owner` 字段（系统自动填充） |
| **Developer** | 包含 `owner` 字段，完整暴露 | 暴露 `owner` 字段，无限制 |

实现对应领域模型中的 `OwnerFieldSchema`：
- `visibleInQueryOutput = true`
- `visibleInMutationInput = false`（EndUser 视角）

---

## 验收标准

### AC-1：EndUserRef 字段约束

- [ ] 新建 `EndUserRef` 字段，DB 层生成对应外键约束（`REFERENCES private_{projectSlug}.users(id)`）
- [ ] 同一 Model 尝试添加第二个 `EndUserRef` 字段 → 报错提示"每个 Model 只允许一个归属字段"

### AC-7：owner 字段可见性

- [ ] EndUser 调用 Query → 返回结果中包含 `owner` 字段（值为自己的 ID）
- [ ] EndUser 调用 CreateOne / UpdateOne → input 类型中**不暴露** `owner` 字段
- [ ] Developer 调用 Query / Mutation → `owner` 字段完整暴露，行为无变化

---

## 领域模型关键元素

```
FieldDefinition
  + format: FieldFormat (新增 END_USER_REF)
  + isEndUserRef(): Boolean
  - 不变量：isEndUserRef() == true 时 name 必须为 "owner"

FieldFormat (Enum)
  新增：END_USER_REF

FieldError (Enum)
  新增：END_USER_REF_ALREADY_EXISTS

OwnerFieldSchema (Value Object)
  + visibleInQueryOutput: Boolean = true
  + visibleInMutationInput: Boolean = false
```

---

## 不做什么（本子页 Out of scope）

- 自动生成 owner 字段（见 `02-model-owner-lifecycle.md`）
- Runtime WHERE 注入逻辑（见 `04-runtime-rls-injection.md`）
- 多个 EndUserRef 字段支持（后续扩展）
