# .agents/ 快速参考指南

快速查找命令/Skill 创建指南。

---

## 我想创建一个新的...

### ✨ Skill（执行特定任务的专门模块）

**场景**：
- 用户经常需要执行某个特定的工作流
- 有明确的输入、步骤和输出
- 需要被多次调用和重复使用

**文件路径**：
```
skills/my-skill/SKILL.md
```

**基本模板**：
```yaml
---
name: my-skill
description: >
  该 Skill 做什么。
  当用户提到 [关键词1]、[关键词2] 时，使用此 Skill。
---

# My Skill

## 为什么需要

背景说明

## 工作步骤

### 步骤 1
...

### 步骤 2
...

## 常见问题
```

**文件结构**：
```
skills/my-skill/
├── SKILL.md           # 必须
├── evals/             # 可选：测试用例
│   └── evals.json
├── scripts/           # 可选：脚本
│   └── helper.py
└── references/        # 可选：参考文档
    └── guide.md
```

**YAML 字段**：
- `name` (必须) - Skill 唯一标识，用于 `/skill-name` 触发
- `description` (必须) - 包含触发词汇和功能说明
- `compatibility` (可选) - 依赖的工具/MCP

---

### 🤖 Agent（多代理场景中的角色）

**场景**：
- 定义一个专门的角色（如"后端 Worker"、"产品经理"）
- 在多代理系统中被分派工作
- 有明确的职责边界和工作流程

**文件路径**：
```
agents/my-agent.md
```

**基本模板**：
```yaml
---
name: my-agent
description: Agent 的简短描述

Examples:
- user: "用户输入示例"
  assistant: "AI 如何使用此 Agent 的示例"
  <commentary>
  为什么这个例子需要这个 Agent
  </commentary>

tool: *
---

# My Agent

## 职责边界

### 做什么
- 任务1
- 任务2

### 不做什么
- 任务3
- 任务4

## 工作流程

1. 第一步
2. 第二步
3. ...
```

**YAML 字段**：
- `name` (必须) - Agent 标识
- `description` (必须) - 简短描述
- `Examples` (必须) - 至少 2-3 个示例
- `tool` - 可用工具（`*` 表示全部）

---

### 📋 Command（用户可执行的命令）

**场景**：
- 定义一个特殊的工作流程或工具集
- 需要详细的步骤指导
- 用于引导用户完成复杂的多步骤任务

**文件路径**：
```
commands/my-command.md
```

**基本模板**：
```yaml
---
name: my-command
description: "简短描述"
argument-hint: "<arg1> [--flag value]"
---

# My Command

## 概述

简要说明

## 步骤

### 1. 第一步
...

### 2. 第二步
...

### 3. 第三步
...
```

**YAML 字段**：
- `name` (必须) - 命令名称
- `description` (必须) - 简短描述
- `argument-hint` (必须) - 参数语法提示

---

### 🚫 Hook（执行生命周期钩子）

**场景**：
- 在工具执行前/后进行预检查或通知
- 验证文件写入、阻止危险操作
- 集成外部系统通知

**文件路径**：
```
hooks/my-hook.py
```

**基本模板** (Python PreToolUse)：
```python
#!/usr/bin/env python3
"""
PreToolUse hook: 功能说明。

从 stdin 读取 JSON，检查...
退出码: 0 = 允许, 2 = 阻止
"""
import json
import sys

def main():
    data = json.load(sys.stdin)
    
    # 提取信息并验证
    if should_block(data):
        print(json.dumps({
            "continue": False,
            "stopReason": "原因"
        }))
        sys.exit(2)
    
    sys.exit(0)

if __name__ == "__main__":
    main()
```

---

### 📐 Rule（代码规范检查规则）

**场景**：
- 针对特定文件路径的代码规范
- 在修改匹配路径的文件时自动触发
- 提供规范和最佳实践指导

**文件路径**：
```
rules/backend/my-rule.md
```

**基本模板**：
```yaml
---
paths:
  - "internal/**/*.go"
  - "path/to/**/*.ts"
---

# 规范标题

规范描述文本。

Refer to @ai-metadata/path/to/detailed-doc.md for details.

## 触发场景

- 场景 1
- 场景 2
```

**YAML 字段**：
- `paths` (必须) - 文件路径模式（glob）

---

## 最佳实践

### Skill 描述应该"推荐性强"

❌ 不好：
```
description: "提取 design token 的工具"
```

✅ 好：
```
description: >
  在实现任何 React 组件前，必须使用此 skill 从 prototype HTML/CSS 中提取 design token。
  当用户提到 "原型"、"还原"、"实现页面" 时，立刻调用此 skill。
```

### 使用 Frontmatter + Markdown 结构

```yaml
---
name: ...          # YAML 元数据
description: ...
---

# Markdown 正文
```

### 嵌套 Skill

在 `front/` 下可以有多个子 skill：
```
skills/front/
├── design-token-extractor/
│   └── SKILL.md
├── contract-sync/
│   └── SKILL.md
└── ui-ux-pro-max/
    └── SKILL.md
```

### 引用外部文档

```
Refer to @ai-metadata/backend/development/architecture.md for details.
```

---

## 文件检查清单

创建新文件时，检查以下项目：

- [ ] YAML frontmatter 格式正确（`---` 包围）
- [ ] `name` 字段使用 kebab-case（小写+连字符）
- [ ] `description` 包含明确的触发词汇（对于 Skill）
- [ ] Markdown 标题使用 `# 单个标题`（不要嵌套）
- [ ] 包含足够详细的说明和示例
- [ ] 对于 Skill：有清晰的工作步骤
- [ ] 对于 Agent：有明确的职责边界和示例
- [ ] 对于 Command：有详细的执行步骤
- [ ] 对于 Rule：有明确的 `paths` 和触发场景
- [ ] 文件保存在正确的目录中

---

## 常见问题

**Q: Skill 和 Command 有什么区别？**
A: Skill 是 AI 可以自动触发的模块，Command 是用户显式调用的工作流。Command 通常更复杂，需要更详细的人工指导。

**Q: 我的 Skill 需要测试吗？**
A: 如果有客观的验收标准（如文件转换、数据提取），应该在 `evals/` 目录中添加测试用例。

**Q: 如何让 Skill 被触发？**
A: 在 `description` 中明确列出触发词汇。例如：
```
当用户提到 "后端报错了"、"接口返回错误"、"定位问题" 时，使用此 Skill。
```

**Q: 能在 Agent 中使用 Skill 吗？**
A: 可以。在 Agent 中引用 Skill 的方法是说明："当您遇到 XXX 情况时，使用 `/skill-name` skill"。

**Q: 如何组织复杂的 Skill？**
A: 使用可选的子目录：
```
skills/my-complex-skill/
├── SKILL.md
├── evals/          ← 测试用例
├── scripts/        ← 执行脚本
├── references/     ← 参考文档
└── assets/         ← 静态资源
```

---

## 文件路径速查

| 想做什么 | 路径 |
|---------|------|
| 创建 Skill | `skills/skill-name/SKILL.md` |
| 创建子 Skill（前端） | `skills/front/skill-name/SKILL.md` |
| 创建 Agent | `agents/agent-name.md` |
| 创建 Command | `commands/command-name.md` |
| 创建 Hook | `hooks/hook-name.py` |
| 创建后端规则 | `rules/backend/rule-name.md` |
| 创建前端规则 | `rules/front/rule-name.md` |

