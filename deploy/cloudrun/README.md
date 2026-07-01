# ModelCraft Cloud Run 部署

云托管现在走统一部署结构：

- 根目录只有一份 [`docker-compose.yml`](/data/home/lukemxjia/modelcraft/deploy/docker-compose.yml)
- 环境配置放在 `deploy/configs/cloudrun/*.yaml`
- 构建镜像时通过 `APP_ENV=cloudrun` 把对应 YAML 打进镜像

## 目录约定

```text
deploy/
  docker-compose.yml
  images/
  configs/
    cloudrun/
      backend.yaml
      apisix.yaml
      frontend.yaml
      agent.yaml
  cloudrun/
    apisix/
      Dockerfile
      docker-entrypoint.sh
```

## 使用方式

本地验证 Cloud Run 形态：

```bash
cd deploy
just check cloudrun
just build cloudrun
just up cloudrun
```

等价命令会把以下变量注入统一 compose：

- `APP_ENV=cloudrun`
- `BACKEND_CONTAINER_PORT=80`
- `APISIX_CONTAINER_PORT=80`
- `AGENT_CONTAINER_PORT=80`
- `FRONTEND_CONTAINER_PORT=80`

## 配置来源

各服务最终运行配置都来自镜像内 YAML：

- `configs/cloudrun/backend.yaml`
- `configs/cloudrun/apisix.yaml`
- `configs/cloudrun/frontend.yaml`
- `configs/cloudrun/agent.yaml`

其中：

- backend 直接使用入镜像的 `config.yaml`
- apisix 在启动时读取 YAML 并渲染 `apisix.yaml`
- frontend 和 agent 在启动时从 YAML 导出当前进程环境变量

## 注意事项

- Cloud Run 前端域名需要填入 [configs/cloudrun/apisix.yaml](/data/home/lukemxjia/modelcraft/deploy/configs/cloudrun/apisix.yaml)
- `LLM_API_KEY` 等敏感字段当前仍在 YAML 中，后续应迁移到安全密钥管理方案
- Cloud Run 镜像的内网互访地址约定为 `backend:80`、`apisix:80`、`agent:80`
