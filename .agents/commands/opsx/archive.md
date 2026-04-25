---
name: OPSX: 归档
description: "在实验流程中归档已完成变更"
argument-hint: "[命令参数]"
---

在实验流程中归档已完成变更。

**输入**：可在 `/opsx:archive` 后可选指定变更名（例如：`/opsx:archive add-auth`）。若未提供，可从对话上下文推断；若语义模糊或存在歧义，**必须**提示用户从可用变更中选择。

**步骤**

1. **未提供变更名时，提示用户选择**

   运行 `openspec list --json` 获取可用变更。使用 **AskUserQuestion 工具** 让用户选择。

   仅展示活跃变更（未归档）。
   若可用，展示每个变更使用的 schema。

   **重要**：不要猜测或自动选择变更，始终由用户确认。

2. **检查工件完成状态**

   运行 `openspec status --change "<name>" --json` 检查工件完成情况。

   解析 JSON 获取：
   - `schemaName`：当前工作流 schema
   - `artifacts`：工件列表及其状态（`done` 或其他）

   **若存在非 `done` 工件：**
   - 显示警告并列出未完成工件
   - 提示用户确认是否继续
   - 用户确认后继续

3. **检查任务完成状态**

   读取任务文件（通常为 `tasks.md`）检查未完成任务。

   统计 `- [ ]`（未完成）与 `- [x]`（已完成）数量。

   **若存在未完成任务：**
   - 显示警告并给出未完成数量
   - 提示用户确认是否继续
   - 用户确认后继续

   **若不存在任务文件：**跳过任务相关警告并继续。

4. **评估 delta spec 同步状态**

   检查 `openspec/changes/<name>/specs/` 是否存在 delta specs。若不存在，跳过同步提示并继续。

   **若存在 delta specs：**
   - 将每个 delta spec 与对应主规范 `openspec/specs/<capability>/spec.md` 对比
   - 判断将发生的变更类型（新增、修改、删除、重命名）
   - 在询问前展示汇总结果

   **提示选项：**
   - 若存在待同步变更："立即同步（推荐）"、"跳过同步直接归档"
   - 若已同步："立即归档"、"仍执行同步"、"取消"

   若用户选择同步，使用 Task 工具（`subagent_type: "general-purpose"`，`prompt: "Use Skill tool to invoke openspec-sync-specs for change '<name>'. Delta spec analysis: <include the analyzed delta spec summary>"`）。无论是否同步，后续都继续归档流程。

5. **执行归档**

   若目录不存在，先创建：
   ```bash
   mkdir -p openspec/changes/archive
   ```

   使用当前日期生成目标名：`YYYY-MM-DD-<change-name>`

   **检查目标是否已存在：**
   - 若存在：报错并建议重命名已有归档或改期归档
   - 若不存在：移动变更目录

   ```bash
   mv openspec/changes/<name> openspec/changes/archive/YYYY-MM-DD-<name>
   ```

6. **展示总结**

   展示归档结果摘要：
   - 变更名
   - 使用的 schema
   - 归档位置
   - Spec 同步状态（已同步 / 跳过同步 / 无 delta specs）
   - 任何警告说明（工件/任务未完成）

**成功输出示例**

```text
## 归档完成

**变更：** <change-name>
**Schema：** <schema-name>
**归档位置：** openspec/changes/archive/YYYY-MM-DD-<name>/
**Specs：** ✓ 已同步到主规范

所有工件已完成。所有任务已完成。
```

**成功输出示例（无 Delta Specs）**

```text
## 归档完成

**变更：** <change-name>
**Schema：** <schema-name>
**归档位置：** openspec/changes/archive/YYYY-MM-DD-<name>/
**Specs：** 无 delta specs

所有工件已完成。所有任务已完成。
```

**成功输出示例（含警告）**

```text
## 归档完成（含警告）

**变更：** <change-name>
**Schema：** <schema-name>
**归档位置：** openspec/changes/archive/YYYY-MM-DD-<name>/
**Specs：** 已跳过同步（用户选择）

**警告：**
- 归档时仍有 2 个工件未完成
- 归档时仍有 3 个任务未完成
- 已跳过 Delta Spec 同步（用户选择）

如非预期，请复核该归档。
```

**失败输出示例（归档已存在）**

```text
## 归档失败

**变更：** <change-name>
**目标：** openspec/changes/archive/YYYY-MM-DD-<name>/

目标归档目录已存在。

**可选操作：**
1. 重命名已有归档
2. 若为重复归档则删除已有目录
3. 等待其他日期再归档
```

**护栏**
- 未提供变更名时必须让用户选择
- 使用工件图（`openspec status --json`）判断完成度
- 警告不应阻塞归档，但必须提示并确认
- 移动归档时保留 `.openspec.yaml`（随目录一起移动）
- 清晰总结实际执行结果
- 若用户请求同步，必须使用 Skill 工具调用 `openspec-sync-specs`（agent 驱动）
- 若存在 delta specs，必须先完成同步评估并展示汇总后再询问

---

## 语言约束（强制）
- 本命令流程中新生成或更新的文件内容必须使用中文（简体）。
- 命令、路径、代码标识符可保留原文；解释性文本必须为中文。
