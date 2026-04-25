# ModelCraft CLI — PRD 总览

> **版本**: v1.0  
> **状态**: Draft  
> **目标用户**: AI Agent（第一版本不考虑人类交互）

---

## 1. 问题陈述

当前 ModelCraft EndUser 仅能通过 Web UI 访问数据。AI Agent 无法以编程方式登录、查询和操作 ModelCraft 中的数据。

需要一个 CLI 工具，让 AI Agent 能够：
- 通过命令行认证身份
- 发现可用的资源（项目、数据库、模型、字段）
- 执行数据查询和变更操作
- 获取结构化的 JSON 输出用于程序化处理

---

## 2. 设计目标

| 目标 | 说明 |
|------|------|
| **Agent-First** | 所有输出默认为机器可读的 JSON，TTY 检测决定是否附加人类友好格式 |
| **零歧义** | 命令结构、标志、输出 schema 均为确定性的，Agent 无需猜测 |
| **可自省** | Agent 可通过 `mc schema` 和 `mc describe` 发现所有能力和资源元数据 |
| **薄客户端** | CLI 不做本地校验（如 limit），服务端是唯一权威，错误信息包含足够的自修正信息 |
| **单二进制** | Go + Cobra 编译为单个可执行文件，零运行时依赖 |

---

## 3. 技术约束

| 项 | 约束 |
|----|------|
| 语言 | Go（与后端共享类型定义） |
| CLI 框架 | Cobra |
| 输出格式 | JSON（默认） / YAML（`--output yaml`） |
| 认证 | EndUser JWT（`iss: "mc-enduser"`） |
| 后端协议 | REST（Auth）+ GraphQL（Runtime 数据查询） |
| 后端要求 | 新增公共 EndUser Auth REST 端点（当前仅有 BFF 内部路由） |

---

## 4. 资源路径约定

### 4.1 三级资源层次

```
project . database . model
  └─ projectSlug  └─ dbName  └─ modelName
```

使用 `.` 作为分隔符（而非 `/`），理由：

| 维度 | `.` 分隔符 | `/` 分隔符 |
|------|-----------|-----------|
| 语义 | 逻辑命名空间（DNS、Protobuf、Java） | 物理路径（文件系统、URL） |
| Shell 安全性 | 无路径补全干扰 | 触发路径自动补全 |
| 先例 | Terraform `aws_instance.web`、K8s 内部 | REST URL `/orgs/acme/projects/sales` |

### 4.2 示例

```bash
# 完整路径
mc query sales.maindb.users --where '{"username":{"contains":"alice"}}'

# 缩写（使用当前项目上下文）
mc query maindb.users --where '...'

# 单模型（使用当前项目+数据库上下文）
mc query users --where '...'
```

上下文解析优先级：
1. 命令行参数中的完整路径 `a.b.c`
2. `--project` / `--database` 显式标志
3. credentials.json 中缓存的当前项目
4. 报错并提示设置上下文

---

## 5. 命令树

```
mc
├── auth                          # 认证管理
│   ├── login                     # 登录
│   ├── logout                    # 登出
│   ├── refresh                   # 刷新 token
│   ├── status                    # 查看当前认证状态
│   └── switch-project            # 切换当前项目
│
├── query <resource-path>         # 数据查询（findMany）
├── get <resource-path>           # 单条查询（findUnique）
├── create <resource-path>        # 创建记录
├── update <resource-path>        # 更新记录
├── delete <resource-path>        # 删除记录
├── count <resource-path>         # 计数
├── aggregate <resource-path>     # 聚合查询
│
├── describe <resource-path>      # 资源元数据（字段/类型/关系/限制）
├── catalog                       # 列出可访问的项目/数据库/模型
│   ├── projects                  # 列出可访问项目
│   ├── databases                 # 列出项目内数据库
│   └── models                    # 列出数据库内模型
│
├── schema                        # CLI 自省（Agent 专用）
│   ├── commands                  # 输出完整命令树 JSON
│   ├── query                     # 输出 query 命令 schema
│   └── flags                     # 输出全局标志 schema
│
└── version                       # CLI 版本信息
```

---

## 6. 子页文档

| 文件 | 说明 |
|------|------|
| [01-auth-flow.md](./01-auth-flow.md) | 认证流程：公共端点设计、登录时序、Token 管理 |
| [02-data-commands.md](./02-data-commands.md) | 数据命令：query/get/create/update/delete/count/aggregate |
| [03-discovery-and-introspection.md](./03-discovery-and-introspection.md) | 资源发现与 Agent 自省：catalog/describe/schema |
| [04-error-handling.md](./04-error-handling.md) | 错误处理：统一格式、退出码、Agent 自修正、Limit 机制 |
| [05-architecture.md](./05-architecture.md) | CLI 内部架构与后端变更需求 |

---

## 7. v1 范围

### 包含

- `mc auth login / logout / refresh / status / switch-project`
- `mc query / get / create / update / delete / count / aggregate`
- `mc describe` — 模型元数据
- `mc catalog projects / databases / models` — 资源发现
- `mc schema commands / query / flags` — Agent 自省
- `mc version`
- 统一 JSON 错误输出 + 语义退出码
- Token 自动刷新
- `.` 分隔的资源路径

### 不包含（v2 考虑）

| 项 | 原因 |
|----|------|
| 人类交互模式（REPL / 交互提示） | v1 Agent-only |
| `--where` 简化语法（如 `username=alice`） | v1 仅 JSON 格式 |
| MCP Server 封装 | 待评估需求后决定 |
| 批量操作（createMany / updateMany / deleteMany） | v2 按需添加 |
| 自助注册 | 由管理员创建账号 |
| OAuth / SSO 登录 | v1 仅 username+password |
| 离线缓存 | 薄客户端不缓存数据 |
| 插件系统 | 过度工程 |

---

## 8. 成功指标

| 指标 | 目标 |
|------|------|
| Agent 可自主完成登录到查询的完整流程 | 无需人类介入 |
| Agent 遇到错误能自修正 | 错误信息包含 `suggestion` 和 `retryable` |
| Agent 可发现所有资源和命令 | 通过 `mc schema` + `mc describe` + `mc catalog` |

---

## 9. 开放问题

| # | 问题 | 状态 |
|---|------|------|
| 1 | PermissionBundle 是否需要在 CLI 层面区分读写权限？ | 待定 — 当前设计依赖服务端鉴权 |
| 2 | 是否需要 MCP Server 封装以便 Agent 框架直接集成？ | 待评估 |
| 3 | Token 存储是否需要加密（如 keyring）？ | v1 明文 JSON，v2 考虑 |
| 4 | 是否支持 `--where` 简化语法作为 JSON 的补充？ | v2 人类模式时考虑 |
| 5 | 是否需要 `mc exec` 支持原始 GraphQL 查询？ | 待评估 |
