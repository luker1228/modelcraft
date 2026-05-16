# APISIX Gateway Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 用 APISIX 完全替代自维护的 `modelcraft-gateway`，将 Cookie 管理移入后端，JWT payload 加 `key` 字段以适配 APISIX `jwt-auth` 插件。

**Architecture:** APISIX 接收所有外部流量，使用 `jwt-auth` 插件验签（ES256），通过 `serverless-pre-function` Lua 脚本从 claims 提取 `user_id` 注入 `X-User-ID`，`Authorization` 原样透传。后端的 `auth` handler 在响应中自己管理 `Set-Cookie`，不再依赖 Gateway 中间层。

**Tech Stack:** APISIX 3.9 (standalone 模式), etcd 3.5, Go 1.21 (后端改动), Docker Compose

---

## 文件变更地图

### 新建文件
- `apisix/config.yaml` — APISIX 启动配置（etcd、插件列表、OTel）
- `apisix/apisix.yaml` — 声明式路由 + 插件（standalone 模式）

### 修改文件
- `modelcraft-backend/internal/domain/auth/platform_claims.go` — 新增 `Key` 字段
- `modelcraft-backend/internal/domain/auth/jwt_signer.go` — `IssueAccessToken` 传入 `key`
- `modelcraft-backend/pkg/config/config.go` — 新增 `CookieConfig`
- `modelcraft-backend/configs/config.yaml` — 新增 cookie 配置节
- `modelcraft-backend/internal/interfaces/http/handlers/auth/handler.go` — Login/Refresh/Logout 写 Set-Cookie
- `modelcraft-backend/internal/interfaces/http/handlers/enduser/auth_handler.go` — EndUserLogin/Refresh/Logout 写 Set-Cookie
- `docker-compose.yml` — 新增 etcd + apisix 服务，移除 gateway 服务

---

## Task 1：后端 — PlatformClaims 加 `key` 字段

**Files:**
- Modify: `modelcraft-backend/internal/domain/auth/platform_claims.go`
- Modify: `modelcraft-backend/internal/domain/auth/platform_claims_test.go`
- Modify: `modelcraft-backend/internal/domain/auth/jwt_signer.go`

**背景：** APISIX `jwt-auth` 插件匹配 Consumer 时依赖 JWT payload 中的 `key` 字段。`PlatformClaims` 目前无此字段，需要加上并在签发时填入固定值 `"mc-user"`。

- [ ] **Step 1: 在 `platform_claims.go` 增加 `Key` 字段**

  ```go
  // platform_claims.go — 在 PlatformClaims struct 中加一行
  type PlatformClaims struct {
      UserID          string            `json:"user_id"`
      OrgName         string            `json:"org_name"`
      Key             string            `json:"key"`                          // APISIX jwt-auth Consumer key
      EndUserAdminIDs map[string]string `json:"end_user_admin_ids,omitempty"`
      jwt.RegisteredClaims
  }
  ```

- [ ] **Step 2: 在 `jwt_signer.go` 的 `IssueAccessToken` 中填入 `Key`**

  在 `claims := &PlatformClaims{...}` 块中加一行 `Key: "mc-user",`：

  ```go
  claims := &PlatformClaims{
      UserID:          userID,
      OrgName:         orgName,
      Key:             "mc-user",   // ← 新增
      EndUserAdminIDs: endUserAdminIDs,
      RegisteredClaims: jwt.RegisteredClaims{
          Issuer:    string(IssuerPlatform),
          Subject:   userID,
          Audience:  aud,
          IssuedAt:  jwt.NewNumericDate(now),
          ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
      },
  }
  ```

- [ ] **Step 3: 更新 `platform_claims_test.go`，验证 `key` 字段在签发/解析后正确**

  在现有 `TestPlatformClaims_Validate` 中找一个成功用例，补充断言：

  ```go
  // 在某个 c := &PlatformClaims{...} 的成功测试 case 加上：
  assert.Equal(t, "mc-user", claims.Key)
  ```

  并在 `jwt_signer_test.go` 的 issue+parse 往返测试中加：

  ```go
  assert.Equal(t, "mc-user", claims.Key)
  ```

- [ ] **Step 4: 运行 domain/auth 包测试，确保全部通过**

  ```bash
  cd modelcraft-backend && go test ./internal/domain/auth/... -v
  ```

  期望：所有测试 PASS，无编译错误。

- [ ] **Step 5: Commit**

  ```bash
  git add modelcraft-backend/internal/domain/auth/platform_claims.go \
          modelcraft-backend/internal/domain/auth/jwt_signer.go \
          modelcraft-backend/internal/domain/auth/platform_claims_test.go \
          modelcraft-backend/internal/domain/auth/jwt_signer_test.go
  git commit -m "feat(auth): add key field to PlatformClaims for APISIX jwt-auth"
  ```

