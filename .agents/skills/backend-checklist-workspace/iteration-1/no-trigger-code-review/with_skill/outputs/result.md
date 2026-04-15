# Code Review: internal/app/modeldesign/field_app.go

## 前置说明：文件不存在

**`internal/app/modeldesign/field_app.go` 不存在。**

通过 glob 搜索，`internal/app/modeldesign/` 目录下字段相关的 app 层代码实际分布在：

- `field_builder.go` — 仅包含 `package modeldesign` 声明，内容为空
- `model_app.go` — 包含所有字段相关 app 层方法（AddFieldSync、UpdateFieldSync、RemoveFieldSync 等）

本 review 对象调整为 **`internal/app/modeldesign/model_app.go`** 中与字段操作相关的方法（约 700 行），这是最接近用户意图的目标文件。

---

## 是否触发了 backend-checklist skill

**未触发 backend-checklist skill。**

理由：
- 用户的触发词是「review 代码质量，看有没有逻辑错误和错误处理漏洞」
- backend-checklist skill 的触发词明确限定为：「加入错题本」、「记录这个错误」、「用错题本 check」、「checklist review」、「有没有历史 bug」、「check 历史 bug」
- 本次任务是**通用代码质量 review**，不是 checklist 驱动的历史 bug 模式比对
- skill 描述中"后端代码实现完成后，也应主动提示用户是否需要跑一遍 checklist review"——这是一个建议提示，而非自动触发

**不过**，作为 review 的补充，我已手动读取了 `ai-metadata/backend/common-mistakes.md`，并将其中的规则纳入了本次 review 的检查范围（见下文 BM-20260415-0001 规则检查）。

---

## 代码质量 Review

### 1. 逻辑错误

#### 1.1 `addPhysicFields` 中字段状态更新逻辑反转（HIGH）

**位置**: `model_app.go:494-501`

```go
updateReq := modeldesign.UpdateFieldsStatusRequest{
    ModelId: model.ID,
    Name:    fieldNames,
    Status:  modeldesign.FieldStatusToDelete,  // ← 部署成功后设置为"待删除"？
}
if err := s.modelRepo.UpdateFieldsStatus(ctx, updateReq); err != nil {
    return fmt.Errorf("更新字段状态失败: %w", err)
}
```

部署成功后，字段状态被设置为 `FieldStatusToDelete`，而非 `FieldStatusDeploySuccess`。这与 `addVirtualFields`（同文件 L661-668）和 `addRelationFields`（L717-723）的逻辑相反，后两者在保存后都设置了 `FieldStatusDeploySuccess`。

**推测**：这可能是意图预留一个"待确认部署成功"状态，但代码注释缺失，语义极度混乱。若 `FieldStatusToDelete` 字面意思是"待删除"，则部署成功的字段被错误标记为待删除状态，可能导致后续查询或清理逻辑误删。

**建议**：明确状态流转语义，若物理字段部署成功后状态应为可用，应与其他路径保持一致，使用 `FieldStatusDeploySuccess`。

---

#### 1.2 `SyncModelSchemaFromJSON` 中字段类型比较可能 panic（HIGH）

**位置**: `model_app.go:1371`

```go
if existingField.Type.Format != schemaField.Type.Format {
```

`existingField.Type` 和 `schemaField.Type` 均可能为 nil（`FieldDefinition.Type` 是指针类型）。代码没有 nil 检查，直接访问 `.Format` 会导致 nil pointer dereference panic。

**建议**：
```go
existingType := ""
if existingField.Type != nil {
    existingType = string(existingField.Type.Format)
}
schemaType := ""
if schemaField.Type != nil {
    schemaType = string(schemaField.Type.Format)
}
if existingType != schemaType {
    ...
}
```

---

#### 1.3 `collectFieldsToDelete` 系统字段硬编码（MEDIUM）

**位置**: `model_app.go:1437-1441`

```go
systemFieldNames := map[string]bool{
    "id":        true,
    "createdAt": true,
    "updatedAt": true,
}
```

系统字段列表硬编码在此处，而 `modeldesign.GetSystemFields()` 是正式定义（L151-163）。若系统字段后续扩展，此处不会自动同步，会导致系统字段被误删。

**建议**：
```go
systemFieldNames := make(map[string]bool)
for _, f := range modeldesign.GetSystemFields() {
    systemFieldNames[f.Name] = true
}
```

---

### 2. 错误处理漏洞

#### 2.1 `addPhysicFields` 中枚举关联创建不在事务中（HIGH）

**位置**: `model_app.go:503-525`

