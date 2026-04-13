# AI Agents Documentation

This document provides guidelines for AI agents working with this project.

## Project Structure

This is the ModelCraft project with separate frontend and backend codebases:

- **Backend (Go)**: Code in `./modelcraft-backend`. See [AGENTS.md](modelcraft-backend/AGENTS.md)
  @./modelcraft-backend/AGENTS.md
- **Frontend**: Code in `./modelcraft-front`. See [AGENTS.md](modelcraft-front/AGENTS.md)
  @./modelcraft-front/AGENTS.md

## AI Metadata

`ai-metadata/` 是项目中**唯一的**知识文档存放位置，不存在其他知识文档目录。

完整路径索引见 @./ai-metadata/index.md

知识按模块组织：

- **Backend**: @./ai-metadata/backend/README.md
- **Frontend**: @./ai-metadata/front/development/README.md

## Git 仓库结构

本项目采用 **git submodule** + **git subtree** 混合管理。

### Submodule（子项目）

根仓库通过 submodule 引入两个子项目：

| 子模块 | 路径 | 远程仓库 | 追踪分支 |
|--------|------|----------|----------|
| Backend (Go) | `./modelcraft-backend` | `modelcraft-go` (main) | `main` |
| Frontend (Next.js) | `./modelcraft-front` | `modelcraft-front` | 默认 |

```bash
# 克隆项目（含子模块）
git clone --recurse-submodules <repo-url>

# 已有仓库拉取子模块
git submodule update --init --recursive

# 更新子模块到远程最新
git submodule update --remote
```

### Subtree（API Contract 共享）

后端 `api/` 目录是 API Contract 的**唯一真相源**，通过 **git subtree** 与前端共享：

```
modelcraft-backend/api/          ← 唯一真相源（后端仓库）
        │
        │  git subtree push --prefix=api
        ▼
modelcraft-api-contracts        ← 共享仓库 (contracts remote)
        │
        │  git subtree add/pull --prefix=contract
        ▼
modelcraft-front/contract/      ← 前端消费端（只读）
```

| 项目 | Remote 名称 | Subtree 前缀 | Squash |
|------|-------------|-------------|--------|
| Backend | `contracts` | `api/` | 否 |
| Frontend | `contracts` | `contract/` | 是 |

共享仓库: `https://git.woa.com/lukemxjia/modelcraft-api-contracts.git`

> 详见 @./ai-metadata/backend/development/contract-sync.md

### 提交顺序

1. **子项目内提交** — 在 `./modelcraft-backend` 或 `./modelcraft-front` 中提交代码变更
2. **subtree push（如需）** — 后端修改 API 后执行 `git subtree push --prefix=api contracts main`
3. **subtree pull（如需）** — 前端执行 `git subtree pull --prefix=contract contracts main --squash`
4. **根项目提交** — 回到根项目，`git add modelcraft-backend modelcraft-front` 后提交子模块引用

## Git Rules

- Never use `git commit --no-verify`. Pre-commit hooks must always run. If a hook fails, fix the underlying issue instead of bypassing it.
- **前端禁止直接修改 `contract/`** — 所有变更必须通过 subtree pull 获取。
- **先 push 再 pull** — 后端必须先 `subtree push`，前端才能 `subtree pull`。

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
- Each subproject (`modelcraft-backend/`, `modelcraft-front/`) follows the same symlink structure independently.

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

This project has a graphify knowledge graph at graphify-out/.

Rules:
- Before answering architecture or codebase questions, read graphify-out/GRAPH_REPORT.md for god nodes and community structure
- If graphify-out/wiki/index.md exists, navigate it instead of reading raw files
- After modifying code files in this session, run `python3 -c "from graphify.watch import _rebuild_code; from pathlib import Path; _rebuild_code(Path('.'))"` to keep the graph current
