---
name: doPropose
description: "从模糊想法到结构化方案：先用 idea-refine 发散精炼创意，再用 spec-driven-development 产出规范文档"
argument-hint: "[idea or topic]"
---

将用户的模糊想法或主题经过两阶段处理，输出可执行的规范化方案。

## 流程

### Step 1 — `idea-refine` 发散与精炼

- 输入：用户提供的想法或主题（`$ARGUMENTS`）
- 使用 `idea-refine` skill 对想法进行结构化发散和收敛思考
- 产出：精炼后的核心需求和方向
- 人工确认后继续

### Step 2 — `spec-driven-development` 规范化

- 输入：Step 1 精炼后的需求
- 使用 `spec-driven-development` skill 将需求转化为正式规范文档（spec）
- 产出：可供开发使用的 spec 文档

## 约束

- 必须先完成 Step 1 并经用户确认，再进入 Step 2
- 如果用户未提供任何输入（`$ARGUMENTS` 为空），先提示用户描述想法，不继续执行
