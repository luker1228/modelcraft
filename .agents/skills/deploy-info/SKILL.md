---
name: deploy-info
description: >
  ModelCraft 部署与环境配置速查。当需要了解服务端口、Docker 配置、环境变量、
  健康检查端点、数据库连接等部署相关信息时使用此 skill。
  触发场景包括：
  (1) 用户询问某个服务的端口号（"后端在哪个端口"、"MySQL 端口是多少"），
  (2) 用户需要了解 .env 配置项的含义或默认值，
  (3) 用户需要查看 Docker Compose 服务列表或启动命令，
  (4) 用户需要了解本地开发 vs Docker 部署的环境差异，
  (5) 用户需要了解健康检查、数据库连接等运维信息，
  (6) 用户提到"部署"、"docker-compose"、"服务端口"、"环境配置"等关键词。
globs:
  - "**/.env*"
  - "**/docker-compose*.yml"
  - "**/Dockerfile"
---

# 部署与环境配置速查

## 服务端口

| 服务 | 端口 | 健康检查 |
|------|------|----------|
| **ModelCraft API (Go)** | `8080` | `GET /health` |
| **Frontend (Next.js)** | `3000` | - |
| **MySQL (Docker 部署)** | `6033` → 容器 `3306` | `mysqladmin ping` |
| **MySQL (本地开发)** | `3307` → 容器 `3306` | `mysqladmin ping` |
| **Redis** | `6379` | `redis-cli ping` |
| **phpMyAdmin** | `8081` → 容器 `80` | - |

## 两种 Docker Compose 模式

### 本地开发 (`docker-compose.local.yml`)

仅启动第三方服务，Go 应用通过 `just run` 在宿主机运行：

```bash
docker compose -f docker-compose.local.yml up -d
# 包含：MySQL(3307), Redis(6379), phpMyAdmin(8081, 需 --profile tools)
```

- MySQL 单实例，承载应用数据库
- App 通过 `127.0.0.1:3307` 连接 MySQL

### 完整部署 (`docker-compose.yml`)

所有服务容器化：

```bash
docker compose up -d                          # 基础服务
docker compose --profile tools up -d          # 含 phpMyAdmin
```

- MySQL 独立实例 (`modelcraft-mysql`, 端口 6033)
- App 通过 `modelcraft-mysql:3306` (容器网络) 连接 MySQL
- MySQL 初始化 schema：挂载 `./db/schema/mysql` 到 `/docker-entrypoint-initdb.d`

## 核心环境变量

### 后端 (`.env` / `.env.dev`)

| 变量 | 本地开发默认值 | Docker 默认值 | 说明 |
|------|---------------|---------------|------|
| `PORT` | `8080` | `8080` | 应用端口 |
| `DB_HOST` | `127.0.0.1` | `modelcraft-mysql` | MySQL 主机 |
| `DB_PORT` | `3307` | `3306` | MySQL 端口 |
| `DB_USERNAME` | `root` | `modelcraft` | MySQL 用户 |
| `DB_PASSWORD` | `modelcraft123` | *(必填)* | MySQL 密码 |
| `DB_NAME` / `DB_DATABASE` | `modelcraft` | `modelcraft` | 数据库名 |
| `DB_MIGRATE_ON_STARTUP` | - | `true` | Docker 启动时自动迁移 |
| `REDIS_HOST` | `localhost` | `modelcraft-redis` | Redis 主机 |
| `REDIS_PORT` | `6379` | `6379` | Redis 端口 |
| `JWT_SECRET` | *(必填)* | *(必填)* | JWT 签名密钥 |
| `CRYPTO_AES_KEY` | `12345678901234567890123456789012` | *(必填, 32字节)* | AES-256 密钥 |
| `CASDOOR_ENDPOINT` | `http://9.135.32.8:8000` | - | Casdoor 地址 |
| `CASDOOR_CLIENT_ID` | `765f754c2a59662a442b` | *(必填)* | OAuth Client ID |
| `CASDOOR_CLIENT_SECRET` | `b09e8443c481f8545bab5312e9ffaa47f88179e0` | *(必填)* | OAuth Client Secret |
| `AUTH_DESIGN_ENABLED` | `true` | - | 设计时鉴权开关 |

