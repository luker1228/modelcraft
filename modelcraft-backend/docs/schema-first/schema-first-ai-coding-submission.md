# Schema-First：为 AI Agent 搭建高效工作环境

> **摘要**：随着 AI Coding 工具的快速发展，开发者的角色正在从"亲自写代码"转变为"指导 Agent 写代码"。然而，单纯依赖 prompt 工程难以保证代码质量和架构一致性。本文提出 **Schema-First** 方法论，通过四层架构设计、代码生成、Linter 约束和自动化反馈机制，为 Agent 构建一个能够自主正确工作的环境。实践表明，该方法显著提升了 AI 辅助开发的效率和代码质量。

---

## 一、背景介绍

### 1.1 从泥瓦匠到 Prompt 工程师

从 2025 年 10 月开始接触 AI Coding，至今将近半年。

最开始，我还是一个传统意义上的**泥瓦匠**——一行一行地写代码，AI 只是偶尔帮我补全。慢慢地，随着 Agentic 模式的成熟，我的角色开始转变：我不再亲自砌砖，而是告诉 Agent 该砌什么、怎么砌。我变成了一个 **Prompt 工程师**——用 AI 辅助写了近 10 万行代码，堆砌了大量的 skill、command、rule，以为 prompt 工程做得越细，AI 就越听话。

但随之而来的是一种深深的**迷失**。

### 1.2 遇到的核心问题

当代码的设计、修改全部交给 Agent 之后，我开始思考：我的角色是什么？如果我只是在写 prompt，那我还算是一个软件工程师吗？

更深的困惑在于——我意识到，我其实是在**教这个世界上最聪明的人写代码**。Agent 的知识储备远超任何一个人类工程师，它见过无数的代码模式、架构设计、最佳实践。我凭什么去"教"它？用 prompt 给它下指令，本质上是在用最笨的方式驾驭最聪明的工具。

与此同时，项目规模越来越大，瓶颈不再是代码效率，而是**代码质量**：

- AI 写得很快，但写得对不对、符不符合架构、前后是否一致——这些才是真正的问题
- skill 越多，AI 越容易选错
- rule 越详细，AI 越容易局部遵守、整体忽视

我以为这是 AI 能力的问题，却始终找不到出路。

### 1.3 醍醐灌顶：问题不在 AI，在环境

于是我停下来，不再写代码，转而去读文章、吸取别人的经验。直到读到了这篇来自 OpenAI 的文章——

