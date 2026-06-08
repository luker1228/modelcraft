# Constrained GraphQL Runtime Patent Package Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the approved constrained-GraphQL-runtime spec into a concrete patent-preparation package consisting of an invention disclosure, draft claims, figure outline, and ModelCraft evidence map.

**Architecture:** Keep the approved design spec as the single source of truth, then derive four focused downstream artifacts from it. Write each artifact as a standalone document with a narrow purpose so legal review, technical review, and later claim revision can happen independently without rewriting the core idea from scratch.

**Tech Stack:** Markdown, `rg`, `sed`, `git`

---

## File Structure

- Reference: `docs/superpowers/specs/2026-06-07-constrained-graphql-runtime-patent-design.md`
  Responsibility: approved technical direction; do not fork the core concept outside this file during execution.

- Create: `docs/patents/2026-06-07-constrained-graphql-runtime-invention-disclosure.md`
  Responsibility: attorney-facing invention disclosure covering problem, solution, novelty, technical effect, embodiments, and terminology.

- Create: `docs/patents/2026-06-07-constrained-graphql-runtime-claims-draft.md`
  Responsibility: first-pass independent and dependent claim language derived from the approved spec.

- Create: `docs/patents/2026-06-07-constrained-graphql-runtime-figures.md`
  Responsibility: drawing list, figure captions, and exact block/arrow contents for later diagram production.

- Create: `docs/patents/2026-06-07-constrained-graphql-runtime-modelcraft-evidence-map.md`
  Responsibility: map each claimed concept to concrete ModelCraft concepts, docs, and code anchors so later reviewers can validate enablement.

---

### Task 1: Create the invention disclosure draft

**Files:**
- Create: `docs/patents/2026-06-07-constrained-graphql-runtime-invention-disclosure.md`
- Reference: `docs/superpowers/specs/2026-06-07-constrained-graphql-runtime-patent-design.md`

- [ ] **Step 1: Create the patents directory**

Run:

```bash
mkdir -p docs/patents
```

Expected: command exits `0` and `docs/patents` exists.

- [ ] **Step 2: Draft the invention disclosure with the exact section structure below**

```md
# Constrained GraphQL Runtime Invention Disclosure

**Date:** 2026-06-07
**Source Spec:** `docs/superpowers/specs/2026-06-07-constrained-graphql-runtime-patent-design.md`
**Project:** ModelCraft

## 1. Invention Name

一种基于模型元数据或数据库表元数据和权限策略生成受约束 GraphQL 访问边界并形成查询修正闭环的数据访问控制方法

## 2. Technical Field

本发明涉及数据库访问控制、接口动态生成和计算机实现的数据安全技术，尤其涉及一种基于模型元数据或数据库表元数据并结合权限策略动态生成受约束 GraphQL 访问边界的方法。

## 3. Background

现有数据平台通常通过自由 SQL、近似自由查询 DSL 或粗粒度接口暴露数据访问能力。这类方案存在三个问题：

1. 自由查询表达容易带来 SQL 注入、越权字段访问和越权条件构造风险。
2. 现有权限控制常停留在表级、接口级或字段可见级，难以细化到“字段是否可过滤、可用哪些操作符、是否允许排序/聚合/关联展开”。
3. 在 AI 场景中，调用方通常不知道什么能查、怎么查、查错后如何修改，因此难以稳定生成合法请求。

## 4. Core Technical Problem

如何在保留现有 GraphQL 作为查询承载协议的前提下，不开放自由 SQL，同时基于模型定义或数据库真实表结构自动暴露受限数据访问能力，并在越界查询发生时返回可用于修正的语义反馈，以形成安全、细粒度、可收敛的数据访问闭环。

## 5. Core Technical Solution

本发明通过以下协同机制解决上述问题：

1. 获取模型元数据，或者直接读取数据库表、列、主键、索引和外键形成数据库表元数据；
2. 基于前述元数据生成目标数据对象对应的 GraphQL 基础访问能力；
3. 获取调用主体对应的权限策略，并将所述权限策略与前述元数据联合裁剪为主体专属受约束 GraphQL 访问边界；
4. 向调用方暴露所述访问边界，用于指导查询构造；
5. 当调用方提交超界 GraphQL 查询时，返回指示超界位置与超界原因的语义反馈；
6. 调用方依据语义反馈修正查询，直至获得合法请求并完成数据访问。

## 6. Distinguishing Points

与普通 GraphQL 平台相比，本发明的区别不在于 GraphQL 语法本身，而在于：

1. 访问边界按主体动态生成，而非固定 schema 全量暴露；
2. 元数据来源既可以是平台内部模型，也可以直接来自数据库真实表结构；
3. 约束细化到字段、条件、操作符、结果结构和关联访问层；
4. 系统提供的是查询修正闭环，而非仅在失败时返回泛化错误。

## 7. Technical Effects

1. 降低自由查询导致的注入与越权风险；
2. 实现比表级或字段可见级更细的访问控制；
3. 减少人工配置接口的成本；
4. 使 AI 能在现有 GraphQL 协议上收敛到合法查询。

## 8. Preferred Embodiments

### 8.1 Metadata-First Embodiment

平台维护模型元数据，并基于模型元数据生成 GraphQL 访问能力，再按权限策略裁剪。

### 8.2 Database-Introspection Embodiment

系统直接读取数据库表、列、主键、索引和外键信息，生成基础 GraphQL 访问能力，再按权限策略裁剪。

### 8.3 Hybrid Embodiment

平台语义化模型定义与数据库真实表元数据联合用于生成和完善 GraphQL 访问边界。

## 9. Non-Core But Useful Notes

1. GraphQL 的原生类型校验不是本发明的创新点。
2. AI 自动修正不是独立发明点，而是访问边界与语义反馈闭环带来的效果。
3. “安全执行”不是本发明唯一卖点，真正核心是边界生成与反馈闭环。
```

