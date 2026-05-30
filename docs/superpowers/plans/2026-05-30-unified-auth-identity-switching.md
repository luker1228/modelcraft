# 统一认证 & 管理/用户视图无缝切换 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 合并两套登录逻辑为同一后端实现，前端共享单一 auth store，`is_admin=true` 用户在管理页/用户页之间路由跳转即可切换身份，无需重新登录。

**Architecture:** 后端 `EndUserLogin` 直接复用 `TokenService.Login` 的核心逻辑（同一 `users` 表、同一 `JWTSigner.IssueAccessToken`），两条登录路径指向同一 handler。登录响应移除 `projects` 字段。前端建立单一 Zustand auth store，从 JWT payload 解析 `isAdmin`，路由 guard 按 `isAdmin` 放行。

**Tech Stack:** Go（后端），Next.js + Zustand（前端），jwt-decode（前端 JWT 解析）

**Spec:** `docs/superpowers/specs/2026-05-30-unified-auth-identity-switching-design.md`

---

## 文件变更清单

### 后端（`modelcraft-backend/`）

| 操作 | 文件 |
|------|------|
| Modify | `internal/interfaces/http/handlers/enduser/auth_handler.go` |
| Modify | `internal/app/enduser/commands.go`（移除 `LoginResult.Projects`、`RefreshResult.Projects`） |
| Modify | `internal/app/enduser/end_user_auth_service.go`（`LoginEndUser` 不再查 projects、`doRefreshInTx` 不再返回 projects） |
| Test | `internal/app/enduser/end_user_auth_service_test.go` |

### 前端（`modelcraft-front/`）

> 注意：前端文件路径需在实施时根据实际目录确认，以下为预期路径。

| 操作 | 文件 |
|------|------|
| Modify/Create | `src/store/authStore.ts`（单一 store，解析 `isAdmin`） |
| Modify | 管理路由 layout（`src/app/org/[orgName]/layout.tsx` 或类似） |
| Modify | 两个登录页的 `onLoginSuccess` 回调 |
| Modify/Create | 视图切换按钮组件（管理页 + 用户页导航） |

---

## Task 1：后端——移除 `LoginResult.Projects` 和 `RefreshResult.Projects`

**Files:**
- Modify: `modelcraft-backend/internal/app/enduser/commands.go`
- Modify: `modelcraft-backend/internal/app/enduser/end_user_auth_service.go`

- [ ] **Step 1: 修改 `commands.go`，移除 projects 字段**

编辑 `modelcraft-backend/internal/app/enduser/commands.go`，将 `LoginResult` 和 `RefreshResult` 中的 `Projects` 字段删除：

```go
// LoginResult represents the result of a successful login.
type LoginResult struct {
	UserID       string
	OrgName      string
	AccessToken  string
	// Projects 字段已移除：登录不再返回 project 列表
	RefreshToken string
	ExpiresAt    time.Time
}

// RefreshResult represents the result of a successful token refresh.
type RefreshResult struct {
	UserID       string
	AccessToken  string
	// Projects 字段已移除：刷新 token 不再返回 project 列表
	RefreshToken string
	ExpiresAt    time.Time
}
```

- [ ] **Step 2: 修改 `LoginEndUser`，移除 project 查询逻辑**

在 `modelcraft-backend/internal/app/enduser/end_user_auth_service.go` 中，找到 `LoginEndUser` 方法（约 162 行），将 project 查询和 `accessToken` 条件判断替换为直接签发 token：

