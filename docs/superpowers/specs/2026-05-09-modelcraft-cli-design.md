# 设计文档：ModelCraft Agent-First CLI v1

**日期：** 2026-05-09  
**状态：** 已批准，待实现  
**背景：** ModelCraft 现有 EndUser 数据访问主要依赖 Web UI 与 GraphQL 接口。为了让 AI Agent 和自动化脚本稳定完成登录、资源发现、结构理解与只读查询，需要一个面向机器消费的跨平台 CLI。

---

## 问题陈述

当前系统缺少一个稳定、可脚本化、可自省的命令行入口，使 AI Agent 在以下场景中需要自行拼接 Gateway / GraphQL 请求，成本高且容易出错：

- 登录并维护 EndUser 会话
- 列出当前用户可访问的 project / database / model
- 查询模型结构，理解字段与查询能力
- 执行只读数据查询并消费结构化结果
- 在参数错误、上下文缺失、token 过期时执行自修正

此外，CLI 设计必须符合当前系统的真实边界：

- 请求链路必须遵守 `CLI -> Gateway -> Backend`
- EndUser 登录态是 **org 级别 token**，不是“当前 project 级 token”
- 现有 `select-project` 仅做 project 访问校验，不负责重签 access token
- Gateway 现有 EndUser Auth 逻辑偏浏览器 cookie 模式，CLI 需要一套非 cookie 的协议

---

## 目标

- 提供一个 **agent-first** 的 CLI，默认输出稳定 JSON，方便 AI Agent 和脚本消费
- 打通 `login -> catalog -> describe -> query` 的完整只读链路
- 使用英文错误码与错误信息，便于与 GraphQL / Gateway / Backend 错误体系对齐
- 在 Windows、macOS、Linux 上以单二进制方式运行，Windows 作为一等目标
- 允许 CLI 本地维护单一登录档案和可选的默认 project 上下文
- 通过 `mc schema` 提供 CLI 自身的静态自省能力
- 通过 `mc describe` 提供资源级自省能力

---

## 非目标

以下内容不在 v1 范围内：

- `create / update / delete` 写操作
- API Key 认证
- 多 profile 管理
- 当前 database 上下文
- 交互式 REPL / TUI
- 自动抓取全部分页数据
- 本地缓存业务数据
- 中英文错误切换
- 浏览器式 cookie 登录流

---

## 设计决策摘要

本设计确认以下关键决策：

- **实现语言：** Go
- **CLI 框架：** Cobra
- **发布形式：** 各平台单二进制
- **链路边界：** Gateway-first，CLI 不直连 Backend
- **CLI 类型：** agent-first，同时兼顾人工排障
- **范围：** v1 只读
- **登录态存储：** 本地保存 `accessToken + refreshToken`
- **上下文：** 仅维护单一 `currentProject`，且为可选本地默认值
- **档案模型：** 单档案
- **错误语言：** 全英文
- **Windows 目标：** 一等支持对象

---

## 方案比较

### 方案 A：Gateway-Native Agent CLI（选中）

- CLI 统一请求 Gateway
- Gateway 为 CLI 提供非 cookie 的 EndUser Auth passthrough
- CLI 本地保存 token 与 project 列表
- `schema` 本地生成，`describe` 走服务端

**优点：**
- 与现有安全边界一致
- 适合 agent / CI / 自动化脚本
- 不依赖浏览器 cookie 机制
- 后续扩展写操作不会推翻基础设计

**代价：**
- Gateway 需要补充 CLI 专用认证协议
- `describe` 需要组合 runtime introspection 与 catalog 信息

### 方案 B：Backend-Direct CLI（未选）

- CLI 直接调用 Backend

**不选原因：**
- 违背当前系统链路约束
- 会导致 CLI 与前端 / Gateway 的安全边界分叉

### 方案 C：Cookie-Reuse CLI（未选）

- CLI 复用 Gateway 当前浏览器 cookie 登录协议

**不选原因：**
- 原生 CLI 不适合以浏览器 cookie 作为核心会话模型
- 会增加 Windows / PowerShell / agent 执行环境中的状态复杂度

---

## 认证与上下文模型

### 1. 登录态语义

CLI 的登录态是 **org 级** EndUser 会话，包含：

- `server`
- `orgName`
- `userId`
- `accessToken`
- `refreshToken`
- `expiresAt`
- `projects[]`

这里的 `accessToken` **不携带“当前 project”状态**。当前代码中，EndUser token 的签发只写入 `user_id`、`org_name` 和 `aud=end_user`，并未把 `projectSlugs` 写入 JWT claims。`projectSlugs` 目前仅用于应用层在签发前收集可访问项目集合，并未形成 project 级 token。