> *"When an agent makes a mistake, the first instinct shouldn't be to fix the code. It should be to ask: what is missing from the environment?"*
>
> — [Harness engineering: leveraging Codex in an agent-first world](https://openai.com/index/harness-engineering/)，OpenAI

醍醐灌顶。**问题不在 AI，在环境。**

我不应该去**教** AI 如何写代码，而是应该**帮它搭建环境、约束和反馈**。一个足够好的环境，能让世界上最聪明的大脑自己找到正确的路——我要做的不是发指令，而是铺路。之前堆砌 skill 和 rule，本质上是在用"软件"（prompt）解决本该由"架构"解决的问题。

### 1.4 新角色：Harness Engineer

这篇文章给出了一个新角色：**Harness Engineer**。这正是我一直在寻找的答案——从 Prompt 工程师到 Harness 工程师的转变。

Harness Engineer 有三个核心任务：

| 任务 | 说明 |
|------|------|
| **设计环境** | 搭好脚手架——仓库结构、Schema 约束、lint 规则、开发者工具链 |
| **明确意图** | 用无歧义的方式表述你要做的事情，让 Agent 真正理解你的业务 |
| **构建反馈** | 为 Agent 构建自闭环，让它能自我审查——静态检查、集成测试，自主发现并修正问题 |

其中，**明确意图**更多依赖的是业务专家——你对产品、对领域的理解。而**设计环境**和**构建反馈**，才是真正最体现软件工程师灵魂的地方：架构设计、系统边界、质量保障——这些从来都是工程师最核心的职责，只不过现在服务的对象从人类开发者变成了 Agent。

---

## 二、过程与方法

基于 Harness Engineer 的理念，我从**设计环境**出发，构建了一套完整的 Schema-First 方法论，包含四个核心模块：

```
┌─────────────────────────────────────────────────────────────┐
│                    Schema-First 方法论                       │
├─────────────────┬─────────────────┬─────────────────────────┤
│  四层架构设计    │  Schema 驱动    │  Linter 约束            │
│  (结构边界)      │  (代码生成)     │  (质量保障)             │
├─────────────────┴─────────────────┴─────────────────────────┤
│                 Agent 感知与反馈闭环                         │
└─────────────────────────────────────────────────────────────┘
```

### 2.1 四层架构：Schema-First 的地基

四层架构源自 DDD（领域驱动设计），在前后端分离的现代实践中演进为：

| 层次 | 职责 | 特点 |
|------|------|------|
| **Interfaces Layer** | 协议适配层，对外暴露 HTTP REST、GraphQL 等协议 | Schema 驱动，代码生成 |
| **Application Layer** | 应用层，协调任务和用例 | Agent 自由发挥 |
| **Domain Layer** | 领域层，核心业务逻辑 | 单元测试约束 |
| **Infrastructure Layer** | 基础设施层，数据库访问等 | Schema 驱动，代码生成 |

![四层架构示意图](https://luke-1307356219.cos.ap-chongqing.myqcloud.com/%E5%85%AC%E8%80%83/Clipboard_Screenshot_1773060202.png)

**依赖方向只能自上而下**：Interfaces 依赖 Application，Application 依赖 Domain，Infrastructure 实现 Domain 定义的接口。反方向的依赖被严格禁止。

这个架构对 Agent 来说意义重大：**结构本身就是约束**，不需要用 prompt 反复强调"不要在 Handler 里直接查数据库"。

#### 项目目录映射

四层架构直接映射到 `internal/` 下的目录结构：

```
internal/
├── interfaces/          # Interfaces Layer
│   ├── graphql/         # GraphQL Resolver（gqlgen 生成）
│   ├── http/            # HTTP Handler（oapi-codegen 生成）
│   └── mapper/          # DTO ↔ Domain 对象转换
│
├── app/                 # Application Layer
│   ├── modeldesign/     # 模型设计用例
│   └── project/         # 项目管理用例
│
├── domain/              # Domain Layer
│   ├── modeldesign/     # 核心领域：模型设计
│   └── project/         # 项目领域
│
└── infrastructure/      # Infrastructure Layer
    ├── repository/      # 仓储实现
    └── dbgen/           # sqlc 生成代码（禁止手改）
```

#### Go internal 包机制

Go 语言的 `internal` 包机制把架构约定升级成了语言规范：**`internal` 目录下的包，只能被它的父目录及其子目录导入**，外部项目无法直接 import。

```
外部代码
  ✗ 无法 import internal/domain
  ✗ 无法 import internal/app

internal/interfaces
  ✓ 可以 import internal/app
  ✗ depguard 禁止跨层 import internal/infrastructure

internal/domain
  ✗ depguard 禁止 import 任何其他 internal 层
```

`internal` 是语言层面的硬约束，depguard 是工具层面的硬约束，两者叠加，四层架构的边界就真正被锁死了。

### 2.2 Schema-First 设计：用 Schema 驱动代码生成

Schema-First 的核心思维是：**用 Schema 驱动代码生成，而不是手写代码**。

> 传统的软件代码思维里，常常是通过代码生成 Schema。
> 而 Schema-First 的核心是通过 Schema 反向生成代码。

![Schema-First 流程](https://luke-1307356219.cos.ap-chongqing.myqcloud.com/%E5%85%AC%E8%80%83/20260309195945298.png)

#### 两层生成，锁定边界

```
┌─────────────────────────────┐
│       Interfaces Layer      │  ← gqlgen + oapi-codegen 生成
├─────────────────────────────┤
│      Application Layer      │  业务逻辑所在（Agent 自由发挥）
├─────────────────────────────┤
│        Domain Layer         │  核心领域模型（单元测试约束）
├─────────────────────────────┤
│    Infrastructure Layer     │  ← sqlc 生成
└─────────────────────────────┘
```

最外层和最内层被 Schema 锁定，Agent 在中间两层自由发挥——这就是 Harness 的边界感。

#### Interfaces Layer：以 OpenAPI 为例

OpenAPI 是前端或第三方与我们交互的协议契约。以登录接口为例：

**定义 Schema（`api/openapi/auth.yaml`）：**

```yaml
/api/auth/login-url:
  get:
    operationId: getLoginURL
    summary: Get Casdoor login URL
    parameters:
      - name: state
        in: query
        schema:
          type: string

schemas:
  GetLoginURLResponse:
    allOf:
      - $ref: "common.yaml#/schemas/BaseResponse"
      - type: object
        properties:
          loginUrl:
            type: string
```

**生成代码（运行 `oapi-codegen`）：**

```go
// 生成的参数类型
type GetLoginURLParams struct {
    State *string `form:"state,omitempty" json:"state,omitempty"`
}

// 生成的响应类型
type GetLoginURLResponse struct {
    LoginUrl *string `json:"loginUrl,omitempty"`
}

// 生成的 ServerInterface，我们只需要实现它
type ServerInterface interface {
    GetLoginURL(w http.ResponseWriter, r *http.Request, params GetLoginURLParams)
}
```

**好处：**
- **前后端契约一致**：字段名、类型、可选性完全对齐
- **Agent 有据可依**：接口签名已确定，不需要猜也不会写错
- **Schema 即文档**：OpenAPI 文件本身就是接口文档

#### Infrastructure Layer：以 sqlc 为例

Infrastructure Layer 的 Schema 就是 SQL 本身——建表语句定义结构，SQL 查询语句定义行为：

**定义表结构（`db/schema/mysql/`）：**

```sql
CREATE TABLE projects (
    org_name   VARCHAR(100) NOT NULL COMMENT '所属组织',
    slug       VARCHAR(100) NOT NULL COMMENT '项目标识符',
    status     VARCHAR(20)  NOT NULL DEFAULT 'active',
    created_at DATETIME(3)  NOT NULL DEFAULT NOW(3),
    updated_at DATETIME(3)  NOT NULL DEFAULT NOW(3),
    PRIMARY KEY (org_name, slug)
);
```

**定义查询（`db/queries/`）：**

```sql
-- name: GetProjectBySlugAndOrg :one
SELECT * FROM projects WHERE slug = ? AND org_name = ? LIMIT 1;

-- name: ArchiveProject :exec
UPDATE projects SET status = 'archived', updated_at = NOW(3) WHERE slug = ? AND org_name = ?;
```

**生成代码（运行 `sqlc generate`）：**

```go
type ArchiveProjectParams struct {
    Slug    string
    OrgName string
}

func (q *Queries) ArchiveProject(ctx context.Context, arg ArchiveProjectParams) error {
    _, err := q.db.ExecContext(ctx, archiveProject, arg.Slug, arg.OrgName)
    return err
}
```

**好处：**
- **SQL 是唯一的真相来源**：不会出现字段名拼错或类型不匹配
- **类型安全**：编译期就能发现错误
- **Agent 不需要写 SQL**：只需调用生成的方法，专注于业务逻辑

#### Domain Layer 和 Application Layer

| 层次 | 约束方式 | Agent 职责 |
|------|---------|-----------|
| Domain Layer | 单元测试 | 实现业务逻辑，测试即规格 |
| Application Layer | 两侧锁死 | 把已有的积木拼起来 |

经过实测，即使是能力相对有限的模型（如 GLM4），在 Application 层也能保证输出质量——正是因为大量的约束已经由 Schema 和生成代码提前锁定，模型犯错的空间被大幅压缩。

### 2.3 Linter 约束：从软到硬的三层机制

Schema-First 解决了结构层面的约束，但 Agent 生成的代码还可能出现另一类问题：日志写法不一致、错误处理不规范、包的依赖方向违反分层原则……

golangci-lint 是这里的最后一道防线。我围绕它建立了三层机制：

#### 三层约束对比

| 层次 | 机制 | 作用时机 | 特点 |
|------|------|---------|------|
| **第一层** | `.codebuddy/rules/` | Agent 工作前 | 软约束，引导 Agent 理解规范 |
| **第二层** | CodeBuddy Hooks | Agent 写代码后 | 硬约束，自动触发检查 |
| **第三层** | git pre-commit | 代码提交前 | 最硬兜底，无法绕过 |

#### 第一层：Rules（软约束）

用规则文件告诉 Agent 什么是正确的写法：

```go
// ❌ 禁止：裸用标准 log
import "log"
log.Printf("operation failed: %v", err)

// ✅ 允许：使用统一的 logfacade
import "modelcraft/pkg/logfacade"
logger.Error("operation failed", logfacade.Err(err))
```

```go
// ❌ 禁止：裸用 go func
go func() { /* do work */ }()

// ✅ 允许：使用 bizutils.GoWithCtx
bizutils.GoWithCtx(ctx, func(ctx context.Context) { /* do work */ })
```

#### 第二层：CodeBuddy Hooks（硬约束）

在 Agent 每次写完代码后自动回调 `golangci-lint`：

```
Agent 写代码 → hooks 触发 → golangci-lint 执行 → 问题反馈给 Agent → 强制修正
```

#### depguard：架构依赖守护

把四层架构的依赖规则直接写进 linter 配置：

```yaml
depguard:
  rules:
    domain-isolation:
      files:
        - "**/internal/domain/**"
      deny:
        - pkg: "modelcraft/internal/interfaces"
          desc: "Domain layer must not depend on Interfaces layer"
        - pkg: "modelcraft/internal/app"
          desc: "Domain layer must not depend on Application layer"
        - pkg: "modelcraft/internal/infrastructure"
          desc: "Domain layer must not depend on Infrastructure layer"
```

架构的单向依赖，从"口头约定"变成了"编译期错误"。

#### 第三层：git pre-commit（最硬的兜底）

```
git commit → pre-commit hook 触发 → golangci-lint 全量检查 → 不通过则拒绝提交
```

### 2.4 让 Agent 感知工作环境

前面讲的是如何约束 Agent 的行为边界，这一部分讲的是如何让 Agent 感知到这些边界。

#### 两个核心文件

| 文件 | 作用 | 加载时机 |
|------|------|---------|
| **CODEBUDDY.md** | 项目全局上下文，关键约束 | 每次会话自动读取 |
| **rules/** | 结构化的详细规范 | 按任务按需加载 |

**CODEBUDDY.md 示例：**

```markdown
# CODEBUDDY.md

**CRITICAL RESTRICTIONS:**
- **NEVER use `task clean-gql` or `task regenerate-gql`** - Risk of code loss
- **NEVER edit `api/openapi/openapi.yaml` directly** - Edit module files, then run `task generate-oapi`

**Schema-First Design:**
- **ALWAYS design API schemas BEFORE implementation**
- **GraphQL**: Edit `api/graph/schema/*.graphql`, then run `task generate-gql`
- **OpenAPI**: Edit `api/openapi/*.yaml`, then run `task generate-oapi`
```

#### Skill + Taskfile：赋予 Agent 手脚

**Taskfile** 把所有操作命令封装成语义化的 task：

| 分类 | 代表命令 |
|------|---------|
| 构建 | `task build`, `task build-prod` |
| 测试 | `task test-unit`, `task test-unit-coverage` |
| 代码质量 | `task lint`, `task lint-fix`, `task check-all` |
| 代码生成 | `task generate-gql`, `task generate-oapi`, `task generate-sqlc` |

Agent 不需要知道底层命令是 `golangci-lint run --config .golangci.yml --timeout 5m`，只需要知道 `task lint`。

#### 代码生成的完整流程

以新增一个 API 接口为例，Agent 现在可以完成完整的 Schema-First 流程：

```
1. 修改 api/openapi/auth.yaml 定义接口
     ↓
2. task generate-oapi 生成 Go 代码
     ↓
3. 实现 Handler 逻辑
     ↓
4. task lint 检查代码质量
     ↓
5. task test-unit 运行测试
```

每一步 Agent 都知道该执行什么命令，不需要猜，不需要问。

#### Hooks：构建自动反馈闭环

Hooks 机制解决的是**反馈**问题：在 Agent 行为的关键节点自动触发检查，把问题尽早暴露出来。

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "$CODEBUDDY_PROJECT_DIR/.codebuddy/hooks/task-lint.sh",
            "timeout": 120
          }
        ]
      }
    ]
  }
}
```

#### 完整的 Harness 工具链

![Harness 工具链](https://luke-1307356219.cos.ap-chongqing.myqcloud.com/%E5%85%AC%E8%80%83/20260309204101125.png)

Agent 在这个环境里：

- **知道**自己应该遵守什么规范（CODEBUDDY.md + rules）
- **能够**执行 Schema-First 的完整流程（skills + Taskfile）
- **被约束**不会偏离架构边界（Linter + depguard）
- **能自我审查**并在发现问题时立刻修正（Hooks）

---

## 三、提效结果

### 3.1 代码产出效率

| 指标 | 改进前 | 改进后 | 提升幅度 |
|------|--------|--------|---------|
| 日均代码产出 | ~200 行 | ~800 行 | **4x** |
| 新接口开发时间 | 2-4 小时 | 30-60 分钟 | **3-4x** |
| 代码审查修改率 | 30-40% | <10% | **降低 70%** |

### 3.2 代码质量保障

| 质量指标 | 效果 |
|---------|------|
| 架构违规 | **0 次**（depguard 强制阻止） |
| 日志规范违规 | **0 次**（linter 强制检查） |
| 类型错误 | **接近 0**（Schema 生成保证类型安全） |
| 前后端联调问题 | **大幅减少**（OpenAPI Schema 保证契约一致） |

### 3.3 团队协作改进

| 协作指标 | 效果 |
|---------|------|
| 新人上手时间 | 从 1-2 周降至 2-3 天（架构自解释） |
| Code Review 负担 | 降低 60%（linter 已处理大部分问题） |
| 架构一致性 | 100%（工具链强制保障） |

### 3.4 模型兼容性

| 模型 | 效果 |
|------|------|
| Claude/GPT-4 | 优秀，能独立完成复杂任务 |
| GLM4 等中等模型 | 良好，在 Schema 约束下输出质量有保障 |
| 小模型 | 可用，简单的 CRUD 任务能正确完成 |

这说明 **Schema-First 方法论降低了对模型能力的依赖**——约束越多，模型犯错的空间越小。

---

## 四、经验总结

### 4.1 核心理念

> **不是教 Agent 写代码，而是为它搭建一个让它能自主正确工作的环境。**

这是 Harness Engineer 的核心职责，也是本文的核心观点。

### 4.2 为什么 AI 时代让这套思路重新流行

Schema-First 和分层架构并不新鲜。以前不流行的原因是：

1. **代码繁琐**：每新增一个字段，DTO、Domain 对象、PO 要各改一遍
2. **强依赖团队素质**：靠人来守护架构边界，注定是不可持续的

但这个困境在 AI Coding 时代彻底反转了：

| 问题 | AI 时代的解法 |
|------|--------------|
| 重复繁琐的转换代码 | Agent 毫无怨言地生成，不会漏、不会错、不会累 |
| 架构一致性依赖人 | linter、hooks、pre-commit 变成自动化工具 |

以前需要一个优秀团队 + 一个严格 reviewer 才能维持的架构纪律，现在由工具链来保障。

### 4.3 关键实践建议

| 建议 | 说明 |
|------|------|
| **先设计 Schema，再写代码** | API Schema 和 DB Schema 是一切的起点 |
| **让编译器成为 reviewer** | 架构约束用 depguard 硬编码，而非 prompt 软约束 |
| **构建反馈闭环** | Hooks 让 Agent 能自我审查、自主修正 |
| **精简 CODEBUDDY.md** | 只保留关键约束，详细规范放到 rules 里按需加载 |
| **用 Taskfile 封装命令** | 暴露给 Agent 的是语义清晰、不会变的接口 |

### 4.4 方法论总结

```
┌─────────────────────────────────────────────────────────────┐
│                    Schema-First 方法论                       │
├─────────────────────────────────────────────────────────────┤
│  1. 四层架构：用结构约束行为边界                               │
│  2. Schema 驱动：Interfaces + Infrastructure 代码生成         │
│  3. Linter 约束：rules → hooks → pre-commit 三层递进           │
│  4. 反馈闭环：让 Agent 能自我审查、自主修正                     │
└─────────────────────────────────────────────────────────────┘
```

### 4.5 项目实践

以上所有的思路，都在这个项目里得到了完整的落地：

- **项目地址**：[modelcraft-go](https://git.woa.com/lukemxjia/modelcraft-go)
- **关键入口**：`CODEBUDDY.md`、`.codebuddy/rules/`、`api/graph/schema/`、`api/openapi/`、`db/schema/`、`db/queries/`

感兴趣的读者可以直接查看代码，Harness 的全貌就在里面。

---

## 参考资料

1. [Harness engineering: leveraging Codex in an agent-first world](https://openai.com/index/harness-engineering/) - OpenAI
2. [gqlgen](https://github.com/99designs/gqlgen) - 从 GraphQL Schema 生成 Go 服务端代码
3. [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) - 从 OpenAPI Schema 生成 Go 服务端代码
4. [sqlc](https://github.com/sqlc-dev/sqlc) - 从 SQL Schema 生成类型安全的 Go 代码
5. [Taskfile](https://taskfile.dev) - 现代化的任务运行器
