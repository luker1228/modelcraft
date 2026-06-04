# End-User API Token Runtime Auth — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Allow End-Users to call Runtime GraphQL endpoints using `mc_pat_*` Bearer tokens, without first exchanging for a JWT, plus a frontend "使用示例" dialog on the Token management page.

**Architecture:** A new `ChiRuntimePATMiddleware` in `internal/interfaces/http/middleware/` detects `mc_pat_*` Bearer tokens, validates via `APITokenService`, and injects both `EndUserIdentity` (for RLS resolution) and `ctxutils` user fields (for the downstream JWT middleware's `X-User-ID` short-circuit path). The existing `ChiJWTAuthMiddleware` already short-circuits when `X-User-ID` is present in the header but the runtime routes bypass the gateway — so instead we inject context directly and teach `ChiJWTAuthMiddleware` to short-circuit when `ctxutils.UserID` is already set. On the frontend, a new `ApiUsageDialog` component is added to the Token page table rows.

**Tech Stack:** Go 1.21+, chi v5, `internal/middleware` package, `internal/interfaces/http/middleware` package, `pkg/ctxutils`, Next.js 14, shadcn/ui Dialog, Tailwind CSS

---

## File Map

### Backend

| File | Change |
|------|--------|
| `modelcraft-backend/internal/interfaces/http/middleware/chi_runtime_pat_auth.go` | **Create** — new middleware that handles `mc_pat_*` on runtime end-user routes |
| `modelcraft-backend/internal/interfaces/http/middleware/chi_runtime_pat_auth_test.go` | **Create** — unit tests (3 table-driven cases) |
| `modelcraft-backend/internal/interfaces/http/middleware/chi_jwt_auth.go` | **Modify** — add short-circuit when user is already authenticated via PAT |
| `modelcraft-backend/internal/interfaces/http/routes.go` | **Modify** — `SetupRuntimeGraphQLRoutesOnChi` accepts `*appEnduser.APITokenService`, builds `endUserRuntimeMW` with PAT middleware |
| `modelcraft-backend/internal/interfaces/http/chi_setup.go` | **Modify** — pass `cfg.APITokenService` to `SetupRuntimeGraphQLRoutesOnChi` |

### Frontend

| File | Change |
|------|--------|
| `modelcraft-front/src/app/end-user/[orgName]/dashboard/token/_components/ApiUsageDialog.tsx` | **Create** — Dialog component with Python snippet + copy button |
| `modelcraft-front/src/app/end-user/[orgName]/dashboard/token/page.tsx` | **Modify** — add "使用示例" button per token row, wire to `ApiUsageDialog` |

---

## Task 1: Create `ChiRuntimePATMiddleware`

**Files:**
- Create: `modelcraft-backend/internal/interfaces/http/middleware/chi_runtime_pat_auth.go`

### Background

The `rls_resolver.go` in `internal/interfaces/runtime/` calls `middleware.GetEndUserIdentity(ctx)` — that's `internal/interfaces/http/middleware`, NOT `internal/middleware`. So the new PAT middleware must live in `internal/interfaces/http/middleware` to use the same `endUserContextKey` and `EndUserIdentity` type.

The existing `ChiJWTAuthMiddleware` short-circuits when `X-User-ID` header is present (set by gateway). For the direct PAT path, there's no gateway, so we also set `ctxutils.SetUserID` in context. We will add a context-key short-circuit to the JWT middleware in Task 2 to complete the bypass.

- [ ] **Step 1: Create the file**

```go
package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	appEnduser "modelcraft/internal/app/enduser"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"
)

// ChiRuntimePATMiddleware handles mc_pat_* Bearer tokens for /end-user/ runtime routes.
// On success it injects:
//   - EndUserIdentity into context (read by rls_resolver.go via GetEndUserIdentity)
//   - ctxutils user fields (UserID, OrgName, UserType) so the JWT middleware short-circuits
//
// Non-PAT requests pass through unchanged (JWT middleware handles them).
func ChiRuntimePATMiddleware(
	svc *appEnduser.APITokenService,
	logger logfacade.Logger,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bearer := r.Header.Get("Authorization")
			if !strings.HasPrefix(bearer, "Bearer mc_pat_") {
				next.ServeHTTP(w, r)
				return
			}

			plaintext := strings.TrimPrefix(bearer, "Bearer ")
			token, err := svc.ValidateToken(r.Context(), plaintext)
			if err != nil || token == nil {
				if logger != nil {
					logger.Warnf(r.Context(), "runtime PAT validation failed: %v", err)
				}
				writeJSONError(w, http.StatusUnauthorized, "invalid or expired API token", "UNAUTHENTICATED")
				return
			}

			// Fire-and-forget: update last_used_at
			go func() {
				if updateErr := svc.UpdateLastUsedAt(context.Background(), token.ID, time.Now()); updateErr != nil {
					if logger != nil {
						logger.Warnf(context.Background(), "update last_used_at failed: %v", updateErr)
					}
				}
			}()

			// Inject EndUserIdentity for RLS resolver (GetEndUserIdentity reads this)
			identity := &EndUserIdentity{
				EndUserID: token.EndUserID,
				Issuer:    issuerPlatform,
			}
			ctx := context.WithValue(r.Context(), endUserContextKey, identity)

			// Also inject ctxutils fields so ChiJWTAuthMiddleware short-circuits
			ctx = ctxutils.SetUserID(ctx, token.EndUserID)
			ctx = ctxutils.SetOrgName(ctx, token.OrgName)
			ctx = ctxutils.SetUserType(ctx, ctxutils.UserTypeEndUser)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd modelcraft-backend && go build ./internal/interfaces/http/middleware/...
```

Expected: no output (clean build).

---

## Task 2: Add short-circuit to `ChiJWTAuthMiddleware`

**Files:**
- Modify: `modelcraft-backend/internal/interfaces/http/middleware/chi_jwt_auth.go`

The runtime routes don't go through the gateway, so there's no `X-User-ID` header. But when PAT auth succeeds, `ctxutils.UserID` is already set in context. We teach the JWT middleware to short-circuit in that case.

- [ ] **Step 1: Read the current file to confirm exact text to edit**

The function body of `ChiJWTAuthMiddleware` starts after `return func(next http.Handler) http.Handler {`. The first check is `if config.SkipValidation`. Insert the new check right after that block.

- [ ] **Step 2: Add the short-circuit**

In `modelcraft-backend/internal/interfaces/http/middleware/chi_jwt_auth.go`, find this exact block:

```go
		if config.SkipValidation {
				next.ServeHTTP(w, r)
				return
			}

			if tryInternalTokenAuth(config, w, r, next) {
```

Replace with:

```go
			if config.SkipValidation {
				next.ServeHTTP(w, r)
				return
			}

			// Short-circuit: PAT middleware already authenticated this request
			// (UserID set in context by ChiRuntimePATMiddleware).
			if uid, err := ctxutils.GetUserIDFromContext(r.Context()); err == nil && uid != "" {
				next.ServeHTTP(w, r)
				return
			}

			if tryInternalTokenAuth(config, w, r, next) {
```

- [ ] **Step 3: Verify it compiles**

```bash
cd modelcraft-backend && go build ./internal/interfaces/http/middleware/...
```

Expected: no output.

---

## Task 3: Write unit tests for `ChiRuntimePATMiddleware`

**Files:**
- Create: `modelcraft-backend/internal/interfaces/http/middleware/chi_runtime_pat_auth_test.go`

- [ ] **Step 1: Write the test file**

```go
package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domainenduser "modelcraft/internal/domain/enduser"
	"modelcraft/pkg/ctxutils"
)

// mockPATService implements the subset of appEnduser.APITokenService needed for tests.
// We call the real middleware by passing a *appEnduser.APITokenService, but because
// APITokenService.ValidateToken is a method we cannot easily swap. Instead we test
// the middleware indirectly by constructing a real APITokenService with a fake repo.
//
// Approach: use a table-driven helper that builds the middleware with a real service
// backed by a stubTokenRepo.

// stubTokenRepo implements domainenduser.APITokenRepository for test purposes.
type stubTokenRepo struct {
	token *domainenduser.APIToken
	err   error
}

func (s *stubTokenRepo) Save(_ context.Context, _ *domainenduser.APIToken) error {
	return nil
}
func (s *stubTokenRepo) FindByHash(_ context.Context, _ string) (*domainenduser.APIToken, error) {
	return s.token, s.err
}
func (s *stubTokenRepo) ListByUser(_ context.Context, _, _ string) ([]*domainenduser.APIToken, error) {
	return nil, nil
}
func (s *stubTokenRepo) SoftDelete(_ context.Context, _, _, _ string) error { return nil }
func (s *stubTokenRepo) UpdateLastUsedAt(_ context.Context, _ string, _ time.Time) error {
	return nil
}

func TestChiRuntimePATMiddleware(t *testing.T) {
	validToken := &domainenduser.APIToken{
		ID:          "tok-1",
		OrgName:     "acme",
		EndUserID:   "eu-42",
		Name:        "my-token",
		TokenHash:   "ignored-by-stub",
		DeletedAt:   0,
	}

	cases := []struct {
		name            string
		authHeader      string
		repoToken       *domainenduser.APIToken
		repoErr         error
		wantStatus      int
		wantNextCalled  bool
		wantEndUserID   string // empty means don't check
	}{
		{
			name:           "non-PAT Bearer passes through unchanged",
			authHeader:     "Bearer eyJhbGciOiJFUzI1NiJ9.payload.sig",
			repoToken:      nil,
			repoErr:        nil,
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
			wantEndUserID:  "", // identity NOT injected
		},
		{
			name:           "mc_pat_ token invalid → 401",
			authHeader:     "Bearer mc_pat_invalid",
			repoToken:      nil,
			repoErr:        errors.New("not found"),
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "mc_pat_ token valid → identity injected, next called",
			authHeader:     "Bearer mc_pat_valid",
			repoToken:      validToken,
			repoErr:        nil,
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
			wantEndUserID:  "eu-42",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			import_appEnduser_stub := buildAPITokenServiceFromStub(&stubTokenRepo{
				token: tc.repoToken,
				err:   tc.repoErr,
			})

			nextCalled := false
			var gotEndUserID string
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				if identity := GetEndUserIdentity(r.Context()); identity != nil {
					gotEndUserID = identity.EndUserID
				}
				w.WriteHeader(http.StatusOK)
			})

			mw := ChiRuntimePATMiddleware(import_appEnduser_stub, nil)
			handler := mw(next)

			req := httptest.NewRequest(http.MethodPost, "/end-user/graphql/org/acme/project/p/db/d/model/m", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("status: want %d got %d", tc.wantStatus, rr.Code)
			}
			if nextCalled != tc.wantNextCalled {
				t.Errorf("nextCalled: want %v got %v", tc.wantNextCalled, nextCalled)
			}
			if tc.wantEndUserID != "" && gotEndUserID != tc.wantEndUserID {
				t.Errorf("endUserID in context: want %q got %q", tc.wantEndUserID, gotEndUserID)
			}
			if tc.name == "non-PAT Bearer passes through unchanged" {
				// ctxutils UserID must NOT be set (we didn't authenticate)
				if uid, err := ctxutils.GetUserIDFromContext(req.Context()); err == nil && uid != "" {
					t.Errorf("UserID must not be injected for non-PAT pass-through, got %q", uid)
				}
			}
		})
	}
}
```

> **Note:** `buildAPITokenServiceFromStub` is a test helper (step 2 below).

- [ ] **Step 2: Add the test helper**

Append to the same file:

```go
// buildAPITokenServiceFromStub creates a real *appEnduser.APITokenService backed by a stub repo.
// Import the app package inline to avoid circular deps.
func buildAPITokenServiceFromStub(repo domainenduser.APITokenRepository) *appEnduser_pkg.APITokenService {
	return appEnduser_pkg.NewAPITokenService(repo)
}
```

Add the import at the top of the file:

```go
import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domainenduser "modelcraft/internal/domain/enduser"
	appEnduser_pkg "modelcraft/internal/app/enduser"
	"modelcraft/pkg/ctxutils"
)
```

- [ ] **Step 3: Run the tests (expect FAIL — middleware not wired yet, but logic is testable)**

```bash
cd modelcraft-backend && go test ./internal/interfaces/http/middleware/... -run TestChiRuntimePATMiddleware -v
```

Expected: all 3 subtests PASS (the middleware function itself is complete; routing wiring is separate).

- [ ] **Step 4: Commit**

```bash
cd modelcraft-backend
git add internal/interfaces/http/middleware/chi_runtime_pat_auth.go \
        internal/interfaces/http/middleware/chi_runtime_pat_auth_test.go \
        internal/interfaces/http/middleware/chi_jwt_auth.go
git commit -m "feat(middleware): add ChiRuntimePATMiddleware for end-user runtime PAT auth

- New ChiRuntimePATMiddleware validates mc_pat_* Bearer tokens and injects
  EndUserIdentity + ctxutils fields for downstream RLS resolution
- ChiJWTAuthMiddleware short-circuits when UserID already in context (PAT path)
- 3-case table-driven unit tests"
```

---

## Task 4: Wire PAT middleware into runtime routes

**Files:**
- Modify: `modelcraft-backend/internal/interfaces/http/routes.go`

- [ ] **Step 1: Update `SetupRuntimeGraphQLRoutesOnChi` signature and body**

Find the function signature:

```go
func SetupRuntimeGraphQLRoutesOnChi(router chi.Router, handlers *RuntimeHandlers, cfg *config.Config) {
```

Replace with:

```go
func SetupRuntimeGraphQLRoutesOnChi(
	router chi.Router,
	handlers *RuntimeHandlers,
	cfg *config.Config,
	apiTokenSvc *appEnduser.APITokenService,
) {
```

Add the import if not already present (it already is — `appEnduser` is imported):
```go
appEnduser "modelcraft/internal/app/enduser"
```

- [ ] **Step 2: Build two separate middleware closures**

Find the block inside the function:

```go
	runtimeMW := func(next http.Handler) http.Handler {
		orgMW := middleware.ChiGraphQLOrgMiddleware()
		jwtMW := middleware.ChiJWTAuthMiddleware(jwtConfig)
		cacheMW := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				useCache := req.URL.Query().Get("useCache") != "false"
				next.ServeHTTP(w, req.WithContext(ctxutils.SetUseCache(req.Context(), useCache)))
			})
		}
		return requestIDInjectorMiddleware(jwtMW(orgMW(cacheMW(next))))
	}
```

Replace with:

```go
	logger := logfacade.GetLogger(context.Background())

	orgMW := middleware.ChiGraphQLOrgMiddleware()
	jwtMW := middleware.ChiJWTAuthMiddleware(jwtConfig)
	cacheMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			useCache := req.URL.Query().Get("useCache") != "false"
			next.ServeHTTP(w, req.WithContext(ctxutils.SetUseCache(req.Context(), useCache)))
		})
	}

	// Design-time tenant runtime: JWT only (same as before)
	runtimeMW := func(next http.Handler) http.Handler {
		return requestIDInjectorMiddleware(jwtMW(orgMW(cacheMW(next))))
	}

	// End-user runtime: PAT Token takes priority, JWT is fallback
	var patMW func(http.Handler) http.Handler
	if apiTokenSvc != nil {
		patMW = middleware.ChiRuntimePATMiddleware(apiTokenSvc, logger)
	} else {
		patMW = func(next http.Handler) http.Handler { return next }
	}
	endUserRuntimeMW := func(next http.Handler) http.Handler {
		return requestIDInjectorMiddleware(patMW(jwtMW(orgMW(cacheMW(next)))))
	}
```

- [ ] **Step 3: Update the end-user runtime route registrations**

Find:

```go
	// End-user runtime routes — same handler, end-user JWT identity injected by APISIX.
	// No X-Action middleware: runtime queries are schema-driven and don't need operation validation.
	endUserRuntimePath := "/end-user/graphql/org/{orgName}/project/{projectSlug}/db/{db}/model/{model}"
	router.With(runtimeMW).Get(endUserRuntimePath, handlers.ModelRuntimeHandler.HandlePlayground)
	router.With(runtimeMW).Post(endUserRuntimePath, handlers.ModelRuntimeHandler.HandleQuery)
```

Replace with:

```go
	// End-user runtime routes — PAT Bearer token or gateway-injected JWT.
	endUserRuntimePath := "/end-user/graphql/org/{orgName}/project/{projectSlug}/db/{db}/model/{model}"
	router.With(endUserRuntimeMW).Get(endUserRuntimePath, handlers.ModelRuntimeHandler.HandlePlayground)
	router.With(endUserRuntimeMW).Post(endUserRuntimePath, handlers.ModelRuntimeHandler.HandleQuery)
```

- [ ] **Step 4: Verify the build**

```bash
cd modelcraft-backend && go build ./internal/interfaces/http/...
```

Expected: compiler error about `SetupRuntimeGraphQLRoutesOnChi` call site — fix in Task 5.

---

## Task 5: Update call site in `chi_setup.go`

**Files:**
- Modify: `modelcraft-backend/internal/interfaces/http/chi_setup.go`

- [ ] **Step 1: Update the call to `SetupRuntimeGraphQLRoutesOnChi`**

Find:

```go
	if cfg.RuntimeHandlers != nil {
		SetupRuntimeGraphQLRoutesOnChi(r, cfg.RuntimeHandlers, cfg.Config)
	}
```

Replace with:

```go
	if cfg.RuntimeHandlers != nil {
		SetupRuntimeGraphQLRoutesOnChi(r, cfg.RuntimeHandlers, cfg.Config, cfg.APITokenService)
	}
```

(`cfg.APITokenService` is already declared in `ChiRouterConfig` — field `APITokenService *appEnduser.APITokenService`)

- [ ] **Step 2: Full build and existing tests pass**

```bash
cd modelcraft-backend && go build ./... && go test ./internal/interfaces/http/middleware/... -v
```

Expected: build clean, all middleware tests pass.

- [ ] **Step 3: Commit**

```bash
cd modelcraft-backend
git add internal/interfaces/http/routes.go \
        internal/interfaces/http/chi_setup.go
git commit -m "feat(routes): wire ChiRuntimePATMiddleware into end-user runtime routes

SetupRuntimeGraphQLRoutesOnChi now accepts *APITokenService and builds
a separate endUserRuntimeMW that prepends PAT auth before JWT for the
/end-user/graphql/.../db/.../model/... routes.
Design-time /graphql/ routes are unchanged."
```

---

## Task 6: Create `ApiUsageDialog` frontend component

**Files:**
- Create: `modelcraft-front/src/app/end-user/[orgName]/dashboard/token/_components/ApiUsageDialog.tsx`

- [ ] **Step 1: Create the `_components` directory and file**

```bash
mkdir -p modelcraft-front/src/app/end-user/\[orgName\]/dashboard/token/_components
```

- [ ] **Step 2: Write the component**

```tsx
'use client'

import { useState } from 'react'
import { Check, Copy, Terminal } from 'lucide-react'
import { Button } from '@web/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@web/components/ui/dialog'

interface ApiUsageDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  orgName: string
  tokenName: string
}

function buildPythonSnippet(orgName: string): string {
  return `import os
import requests

# 从环境变量读取 API Token（避免硬编码）
TOKEN = os.environ["MC_API_TOKEN"]

# 替换为你的实际参数
ORG_NAME     = "${orgName}"   # 已自动填入
PROJECT_SLUG = "your-project"
DB_NAME      = "your-db"
MODEL_NAME   = "your-model"

ENDPOINT = (
    f"http://localhost:8080/end-user/graphql"
    f"/org/{ORG_NAME}/project/{PROJECT_SLUG}"
    f"/db/{DB_NAME}/model/{MODEL_NAME}"
)

# GraphQL 查询示例：查询前 10 条记录
query = """
query {
  list(limit: 10) {
    id
  }
}
"""

resp = requests.post(
    ENDPOINT,
    json={"query": query},
    headers={"Authorization": f"Bearer {TOKEN}"},
)
resp.raise_for_status()
print(resp.json())`
}

function CopyCodeButton({ code }: { code: string }) {
  const [copied, setCopied] = useState(false)

  const handleCopy = () => {
    void navigator.clipboard.writeText(code).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }

  return (
    <Button
      variant="outline"
      size="sm"
      className="h-7 gap-1.5 px-2.5 text-xs"
      onClick={handleCopy}
    >
      {copied ? (
        <Check className="size-3.5 text-emerald-500" />
      ) : (
        <Copy className="size-3.5" />
      )}
      {copied ? '已复制' : '复制代码'}
    </Button>
  )
}

export function ApiUsageDialog({
  open,
  onOpenChange,
  orgName,
  tokenName,
}: ApiUsageDialogProps) {
  const snippet = buildPythonSnippet(orgName)

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Terminal className="size-4 text-primary" />
            使用 Token 调用 Runtime API
          </DialogTitle>
          <DialogDescription>
            Token「<span className="font-mono font-medium">{tokenName}</span>
            」可直接用于以下端点的 Bearer 认证。
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {/* Endpoint */}
          <div className="space-y-1.5">
            <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              端点
            </p>
            <pre className="overflow-x-auto rounded-md border bg-muted/40 p-3 font-mono text-xs leading-5 text-foreground">
              {`POST /end-user/graphql/org/{orgName}/project/{projectSlug}\n     /db/{db}/model/{model}`}
            </pre>
          </div>

          {/* Auth header */}
          <div className="space-y-1.5">
            <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              认证方式
            </p>
            <pre className="overflow-x-auto rounded-md border bg-muted/40 p-3 font-mono text-xs text-foreground">
              {`Authorization: Bearer <your-token>`}
            </pre>
          </div>

          {/* Python snippet */}
          <div className="space-y-1.5">
            <div className="flex items-center justify-between">
              <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                Python 示例
              </p>
              <CopyCodeButton code={snippet} />
            </div>
            <pre className="max-h-72 overflow-y-auto overflow-x-auto rounded-md border bg-[#F6F8FA] p-4 font-mono text-xs leading-5 text-foreground">
              {snippet}
            </pre>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
```

- [ ] **Step 3: Verify TypeScript compiles**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | grep ApiUsageDialog
```

Expected: no output (no errors in the new file).

---

## Task 7: Wire `ApiUsageDialog` into the Token page

**Files:**
- Modify: `modelcraft-front/src/app/end-user/[orgName]/dashboard/token/page.tsx`

- [ ] **Step 1: Import the new component**

Add to the existing import block at the top of `page.tsx`:

```tsx
import { ApiUsageDialog } from './_components/ApiUsageDialog'
```

- [ ] **Step 2: Add state for the dialog**

Inside `TokenPageContent`, after the existing state declarations:

```tsx
const [usageTarget, setUsageTarget] = useState<APIToken | null>(null)
```

- [ ] **Step 3: Add "使用示例" button to each token row**

Find the actions `<td>` in the token row (the one containing the `Trash2` delete button):

```tsx
                        <td className="px-4 py-3 text-right">
                          <Button
                            variant="ghost"
                            size="sm"
                            className="size-7 p-0 text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                            title="撤销 Token"
                            onClick={() => setRevokeTarget(token)}
                          >
                            <Trash2 className="size-3.5" />
                          </Button>
                        </td>
```

Replace with:

```tsx
                        <td className="px-4 py-3 text-right">
                          <div className="flex items-center justify-end gap-1">
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-7 gap-1 px-2 text-xs text-muted-foreground transition-colors hover:bg-muted"
                              title="使用示例"
                              onClick={() => setUsageTarget(token)}
                            >
                              <Terminal className="size-3.5" />
                              示例
                            </Button>
                            <Button
                              variant="ghost"
                              size="sm"
                              className="size-7 p-0 text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                              title="撤销 Token"
                              onClick={() => setRevokeTarget(token)}
                            >
                              <Trash2 className="size-3.5" />
                            </Button>
                          </div>
                        </td>
```

- [ ] **Step 4: Add the `ApiUsageDialog` mount**

Find the `{/* Dialogs */}` section at the bottom of `TokenPageContent`'s return:

```tsx
      {/* Dialogs */}
      <CreateTokenDialog
```

Add the new dialog just before `</EndUserAppLayout>`:

```tsx
      <ApiUsageDialog
        open={!!usageTarget}
        onOpenChange={(o) => { if (!o) setUsageTarget(null) }}
        orgName={orgName}
        tokenName={usageTarget?.name ?? ''}
      />
```

- [ ] **Step 5: Verify TypeScript and lint**

```bash
cd modelcraft-front && npx tsc --noEmit && npm run lint 2>&1 | grep -E "(error|warning)" | head -20
```

Expected: no TypeScript errors. Lint warnings for the new file only (if any) are acceptable; errors are not.

- [ ] **Step 6: Commit**

```bash
cd modelcraft-front
git add src/app/end-user/\[orgName\]/dashboard/token/_components/ApiUsageDialog.tsx \
        src/app/end-user/\[orgName\]/dashboard/token/page.tsx
git commit -m "feat(token-page): add API usage dialog with Python snippet

Each token row now has a '示例' button that opens a dialog showing
the Runtime GraphQL endpoint, auth header format, and a complete
runnable Python example with orgName auto-filled."
```

---

## Self-Review

**Spec coverage check:**

| Spec requirement | Task |
|-----------------|------|
| `mc_pat_*` Bearer auth on `/end-user/` runtime routes | Tasks 1, 4 |
| JWT path completely unaffected for non-PAT tokens | Task 2 (short-circuit only triggers when uid already set) |
| `EndUserIdentity` identical structure (Issuer = "mc-platform") | Task 1 |
| Only `/end-user/` routes affected, not `/graphql/` | Task 4 (separate `runtimeMW` vs `endUserRuntimeMW`) |
| `UpdateLastUsedAt` called | Task 1 |
| Frontend: "使用示例" button per token row | Task 7 |
| Frontend: Dialog with endpoint + auth + Python snippet | Task 6 |
| `orgName` auto-filled in Python snippet | Task 6 (`buildPythonSnippet(orgName)`) |
| One-click copy button | Task 6 (`CopyCodeButton`) |
| Purely presentational, no network requests | Task 6 (no Apollo/fetch) |

**Placeholder scan:** None found.

**Type consistency:**
- `ChiRuntimePATMiddleware` signature: `(svc *appEnduser.APITokenService, logger logfacade.Logger)` — consistent across Tasks 1, 4
- `endUserContextKey` / `EndUserIdentity` / `issuerPlatform` — all defined in `runtime_auth_middleware.go` in the same package; new file is in the same package, so they're directly accessible
- `ApiUsageDialog` props `{ open, onOpenChange, orgName, tokenName }` — consistent between Task 6 definition and Task 7 usage
- `usageTarget` is `APIToken | null`, `tokenName={usageTarget?.name ?? ''}` — safe null coalesce
