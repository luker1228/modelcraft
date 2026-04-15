# ModelCraft .agents/ 目录结构分析

## 目录总览

```
.agents/
├── commands/          # 用户可执行的命令
├── skills/            # AI Skill 库
├── agents/            # Agent 定义
├── hooks/             # 执行钩子（Python/Shell 脚本）
└── rules/             # 代码规范规则（Markdown + YAML）
```

---

## 1. commands/ 目录

### 目录内容
- `multi-agent.md` (5.0K)

### 格式规范 (以 multi-agent.md 为例)

```yaml
---
name: Multi-Agent                          # 命令名称
description: "描述文本"                    # 命令简短描述
argument-hint: "<参数提示>"                # 参数语法提示
---

# 标题

命令的详细说明（Markdown 格式）

## 章节

具体的执行步骤、指导和规则
```

**特点**：
- YAML frontmatter 包含 name、description、argument-hint
- 后跟 Markdown 详细文档
- 用于定义用户可执行的特殊命令工作流
- 包含多个章节、清晰的步骤指导

---

## 2. skills/ 目录

### 目录结构

```
skills/
├── backend-debug/
│   └── SKILL.md                          # 只有 SKILL.md 文件
├── backend-develop/
│   └── SKILL.md
├── front/
│   ├── design-token-extractor/
│   │   ├── SKILL.md
│   │   ├── evals/                        # 可选：测试用例目录
│   │   │   ├── evals.json
│   │   │   └── trigger-evals.json
│   │   └── workspace/                    # 可选：工作目录
│   ├── theme-factory/
│   └── ... (其他子 skill)
├── skill-creator/
│   └── SKILL.md
├── skill-manager/
│   └── SKILL.md
└── ... (其他 skill)
```

### Skill 文件格式 (SKILL.md)

```yaml
---
name: skill-name                           # 唯一标识，用于 /skill-name 触发
description: >
  多行描述文本。包含：
  1. Skill 的功能说明
  2. 触发场景（何时使用）
  3. 触发词汇（关键字）
---

# Skill 标题

Markdown 正文，包含：
- 背景/为什么需要这个 Skill
- 详细的执行步骤
- 代码示例
- 最佳实践
- 常见问题
- 快速参考表
```

### Skill 示例分析

#### 类型 1: 简单流程型 (backend-debug)
- 只有 SKILL.md
- 定义线性的问题排查流程
- 包含日志查看、代码定位、修复验证等步骤

**特征**：
```markdown
## 第一步：...
## 第二步：...
## 第三步：...
## 常见错误速查
```

#### 类型 2: 复杂工程型 (design-token-extractor)
- SKILL.md + evals/ + workspace/
- 更复杂的任务，需要测试用例
- evals/ 包含测试数据 (JSON 格式)
- workspace/ 可能包含配置或资源文件

**特征**：
```
├── SKILL.md (详细说明 + 实现规则)
├── evals/
│   ├── evals.json (测试用例)
│   └── trigger-evals.json (触发器测试)
└── workspace/
```

#### 类型 3: 项目型 (front/ 下的子 skill)
- 前端相关的专门 Skill
- 例如：artifact-preview、front-contract-pull、ui-ux-pro-max

#### 类型 4: 元 Skill (skill-creator)
- 用于创建/改进其他 Skill
- 定义 Skill 的创建流程

### Skill 描述的关键格式特征

根据 skill-creator 的说明，描述应该包括：

```
Skill 的功能说明。说明何时使用、触发词汇。

当用户提到 [关键词1]、[关键词2]、[关键词3] 时，应该使用此 Skill。
也包含 [其他自动触发场景]。
```

**"推荐实践"**：
- 描述应该"有点推荐性"（略显主动）
- 包含明确的触发条件
- 用具体的触发词汇而不是通用描述

---

## 3. agents/ 目录

### 目录内容