因此，CLI 不应把“切换当前 project”建模为换 token 行为。

### 2. 自动续期

CLI 在每次请求前检查 `expiresAt`：

- 距过期大于 60 秒：直接使用 `accessToken`
- 距过期小于等于 60 秒：自动调用 `refresh`
- `refresh` 失败：返回认证错误，提示重新登录

### 3. project 上下文

CLI 可选维护一个本地默认值 `currentProject`。

关键规则：

- `currentProject` 不是认证态本身，只是命令默认值
- `login` 成功后 **默认不设置** `currentProject`
- 即使用户没有任何可访问 project，也允许登录成功
- `auth switch-project <slug>` 仅修改本地 `currentProject`
- `switch-project` 不触发 token exchange，不刷新 access token
- `switch-project` 设置前应校验目标 project 是否可访问

### 4. 无 project 的行为

如果用户没有可访问 project，CLI 应表现为：

- `projects: []`
- `currentProject: null`
- `catalog projects` 正常返回空列表
- project 级命令在缺少 `--project` 且 `currentProject` 为空时，返回 `NO_PROJECT_CONTEXT`

### 5. 单档案存储

v1 仅支持一套本地登录档案，建议文件结构如下：

```json
{
  "server": "https://gateway.example.com",
  "orgName": "acme",
  "userId": "01944...",
  "accessToken": "eyJ...",
  "refreshToken": "...",
  "expiresAt": "2026-05-09T12:00:00Z",
  "projects": [
    { "slug": "sales", "title": "Sales" },
    { "slug": "hr", "title": "HR" }
  ],
  "currentProject": null
}
```

环境变量覆盖优先级高于本地文件，v1 支持：

- `MC_SERVER`
- `MC_ORG`
- `MC_ACCESS_TOKEN`
- `MC_PROJECT`（可选）

v1 不做“环境变量覆盖后再回写本地档案”的复杂同步逻辑。

---

## 命令面设计

### 1. 命令树

```text
mc
├── auth
│   ├── login
│   ├── logout
│   ├── refresh
│   ├── status
│   └── switch-project
│
├── catalog
│   ├── projects
│   ├── databases
│   └── models
│
├── describe
├── schema
│   ├── commands
│   ├── query
│   └── flags
│
├── query
├── get
├── count
├── aggregate
│
└── version
```

### 2. v1 包含的命令

- `auth login`
- `auth logout`
- `auth refresh`
- `auth status`
- `auth switch-project`
- `catalog projects`
- `catalog databases`
- `catalog models`
- `describe`
- `schema commands`
- `schema query`
- `schema flags`
- `query`
- `get`
- `count`
- `aggregate`
- `version`

### 3. v1 不包含的命令

- `create`
- `update`
- `delete`

---

## 资源路径规则

### 1. 资源路径格式

v1 使用三段式资源路径：

```text
<project>.<database>.<model>
```

支持两种输入形式：

- `project.database.model`
- `database.model`

v1 **不支持** 单段 `model`。

### 2. 为什么不支持单段 `model`

单段路径必须依赖：

- 当前 project 上下文
- 当前 database 上下文

但本设计已明确：

- v1 只维护 `currentProject`
- v1 不维护 `currentDatabase`

因此若支持单段 `model`，会把状态模型复杂化并引入歧义。v1 保持至少 `database.model` 更稳。

### 3. project 解析优先级

统一优先级如下：

1. 资源路径里的 `project`
2. `--project`
3. 本地 `currentProject`
4. 报错 `NO_PROJECT_CONTEXT`

### 4. database 规则

v1 不维护默认 database。涉及 model 的命令至少需要提供：

- `project.database.model`
- 或 `database.model` + `project` 来源可补全

---

## Catalog / Describe / Schema 设计

### 1. `catalog`

- `catalog projects`
  - 不依赖 `currentProject`
  - 返回当前登录态的可访问项目列表

- `catalog databases`
  - 需要 `project`
  - `project` 可来自 `--project` 或 `currentProject`

- `catalog models`
  - 需要 `project` 和 `database`
  - `project` 可补全，`database` 必须显式提供

v1 可支持 `catalog` 无子命令返回汇总视图，但默认只展开到必要摘要，不递归拉取全量模型树，避免大 org 下响应过重。

### 2. `schema`

`mc schema` 是 **CLI 自省**，不请求后端。

- `schema commands`
- `schema query`
- `schema flags`

均由 CLI 根据 Cobra 命令树和命令元数据静态生成。

设计收益：

- 不依赖登录态
- 响应稳定且快速
- 适合 agent 启动时理解 CLI 能力

