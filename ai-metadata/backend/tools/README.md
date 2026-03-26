# 工具手册

> **优先级: 辅助** - 项目使用的脚本、工具及其安装说明。

## 概述

工具手册介绍项目中使用的各种开发工具、自动化脚本，以及它们的安装和使用方法。

## 文档列表

| 文档 | 说明 |
|------|------|
| [tools-installation.md](./tools-installation.md) | 开发工具安装指南 |
| [justfile-guide.md](./justfile-guide.md) | Justfile 命令参考 |
| [scripts.md](./scripts.md) | 脚本工具说明 |
| [ide-setup.md](./ide-setup.md) | IDE 配置指南 |

## 核心工具

| 工具 | 用途 | 安装方式 |
|------|------|----------|
| goenv | Go 版本管理 | git clone |
| just | 命令运行器 | npm / cargo |
| Docker Compose | 容器编排 | curl 下载 |
| Atlas | 数据库迁移 | 脚本安装 |
| golangci-lint | 代码检查 | go install |
| gofumpt | 代码格式化 | go install |
| sqlc | SQL 代码生成 | go install |
| **jq** | **JSON 处理** | **apt/brew/yum** |

## 快速安装

### 必需工具

```bash
# 安装 goenv (Go 版本管理)
git clone https://github.com/go-nv/goenv.git ~/.goenv

# 安装 just (命令运行器)
npm install -g rust-just
# 或使用 cargo:
# cargo install just

# 安装 Docker Compose (容器编排)
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 安装 Atlas (数据库迁移)
curl -sSf https://atlasgo.sh | sh
```

### 开发工具 (通过 just 安装)

```bash
# 安装所有开发工具
just install-tools

# 工具会自动按需安装 (运行命令时检查)
```

## 常用 just 命令

```bash
just --list           # 查看所有可用命令
just <recipe>         # 运行指定命令
just --show <recipe>  # 查看命令详情
just --dry-run <recipe> # 干运行（只显示命令不执行）
```

### 应用管理

```bash
just run              # 运行应用
just build            # 构建应用
just dev              # 热重载开发
```

### 代码质量

```bash
just lint             # 代码检查
just lint-fix         # 自动修复
just test-unit        # 运行测试
just test-coverage    # 覆盖率检查
```

### 数据库

```bash
just db create        # 创建数据库
just db up            # 应用 schema
just db status        # 查看状态
```

### 环境管理

```bash
just env-list         # 列出环境
just env-create dev   # 创建环境
just env-switch dev   # 切换环境
```

### 代码生成

```bash
just generate-gql     # 生成 GraphQL 代码
just generate-sqlc    # 生成 SQL 代码
just generate-oapi    # 生成 OpenAPI 代码
```

## 脚本目录

```
scripts/
├── db-env.sh              # 数据库环境变量加载
├── migrate.sh             # 数据库迁移脚本
├── check-domain-coverage.sh # 覆盖率检查
└── auto-fix-coverage.sh   # 自动补充测试
```

## 查看所有可用命令

```bash
# 使用 just
just --list

# 查看命令详情
just --show <recipe>
```

## 阅读顺序

1. 先阅读 `tools-installation.md` 安装必需工具
2. 再阅读 `justfile-guide.md` 了解命令系统
3. 按需阅读其他文档

## 环境要求

- Go 1.22+
- Docker (可选，用于容器化部署)
- MySQL 8.0+ (或通过 Docker 运行)

## 相关文档

- 部署流程请参考 [4-deployment](../4-deployment/)
- 开发规范请参考 [2-development](../2-development/)