```
agents/
├── agents-md-purifier.md
├── backend-api.md
├── backend-reviewer.md
├── backend-worker.md         # 13.3K - 最详细的例子
├── front-architect.md
├── front-worker.md
├── metadata-index-keeper.md
├── pm.md                      # 5.0K
└── prd-page-splitter.md
```

### Agent 文件格式

```yaml
---
name: agent-name                # Agent 在多代理场景中的标识
description: Agent 的简短描述

Examples:                        # 使用示例，包含用户输入和 Assistant 响应
- user: "用户输入1"
  assistant: "AI 响应1"
  <commentary>
  说明这个例子为什么需要这个 Agent
  </commentary>

- user: "用户输入2"
  assistant: "AI 响应2"
  <commentary>
  说明这个例子为什么需要这个 Agent
  </commentary>

tool: *                         # 工具配置（* 表示所有工具可用）
---

# Agent 详细说明

用 Markdown 写的完整 Agent 角色定义、职责边界、工作流程。

## 职责边界

### 做什么
- ...
- ...

### 不做什么
- ...
- ...

## 工作流程

1. 第一步
2. 第二步
3. ...
```

### Agent 示例分析

#### pm.md (产品经理 Agent)
**特点**：
- 定义了思维框架和沟通方式
- 包含 PRD 模板
- 明确了和用户的交互方式

**结构**：
```
---
name: pm
Examples: [3 个示例]
tool: *
---

# 标题

## 思维框架
## 沟通方式
## PRD 格式
## 竞品参考
## 开始
```

#### backend-worker.md (后端实现 Worker)
**特点**：
- 最复杂的 Agent，13.3K 大小
- 明确的职责边界（做什么/不做什么）
- 包含 TDD 开发节奏
- 参考文档表
- 技术栈清单

**结构**：
```
---
name: backend-worker
Examples: [4 个示例，展示代码实现场景]
tool: *
---

# 后端实现 Worker

## 职责边界
### 做什么
### 不做什么

## TDD 开发节奏

## 参考文档表

## 技术栈（固定）

## 关键模式
- Domain 层 Repository 接口设计
- Infrastructure 层 sqlc 包装
- Application 层事务管理
- Interfaces 层 GraphQL resolver

## 使用技能表
```

---

## 4. hooks/ 目录

### 目录内容

```
hooks/
├── check-documentation.py      # PreToolUse 钩子 - Python
├── notify-wecom.sh             # Shell 脚本
├── notify-wecom-stop.sh        # Shell 脚本
└── rtk-rewrite.sh              # Shell 脚本
└── task-lint.sh                # Shell 脚本
```

### Hook 脚本格式 (Python 示例)

```python
#!/usr/bin/env python3
"""
PreToolUse hook: 钩子触发时机和功能说明。

从 stdin 读取 JSON 输入（事件类型），检查/处理...

退出码: 0 = 允许, 2 = 阻止 (或其他自定义码)
"""
import json
import sys

def main():
    # 读取输入
    data = json.load(sys.stdin)
    
    # 提取信息
    file_path = data.get("tool_input", {}).get("file_path", "")
    
    # 验证逻辑
    if should_block(file_path):
        print(json.dumps({
            "continue": False,
            "stopReason": "原因说明"
        }))
        sys.exit(2)
    
    sys.exit(0)

if __name__ == "__main__":
    main()
```

### Hook 的作用类型

1. **PreToolUse** - 工具执行前的预检查 (check-documentation.py)
   - 阻止写入敏感文件
   - 验证文件创建策略
   
2. **通知钩子** - 执行后的通知 (notify-wecom.sh)
   - 集成消息通知服务

3. **代码转换钩子** - 自动代码改写 (rtk-rewrite.sh)

---

## 5. rules/ 目录

### 目录结构

