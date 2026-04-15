# .agents 目录结构探索总结

## 📁 已生成的文档

在 `/data/home/lukemxjia/modelcraft/.agents/` 目录中已生成两份详细文档：

1. **DIRECTORY_STRUCTURE.md** (11KB)
   - 完整的目录结构分析
   - 每个目录的格式规范详解
   - Frontmatter 字段说明
   - 5 个目录的深入分析

2. **QUICK_REFERENCE.md**
   - 快速查找指南
   - 各类型文件创建模板
   - 最佳实践
   - 常见问题解答
   - 文件路径速查表

---

## 🎯 核心发现

### 目录组成（5 个目录）

| 目录 | 用途 | 文件类型 | 例子 |
|------|------|--------|------|
| **commands/** | 用户可执行命令 | .md | multi-agent.md |
| **skills/** | AI 专门 Skill 库 | SKILL.md | backend-debug/ |
| **agents/** | 多代理角色定义 | .md | pm.md, backend-worker.md |
| **hooks/** | 执行生命周期钩子 | .py/.sh | check-documentation.py |
| **rules/** | 代码规范规则 | .md | backend/architecture.md |

### 关键格式模式

**1. Frontmatter + Markdown 二段式**
```yaml
---
name: 标识
description: 功能说明
# 其他字段
---

# Markdown 正文
```

**2. 每个目录的 Frontmatter 字段**

| 目录 | 必需字段 | 可选字段 |
|------|--------|--------|
| commands/ | name, description, argument-hint | - |
| skills/ | name, description | compatibility |
| agents/ | name, description, Examples, tool | - |
| hooks/ | (无) | - |
| rules/ | paths | - |

**3. Skill 和 Agent 的关键差异**

- **Skill**: 自动触发，有明确的工作流程，通常是功能性的
- **Agent**: 手动分派角色，有职责边界，在多代理场景中使用
- **Command**: 用户显式调用，包含详细的引导步骤

---

## 📋 现存 Skill 分类

### 简单流程型 Skill
- backend-debug（日志查看 → 代码定位 → 修复验证）
- backend-develop（后端开发完整指南）
- db-develop（数据库开发指南）

**特征**: 只有 SKILL.md 文件

### 复杂工程型 Skill
- design-token-extractor（含 evals/ + workspace/）

**特征**: SKILL.md + evals/ + workspace/

### 项目型 Skill（前端）
- front/artifact-preview
- front/ui-ux-pro-max
- front-contract-pull（顶层，非 front/ 子目录）
- 等

**特征**: 在 front/ 下的子目录结构

---

## 📝 现存 Agent 分析

| Agent | 大小 | 特点 |
|-------|------|------|
| pm.md | 5.0K | 思维框架 + PRD 模板 |
| backend-worker.md | 13.3K | 最复杂，有职责边界、TDD 流程 |
| backend-api.md | - | API 设计角色 |
| backend-reviewer.md | 7.2K | 代码审查角色 |
| front-worker.md | 13.8K | 前端实现角色 |
| front-architect.md | 7.5K | 前端架构角色 |

**共同特征**: 
- 都有 Examples 部分（2-3 个示例）
- tool: * （可访问所有工具）
- 明确的职责边界

---

## 🔧 Hook 和 Rule 说明

### Hooks (5 个)
1. **check-documentation.py** - PreToolUse：阻止敏感文件写入
2. **notify-wecom.sh** - 发送通知
3. **notify-wecom-stop.sh** - 停止通知
4. **rtk-rewrite.sh** - 代码转换
5. **task-lint.sh** - 任务 lint

### Rules 结构
- **backend/** (8 个规范)：架构、注释、错误处理等
- **front/** (3 个)：Apollo 稳定性、布局、样式

---

## 🎓 创建新命令的标准流程

### 创建 Skill 的 4 步

1. **确定场景** - 用户何时会使用它，有哪些关键词
2. **写 SKILL.md** - 包含 frontmatter + 详细步骤
3. **添加触发词汇** - 在 description 中明确列出
4. **可选：添加资源** - evals/、scripts/、references/ 等

### 创建 Agent 的 4 步

1. **定义角色** - 这个 Agent 的职责是什么
2. **确定边界** - 做什么/不做什么
3. **写示例** - 至少 2-3 个 Examples
4. **详细说明** - 工作流程和模式

### 创建 Command 的 3 步

1. **设计参数** - 需要哪些输入
2. **规划步骤** - 分几步完成
3. **写指南** - 详细的步骤说明

---

## 💡 最佳实践

### Skill 描述应该"推荐性强"

好的例子：
```
在实现任何 React 组件前，必须使用此 skill 从 prototype HTML/CSS 中提取 design token。
当用户提到 "原型"、"还原"、"实现页面"、"HTML 转 React" 时，立刻调用此 skill。
```

不好的例子：
```
提取设计 token 的工具。
```

### 使用嵌套结构组织复杂 Skill

```
skills/front/
├── design-token-extractor/
│   ├── SKILL.md
│   ├── evals/
│   └── workspace/
└── theme-factory/
```

### 使用 Frontmatter 引用

```
Refer to @ai-metadata/backend/development/architecture.md for details.
```

---

## 📚 相关文档位置

两份生成的文档已保存在：

```
/data/home/lukemxjia/modelcraft/.agents/
├── DIRECTORY_STRUCTURE.md      # 完整分析（11KB）
├── QUICK_REFERENCE.md          # 快速指南（推荐先看）
└── EXPLORATION_SUMMARY.md      # 本文档（总结）
```

---

## 🚀 后续步骤建议

1. **先读 QUICK_REFERENCE.md** - 快速了解各类型文件的创建方法
2. **查看现存示例** - 特别是 backend-debug/SKILL.md 和 backend-worker.md
3. **遵循模板创建** - 使用文档中的标准模板创建新命令/Skill
4. **验证格式** - 检查生成的清单项确保格式正确

