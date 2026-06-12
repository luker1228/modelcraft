# ModelCraft CLI (`mc`) 设计指南

> **定位**：面向 AI Agent 的数据访问工具。Agent 通过 `describe` 获取模型 schema，通过 `run` 执行查询，所有输出均为结构化 JSON。

---

## 核心设计

```
API Token (PAT)
    ↓
mc auth login --token mc_pat_xxx
    ↓
~/.config/modelcraft/credentials.json
    ↓
mc describe <path>   →  获取 GraphQL schema（AI 理解数据结构）
mc run <path> <query>  →  执行 GraphQL 查询（AI 读写数据）
```

CLI 不管理 token 生命周期——PAT 永不过期，由管理员在控制台手动吊销。

---

## 认证：API Token 登录

### 登录

```bash
mc auth login --token mc_pat_xxx
mc auth login --token mc_pat_xxx --server http://gateway:9080
```

`--server` 默认为 `http://lukemxjia.devcloud.woa.com:9080`，本地部署时需显式指定。

### 凭证文件

登录成功后，凭证写入本地 JSON 文件：

**默认路径：`~/.config/modelcraft/credentials.json`**

```json
{
  "server": "http://gateway:9080",
  "orgName": "acme",
  "userId": "u1",
  "accessToken": "mc_pat_xxx",
  "currentProject": "sales"
}
```

| 字段 | 说明 |
|------|------|
| `server` | Gateway 地址，后续命令从此读取，无需重复传 |
| `orgName` | Org 标识，用于构建所有请求 URL |
| `userId` | End User ID |
| `accessToken` | PAT 原文，直接用于 Bearer 认证 |
| `currentProject` | 默认 project 上下文，可被 `--project` 临时覆盖 |

可用 `--credentials <path>` 指定自定义路径，适合多账号或 CI 场景：

```bash
mc auth login --token mc_pat_xxx --credentials /tmp/mc-agent.json
mc describe sales.crm.users --credentials /tmp/mc-agent.json
```

### 其他 auth 命令

```bash
mc auth status                   # 查看当前凭证内容
mc auth switch-project <slug>    # 更新 currentProject（实时调后台验证）
mc auth logout                   # 删除凭证文件
```

---

## 核心接口一：`describe` — 获取 GraphQL Schema

Agent 在执行查询前，通过 `describe` 了解模型的字段结构和可用操作。

```bash
mc describe <projectSlug>.<databaseName>.<modelName>
mc describe <databaseName>.<modelName>          # 省略 projectSlug，使用 currentProject
mc describe crm.users --project sales           # --project 临时覆盖
```

**输出示例：**

```json
{
  "ok": true,
  "data": {
    "types": [
      {
        "name": "users",
        "kind": "OBJECT",
        "fields": [
          { "name": "id",    "type": { "kind": "NON_NULL", "ofType": { "kind": "SCALAR", "name": "ID" } } },
          { "name": "name",  "type": { "kind": "SCALAR", "name": "String" } },
          { "name": "email", "type": { "kind": "SCALAR", "name": "String" } }
        ]
      }
    ]
  },
  "meta": { "project": "sales", "database": "crm", "model": "users" }
}
```

内部调用 GraphQL `__schema` introspection，返回完整类型信息。

---

## 核心接口二：`run` — 执行 GraphQL 查询

Agent 拿到 schema 后，用 `run` 执行任意 GraphQL 查询。

```bash
# 直接传查询体
mc run sales.crm.users '{ findMany(take: 5) { id name email } }'

# stdin 传入（适合 Agent 生成复杂查询）
echo '{ count { count } }' | mc run sales.crm.users

# 省略 projectSlug（使用 currentProject）
mc run crm.users '{ findMany { id name } }'
```

**输出示例：**

```json
{
  "ok": true,
  "data": {
    "findMany": [
      { "id": "1", "name": "Alice", "email": "alice@example.com" },
      { "id": "2", "name": "Bob",   "email": "bob@example.com" }
    ]
  },
  "meta": { "project": "sales", "database": "crm", "model": "users" }
}
```

---

## 路径格式

两个核心命令统一使用**资源路径**寻址模型：

```
<projectSlug>.<databaseName>.<modelName>    # 完整路径
<databaseName>.<modelName>                  # 省略 projectSlug（需已设置 currentProject）
```

示例：

```bash
mc describe sales.crm.users        # projectSlug=sales, databaseName=crm, modelName=users
mc describe crm.users              # 省略 sales，从 credentials.json 的 currentProject 读取
mc describe crm.users --project sales  # --project 临时覆盖，不修改凭证文件
```

---

## 资源发现（辅助）

Agent 首次使用时，可通过以下命令探索可访问资源：

```bash
mc catalog projects                           # 列出可访问的项目
mc catalog databases --project sales          # 列出项目内数据库
mc catalog models --project sales --database crm  # 列出数据库内模型
```

典型的 Agent 启动流程：

```bash
mc catalog projects                  # 1. 发现有哪些项目
mc auth switch-project sales         # 2. 设置 currentProject
mc catalog databases                 # 3. 发现有哪些数据库
mc catalog models --database crm     # 4. 发现有哪些模型
mc describe crm.users                # 5. 了解 crm.users 结构（省略 projectSlug）
mc run crm.users '{ findMany { id } }'  # 6. 执行查询
```

---

## CLI 自省

Agent 可通过 `schema commands` 获取 CLI 完整命令树和 flag 列表，无需硬编码：

```bash
mc schema commands
```

---

## 输出格式

所有命令输出统一 JSON 包装：

```json
// 成功
{ "ok": true, "data": { ... }, "meta": { ... } }

// 失败
{ "ok": false, "error": { "code": "...", "message": "...", "retryable": true, "suggestion": "..." } }
```

### 退出码

| 码 | 错误码 | 含义 |
|----|--------|------|
| `0` | — | 成功 |
| `2` | `INVALID_ARGUMENT` / `MISSING_REQUIRED_FLAG` | 参数错误 |
| `3` | `UNAUTHENTICATED` | 未登录或 PAT 已被吊销 |
| `4` | `PERMISSION_DENIED` | 权限不足 |
| `5` | `NOT_FOUND` / `NO_PROJECT_CONTEXT` | 资源不存在或未设置 project |
| `7` | 其他 | 未知错误 |

---

## 环境变量

可覆盖凭证文件中的字段，适合 CI 或多环境切换：

| 变量 | 覆盖字段 | 说明 |
|------|---------|------|
| `MC_ACCESS_TOKEN` | `accessToken` | 直接传 PAT，无需凭证文件 |
| `MC_SERVER` | `server` | Gateway 地址 |
| `MC_ORG` | `orgName` | Org slug |
| `MC_PROJECT` | `currentProject` | project 上下文 |

CI 场景最简配置（无需凭证文件）：

```bash
MC_ACCESS_TOKEN=mc_pat_xxx MC_PROJECT=sales \
  mc run crm.users '{ findMany(take: 10) { id name } }'
```
