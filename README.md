# ModelCraft CLI (`mc`)

`mc` 是 ModelCraft 的终端工具，面向 end-user 场景，支持：

- PAT Token / 用户名密码两种登录方式
- 项目 / 数据库 / 模型目录发现（catalog）
- 模型字段 introspection（describe）
- 运行时 GraphQL 查询执行（run）

所有输出均为结构化 JSON，方便 AI Agent 或脚本消费。

---

## 目录

1. [前提条件](#前提条件)
2. [部署后端服务](#部署后端服务)
3. [构建 CLI](#构建-cli)
4. [快速开始](#快速开始)
5. [命令参考](#命令参考)
6. [路径格式](#路径格式)
7. [典型工作流](#典型工作流)
8. [认证架构](#认证架构)
9. [输出格式与退出码](#输出格式与退出码)
10. [全局参数与环境变量](#全局参数与环境变量)
11. [常见错误](#常见错误)
12. [运行测试](#运行测试)

---

## 前提条件

| 工具 | 最低版本 | 用途 |
|------|---------|------|
| [Go](https://go.dev/doc/install) | 1.21+ | 编译 CLI |
| [Docker](https://docs.docker.com/get-docker/) | 24+ | 运行后端服务 |
| [just](https://github.com/casey/just) | 1.14+ | 任务运行器（可选，方便起停服务） |
| [atlas](https://atlasgo.io/getting-started) | latest | 数据库 Schema 迁移（`db-up` 时需要） |
| [mysql client](https://dev.mysql.com/downloads/) | 8.0+ | `db-up` / `db-reset` 时需要 |

---

## 部署后端服务

### 1. 准备环境配置文件

后端所有服务的配置均位于 `deploy/env/`。每个服务都有一个 `.example` 示例文件：

```bash
cd deploy/env
cp backend.local.env.example  backend.local.env
cp mysql.local.env.example    mysql.local.env
cp apisix.local.env.example   apisix.local.env
cp frontend.local.env.example frontend.local.env
cp agent.local.env.example    agent.local.env
```

根据实际环境编辑各 `.env` 文件，重点关注：

- `mysql.local.env`：设置 `MYSQL_ROOT_PASSWORD`
- `backend.local.env`：设置数据库连接串和密钥
- `apisix.local.env`：设置 JWT 签名密钥（需与 backend 一致）

### 2. 启动所有服务

```bash
# 进入部署目录（必须在此目录下执行）
cd deploy

# 构建镜像并后台启动所有服务
just deploy
# 等价于: docker compose -f compose/docker-compose.local.yml up -d --build
```

### 3. 初始化数据库 Schema

服务首次启动后，需要应用 Schema：

```bash
cd deploy
just db-up
```

### 4. 验证服务状态

```bash
cd deploy
just ps          # 查看各容器状态
just logs        # 跟踪所有服务日志
```

| 服务 | 地址 | 说明 |
|------|------|------|
| Backend | `http://localhost:8080` | Go 后端 API |
| Gateway (APISIX) | `http://localhost:9080` | CLI 和前端统一入口 |
| Frontend | `http://localhost:3100` | 管理控制台（创建 End User / PAT） |
| MySQL | `localhost:6033` | 数据库（本地调试用） |
| phpMyAdmin | `http://localhost:8081` | 数据库管理 UI（需 `just tools` 启动） |

### 5. 常用运维命令

```bash
cd deploy

just up          # 启动（不重新构建）
just down        # 停止并移除容器
just restart     # 重启所有服务
just clean       # 停止并清除所有数据卷（慎用，会清空 MySQL）
just backend     # 单独重新构建并重启后端
just frontend    # 单独重新构建并重启前端
just db-reset    # 重建数据库（慎用）
```

---

## 构建 CLI

```bash
# 进入 CLI 目录
cd modelcraft-cli

# 编译（输出到 bin/mc）
go build -o ./bin/mc .

# 验证
./bin/mc --help
```

可选：将 `bin/mc` 加入 `PATH`，方便全局使用：

```bash
sudo cp ./bin/mc /usr/local/bin/mc
# 或
export PATH="$PWD/bin:$PATH"
```

> **Windows / macOS 交叉编译**
>
> ```bash
> GOOS=darwin GOARCH=arm64 go build -o ./bin/mc-darwin-arm64 .
> GOOS=windows GOARCH=amd64 go build -o ./bin/mc.exe .
> ```

---

## 快速开始

### 方式一：PAT Token 登录（推荐）

在管理控制台（`http://localhost:3100`）中为 End User 创建 PAT，然后：

```bash
mc auth login --server http://localhost:9080 --token 'mc_pat_xxx'
```

PAT 会话无过期时间，无需 refresh，适合 CI / Agent 使用。

### 方式二：用户名/密码登录

```bash
mc auth login \
  --server http://localhost:9080 \
  --username alice \
  --password 'your-password'
```

> `--server` 默认指向 `https://lukemxjia.devcloud.woa.com`，本地部署时需显式指定。

登录成功后凭证保存至 `~/.config/modelcraft/credentials.json`，后续命令无需重复指定 `--server`。

---

## 命令参考

### auth — 身份认证

```bash
mc auth login           # 登录（PAT 或用户名/密码）
mc auth status          # 查看当前登录状态
mc auth logout          # 登出并清除本地凭证
mc auth refresh         # 刷新 access token（仅密码登录有效）
mc auth switch-project  # 设置默认 project 上下文
```

#### `mc auth login`

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--server` | Gateway 地址 | `https://lukemxjia.devcloud.woa.com` |
| `--token` | PAT（`mc_pat_xxx`），指定后跳过用户名/密码 | — |
| `--username` | End User 用户名 | — |
| `--password` | End User 密码 | — |
| `--org` | Org slug（可选，通常由用户名自动解析） | — |
| `--credentials` | 凭证文件路径 | `~/.config/modelcraft/credentials.json` |

#### `mc auth switch-project <slug>`

```bash
mc auth switch-project sales
```

设置后续命令的默认 project，无需每次传 `--project`。

---

### catalog — 资源发现

```bash
mc catalog projects                         # 列出可访问项目
mc catalog databases [--project <slug>]     # 列出项目内数据库
mc catalog models --database <name> [--project <slug>]  # 列出数据库内模型
```

---

### describe — 字段 introspection

```bash
mc describe sales.crm.users
mc describe crm.users --project sales
```

通过 GraphQL introspection 返回模型所有字段、类型及可用 Query/Mutation 操作。

---

### run — 执行 GraphQL 查询

```bash
# 直接传入查询体
mc run sales.crm.users '{ findMany(take: 5) { id name } }'

# 通过 stdin 传入（便于复杂查询或管道组合）
echo '{ count { count } }' | mc run sales.crm.users

# 从文件读取
cat query.graphql | mc run sales.crm.users
```

---

### schema — 导出命令 schema

```bash
mc schema commands    # 输出 CLI 命令与 flag 的完整 JSON schema
```

AI Agent 可通过此命令自省可用操作，无需硬编码命令列表。

---

### version — 版本信息

```bash
mc version
```

---

## 路径格式

运行时命令（`describe`、`run`）使用统一路径格式：

```
project.database.model    # 完整路径
database.model            # 省略 project（需预先设置默认 project）
```

示例：

```bash
mc describe sales.crm.users           # project=sales, db=crm, model=users
mc run crm.orders --project sales     # 通过 --project 覆盖
```

---

## 典型工作流

```bash
# ─── 1. 登录 ───────────────────────────────────────────────
mc auth login --server http://localhost:9080 --token 'mc_pat_xxx'

# ─── 2. 确认身份 ────────────────────────────────────────────
mc auth status

# ─── 3. 查看可访问项目 ─────────────────────────────────────
mc catalog projects

# ─── 4. 设置默认 project（后续省略 --project）─────────────
mc auth switch-project sales

# ─── 5. 发现数据库与模型 ────────────────────────────────────
mc catalog databases
mc catalog models --database crm

# ─── 6. 查看模型字段 ────────────────────────────────────────
mc describe sales.crm.users

# ─── 7. 执行查询 ────────────────────────────────────────────
mc run sales.crm.users '{ findMany(take: 5) { id name email } }'

# ─── 8. 组合管道 ────────────────────────────────────────────
mc run sales.crm.users '{ count { count } }' | jq '.data.count.count'
```

---

## 认证架构

```
CLI (mc_pat_xxx / JWT)
  → APISIX Gateway (:9080)        # 验证 Token，注入 X-User-ID
      → Backend (:8080)           # 信任 X-User-ID，执行业务逻辑
          → MySQL (:6033)
```

| 登录方式 | 本地存储 | 过期策略 |
|---------|---------|---------|
| PAT (`--token`) | PAT 原文作为 access_token，无 refresh_token | 永不过期，由管理员手动吊销 |
| 用户名/密码 | JWT access_token + refresh_token | access_token 过期时自动 `refresh` |

---

## 输出格式与退出码

所有命令输出统一 JSON 包装：

```json
// 成功
{"ok": true, "data": {...}, "meta": {"project": "sales"}}

// 失败
{"ok": false, "error": {"code": "UNAUTHENTICATED", "message": "No local session found.", "retryable": true, "suggestion": "Run 'mc auth login'."}}
```

### 退出码

| 码 | 含义 |
|----|------|
| `0` | 成功 |
| `2` | 参数错误（`INVALID_ARGUMENT` / `MISSING_REQUIRED_FLAG`） |
| `3` | 未认证（`UNAUTHENTICATED`） |
| `4` | 权限不足（`PERMISSION_DENIED`） |
| `5` | 资源不存在（`NOT_FOUND` / `NO_PROJECT_CONTEXT`） |
| `7` | 未知错误 |

---

## 全局参数与环境变量

| 参数 | 环境变量 | 说明 | 默认值 |
|------|---------|------|--------|
| `--credentials` | — | 凭证文件路径 | `~/.config/modelcraft/credentials.json` |
| `--project` | `MC_PROJECT` | 临时覆盖 project 上下文 | 从凭证文件读取 |
| — | `MC_SERVER` | 覆盖 Gateway 地址 | 从凭证文件读取 |
| — | `MC_ORG` | 覆盖 Org slug | 从凭证文件读取 |
| — | `MC_ACCESS_TOKEN` | 覆盖 access token（CI 场景） | 从凭证文件读取 |

---

## 常见错误

| 错误码 | 原因 | 解决方案 |
|--------|------|----------|
| `UNAUTHENTICATED` | 未登录或 token 过期 | `mc auth login` |
| `NO_PROJECT_CONTEXT` | 未设置默认 project | `--project <slug>` 或 `mc auth switch-project <slug>` |
| `MISSING_REQUIRED_FLAG` | 缺少必填参数 | `mc <command> --help` |
| `PROJECT_NOT_FOUND` | project slug 不在授权列表 | `mc catalog projects` 查看可访问列表 |
| `SERVICE_UNAVAILABLE` | 网关不可达 | 检查 `--server` 地址和服务状态（`just ps`） |

---

## 运行测试

```bash
cd modelcraft-cli

# 单元测试
just test
# 等价于: go test ./cmd/... ./internal/...

# 集成测试（需要先编译二进制）
just test-integration
# 等价于: go test ./tests/integration/... -v -timeout 60s

# 全部测试
just test-all

# 过滤特定测试
go test ./... -v -run TestAuth
```
