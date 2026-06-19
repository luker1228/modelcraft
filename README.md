# ModelCraft

> 低代码平台后端底座之一 —— **数据模型**。
> 把你的数据库变成可治理、可查询、可鉴权的 GraphQL API。

ModelCraft 负责数据模型的设计、托管与运行时查询。管理控制台负责设计时操作（建模、
托管数据库、配置 RLS 策略），而 **end-user 链路**只对外暴露数据消费能力：通过 PAT
（Personal Access Token）发现项目/数据库/模型，并在运行时 GraphQL endpoint 上执行
`findMany` / `count` 等只读查询，RLS（行级安全）自动按调用者身份过滤可见行。

## 它解决什么问题

| 能力 | 说明 | 谁来做 |
|------|------|--------|
| **模型设计** | 创建/导入模型，定义字段、枚举、外键、分组 | 管理控制台（JWT） |
| **数据库托管** | 接入 MySQL 集群，反向工程生成模型，下发 DDL | 管理控制台（JWT） |
| **行级安全 (RLS)** | CEL 表达式声明行级可见范围，运行时自动注入 | 管理控制台（JWT） |
| **资源发现** | 列出可访问的项目、数据库、模型 | End User（PAT） |
| **运行时查询** | 模型级 GraphQL endpoint，`findMany` / `count` 等 | End User（PAT） |
| **多租户隔离** | Org → Project → Database → Model 四级隔离 | — |

## 三种接入方式（end-user 链路）

ModelCraft 对外只暴露 **end-user 链路**：管理控制台完成建模后，数据消费者通过 PAT
查询数据。你有三种方式接入这条链路：

```
                         ┌──────────────────────────┐
                         │   APISIX Gateway (:9080)  │  PAT→JWT 转换 + 身份注入
                         └───────────┬──────────────┘
                                     │ end-user 链路
            ┌────────────────────────┼────────────────────────┐
            │                        │                        │
     ① 官方 CLI (`mc`)        ② 直接调 GraphQL            ③ 自研 SDK
     适合：终端、脚本、CI      适合：后端服务、Agent        适合：封装成语言 SDK
```

### ① 官方 CLI（`mc`）

面向 end-user 的命令行工具，结构化 JSON 输出，适合 AI Agent 与脚本消费。

```bash
# 登录（PAT 永不过期，适合 CI）
mc auth login --server http://localhost:9080 --token 'mc_pat_xxx'

# 发现资源
mc catalog projects                    # 列出可访问项目
mc catalog databases --project sales   # 列出项目内数据库
mc catalog models --database crm       # 列出数据库内模型

# 查看模型结构
mc describe sales.crm.users

# 执行运行时查询
mc run sales.crm.users '{ findMany(take: 5) { id name } }'

# 扮演其他用户（注入 X-MC-Auth-* header，用于 RLS 测试）
mc run sales.crm.users '{ findMany { id } }' \
  --as-userid user_abc --as-username alice --as-role admin
```

CLI 源码与完整文档：[`modelcraft-cli/`](./modelcraft-cli)

### ② 直接调用 GraphQL

不使用 CLI 时，可直接用 `curl` 或任意 GraphQL 客户端调用 end-user endpoint。
PAT 对外只暴露两类操作：**身份验证**（whoami）、**资源发现 + 运行时查询**（GraphQL）。
设计时操作（建模、RLS 配置等）在管理控制台完成，不通过 PAT 暴露。

**身份验证** —— PAT 换取身份（whoami 返回 userId / orgName / 短效 JWT）：

```bash
curl http://localhost:9080/api/tenant/auth/whoami \
  -H "Authorization: Bearer mc_pat_xxx"
```

OpenAPI Spec（仅认证）：[`modelcraft-backend/api/openapi/`](./modelcraft-backend/api/openapi)

**资源发现** —— 列出项目、数据库、模型：

