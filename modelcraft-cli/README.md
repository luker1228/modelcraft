# ModelCraft CLI (`mc`) 使用文档

`mc` 是 ModelCraft 的终端工具，面向 end-user 场景，支持：
- 登录与会话管理
- 项目/数据库/模型目录发现
- 模型能力 introspection
- 运行时 GraphQL 查询执行

## 1. 构建与运行

在仓库根目录执行：

```bash
cd modelcraft-cli
go run . --help
```

或构建二进制：

```bash
cd modelcraft-cli
go build -o ./bin/mc .
./bin/mc --help
```

## 2. 快速开始

```bash
# 1) 登录
mc auth login \
  --server https://gateway.example.com \
  --org acme \
  --username alice \
  --password '***'

# 2) 查看可访问项目
mc catalog projects

# 3) 选择默认项目上下文（后续可省略 --project）
mc auth switch-project sales

# 4) 查看数据库与模型
mc catalog databases
mc catalog models --database crm

# 5) 先看模型可查询字段
mc describe sales.crm.users

# 6) 执行查询
mc run sales.crm.users '{ findMany(take: 5) { id name } }'
```

## 3. 命令总览

- `mc auth`：登录、登出、刷新 token、查看状态、切换默认项目
- `mc catalog`：发现可访问的 `projects/databases/models`
- `mc describe <path>`：基于 GraphQL introspection 输出模型字段
- `mc run <path> [query]`：执行 GraphQL 查询（支持参数或 stdin）
- `mc schema commands`：导出 CLI 命令与 flag schema（JSON）
- `mc version`：显示版本信息

## 4. 路径格式

`<path>` 支持两种格式：
- `project.database.model`
- `database.model`（要求已设置默认 project）

示例：
- `sales.crm.users`
- `crm.users --project sales`

## 5. 常用示例

```bash
# 登录状态
mc auth status

# 按项目列数据库
mc catalog databases --project sales

# 按数据库列模型
mc catalog models --project sales --database crm

# 通过 stdin 传查询
echo '{ count }' | mc run sales.crm.users

# 自定义凭证文件
mc auth status --credentials /tmp/mc-credentials.json
```

## 6. 全局/共享参数

多个命令支持：
- `--credentials`：凭证文件路径（默认由 CLI 内置路径决定）
- `--project`：临时覆盖 project 上下文

查看单命令详情：

```bash
mc <command> --help
```

## 7. 输出与错误

CLI 默认输出 JSON：
- 成功：`{"ok":true,"data":...}`
- 失败：`{"ok":false,"error":...}`

参数错误时通常会给出可重试建议，例如：
- `Run 'mc <command> --help' to inspect valid arguments and flags.`

## 8. 常见问题

- 报 `UNAUTHENTICATED`：先执行 `mc auth login`
- 报 `NO_PROJECT_CONTEXT`：传 `--project <slug>` 或执行 `mc auth switch-project <slug>`
- `mc run` 没有 query：第二个参数传入查询，或从 stdin 管道输入