---

## Task 2：后端 — Cookie 配置项

**Files:**
- Modify: `modelcraft-backend/pkg/config/config.go`
- Modify: `modelcraft-backend/configs/config.yaml`

**背景：** 目前 Cookie 属性（domain、secure、SameSite）硬编码在 Gateway。移入后端后需要通过配置控制，以区分本地开发（`secure=false`）和生产（`secure=true`）。

- [ ] **Step 1: 在 `config.go` 新增 `CookieConfig` 并挂到 `AuthConfig`**

  在 `// AuthConfig 认证配置` 的 `AuthConfig` struct 前插入：

  ```go
  // CookieConfig controls Set-Cookie attributes for refresh tokens.
  type CookieConfig struct {
      Domain   string `mapstructure:"domain"`   // Cookie domain (empty = use request host)
      Secure   bool   `mapstructure:"secure"`   // Secure flag; set true in production (HTTPS)
      SameSite string `mapstructure:"same_site"` // "strict" | "lax" | "none"
  }
  ```

  然后在 `AuthConfig` 中加一行：

  ```go
  type AuthConfig struct {
      InternalToken string            `mapstructure:"internal_token"`
      Cookie        CookieConfig      `mapstructure:"cookie"`          // ← 新增
      Design        DesignAuthConfig  `mapstructure:"design"`
      Runtime       RuntimeAuthConfig `mapstructure:"runtime"`
  }
  ```

- [ ] **Step 2: 在 `configs/config.yaml` 新增 `cookie` 配置节**

  在 `auth:` 配置节下追加（本地开发默认 secure=false）：

  ```yaml
  auth:
    # ... 已有配置保持不变 ...
    cookie:
      domain: ""          # 留空则使用请求的 host
      secure: false       # 生产环境通过环境变量 AUTH_COOKIE_SECURE=true 覆盖
      same_site: "strict"
  ```

- [ ] **Step 3: 在 `config.go` 绑定环境变量（在 LoadConfigWithOptions 中已有 BindEnv 区域追加）**

  找到文件中现有的 `v.BindEnv` 调用集中位置，追加：

  ```go
  _ = v.BindEnv("auth.cookie.secure", "AUTH_COOKIE_SECURE")
  _ = v.BindEnv("auth.cookie.domain", "AUTH_COOKIE_DOMAIN")
  _ = v.BindEnv("auth.cookie.same_site", "AUTH_COOKIE_SAME_SITE")
  ```

- [ ] **Step 4: 编译后端，确保无报错**

  ```bash
  cd modelcraft-backend && go build ./...
  ```

  期望：编译成功，无错误。

- [ ] **Step 5: Commit**

  ```bash
  git add modelcraft-backend/pkg/config/config.go \
          modelcraft-backend/configs/config.yaml
  git commit -m "feat(config): add CookieConfig for refresh token Set-Cookie attributes"
  ```

---

## Task 3：后端 — Tenant auth handler 写 Set-Cookie

**Files:**
- Modify: `modelcraft-backend/internal/interfaces/http/handlers/auth/handler.go`
- Modify: `modelcraft-backend/internal/interfaces/http/server.go`（或 routes.go，确认 handler 构造点）

**背景：** 当前 `HandleLogin` 把 `refreshToken` 放在响应 JSON 里，由 Gateway 提取写 Cookie。Gateway 去掉后，后端自己在响应时 `Set-Cookie`，并从 JSON 响应中删除 `refreshToken`。`HandleRefresh` 从请求 Cookie 读取 refresh token，`HandleLogout` 清除 Cookie。

- [ ] **Step 1: 确认 `Handler` 可以访问 `CookieConfig`，修改构造函数**

  在 `handler.go` 顶部找到 `Handler` struct 和 `NewHandler`，加入 `cookieCfg`：

  ```go
  type Handler struct {
      tokenService *appAuth.TokenService
      cookieCfg    config.CookieConfig   // ← 新增
      logger       logfacade.Logger
  }

  func NewHandler(tokenService *appAuth.TokenService, cookieCfg config.CookieConfig, logger logfacade.Logger) *Handler {
      return &Handler{
          tokenService: tokenService,
          cookieCfg:    cookieCfg,
          logger:       logger,
      }
  }
  ```

  在 import 中加 `"modelcraft/pkg/config"`（如未导入）。

