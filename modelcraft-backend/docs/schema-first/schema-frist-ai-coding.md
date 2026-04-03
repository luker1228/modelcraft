# Schema-First：为 Agent 搭建工作环境

从 2025 年 10 月开始接触 AI Coding，至今将近半年。

最开始，我还是一个传统意义上的**泥瓦匠**——一行一行地写代码，AI 只是偶尔帮我补全。慢慢地，随着 Agentic 模式的成熟，我的角色开始转变：我不再亲自砌砖，而是告诉 Agent 该砌什么、怎么砌。我变成了一个 **prompt 工程师**——用 AI 辅助写了近 10 万行代码，堆砌了大量的 skill、command、rule，以为 prompt 工程做得越细，AI 就越听话。

但随之而来的是一种深深的**迷失**。

当代码的设计、修改全部交给 Agent 之后，我开始思考：我的角色是什么？如果我只是在写 prompt，那我还算是一个软件工程师吗？

更深的困惑在于——我意识到，我其实是在**教这个世界上最聪明的人写代码**。Agent 的知识储备远超任何一个人类工程师，它见过无数的代码模式、架构设计、最佳实践。我凭什么去"教"它？用 prompt 给它下指令，本质上是在用最笨的方式驾驭最聪明的工具。

与此同时，项目规模越来越大，瓶颈不再是代码效率，而是**代码质量**。AI 写得很快，但写得对不对、符不符合架构、前后是否一致——这些才是真正的问题。skill 越多，AI 越容易选错；rule 越详细，AI 越容易局部遵守、整体忽视。我以为这是 AI 能力的问题，却始终找不到出路。

于是我停下来，不再写代码，转而去读文章、吸取别人的经验。

直到读到了这篇来自 OpenAI 的文章——