物理字段通过 `s.modelRepo.AddFields` 写入（有补偿回滚），但随后的枚举关联记录 `enumAssocRepo.Create` 在事务之外执行。若此处失败，字段已落库但枚举关联未建立，造成数据不一致（字段存在但查不到关联枚举）。

此处与 `removeEnumLabelFieldWithRelation` 形成对比——删除时有事务保护，创建时却无事务。

**建议**：将枚举关联的创建纳入与字段落库相同的事务，或至少在失败时执行补偿删除字段和关联。

---

#### 2.2 `RemoveFieldSync` 中删除物理字段时缺少事务保护（MEDIUM）

**位置**: `model_app.go:880-900`

删除物理字段的流程：
1. `UpdateFieldsStatus` → 标记为待删除（平台 DB）
2. `DeployModelToRemoveFields` → 执行 DROP COLUMN（客户 DB）
3. `DeleteFields` → 删除平台 DB 记录

若步骤 2 成功（客户 DB 列已删除）但步骤 3 失败，会出现客户 DB 无该列但平台 DB 仍有字段记录的不一致状态。与 `addPhysicFields` 中的补偿机制相比，删除路径没有类似保护。

**建议**：使用事务封装步骤 1 和 3，或在步骤 3 失败时记录告警并提供修复机制。

---

#### 2.3 `DeleteFieldEnumRelation` 中读取字段存在竞态（LOW）

**位置**: `model_app.go:1177-1188`

检查 label 字段是否引用该 relation（read），然后删除 relation（write）之间无锁保护。极端情况下，两个请求并发执行时可能同时通过检查并都完成删除，产生竞态。虽然实际业务中并发概率低，但若有数据库唯一约束支撑则可接受，否则建议在事务内执行 select-for-update。

---

#### 2.4 `addVirtualFields` 中 `fieldEnumRelRepo` 为 nil 时的判断时机过晚（LOW）

**位置**: `model_app.go:606-608`

```go
if s.fieldEnumRelRepo == nil {
    return bizerrors.NewErrorFromContext(ctx, bizerrors.SystemError, "field enum relation repository is not configured")
}
```

这个 nil 检查在循环体内，每次循环都检查。若有多个 ENUM_LABEL 字段，第一个字段通过了 `field.IsEnumLabelField()` 检查后才发现 repo 未注入，而不是在进入循环前就快速失败。建议在函数入口处统一检查依赖。

---

### 3. 错题本规则检查（BM-20260415-0001）

规则：**凡是查询带 `project_slug` 的 SQL，必须同时带 `org_name`。**

本次 review 的文件是 app 层，不直接写 SQL，数据访问均通过 Repository 接口。app 层在查询时通过 `ctxutils.GetOrgNameFromContext` 获取 orgName 并传递给 repo，检查了以下关键调用：

- `CreateModelSync` L106: 正确获取 orgName 并传入 `clusterRepo.GetByProjectKey`
- `AddFieldSync` L360: 正确获取 orgName
- `CreateFieldEnumRelation` L1065: 正确获取 orgName 并通过 model 对象透传

**结论**：app 层对 orgName 的传递是正确的，未发现跨租户隔离问题。具体 SQL 隔离需在 repository 层确认。

---

## 总结

| 编号 | 严重程度 | 类型 | 位置 | 描述 |
|------|----------|------|------|------|
| #1 | HIGH | 逻辑错误 | L494-501 | 物理字段部署成功后状态被错误设置为 FieldStatusToDelete |
| #2 | HIGH | 潜在 panic | L1371 | Type.Format 访问前未做 nil 检查 |
| #3 | HIGH | 错误处理 | L503-525 | 枚举关联创建不在事务中，可能导致数据不一致 |
| #4 | MEDIUM | 逻辑缺陷 | L1437-1441 | 系统字段名硬编码，未复用 GetSystemFields() |
| #5 | MEDIUM | 错误处理 | L880-900 | 删除物理字段时平台DB操作无事务保护 |
| #6 | LOW | 竞态 | L1177-1188 | 删除 FieldEnumRelation 时存在 check-then-act 竞态 |
| #7 | LOW | 代码质量 | L606-608 | nil 检查位置在循环内，未在入口快速失败 |

最值得关注的是 **#1**（状态反转，语义不明，需要确认是否为设计意图）和 **#3**（创建与删除路径事务保护不对称）。

---

*注：本 review 已参考 `ai-metadata/backend/common-mistakes.md` 中的 BM-20260415-0001 规则，但未正式调用 backend-checklist skill 的 review 流程（触发词不匹配）。如需完整 checklist review，可使用「用错题本 check」触发该 skill。*