- [ ] **Step 2: 添加 `setRefreshCookie` 和 `clearRefreshCookie` 辅助方法**

  ```go
  const tenantRefreshCookieName = "mc_refresh_token"

  func (h *Handler) setRefreshCookie(w http.ResponseWriter, token string) {
      sameSite := http.SameSiteStrictMode
      switch h.cookieCfg.SameSite {
      case "lax":
          sameSite = http.SameSiteLaxMode
      case "none":
          sameSite = http.SameSiteNoneMode
      }
      http.SetCookie(w, &http.Cookie{
          Name:     tenantRefreshCookieName,
          Value:    token,
          Path:     "/",
          Domain:   h.cookieCfg.Domain,
          HttpOnly: true,
          Secure:   h.cookieCfg.Secure,
          SameSite: sameSite,
          MaxAge:   30 * 24 * 60 * 60, // 30 days
      })
  }

  func (h *Handler) clearRefreshCookie(w http.ResponseWriter) {
      http.SetCookie(w, &http.Cookie{
          Name:     tenantRefreshCookieName,
          Value:    "",
          Path:     "/",
          Domain:   h.cookieCfg.Domain,
          HttpOnly: true,
          Secure:   h.cookieCfg.Secure,
          MaxAge:   -1,
      })
  }
  ```

- [ ] **Step 3: 修改 `HandleLogin`：写 Cookie，删除 JSON 中的 `refreshToken`**

  找到 `HandleLogin` 末尾的 `writeJSON(w, http.StatusOK, map[string]any{...})` 调用，改为：

  ```go
  h.setRefreshCookie(w, result.RefreshToken)

  writeJSON(w, http.StatusOK, map[string]any{
      "requestId":   requestID,
      "userId":      result.UserID,
      "userName":    userName,
      "orgName":     orgName,
      "accessToken": result.AccessToken,
      "expiresIn":   result.ExpiresIn,
      // refreshToken intentionally omitted — stored in httpOnly cookie
  })
  ```

- [ ] **Step 4: 修改 `HandleRefresh`：从 Cookie 读取 refresh token**

  将当前 `HandleRefresh` 中 `var req struct { RefreshToken string ... }; json.NewDecoder(r.Body).Decode(&req)` 改为从 Cookie 读取：

  ```go
  func (h *Handler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
      requestID := ctxutils.GetRequestID(r.Context())

      cookie, err := r.Cookie(tenantRefreshCookieName)
      if err != nil || cookie.Value == "" {
          writeAuthError(w, http.StatusUnauthorized, requestID, "REFRESH_MISSING", "refresh token not found")
          return
      }

      result, err := h.tokenService.Refresh(r.Context(), appAuth.RefreshCommand{
          RefreshToken: cookie.Value,
      })
      if err != nil {
          h.clearRefreshCookie(w)
          h.handleBusinessError(w, r, requestID, err, "Refresh failed")
          return
      }

      h.setRefreshCookie(w, result.RefreshToken)
      writeJSON(w, http.StatusOK, map[string]any{
          "requestId":   requestID,
          "accessToken": result.AccessToken,
          "expiresIn":   result.ExpiresIn,
      })
  }
  ```

- [ ] **Step 5: 修改 `HandleLogout`：从 Cookie 读 refresh token，清除 Cookie**

  ```go
  func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
      cookie, _ := r.Cookie(tenantRefreshCookieName)
      if cookie != nil && cookie.Value != "" {
          _ = h.tokenService.Logout(r.Context(), appAuth.LogoutCommand{
              RefreshToken: cookie.Value,
          })
      }
      h.clearRefreshCookie(w)
      w.WriteHeader(http.StatusNoContent)
  }
  ```

- [ ] **Step 6: 找到 `NewHandler` 调用点，传入 `CookieConfig`**

  在 `modelcraft-backend/internal/interfaces/http/routes.go`（或 server.go）找到：
  ```go
  authHandler := auth.NewHandler(tokenService, logger)
  ```
  改为：
  ```go
  authHandler := auth.NewHandler(tokenService, cfg.Auth.Cookie, logger)
  ```

- [ ] **Step 7: 编译后端**

  ```bash
  cd modelcraft-backend && go build ./...
  ```

  期望：编译成功。

- [ ] **Step 8: Commit**

  ```bash
  git add modelcraft-backend/internal/interfaces/http/handlers/auth/handler.go \
          modelcraft-backend/internal/interfaces/http/routes.go
  git commit -m "feat(auth): move tenant refresh cookie management to backend"
  ```

---

## Task 4：后端 — EndUser auth handler 写 Set-Cookie

**Files:**
- Modify: `modelcraft-backend/internal/interfaces/http/handlers/enduser/auth_handler.go`