```
rules/
├── backend/
│   ├── architecture.md         # DDD 分层架构规范
│   ├── comments.md             # 代码注释规范
│   ├── context-handling.md     # 上下文处理规范
│   ├── error-handling.md       # 错误处理规范
│   ├── logging.md              # 日志规范
│   ├── repo-develop.md         # Repository 模式规范
│   ├── sqlc-custom-types.md    # sqlc 自定义类型规范
│   └── type-conversion.md      # 类型转换规范
├── front/
│   ├── apollo-client-stability.md
│   ├── frontend-layout/        # 前端布局相关规范
│   └── styling/                # 样式相关规范
```

### Rule 文件格式

```yaml
---
paths:                          # 此规则适用的文件路径模式
  - "internal/**/*.go"
  - "path/to/**/*.ts"
---

# 规范标题

规范描述文本。

Refer to @ai-metadata/path/to/detailed-doc.md for detailed information.

## 触发场景

- 场景 1
- 场景 2
- 场景 3
```

### Rule 示例分析 (architecture.md)

```yaml
---
paths:
  - "internal/**/*.go"          # 仅适用于内部 Go 文件
---

# 架构分层规范

开发任何 `internal/` 目录下的代码时，必须遵守 DDD 分层依赖规则。

Refer to @ai-metadata/backend/development/architecture.md 
for detailed layer responsibilities...

## 触发场景

- 新增或修改 `internal/` 目录下任何文件时
- 添加跨层 import 语句时
- 设计新模块、服务或仓储时
```

---

## 格式总结表

| 目录 | 文件类型 | 格式 | YAML 字段 | 功能 |
|------|---------|------|----------|------|
| commands/ | .md | Frontmatter + Markdown | name, description, argument-hint | 定义可执行命令 |
| skills/ | SKILL.md | Frontmatter + Markdown | name, description, (optional: compatibility) | 执行特定任务的 Skill |
| agents/ | .md | Frontmatter + Markdown | name, description, Examples, tool | 多代理场景中的角色定义 |
| hooks/ | .py/.sh | 脚本 | (无 Frontmatter) | 执行生命周期钩子 |
| rules/ | .md | Frontmatter + Markdown | paths | 代码规范检查规则 |

---

## 关键设计模式

### 1. Frontmatter + Markdown 二段式
几乎所有 Markdown 类文件都采用：
- **YAML Frontmatter** - 元数据（name、description、paths 等）
- **Markdown Body** - 详细内容

### 2. 嵌套结构支持
- `skills/front/` 下可以有多个子 skill 目录
- 每个子 skill 独立有 SKILL.md
- 可选的 evals/ 和 workspace/ 目录

### 3. 引用机制
- 使用 `@ai-metadata/` 前缀引用外部文档
- `Refer to @path/to/doc.md` 语法

### 4. 可选资源层
较复杂的 Skill/Agent 可以包含：
- `evals/` - 测试用例（JSON）
- `workspace/` - 工作文件
- `scripts/` - 可执行脚本
- `references/` - 参考文档
- `assets/` - 静态资源

---

## 创建新命令/Skill 的模板

### 命令模板 (commands/)

```yaml
---
name: your-command-name
description: "简短描述 (一行)"
argument-hint: "<arg1> [--flag <value>]"
---

# 你的命令

## 概述

简要说明

## 步骤

### 1. 第一步
...

### 2. 第二步
...
```

### Skill 模板 (skills/)

```yaml
---
name: your-skill-name
description: >
  功能说明。
  当用户提到 [关键词1]、[关键词2]、[关键词3]、
  或者 [场景描述] 时，使用此 Skill。
  也自动触发于 [其他场景]。
---

# Skill 标题

## 为什么需要

背景说明

## 工作流程

### 第一步：...
### 第二步：...

## 常见问题

## 参考
```

### 目录结构

```
skills/your-skill/
├── SKILL.md                    # 必须
├── evals/                      # 可选：测试用例
│   └── evals.json
├── scripts/                    # 可选：工具脚本
│   └── helper.py
├── references/                 # 可选：参考文档
│   └── guide.md
└── assets/                     # 可选：资源文件
```
