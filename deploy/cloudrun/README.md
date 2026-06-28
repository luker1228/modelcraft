# ModelCraft 微信云托管部署指南

## 架构

```
微信云托管
 ┌──────────────────────────────────────────────────┐
 │                                                  │
 │  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
 │  │ frontend  │  │  apisix   │  │  agent   │       │
 │  │ (Next.js) │  │ (Gateway) │  │ (Python) │       │
 │  │  :80      │  │  :80      │  │  :80     │       │
 │  └─────┬─────┘  └────┬─────┘  └────┬─────┘       │
 │        │             │              │              │
 │        └─────┬───────┘              │              │
 │              │                      │              │
 │        ┌─────▼─────┐               │              │
 │        │  backend   │◄──────────────┘              │
 │        │  (Go API)  │                              │
 │        │  :80       │                              │
 │        └─────┬─────┘                               │
 │              │                                      │
 └──────────────┼──────────────────────────────────────┘
                │
       ┌────────▼────────┐
       │  自建 MySQL       │
       │  (CVM / 外网)     │
       └─────────────────┘
```

## 服务清单

| # | 服务名 | 技术栈 | Dockerfile 位置 | 默认端口 |
|---|--------|--------|----------------|---------|
| 1 | `backend` | Go | `modelcraft-backend/Dockerfile` | 8080 |
| 2 | `apisix` | APISIX 3.9 | `deploy/cloudrun/apisix/Dockerfile` | 9080 |
| 3 | `frontend` | Next.js | `modelcraft-front/Dockerfile` | 3000 |
| 4 | `agent` | Python | `modelcraft-agent/Dockerfile` | 8000 |

> 云托管会注入 `PORT=80`，所有服务自动监听 80 端口。

## 部署步骤

### 1. 前置条件

- [ ] 微信云托管环境已创建
- [ ] 自建 MySQL 可从公网或通过云托管 VPC 访问
- [ ] 云托管已绑定 Git 仓库（或准备好手动构建）

### 2. 数据库初始化

在自建 MySQL 上创建数据库并执行 Schema 迁移：

```bash
# 连接 MySQL
mysql -h <your-mysql-host> -u root -p

# 创建数据库
CREATE DATABASE IF NOT EXISTS modelcraft CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'modelcraft'@'%' IDENTIFIED BY '<password>';
GRANT ALL PRIVILEGES ON modelcraft.* TO 'modelcraft'@'%';
FLUSH PRIVILEGES;
```

Schema 迁移由 Backend 在启动时自动执行（`DB_MIGRATE_ON_STARTUP=true`），
也可以通过 Atlas 手动执行：

```bash
cd deploy
atlas schema apply \
  -u "mysql://modelcraft:<password>@<host>:3306/modelcraft" \
  --to "file://../modelcraft-backend/db/schema/mysql/" \
  --dev-url "mysql://modelcraft:<password>@<host>:3306/modelcraft_dev" \
  --auto-approve
```

### 3. 服务部署顺序

**必须按顺序部署**，因为存在依赖关系：

```
1. backend  ──→  2. apisix  ──→  3. frontend + agent (可并行)
```

| 步骤 | 服务 | 构建上下文 | Dockerfile 路径 |
|------|------|-----------|-----------------|
| 1 | backend | `modelcraft-backend/` | `Dockerfile` |
| 2 | apisix | `deploy/cloudrun/apisix/` | `Dockerfile` |
| 3 | frontend | `modelcraft-front/` | `Dockerfile` |
| 4 | agent | `modelcraft-agent/` | `Dockerfile` |

### 4. 环境变量配置

在云托管控制台「服务配置 → 环境变量」中，逐服务添加环境变量。

模板文件位于 `deploy/cloudrun/env/`：

| 服务 | 模板 |
|------|------|
| backend | `env/backend.env` |
| apisix | `env/apisix.env` |
| frontend | `env/frontend.env` |
| agent | `env/agent.env` |

> 将模板中 `<...>` 占位符替换为实际值。

### 5. 构建配置

#### 云托管控制台 / Git 自动构建