**背景：** 与 Task 3 类似，但针对 EndUser 认证流程。EndUser 的 refresh cookie 名为 `mc_enduser_refresh_token`。`EndUserSelectProject` 也需要更新（它会签发新的 token）。

- [ ] **Step 1: 在 `auth_handler.go` 中给 `AuthHandler` 加 `cookieCfg` 并更新构造函数**

  ```go
  type AuthHandler struct {
      authService *appEnduser.EndUserAuthAppService
      jwtSigner   *domainAuth.JWTSigner
      cookieCfg   config.CookieConfig   // ← 新增
      logger      logfacade.Logger
  }

  func NewAuthHandler(
      authService *appEnduser.EndUserAuthAppService,
      jwtSigner *domainAuth.JWTSigner,
      cookieCfg config.CookieConfig,   // ← 新增
      logger logfacade.Logger,
  ) *AuthHandler {
      return &AuthHandler{
          authService: authService,
          jwtSigner:   jwtSigner,
          cookieCfg:   cookieCfg,
          logger:      logger,
      }
  }
  ```

- [ ] **Step 2: 添加 `setEndUserRefreshCookie` 和 `clearEndUserRefreshCookie` 辅助方法**

  ```go
  const endUserRefreshCookieName = "mc_enduser_refresh_token"

  func (h *AuthHandler) setEndUserRefreshCookie(w http.ResponseWriter, token string) {
      sameSite := http.SameSiteStrictMode
      switch h.cookieCfg.SameSite {
      case "lax":
          sameSite = http.SameSiteLaxMode
      case "none":
          sameSite = http.SameSiteNoneMode
      }
      http.SetCookie(w, &http.Cookie{
          Name:     endUserRefreshCookieName,
          Value:    token,
          Path:     "/",
          Domain:   h.cookieCfg.Domain,
          HttpOnly: true,
          Secure:   h.cookieCfg.Secure,
          SameSite: sameSite,
          MaxAge:   30 * 24 * 60 * 60,
      })
  }

  func (h *AuthHandler) clearEndUserRefreshCookie(w http.ResponseWriter) {
      http.SetCookie(w, &http.Cookie{
          Name:     endUserRefreshCookieName,
          Value:    "",
          Path:     "/",
          Domain:   h.cookieCfg.Domain,
          HttpOnly: true,
          Secure:   h.cookieCfg.Secure,
          MaxAge:   -1,
      })
  }
  ```

- [ ] **Step 3: 修改 `EndUserLogin`：写 Cookie，从响应 JSON 删除 `refreshToken`**

  找到 `h.writeJSON(w, http.StatusOK, buildTokenResponse(..., result.RefreshToken, ...))` 这行，改为：

  ```go
  h.setEndUserRefreshCookie(w, result.RefreshToken)
  h.writeJSON(w, http.StatusOK, buildTokenResponse(requestID, result.UserID,
      result.AccessToken, "" /* no refresh in body */, result.ExpiresAt,
      toProjectList(result.Projects), ""))
  ```

- [ ] **Step 4: 修改 `EndUserRegister`：同上，写 Cookie，清空 JSON 中的 refreshToken**

  与 Step 3 相同的改法，找到 `buildTokenResponse(..., result.RefreshToken, ...)` 那行替换。

- [ ] **Step 5: 修改 `EndUserRefreshToken`：从 Cookie 读取 refresh token**

  ```go
  func (h *AuthHandler) EndUserRefreshToken(w http.ResponseWriter, r *http.Request) {
      ctx := r.Context()
      requestID := ctxutils.GetRequestID(ctx)

      cookie, err := r.Cookie(endUserRefreshCookieName)
      if err != nil || cookie.Value == "" {
          h.writeError(w, http.StatusUnauthorized, requestID, "REFRESH_MISSING", "refresh token not found")
          return
      }

      // 需要 orgName + projectSlug（来自 request body）
      var req struct {
          OrgName     string `json:"orgName"`
          ProjectSlug string `json:"projectSlug"`
      }
      if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
          h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "invalid request body")
          return
      }

      result, err := h.authService.RefreshEndUserToken(ctx, appEnduser.RefreshCommand{
          RefreshToken: cookie.Value,
          OrgName:      req.OrgName,
          ProjectSlug:  req.ProjectSlug,
      })
      if err != nil {
          h.clearEndUserRefreshCookie(w)
          h.handleBizError(w, r, requestID, err, "end-user refresh failed")
          return
      }

      h.setEndUserRefreshCookie(w, result.RefreshToken)
      h.writeJSON(w, http.StatusOK, buildTokenResponse(requestID, result.UserID,
          result.AccessToken, "" /* no refresh in body */, result.ExpiresAt,
          toProjectList(result.Projects), ""))
  }
  ```