```go
// LoginEndUser handles end-user login.
func (s *EndUserAuthAppService) LoginEndUser(ctx context.Context, cmd LoginCommand) (*LoginResult, error) {
	userRepo := s.repoFactory.NewEndUserRepository(s.db, cmd.OrgName, "")
	resolvedOrgName := cmd.OrgName

	identifier := cmd.Username
	idType := IdentifierTypeUsername
	if cmd.IdentifierType != "" {
		identifier = cmd.Identifier
		idType = cmd.IdentifierType
	} else if cmd.Identifier != "" {
		identifier = cmd.Identifier
	}

	var user *enduser.EndUser
	var err error
	switch idType {
	case IdentifierTypePhone:
		if resolvedOrgName == "" {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid, "orgName is required for phone login")
		}
		if identifier == "" {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid, "phone is required")
		}
		user, err = userRepo.GetByPhone(ctx, resolvedOrgName, identifier)
	case IdentifierTypeUsername:
		if resolvedOrgName != "" {
			user, err = userRepo.GetByUsername(ctx, resolvedOrgName, identifier)
		} else {
			user, err = userRepo.GetByUsernameGlobal(ctx, identifier)
		}
	default:
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid, "unsupported identifier type: "+string(idType))
	}
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if user == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserInvalidCredentials)
	}
	if resolvedOrgName == "" {
		resolvedOrgName = user.OrgName
	}
	if !user.VerifyPassword(cmd.Password) {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserInvalidCredentials)
	}
	if !user.IsActive() {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserAccountDisabled)
	}

	tokenResult, err := s.issueAccessToken(ctx, user.ID, resolvedOrgName, nil, user.IsAdmin)
	if err != nil {
		return nil, err
	}

	plaintext, tokenHash, err := generateRefreshToken()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate refresh token")
	}

	sessionID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate session id")
	}
	expiresAt := time.Now().Add(s.refreshTTL)
	token := &domainAuth.RefreshToken{
		ID:        sessionID,
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	refreshTokenRepo := s.repoFactory.NewRefreshTokenRepository(s.db)
	if err := refreshTokenRepo.Save(ctx, token); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	s.logger.Infof(ctx, "EndUser login: id=%s, identifier=%s, org=%s", user.ID, identifier, resolvedOrgName)

	return &LoginResult{
		UserID:       user.ID,
		OrgName:      resolvedOrgName,
		AccessToken:  tokenResult.AccessToken,
		RefreshToken: plaintext,
		ExpiresAt:    expiresAt,
	}, nil
}
```

- [ ] **Step 3: 修改 `issueAccessToken` 签名，移除 `projectSlugs` 参数（或传 nil）**

在同文件中，`issueAccessToken` 当前接收 `projectSlugs []string`。将调用改为传 `nil`（`EndUserTokenIssueInput.ProjectSlugs` 保留字段不删除，避免接口破坏）：

```go
// issueAccessToken issues an access token. projectSlugs is unused in current JWT but kept for future use.
func (s *EndUserAuthAppService) issueAccessToken(
	ctx context.Context,
	userID, orgName string,
	projectSlugs []string,
	isAdmin bool,
) (*EndUserTokenIssueResult, error) {
	// 函数体不变
```

实际上函数签名不变，只是调用处不再传 project slugs，传 `nil` 即可。

- [ ] **Step 4: 修改 `doRefreshInTx`，移除 project 查询，直接签发 token**

找到 `doRefreshInTx`（约 331 行），替换 `loadUserAndProjects` + `buildAccessToken` 调用：

```go
func (s *EndUserAuthAppService) doRefreshInTx(
	ctx context.Context,
	txDB SQLDBTX,
	tokenHash, orgName string,
) (*RefreshResult, error) {
	txRefreshTokenRepo := s.repoFactory.NewRefreshTokenRepository(txDB)

	token, err := txRefreshTokenRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if token == nil || !token.IsValid() {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserInvalidRefreshToken)
	}

	// 只查 user 本身，不再查 projects
	userRepo := s.repoFactory.NewEndUserRepository(s.db, orgName, "")
	user, err := userRepo.GetByID(ctx, orgName, token.UserID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if user == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserInvalidRefreshToken)
	}
	if !user.IsActive() {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserAccountDisabled)
	}

	tokenResult, err := s.issueAccessToken(ctx, token.UserID, orgName, nil, user.IsAdmin)
	if err != nil {
		return nil, err
	}

	if err := txRefreshTokenRepo.Revoke(ctx, token.ID); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	newPlaintext, newTokenHash, err := generateRefreshToken()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate new refresh token")
	}

	newTokenID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate token id")
	}
	expiresAt := time.Now().Add(s.refreshTTL)
	newToken := &domainAuth.RefreshToken{
		ID:        newTokenID,
		UserID:    token.UserID,
		TokenHash: newTokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	if err := txRefreshTokenRepo.Save(ctx, newToken); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	return &RefreshResult{
		UserID:       token.UserID,
		AccessToken:  tokenResult.AccessToken,
		RefreshToken: newPlaintext,
		ExpiresAt:    expiresAt,
	}, nil
}
```