每个服务在云托管控制台配置：

| 服务 | 构建目录 | Dockerfile | 输出 |
|------|---------|-----------|------|
| backend | `modelcraft-backend/` | `modelcraft-backend/Dockerfile` | 镜像 |
| apisix | `deploy/cloudrun/apisix/` | `deploy/cloudrun/apisix/Dockerfile` | 镜像 |
| frontend | `modelcraft-front/` | `modelcraft-front/Dockerfile` | 镜像 |
| agent | `modelcraft-agent/` | `modelcraft-agent/Dockerfile` | 镜像 |

#### 本地构建推送

如果你使用 `docker build` + 手动推送镜像：

```bash
# Backend
docker build -t ccr.ccs.tencentyun.com/tcb-100021144463-xnqx/backend:latest \
  -f modelcraft-backend/Dockerfile modelcraft-backend/

# APISIX
docker build -t ccr.ccs.tencentyun.com/tcb-100021144463-xnqx/apisix:latest \
  -f deploy/cloudrun/apisix/Dockerfile deploy/cloudrun/apisix/

# Frontend
docker build -t ccr.ccs.tencentyun.com/<namespace>/modelcraft-frontend:latest \
  -f modelcraft-front/Dockerfile modelcraft-front/

# Agent
docker build -t ccr.ccs.tencentyun.com/<namespace>/modelcraft-agent:latest \
  -f modelcraft-agent/Dockerfile modelcraft-agent/
```

### 6. 服务间通信配置

云托管的服务通过 **K8s Service DNS** 互访（格式：`http://<service-name>:80`）。

关键配置：

| 服务 | 环境变量 | 值 |
|------|---------|-----|
| frontend | `BACKEND_URL` | `http://apisix:80` |
| agent | `GATEWAY_URL` | `http://apisix:80` |
| apisix | upstream backend | `http://backend:80` |

> **注意**：`apisix.yaml` 中 upstream backend 的地址需要从容器名改为云托管服务名。
> 参见下方「APISIX 配置调整」。

## APISIX 配置调整

### CORS 允许域名

`apisix.yaml` 中的 CORS `allow_origins` 需要改为云托管前端域名：

```yaml
# 原来（本地开发）
allow_origins: "http://localhost:3000"

# 改为（云托管）
allow_origins: "https://<frontend-service>.<env-id>.ap-shanghai.run.tcloudbase.com"
```

### Backend Upstream

确保 `apisix.yaml` 中 upstream 指向正确的后端地址：

```yaml
upstreams:
  - id: backend
    nodes:
      "backend:80": 1   # 云托管 K8s Service DNS
```

> 当前配置使用 `upstream_id: backend`，如果在 `apisix.yaml` 中已定义 upstream，只需确认 nodes 地址正确。

## 健康检查

所有服务都配置了 Docker HEALTHCHECK，云托管会自动探测：

| 服务 | 健康检查端点 | 探测间隔 |
|------|------------|---------|
| backend | `GET /health` | 30s |
| frontend | `GET /` | 30s |
| agent | `GET /healthz` | 30s |
| apisix | `GET /` | 15s |

## 常见问题

### Q: 构建 APISIX 镜像时找不到 config.yaml

APISIX 的 `config.yaml` 和 `apisix.yaml` 在仓库根目录 `apisix/` 下。
构建时需要先将它们复制到 `deploy/cloudrun/apisix/`：

```bash
cp apisix/config.yaml apisix/apisix.yaml deploy/cloudrun/apisix/
```

### Q: Backend 连接不上自建 MySQL

1. 确认 MySQL 允许云托管出口 IP 访问（防火墙/安全组规则）
2. 确认 `DB_HOST` 环境变量设置正确（使用 IP 或可解析域名）
3. 查看 Backend 日志排查连接错误

### Q: CORS 报错

1. 确认 `apisix.yaml` 中 CORS `allow_origins` 包含前端实际域名
2. 确认 `FRONTEND_URL` 环境变量正确
3. 重新部署 APISIX 服务使配置生效
