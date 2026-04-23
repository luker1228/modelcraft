---
name: doPlan
description: "将需求或 spec 文档拆解为有序任务列表，使用 planning-and-task-breakdown 输出可并行执行的开发计划"
argument-hint: "[spec file path or requirement description]"
---

基于已有需求或 spec 文档，生成结构化的开发任务分解和执行计划。

## 流程

### Step 1 — 输入校验

- 检查 `$ARGUMENTS` 是否提供了 spec 文件路径或需求描述
- 如果是文件路径，校验文件存在且可读；缺失则立即停止并提示
- 如果未提供任何输入，提示用户补充需求说明或 spec 文件路径

### Step 2 — `planning-and-task-breakdown` 任务拆解

- 输入：`$ARGUMENTS`（spec 文件或需求描述）
- 使用 `planning-and-task-breakdown` skill 将需求拆解为有序、可执行的任务
- 产出：
  - 任务列表（含依赖关系与并行机会）
  - 每个任务的 owner 建议（前端 / 后端 / 共享）
  - 执行顺序与波次建议

## 约束

- 如果缺少必要输入，立即停止并提示具体缺失信息
- 输出的任务粒度应足够细，可直接分配给具体 agent 执行