- [ ] **Step 5: 删除不再使用的 `loadUserAndProjects`、`buildAccessToken`、`toAppAccessibleProjects` 函数**

检查这三个函数是否还被其他地方引用：

```bash
cd modelcraft-backend && grep -rn "loadUserAndProjects\|buildAccessToken\|toAppAccessibleProjects" --include="*.go"
```

如果只在 `end_user_auth_service.go` 内部使用，直接删除。

- [ ] **Step 6: 编译检查**

```bash
cd modelcraft-backend && go build ./...
```

预期：无编译错误。

- [ ] **Step 7: Commit**

```bash
git add modelcraft-backend/internal/app/enduser/commands.go \
        modelcraft-backend/internal/app/enduser/end_user_auth_service.go
git commit -m "refactor(enduser): remove projects from login/refresh result"
```

---

## Task 2：后端——修改 `EndUserLogin` handler，去掉 projects 响应字段

**Files:**
- Modify: `modelcraft-backend/internal/interfaces/http/handlers/enduser/auth_handler.go`

- [ ] **Step 1: 修改 `EndUserLogin` handler，移除 `toProjectList` 调用**

在 `auth_handler.go` 的 `EndUserLogin` 方法（约 54 行），将响应构建改为：

```go
func (h *AuthHandler) EndUserLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	var req struct {
		OrgName        string `json:"orgName"`
		Username       string `json:"username"`
		Identifier     string `json:"identifier"`
		IdentifierType string `json:"identifierType"`
		Password       string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "invalid request body")
		return
	}
	result, err := h.authService.LoginEndUser(ctx, appEnduser.LoginCommand{
		OrgName:        req.OrgName,
		Username:       req.Username,
		Identifier:     req.Identifier,
		IdentifierType: appEnduser.IdentifierType(req.IdentifierType),
		Password:       req.Password,
	})
	if err != nil {
		h.handleBizError(w, r, requestID, err, "end-user login failed")
		return
	}

	h.shared.SetRefreshCookie(w, result.RefreshToken)
	h.writeJSON(w, http.StatusOK, map[string]any{
		"requestId":   requestID,
		"userId":      result.UserID,
		"orgName":     result.OrgName,
		"accessToken": result.AccessToken,
		"expiresAt":   result.ExpiresAt.UTC().Format(time.RFC3339),
	})
}
```

- [ ] **Step 2: 修改 `EndUserRefreshToken` handler，同样移除 projects**

找到 `EndUserRefreshToken`（约 142 行），将响应改为：

```go
h.shared.SetRefreshCookie(w, result.RefreshToken)
h.writeJSON(w, http.StatusOK, map[string]any{
	"requestId":   requestID,
	"userId":      result.UserID,
	"orgName":     req.OrgName,
	"accessToken": result.AccessToken,
	"expiresAt":   result.ExpiresAt.UTC().Format(time.RFC3339),
})
```

- [ ] **Step 3: 删除不再使用的 `projectItem`、`toProjectList`、`buildTokenResponse` 函数**

检查引用：

```bash
cd modelcraft-backend && grep -rn "projectItem\|toProjectList\|buildTokenResponse" --include="*.go"
```

如果仅在 `auth_handler.go` 内部使用，直接删除这三个函数（约 266-308 行）。

- [ ] **Step 4: 编译检查**

```bash
cd modelcraft-backend && go build ./...
```

预期：无编译错误。

- [ ] **Step 5: Commit**

```bash
git add modelcraft-backend/internal/interfaces/http/handlers/enduser/auth_handler.go
git commit -m "refactor(enduser): remove projects from login/refresh HTTP response"
```

---

## Task 3：后端——写测试验证登录响应不含 projects

**Files:**
- Modify: `modelcraft-backend/internal/app/enduser/end_user_auth_service_test.go`

- [ ] **Step 1: 在测试文件里找到（或新增）登录成功测试**

打开 `modelcraft-backend/internal/app/enduser/end_user_auth_service_test.go`，找到 `TestLoginEndUser` 或类似函数。

- [ ] **Step 2: 添加断言——`LoginResult` 不含 projects 字段**