### 3. `describe`

`describe` 是 **资源自省**，必须请求服务端。

#### `describe <project>.<database>`

通过 catalog / project 级 GraphQL 获取：

- database 名称
- model 列表
- 基础摘要信息

#### `describe <project>.<database>.<model>`

以 **runtime GraphQL introspection 为主，catalog 信息为辅**。

runtime introspection 负责提供：

- field 名称
- field 类型
- required / list 信息
- query / filter / orderBy 等输入能力

catalog 负责补充：

- model title
- database 名称
- 其他现成的业务元数据

### 4. `describe` 的边界

v1 的 `describe` 目标是“对 agent 足够有用”，不是“一次性返回所有理想元数据”。

因此：

- 若当前后端没有现成来源，v1 不强行承诺返回 `limits.maxTake`
- 若某些展示型元数据无法稳定拿到，允许 v1 暂不提供
- 重点保证 agent 能理解字段、类型、过滤能力和资源位置

---

## 输出格式与错误模型

### 1. 输出通道规则

- `stdout`：仅输出业务结果（JSON / YAML）
- `stderr`：仅输出诊断信息
- 默认 `--output json`
- 支持 `--output yaml`
- 支持 `--compact` 输出单行 JSON

v1 不做 TTY 下自动切换人类格式，避免 stdout schema 随终端环境变化。

### 2. 成功输出结构

所有成功响应统一为：

```json
{
  "ok": true,
  "data": {},
  "meta": {}
}
```

说明：

- `data`：主业务结果
- `meta`：上下文、分页、补充信息

### 3. 错误输出结构

```json
{
  "ok": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message.",
    "retryable": true,
    "suggestion": "Actionable next step.",
    "details": {}
  }
}
```

规则：

- `code`：英文 `UPPER_SNAKE_CASE`
- `message`：英文
- `suggestion`：英文
- `details`：英文键名的结构化字段

### 4. 退出码

- `0`：成功
- `2`：参数 / 输入错误
- `3`：认证错误
- `4`：权限错误
- `5`：资源不存在 / 上下文缺失
- `6`：服务端限制 / 可调整约束
- `7`：服务端或网络异常

`NO_PROJECT_CONTEXT` 归类为 `5`。

### 5. 关键错误样例

- `NO_PROJECT_CONTEXT`
- `MODEL_NOT_FOUND`
- `DATABASE_NOT_FOUND`
- `INVALID_JSON_FLAG`
- `TOKEN_EXPIRED`
- `UNAUTHENTICATED`
- `PERMISSION_DENIED`
- `TAKE_EXCEEDS_LIMIT`
- `SERVICE_UNAVAILABLE`

CLI 应优先复用 GraphQL / Gateway / Backend 现有英文错误码；仅在 CLI 本地错误场景下补充自身错误码。

### 6. `--verbose`

`--verbose` 打开后，stderr 可输出：

- 目标 URL
- 资源解析结果
- token 是否被自动 refresh
- request ID
- GraphQL operation 名称
- 请求耗时

即使启用 `--verbose`，stdout schema 仍保持不变。

---

## Gateway 协议设计

### 1. 背景

当前 Gateway 的 EndUser Auth 更偏浏览器 cookie 流程：

- login / refresh 时会提取或写入 refresh cookie
- CLI 不适合把 cookie 作为主会话模型

### 2. v1 目标

保留 `CLI -> Gateway -> Backend` 的边界，同时为 CLI 提供 **非 cookie 的 JSON 透传协议**。

### 3. 推荐做法

在 Gateway 中新增一组 **CLI 专用 EndUser Auth passthrough handler**，与现有浏览器 handler 并列存在，而不是复用同一路由再分支逻辑。

CLI 专用 handler 的职责：

- 将 Backend 的 JSON token 响应原样透传给 CLI
- 不写入 refresh cookie
- 不依赖浏览器状态
- 保持统一认证入口仍在 Gateway

v1 需要覆盖：

- `login`
- `refresh`
- `logout`
- `me`
- `select-project`（仅用于 project 可访问性校验）

---

## CLI 内部架构

推荐目录：

```text
modelcraft-cli/
├── cmd/
├── internal/auth/
├── internal/client/
├── internal/config/
├── internal/output/
├── internal/resource/
├── internal/schema/
└── main.go
```

模块职责：

- `cmd/`
  - Cobra 命令定义
  - 参数接线

- `internal/auth/`
  - 登录态加载
  - token 过期检查
  - 自动 refresh
  - project 上下文切换

- `internal/client/`
  - Gateway REST / GraphQL 请求封装

- `internal/config/`
  - 单档案 credentials 持久化

