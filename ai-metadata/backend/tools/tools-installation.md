# 开发工具安装手册

本文档介绍 ModelCraft Go 项目所需的核心开发工具安装方法。

---

## 1. goenv - Go 版本管理工具

goenv 用于管理多个 Go 版本，类似于 Node.js 的 nvm。

### 安装方法

```bash
# 1. 克隆 goenv 仓库
git clone https://github.com/go-nv/goenv.git ~/.goenv

# 2. 配置环境变量 (添加到 ~/.bashrc 或 ~/.zshrc)
echo 'export GOENV_ROOT="$HOME/.goenv"' >> ~/.bashrc
echo 'export PATH="$GOENV_ROOT/bin:$PATH"' >> ~/.bashrc
echo 'eval "$(goenv init -)"' >> ~/.bashrc
echo 'export PATH="$GOROOT/bin:$PATH"' >> ~/.bashrc
echo 'export PATH="$GOPATH/bin:$PATH"' >> ~/.bashrc

# 3. 重新加载配置
source ~/.bashrc
```

### 常用命令

```bash
# 查看可安装的 Go 版本
goenv install -l

# 安装指定版本
goenv install 1.22.0

# 设置全局版本
goenv global 1.22.0

# 设置当前目录版本 (会创建 .go-version 文件)
goenv local 1.22.0

# 查看当前版本
goenv version

# 查看已安装的版本
goenv versions

# 刷新 shims (安装新的 Go 工具后需要执行)
goenv rehash
```

### 验证安装

```bash
goenv --version
go version
```

---

## 2. just - 命令运行器

just 是一个用 Rust 编写的命令运行器，类似于 Make，但更简单易用。项目使用 justfile 定义所有开发任务。

> 详细命令参考请查看 [justfile-guide.md](./justfile-guide.md)

### 安装方法

```bash
# 通过 npm 安装
npm install -g rust-just

# 或通过 cargo 安装
cargo install just
```

### 常用命令

```bash
# 查看所有可用命令
just --list
just -l

# 运行指定命令
just <recipe>

# 运行带参数的命令
just <recipe> key=value

# 查看命令详情
just --show <recipe>

# 干运行（只显示命令不执行）
just --dry-run <recipe>
```

### 项目常用命令

```bash
# 运行应用
just run
just run force=true   # 强制杀掉占用端口

# 运行测试
just test-unit

# 代码检查
just lint

# 数据库管理
just db up            # 应用 schema
just db status        # 查看状态

# 环境管理
just env-list
just env-switch dev
just env-create dev
```

### 验证安装

```bash
just --version
```

---

## 3. Atlas - 数据库 Schema 管理工具

Atlas 是一个现代化的数据库 Schema 管理工具，支持声明式迁移和版本化迁移。

- 📖 官方文档: https://atlasgo.io/docs
- 📖 版本化迁移: https://atlasgo.io/versioned/diff

### 安装方法

#### 推荐：使用官方安装脚本

```bash
curl -sSf https://atlasgo.sh | sh
```

#### 其他安装方式

```bash
# macOS (Homebrew)
brew install ariga/tap/atlas

# Docker
docker pull arigaio/atlas

# Linux amd64 (手动下载)
curl -Lo atlas https://release.ariga.io/atlas/atlas-linux-amd64-latest
chmod +x atlas
sudo mv atlas /usr/local/bin/
```

### 验证安装

```bash
atlas version
```

---

### 迁移模式

Atlas 支持两种迁移模式：

| 模式 | 说明 | 适用场景 |
|------|------|----------|
| **声明式 (Declarative)** | 直接应用目标 Schema | 开发环境、快速迭代 |
| **版本化 (Versioned)** | 生成迁移文件，逐步执行 | 生产环境、团队协作 |

---

### 声明式迁移 (Declarative)

直接将数据库同步到目标 Schema 状态：

