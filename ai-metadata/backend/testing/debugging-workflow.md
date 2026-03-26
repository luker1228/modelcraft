# 🔧 开发调试流程

> 日常开发调试的关键命令和问题排查方法。

---

## 1️⃣ 本地运行

```bash
just run                              # 后台运行
just run true                         # 强制运行（先停止现有进程）
just stop                             # 停止服务
just logs                             # 查看日志
```

**说明**:
- 后台运行，日志输出到 `logs/server.log`
- `just run true` - 自动停止现有进程再启动（推荐使用）

## 3️⃣ 数据库问题

### 常用命令

```bash
just db status                        # 查看状态
just db up                            # 应用 Schema
just db reset                         # 重置数据库（⚠️ 删除所有数据）
just db login                         # 登录 MySQL CLI
```

### 使用不同环境

```bash
just db status .env.autotest          # 测试环境
just db reset .env.autotest
```

### 常见问题

**问题 1**: 连接超时 `connection refused`

```bash
just deploy-infra status              # 检查 MySQL 是否运行
just deploy-infra start               # 启动 MySQL
```

**问题 2**: 数据库不存在 `Unknown database`

```bash
just db up                            # 自动创建并应用 Schema
```

**问题 3**: Schema 冲突

```bash
just db reset                         # 重置数据库
```

---

## 4️⃣ 常见问题快速修复

| 问题 | 解决方案 |
|------|----------|
| 端口被占用 | `just run --force` 或 `just run force=true` |
| GraphQL Schema 不生效 | `just generate-gql && just restart` |
| 依赖包缺失 | `just deps` |
| 测试失败 | `just test-unit-verbose` 或 `just db reset .env.autotest` |
| 代码生成失败 | `just install-tools` |

---

### Request ID 追踪

```bash
# 1. 发起请求时设置自定义 request_id
curl -i http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: my_debug_request_001" \
  -d '{"query":"..."}'

# 2. 查看完整调用链路
just log-cat my_debug_request_001
```
```