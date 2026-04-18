# 新建 Model 自动生成 owner 字段

> 依赖：`01-enduserref-field.md`
> 对应主 PRD 章节：M2、M5

---

## 背景

开发者用 ModelCraft 新建 Model 时，应**零配置开箱即用**数据隔离。系统在 Model 创建时自动添加 `owner` 字段，无需开发者手动操作。

相对地，从已有数据库表**导入**的 Model 不应被强制注入 owner 字段，行为保持与导入前完全一致（向后兼容）。

---

## 需求

### M2：新建 Model 自动生成 owner

- 新建 Model（`createdVia = NEW`）时，系统自动向字段列表中添加名为 `owner` 的 `EndUserRef` 字段
- 该字段由开发者**可以删除**（不是锁定字段）
- 删除即关闭 RLS，无单独的 RLS 开关

### M5：导入 Model 不默认开启 RLS

- 从已有数据库表导入的 Model（`createdVia = IMPORTED`）不自动添加 `owner` 字段
- 当前版本**不支持**为导入 Model 手动配置 RLS（后续扩展）
- 导入后 Runtime 行为与导入前完全一致（无 WHERE 注入）

### 删除 owner 字段的交互

- 开发者删除 `EndUserRef` 字段时，前端展示**二次确认弹窗**
- 弹窗提示文案：「删除后数据隔离将关闭，该 Model 的所有终端用户将可访问全量数据。」
- 对应后端行为：`Model.removeField()` 返回 `RemoveFieldResult { warning: RLS_WILL_DISABLE }`

---

## 验收标准

### AC-2：新建 Model 自动生成 owner

- [ ] 新建 Model 后，字段列表中自动存在名为 `owner` 的 `EndUserRef` 字段
- [ ] 导入 Model 后，字段列表中**无** `owner` 字段
- [ ] 删除 `owner` 字段时，弹出二次确认弹窗，提示"删除后数据隔离将关闭"

### AC-8：向后兼容

- [ ] 升级后，现有无 EndUserRef 字段的 Model，Runtime 行为完全不变
- [ ] 现有 BDD / 集成测试全部通过，无需修改

---

## 用户故事对应

**Story 1**：新建 Model，零配置开箱即用
> 新建 `orders` Model 后，字段列表中自动存在 `owner` 字段。

**Story 2**：删除 owner 字段，关闭 RLS
> 删除 `EndUserRef` 字段后，该 Model 的 Runtime 请求不再有 WHERE 注入，所有 EndUser 可访问全量数据。

**Story 5**：从表导入的 Model 不强制 RLS
> 导入 Model 后，字段列表中无 `owner` 字段，Runtime 行为与导入前完全一致。

---

## 领域模型关键元素

```
Model (Aggregate Root)
  + createdVia: ModelCreationSource
  - 不变量：createdVia=NEW 时自动携带 owner 字段
  - 不变量：createdVia=IMPORTED 时不生成 owner 字段
  + isRLSEnabled(): Boolean
    → 存在 format=END_USER_REF 的字段 → true
    → 不存在 → false（开放模式）
  + removeField(name: String): RemoveFieldResult

ModelCreationSource (Enum)
  NEW      ← 自动添加 owner
  IMPORTED ← 不添加 owner

RemoveFieldWarning (Enum)
  RLS_WILL_DISABLE ← 前端展示二次确认弹窗
```

---

## 不做什么（本子页 Out of scope）

- Runtime WHERE 注入逻辑（见 `04-runtime-rls-injection.md`）
- JWT 认证机制（见 `03-runtime-jwt-auth.md`）
- 为导入 Model 手动添加 EndUserRef 支持（后续扩展）