- [ ] **Step 6: 修改 `EndUserLogout`：从 Cookie 读 refresh token，清除 Cookie**

  ```go
  func (h *AuthHandler) EndUserLogout(w http.ResponseWriter, r *http.Request) {
      ctx := r.Context()
      requestID := ctxutils.GetRequestID(ctx)

      cookie, _ := r.Cookie(endUserRefreshCookieName)
      if cookie != nil && cookie.Value != "" {
          var req struct {
              OrgName string `json:"orgName"`
          }
          _ = json.NewDecoder(r.Body).Decode(&req)
          _ = h.authService.LogoutEndUser(ctx, appEnduser.LogoutCommand{
              RefreshToken: cookie.Value,
              OrgName:      req.OrgName,
          })
      }

      h.clearEndUserRefreshCookie(w)
      w.WriteHeader(http.StatusNoContent)
  }
  ```

- [ ] **Step 7: 修改 `EndUserSelectProject`：从 Cookie 读 refresh token，写新 Cookie**

  找到当前 `EndUserSelectProject` 方法中从 request body 读 `refreshToken` 的逻辑，改为从 Cookie 读取，并在成功响应后写新 Cookie（如果返回了新 refresh token）：

  ```go
  func (h *AuthHandler) EndUserSelectProject(w http.ResponseWriter, r *http.Request) {
      ctx := r.Context()
      requestID := ctxutils.GetRequestID(ctx)

      cookie, err := r.Cookie(endUserRefreshCookieName)
      if err != nil || cookie.Value == "" {
          h.writeError(w, http.StatusUnauthorized, requestID, "REFRESH_MISSING", "refresh token not found")
          return
      }

      var req struct {
          OrgName     string `json:"orgName"`
          ProjectSlug string `json:"projectSlug"`
      }
      if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
          h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "invalid request body")
          return
      }

      result, err := h.authService.SelectProject(ctx, appEnduser.SelectProjectCommand{
          RefreshToken: cookie.Value,
          OrgName:      req.OrgName,
          ProjectSlug:  req.ProjectSlug,
      })
      if err != nil {
          h.handleBizError(w, r, requestID, err, "end-user select-project failed")
          return
      }

      if result.RefreshToken != "" {
          h.setEndUserRefreshCookie(w, result.RefreshToken)
      }
      h.writeJSON(w, http.StatusOK, buildTokenResponse(requestID, result.UserID,
          result.AccessToken, "" /* no refresh in body */, result.ExpiresAt,
          toProjectList(result.Projects), result.ProjectSlug))
  }
  ```

- [ ] **Step 8: 更新 `NewAuthHandler` 调用点（在 routes.go）**

  找到：
  ```go
  endUserAuthHandler := enduserHandlers.NewAuthHandler(endUserAuthAppService, jwtSigner, logger)
  ```
  改为：
  ```go
  endUserAuthHandler := enduserHandlers.NewAuthHandler(endUserAuthAppService, jwtSigner, cfg.Auth.Cookie, logger)
  ```

- [ ] **Step 9: 编译后端**

  ```bash
  cd modelcraft-backend && go build ./...
  ```

  期望：编译成功。

- [ ] **Step 10: Commit**

  ```bash
  git add modelcraft-backend/internal/interfaces/http/handlers/enduser/auth_handler.go \
          modelcraft-backend/internal/interfaces/http/routes.go
  git commit -m "feat(enduser): move end-user refresh cookie management to backend"
  ```

---

## Task 5：APISIX 配置文件

**Files:**
- Create: `apisix/config.yaml`
- Create: `apisix/apisix.yaml`

**背景：** APISIX standalone 模式从 `apisix.yaml` 直接加载路由，不依赖 Admin API。`config.yaml` 配置 etcd 地址、启用的插件列表、OTel。

- [ ] **Step 1: 创建 `apisix/config.yaml`**

  ```yaml
  # apisix/config.yaml
  apisix:
    node_listen: 9080
    enable_ipv6: false

  etcd:
    host:
      - "http://etcd:2379"
    prefix: /apisix
    timeout: 30

  deployment:
    role: traditional
    role_traditional:
      config_provider: etcd

  plugin_attr:
    opentelemetry:
      resource:
        service.name: modelcraft-apisix
      collector:
        address: "${OTEL_EXPORTER_OTLP_ENDPOINT:-}"  # 空则禁用导出

  plugins:
    - cors
    - request-id
    - opentelemetry
    - jwt-auth
    - serverless-pre-function
    - proxy-rewrite
    - response-rewrite
  ```

