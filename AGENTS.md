# AI Agents Documentation

This document provides guidelines for AI agents working with this project.

## Project Structure

This is the ModelCraft monorepo with all components co-located:

- **Backend (Go)**: Code in `./modelcraft-backend`. See [AGENTS.md](modelcraft-backend/AGENTS.md)
  @./modelcraft-backend/AGENTS.md
- **Frontend**: Code in `./modelcraft-front`. See [AGENTS.md](modelcraft-front/AGENTS.md)
  @./modelcraft-front/AGENTS.md
- **Gateway**: Code in `./modelcraft-gateway`
- **CLI**: Code in `./modelcraft-cli`

## AI Metadata

`ai-metadata/` 是项目中**唯一的**知识文档存放位置，不存在其他知识文档目录。

除非用户明确要求，agent 不得随意创建总结/复盘/说明类文档。

完整路径索引见 @./ai-metadata/index.md

知识按模块组织：

- **Backend**: @./ai-metadata/backend/README.md
- **Frontend**: @./ai-metadata/front/development/README.md

## Git 仓库结构

本项目采用 **monorepo** 结构，所有组件在同一 git 仓库中管理：

```
modelcraft/                    ← 根仓库 (origin: modelcraft.git)
├── modelcraft-backend/        ← Go 后端
├── modelcraft-front/          ← Next.js 前端
├── modelcraft-gateway/        ← Gateway 服务
└── modelcraft-cli/            ← CLI 工具
```

### API Contract 共享

后端 `modelcraft-backend/api/` 目录是 API Contract 的**唯一真相源**，前端通过 `front-contract-pull` skill 同步：

```
modelcraft-backend/api/          ← 唯一真相源
        │
        │  front-contract-pull skill
        ▼
modelcraft-front/contract/       ← 前端消费端（只读）
```

> 详见 @./ai-metadata/backend/development/contract-sync.md

### 提交规范

- 所有组件在同一仓库提交，无需跨仓库操作
- 修改 API 后运行 `front-contract-pull` 同步前端 contract

## Git Rules

- Never use `git commit --no-verify`. Pre-commit hooks must always run. If a hook fails, fix the underlying issue instead of bypassing it.
- **前端禁止直接修改 `contract/`** — 所有变更必须通过 `front-contract-pull` skill 获取。

Each subproject has its own pre-commit hook:

- **Backend** (`./modelcraft-backend`): runs `just lint`. If it fails, run `just lint-fix` to auto-fix, then re-verify with `just lint`.
- **Frontend** (`./modelcraft-front`): runs `npx lint-staged` via Husky. If it fails, fix the reported lint errors manually and re-commit.

## No Absolute Paths

- Do not use absolute paths (e.g., `/root/modelcraft_project/...`). Always use relative paths (e.g., `./modelcraft-backend/...`) when referencing files or directories.

## Use @ References for Documentation

> Refer to @docs/architecture.md for system flow.
or
- Refer to @docs/state-management.md before editing state.

Please refer to the respective documentation for detailed coding styles, patterns, and conventions for each component.

## Single Source of Truth

**`AGENTS.md` is the single source of truth** for AI agent configuration in this project.

All agent-specific directories (`.claude`, `.codebuddy`) are **symlinks** that point to `./.agents/`:

```
.claude/agents    -> .agents/agents
.claude/commands  -> .agents/commands
.claude/hooks     -> .agents/hooks
.claude/rules     -> .agents/rules
.claude/skills    -> .agents/skills

.codebuddy/agents   -> .agents/agents
.codebuddy/commands -> .agents/commands
.codebuddy/hooks    -> .agents/hooks
.codebuddy/rules    -> .agents/rules
.codebuddy/skills   -> .agents/skills
```

**Rules:**
- Edit content only in `./.agents/` — never edit through the symlink paths.

## Writing Rules

When creating or editing rules, use the frontmatter format with `paths` to scope the rule to specific files:

```markdown
---
paths:
  - "src/api/**/*.ts"
  - "src/services/**/*.ts"
---

# Rule Title

Rule content here...
```

This ensures rules are only loaded when relevant files are being edited, improving performance and context relevance.

## graphify

any input to knowledge graph. Trigger: `/graphify`
When the user types `/graphify`, invoke the Skill tool with `skill: "graphify"` before doing anything else.

This project has a graphify knowledge graph at graphify-out/.

Rules:
- Before answering architecture or codebase questions, read graphify-out/GRAPH_REPORT.md for god nodes and community structure
- If graphify-out/wiki/index.md exists, navigate it instead of reading raw files
