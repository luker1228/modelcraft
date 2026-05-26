COMPOSE_FILE := compose/docker-compose.local.yml
COMPOSE := $(shell if docker compose version >/dev/null 2>&1; then echo "docker compose"; else echo "docker-compose"; fi)

.PHONY: docker docker-up docker-down docker-build docker-deploy docker-logs docker-ps docker-restart docker-clean

## 启动所有服务（后台）
docker:
	cd deploy && $(COMPOSE) -f $(COMPOSE_FILE) up -d

## 启动所有服务并重新构建镜像
docker-build:
	cd deploy && $(COMPOSE) -f $(COMPOSE_FILE) up -d --build

## 兼容入口：部署并重建镜像
docker-deploy: docker-build

## 停止并移除容器
docker-down:
	cd deploy && $(COMPOSE) -f $(COMPOSE_FILE) down

## 查看运行中的容器状态
docker-ps:
	cd deploy && $(COMPOSE) -f $(COMPOSE_FILE) ps

## 查看所有服务日志（跟踪）
docker-logs:
	cd deploy && $(COMPOSE) -f $(COMPOSE_FILE) logs -f

## 重启所有服务
docker-restart:
	cd deploy && $(COMPOSE) -f $(COMPOSE_FILE) restart

## 停止容器并删除数据卷（危险：会清除 MySQL 数据）
docker-clean:
	cd deploy && $(COMPOSE) -f $(COMPOSE_FILE) down -v

## 单独构建并重启 backend
docker-backend:
	cd deploy && $(COMPOSE) -f $(COMPOSE_FILE) up -d --build backend

## 单独构建并重启 frontend
docker-frontend:
	cd deploy && $(COMPOSE) -f $(COMPOSE_FILE) up -d --build frontend

## 单独构建并重启 agent
docker-agent:
	cd deploy && $(COMPOSE) -f $(COMPOSE_FILE) up -d --build modelcraft-agent

## 启用 tools profile（含 phpmyadmin）
docker-tools:
	cd deploy && $(COMPOSE) -f $(COMPOSE_FILE) --profile tools up -d