```bash
# 检查当前 schema
atlas schema inspect -u "mysql://user:pass@localhost:3306/dbname"

# 应用 schema (自动计算差异并执行)
atlas schema apply \
  -u "mysql://user:pass@localhost:3306/dbname" \
  --to file://db/schema/mysql/ \
  --dev-url "mysql://user:pass@localhost:3306/dbname_dev"

# 预览变更 (不实际执行)
atlas schema diff \
  --from "mysql://user:pass@localhost:3306/dbname" \
  --to file://db/schema/mysql/
```

---

### 版本化迁移 (Versioned)

生成可追踪的迁移文件，适合团队协作和生产环境。

> 📖 详细文档: https://atlasgo.io/versioned/diff

#### 1. 初始化迁移目录

```bash
# 创建迁移目录
mkdir -p db/migrations
```

#### 2. 生成迁移文件 (atlas migrate diff)

```bash
# 从 Schema 文件生成迁移
atlas migrate diff <migration_name> \
  --dir "file://db/migrations" \
  --to "file://db/schema/mysql/" \
  --dev-url "mysql://user:pass@localhost:3306/dbname_dev"

# 示例：添加用户表
atlas migrate diff add_users_table \
  --dir "file://db/migrations" \
  --to "file://db/schema/mysql/" \
  --dev-url "mysql://root:password@localhost:3306/modelcraft_dev"
```

生成的文件结构：
```
db/migrations/
├── 20240101120000_add_users_table.sql   # 迁移 SQL
└── atlas.sum                             # 校验文件
```

#### 3. 应用迁移

```bash
# 应用所有待执行的迁移
atlas migrate apply \
  --dir "file://db/migrations" \
  --url "mysql://user:pass@localhost:3306/dbname"

# 预览将要执行的迁移 (dry-run)
atlas migrate apply \
  --dir "file://db/migrations" \
  --url "mysql://user:pass@localhost:3306/dbname" \
  --dry-run
```

#### 4. 查看迁移状态

```bash
# 查看迁移历史和状态
atlas migrate status \
  --dir "file://db/migrations" \
  --url "mysql://user:pass@localhost:3306/dbname"
```

#### 5. 迁移回滚

```bash
# 回滚到指定版本
atlas migrate down \
  --dir "file://db/migrations" \
  --url "mysql://user:pass@localhost:3306/dbname" \
  --to-version 20240101120000
```

---

### 常用命令速查

| 命令 | 说明 |
|------|------|
| `atlas schema inspect` | 查看数据库当前 schema |
| `atlas schema apply` | 声明式应用 schema |
| `atlas schema diff` | 比较 schema 差异 |
| `atlas migrate diff` | 生成版本化迁移文件 |
| `atlas migrate apply` | 应用版本化迁移 |
| `atlas migrate status` | 查看迁移状态 |
| `atlas migrate down` | 回滚迁移 |
| `atlas migrate hash` | 重新计算校验和 |

---

### 项目中使用 (通过 just)

```bash
# 创建数据库
just db create

# 应用 schema (声明式)
just db up

# 查看数据库状态
just db status

# 重置数据库
just db reset

# 检查迁移文件规范
just db lint
```

---

### 配置文件 (可选)

可以创建 `atlas.hcl` 简化命令：

```hcl
# atlas.hcl
env "local" {
  src = "file://db/schema/mysql/"
  url = "mysql://root:password@localhost:3306/modelcraft"
  dev = "mysql://root:password@localhost:3306/modelcraft_dev"
  
  migration {
    dir = "file://db/migrations"
  }
}
```

使用配置文件：

```bash
# 使用 local 环境配置
atlas migrate diff add_feature --env local
atlas migrate apply --env local
```

---

## 4. jq - JSON 处理工具

jq 是命令行 JSON 处理器，用于解析、过滤和转换 JSON 数据。

### 安装方法