> *"When an agent makes a mistake, the first instinct shouldn't be to fix the code. It should be to ask: what is missing from the environment?"*
>
> — [Harness engineering: leveraging Codex in an agent-first world](https://openai.com/index/harness-engineering/)，OpenAI

醍醐灌顶。问题不在 AI，在环境。

我不应该去**教** AI 如何写代码，而是应该**帮它搭建环境、约束和反馈**。一个足够好的环境，能让世界上最聪明的大脑自己找到正确的路——我要做的不是发指令，而是铺路。之前堆砌 skill 和 rule，本质上是在用"软件"（prompt）解决本该由"架构"解决的问题。

这篇文章给出了一个新角色：**Harness Engineer**。这正是我一直在寻找的答案——从 prompt 工程师到 Harness 工程师的转变。

Harness Engineer 有三个核心任务：

1. **设计环境**：搭好脚手架——仓库结构、Schema 约束、lint 规则、开发者工具链
2. **明确意图**：用无歧义的方式表述你要做的事情，让 Agent 真正理解你的业务
3. **构建反馈**：为 Agent 构建自闭环，让它能自我审查——静态检查、集成测试，自主发现并修正问题

其中，**明确意图**更多依赖的是业务专家——你对产品、对领域的理解。而**设计环境**和**构建反馈**，才是真正最体现软件工程师灵魂的地方：架构设计、系统边界、质量保障——这些从来都是工程师最核心的职责，只不过现在服务的对象从人类开发者变成了 Agent。

这篇文章正是从**设计环境**出发，展开两件具体的事：

1. **Schema-First 设计**：通过四层架构和代码生成，从结构上锁定 Agent 的工作边界
2. **Linter 约束**：Schema 管不到的地方，用静态检查从软到硬逐层兜底


## **Part 1：软件设计的 Schema-First**

### 四层架构：Schema-First 的地基

在讲 Schema-First 的具体设计之前，有必要先介绍它所依托的架构基础。

四层架构源自 DDD（领域驱动设计），最初的定义是：

- **User Interface Layer**：用户界面层，负责与用户交互
- **Application Layer**：应用层，协调任务和用例，不包含业务规则
- **Domain Layer**：领域层，系统的核心，包含业务逻辑和领域模型
- **Infrastructure Layer**：基础设施层，提供技术实现（数据库、消息队列等）

随着前后端分离的普及，第一层发生了演进。UI 层不再是服务端的职责，取而代之的是**协议适配层**——同一套业务逻辑，可能需要同时对外暴露 HTTP REST、GraphQL、Protobuf 等不同协议。这一层的本质从"渲染界面"变成了"适配协议"，因此更准确的叫法是 **Interfaces Layer**。

![](https://luke-1307356219.cos.ap-chongqing.myqcloud.com/%E5%85%AC%E8%80%83/Clipboard_Screenshot_1773060202.png)

**依赖方向只能自上而下**：Interfaces 依赖 Application，Application 依赖 Domain，Infrastructure 实现 Domain 定义的接口。反方向的依赖被严格禁止——Domain 层尤其纯粹，它不知道自己是被 HTTP 调用的还是被 GraphQL 调用的，也不知道数据存在 MySQL 还是 PostgreSQL。

这个架构对 Agent 来说意义重大。它不仅是人的编码规范，更是 Agent 的行为边界——结构本身就是约束，不需要用 prompt 反复强调"不要在 Handler 里直接查数据库"。

### 项目实现举例

四层架构直接映射到 `internal/` 下的目录结构：

```
internal/
├── interfaces/          # Interfaces Layer
│   ├── graphql/         # GraphQL Resolver（gqlgen 生成）
│   ├── http/            # HTTP Handler（oapi-codegen 生成）
│   ├── mapper/          # DTO ↔ Domain 对象转换
│   └── runtime/         # 运行时协议适配
│
├── app/                 # Application Layer
│   ├── modeldesign/     # 模型设计用例
│   ├── project/         # 项目管理用例
│   └── ...
│
├── domain/              # Domain Layer
│   ├── modeldesign/     # 核心领域：模型设计
│   ├── project/         # 项目领域
│   ├── shared/          # 跨域共享
│   └── ...
│
└── infrastructure/      # Infrastructure Layer
    ├── repository/      # 仓储实现
    ├── dbgen/           # sqlc 生成代码（禁止手改）
    └── database/        # 数据库连接、迁移
```

### Go internal 包机制

这个目录结构能真正发挥约束作用，离不开 Go 语言的 **`internal` 包机制**。

为什么要用 `internal`？普通目录也可以放代码，但没有任何访问限制——任何人都可以 import，包括 Agent。如果四层架构只是目录命名约定，Agent 完全可以在 Domain 层里直接 import Infrastructure 的包，编译器不会报错，人工 review 也可能漏掉。

`internal` 把这个约定升级成了语言规范：**`internal` 目录下的包，只能被它的父目录及其子目录导入**，外部项目无法直接 import。

这带来了两个好处：

1. **对外隐藏实现细节**：整个四层架构是内部实现，外部无法绕过 Interfaces 层直接调用 Domain
2. **结合 depguard 强化分层约束**：`internal` 机制挡住了外部，depguard 则进一步约束内部各层之间的依赖方向

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

> 如何让项目长期维护这个结构而不被破坏？答案在后面的 [depguard 章节](#depguard架构依赖守护)。

Schema-First 的设计，正是沿着这个架构展开的。这是Schema-First设计的雏形，即利用工具，测试，linter，编译器来约束环境

### Schema-First 设计

Schema-First 的核心思维是：**用 Schema 驱动代码生成，而不是手写代码**。

> 传统的软件代码思维里，常常是通过代码生成schema。
> 而Schema-First的核心是通过schema反向生成代码

![](https://luke-1307356219.cos.ap-chongqing.myqcloud.com/%E5%85%AC%E8%80%83/20260309195945298.png)

在四层架构里，有两层天然适合这种思维：

- **Interfaces Layer**：API 契约是公开的、结构化的，完全可以从 Schema 生成
- **Infrastructure Layer**：数据库表结构是确定的，ORM 模型和查询代码可以从 Schema 生成

恰好有三个库完美契合这两层：

**Interfaces Layer：**
- [gqlgen](https://github.com/99designs/gqlgen) — 从 GraphQL Schema 生成 Go 服务端代码
- [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) — 从 OpenAPI Schema 生成 Go 服务端代码

**Infrastructure Layer：**
- [sqlc](https://github.com/sqlc-dev/sqlc) — 从 SQL Schema 和 SQL 查询语句生成类型安全的 Go 代码

三个库的共同理念一致：**写 Schema，生成代码，禁止手改生成文件**。

#### 两层生成，锁定边界

```
┌─────────────────────────────┐
│       Interfaces Layer      │  ← gqlgen + oapi-codegen 生成
├─────────────────────────────┤
│      Application Layer      │  业务逻辑所在
├─────────────────────────────┤
│        Domain Layer         │  核心领域模型
├─────────────────────────────┤
│    Infrastructure Layer     │  ← sqlc 生成
└─────────────────────────────┘
```

最外层和最内层被 Schema 锁定，Agent 在中间两层自由发挥——这就是 Harness 的边界感。

#### Interfaces Layer：以 OpenAPI 为例

OpenAPI 是前端或第三方与我们交互的协议契约。Schema-First 的起点就从这里开始。

以登录接口为例，我们先在 `api/openapi/auth.yaml` 里定义：

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

运行 `oapi-codegen --config api/openapi/oapi-codegen.yaml api/openapi/openapi.yaml` 之后，自动生成：

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
    // ...
}
```

**这样做的好处：**

- **前后端契约一致**：前端拿到同一份 OpenAPI Schema 生成 SDK，字段名、类型、可选性完全对齐，不会出现联调时对不上的问题
- **Agent 有据可依**：Agent 实现 Handler 时，接口签名已经确定，参数类型已经生成，不需要猜也不会写错
- **重构不破坏契约**：内部实现随便改，只要 `ServerInterface` 还能编译通过，对外协议就没有变化
- **Schema 即文档**：OpenAPI 文件本身就是接口文档，不需要额外维护

#### Infrastructure Layer：以 sqlc 为例

Infrastructure Layer 的 Schema 就是 SQL 本身——建表语句定义结构，SQL 查询语句定义行为，两者共同驱动代码生成。

以项目表为例，先在 `db/schema/mysql/` 里定义表结构：

```sql
CREATE TABLE projects (
    org_name   VARCHAR(100) NOT NULL COMMENT '所属组织',
    slug       VARCHAR(100) NOT NULL COMMENT '项目标识符',
    status     VARCHAR(20)  NOT NULL DEFAULT 'active' COMMENT '项目状态：active/archived',
    created_at DATETIME(3)  NOT NULL DEFAULT NOW(3),
    updated_at DATETIME(3)  NOT NULL DEFAULT NOW(3),
    PRIMARY KEY (org_name, slug)
);
```

再在 `db/queries/` 里定义查询：

```sql
-- name: GetProjectBySlugAndOrg :one
SELECT * FROM projects
WHERE slug = ? AND org_name = ?
LIMIT 1;

-- name: ArchiveProject :exec
UPDATE projects
SET status = 'archived', updated_at = NOW(3)
WHERE slug = ? AND org_name = ?;
```

运行 `sqlc generate` 之后，自动生成：

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

**这样做的好处：**

- **SQL 是唯一的真相来源**：表结构、字段名、类型全部在 SQL 里定义，Go 代码从 SQL 生成，不会出现字段名拼错或类型不匹配
- **类型安全**：参数结构体由 sqlc 生成，编译期就能发现错误，不需要运行时才暴露问题
- **Agent 不需要写 SQL**：查询语句已经写好，Agent 只需要调用生成的方法，专注于业务逻辑
- **无 ORM 魔法**：SQL 完全可读、可审查，没有 ORM 隐式行为带来的不确定性

#### Domain Layer：唯一的约束是单元测试

Interfaces 和 Infrastructure 两层被 Schema 和代码生成锁定了。但 Domain Layer 不同——它是整个项目的核心，没有任何外部依赖，也没有 Schema 可以驱动生成。

这里没有框架约束，没有生成代码，Agent 拥有最大的自由度。而自由度越大，出错的空间也越大。

对 Domain Layer 最有效的约束，是**单元测试**。

- 没有数据库依赖，测试可以完全内存运行，速度极快
- 测试即规格——Agent 看到测试用例，就知道这段逻辑的期望行为是什么
- TDD 在这里不是仪式，而是 Harness 的一部分：先写测试，给 Agent 一个明确的靶心，再让它实现

```
Domain Layer
    ↑
单元测试（*_test.go）  ← 这里是 Harness 对 Domain 的唯一锁定
```

没有测试覆盖的 Domain 逻辑，就是 Harness 里的一个漏洞。

#### Application Layer：模型自由发挥的地方

Application Layer 是四层里唯一没有 Schema 约束、也没有代码生成的层——它的工作就是**调包侠**：编排 Domain 层的业务逻辑和 Infrastructure 层的数据访问，组合成一个完整的用例。

```go
func (uc *CreateProjectUseCase) Execute(ctx context.Context, input CreateProjectInput) error {
    // 调用 Domain 层：业务规则校验
    project, err := uc.projectService.NewProject(input.OrgName, input.Slug)
    if err != nil {
        return err
    }
    // 调用 Infrastructure 层：持久化
    return uc.projectRepo.Save(ctx, project)
}
```

这一层看起来最"自由"，但实际上它的边界已经被两侧锁死：

- 能调用什么方法？——由 Domain 层的接口定义决定
- 能存取什么数据？——由 sqlc 生成的方法决定
- 能接收什么参数、返回什么结果？——由 Interfaces 层的生成类型决定

Agent 在这里的工作高度确定：**把已有的积木拼起来**。没有需要发明的东西，没有需要猜的字段，只有组合。

经过实测，即使是能力相对有限的模型（如 GLM4），在这一层也能保证输出质量——正是因为大量的约束已经由 Schema 和生成代码提前锁定，模型犯错的空间被大幅压缩。

### 状态的变化与转移：软件设计的核心

回顾四层架构和 Schema-First 的设计，其实揭露了一个软件设计更本质的规律：**软件的本质是状态的变化与转移**。

一个请求从进入系统到落库，经历的是一条完整的状态转移链：

```
OpenAPI Schema
     ↓  (oapi-codegen 生成)
DTO 对象（接口层，面向协议）
     ↓  (mapper 转换)
Domain 对象（领域层，面向业务）
     ↓  (mapper 转换)
PO 对象（持久化对象，面向存储）
     ↓  (sqlc 生成)
数据库 DB Schema
```

每一层的对象形态不同，因为它服务的关注点不同：DTO 关心的是协议兼容，Domain 对象关心的是业务规则，PO 关心的是存储效率。强行用一个对象贯穿所有层，表面上省事，实际上是把不同层的关注点混在一起，最终让任何一层都无法单独演进。

Schema-First 的价值正在于此——**每一层的 Schema 都是对这一层状态的精确描述**，层与层之间通过 mapper 显式转换。状态在哪里、长什么样、怎么流转，全部有据可查。

对 Agent 来说，这条链路意味着：每一步该做什么，输入是什么，输出是什么，都已经被 Schema 定义清楚了。它不需要理解整个系统，只需要在自己负责的那一段链路上正确地完成转换。

#### 为什么 AI 时代让这套思路重新流行

这套思路并不新鲜。为什么以前不流行？

表面原因是代码繁琐——每新增一个字段，DTO、Domain 对象、PO 要各改一遍，mapper 也要跟着改，人容易出错。但更深层的原因是：**这套思路强依赖团队素质，以及一个严格的 reviewer**。

每个人对分层的理解不同，每次 code review 的标准不同，新人来了要重新培训，项目紧了规范就松了。《人月神话》早就告诉我们：**你无法通过增加人手来保证一致性**，沟通成本会随团队规模指数级增长。靠人来守护架构边界，注定是不可持续的。

所以很多团队选择了妥协：用一个对象走天下，或者引入 ORM 的魔法映射，用便利性换走了清晰的边界。

但这个困境在 AI Coding 时代彻底反转了，而且是从两个维度同时反转：

**第一，重复繁琐的转换代码不再是负担。** 给 Agent 清晰的 Schema 定义，它可以毫无怨言地生成所有的转换代码，不会漏、不会错、不会累。以前压在人身上的那些枯燥工作，现在是 Agent 最擅长的事。

**第二，架构一致性不再依赖人来守护。** linter、hooks、pre-commit——这套机制把"严格的 reviewer"变成了自动化工具。它不会累，不会心软，不会因为项目赶进度就放水。每一次提交都经过同样严格的检查，跟团队规模无关，跟人的状态无关。

以前需要一个优秀团队 + 一个严格 reviewer 才能维持的架构纪律，现在由工具链来保障。这才是 Schema-First 在 AI 时代真正流行的原因。

而那个"严格的 reviewer"，在 AI 时代只需要一个 skill 就能解决。`backend-patterns` skill 把四层架构的依赖规则、API 设计规范、错误处理模式、日志模式、事务边界——所有人工 review 才能发现的问题——全部固化成 Agent 每次开始工作时能读到的上下文。它不会累，不会忘，不会因为今天心情不好就放水。

---

## Part 2：Linter 约束

### Linter 兜底：golangci-lint

Schema-First 解决了结构层面的约束，但 Agent 生成的代码还可能出现另一类问题：日志写法不一致、错误处理不规范、包的依赖方向违反分层原则……这些是 Schema 无法覆盖的细节。

golangci-lint 是这里的最后一道防线。我围绕它建立了三层机制，从软到硬逐层约束 Agent：

#### 第一层：rules（软约束）

用 `.codebuddy/rules/` 里的规则文件，告诉 Agent 什么是正确的写法：

- `code-style/coding-style.md` — 日志的最佳实践、错误处理的最佳实践
- `api-design/graphql-patterns.md` — GraphQL 错误类型设计规范

举几个典型的例子：

**日志：禁止裸用标准 `log`，必须使用统一的 `logfacade`**

```go
// ❌ 禁止
import "log"
log.Printf("operation failed: %v", err)

// ✅ 允许
import "modelcraft/pkg/logfacade"
logger.Error("operation failed", logfacade.Err(err))
```

**日志：堆栈跟踪只在错误转换边界使用，Service 层禁止**

```go
// ❌ 禁止：Service 层不应记录堆栈
func (s *Service) Create(ctx context.Context) error {
    logger.Error("failed", logfacade.Err(err), logfacade.Stack(err))
}

// ✅ 允许：错误转换为 GraphQL 类型前记录堆栈
func (r *resolver) CreateProject(ctx context.Context) (*Payload, error) {
    if err != nil {
        logger.Error("failed", logfacade.Err(err), logfacade.Stack(err))
        return &Payload{Error: toGraphQLError(err)}, nil
    }
}
```

**错误处理：禁止裸用标准 `errors`，必须使用 `pkg/bizerrors`**

```go
// ❌ 禁止
import "errors"
return nil, errors.New("invalid input")

// ✅ 允许
import pkgerrors "modelcraft/pkg/bizerrors"
return nil, pkgerrors.Wrapf(err, "get user %s", id)
```

**并发：禁止裸用 `go func`，必须使用 `bizutils.GoWithCtx`**

```go
// ❌ 禁止
go func() { /* do work */ }()

// ✅ 允许
bizutils.GoWithCtx(ctx, func(ctx context.Context) { /* do work */ })
```

这些规则 Agent 会主动读取并遵守。但"主动遵守"不够可靠，所以才有后面的 hooks 和 pre-commit 做强制兜底。

#### 第二层：CodeBuddy Hooks（硬约束）

利用 CodeBuddy 的 hooks 机制，在 Agent 每次写完代码后自动回调 `golangci-lint`：

```
Agent 写代码
     ↓
hooks 触发
     ↓
golangci-lint 执行
     ↓
fmt 格式统一 + 规范检查 + 静态分析
     ↓
问题反馈给 Agent，强制修正
```

这一层不依赖 Agent 的"自觉"，而是把检查嵌入到工作流里。格式问题、规范问题在会话内就被消灭，不会流出去。

golangci-lint 还支持自定义 linter，可以针对项目的特定规范做静态检查。

> 自定义 linter 的实现，参考：https://km.woa.com/group/51993/articles/show/623020

#### depguard：架构依赖守护

前面留了一个坑——如何让项目长期维护四层架构的单向依赖而不被破坏。答案就在 golangci-lint 的 **depguard** 功能里。

depguard 可以针对不同的包路径，精确控制它允许或禁止导入哪些包。利用这个机制，可以把四层架构的依赖规则直接写进 linter 配置，让编译工具链来强制执行：

```yaml
depguard:
  rules:
    # Domain 层：不允许依赖任何外层
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

    # Application 层：只允许依赖 Domain，不允许依赖 Interfaces 和 Infrastructure 的具体实现
    app-isolation:
      files:
        - "**/internal/app/**"
      deny:
        - pkg: "modelcraft/internal/interfaces"
          desc: "Application layer must not depend on Interfaces layer"

    # 禁止使用标准 log 包（已有）
    no-std-log:
      files:
        - "!**/main.go"
      deny:
        - pkg: "log"
          desc: "Use modelcraft/pkg/logfacade instead"
```

这样一来，如果 Agent 在 Domain 层里不小心 import 了 Infrastructure 的包，golangci-lint 会立刻报错——不需要人工 review，不需要靠规范文档约束，工具链直接拒绝。

架构的单向依赖，从"口头约定"变成了"编译期错误"。这是前面留的坑的答案。

#### 第三层：git pre-commit（最硬的兜底）

即使前两层都绕过了，git pre-commit hook 是最后一道门：提交前强制跑 `golangci-lint`，不通过不让提交。

```
git commit
     ↓
pre-commit hook 触发
     ↓
golangci-lint 全量检查
     ↓
不通过 → 拒绝提交
```

#### 从软到硬

这三层不是冗余，而是互补：

| 层次 | 机制 | 作用时机 | 特点 |
|------|------|---------|------|
| rules | `.codebuddy/rules/` | Agent 工作前 | 引导，依赖 Agent 理解 |
| hooks | CodeBuddy Hooks | Agent 写代码后 | 自动触发，实时反馈 |
| pre-commit | git hook | 代码提交前 | 强制兜底，无法绕过 |

Schema-First 定义了代码应该长什么样，这三层 linter 机制保证了代码真的长成那个样子。两者合在一起，构成了一个完整的 Harness。

还有第四点，关乎环境，关乎团队——**集成测试的部署**。Agent 写完代码，需要一个真实的环境来跑集成测试，验证各层之间的协作是否符合预期。这不只是 Agent 的问题，也是团队协作的基础设施问题。这里不展开，但它同样是 Harness 不可缺少的一环。

---

## Part 3：让 Agent 感知工作环境

接下来，就是给 Agent 装上手脚。前两部分讲的是如何通过 Schema 和 Linter 约束 Agent 的行为边界，这一部分讲的是如何让 Agent 感知到这些边界——让它知道自己在什么环境里工作、有哪些工具可用、应该遵守什么规范。

这里以 CodeBuddy CLI 为例，但核心思路适用于任何支持 memory 和 tool 机制的 AI Coding 工具（ps：虽然这个有些基础，但是这是整个环节必要的一环）。

### 两个核心文件：CODEBUDDY.md 和 rules

每次会话开启前，CodeBuddy 会自动加载两类文件：

1. **CODEBUDDY.md**（或 CLAUDE.md）：项目级的"记忆文件"，放在项目根目录，每次会话自动读取
2. **rules/**：结构化的规则文件，放在 `.codebuddy/rules/` 目录下，按需加载

这两个入口是 Agent 感知环境的起点。

#### CODEBUDDY.md：项目全局上下文

这个文件是 Agent 了解你项目的第一入口。它应该包含：

- **Schema-First 设计的关键约束**：哪些文件是自动生成的、禁止手动修改
- **关键命令入口**：告诉 Agent 用什么命令执行常见操作
- **架构概览**：四层结构、领域边界、API 协议分工

```markdown
# CODEBUDDY.md

**CRITICAL RESTRICTIONS:**
- **NEVER use `task clean-gql` or `task regenerate-gql`** - Risk of code loss
- **NEVER edit `api/openapi/openapi.yaml` directly** - Edit module files, then run `task generate-oapi`

**Schema-First Design:**
- **ALWAYS design API schemas BEFORE implementation**
- **GraphQL** (for business domain): Edit `api/graph/schema/*.graphql`, then run `task generate-gql`
- **OpenAPI** (for user/org management): Edit `api/openapi/*.yaml`, then run `task generate-oapi`

**Architecture:**
- Interfaces Layer: `internal/interfaces/`
- Application Layer: `internal/app/`
- Domain Layer: `internal/domain/`
- Infrastructure Layer: `internal/infrastructure/`
```

这个文件越精炼越好——Agent 每次都会读，太长反而会稀释重要信息。把详细规范放到 rules 里按需加载。随着模型越来越强，我经常需要删除这里的内容，现在只需要把project相关信息写在这里即可。

#### rules/：按场景加载的详细规范

rules 目录下的文件是分类的规范，Agent 根据当前任务按需读取：

```
.codebuddy/rules/
├── api-design/
│   ├── graphql-patterns.md      # GraphQL 错误类型、Payload 设计
│   └── openapi-patterns.md      # OpenAPI 响应格式、错误码设计
├── code-style/
│   ├── coding-style.md          # 日志规范、错误处理规范
│   └── repository-patterns.md   # Repository 层设计模式
└── ...
```

rules 是 Agent 的"扩展记忆"——需要的时候加载，不需要的时候不占用上下文。

### Skill + Taskfile：赋予 Agent 手脚

让 Agent 知道约束只是第一步，还要让它能够执行操作。这就需要把项目的工具链暴露给 Agent。

#### Taskfile：统一的命令入口

前面讲过 Taskfile 的作用——把所有操作命令封装成语义化的 task。对 Agent 来说，它不需要知道 `golangci-lint run --config .golangci.yml --timeout 5m`，只需要知道 `task lint`。

但光有 Taskfile 不够——Agent 需要知道这些 task 的存在。

#### taskfile skill：教会 Agent 使用工具

我为 Taskfile 写了一个专门的 skill

这样 Agent 就拥有了"手脚"——它可以通过 OpenAPI Schema 生成代码、通过 DB Schema 生成 sqlc 代码、通过 task lint 检查质量。

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

### Hooks：构建自动反馈闭环

让 Agent 知道工具在哪里只是单向的——Agent 执行了操作，但我们不知道它做得对不对。Hooks 机制解决的是**反馈**问题：在 Agent 行为的关键节点自动触发检查，把问题尽早暴露出来。

#### 为什么要尽早反馈？

Agent 的一个常见问题是：上下文越来越长，问题越积越多，到后面想修都不知道从哪下手。Harness 工程师的核心职责之一就是**构建自闭环**——让 Agent 能自我审查、自主修正。

Hooks 就是实现自闭环的关键机制。

#### PostToolUse：每次写文件后触发 Linter

CodeBuddy 支持 `PostToolUse` 和 `afterSearchRplaceFileEdit` 这两个 hook，在 Agent 编辑文件后触发。我配置了一个 hook，每次 Agent 修改 `.go` 文件后自动运行 `task-lint`：

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

### 完整的 Harness 工具链

把前面所有的部分拼在一起，一个完整的 Agent 工作环境就构建好了：

![](https://luke-1307356219.cos.ap-chongqing.myqcloud.com/%E5%85%AC%E8%80%83/20260309204101125.png)

Agent 在这个环境里：

- **知道**自己应该遵守什么规范（CODEBUDDY.md + rules）
- **能够**执行 Schema-First 的完整流程（skills + Taskfile）
- **被约束**不会偏离架构边界（Linter + depguard）
- **能自我审查**并在发现问题时立刻修正（Hooks）

这就是 Harness Engineer 的工作成果：**不是教 Agent 写代码，而是为它搭建一个让它能自主正确工作的环境。**

---

## Part 4：加餐

### Taskfile——把大脑从命令记忆中解放出来

记忆命令这件事，交给 Agent 来做比人类更合适。

我用的是 [Taskfile](https://taskfile.dev)（也可以用 Makefile，思路一样）——把项目所有的操作命令封装成语义化的 task，收敛到一个 `Taskfile.yml` 里统一管理。Makefile 是更传统的选择，几乎不需要额外安装；Taskfile 语法更现代，支持变量、依赖、跨平台，两者都能达到同样的目的。

整个项目的开发、测试、构建、部署、数据库管理，全部通过 task 入口：

| 分类 | 代表命令 |
|------|---------|
| 构建 | `task build`, `task build-prod` |
| 开发 | `task dev`, `task run`, `task restart` |
| 测试 | `task test-unit`, `task test-unit-pkg`, `task test-unit-coverage` |
| 代码质量 | `task lint`, `task lint-fix`, `task check-all` |
| 代码生成 | `task generate-gql`, `task generate-oapi`, `task generate-sqlc` |
| 数据库 | `task dbmigrate-up`, `task dbmigrate-create`, `task login-db` |
| 部署 | `task deploy-local`, `task deploy-docker`, `task docker-up` |
| 环境管理 | `task envswitch`, `task envlist`, `task envdiff` |

统一入口的好处不只是方便人——更重要的是对 Agent 友好。Agent 不需要知道底层命令是 `go build` 还是 `CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo ...`，它只需要知道"构建用 `task build`，跑测试用 `task test-unit`"。

命令的复杂度被封装在 Taskfile 里，暴露给 Agent 的是一套语义清晰、不会变的接口。这也是 Harness 工具可用性的一部分。

光有 Taskfile 还不够——Agent 需要知道这些 task 的存在。我为此写了一个 `taskfile` skill，进一步增强了智能化程度：

- **自动同步检查**：每次使用前先跑 `task --list | wc -l`，对比 skill 里记录的 task 数量，如果 `Taskfile.yml` 有新增，自动更新 skill 的引用文档，保证 Agent 的知识不过期
- **场景化工作流**：skill 里预定义了常见场景的命令组合，比如"提交前质量检查"直接给出 `task fmt && task vet && task lint`，Agent 不需要自己推断
- **关键限制内置**：把"永远不要跑 `task clean-gql`"这类危险操作直接写进 skill，Agent 调用时自动规避
- **参数提示**：部分 task 支持参数（如 `task run FORCE=true`、`task test-unit-pkg PKG=./path/to/pkg`），skill 里统一列明，Agent 不需要猜参数名

Taskfile 解决了"命令怎么跑"的问题，skill 解决了"Agent 怎么知道"的问题。两者配合，工具可用性才算真正闭环。

### Hooks——防止 Agent 乱写文件

Agent 有一个常见的坏习惯：动不动就创建 `.md` 文件。任务完成了写个总结、遇到问题写个说明、加个功能写个 README——没人要求它写，它自己就写了。

这类文件既没有用，又污染了项目目录。

CodeBuddy 的 hooks 机制可以在工具调用前拦截，用来做这类行为管控。我配置了一个 `PreToolUse` hook，挂载在 `Write` 工具上：

```json
{
  "matcher": "Write",
  "hooks": [
    {
      "type": "command",
      "command": "python3 $CODEBUDDY_PROJECT_DIR/.codebuddy/hooks/check-documentation.py",
      "timeout": 30
    }
  ]
}
```

脚本逻辑很简单：检查 Agent 要写的文件是不是 `.md` 或 `.txt`，如果是，对照白名单——白名单之外的一律拦截：

```
Agent 调用 Write 工具
     ↓
check-documentation.py 执行
     ↓
是 .md/.txt 文件？
     ↓ 是
在白名单里？（README.md / CODEBUDDY.md / docs/ / .codebuddy/ 等）
     ↓ 否
permissionDecision: deny → 写入被拒绝
```

白名单明确列出了允许创建的文档类型：

- `README.md`、`CODEBUDDY.md`、`CHANGELOG.md` 等标准文件
- `.codebuddy/` 目录下的所有文件（rules、skills、commands）
- `docs/` 和 `openspec/` 目录

白名单之外的任何 `.md` 文件，Agent 都写不进去。

这个 hook 的本质和 Schema-First、linter 是同一种思维：**不依赖 Agent 的自觉，用环境机制来强制约束行为**。

### 项目实践

以上所有的思路，都在这个项目里得到了完整的落地：

**[modelcraft-go](https://git.woa.com/lukemxjia/modelcraft-go)**

感兴趣的可以直接看代码，从 `CODEBUDDY.md`、`.codebuddy/rules/`、`api/graph/schema/`、`api/openapi/`、`db/schema/`、`db/queries/` 这几个入口开始，Harness 的全貌就在里面。


