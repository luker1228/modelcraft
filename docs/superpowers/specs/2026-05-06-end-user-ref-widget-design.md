# END_USER_REF Widget 设计文档

**日期**：2026-05-06  
**状态**：已确认  
**范围**：JSON Schema 契约扩展 + 前端 widget 注册 + 后端安全边界

---

## 背景

`owner` 字段是所有模型的系统字段，格式为 `END_USER_REF`，用于标识记录所属的 EndUser（RLS 数据所有权）。

当前问题：
- `END_USER_REF` 格式在 JSON Schema 生成时没有 `x-mc.widget`，前端退化为普通文本框
- EndUser 端插入记录时需要手动填写 UUID，体验差且存在安全风险
- Tenant 管理端需要通过下拉选择器指定 EndUser，目前没有对应控件

---

## 设计方案

采用**方案 B：单 Widget + 前端 workspaceMode 分支**。

- 后端始终输出 `x-mc.widget = "end-user-ref"`，schema 生成器保持幂等，无需感知调用方身份
- 前端根据 `workspaceMode`（`end_user` / `design`）决定渲染行为
- 符合现有 `workspace-mode-boundary.md` 的边界约定

---

## §1 后端 JSON Schema 变更

### 1.1 `decideWidget()` 新增分支

**文件**：`modelcraft-backend/internal/domain/modeldesign/jsonschema_generator.go`

在 `decideWidget()` 的 `switch field.Type.Format` 中新增：

```go
case FormatEndUserRef:
    return "end-user-ref"
```

### 1.2 生成结果示例

```json
"owner": {
  "type": "string",
  "title": "Owner",
  "description": "System Field",
  "x-mc": {
    "format": "END_USER_REF",
    "widget": "end-user-ref",
    "isPrimary": false,
    "isUnique": false,
    "nullable": false,
    "displayOrder": ""
  }
}
```

### 1.3 契约文档更新

`ai-metadata/backend/design/domain-model/8-runtime/jsonschema-contract.md` 的 widget 表格追加：

| `x-mc.widget` 值 | 触发条件 | 前端控件 |
|---|---|---|
| `"end-user-ref"` | `format = END_USER_REF` | EndUser 端：隐藏；Tenant 管理端：EndUser 下拉选择器 |

---

## §2 前端变更

### 2.1 类型扩展

**文件**：`modelcraft-front/src/types/xmc.ts`

```ts
export type XMCWidget =
  | 'enum-select'
  | 'date'
  | 'datetime-local'
  | 'time'
  | 'textarea'
  | 'relation-selector'
  | 'relation-multi-readonly'
  | 'end-user-ref'   // 新增
```

### 2.2 UI Schema 构建分支

**文件**：`modelcraft-front/src/web/components/.../build-ui-schema.ts`

```ts
case 'end-user-ref':
  if (workspaceMode === 'end_user') {
    // EndUser 端：完全隐藏，不渲染表单项，不进 payload
    return { 'ui:widget': 'hidden' }
  } else {
    // Tenant 管理端：渲染 EndUser 下拉选择器
    return { 'ui:widget': 'EndUserSelectorWidget' }
  }
```

### 2.3 Payload 过滤

- **EndUser 端**：`owner` 字段 `ui:widget = hidden`，不出现在表单值，不进 payload。无需额外改动 `sanitizeMutationInputData`。
- **Tenant 管理端**：`owner` 字段由 `EndUserSelectorWidget` 填入 UUID，正常进入 payload。

### 2.4 `EndUserSelectorWidget` 组件（新增）

- 标准 RJSF custom widget
- 调用现有 `endUsers` / `endUsersByProject` GraphQL query
- 渲染 `<Select>`，value 为 EndUser UUID，label 为 EndUser email 或 displayName
- 注册到 RJSF widgets 映射表

---

## §3 后端安全边界

### EndUser 端（Runtime API）

后端 runtime insert handler 强制从 JWT 读取当前 EndUser ID 并写入 `owner`，**覆盖** payload 中任何传入值：

```go
insertData["owner"] = ctx.EndUserID()  // 从 JWT 读取，防篡改
```

### Tenant 管理端（Admin API）

直接使用 payload 传入的 `owner` 值，但需校验该 UUID 是当前 Project 下有效的 EndUser ID。

### 两端对比

| 端 | `owner` 来源 | 后端行为 |
|---|---|---|
| EndUser 端 | 前端不传（hidden） | 从 JWT 自动注入，覆盖任何传入值 |
| Tenant 管理端 | 前端下拉选择传入 | 直接使用，校验是否为有效 EndUser |

---

## 改动文件清单

| 文件 | 类型 | 说明 |
|---|---|---|
| `modelcraft-backend/internal/domain/modeldesign/jsonschema_generator.go` | 修改 | `decideWidget()` 新增 `FormatEndUserRef` 分支 |
| `modelcraft-backend/internal/domain/modelruntime/...` | 修改 | insert handler 注入 `owner` from JWT |
| `modelcraft-front/src/types/xmc.ts` | 修改 | `XMCWidget` 新增 `"end-user-ref"` |
| `modelcraft-front/src/web/components/.../build-ui-schema.ts` | 修改 | 新增 `end-user-ref` case，按 workspaceMode 分支 |
| `modelcraft-front/src/web/components/EndUserSelectorWidget.tsx` | 新增 | RJSF custom widget，EndUser 下拉选择 |
| `ai-metadata/backend/design/domain-model/8-runtime/jsonschema-contract.md` | 修改 | widget 表格追加 `end-user-ref` 行 |

---

## 不在范围内

- 表格列展示 owner 字段：当前保持展示原始 UUID，后续迭代
- EndUser 端 UPDATE 时 owner 字段的保护逻辑（随安全迭代补充）
