# API Contract 共享规范

> **优先级: 中** - 定义后端与前端之间 API Contract 的共享机制和工作流程。

## 概述

后端 `api/` 目录是 API Contract 的**唯一真相源**，通过 **git subtree** 与前端共享，消除手动同步。

## 架构

```
modelcraft-backend/api/          ← 唯一真相源（后端仓库）
        │
        │  git subtree push --prefix=api
        ▼
modelcraft-api-contracts        ← 共享仓库（git.woa.com/lukemxjia/modelcraft-api-contracts.git）
        │
        │  git subtree add/pull --prefix=contract
        ▼
modelcraft-front/contract/      ← 前端消费端（只读）
```

### 目录内容

```
api/  (backend: modelcraft-backend/api/)
├── graph/
│   ├── org/schema/              # Org 域 GraphQL Schema（5 个 .graphql 文件）
│   │   ├── base.graphql
│   │   ├── permission.graphql
│   │   ├── project.graphql
│   │   ├── schema.graphql
│   │   └── user_management.graphql
│   └── project/schema/          # Project 域 GraphQL Schema（7 个 .graphql 文件）
│       ├── base.graphql
│       ├── cluster.graphql
│       ├── enum.graphql
│       ├── field.graphql
│       ├── logical_foreign_key.graphql
│       ├── model.graphql
│       └── schema.graphql
└── openapi/                     # REST API OpenAPI 规范
    ├── auth.yaml
    ├── common.yaml
    ├── openapi-root.yaml
    ├── org.yaml
    ├── user.yaml
    ├── webhook.yaml
    ├── oapi-codegen.yaml        # 后端专属：代码生成器配置
    ├── openapi.yaml             # 后端专属：合并后的完整 spec
    ├── examples/                # 后端专属：示例数据
    └── README.md
```

### 共享仓库配置

| 项目 | Remote 名称 | Subtree 前缀 | Squash |
|------|-------------|-------------|--------|
| Backend (`modelcraft-go`) | `contracts` | `api/` | 否 |
| Frontend (`modelcraft-front`) | `contracts` | `contract/` | 是 |

共享仓库 URL: `https://git.woa.com/lukemxjia/modelcraft-api-contracts.git`

---

## 日常同步工作流

### 后端修改 API 后

```bash
# 1. 正常修改 api/ 下的文件并提交
git add api/
git commit -m "feat(api): update model.graphql with new fields"

# 2. 推送到共享仓库
git subtree push --prefix=api contracts main
```

### 前端拉取更新

```bash
# 1. 从共享仓库拉取最新 contract
git subtree pull --prefix=contract contracts main --squash

# 2. 推送前端仓库
git push origin main
```

### 根项目更新子模块

```bash
# 在根项目目录下
git add modelcraft-backend modelcraft-front
git commit -m "chore: update submodules (sync api contracts)"
```

---

## 关键规则

1. **后端 `api/` 是唯一真相源** — 所有 API Contract 变更只能从后端发起
2. **前端禁止直接修改 `contract/`** — 所有变更必须通过 subtree pull 获取
3. **先 push 再 pull** — 后端必须先 `subtree push`，前端才能 `subtree pull`
4. **Squash 策略** — 后端 push 不用 squash（保留历史），前端 pull 用 squash（保持前端历史整洁）
5. **Squash 一致性** — 一旦开始某种 squash 策略，不能中途切换

## GraphQL Schema 组织

后端有**两套独立的 GraphQL Schema**，分别服务在不同 endpoint：

| Schema | 目录 | 服务 URL | 包含内容 |
|--------|------|----------|----------|
| Org GraphQL | `api/graph/org/schema/` | `/graphql/org/{orgName}/` | 项目/集群/用户/角色管理 |
| Project GraphQL | `api/graph/project/schema/` | `/graphql/org/{orgName}/project/{projectSlug}/` | 模型/字段/枚举/外键/分组 |

前端调用时**必须使用正确的 endpoint**：
- Org 相关操作（项目/集群 CRUD）→ Org endpoint
- Project 相关操作（模型/字段/枚举 CRUD）→ Project endpoint

---

## 后端专属文件说明

共享仓库中包含以下后端专属文件，前端**不需要使用**但会随 subtree 一起拉取：

| 文件 | 说明 |
|------|------|
| `openapi/oapi-codegen.yaml` | Go 代码生成器配置 |
| `openapi/openapi.yaml` | 合并后的完整 OpenAPI spec（生成产物） |
| `openapi/examples/` | 请求/响应示例数据 |
| `openapi/README.md` | OpenAPI 模块说明 |

## 首次设置（已完成）

如需在新环境中重建 subtree，执行以下步骤：

### 后端

```bash
git remote add contracts https://git.woa.com/lukemxjia/modelcraft-api-contracts.git
git subtree push --prefix=api contracts main
```

### 前端

```bash
git remote add contracts https://git.woa.com/lukemxjia/modelcraft-api-contracts.git
git rm -r contract/
git commit -m "chore: remove manually-synced contract directory"
git subtree add --prefix=contract contracts main --squash
```

## 参考索引

| 主题 | 文件 |
|------|------|
| 后端 API 目录 | `modelcraft-backend/api/` |
| 前端 Contract 目录 | `modelcraft-front/contract/` |
| 共享仓库 | `git.woa.com/lukemxjia/modelcraft-api-contracts.git` |
| 后端 GraphQL 路由 | `internal/interfaces/http/routes.go` |
| gqlgen Org 配置 | `gqlgen.org.yml` |
| gqlgen Project 配置 | `gqlgen.project.yml` |
