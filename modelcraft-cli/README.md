# ModelCraft CLI (`mc`)

`mc` 是 ModelCraft 的终端工具，面向 end-user 场景，支持：

- PAT Token / 用户名密码两种登录方式
- 项目 / 数据库 / 模型目录发现
- 模型字段 introspection
- 运行时 GraphQL 查询执行

---

## 构建

```bash
cd modelcraft-cli
go build -o ./bin/mc .
./bin/mc --help
```

---

## 快速开始

### 方式一：PAT Token 登录（推荐）

在控制台「Token 管理」页面创建 PAT，然后：

```bash
mc auth login --token 'mc_pat_xxx'
```

PAT 会话无过期时间，无需 refresh。

### 方式二：用户名/密码登录

```bash
mc auth login \
  --server https://gateway.example.com \
  --username alice \
  --password '***'
```

`--server` 默认为 `https://lukemxjia.devcloud.woa.com`。

---

## 命令一览

| 命令 | 说明 |
|------|------|
| `mc version` | 显示版本信息 |
| `mc auth login` | 登录（PAT 或用户名/密码） |
| `mc auth status` | 查看当前登录状态 |
| `mc auth logout` | 登出并清除本地会话 |
| `mc auth refresh` | 用 refresh token 刷新 access token（仅密码登录） |
| `mc auth switch-project <slug>` | 设置默认 project 上下文 |
| `mc catalog projects` | 列出可访问的项目 |
| `mc catalog databases` | 列出项目内数据库 |
| `mc catalog models` | 列出数据库内模型 |
| `mc describe <path>` | 查看模型字段（GraphQL introspection） |
| `mc query <path>` | 查询模型数据（结构化，自动 findMany） |
| `mc run <path> [query]` | 执行任意 GraphQL 查询 |
| `mc schema commands` | 导出 CLI 命令与 flag schema（JSON） |

---

## 路径格式

运行时命令（`describe`、`query`、`run`）使用统一路径格式：

```
project.database.model      # 完整路径
database.model              # 省略 project（需预先设置默认 project）
```

示例：
```
sales.crm.users
crm.users --project sales
```

---

## 典型工作流

```bash
# 1. 登录
mc auth login --token 'mc_pat_xxx'

# 2. 查看可访问项目
mc catalog projects

# 3. 设置默认 project（后续省略 --project）
mc auth switch-project sales

# 4. 发现数据库与模型
mc catalog databases
mc catalog models --database crm

# 5. 查看模型字段
mc describe sales.crm.users

# 6. 查询数据
mc query sales.crm.users

# 7. 自定义 GraphQL 查询
mc run sales.crm.users '{ findMany { items { __typename } totalCount } }'

# 通过 stdin 传入查询
echo '{ count { count } }' | mc run sales.crm.users
```

---

## 认证架构

```
CLI (mc_pat_xxx)
  → APISIX Gateway (9080)           # 验证 PAT，注入 X-User-ID
      → Backend (8080)              # 信任 X-User-ID，执行业务逻辑
```

- **PAT 登录**：CLI 调用 `/api/cli/end-user/auth/whoami` 获取身份，本地保存 PAT 作为 access token，无 refresh token
- **密码登录**：获取 JWT access token + refresh token，access token 过期前自动 refresh
- **APISIX 职责**：验证所有 token（JWT 或 PAT），统一转换为 `X-User-ID` 头注入后端

---

## 输出格式

所有命令输出 JSON 包裹格式：

```json
// 成功
{"ok": true, "data": {...}, "meta": {...}}

// 失败
{"ok": false, "error": {"code": "...", "message": "...", "retryable": true, "suggestion": "..."}}
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

## 全局参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--credentials` | 凭证文件路径 | `~/.config/modelcraft/credentials.json` |
| `--project` | 临时覆盖 project 上下文 | 从凭证文件读取 |

环境变量也可覆盖：`MC_SERVER`、`MC_ORG`、`MC_PROJECT`、`MC_ACCESS_TOKEN`

---

## 常见错误

| 错误码 | 原因 | 解决方案 |
|--------|------|----------|
| `UNAUTHENTICATED` | 未登录或 token 过期 | `mc auth login` |
| `NO_PROJECT_CONTEXT` | 未设置默认 project | `--project <slug>` 或 `mc auth switch-project` |
| `MISSING_REQUIRED_FLAG` | 缺少必填参数 | 查看 `mc <command> --help` |
| `PROJECT_NOT_FOUND` | project slug 不在授权列表 | `mc catalog projects` 查看可访问列表 |
| `SERVICE_UNAVAILABLE` | 网关不可达 | 检查网络或 server 地址 |

---

## 运行测试

```bash
cd modelcraft-cli
go test ./...           # 全部单元+集成测试（66 个）
go test ./... -v -run TestAuth   # 过滤特定测试
```
