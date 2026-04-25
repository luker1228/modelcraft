---
name: OPSX: 发起提案
description: "发起一个新变更：一步创建并生成全部工件"
argument-hint: "[命令参数]"
---

发起一个新变更：创建变更并一步生成全部工件。

我会创建包含以下工件的变更：
- `proposal.md`（做什么、为什么）
- `design.md`（如何实现）
- `tasks.md`（实施步骤）

准备进入实现时，执行 `/opsx:apply`。

---

**语言约束（强制）**
- 本命令生成或更新的所有工件文件内容必须使用中文（简体）。
- 代码、命令、路径、配置键名、协议字段名可保留原文；说明性文字必须为中文。

---

**输入**：`/opsx:propose` 后的参数可以是变更名（kebab-case），也可以是用户想构建内容的自然语言描述。

**步骤**

1. **如果未提供输入，先询问用户要做什么**

   使用 **AskUserQuestion 工具**（开放问题，不提供预设选项）询问：
   > "你希望推进什么变更？请描述你想构建或修复的内容。"

   根据用户描述推导 kebab-case 名称（例如："新增用户认证" → `add-user-auth`）。

   **重要**：在没有理解用户目标前，不要继续。

2. **创建变更目录**
   ```bash
   openspec new change "<name>"
   ```
   该命令会在 `openspec/changes/<name>/` 创建脚手架及 `.openspec.yaml`。

3. **获取工件构建顺序**
   ```bash
   openspec status --change "<name>" --json
   ```
   解析 JSON，获取：
   - `applyRequires`：实现前必须完成的工件 ID 列表（例如：`["tasks"]`）
   - `artifacts`：全部工件及其状态、依赖关系

4. **按依赖顺序创建工件，直到可执行 apply**

   使用 **TodoWrite 工具** 跟踪工件进度。

   按依赖关系循环处理（优先处理无未满足依赖的工件）：

   a. **对于每个状态为 `ready` 的工件（依赖已满足）**：
      - 获取说明：
        ```bash
        openspec instructions <artifact-id> --change "<name>" --json
        ```
      - 说明 JSON 包含：
        - `context`：项目背景（仅作为约束，不写入输出）
        - `rules`：工件规则（仅作为约束，不写入输出）
        - `template`：输出文件结构模板
        - `instruction`：该工件类型的写作指导
        - `outputPath`：输出路径
        - `dependencies`：需先读取的已完成依赖工件
      - 读取已完成依赖文件作为上下文
      - 使用 `template` 结构创建工件文件（正文使用中文）
      - 按 `context` 和 `rules` 约束写作，但不要把它们原样写入文件
      - 简短汇报进度："已创建 <artifact-id>"

   b. **持续到 `applyRequires` 全部完成**：
      - 每创建一个工件后重新执行 `openspec status --change "<name>" --json`
      - 检查 `applyRequires` 中每个工件在 `artifacts` 中是否 `status: "done"`
      - 当全部满足后停止

   c. **若某工件需要用户补充信息**（上下文不清晰）：
      - 使用 **AskUserQuestion 工具** 提问澄清
      - 获得答复后继续创建

5. **展示最终状态**
   ```bash
   openspec status --change "<name>"
   ```

**输出**

完成全部工件后，汇总以下内容（中文）：
- 变更名称与路径
- 已创建工件及简要说明
- 就绪状态："所有工件已创建，已可进入实现。"
- 提示："运行 `/opsx:apply` 开始实现。"

**工件创建指南**

- 严格遵循 `openspec instructions` 中 `instruction` 字段对工件类型的要求
- schema 决定工件内容范围，按 schema 执行
- 创建新工件前先读取依赖工件获取上下文
- 使用 `template` 作为输出结构，并完整填充
- **重要**：`context` 与 `rules` 是给你的约束，不是输出内容
  - 不要把 `<context>`、`<rules>`、`<project_context>` 等块复制到工件中
  - 它们只用于指导写作，不应出现在最终文件

**护栏**
- 创建实现所需的全部工件（由 schema 的 `apply.requires` 定义）
- 创建任何工件前都要先读取其依赖工件
- 若上下文关键缺失，优先询问用户；否则应做合理决策保持推进
- 若同名变更已存在，询问用户是继续该变更还是创建新变更
- 每次写入后确认工件文件已存在，再进入下一个工件
- 默认所有新建工件正文使用中文（简体）