```bash
# macOS (Homebrew)
brew install jq

# Ubuntu/Debian
sudo apt-get install -y jq

# CentOS/RHEL
sudo yum install -y jq

# 下载二进制文件
curl -sL https://stedolan.github.io/jq/download/jq-linux64 -o jq
chmod +x jq
sudo mv jq /usr/local/bin/
```

### 常用命令

```bash
# 格式化 JSON
echo '{"name":"test"}' | jq '.'

# 提取字段
echo '{"name":"test","age":18}' | jq '.name'

# 过滤数组
echo '[1,2,3,4,5]' | jq '.[] | select(. > 2)'

# 解析 JSON 文件
jq '.users[]' data.json

# 聚合操作
cat data.json | jq 'map(.price) | add'
```

### 验证安装

```bash
jq --version
```

---

## 5. just - 命令运行器

just 是一个用 Rust 编写的命令运行器，类似于 Make，但更简单易用。项目使用 justfile 定义任务。

- 📖 GitHub: https://github.com/casey/just

### 安装方法

```bash
npm install -g rust-just
```

### 常用命令

```bash
# 查看所有可用任务
just --list
just -l

# 运行指定任务
just <task-name>

# 查看任务详情
just --show <task-name>

# 干运行（只显示命令不执行）
just --dry-run <task-name>

# 查看帮助
just --help
```

### 验证安装

```bash
just --version
```

---

## 5. Docker Compose - 容器编排工具

Docker Compose 是用于定义和运行多容器 Docker 应用程序的工具。

### 安装方法

```bash
# 下载最新版 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose

# 添加执行权限
sudo chmod +x /usr/local/bin/docker-compose

# 验证安装
docker-compose --version
```

---

## 快速安装脚本

以下脚本可一键安装所有工具：

```bash
#!/bin/bash

echo "🔧 Installing development tools..."

# Install just
if ! command -v just &> /dev/null; then
    echo "📦 Installing just..."
    npm install -g rust-just
fi

# Install Atlas
if ! command -v atlas &> /dev/null; then
    echo "📦 Installing Atlas..."
    curl -sSf https://atlasgo.sh | sh
fi

# Install goenv
if [ ! -d "$HOME/.goenv" ]; then
    echo "📦 Installing goenv..."
    git clone https://github.com/go-nv/goenv.git ~/.goenv
    echo 'export GOENV_ROOT="$HOME/.goenv"' >> ~/.bashrc
    echo 'export PATH="$GOENV_ROOT/bin:$PATH"' >> ~/.bashrc
    echo 'eval "$(goenv init -)"' >> ~/.bashrc
    echo 'export PATH="$GOROOT/bin:$PATH"' >> ~/.bashrc
    echo 'export PATH="$GOPATH/bin:$PATH"' >> ~/.bashrc
    source ~/.bashrc
fi

# Install Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo "📦 Installing Docker Compose..."
    sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
fi

echo "✅ All tools installed!"
echo ""
echo "Please run: source ~/.bashrc"
```

---

## 常见问题

### Q: goenv install 失败

确保已安装编译依赖：

```bash
# Ubuntu/Debian
sudo apt-get install -y build-essential

# CentOS/RHEL
sudo yum groupinstall -y "Development Tools"
```

### Q: just 命令找不到

检查安装方式：

```bash
# 通过 npm 安装
npm install -g rust-just

# 或通过 cargo 安装 (需要 Rust)
cargo install just

# 检查 PATH
which just
```

### Q: Atlas 连接数据库失败

检查数据库 URL 格式：

```bash
# MySQL
mysql://user:password@host:port/database

# PostgreSQL
postgres://user:password@host:port/database?sslmode=disable
```

---

## 参考链接

- [goenv 官方文档](https://github.com/go-nv/goenv)
- [just 官方文档](https://github.com/casey/just)
- [Atlas 官方文档](https://atlasgo.io/docs)
- [Docker Compose GitHub](https://github.com/docker/compose)
