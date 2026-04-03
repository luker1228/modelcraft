# ModelCraft - 模型设计平台

ModelCraft 是一个基于 Go 和 GraphQL 的现代化模型设计平台，提供可视化的数据模型设计、管理和部署功能。

## 🚀 快速开始

### 前置要求

- Docker 20.10+
- Docker Compose 2.0+
- 至少 2GB 可用内存

### 一键启动（默认模式）

使用内置的 MySQL 容器快速启动：

```bash
# 构建并启动所有服务
docker compose up -d

# 查看服务状态
docker compose ps

# 查看应用日志
docker compose logs -f modelcraft
```

服务启动后访问：http://localhost:8080

## 📋 MySQL 配置指南

ModelCraft 支持两种 MySQL 部署模式：

### 1. 内置 MySQL 模式（默认）

使用 Docker Compose 中内置的 MySQL 容器，适合开发和测试环境。

**优点：**
- 一键启动，无需额外配置
- 自动数据持久化
- 内置数据库迁移

### 2. 外部 MySQL 模式（生产环境推荐）

使用外部已部署的 MySQL 实例，适合生产环境。

## 🔧 使用外部 MySQL 数据库

### 配置加载机制

ModelCraft 使用 **godotenv** 库加载环境变量，配置加载优先级为：

1. **系统环境变量**（最高优先级）
2. **`.env` 文件**（由 godotenv 加载）
3. **`config.yaml` 默认值**（最低优先级）

这种设计确保了灵活性和安全性：
- 开发环境使用 `.env` 文件
- 生产环境使用系统环境变量或 Docker 环境变量
- 敏感信息永不提交到代码仓库

### 方法一：通过环境变量配置（推荐）

创建或编辑 `.env` 文件：

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑 .env 文件，添加以下配置
```

在 `.env` 文件中配置外部MySQL：

```bash
# ============================================
# 外部 MySQL 配置
# ============================================

# 启用外部 MySQL 模式
USE_EXTERNAL_MYSQL=true

# 外部 MySQL 连接配置
EXTERNAL_MYSQL_HOST=your-mysql-host.example.com
EXTERNAL_MYSQL_PORT=3306
EXTERNAL_MYSQL_USER=your-username
EXTERNAL_MYSQL_PASSWORD=your-password
EXTERNAL_MYSQL_DATABASE=modelcraft

# ============================================
# 其他相关配置
# ============================================

# 应用配置
GIN_MODE=release
LOG_LEVEL=info

# Redis 配置
REDIS_HOST=modelcraft-redis
REDIS_PORT=6379
REDIS_PASSWORD=

# 安全配置（生产环境必须修改）
JWT_SECRET=your-production-jwt-secret-key
CRYPTO_AES_KEY=your-32-byte-aes-encryption-key

# 时区
TZ=Asia/Shanghai

# ============================================
# 特殊场景配置
# ============================================

# 如果外部 MySQL 在本地宿主机，使用以下配置
# EXTERNAL_MYSQL_HOST=host.docker.internal

# 如果外部 MySQL 在 Docker 网络中，使用服务名
# EXTERNAL_MYSQL_HOST=your-mysql-service-name

# 如果使用 SSL 连接
# EXTERNAL_MYSQL_SSL=true
```

**配置说明：**
- 复制 `.env.example` 为 `.env` 并修改上述配置
- 所有以 `EXTERNAL_MYSQL_` 开头的变量仅在 `USE_EXTERNAL_MYSQL=true` 时生效
- 安全相关的密钥（JWT_SECRET、CRYPTO_AES_KEY）在生产环境必须修改
启动服务（跳过内置 MySQL 容器）：

```bash
# 使用 --scale 跳过 MySQL 容器
docker compose --scale modelcraft-mysql=0 up -d

# 或者使用 profile 机制
docker compose --profile tools up -d modelcraft modelcraft-redis
```

### 方法二：通过 Docker Compose 环境变量

直接在启动命令中设置环境变量：

```bash
docker compose --scale modelcraft-mysql=0 up -d \
  -e USE_EXTERNAL_MYSQL=true \
  -e EXTERNAL_MYSQL_HOST=your-mysql-host \
  -e EXTERNAL_MYSQL_USER=your-username \
  -e EXTERNAL_MYSQL_PASSWORD=your-password \
  -e EXTERNAL_MYSQL_DATABASE=modelcraft
```

### 方法三：修改配置文件（推荐）

编辑 `configs/config.docker.yaml` 文件，启用自动迁移功能：

```yaml
# 数据库配置
database:
  type: "mysql"
  host: "your-mysql-host.example.com"
  port: 3306
  username: "your-username"
  password: "your-password"
  database: "modelcraft"
  charset: "utf8mb4"
  
  # 连接池配置
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600
  
  # 启用自动数据库迁移（推荐）
  migrate_on_startup: true
  log_level: "info"

# 其他配置
server:
  port: "8080"
  mode: "release"

logger:
  level: "info"
  output_path: "stdout"
```

**自动迁移的优势：**
- 应用启动时自动检查并执行数据库迁移
- 无需手动执行SQL脚本
- 支持增量迁移，不会重复执行已应用的变更
- 迁移失败时应用会停止启动，确保数据一致性

## 🗄️ 外部 MySQL 数据库初始化

在使用外部 MySQL 之前，需要手动初始化数据库：

### 1. 创建数据库和用户

在外部 MySQL 实例上执行：

```sql
-- 创建数据库
CREATE DATABASE IF NOT EXISTS modelcraft 
DEFAULT CHARACTER SET utf8mb4 
COLLATE utf8mb4_unicode_ci;

