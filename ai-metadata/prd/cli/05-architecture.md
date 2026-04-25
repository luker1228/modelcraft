# CLI 内部架构与后端变更

---

## 1. CLI 项目结构

```
modelcraft-cli/
├── cmd/                    # Cobra 命令定义
│   ├── root.go             # 根命令、全局标志、TTY 检测
│   ├── auth.go             # auth 子命令组
│   ├── query.go            # query 命令
│   ├── get.go              # get 命令
│   ├── create.go           # create 命令
│   ├── update.go           # update 命令
│   ├── delete.go           # delete 命令
│   ├── count.go            # count 命令
│   ├── aggregate.go        # aggregate 命令
│   ├── describe.go         # describe 命令
│   ├── catalog.go          # catalog 命令组
│   ├── schema.go           # schema 命令组（Agent 自省）
│   └── version.go          # version 命令
├── internal/
│   ├── client/             # HTTP/GraphQL 客户端
│   │   ├── auth.go         # Auth REST 客户端
│   │   ├── runtime.go      # Runtime GraphQL 客户端
│   │   └── catalog.go      # Catalog GraphQL 客户端
│   ├── config/             # 配置管理
│   │   └── credentials.go  # Token 存储/加载/刷新
│   ├── output/             # 输出格式化
│   │   ├── json.go         # JSON 输出（pretty / compact）
│   │   ├── yaml.go         # YAML 输出
│   │   └── error.go        # 统一错误格式化 + 退出码
│   └── resource/           # 资源路径解析
│       └── path.go         # "a.b.c" 路径解析器
├── main.go
├── go.mod
├── go.sum
└── Makefile
```

---

## 2. 核心模块说明

### 2.1 资源路径解析器 (`internal/resource/path.go`)

解析 `.` 分隔的资源路径，结合上下文补全缺失段：

```go
type ResourcePath struct {
    Project  string // 项目 slug
    Database string // 数据库名
    Model    string // 模型名
}

// Parse 解析资源路径字符串
// "a.b.c" → {Project:"a", Database:"b", Model:"c"}
// "b.c"   → {Database:"b", Model:"c"} (Project 从上下文补全)
// "c"     → {Model:"c"} (Project+Database 从上下文补全)
func Parse(path string, ctx *Config) (*ResourcePath, error)
```

### 2.2 Runtime GraphQL 客户端 (`internal/client/runtime.go`)

将 CLI 参数转换为 GraphQL 请求：

```go
type RuntimeClient struct {
    httpClient *http.Client
    serverURL  string
}

// Query 执行 findMany 查询
func (c *RuntimeClient) Query(ctx context.Context, path ResourcePath, opts QueryOptions) (*QueryResult, error)

// Get 执行 findUnique 查询
func (c *RuntimeClient) Get(ctx context.Context, path ResourcePath, opts GetOptions) (*GetResult, error)

// Create 执行 createOne 变更
func (c *RuntimeClient) Create(ctx context.Context, path ResourcePath, data map[string]interface{}) (*MutationResult, error)
```

GraphQL 请求发送到：
```
POST {serverURL}/graphql/org/{orgName}/project/{project}/db/{database}/model/{model}
Authorization: Bearer {accessToken}
```

### 2.3 输出格式化 (`internal/output/`)

所有命令通过统一的输出层返回结果：

```go
// Success 输出成功响应
func Success(data interface{}, meta *Meta) {
    result := map[string]interface{}{
        "ok":   true,
        "data": data,
    }
    if meta != nil {
        result["meta"] = meta
    }
    writeJSON(result)
}

// Error 输出错误响应并设置退出码
func Error(code string, message string, retryable bool, suggestion string, details map[string]interface{}) {
    result := map[string]interface{}{
        "ok": false,
        "error": map[string]interface{}{
            "code":       code,
            "message":    message,
            "retryable":  retryable,
            "suggestion": suggestion,
        },
    }
    if details != nil {
        result["error"].(map[string]interface{})["details"] = details
    }
    writeJSON(result)
    os.Exit(exitCodeFor(code))
}
```

### 2.4 Token 管理 (`internal/config/credentials.go`)

```go
type Credentials struct {
    Server       string `json:"server"`
    OrgName      string `json:"orgName"`
    ProjectSlug  string `json:"projectSlug"`
    UserID       string `json:"userId"`
    AccessToken  string `json:"accessToken"`
    RefreshToken string `json:"refreshToken"`
    ExpiresAt    string `json:"expiresAt"`
}

// Load 从 ~/.config/modelcraft/credentials.json 加载
func Load() (*Credentials, error)

// Save 保存到 ~/.config/modelcraft/credentials.json
func (c *Credentials) Save() error

// EnsureValidToken 检查 token 有效性，必要时自动刷新
func (c *Credentials) EnsureValidToken(authClient *AuthClient) error
```