- [ ] **Step 2: 创建 `apisix/apisix.yaml`（standalone 模式路由配置）**

  ```yaml
  # apisix/apisix.yaml
  # Standalone 模式：APISIX 直接从此文件加载，不走 etcd / Admin API
  # 修改后需重启 APISIX 容器生效

  consumers:
    - username: mc-user
      plugins:
        jwt-auth:
          key: mc-user
          algorithm: ES256
          public_key: "${JWT_PUBLIC_KEY}"   # PEM 公钥，运行时由环境变量注入

  routes:
    # ── 公开路由：auth 端点，不鉴权 ──────────────────────────────────────────
    - id: tenant-auth
      uri: /api/tenant/auth/*
      upstream_id: backend
      plugins:
        cors:
          allow_origins: "${FRONTEND_URL:-http://localhost:3000}"
          allow_methods: "GET,POST,PUT,DELETE,OPTIONS"
          allow_headers: "Authorization,Content-Type,X-Request-Id,X-Client-Request-Id"
          allow_credential: true
        request-id:
          header_name: X-Request-Id
          include_in_response: true

    - id: enduser-auth
      uri: /api/end-user/auth/*
      upstream_id: backend
      plugins:
        cors:
          allow_origins: "${FRONTEND_URL:-http://localhost:3000}"
          allow_methods: "GET,POST,PUT,DELETE,OPTIONS"
          allow_headers: "Authorization,Content-Type,X-Request-Id,X-Client-Request-Id"
          allow_credential: true
        request-id:
          header_name: X-Request-Id
          include_in_response: true

    - id: cli-enduser-auth
      uri: /api/cli/end-user/auth/*
      upstream_id: backend
      plugins:
        cors:
          allow_origins: "*"
          allow_methods: "GET,POST,OPTIONS"
          allow_headers: "Authorization,Content-Type,X-Request-Id"
        request-id:
          header_name: X-Request-Id
          include_in_response: true

    # ── 受保护路由：GraphQL + REST ────────────────────────────────────────────
    - id: graphql-all
      uri: /graphql/*
      upstream_id: backend
      plugins:
        cors:
          allow_origins: "${FRONTEND_URL:-http://localhost:3000}"
          allow_methods: "POST,OPTIONS"
          allow_headers: "Authorization,Content-Type,X-Request-Id,X-Client-Request-Id"
          allow_credential: true
        request-id:
          header_name: X-Request-Id
          include_in_response: true
        jwt-auth:
          hide_credentials: false   # 保留 Authorization header，后端继续接收
        serverless-pre-function:
          phase: rewrite
          functions:
            - |
              return function(conf, ctx)
                local core = require("apisix.core")
                local jwt = require("resty.jwt")

                local auth_header = core.request.header(ctx, "Authorization")
                if not auth_header then return end

                local token = auth_header:match("Bearer%s+(.+)")
                if not token then return end

                -- JWT 已经被 jwt-auth 插件验证，直接 decode payload
                local jwt_obj = jwt:load_jwt(token)
                if jwt_obj and jwt_obj.payload and jwt_obj.payload.user_id then
                  core.request.set_header(ctx, "X-User-ID", jwt_obj.payload.user_id)
                end
              end

    - id: api-user
      uri: /api/user/*
      upstream_id: backend
      plugins:
        cors:
          allow_origins: "${FRONTEND_URL:-http://localhost:3000}"
          allow_methods: "GET,POST,PUT,PATCH,DELETE,OPTIONS"
          allow_headers: "Authorization,Content-Type,X-Request-Id,X-Client-Request-Id"
          allow_credential: true
        request-id:
          header_name: X-Request-Id
          include_in_response: true
        jwt-auth:
          hide_credentials: false
        serverless-pre-function:
          phase: rewrite
          functions:
            - |
              return function(conf, ctx)
                local core = require("apisix.core")
                local jwt = require("resty.jwt")
                local auth_header = core.request.header(ctx, "Authorization")
                if not auth_header then return end
                local token = auth_header:match("Bearer%s+(.+)")
                if not token then return end
                local jwt_obj = jwt:load_jwt(token)
                if jwt_obj and jwt_obj.payload and jwt_obj.payload.user_id then
                  core.request.set_header(ctx, "X-User-ID", jwt_obj.payload.user_id)
                end
              end

  upstreams:
    - id: backend
      type: roundrobin
      nodes:
        "backend:8080": 1

  #END
  ```

  > **注意：** `apisix.yaml` 中 `${JWT_PUBLIC_KEY}` 是 APISIX 的环境变量插值语法，需要在 Docker Compose 中通过 `env` 传入，APISIX 启动时自动替换。

