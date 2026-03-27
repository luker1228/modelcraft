# Spec: AGENTS.md 知识分层规范 & Spec Review Agent

> 日期: 2026-03-27
> 状态: Draft

## 背景

当前 AGENTS.md 文件混合了通用规则和模块知识内容，职责不清。需要明确分层：

- **AGENTS.md** = 通用规则 + 地图索引（不含模块知识）
- **ai-metadata/{module}/** = 模块知识（按 backend/front 分目录）

### CODEBUDDY.md Symlink

项目中 `CODEBUDDY.md` 是指向 `AGENTS.md` 的 symlink，被 CodeBuddy 作为 workspace rules 加载到 agent 上下文中。因此对 AGENTS.md 的任何修改都会直接影响 agent 的行为。迁移时需确保 `@` 引用在通过 CODEBUDDY.md 加载时仍然正确解析。

### 现有 Gardener Agent

`ai-metadata/backend/GARDENER.md` 已存在知识库维护逻辑（README 地图核对、代码链接检查、文档一致性检查）。新的 Spec Review Agent 不重复 Gardener 的工作，而是专注于 **AGENTS.md 结构合规性**，两者职责边界：

| 职责 | Spec Review Agent | Gardener Agent |
|------|-------------------|----------------|
| AGENTS.md 内容合规 | **负责** | 不涉及 |
| @ 引用路径有效性 | **负责** | 不涉及 |
| 知识归属判定 | **负责** | 不涉及 |
| 文档内容准确性 | 不涉及 | **负责** |
| 代码-文档一致性 | 不涉及 | **负责** |
| README 索引维护 | 不涉及 | **负责** |

**注意**：Gardener Agent 自身的扫描路径也需更新。当前 `ai-metadata/backend/GARDENER.md` 中的路径引用使用旧格式（`1-design`, `2-development` 等），迁移后需同步更新为 `backend/design/`, `backend/development/` 等新路径。此更新作为 Phase 3（修复引用）的一部分执行。

---

## 1. 知识分层规范

### 1.1 判定标准：规则 vs 知识

区分"通用规则"（保留在 AGENTS.md）和"模块知识"（移至 ai-metadata/）的测试：

> **规则测试**：如果删除这段内容，agent 会不会做出**错误行为**（不只是缺乏信息）？如果是 → 规则，保留在 AGENTS.md。如果只是缺乏上下文 → 知识，移至 ai-metadata/。

**示例**：
- "Never use `task regenerate-gql`" → 规则（删除后 agent 会执行错误命令）
- "Use sqlc for DB queries" → 知识（删除后 agent 只是不知道用什么工具，但不会做错事）
- "先提交子项目再提交根项目" → 规则（删除后 agent 会破坏 Git 工作流）
- "项目使用 Chi 框架" → 知识（删除后 agent 缺少上下文，但不会犯错）

### 1.2 根 AGENTS.md（`./AGENTS.md`）

**只允许包含**：
- Git 提交规则（commit 顺序、hook 规则）
- 路径规则（No Absolute Paths）
- 文档引用规范（Use @ References）
- Symlink / Single Source of Truth 规则
- Writing Rules（frontmatter 格式说明）
- Module Knowledge Map（子项目 + ai-metadata 索引）

**禁止包含**：
- 技术栈描述、架构说明、代码规范等模块知识
- 任何可以在 ai-metadata/ 中找到的详细内容

### 1.3 子项目 AGENTS.md（`./modelcraft-go/AGENTS.md`, `./modelcraft-front/AGENTS.md`）

**只允许包含**：
- 子项目特有的操作规则（通过 1.1 判定测试的规则）
- ai-metadata 知识索引（`@ai-metadata/` 引用列表）

**禁止包含**：
- 技术栈表格、架构分层图、核心原则列表等知识内容
- 任何重复 ai-metadata 中已有文档的内容

### 1.4 ai-metadata/{module}/ 目录

每个模块目录下包含完整的模块知识，按主题分子目录：

```
ai-metadata/
├── backend/           # Go 后端知识
│   ├── README.md      # 知识库总览 + 阅读顺序
│   ├── design/        # 设计理念（最高优先级）
│   ├── development/   # 开发规范
│   ├── testing/       # 测试策略
│   ├── deployment/    # 部署指南
│   └── tools/         # 工具手册
├── front/             # 前端知识
│   ├── style/         # 设计系统
│   └── development/   # 开发规范
└── (future modules)   # 新模块按相同模式添加
```

### 1.5 知识迁移规则

迁移内容判定标准：

| 内容类型 | 放在哪里 |
|---------|---------|
| 通用规则（Git、路径、引用） | 根 AGENTS.md |
| 子项目操作规则（lint 修复流程等） | 子项目 AGENTS.md |
| 技术栈、架构、设计理念 | ai-metadata/{module}/ |
| 代码规范、错误处理、测试 | ai-metadata/{module}/ |
| 部署、工具安装 | ai-metadata/{module}/ |

---

## 2. 变更清单

### 迁移顺序

```
Phase 1: 迁移知识 → 确保 ai-metadata/ 中内容完整
Phase 2: 精简 AGENTS.md → 移除知识内容，保留规则和索引
Phase 3: 修复引用 → 确保 @ 路径正确指向 ai-metadata/
Phase 4: 验证 → 运行 Spec Review Agent 确认合规
```

### 回滚方案

迁移只移动内容位置，不删除文件。如果出现问题，从 git 历史恢复 AGENTS.md 即可——ai-metadata/ 文件保持不变。

### 2.1 根 AGENTS.md（`./AGENTS.md`）

**移除**：
- Project Structure 段落中的详细描述文字

**新增**：
- Module Knowledge Map 索引段（纯链接 + 简短描述）

**保留不变**：
- Git Rules
- No Absolute Paths
- Use @ References
- Single Source of Truth
- Writing Rules

### 2.2 Backend AGENTS.md（`./modelcraft-go/AGENTS.md`）

**移除**：
- Agent 工作规则段中的"身份"声明（"你是 CodeBuddy..."）
- 核心原则段（项目定位、GraphQL 入口、数据库、认证等）
- 技术栈表格（| 语言 | Go 1.25.1 | ...）
- 架构分层表格（| 层 | 路径 | 职责 | ...）
- 完整文档表格

**保留（重写为简洁格式）**：
- Lint 修复流程: "运行 `just lint` 遇到问题时，先执行 `just lint-fix` 尝试自动修复，再重新运行 `just lint` 验证结果。"

**修复引用**：
- 所有 `@ai-metadata/2-development/...` 格式的旧路径 → `@ai-metadata/backend/development/...`
- 所有 `@ai-metadata/5-tools/...` 格式的旧路径 → `@ai-metadata/backend/tools/...`

**新增**：
- 简化的 @ 引用索引（指向 ai-metadata/backend/README.md）

### 2.3 Frontend AGENTS.md（`./modelcraft-front/AGENTS.md`）

**无结构变更**：文件结构已经是纯索引格式，不需要修改。

**上下文说明**：`@` 引用会将设计系统和开发规范文档加载到 agent 上下文中。这是预期行为——知识通过索引加载而非内联。前端 AGENTS.md 的 @ 引用列表是否需要精简（减少上下文占用）不在本 spec 范围内，可作为后续优化。

---

## 3. Spec Review Agent

### 3.1 位置

`.agents/skills/spec-review/SKILL.md`

遵循现有 skill 目录命名规范：与 `openspec-*` skills 同级放置，使用 kebab-case 命名。`

### 3.2 与 Gardener Agent 的关系

Spec Review Agent 专注于 **AGENTS.md 结构合规性**，不涉及内容准确性审核（由 Gardener 负责）。

### 3.3 触发条件

- 修改任何 AGENTS.md 文件后
- 新增或修改 ai-metadata/ 下的知识文件后
- 新增模块目录到 ai-metadata/ 后

### 3.4 检查项

#### 检查 1: AGENTS.md 纯净度

验证 AGENTS.md 是否只包含允许的内容：

- 根 AGENTS.md 不应包含技术栈、架构、代码规范等知识
- 子项目 AGENTS.md 不应包含重复 ai-metadata 中已有文档的内容
- 允许：操作规则（通过 1.1 判定测试通过的规则）
- 允许：地图索引（@ 引用列表）

#### 检查 2: 知识归属正确性

验证知识内容是否在 ai-metadata/ 下：

- 技术/架构/规范知识 → ai-metadata/{module}/（不在 AGENTS.md）
- 操作规则 → AGENTS.md（不在 ai-metadata/）

#### 检查 3: 引用完整性

验证所有 @ 引用指向实际存在的文件：

- AGENTS.md 中的 @ 引用是否有效（包括 @./ 和 @ai-metadata/ 路径）
- 检查引用路径是否使用相对路径
- 检测过时的引用格式（如旧的编号前缀路径 `@ai-metadata/2-development/`）

#### 检查 4: 知识覆盖度

对比代码结构和 ai-metadata 文档：

- 检测新增的代码目录/模块是否有对应文档
- 检测 ai-metadata 中的文档引用是否有对应源码
- 标记可能的文档缺失

#### 检查 5: README 索引

验证每个 ai-metadata/{module}/ 目录的 README.md：

- 是否存在
- 是否包含文档索引（所有子目录/文件的列表）
- 是否说明阅读顺序和优先级

#### 检查 6: 孤立文件

检测 ai-metadata/ 下未被任何 AGENTS.md 或 README.md 引用的文件，标记可能的孤立文档。

### 3.5 输出格式

```
## Spec Review Report

### [PASS/FAIL] AGENTS.md 纯净度
- ./AGENTS.md: OK
- ./modelcraft-go/AGENTS.md: ISSUE - 包含技术栈表格（应移至 ai-metadata/backend/）

### [PASS/FAIL] 知识归属
- ./modelcraft-go/AGENTS.md: 3 个知识段需要迁移

### [PASS/FAIL] 引用完整性
- Broken: @ai-metadata/backend/tools/missing.md (文件不存在)
- Stale: @ai-metadata/2-development/architecture.md (旧路径格式)

### [PASS/FAIL] 知识覆盖度
- Missing: internal/domain/newmodule/ 无对应文档

### [PASS/FAIL] README 索引
- ./ai-metadata/backend/README.md: OK

### [PASS/FAIL] 孤立文件
- ./ai-metadata/front/development/bff-design.md (未被引用)
```

### 3.6 触发方式

作为 skill 被 CodeBuddy agent 调用，或在 pre-commit hook 中自动触发。

### 3.7 Hook 兼容性

现有 `.agents/hooks/check-documentation.py` 的 `ALLOWED_MD_DIRS` 不包含项目根目录。这意味着编辑已有的 AGENTS.md 不受影响（文件已存在），但如果需要为新子项目创建 AGENTS.md，需要将根路径添加到允许列表。