---

## 3. 后端变更需求

### 3.1 新增公共 EndUser Auth REST 端点

**位置**: `modelcraft-backend/internal/interfaces/http/routes.go`

新增路由组，复用现有 `AuthHandler` 逻辑，但不使用 `internalTokenMW`：

```go
// SetupPublicEndUserAuthRoutes registers public End-User auth routes for CLI access.
func SetupPublicEndUserAuthRoutes(router chi.Router, handlers *DesignHandlers) {
    router.Route("/api/end-user/auth", func(r chi.Router) {
        r.Use(requestIDInjectorMiddleware)
        // 注意：不使用 internalTokenMW
        // 可选：增加 rate limiting
        r.Post("/login", handlers.EndUserAuthHandler.Login)
        r.Post("/select-project", handlers.EndUserAuthHandler.SelectProject)
        r.Post("/refresh", handlers.EndUserAuthHandler.Refresh)
        r.Post("/logout", handlers.EndUserAuthHandler.Logout)
        r.Get("/me", handlers.EndUserAuthHandler.Me)
    })
}
```

现有 handler 已支持从 `X-Org-Name` header 获取 org 信息，无需修改 handler 逻辑。

### 3.2 Runtime GraphQL — 无需变更

CLI 数据查询直接调用现有端点：

```
POST /graphql/org/{orgName}/project/{projectSlug}/db/{db}/model/{model}
```

`RuntimeAuthMiddleware` 已支持 `Authorization: Bearer <jwt>` 认证，验证 `iss: "mc-enduser"`。

### 3.3 EndUser GraphQL（Catalog）— 无需变更

`modelDatabaseCatalog` 和 `modelCatalog` 查询已存在于 EndUser GraphQL schema：
- 文件: `api/graph/end_user/schema/catalog.graphql`
- 端点: `/graphql/end-user/org/{orgName}/project/{projectSlug}`

### 3.4 Describe 数据来源

`mc describe` 所需的模型元数据（字段、类型、关系、限制）需要新增一个 API 或复用现有能力：

| 方案 | 说明 |
|------|------|
| **方案 A**: 扩展 Catalog GraphQL | 在 EndUser GraphQL 中新增 `modelSchema` 查询，返回字段详情 |
| **方案 B**: Runtime GraphQL Introspection | 利用 GraphQL 标准 `__schema` 自省（Runtime Schema 是动态生成的） |
| **方案 C**: 新增 REST 端点 | `GET /api/end-user/model-schema/{project}/{db}/{model}` |

推荐 **方案 B**：Runtime GraphQL 已动态生成完整 Schema（含 where/orderBy input types），CLI 可通过标准 GraphQL Introspection 获取字段信息，无需后端额外开发。Limit 信息可作为额外 REST 端点或 Catalog 扩展提供。

---

## 4. 构建与发布

### 4.1 构建

```makefile
# Makefile
VERSION := $(shell git describe --tags --always)

build:
    go build -ldflags "-X main.version=$(VERSION)" -o mc ./main.go

build-all:
    GOOS=linux   GOARCH=amd64 go build -o dist/mc-linux-amd64 ./main.go
    GOOS=darwin  GOARCH=amd64 go build -o dist/mc-darwin-amd64 ./main.go
    GOOS=darwin  GOARCH=arm64 go build -o dist/mc-darwin-arm64 ./main.go
    GOOS=windows GOARCH=amd64 go build -o dist/mc-windows-amd64.exe ./main.go
```

### 4.2 安装

```bash
# 从源码安装
go install github.com/modelcraft/cli@latest

# 或下载预编译二进制
curl -fsSL https://mc.example.com/install.sh | bash
```

---

## 5. 依赖关系

```
mc auth login
    ↓ 依赖
后端 /api/end-user/auth/* 公共端点（新增）
    ↓ 复用
现有 AuthHandler + EndUserAuthAppService

mc query / get / create / update / delete
    ↓ 依赖
现有 Runtime GraphQL 端点（无变更）
    ↓ 认证
现有 RuntimeAuthMiddleware（无变更）

mc catalog
    ↓ 依赖
现有 EndUser GraphQL Catalog 查询（无变更）

mc describe
    ↓ 依赖
Runtime GraphQL Introspection（内建）
    + 可选 Limit 信息端点（新增或扩展）
```