- [ ] **Step 3: Commit**

  ```bash
  git add apisix/config.yaml apisix/apisix.yaml
  git commit -m "feat(apisix): add APISIX standalone config and routes"
  ```

---

## Task 6：更新 docker-compose.yml

**Files:**
- Modify: `docker-compose.yml`
- Modify: `.env.example`

**背景：** 在根 `docker-compose.yml` 中新增 `etcd` + `apisix` 服务，修改 `gateway` 服务为注释掉（保留便于回退），更新 `frontend` 和 `modelcraft-agent` 的依赖指向 apisix。

- [ ] **Step 1: 在 `docker-compose.yml` 新增 etcd 和 apisix 服务**

  在 `services:` 下，`backend:` 之前插入：

  ```yaml
    # ---------------------------------------------------------------------------
    # etcd（APISIX 配置存储）
    # ---------------------------------------------------------------------------
    etcd:
      image: bitnami/etcd:3.5
      container_name: modelcraft-etcd
      environment:
        ALLOW_NONE_AUTHENTICATION: "yes"
      volumes:
        - etcd_data:/bitnami/etcd
      networks:
        - modelcraft-network
      restart: unless-stopped
      healthcheck:
        test: ["CMD", "etcdctl", "endpoint", "health"]
        interval: 10s
        timeout: 5s
        retries: 5

    # ---------------------------------------------------------------------------
    # APISIX（API Gateway，替代 modelcraft-gateway）
    # ---------------------------------------------------------------------------
    apisix:
      image: apache/apisix:3.9.0-debian
      container_name: modelcraft-apisix
      ports:
        - "${GATEWAY_PORT:-9080}:9080"
        - "9180:9180"   # Admin API，仅本地调试用，生产环境不暴露
      environment:
        - JWT_PUBLIC_KEY=${JWT_PUBLIC_KEY}
        - FRONTEND_URL=${FRONTEND_URL:-http://localhost:3000}
        - OTEL_EXPORTER_OTLP_ENDPOINT=${OTEL_EXPORTER_OTLP_ENDPOINT:-}
      volumes:
        - ./apisix/config.yaml:/usr/local/apisix/conf/config.yaml:ro
        - ./apisix/apisix.yaml:/usr/local/apisix/conf/apisix.yaml:ro
      depends_on:
        etcd:
          condition: service_healthy
        backend:
          condition: service_healthy
      networks:
        - modelcraft-network
      restart: unless-stopped
      healthcheck:
        test: ["CMD", "curl", "-f", "http://localhost:9080/healthz"]
        interval: 15s
        timeout: 5s
        retries: 3
        start_period: 20s
  ```

- [ ] **Step 2: 注释掉原 `gateway` 服务（保留以便回退）**

  将 `gateway:` 服务块整体注释，头部加说明：

  ```yaml
    # ---------------------------------------------------------------------------
    # modelcraft-gateway（已被 APISIX 替代，保留用于紧急回退）
    # ---------------------------------------------------------------------------
    # gateway:
    #   build: ...
    #   （原内容保留）
  ```

- [ ] **Step 3: 更新 `frontend` 的 depends_on 和构建参数**

  将 `frontend` 中 `depends_on: gateway` 改为 `depends_on: apisix`，并将 `BACKEND_URL` 指向 apisix：

  ```yaml
    frontend:
      build:
        args:
          - BACKEND_URL=http://apisix:9080   # ← 改这里
      depends_on:
        apisix:                              # ← 改这里
          condition: service_healthy
  ```

- [ ] **Step 4: 更新 `modelcraft-agent` 的 depends_on 和 GATEWAY_URL**

  ```yaml
    modelcraft-agent:
      environment:
        - GATEWAY_URL=http://apisix:9080   # ← 改这里
      depends_on:
        apisix:                            # ← 改这里
          condition: service_healthy
  ```

- [ ] **Step 5: 在 `volumes:` 新增 `etcd_data`**

  ```yaml
  volumes:
    mysql_data:
    redis_data:
    backend_logs:
    etcd_data:     # ← 新增
  ```

- [ ] **Step 6: 更新 `.env.example`，新增 APISIX 相关变量说明**

  在 `.env.example` 适当位置加：

  ```bash
  # APISIX Gateway（替代 modelcraft-gateway）
  GATEWAY_PORT=9080
  # JWT_PUBLIC_KEY 同后端共用，已在上方定义
  # OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317
  
  # 后端 Cookie 配置（APISIX 迁移后生效）
  AUTH_COOKIE_SECURE=false    # 生产环境改为 true
  AUTH_COOKIE_DOMAIN=         # 留空则使用请求 host
  AUTH_COOKIE_SAME_SITE=strict
  ```

