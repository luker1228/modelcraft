# .agents 格式速查表

快速对比各类型文件的格式差异。

---

## 📊 五类文件对比表

| 特性 | Command | Skill | Agent | Hook | Rule |
|------|---------|-------|-------|------|------|
| **文件名** | .md | SKILL.md | .md | .py/.sh | .md |
| **位置** | commands/ | skills/ | agents/ | hooks/ | rules/ |
| **Frontmatter** | ✓ | ✓ | ✓ | ✗ | ✓ |
| **自动触发** | 用户调用 | 自动 | 分派 | 钩子 | 自动 |
| **文件数** | 1 个 | 可多个 | 1 个 | 1 个 | 1 个 |

---

## 🏷️ Frontmatter 字段速查

### Command

```yaml
---
name: multi-agent
description: "Analyze task dependencies and dispatch work to multiple agents in parallel"
argument-hint: "<task description or list> [--agent <agent-type>]"
---
```

**字段说明：**
- `name`: 命令唯一标识
- `description`: 一行简短描述
- `argument-hint`: 参数用法提示

### Skill

```yaml
---
name: backend-debug
description: >
  排查和修复 ModelCraft 后端错误。当用户提供 GraphQL 响应中的错误...
  当用户提到 "后端报错了"、"接口返回错误" 时，使用此 skill。
  遇到上述情形时，主动使用此 skill。
---
```

**字段说明：**
- `name`: Skill 唯一标识（用于 /skill-name 触发）
- `description`: 多行描述，需包含触发场景和触发词汇
- `compatibility` (可选): 依赖的工具/MCP

### Agent

```yaml
---
name: backend-worker
description: 后端实现 worker，负责将技术方案文档转化为可运行的 Go 代码。

Examples:
- user: "按照这份技术方案，实现 LogicalForeignKey 的 Repository 层"
  assistant: "我来用 backend-worker agent 实现 Repository。"
  <commentary>
  有明确的实现任务和技术方案作为输入，backend-worker 负责写具体代码。
  </commentary>

- user: "给 DataModel 添加 description 字段..."
  assistant: "让我调用 backend-worker agent 来完成这个字段扩展。"
  <commentary>
  跨层修改但有明确范围，交给 backend-worker。
  </commentary>

tool: *
---
```

**字段说明：**
- `name`: Agent 标识
- `description`: 简短角色描述
- `Examples`: 至少 2-3 个使用示例（必须）
- `tool`: 可用工具（* = 全部）

### Rule

```yaml
---
paths:
  - "internal/**/*.go"
---
```

**字段说明：**
- `paths`: 此规则适用的文件路径模式（glob）

---

## 📁 文件结构对比

### 最简单（Command）

```
commands/
└── my-command.md
```

### 简单（Skill）

```
skills/
└── my-skill/
    └── SKILL.md
```

### 复杂（Skill）

```
skills/
└── my-skill/
    ├── SKILL.md
    ├── evals/
    │   ├── evals.json
    │   └── trigger-evals.json
    ├── scripts/
    │   └── helper.py
    ├── references/
    │   └── guide.md
    └── assets/
        └── resource.txt
```

### Agent（固定）

```
agents/
└── my-agent.md
```

### Hook（固定）

```
hooks/
└── my-hook.py
```

### Rule（固定）

```
rules/
└── backend/
    └── my-rule.md
```

---

## 🎯 何时使用哪个？

```
用户想要执行特定的任务工作流？
├─ YES, 用户会多次重复执行 → 创建 SKILL
├─ YES, 用户首次需要详细指导 → 创建 COMMAND
└─ NO

AI 需要被分派一个角色？
├─ YES, 在多代理场景中 → 创建 AGENT
└─ NO

需要在工具执行时拦截或通知？
└─ YES → 创建 HOOK

需要对特定文件路径的代码规范？
└─ YES → 创建 RULE
```

---

## 📝 Markdown 正文结构对比

### Command 的结构

```markdown
# 命令标题

## 概述

简要说明

## 步骤

### 1. 第一步
详细说明

### 2. 第二步
详细说明

## 常见问题
```

### Skill 的结构

```markdown
# Skill 标题

## 为什么需要

背景说明

## 第一步：...
详细步骤

## 第二步：...
详细步骤

## 常见错误速查

| 错误 | 原因 | 解决方案 |
|------|------|--------|
```

### Agent 的结构

```markdown
# Agent 标题

## 职责边界

### 做什么
- 列表项

### 不做什么
- 列表项

## 工作流程

1. 第一步
2. 第二步

## 使用技能表
```

### Rule 的结构

```markdown
# 规范标题

规范描述文本。

Refer to @ai-metadata/path/to/doc.md

## 触发场景

- 场景 1
- 场景 2
```

---

## 🔑 关键区别速记

| 维度 | Command | Skill | Agent |
|------|---------|-------|-------|
| **触发方式** | /command-name | 自动（关键词） | 手动分派 |
| **职责** | 工作流指导 | 单一任务 | 角色扮演 |
| **复用性** | 中等 | 高 | 中等 |
| **Examples** | 无 | 可选 | 必须 |
| **职责边界** | 无 | 隐含 | 明确 |
| **大小** | 中 | 小-中 | 中-大 |

---

## ✅ 质量检查清单

创建前检查：

- [ ] 选择了正确的类型？
- [ ] 文件保存在正确的目录？
- [ ] YAML frontmatter 格式正确？
- [ ] `name` 使用 kebab-case？
- [ ] `description` 足够清晰？
- [ ] Markdown 结构合理？
- [ ] 包含足够的示例/说明？
- [ ] 没有拼写或语法错误？

对于 Skill：
- [ ] description 包含触发词汇？
- [ ] 有清晰的工作步骤？
- [ ] 是否需要 evals/ 目录？

对于 Agent：
- [ ] 有 2-3 个 Examples？
- [ ] Examples 中有 `<commentary>`？
- [ ] 明确了职责边界？

对于 Command：
- [ ] 有 argument-hint？
- [ ] 步骤足够详细？
- [ ] 包含常见问题？

对于 Rule：
- [ ] paths 使用 glob 模式？
- [ ] 触发场景清晰？

对于 Hook：
- [ ] 脚本可执行权限？
- [ ] 正确的 shebang 行？
- [ ] 清晰的注释说明？