- [ ] **Step 3: Verify the disclosure file exists and contains the core terms**

Run:

```bash
rg -n "数据库表元数据|受约束 GraphQL 访问边界|查询修正闭环" docs/patents/2026-06-07-constrained-graphql-runtime-invention-disclosure.md
```

Expected: at least 3 matching lines, including one title line and one solution/body line.

- [ ] **Step 4: Commit the disclosure draft**

```bash
git add docs/patents/2026-06-07-constrained-graphql-runtime-invention-disclosure.md
git commit -m "docs: add constrained graphql runtime invention disclosure"
```

---

### Task 2: Draft the claim set

**Files:**
- Create: `docs/patents/2026-06-07-constrained-graphql-runtime-claims-draft.md`
- Reference: `docs/superpowers/specs/2026-06-07-constrained-graphql-runtime-patent-design.md`
- Reference: `docs/patents/2026-06-07-constrained-graphql-runtime-invention-disclosure.md`

- [ ] **Step 1: Draft the claims document with one method claim, one system claim, and dependent claims**

```md
# Constrained GraphQL Runtime Claims Draft

**Date:** 2026-06-07
**Project:** ModelCraft

## Independent Claim 1: Method

一种基于模型元数据或数据库表元数据和权限策略生成受约束 GraphQL 访问能力并形成查询修正闭环的数据访问控制方法，其特征在于，包括：

1. 获取目标数据对象的模型元数据，或者读取目标数据对象对应数据库表的表元数据，所述元数据至少包括字段定义、字段类型、对象关系以及字段对应的查询能力属性；
2. 获取与调用主体对应的权限策略，所述权限策略用于限定所述调用主体对目标数据对象的字段访问范围、查询操作范围和条件使用范围；
3. 基于所述模型元数据或所述表元数据生成目标数据对象对应的 GraphQL 访问能力，并基于所述权限策略对所述 GraphQL 访问能力进行裁剪，得到面向所述调用主体的受约束 GraphQL 访问边界；
4. 向调用方提供所述受约束 GraphQL 访问边界，用于指导调用方构造针对目标数据对象的 GraphQL 查询请求；
5. 在接收到所述 GraphQL 查询请求超出所述受约束 GraphQL 访问边界时，输出指示超界位置及超界原因的语义反馈信息，所述语义反馈信息用于指导调用方修正所述 GraphQL 查询请求；
6. 在接收到满足所述受约束 GraphQL 访问边界的 GraphQL 查询请求时，执行对应的数据访问操作并返回查询结果。

## Independent Claim 2: System

一种数据访问控制系统，其特征在于，包括：

1. 元数据获取模块，用于获取模型元数据，或者读取数据库表元数据；
2. 权限策略获取模块，用于获取调用主体对应的权限策略；
3. 访问边界生成模块，用于基于所述元数据和所述权限策略生成受约束 GraphQL 访问边界；
4. 反馈模块，用于在查询超界时输出指示超界位置及超界原因的语义反馈信息；
5. 数据访问模块，用于在查询满足所述访问边界时执行对应的数据访问操作并返回查询结果。

## Dependent Claims

2. 根据权利要求1所述的方法，其中，所述表元数据至少包括表名、列名、列类型、主键、唯一约束、索引和外键关系中的至少一种。

3. 根据权利要求1所述的方法，其中，所述查询能力属性至少包括字段是否允许返回、是否允许作为过滤条件、允许的过滤操作符、是否允许排序、是否允许聚合以及是否允许关联展开中的至少一种。

4. 根据权利要求1所述的方法，其中，所述受约束 GraphQL 访问边界根据主体身份、租户、项目、角色或资源上下文动态裁剪生成。

5. 根据权利要求1所述的方法，其中，所述语义反馈信息至少包括违规字段、违规条件、违规操作符、违规关联路径、缺失权限类型或可替代查询方式中的至少一种。

6. 根据权利要求1所述的方法，其中，所述调用方包括 AI 调用方，所述 AI 调用方基于所述语义反馈信息重构下一次 GraphQL 查询请求。

7. 根据权利要求1所述的方法，其中，所述元数据由平台维护的模型元数据与数据库表元数据联合确定。
```