在登录成功的测试 case 里，确认 `result.Projects` 字段已不存在（编译即验证），并断言 `result.AccessToken` 非空、`result.OrgName` 正确：

```go
func TestLoginEndUser_NoProjectsInResult(t *testing.T) {
    // ... 现有 setup ...
    result, err := svc.LoginEndUser(ctx, LoginCommand{
        OrgName:    "test-org",
        Identifier: "testuser",
        IdentifierType: IdentifierTypeUsername,
        Password:   "Password123!",
    })
    require.NoError(t, err)
    assert.NotEmpty(t, result.AccessToken)
    assert.Equal(t, "test-org", result.OrgName)
    assert.NotEmpty(t, result.RefreshToken)
    // LoginResult 已无 Projects 字段，编译即验证
}
```

- [ ] **Step 3: 运行测试**

```bash
cd modelcraft-backend && go test ./internal/app/enduser/... -v -run TestLogin
```

预期：PASS。

- [ ] **Step 4: Commit**

```bash
git add modelcraft-backend/internal/app/enduser/end_user_auth_service_test.go
git commit -m "test(enduser): verify login result has no projects field"
```

---

## Task 4：前端——建立单一 auth store，解析 `isAdmin`

**Files:**
- Modify/Create: `modelcraft-front/src/store/authStore.ts`（路径以实际为准）

> 在实施前先运行 `find modelcraft-front/src/store -name "*.ts" | head -20` 确认现有 store 文件路径。

- [ ] **Step 1: 确认现有 auth store 位置**

```bash
find modelcraft-front/src -name "*auth*" -o -name "*Auth*" | grep -i store | head -10
```

- [ ] **Step 2: 在 auth store 中增加 `isAdmin` 字段，从 JWT 解析**

在找到的 auth store 文件中，增加 `isAdmin` 字段和解析逻辑。使用 `atob` 解析 JWT payload（无需安装额外包）：

```typescript
function parseJwtIsAdmin(token: string): boolean {
  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
    return payload.is_admin === true
  } catch {
    return false
  }
}
```

在 store 的 `setToken`（或类似）方法里，调用 `parseJwtIsAdmin` 并存入状态：

```typescript
interface AuthState {
  accessToken: string | null
  userId: string | null
  orgName: string | null
  isAdmin: boolean          // 从 JWT is_admin claim 解析
  expiresAt: number | null
  setToken: (token: string, userId: string, orgName: string) => void
  clearToken: () => void
}

// setToken 实现：
setToken: (token, userId, orgName) => set({
  accessToken: token,
  userId,
  orgName,
  isAdmin: parseJwtIsAdmin(token),
  expiresAt: /* 从 JWT exp 解析，或由 expiresIn 计算 */,
}),
```

- [ ] **Step 3: 运行前端 lint 确认无类型错误**

```bash
cd modelcraft-front && npx tsc --noEmit
```

预期：无错误。

- [ ] **Step 4: Commit**

```bash
git add modelcraft-front/src/store/authStore.ts
git commit -m "feat(front/auth): add isAdmin field parsed from JWT in auth store"
```

---

## Task 5：前端——登录成功后按 `isAdmin` 跳转

**Files:**
- Modify: 管理端登录页（`src/app/login/page.tsx` 或类似）
- Modify: 用户端登录页（`src/app/end-user/[orgName]/login/page.tsx` 或类似）

> 实施前先运行 `find modelcraft-front/src/app -name "page.tsx" | grep -i login` 确认路径。

- [ ] **Step 1: 确认两个登录页文件路径**

```bash
find modelcraft-front/src/app -name "page.tsx" | xargs grep -l "login\|Login" 2>/dev/null | head -10
```

- [ ] **Step 2: 修改登录成功回调，按 `isAdmin` 跳转**

在两个登录页的 `onLoginSuccess` 中，用 auth store 的 `isAdmin` 决定跳转目标：

```typescript
// 登录成功后
authStore.setToken(result.accessToken, result.userId, result.orgName)

if (authStore.isAdmin) {
  router.push(`/org/${result.orgName}/dashboard`)
} else {
  router.push(`/end-user/${result.orgName}/projects`)
}
```

- [ ] **Step 3: 运行 lint**

```bash
cd modelcraft-front && npm run lint
```

预期：无错误。

- [ ] **Step 4: Commit**