```bash
curl http://localhost:9080/end-user/graphql/org/<org>/project/<project> \
  -H "Authorization: Bearer mc_pat_xxx" \
  -H "Content-Type: application/json" \
  -H "X-Action: query:CatalogModels" \
  -d '{"query":"query CatalogModels($database:String!){ models(input:{databaseName:$database}){ items{ name } } }","operationName":"CatalogModels","variables":{"database":"crm"}}'
```

> 设计时 GraphQL 路由（`/graphql/org/...`）是租户内部 JWT 鉴权，不对外。
> End User 可用的查询在 Schema 中标注 `@hasPermission(allowEndUser: true)`。

**运行时查询** —— 模型级动态 GraphQL：

```bash
curl http://localhost:9080/end-user/graphql/org/<org>/project/<project>/db/<db>/model/<model> \
  -H "Authorization: Bearer mc_pat_xxx" \
  -H "Content-Type: application/json" \
  -d '{"query":"{ findMany(take: 5) { id name } }"}'
```

> 运行时 endpoint 不需要 `X-Action` header；资源发现端点需要。
> RLS 策略在运行时自动生效，按 EndUser 身份过滤行。

### ③ 自研 SDK

API Contract 是机器可读的，你可以基于它生成任意语言的 SDK：

| Contract | 位置 | 用途 |
|---------|------|------|
| OpenAPI 3.0 | `modelcraft-backend/api/openapi/openapi.yaml` | REST（认证 whoami） |
| GraphQL SDL | `modelcraft-backend/api/graph/**/*.graphql` | end-user GraphQL（资源发现 + 运行时查询） |

参考 CLI 的实现（`modelcraft-cli/internal/client/`）—— 它本身就是一个极简的 Go SDK：
`AuthClient.Whoami` 做 PAT 换身份，`GraphQLClient.Execute` 发 GraphQL 请求。
照此模式用任何语言封装即可。

---

## 目录