- [ ] **Step 2: Verify the claim draft contains both metadata-source variants and both independent claims**

Run:

```bash
rg -n "数据库表元数据|Independent Claim 1|Independent Claim 2|受约束 GraphQL 访问边界" docs/patents/2026-06-07-constrained-graphql-runtime-claims-draft.md
```

Expected: matches for both independent-claim headings and multiple mentions of `数据库表元数据`.

- [ ] **Step 3: Run a consistency check between the spec title and claim title**

Run:

```bash
printf "SPEC:\n"; sed -n '1,40p' docs/superpowers/specs/2026-06-07-constrained-graphql-runtime-patent-design.md | rg "一种基于"
printf "\nCLAIMS:\n"; sed -n '1,80p' docs/patents/2026-06-07-constrained-graphql-runtime-claims-draft.md | rg "一种基于"
```

Expected: both outputs include `模型元数据或数据库表元数据` and `受约束 GraphQL 访问边界`.

- [ ] **Step 4: Commit the claim draft**

```bash
git add docs/patents/2026-06-07-constrained-graphql-runtime-claims-draft.md
git commit -m "docs: add constrained graphql runtime claims draft"
```

---

### Task 3: Create the figures and embodiment flow outline

**Files:**
- Create: `docs/patents/2026-06-07-constrained-graphql-runtime-figures.md`
- Reference: `docs/patents/2026-06-07-constrained-graphql-runtime-invention-disclosure.md`

- [ ] **Step 1: Draft the figure list and exact caption text**

```md
# Constrained GraphQL Runtime Figures Outline

**Date:** 2026-06-07

## Figure 1: Overall system architecture

Caption:
受约束 GraphQL 运行时数据访问控制系统总体架构示意图。

Blocks:
1. 调用方
2. 元数据获取模块
3. 权限策略获取模块
4. 访问边界生成模块
5. GraphQL 查询接口
6. 语义反馈模块
7. 数据访问模块
8. 底层数据库

## Figure 2: Boundary generation flow

Caption:
基于模型元数据或数据库表元数据并结合权限策略生成受约束 GraphQL 访问边界的流程示意图。

Flow:
1. 读取模型元数据或数据库表元数据
2. 提取字段、类型、关系和查询能力属性
3. 获取调用主体权限策略
4. 裁剪 GraphQL 基础访问能力
5. 生成主体专属访问边界

## Figure 3: Query correction closed loop

Caption:
查询超界后的语义反馈与修正闭环流程示意图。

Flow:
1. 调用方构造 GraphQL 查询
2. 判断是否超出访问边界
3. 若超界，返回超界位置与超界原因
4. 调用方依据反馈修正查询
5. 提交合法查询
6. 执行数据访问并返回结果

## Figure 4: Database-introspection embodiment

Caption:
直接读取数据库表、列、主键、索引和外键信息生成 GraphQL 基础访问能力的实施例示意图。

Flow:
1. 读取数据库 schema 信息
2. 生成表对应字段集合
3. 识别主键、唯一约束、索引和外键
4. 生成 GraphQL 基础访问能力
5. 结合权限策略形成受约束访问边界
```

- [ ] **Step 2: Verify the figure file covers both the closed loop and database-introspection embodiment**

Run:

```bash
rg -n "Figure 3|Figure 4|数据库|修正闭环|访问边界" docs/patents/2026-06-07-constrained-graphql-runtime-figures.md
```

Expected: matches for both `Figure 3` and `Figure 4`, plus the keywords `数据库` and `修正闭环`.

- [ ] **Step 3: Commit the figures outline**

```bash
git add docs/patents/2026-06-07-constrained-graphql-runtime-figures.md
git commit -m "docs: add constrained graphql runtime figure outline"
```

---

### Task 4: Create the ModelCraft evidence map

**Files:**
- Create: `docs/patents/2026-06-07-constrained-graphql-runtime-modelcraft-evidence-map.md`
- Reference: `docs/superpowers/specs/2026-06-07-constrained-graphql-runtime-patent-design.md`
- Reference: `graphify-out/GRAPH_REPORT.md`
- Reference: `ai-metadata/backend/design/core-principles.md`
- Reference: `ai-metadata/backend/development/developer-enduser-system.md`
- Reference: `ai-metadata/cli/README.md`

- [ ] **Step 1: Draft the evidence map with exact concept-to-project anchors**

