# 任务执行结果

## 执行动作

用户要求记录一个 Bug：凡是查 `model_enums` 表的 SQL，必须同时带 `org_name` 和 `project_slug`，不能只带 `project_slug`。

按照 `backend-checklist` skill 的 **add** 流程执行：

1. 读取 `ai-metadata/backend/common-mistakes.md`，确认当前已有记录
2. 检查上下文信息是否足够，提取以下内容：
   - **问题描述**：SQL 查询缺少 org_name 过滤导致跨租户数据污染
   - **根因**：`project_slug` 不是全局唯一键，只在同一 org 内唯一
   - **分类**：data-isolation / sql-query
   - **Checklist 规则**：凡是查询带 `project_slug` 的 SQL，必须同时带 `org_name`

## 写入结果

Bug 已成功写入错题本：

- **编号**：`BM-20260415-0001`
- **标题**：SQL 查询缺少 org_name 过滤导致跨租户数据污染
- **严重程度**：CRITICAL
- **分类**：data-isolation / sql-query
- **文件位置**：`ai-metadata/backend/common-mistakes.md`

## 写入内容摘要

```markdown
### BM-20260415-0001 · SQL 查询缺少 org_name 过滤导致跨租户数据污染

**严重程度**: CRITICAL
**分类**: data-isolation / sql-query
**发现日期**: 2026-04-15

#### 问题描述
`GetEnumsByNames` SQL 查询只用 `project_slug` 过滤，缺少 `org_name` 条件。
由于 `project_slug` 在不同 org 之间可能重名，导致 runtime 加载枚举时拿到了
**其他 org 的同名枚举定义**，枚举 options 不一致，引发错误。

#### 错误代码
// ❌ 错误
SELECT * FROM model_enums
WHERE project_slug = ? AND name IN (sqlc.slice('names'));

// ✅ 正确
SELECT * FROM model_enums
WHERE org_name = ? AND project_slug = ? AND name IN (sqlc.slice('names'));

#### 根因
`project_slug` 不是全局唯一键，只在同一 org 内唯一。跨 org 时同一
`project_slug` + 同一 `enum name` 会命中不属于该 org 的记录。

#### ⚡ Checklist 规则（Review 时必查）
> **凡是查询带 `project_slug` 的 SQL，必须同时带 `org_name`。**
> 适用表：`model_enums`、`field_definitions`、`models`、
> `logical_foreign_keys`、`model_field_enum_associations` 等所有 project 域资源表。
```

## Skill 执行评估

- **是否触发了正确的 skill 能力**：是，触发了 `add` 能力
- **是否正确读取了错题本**：是，先读取了 `common-mistakes.md` 确认现有编号
- **是否按照标准模板写入**：是，包含了所有必要字段
- **Checklist 规则质量**：规则以「凡是…必须…」开头，可机械地对照代码判断，明确指定了适用表范围
- **编号格式**：符合 `BM-YYYYMMDD-XXXX` 格式