1. [架构概览](#架构概览)
2. [仓库结构](#仓库结构)
3. [快速部署](#快速部署)
4. [初始化控制台](#初始化控制台)
5. [认证架构](#认证架构)
6. [扮演（Impersonation）](#扮演impersonation)
7. [CLI 命令参考](#cli-命令参考)
8. [运行测试](#运行测试)
9. [发布](#发布)

---

## 架构概览

```
┌─ 管理控制台 (JWT) ──────────────────────────────────────────────┐
│  建模 / 托管数据库 / 配置 RLS / 权限管理        设计时，不对外    │
└────────────────────────────────────────────────────────────────┘

┌─ end-user 链路 (PAT) ───────────────────────────────────────────┐
│  CLI / SDK / 直接 HTTP                                          │
│    │  Bearer mc_pat_xxx                                         │
│    ▼                                                            │
│  APISIX Gateway (:9080)   PAT→JWT 转换，注入 X-User-ID         │
│    │  Bearer <short-lived JWT>                                  │
│    ▼                                                            │
│  Backend (:8080)          end-user GraphQL + 运行时引擎         │
│    │                                                            │
│    ├── MySQL System DB (:6033)   元数据：模型、字段、RLS 策略   │
│    └── MySQL Data Clusters       用户业务数据库（托管的数据源） │
└────────────────────────────────────────────────────────────────┘
```

| 服务 | 端口 | 说明 |
|------|------|------|
| APISIX Gateway | `9080` | 统一入口，PAT→JWT 转换，CORS，请求追踪 |
| Backend | `8080` | Go 后端，REST + GraphQL + 运行时引擎 |
| Frontend | `3100` | Next.js 管理控制台 + End User 入口 |
| MySQL | `6033` | 系统库（元数据） |
| Agent | `8000` | CopilotKit AI Agent（可选） |

---

## 仓库结构

```
modelcraft/                       ← Monorepo
├── modelcraft-backend/           ← Go 后端（唯一真相源：api/）
│   └── api/                        API Contract（OpenAPI + GraphQL SDL）
├── modelcraft-front/             ← Next.js 前端
│   └── contract/                   前端消费的 contract（只读，由 skill 同步）
├── modelcraft-gateway/           ← Gateway 配置
├── modelcraft-cli/               ← 官方 CLI（mc）
├── apisix/                       ← APISIX 声明式路由配置
├── deploy/                       ← Docker Compose 编排 + 环境变量
└── ai-metadata/                  ← 项目知识文档（唯一文档目录）
```

---

## 快速部署

### 前提条件

| 工具 | 最低版本 | 用途 |
|------|---------|------|
| [Go](https://go.dev/doc/install) | 1.24+ | 编译 CLI / 后端 |
| [Docker](https://docs.docker.com/get-docker/) | 24+ | 运行服务 |
| [just](https://github.com/casey/just) | 1.14+ | 任务运行器 |

### 1. 准备环境配置

```bash
cd deploy/env
cp backend.local.env.example  backend.local.env
cp mysql.local.env.example    mysql.local.env
cp apisix.local.env.example   apisix.local.env
cp frontend.local.env.example frontend.local.env
cp agent.local.env.example    agent.local.env
```

按实际环境编辑各 `.env`，重点关注 `mysql.local.env`（`MYSQL_ROOT_PASSWORD`）和
`backend.local.env`（数据库连接串、JWT 密钥）。

### 2. 启动服务

```bash
# 在仓库根目录执行
just deploy/deploy     # 构建镜像并后台启动全部服务
just deploy/db-up      # 首次启动后应用数据库 Schema
just deploy/ps         # 查看容器状态
```

> 部署操作必须在 `./deploy` 目录下执行，默认编排文件为
> `compose/docker-compose.local.yml`。

### 3. 常用运维

```bash
just deploy/up         # 启动（不重新构建）
just deploy/down       # 停止
just deploy/restart    # 重启
just deploy/backend    # 单独重建后端
just deploy/clean      # 清除所有数据卷（慎用）
```

---

## 初始化控制台

服务就绪后，在管理控制台 `http://localhost:3100` 完成首次配置：

1. **注册租户管理员**并登录控制台。
2. **新建项目**（Project，例如 `sales`）。
3. **托管数据库**：在项目内添加数据库集群连接，ModelCraft 自动发现表并生成模型。
4. **创建 End User**：在「用户管理 → End User」为 CLI/API 调用者创建账号。
5. **分配项目访问权限**：将 End User 与项目关联。
6. **创建 PAT**：以 End User 身份登录用户端，在「API Token 管理」创建 `mc_pat_xxx`。

完成后即可接入：

```bash
mc auth login --server http://localhost:9080 --token 'mc_pat_xxx'
```

> RBAC 可选：未配置时，有项目访问权限的 End User 默认拥有该项目下全部只读能力。
> 需要精细化权限时，在「权限管理」中分配角色 / 权限包。

---

## 认证架构

end-user 链路的 PAT 鉴权流程：

```
CLI / SDK (Bearer mc_pat_xxx)
  → APISIX (:9080)
      │  检测 mc_pat_ 前缀 → subrequest backend /api/tenant/auth/whoami
      │  验证通过 → 取回短效 JWT，替换 Authorization header
      │  注入 X-User-ID / X-Org-Name / X-Is-Admin
      ▼
  Backend (:8080)
      │  end-user GraphQL 路由：ChiJWTAuthMiddleware 验证 JWT
      │  运行时 model 路由：RLSContextMiddleware 读取 X-MC-Auth-* header
      ▼
  运行时 GraphQL 引擎 → 按身份执行 RLS 过滤
```

设计时链路（管理控制台）使用独立的 JWT 鉴权，不经过 PAT，也不对外暴露。

| 凭证类型 | 格式 | 过期 | 适用场景 |
|---------|------|------|---------|
| PAT | `mc_pat_` + hex | 永不过期，手动吊销 | end-user 链路：CLI、CI、SDK |
| End User JWT | ES256 签名 | 短效 | 网关→后端内部链路，不对外暴露 |
| 管理员 JWT | ES256 签名 | 短效 | 管理控制台，设计时操作 |

---

## 扮演（Impersonation）

CLI 支持通过全局 flag 注入 `X-MC-Auth-*` header，模拟其他用户的 RLS 上下文，
用于测试行级安全策略：

```bash
mc run sales.crm.users '{ findMany { id name } }' \
  --as-userid user_abc \
  --as-username alice \
  --as-role "admin,manager"
```

| Flag | 注入的 Header | 说明 |
|------|--------------|------|
| `--as-userid` | `X-MC-Auth-Userid-Int` 或 `X-MC-Auth-Userid-Str` | 纯数字 → Int，否则 → Str |
| `--as-username` | `X-MC-Auth-Username` | |
| `--as-role` | `X-MC-Auth-Roles` | 逗号分隔多角色 |

> 扮演 header 仅对运行时 GraphQL 路由（`/end-user/graphql/.../model/...`）生效，
> 由后端 `RLSContextMiddleware` 消费。

---

## CLI 命令参考

### 构建与安装

```bash
cd modelcraft-cli
go build -o ./bin/mc .
sudo cp ./bin/mc /usr/local/bin/mc   # 可选：加入 PATH
```

### auth — 身份认证

```bash
mc auth login --server <url> --token 'mc_pat_xxx'   # 登录
mc auth status                                       # 当前状态
mc auth logout                                       # 清除本地凭证
mc auth switch-project <slug>                        # 设置默认 project
```

### catalog — 资源发现

```bash
mc catalog projects                              # 可访问项目
mc catalog databases [--project <slug>]          # 项目内数据库
mc catalog models --database <name> [--project <slug>]   # 数据库内模型
```

### describe — 模型结构

```bash
mc describe <project>.<database>.<model>
mc describe <database>.<model> --project <slug>
```

### run — 运行时查询

```bash
mc run <path> '{ findMany(take: 5) { id name } }'
echo '{ count { count } }' | mc run <path>     # stdin
```

路径格式：`project.database.model` 或 `database.model`（需预设默认 project）。

### schema — 导出命令 Schema

```bash
mc schema commands    # 输出命令 + flag 的 JSON schema，供 Agent 自省
```

### 全局参数

| 参数 | 环境变量 | 说明 |
|------|---------|------|
| `--credentials` | — | 凭证文件路径，默认 `~/.config/modelcraft/credentials.json` |
| `--project` | `MC_PROJECT` | 覆盖默认 project |
| `--as-userid` | — | 扮演用户 ID |
| `--as-username` | — | 扮演用户名 |
| `--as-role` | — | 扮演角色 |
| — | `MC_SERVER` | 覆盖 Gateway 地址 |
| — | `MC_ORG` | 覆盖 Org slug |
| — | `MC_ACCESS_TOKEN` | 覆盖 access token |

### 输出格式

所有命令输出统一 JSON：

```jsonc
// 成功
{"ok": true, "data": {...}, "meta": {"project": "sales"}}

// 失败
{"ok": false, "error": {"code": "UNAUTHENTICATED", "message": "...", "retryable": true, "suggestion": "..."}}
```

退出码：`0` 成功 · `2` 参数错误 · `3` 未认证 · `4` 权限不足 · `5` 资源不存在 · `7` 未知错误

---

## 运行测试

```bash
# CLI 单元 + 集成测试
just modelcraft-cli/test-all

# 后端测试
just test-unit           # 单元测试
just test-coverage       # 覆盖率
just bdd                 # BDD 验收测试（Cucumber）

# 前端
cd modelcraft-front && npm test
```

---

## 发布

### CLI

```bash
just modelcraft-cli/release v0.2.2    # 打 cli-vX.Y.Z tag，触发 GitHub Actions
```

构建产物（darwin-arm64 + linux-amd64）自动发布为 GitHub Release Assets。

---

## License

[Apache License 2.0](./LICENSE)
