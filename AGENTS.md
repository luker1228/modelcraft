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

知识按模块组织：

- **Backend**: @./ai-metadata/backend/README.md
- **Frontend**: @./ai-metadata/front/development/README.md

## Git Rules

- Never use `git commit --no-verify`. Pre-commit hooks must always run. If a hook fails, fix the underlying issue instead of bypassing it.

Subproject commits must be made before committing in the root project. Always commit changes in `./modelcraft-backend` or `./modelcraft-front` first, then commit in the root.

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
.claude/commands  -> .agents/commands
.claude/hooks     -> .agents/hooks
.claude/rules     -> .agents/rules
.claude/skills    -> .agents/skills

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