### 前端 (`.env`)

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `VITE_BACKEND_URL` | `http://localhost:8080` | 后端 API 地址 |
| `NEXT_PUBLIC_CASDOOR_URL` | `http://9.135.32.8:8000` | Casdoor (浏览器可访问) |
| `NEXT_PUBLIC_CASDOOR_CLIENT_ID` | `765f754c2a59662a442b` | OAuth Client ID |
| `NEXT_PUBLIC_CASDOOR_ORGANIZATION` | `built-in` | Casdoor 组织 |
| `NEXT_PUBLIC_CASDOOR_APP_NAME` | `modelcraft` | Casdoor 应用名 |
| `VITE_COPILOT_ENABLED` | `true` | CopilotKit AI 开关 |

> `NEXT_PUBLIC_*` 变量在构建时嵌入，暴露给浏览器。

### 测试环境 (`.env.autotest`)

与 `.env.dev` 相同，仅 `DB_NAME=modelcraft_test`。

## 配置文件

| 文件 | 位置 | 说明 |
|------|------|------|
| 后端环境变量 | `modelcraft-backend/.env` | 活跃配置（本地开发） |
| 后端环境模板 | `modelcraft-backend/.env.dev.example` | 开发环境模板 |
| Docker 环境模板 | `modelcraft-backend/.env.docker.example` | 最完整的变量参考 |
| 后端 YAML 配置 | `modelcraft-backend/configs/config.yaml` | 应用配置（端口、连接池、日志等） |
| Casdoor 配置 | `modelcraft-backend/casdoor/conf/app.conf` | Casdoor 端口和数据库 |
| 前端环境变量 | `modelcraft-front/.env` | 活跃配置 |
| 前端生产环境 | `modelcraft-front/.env.production` | 生产构建（`VITE_API_BASE_URL` 为空，同源） |

## config.yaml 关键配置

```yaml
server:
  port: 8080
  mode: debug              # debug / release

database:
  host: localhost
  port: 3306
  migrate_on_startup: true
  max_open_conns: 100
  max_idle_conns: 10

redis:
  host: localhost
  port: 6379

jwt:
  expiration: 21600s       # 6 小时
  issuer: modelcraft

auth:
  design:
    enabled: false          # 设计时 API 鉴权（生产应为 true）
  runtime:
    enabled: true           # 运行时 API 始终鉴权
  casdoor:
    endpoint: http://9.135.32.8:8000

logger:
  level: info
  output_path: logs/server.log
  max_size: 100             # MB
  max_backups: 10
  max_age: 7                # 天
  compress: true
```

## 常用运维命令

```bash
# 本地开发启动
docker compose -f docker-compose.local.yml up -d
just run

# 完整 Docker 部署
docker compose up -d

# 查看 Docker 服务状态
docker compose ps
docker compose --profile tools ps

# 查看应用日志
docker compose logs -f modelcraft
just logs

# 重置数据库
just db reset

# 登录 MySQL CLI
just db login
```

## Docker 构建信息

- **基础镜像**: `golang:1.25-alpine` (构建) → `alpine:latest` (运行)
- **CGO**: 启用（MySQL 驱动需要）
- **入口**: `./main -config /app/configs/config.yaml`
- **运行用户**: `appuser:appgroup` (UID/GID 1001)
- **时区**: `Asia/Shanghai`
- **Dockerfile 位置**: `modelcraft-backend/Dockerfile`

## Docker 网络与数据卷

| 资源 | 本地开发 | 完整部署 |
|------|----------|----------|
| 网络 | `modelcraft-local-network` | `modelcraft-network` |
| MySQL 数据 | `mysql_local_data` | `mysql_data` |
| Redis 数据 | `redis_local_data` | `redis_data` |
| Casdoor 数据 | - | `casdoor_data`, `casdoor_mysql_data` |
| 应用挂载 | - | `./configs:/app/configs`, `./logs:/app/logs`, `./data:/app/data` |