-- 创建用户并授权
CREATE USER IF NOT EXISTS 'modelcraft'@'%' 
IDENTIFIED BY 'your-password';

GRANT ALL PRIVILEGES ON modelcraft.* TO 'modelcraft'@'%';
FLUSH PRIVILEGES;
```

### 2. 执行数据库迁移

从项目中的Schema文件初始化数据库结构：

```bash
# 方法1：按顺序执行所有Schema文件
mysql -h your-mysql-host -u your-username -p your-database < db/schema/mysql/01_project.sql
mysql -h your-mysql-host -u your-username -p your-database < db/schema/mysql/02_database_cluster.sql
mysql -h your-mysql-host -u your-username -p your-database < db/schema/mysql/03_model_domain.sql

# 方法2：批量执行所有Schema文件
for file in db/schema/mysql/*.sql; do
    echo "执行Schema文件: $file"
    mysql -h your-mysql-host -u your-username -p your-database < "$file"
done

# 方法3：使用应用自动迁移（推荐）
# 在配置文件中启用自动迁移，应用启动时会自动执行
# database:
#   migrate_on_startup: true
```

**Schema文件说明：**
- `01_project.sql` - 创建项目表并插入默认项目
- `02_database_cluster.sql` - 创建数据库集群表
- `03_model_domain.sql` - 创建模型、字段、关系、枚举等核心表

**执行顺序很重要**，请按数字顺序执行Schema文件。

## 🌐 网络连接配置

### 本地宿主机 MySQL

如果 MySQL 运行在 Docker 宿主机上：

```bash
# 使用 Docker 的特殊主机名
EXTERNAL_MYSQL_HOST=host.docker.internal
EXTERNAL_MYSQL_PORT=3306
```

### 远程服务器 MySQL

确保 Docker 宿主机可以访问远程 MySQL：

```bash
# 测试网络连通性
ping your-mysql-host.example.com

# 测试端口访问
telnet your-mysql-host.example.com 3306

# 测试 MySQL 连接
mysql -h your-mysql-host -u your-username -p
```

### Docker 网络配置

如果需要自定义网络：

```yaml
# 在 docker-compose.yml 中添加自定义网络
networks:
  modelcraft-network:
    driver: bridge
  external-network:
    external: true
    name: your-custom-network
```

## 🔒 安全配置建议

### 生产环境安全

1. **修改默认密码**：
   ```bash
   # 生成强密码
   openssl rand -base64 32
   ```

2. **使用 SSL 连接**：
   ```yaml
   database:
     ssl_mode: "require"
     ssl_ca: "/path/to/ca-cert.pem"
     ssl_cert: "/path/to/client-cert.pem"
     ssl_key: "/path/to/client-key.pem"
   ```

3. **网络隔离**：
   ```bash
   # 使用自定义网络而非默认 bridge
   docker network create modelcraft-network
   ```

## 🐛 常见问题排查

### 连接失败

**错误信息**：`dial tcp: lookup your-mysql-host: no such host`

**解决方案**：
- 检查主机名是否正确
- 确保网络连通性
- 如果是本地宿主机 MySQL，使用 `host.docker.internal`

### 认证失败

**错误信息**：`Access denied for user`

**解决方案**：
- 检查用户名和密码
- 确认用户有数据库访问权限
- 检查 MySQL 的访问控制设置

### 数据库不存在

**错误信息**：`Unknown database 'modelcraft'`

**解决方案**：
- 手动创建数据库
- 执行数据库迁移脚本

## � 使用外部MySQL的最佳实践

### 选择建议

**内置MySQL（推荐用于）：**
- 本地开发和测试环境
- 快速原型验证
- 小规模部署

**外部MySQL（推荐用于）：**
- 生产环境部署
- 已有数据库基础设施
- 需要高可用和备份的场景
- 多个应用共享数据库

### 配置优先级

配置按以下优先级生效：
1. **环境变量**（最高优先级）- 推荐用于生产环境
2. **配置文件** - 推荐用于开发环境
3. **Docker Compose默认值** - 内置MySQL模式

### 安全建议

1. **密码管理**：
   - 使用强密码生成器
   - 定期更换数据库密码
   - 使用环境变量而非硬编码

2. **网络配置**：
   - 配置防火墙规则
   - 使用SSL连接
   - 限制数据库访问IP范围

3. **备份策略**：
   - 定期备份数据库
   - 测试恢复流程
   - 监控数据库性能

### 监控和维护

1. **健康检查**：
   ```bash
   # 检查应用健康状态
   curl http://localhost:8080/health
   
   # 检查数据库连接
   docker compose exec modelcraft sh -c "mysql -h $DATABASE_HOST -u $DATABASE_USERNAME -p$DATABASE_PASSWORD -e 'SELECT 1'"
   ```

2. **日志监控**：
   ```bash
   # 查看应用日志
   docker compose logs -f modelcraft
   
   # 查看数据库连接日志
   docker compose logs modelcraft | grep -i "database\|mysql"
   ```

## �📚 更多资源

- [详细部署文档](docs/DEPLOYMENT.md)
- [Docker 配置文档](DOCKER.md)
- [API 文档](docs/03-runtime/api-guide.md)
- [GraphQL 示例](docs/graphql-examples.md)
- [数据库管理指南](db/README.md)

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

本项目基于 [LICENSE](LICENSE) 文件中的许可证。