```bash
git add modelcraft-front/src/app   # 只 add 改动的登录页
git commit -m "feat(front/auth): redirect after login based on isAdmin"
```

---

## Task 6：前端——管理路由 guard 检查 `isAdmin`

**Files:**
- Modify: `modelcraft-front/src/app/org/[orgName]/layout.tsx`（路径以实际为准）

> 实施前先运行 `find modelcraft-front/src/app/org -name "layout.tsx" | head -5` 确认路径。

- [ ] **Step 1: 确认管理路由 layout 文件**

```bash
find modelcraft-front/src/app/org -name "layout.tsx" | head -5
```

- [ ] **Step 2: 在 layout 中加入 `isAdmin` guard**

在管理路由的 layout（Server Component 用 `redirect`，Client Component 用 `useEffect + router`）中加入：

```typescript
// Client Component 写法示例
'use client'
import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuthStore } from '@/store/authStore'

export default function OrgLayout({ children, params }) {
  const router = useRouter()
  const { isAdmin, orgName } = useAuthStore()

  useEffect(() => {
    if (!isAdmin) {
      router.replace(`/end-user/${params.orgName}/projects`)
    }
  }, [isAdmin, params.orgName, router])

  if (!isAdmin) return null  // 防止闪烁

  return <>{children}</>
}
```

- [ ] **Step 3: 运行 lint + tsc**

```bash
cd modelcraft-front && npm run lint && npx tsc --noEmit
```

预期：无错误。

- [ ] **Step 4: Commit**

```bash
git add modelcraft-front/src/app/org
git commit -m "feat(front/auth): add isAdmin guard to admin route layout"
```

---

## Task 7：前端——视图切换按钮

**Files:**
- Modify: 管理页顶部导航（`src/components/` 或 `src/app/org/[orgName]/` 下的 navbar/header 组件）
- Modify: 用户页顶部导航（`src/app/end-user/[orgName]/` 下的 navbar/header 组件）

> 实施前运行 `find modelcraft-front/src -name "*nav*" -o -name "*Nav*" -o -name "*header*" -o -name "*Header*" | grep -v node_modules | head -20` 确认组件位置。

- [ ] **Step 1: 在管理页导航中加入"切换到用户视图"按钮（仅 `isAdmin=true` 时显示，即永远显示——管理页已有 guard）**

```typescript
import { useRouter } from 'next/navigation'
import { useAuthStore } from '@/store/authStore'

export function AdminNavbar() {
  const router = useRouter()
  const { orgName } = useAuthStore()

  return (
    <nav>
      {/* 现有导航项 */}
      <button
        onClick={() => router.push(`/end-user/${orgName}/projects`)}
      >
        切换到用户视图
      </button>
    </nav>
  )
}
```

- [ ] **Step 2: 在用户页导航中加入"切换到管理视图"按钮（仅 `isAdmin=true` 时显示）**

```typescript
export function EndUserNavbar() {
  const router = useRouter()
  const { orgName, isAdmin } = useAuthStore()

  return (
    <nav>
      {/* 现有导航项 */}
      {isAdmin && (
        <button
          onClick={() => router.push(`/org/${orgName}/dashboard`)}
        >
          切换到管理视图
        </button>
      )}
    </nav>
  )
}
```

- [ ] **Step 3: 运行 lint**

```bash
cd modelcraft-front && npm run lint
```

预期：无错误。

- [ ] **Step 4: Commit**

```bash
git add modelcraft-front/src   # 只 add 改动的导航组件
git commit -m "feat(front/auth): add view-switching buttons to admin and end-user navbars"
```

---

## 验收检查清单

- [ ] `is_admin=true` 用户：登录 → 跳管理页，点"切换到用户视图"→ 跳用户页，不需要重新登录
- [ ] `is_admin=true` 用户：用户页点"切换到管理视图"→ 跳管理页，不需要重新登录
- [ ] `is_admin=false` 用户：登录 → 跳用户页，用户页无"切换到管理视图"按钮
- [ ] `is_admin=false` 用户：直接访问 `/org/xxx/dashboard` → 被 guard 重定向到 `/end-user/xxx/projects`
- [ ] 登录/刷新 API 响应不含 `projects` 字段
- [ ] `go build ./...` 无报错
- [ ] `npm run lint` 无报错