- `internal/output/`
  - 成功输出
  - 错误输出
  - 退出码映射

- `internal/resource/`
  - `project.database.model` / `database.model` 解析

- `internal/schema/`
  - CLI 本地 schema 自省输出
  - `describe` 所需的 introspection 转换辅助

设计原则：

- 命令层不直接拼 HTTP / GraphQL
- 输出层不关心业务语义
- 资源解析逻辑集中管理
- 登录态与 project 上下文逻辑集中管理

---

## 技术选型：为什么是 Go

### 1. 选中方案

- 语言：Go
- CLI 框架：Cobra

### 2. 选择原因

- 与现有团队工具链一致
- 与 Backend 技术栈一致，认知切换最小
- Windows 运行期问题更少
- 原生跨平台单二进制分发友好
- 对本 CLI 的核心问题（auth / HTTP / GraphQL / JSON / config）足够直接
- 从“先做出来 + 后续少返工”的角度，总体技术成本最低

### 3. Windows 目标

Windows 是一等目标，因此应优先保障：

- `mc.exe` 单文件运行
- PowerShell 下 JSON 参数传递可预测
- 用户目录下配置文件写入正常
- stdout / stderr 分离稳定

---

## 测试策略

### 1. 单元测试

覆盖：

- 资源路径解析
- project 优先级决策
- credentials 读写
- token 过期判断
- 错误码映射
- 输出 schema

### 2. client 协议测试

通过 mock 测试：

- auth login / refresh / logout
- catalog databases / models
- query / get / count / aggregate
- GraphQL error -> CLI error 映射

### 3. 最小端到端 smoke

v1 仅保留关键链路：

- login
- catalog projects
- switch-project
- query
- token refresh

不以大规模 E2E 作为 v1 完整性的前提。

### 4. Windows 验证重点

必须显式验证：

- PowerShell 下 `--where` / `--select` / `--orderBy` JSON quoting
- `mc.exe` 在 Windows 下的配置文件读写
- 自动 refresh 与错误输出在 PowerShell 中表现一致

---

## 发布策略

### 1. v1 发布顺序

先发布平台二进制，再补包管理器：

- `windows-amd64`
- `windows-arm64`
- `darwin-amd64`
- `darwin-arm64`
- `linux-amd64`

### 2. `version`

`mc version` 至少输出：

- version
- commit
- build time

### 3. 后续再补的分发入口

- Windows：`winget` / `scoop`
- macOS：`brew`
- 开发者入口：`go install ...@latest`

v1 先保证“下载即可运行”。

---

## 文档要求

v1 至少提供以下文档：

- Quick Start
  - login
  - status
  - switch-project
  - first query

- Command Reference
  - flags
  - examples
  - output schema

- Windows Notes
  - PowerShell JSON quoting
  - config file location
  - 常见认证 / TLS / PATH 问题

---

## 验收标准

v1 视为完成，至少应满足：

- AI Agent 能独立完成：
  - 登录
  - 列出可访问 project
  - 设置或显式指定 project
  - 列出 database 和 model
  - 执行只读查询
  - 在常见错误后根据 `suggestion` 和 `details` 完成一次自修正

- Windows 用户能直接运行下载的 `mc.exe`
- stdout 输出结构在各命令间保持一致
- 自动 refresh 生效
- 无 project 时行为明确，不进行隐式猜测

---

## 影响范围

| 层 | 变更内容 |
|----|---------|
| `modelcraft-cli` | 新建 CLI 工程，包含认证、配置、输出、资源解析、REST/GraphQL client |
| `modelcraft-gateway` | 新增 CLI 专用 EndUser Auth passthrough handler |
| `modelcraft-backend` | 无需改变核心 token 语义；如 `describe` 需要补充元数据，再评估最小增量支持 |
| 文档 | 增补 CLI Quick Start、命令参考、Windows 使用说明 |
| 发布 | 增加多平台 CLI 二进制构建与版本注入 |

---

## 参考约束

- `ai-metadata/prd/cli/00-cli-overview.md`
- `ai-metadata/prd/cli/01-auth-flow.md`
- `ai-metadata/prd/cli/02-data-commands.md`
- `ai-metadata/prd/cli/03-discovery-and-introspection.md`
- `ai-metadata/prd/cli/04-error-handling.md`
- `ai-metadata/prd/cli/05-architecture.md`
- `ai-metadata/backend/development/developer-enduser-system.md`
- `modelcraft-backend/internal/app/enduser/end_user_auth_service.go`
- `modelcraft-backend/internal/interfaces/http/routes.go`
- `modelcraft-gateway/cmd/gateway/main.go`