- [ ] **Step 7: Commit**

  ```bash
  git add docker-compose.yml .env.example
  git commit -m "feat(docker): replace gateway with APISIX + etcd in docker-compose"
  ```

---

## Task 7：本地联调验证

**背景：** 全部配置完成后，启动完整栈验证核心流程。

- [ ] **Step 1: 确保后端 `.env` 中有 `JWT_PUBLIC_KEY` 和 `JWT_PRIVATE_KEY`**

  ```bash
  # 如果没有，用 openssl 生成一对 EC 密钥（P-256）：
  openssl ecparam -name prime256v1 -genkey -noout -out ec-private.pem
  openssl ec -in ec-private.pem -pubout -out ec-public.pem
  # 将 ec-private.pem 内容（单行）写入 .env 的 JWT_PRIVATE_KEY
  # 将 ec-public.pem 内容（单行）写入 .env 的 JWT_PUBLIC_KEY
  # PEM 内容中换行符替换为 \n
  ```

- [ ] **Step 2: 启动完整栈**

  ```bash
  docker compose up -d
  docker compose ps
  ```

  期望：`modelcraft-backend`、`modelcraft-etcd`、`modelcraft-apisix`、`modelcraft-mysql`、`modelcraft-redis` 均 `healthy`。

- [ ] **Step 3: 验证 APISIX 路由可达**

  ```bash
  curl -s http://localhost:9080/healthz
  ```

  期望：返回 `200 OK`。（如 APISIX 没有 /healthz 内置，返回 404 也可以，说明 APISIX 在响应）

- [ ] **Step 4: 验证 tenant 登录（Set-Cookie）**

  ```bash
  curl -s -c cookies.txt -X POST http://localhost:9080/api/tenant/auth/login \
    -H "Content-Type: application/json" \
    -d '{"identifier":"<your-user>","identifierType":"USERNAME","password":"<your-pass>"}' | jq .
  ```

  期望：
  - 响应 JSON 包含 `accessToken`，**不含** `refreshToken`
  - `cookies.txt` 中出现 `mc_refresh_token` cookie

- [ ] **Step 5: 验证 JWT 保护路由（X-User-ID 注入）**

  ```bash
  ACCESS_TOKEN="<从上一步拿到的 accessToken>"
  curl -s -X POST http://localhost:9080/graphql/org/<orgName> \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"query":"{ __typename }"}' | jq .
  ```

  期望：后端正常响应，不返回 401。

- [ ] **Step 6: 验证无 token 请求被拒**

  ```bash
  curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:9080/graphql/org/test \
    -H "Content-Type: application/json" \
    -d '{"query":"{ __typename }"}'
  ```

  期望：返回 `401`。

- [ ] **Step 7: 验证 tenant refresh（Cookie 轮换）**

  ```bash
  curl -s -b cookies.txt -c cookies.txt -X POST http://localhost:9080/api/tenant/auth/refresh \
    -H "Content-Type: application/json" | jq .
  ```

  期望：返回新 `accessToken`，`cookies.txt` 中 `mc_refresh_token` 更新。

- [ ] **Step 8: Commit 验证结果（可选）**

  ```bash
  git commit --allow-empty -m "chore: APISIX gateway migration verified locally"
  ```

---

## Task 8：下线 modelcraft-gateway

**Files:**
- Modify: `docker-compose.yml`（删除注释掉的 gateway 块）

**背景：** 验证通过后，清理遗留代码。`modelcraft-gateway/` 目录保留在 git 历史中，不需要物理删除（以防回退）。

- [ ] **Step 1: 从 `docker-compose.yml` 删除注释掉的 gateway 服务块**

  删除 Task 6 Step 2 中注释掉的 `# gateway:` 块。

- [ ] **Step 2: 更新 gateway 架构文档，标记为已迁移**

  在 `ai-metadata/backend/development/gateway-architecture.md` 顶部加：

  ```markdown
  > ⚠️ **已迁移（2026-05-16）**：`modelcraft-gateway` 已被 APISIX 替代。
  > 本文档保留作历史参考，当前架构见 `apisix/apisix.yaml`。
  ```

- [ ] **Step 3: Final commit**

  ```bash
  git add docker-compose.yml ai-metadata/backend/development/gateway-architecture.md
  git commit -m "chore: remove gateway service, mark architecture doc as migrated"
  ```