```md
# Constrained GraphQL Runtime ModelCraft Evidence Map

**Date:** 2026-06-07

| Claimed Concept | ModelCraft Evidence | Why It Matters |
|---|---|---|
| Design-time / runtime separation | `ai-metadata/backend/design/core-principles.md` | Supports the runtime-boundary generation story instead of static CRUD endpoints. |
| Runtime GraphQL as the access protocol | `ai-metadata/backend/design/core-principles.md` | Confirms the system uses GraphQL rather than REST as the runtime carrier. |
| Developer / EndUser access split | `ai-metadata/backend/development/developer-enduser-system.md` | Supports主体差异化访问 and permission-driven boundary generation. |
| AI-usable data access path | `ai-metadata/cli/README.md` | Shows the project already values schema discovery and structured access for agents. |
| GraphQL schema & runtime communities | `graphify-out/GRAPH_REPORT.md` | Supports that the codebase already has distinct runtime and schema-related architecture hubs. |
| Model-driven runtime docs effort | `docs/superpowers/specs/2026-06-06-model-scoped-runtime-api-doc-design.md` | Supports the claim that runtime ability exposure and query guidance are project-native concerns. |

## Review Notes

1. This file is not the patent itself.
2. This file exists to help later reviewers prove that the application is grounded in a real system.
3. If later code anchors are needed, add them after targeted source inspection rather than guessing.
```

- [ ] **Step 2: Verify each referenced source exists**

Run:

```bash
test -f ai-metadata/backend/design/core-principles.md
test -f ai-metadata/backend/development/developer-enduser-system.md
test -f ai-metadata/cli/README.md
test -f graphify-out/GRAPH_REPORT.md
test -f docs/superpowers/specs/2026-06-06-model-scoped-runtime-api-doc-design.md
```

Expected: all commands exit `0`.

- [ ] **Step 3: Verify the evidence map mentions all four anchor categories**

Run:

```bash
rg -n "runtime|GraphQL|EndUser|AI|schema" docs/patents/2026-06-07-constrained-graphql-runtime-modelcraft-evidence-map.md
```

Expected: matches across the table rows and review notes.

- [ ] **Step 4: Commit the evidence map**

```bash
git add docs/patents/2026-06-07-constrained-graphql-runtime-modelcraft-evidence-map.md
git commit -m "docs: add constrained graphql runtime evidence map"
```

---

### Task 5: Final package review and handoff

**Files:**
- Modify: `docs/patents/2026-06-07-constrained-graphql-runtime-invention-disclosure.md`
- Modify: `docs/patents/2026-06-07-constrained-graphql-runtime-claims-draft.md`
- Modify: `docs/patents/2026-06-07-constrained-graphql-runtime-figures.md`
- Modify: `docs/patents/2026-06-07-constrained-graphql-runtime-modelcraft-evidence-map.md`

- [ ] **Step 1: Run a placeholder scan across the package**

Run:

```bash
rg -n "TODO|TBD|待定|占位|placeholder|implement later|fill in details" docs/patents
```

Expected: no output.

- [ ] **Step 2: Run a terminology consistency check**

Run:

```bash
rg -n "模型元数据或数据库表元数据|受约束 GraphQL 访问边界|查询修正闭环|语义反馈" docs/patents
```

Expected: all four terms appear in the disclosure and claims, and at least two of them appear in the figures or evidence map.

- [ ] **Step 3: Add a package index note at the top of the disclosure**

```md
> Package companions:
> - `docs/patents/2026-06-07-constrained-graphql-runtime-claims-draft.md`
> - `docs/patents/2026-06-07-constrained-graphql-runtime-figures.md`
> - `docs/patents/2026-06-07-constrained-graphql-runtime-modelcraft-evidence-map.md`
```

- [ ] **Step 4: Review `git diff` for the full package**

Run:

```bash
git diff -- docs/patents
```

Expected: only the four patent-package docs appear; no unrelated changes.

- [ ] **Step 5: Commit the reviewed package**

```bash
git add docs/patents
git commit -m "docs: finalize constrained graphql runtime patent package"
```

---

## Self-Review

### Spec Coverage

- Title and main claim direction: covered by Task 1 and Task 2.
- Database-table-metadata source as a core differentiator: covered by Task 1, Task 2, and Task 3 Figure 4.
- Boundary generation plus semantic-feedback closed loop: covered by Task 1, Task 2, and Task 3 Figure 3.
- ModelCraft grounding and enablement: covered by Task 4.

### Placeholder Scan

This plan contains no `TODO`, `TBD`, or deferred-content markers. Every task includes the exact file path, concrete text skeleton, and verification command.

### Type Consistency

The same four core phrases are used throughout the plan:

- `模型元数据或数据库表元数据`
- `受约束 GraphQL 访问边界`
- `查询修正闭环`
- `语义反馈`

No alternate naming is introduced for the same concept.
