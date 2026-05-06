# 🚀 部署指南

> **优先级: 中低** - 定义如何将应用部署到各种环境。

## 概述

部署指南描述了如何配置环境、构建应用、部署到不同环境，以及日常运维操作。

---

## 🐳 Docker 环境要求 (强制)

> **⚠️ 重要：本项目所有基础服务必须使用 Docker 运行，禁止本地安装！**

### 强制规则

| 服务 | 要求 | 说明 |
|------|------|------|
| **MySQL** | ✅ 必须 Docker | 禁止使用本地安装的 MySQL |
| **Redis** | ✅ 必须 Docker | 禁止使用本地安装的 Redis |
| **其他中间件** | ✅ 必须 Docker | MinIO、Kafka 等均需 Docker |

### 为什么禁止本地安装？

1. **环境一致性** - Docker 确保开发、测试、生产环境完全一致
2. **版本控制** - 通过 `docker-compose.yml` 锁定版本，避免版本差异
3. **隔离性** - 不污染本地系统，多项目互不干扰
4. **可复现** - 新成员一条命令即可启动完整环境
5. **易清理** - `docker-compose down -v` 即可完全清理

### 检查 Docker 环境

```bash
# 检查 Docker 是否安装
docker --version

# 检查 Docker Compose 是否安装
docker compose version

# 检查 Docker 服务状态
docker info
```

### 启动基础服务

```bash
# 启动基础设施 (MySQL, Redis)
task deploy:infra

# 查看基础设施状态
task deploy:infra -- --status

# 停止基础设施
task deploy:infra -- --stop
```

### ❌ 错误做法

```bash
# ❌ 禁止：本地安装 MySQL
apt install mysql-server
brew install mysql

# ❌ 禁止：本地安装 Redis
apt install redis-server
brew install redis

# ❌ 禁止：连接本地安装的数据库
DATABASE_URL="localhost:3306"  # 应使用 Docker 容器
```

### ✅ 正确做法

```bash
# ✅ 正确：先切换环境，再启动服务
task env:switch NAME=dev
task deploy:all

# ✅ 正确：连接 Docker 容器中的服务
DATABASE_URL="127.0.0.1:3306"  # Docker 映射端口
REDIS_URL="127.0.0.1:6379"    # Docker 映射端口
```

---

## 📚 文档列表

| 文档 | 说明 |
|------|------|
| [environments.md](./environments.md) | 环境配置说明 |
| [docker-deployment.md](./docker-deployment.md) | Docker 部署指南 |
| [local-development.md](./local-development.md) | 本地开发环境 |
| [production.md](./production.md) | 生产环境部署 |
| [troubleshooting.md](./troubleshooting.md) | 故障排除 |

## 🌍 环境概览

| 环境 | 用途 | 配置文件 |
|------|------|----------|
| local | 本地开发 | `.env.dev` |
| staging | 预发布测试 | `.env.staging` |
| production | 生产环境 | `.env.prod` |

## 🔄 部署流程

```
┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐
│  开发   │ →  │  测试   │ →  │  构建   │ →  │  部署   │
└─────────┘    └─────────┘    └─────────┘    └─────────┘
     │              │              │              │
     ▼              ▼              ▼              ▼
  编写代码      运行测试       构建镜像      部署服务
  本地验证      CI 流水线      推送仓库      健康检查
```

## 🚀 快速部署

### 本地开发环境

```bash
# 1. 检查 Docker 环境 (必须！)
docker --version
docker compose version

# 2. 配置环境变量
task env:create NAME=dev
task env:switch NAME=dev

# 3. 启动所有服务
task deploy:all

# 或分层启动：
task deploy:infra    # 仅启动 MySQL, Redis
task deploy:app      # 仅启动应用 (AuthProvider, ModelCraft)

# 4. 查看部署状态
task deploy:all -- --status
```

### 常用部署命令

```bash
# 启动服务
task deploy:all                    # 启动所有
task deploy:infra                  # 仅基础设施
task deploy:app                    # 仅应用

# 查看状态
task deploy:all -- --status
task deploy:infra -- --status
task deploy:app -- --status

# 停止服务
task deploy:all -- --stop
task deploy:infra -- --stop
task deploy:app -- --stop

# 重启服务
task deploy:all -- --restart
task deploy:infra -- --restart
task deploy:app -- --restart
```

## 🛡️ Gateway 部署联调检查项（强制）

> **硬性约束**：前端请求必须先经过 Gateway，再由 Gateway 转发到 Backend；前端禁止直连 Backend。

### 链路要求

- [ ] 前端 `BACKEND_URL` 配置为 Gateway 地址（如 `http://<gateway-host>:8090`），不能配置为 Backend 直连地址（如 `:8080`）
- [ ] Gateway `BACKEND_URL` 指向 Backend 内网地址
- [ ] Browser/前端网络请求中，不应出现直连 Backend 的请求
- [ ] Backend 不对前端公网暴露直连入口（仅允许 Gateway 访问）

### 联调验收步骤

```bash
# 1) Gateway 健康检查
curl -i http://<gateway-host>:8090/healthz

# 2) Developer 体系联调
# 通过前端 /api/auth/* -> Gateway /api/tenant/auth/*，确认登录/刷新正常

# 3) EndUser 体系联调
# 通过前端 /api/bff/org/{orgName}/end-user/auth/* -> Gateway /api/end-user/auth/*，确认登录/refresh 正常

# 4) 反向验证：禁止前端直连 Backend
# 将前端 BACKEND_URL 临时改为 Backend 直连地址后，请求应视为配置错误并回退。
```

### 常见错误

- 前端 `.env` 将 `BACKEND_URL` 配成 Backend 地址（绕过 Gateway）
- Nginx/Ingress 将 Backend 8080 暴露给浏览器侧流量
- 联调只测通了功能，但未检查是否经过 Gateway 链路

## 📋 部署检查清单

### 部署前

- [ ] 所有测试通过
- [ ] 环境变量已配置
- [ ] 数据库迁移已准备
- [ ] 配置文件已更新

### 部署后

- [ ] 健康检查通过
- [ ] 日志无异常错误
- [ ] 关键功能验证通过
- [ ] 监控告警正常

## 🔧 常用运维命令

```bash
# 查看服务状态
task status

# 查看日志
task logs

# 重启服务
task restart

# 数据库迁移
task db:migrate-up
```

## 📖 阅读顺序

1. **先确保 Docker 环境就绪** (见上方 Docker 环境要求)
2. 再阅读 `local-development.md` 搭建本地环境
3. 阅读 `environments.md` 理解环境配置
4. 按需阅读其他部署相关文档

## ⚠️ 前置要求

### 必备软件

| 软件 | 版本要求 | 必须 |
|------|----------|------|
| **Docker** | 20.10+ | ✅ 必须 |
| **Docker Compose** | 2.0+ | ✅ 必须 |
| Go | 见 `.go-version` | ✅ 必须 |
| Task | 最新版 | ✅ 必须 |

> **注意**: Docker 和 Docker Compose 是强制要求，没有 Docker 环境无法运行项目！

### 阅读前置

阅读本目录前，请确保已理解：
- [2-development](../2-development/) - 开发规范
- [3-testing](../3-testing/) - 测试策略
- [5-tools](../5-tools/) - 工具使用

## 🔗 相关文档

- 工具安装请参考 [5-tools](../5-tools/)
- 架构理解请参考 [2-development](../2-development/)
