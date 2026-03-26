# Justfile 命令参考

> 项目使用 [just](https://github.com/casey/just) 作为命令运行器，替代 Taskfile/Make。

## 快速开始

```bash
just              # 显示所有可用命令
just <recipe>     # 执行命令
just --list       # 列出所有命令及参数
just --show <cmd> # 查看命令详情
just --dry-run <cmd> # 干运行（只显示命令不执行）
```

## 命令分类

### 应用构建与运行

| 命令 | 说明 | 参数 |
|------|------|------|
| `just build` | 构建应用 | - |
| `just build-prod` | 生产环境构建 | - |
| `just build-all` | 跨平台构建 (Linux, macOS, Windows) | - |
| `just run` | 运行应用 | `force=true` 强制杀掉 8080 端口; `stdout=true` 控制台输出 |
| `just dev` | 热重载开发模式 (需要 Air) | - |
| `just start` | 后台启动服务 | - |
| `just stop` | 停止后台服务 | - |
| `just restart` | 重启服务 | - |
| `just status` | 查看服务状态 | - |
| `just logs` | 查看服务日志 | - |
| `just log-cat` | 按 request_id 查看日志 | `<request_id>` |
| `just clean` | 清理构建产物 | - |

### 依赖与工具安装

| 命令 | 说明 |
|------|------|
| `just deps` | 安装 Go 依赖 (`go mod tidy && go mod download`) |
| `just install-tools` | 安装所有开发工具 (gofumpt, golangci-lint, goimports) |

### 代码生成

| 命令 | 说明 | 注意事项 |
|------|------|----------|
| `just generate-gql` | 生成 GraphQL 代码 | 先编辑 `api/graph/schema/*.graphql` |
| `just generate-oapi` | 合并并生成 OpenAPI 代码 | 先编辑 `api/openapi/*.yaml` 模块文件 |
| `just bundle-oapi` | 仅合并 OpenAPI spec | - |
| `just generate-sqlc` | 生成 sqlc 代码 | - |
| `just clean-gql` | 清理生成的 GraphQL 代码 | **谨慎使用** - 可能丢失自定义代码 |
| `just clean-sqlc` | 清理生成的 sqlc 代码 | - |

### 代码质量

| 命令 | 说明 |
|------|------|
| `just lint` | 运行 golangci-lint 代码检查 |
| `just lint-fix` | 运行 lint 并自动修复 |
| `just check-pkg-dep` | 检查包依赖使用规范 (depguard) |
| `just check-all` | 运行所有检查 (lint + test-unit) |

### 测试 - Go 单元测试

| 命令 | 说明 | 参数 |
|------|------|------|
| `just test` / `just test-unit` | 运行单元测试 | - |
| `just test-unit-coverage` | 运行测试并生成覆盖率报告 | - |
| `just test-unit-fast` | 快速测试 (无竞态检测) | - |
| `just test-unit-pkg` | 测试指定包 | `<pkg>` 包路径 |
| `just test-unit-verbose` | 详细测试输出 | - |
| `just test-unit-bench` | 基准测试 | - |
| `just test-unit-clean` | 清理测试缓存和覆盖率文件 | - |

### 测试 - 覆盖率检查

| 命令 | 说明 | 参数 |
|------|------|------|
| `just test-coverage` | Domain 层覆盖率检查 (要求 95%) | `skip_tests="..."` 跳过的测试 |
| `just test-coverage-all` | 全项目覆盖率检查 | - |
| `just test-coverage-html` | 生成 HTML 覆盖率报告 | - |
| `just test-coverage-badge` | 生成覆盖率徽章 | - |
| `just test-coverage-auto-fix` | 自动补充测试到 95% 覆盖率 | `max_iterations="10"` `package=""` |

### 测试用户管理

| 命令 | 说明 | 参数 |
|------|------|------|
| `just test-user-setup` | 创建测试用户 | - |
| `just test-user-cleanup` | 清理测试用户 | `user_id="..."` |

### Docker 命令

| 命令 | 说明 |
|------|------|
| `just docker-build` | 构建 Docker 镜像 |
| `just docker-run` | 运行 Docker 容器 |
| `just docker-up` | 构建并启动所有服务 |
| `just docker-compose-up` | 启动 Docker Compose 服务 |
| `just docker-compose-down` | 停止 Docker Compose 服务 |
| `just docker-compose-logs` | 查看 Docker Compose 日志 |
| `just docker-compose-build` | 构建 Docker Compose 服务 |
| `just docker-compose-restart` | 重启 Docker Compose 服务 |
| `just docker-clean` | 清理 Docker 环境 |
| `just docker-status` | 查看 Docker 服务状态 |
| `just docker-shell` | 进入应用容器 |
| `just docker-app-logs` | 查看应用日志 |
| `just docker-db-logs` | 查看数据库日志 |

### 部署管理

| 命令 | 说明 | 参数 |
|------|------|------|
| `just deploy-infra` | 管理基础设施 (MySQL, Redis) | `action="start\|status\|stop\|restart"` |
| `just deploy-app` | 管理应用服务 | `action="start\|status\|stop\|restart"` |
| `just deploy-all` | 管理所有服务 | `action="start\|status\|stop\|restart"` |

### 数据库管理

统一使用 `just db <action>` 命令：

| 命令 | 说明 | 参数 |
|------|------|------|
| `just db create` | 创建数据库 | `env_file=".env"` |
| `just db drop` | 删除数据库 | `env_file=".env"` |
| `just db up` | 应用 schema (默认) | `env_file=".env"` |
| `just db down` | 回滚提示 | - |
| `just db status` | 查看数据库状态 | `env_file=".env"` |
| `just db reset` | 重置数据库 (drop + recreate) | `env_file=".env"` |
| `just db lint` | 检查迁移文件 | `env_file=".env"` |
| `just db diff` | 创建迁移 diff | `env_file=".env"` |
| `just db login` | 登录数据库 | `env_file=".env"` |

示例：
```bash
just db                    # 应用 schema (默认)
just db up                 # 应用 schema
just db status             # 查看状态
just db reset              # 重置数据库
just db login              # 登录数据库
just db up .env.autotest   # 使用测试环境
```

### 端口管理

| 命令 | 说明 | 参数 |
|------|------|------|
| `just port-kill` | 杀掉指定端口进程 | `port="8080"` |
| `just port-check` | 检查端口占用 | `port="8080"` |

### 环境管理

| 命令 | 说明 | 参数 |
|------|------|------|
| `just env-list` | 列出所有环境文件 | - |
| `just env-create` | 从模板创建新环境 | `<name>` `template=""` |
| `just env-current` | 显示当前活动环境 | - |
| `just env-switch` | 切换环境 | `<name>` `create=""` |
| `just env-diff` | 对比环境差异 | `<name>` |
| `just env-backup` | 备份当前 .env | - |
| `just env-restore` | 从备份恢复 .env | `<file>` |

示例：
```bash
just env-list                      # 查看所有环境
just env-create dev                # 创建 .env.dev
just env-switch dev                # 切换到 dev 环境
just env-switch prod create=true   # 切换并自动创建
just env-diff prod                 # 对比当前与 prod 差异
```

## 常用工作流

### 开发流程

```bash
# 1. 配置环境
just env-create dev          # 创建开发环境
just env-switch dev          # 切换到开发环境

# 2. 启动服务
just deploy-all              # 启动所有服务 (基础设施 + 应用)
# 或分步：
just deploy-infra            # 启动 MySQL, Redis
just deploy-app              # 启动应用

# 3. 开发
just run                     # 普通运行
just run force=true          # 强制运行 (杀掉占用端口)
just run stdout=true         # 控制台输出日志
just dev                     # 热重载模式

# 4. 查看状态
just deploy-all status       # 查看所有服务状态
just logs                    # 查看应用日志
```

### 测试流程

```bash
just test-unit               # 运行所有单元测试
just test-unit-pkg ./internal/domain/project  # 测试指定包
just test-coverage           # 检查覆盖率 (要求 95%)
just test-coverage-html      # 生成 HTML 报告
```

### 提交前检查

```bash
just lint                    # 代码检查
just lint-fix                # 自动修复
just check-all               # 完整检查 (lint + test)
```

### 代码生成

```bash
# GraphQL (修改 api/graph/schema/*.graphql 后)
just generate-gql

# OpenAPI (修改 api/openapi/*.yaml 模块后)
just generate-oapi

# sqlc (修改 SQL 查询后)
just generate-sqlc
```

### 数据库操作

```bash
just db status               # 查看数据库状态
just db up                   # 应用 schema
just db reset                # 重置数据库
just db login                # 登录数据库 CLI
```

## 关键限制

- **NEVER** 直接编辑 `api/openapi/openapi.yaml` - 应编辑模块文件后运行 `just generate-oapi`
- **谨慎使用** `just clean-gql` - 可能丢失自定义的 resolver 代码

## 参数传递语法

just 使用 `key=value` 语法传递参数：

```bash
just run force=true stdout=true
just test-unit-pkg ./internal/domain/project
just db up .env.autotest
just port-kill 3000
```

## 相关文档

- [tools-installation.md](./tools-installation.md) - 开发工具安装指南
- [scripts.md](./scripts.md) - 脚本工具说明
- [ide-setup.md](./ide-setup.md) - IDE 配置指南
